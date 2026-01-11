"""
{{ project_name }} - Fabric Deployment & GCP Setup Tasks

Usage:
    # One-time GCP project setup
    fab setup --project=myproject --billing=XXXXXX-XXXXXX-XXXXXX
    fab setup --project=myproject-staging --billing=XXXXXX-XXXXXX-XXXXXX --staging

    # Day-to-day operations
    fab deploy              # Deploy to production
    fab deploy --env=staging  # Deploy to staging
    fab build               # Build Docker image only
    fab migrate             # Run migrations on Cloud Run
    fab logs                # View Cloud Run logs
    fab secrets-download    # Download secrets from Secret Manager
    fab secrets-upload      # Upload secrets to Secret Manager
    fab db-export           # Export database to GCS
    fab db-import           # Import database from GCS

Configuration:
    Set these environment variables or create a .env file:
    - GCP_PROJECT_ID: Your GCP project ID
    - GCP_REGION: GCP region (default: europe-west2)
    - CLOUD_SQL_INSTANCE: Cloud SQL instance name
    - CLOUD_SQL_PROJECT: Project containing Cloud SQL (if different)
    - GCP_BILLING_ACCOUNT: Billing account for setup (optional, can pass as arg)
"""
import os
import secrets
import string
from fabric import task
from invoke import Context

# Load environment variables
try:
    from dotenv import load_dotenv
    load_dotenv()
except ImportError:
    pass

# Configuration - UPDATE THESE FOR YOUR PROJECT
GCP_PROJECT_ID = os.getenv("GCP_PROJECT_ID", "{{ project_slug }}")
GCP_REGION = os.getenv("GCP_REGION", "{{ gcp_region }}")
CLOUD_SQL_INSTANCE = os.getenv("CLOUD_SQL_INSTANCE", "")
CLOUD_SQL_PROJECT = os.getenv("CLOUD_SQL_PROJECT", "{{ gcp_project }}")
SERVICE_NAME = os.getenv("SERVICE_NAME", "{{ project_slug }}")
GCP_BILLING_ACCOUNT = os.getenv("GCP_BILLING_ACCOUNT", "00139C-8D2D10-3919FA")
GCP_ORGANIZATION_ID = os.getenv("GCP_ORGANIZATION_ID", "")

# Colors for output
GREEN = "\033[0;32m"
YELLOW = "\033[1;33m"
RED = "\033[0;31m"
NC = "\033[0m"


def log_info(msg):
    print(f"{GREEN}[INFO]{NC} {msg}")


def log_warn(msg):
    print(f"{YELLOW}[WARN]{NC} {msg}")


def log_error(msg):
    print(f"{RED}[ERROR]{NC} {msg}")


def generate_password(length=30):
    """Generate a secure random password."""
    alphabet = string.ascii_letters + string.digits
    return ''.join(secrets.choice(alphabet) for _ in range(length))


def get_env_config(env: str) -> dict:
    """Get configuration for the specified environment."""
    configs = {
        "production": {
            "service": SERVICE_NAME,
            "settings": f"{SERVICE_NAME}.settings.cloud_production",
            "secrets_name": "application_settings",
            "min_instances": 0,
            "max_instances": 10,
        },
        "staging": {
            "service": f"{SERVICE_NAME}-staging",
            "settings": f"{SERVICE_NAME}.settings.cloud_staging",
            "secrets_name": "application_settings_staging",
            "min_instances": 0,
            "max_instances": 2,
        },
    }
    return configs.get(env, configs["production"])


@task
def build(c, env="production"):
    """Build Docker image using Cloud Build."""
    config = get_env_config(env)
    image = f"gcr.io/{GCP_PROJECT_ID}/{config['service']}"

    print(f"Building image with Cloud Build: {image}")
    c.run(f"""gcloud builds submit \\
        --tag {image} \\
        --project {GCP_PROJECT_ID} \\
        --timeout=30m""", pty=True)
    print(f"Image built: {image}")


@task
def deploy(c, env="production"):
    """Build and deploy to Cloud Run."""
    config = get_env_config(env)
    image = f"gcr.io/{GCP_PROJECT_ID}/{config['service']}"

    # Build and push
    build(c, env=env)

    # Deploy to Cloud Run
    print(f"Deploying to Cloud Run: {config['service']}")
    cmd = f"""gcloud run deploy {config['service']} \\
        --image {image} \\
        --platform managed \\
        --region {GCP_REGION} \\
        --project {GCP_PROJECT_ID} \\
        --add-cloudsql-instances {CLOUD_SQL_PROJECT}:{GCP_REGION}:{CLOUD_SQL_INSTANCE} \\
        --set-env-vars DJANGO_SETTINGS_MODULE={config['settings']},GCP_PROJECT_ID={GCP_PROJECT_ID} \\
        --min-instances {config['min_instances']} \\
        --max-instances {config['max_instances']} \\
        --allow-unauthenticated"""
    c.run(cmd, pty=True)
    print(f"Deployed: {config['service']}")


@task
def migrate(c, env="production"):
    """Run Django migrations via Cloud Build."""
    config = get_env_config(env)

    print(f"Running migrations for {env}...")
    c.run(f"""gcloud builds submit \\
        --config cloudmigrate.yaml \\
        --project {GCP_PROJECT_ID} \\
        --substitutions _DJANGO_SETTINGS_MODULE={config['settings']} \\
        --timeout=30m""", pty=True)


@task
def logs(c, env="production"):
    """View Cloud Run logs."""
    config = get_env_config(env)
    c.run(f"gcloud run services logs read {config['service']} --region {GCP_REGION} --project {GCP_PROJECT_ID}", pty=True)


@task
def createsuperuser(c, email, password, env="production"):
    """Create a Django superuser via Cloud Build.
    
    Usage: fab createsuperuser --email=admin@example.com --password=secret123
    """
    config = get_env_config(env)

    print(f"Creating superuser {email} for {env}...")

    # Create a temporary cloudbuild config for createsuperuser
    cloudbuild_config = f"""
steps:
  - name: 'gcr.io/google-appengine/exec-wrapper'
    args:
      - '-i'
      - 'gcr.io/{GCP_PROJECT_ID}/{config["service"]}'
      - '-s'
      - '{CLOUD_SQL_PROJECT}:{GCP_REGION}:{CLOUD_SQL_INSTANCE}'
      - '-e'
      - 'DJANGO_SETTINGS_MODULE={config["settings"]}'
      - '-e'
      - 'DJANGO_SUPERUSER_EMAIL={email}'
      - '-e'
      - 'DJANGO_SUPERUSER_PASSWORD={password}'
      - '--'
      - 'python'
      - 'manage.py'
      - 'createsuperuser'
      - '--noinput'
timeout: '600s'
"""
    
    import tempfile
    with tempfile.NamedTemporaryFile(mode='w', suffix='.yaml', delete=False) as f:
        f.write(cloudbuild_config)
        config_file = f.name
    
    try:
        c.run(f"""gcloud builds submit \\
            --config {config_file} \\
            --project {GCP_PROJECT_ID} \\
            --no-source \\
            --timeout=10m""", pty=True)
        print(f"Superuser {email} created successfully!")
    finally:
        os.unlink(config_file)


@task(name="secrets-download")
def secrets_download(c, env="production"):
    """Download secrets from Secret Manager to .env file."""
    config = get_env_config(env)
    output_file = f".env.{env}"

    print(f"Downloading secrets to {output_file}...")
    c.run(f"""gcloud secrets versions access latest \\
        --secret="{config['secrets_name']}" \\
        --project={GCP_PROJECT_ID} \\
        --format="value(payload.data)" > {output_file}""")
    print(f"Secrets saved to {output_file}")


@task(name="secrets-upload")
def secrets_upload(c, env="production", file=None):
    """Upload secrets from .env file to Secret Manager."""
    config = get_env_config(env)
    input_file = file or f".env.{env}"

    print(f"Uploading secrets from {input_file}...")
    c.run(f"""gcloud secrets versions add {config['secrets_name']} \\
        --data-file={input_file} \\
        --project={GCP_PROJECT_ID}""", pty=True)
    print(f"Secrets uploaded to {config['secrets_name']}")


@task(name="db-export")
def db_export(c, database=None):
    """Export database to GCS bucket."""
    db = database or GCP_PROJECT_ID
    bucket = GCP_PROJECT_ID

    print(f"Exporting database {db} to gs://{bucket}/{db}.gz...")
    c.run(f"""gcloud sql export sql {CLOUD_SQL_INSTANCE} \\
        gs://{bucket}/{db}.gz \\
        --database={db} \\
        --project={CLOUD_SQL_PROJECT}""", pty=True)


@task(name="db-import")
def db_import(c, file, database=None):
    """Import database from GCS bucket."""
    db = database or GCP_PROJECT_ID

    print(f"Importing {file} to database {db}...")
    c.run(f"""gcloud sql import sql {CLOUD_SQL_INSTANCE} \\
        {file} \\
        --database={db} \\
        --project={CLOUD_SQL_PROJECT}""", pty=True)


@task(name="db-download")
def db_download(c, database=None):
    """Export and download database locally."""
    db = database or GCP_PROJECT_ID
    bucket = GCP_PROJECT_ID

    # Export to GCS
    db_export(c, database=db)

    # Download locally
    print(f"Downloading gs://{bucket}/{db}.gz...")
    c.run(f"gsutil cp gs://{bucket}/{db}.gz .")
    c.run(f"gunzip -f {db}.gz")
    print(f"Database saved to {db}")


@task(name="media-download")
def media_download(c, bucket=None):
    """Download media files from GCS."""
    bucket = bucket or GCP_PROJECT_ID
    print(f"Downloading media from gs://{bucket}/media...")
    c.run(f"gsutil -m cp -r gs://{bucket}/media .", pty=True)


@task(name="media-upload")
def media_upload(c, bucket=None):
    """Upload media files to GCS."""
    bucket = bucket or GCP_PROJECT_ID
    print(f"Uploading media to gs://{bucket}/media...")
    c.run(f"gsutil -m cp -r media gs://{bucket}/", pty=True)
    c.run(f"gsutil -m acl set -R -a public-read gs://{bucket}/media", pty=True)


@task
def shell(c, env="production"):
    """Open a Django shell on Cloud Run (via Cloud Build)."""
    config = get_env_config(env)
    print("Note: This runs a one-off container. For interactive shell, use local development.")
    c.run(f"""gcloud builds submit \\
        --config cloudshell.yaml \\
        --project {GCP_PROJECT_ID} \\
        --substitutions _DJANGO_SETTINGS_MODULE={config['settings']} \\
        --timeout=30m""", pty=True)


@task
def status(c, env="production"):
    """Show Cloud Run service status."""
    config = get_env_config(env)
    c.run(f"gcloud run services describe {config['service']} --region {GCP_REGION} --project {GCP_PROJECT_ID}", pty=True)


@task
def collectstatic(c):
    """Run collectstatic locally (for debugging)."""
    c.run("python manage.py collectstatic --noinput", pty=True)


# =============================================================================
# GCP PROJECT SETUP TASKS
# =============================================================================


@task
def setup(c, project, billing=None, staging=False, region=None, sql_instance=None, sql_project=None):
    """
    Set up a new GCP project for Django on Cloud Run.

    Creates all GCP resources needed:
    - GCP Project (or uses existing)
    - Cloud Run, Cloud SQL, Secret Manager, Cloud Build APIs
    - Cloud Storage bucket for media files
    - Cloud SQL database (on shared instance)
    - Secret Manager secrets for Django settings
    - IAM permissions for Cloud Run and Cloud Build

    Usage:
        fab setup --project=myproject --billing=XXXXXX-XXXXXX-XXXXXX
        fab setup --project=myproject-staging --billing=XXXXXX-XXXXXX-XXXXXX --staging
    """
    # Configuration
    billing_account = billing or GCP_BILLING_ACCOUNT
    region = region or GCP_REGION
    sql_instance = sql_instance or CLOUD_SQL_INSTANCE
    sql_project = sql_project or CLOUD_SQL_PROJECT
    org_id = GCP_ORGANIZATION_ID

    if not billing_account:
        log_error("Billing account is required.")
        print("Pass --billing=XXXXXX-XXXXXX-XXXXXX or set GCP_BILLING_ACCOUNT env var")
        return

    secrets_name = "application_settings_staging" if staging else "application_settings"
    bucket_name = project

    log_info(f"Setting up GCP project: {project}")
    log_info(f"Region: {region}")
    log_info(f"Staging: {staging}")
    print()

    # Create or select project
    setup_create_project(c, project, org_id)

    # Link billing
    setup_link_billing(c, project, billing_account)

    # Enable APIs
    setup_enable_apis(c, project)

    # Get service account emails
    cloudrun_sa, cloudbuild_sa = setup_get_service_accounts(c, project)

    # IAM permissions
    setup_iam_permissions(c, project, cloudrun_sa, cloudbuild_sa, sql_project)

    # Create database
    db_password = setup_create_database(c, project, sql_instance, sql_project)

    # Create storage bucket
    setup_create_bucket(c, project, bucket_name, region)

    # Create secrets
    setup_create_secrets(c, project, secrets_name, bucket_name, db_password,
                         sql_project, region, sql_instance, cloudrun_sa, cloudbuild_sa)

    # Summary
    print()
    log_info("==========================================")
    log_info("GCP Project Setup Complete!")
    log_info("==========================================")
    print()
    print(f"Project ID:     {project}")
    print(f"Region:         {region}")
    print(f"Database:       {project} on {sql_instance}")
    print(f"Storage Bucket: gs://{bucket_name}")
    print(f"Secrets:        {secrets_name}")
    print()
    print("Next steps:")
    print("  1. Update your .env file:")
    print(f"     GCP_PROJECT_ID={project}")
    print(f"     GCP_REGION={region}")
    env_flag = "--env=staging" if staging else ""
    print(f"  2. Deploy: fab deploy {env_flag}")
    print(f"  3. Run migrations: fab migrate {env_flag}")
    print()
    log_info("Done!")


def setup_create_project(c, project, org_id):
    """Create or select GCP project."""
    log_info("Creating/selecting project...")
    try:
        if org_id:
            c.run(f'gcloud projects create "{project}" --organization "{org_id}"',
                  warn=True, hide=True)
        else:
            c.run(f'gcloud projects create "{project}"', warn=True, hide=True)
    except Exception:
        log_warn("Project already exists or creation failed, continuing...")


def setup_link_billing(c, project, billing_account):
    """Link billing account to project."""
    log_info("Linking billing account...")
    result = c.run(f'gcloud beta billing projects link "{project}" --billing-account "{billing_account}"',
                   warn=True)
    if result.failed:
        log_error("Failed to link billing account")


def setup_enable_apis(c, project):
    """Enable required Cloud APIs."""
    log_info("Enabling Cloud APIs (this may take a few minutes)...")
    apis = [
        "run.googleapis.com",
        "sql-component.googleapis.com",
        "sqladmin.googleapis.com",
        "compute.googleapis.com",
        "cloudbuild.googleapis.com",
        "secretmanager.googleapis.com",
        "storage.googleapis.com",
    ]
    c.run(f'gcloud services --project "{project}" enable {" ".join(apis)}', pty=True)


def setup_get_service_accounts(c, project):
    """Get service account emails for Cloud Run and Cloud Build."""
    result = c.run(f'gcloud projects describe "{project}" --format "value(projectNumber)"',
                   hide=True)
    project_num = result.stdout.strip()
    cloudrun_sa = f"{project_num}-compute@developer.gserviceaccount.com"
    cloudbuild_sa = f"{project_num}@cloudbuild.gserviceaccount.com"
    log_info(f"Cloud Run SA: {cloudrun_sa}")
    log_info(f"Cloud Build SA: {cloudbuild_sa}")
    return cloudrun_sa, cloudbuild_sa


def setup_iam_permissions(c, project, cloudrun_sa, cloudbuild_sa, sql_project):
    """Set up IAM permissions."""
    log_info("Setting up IAM permissions...")

    # Cloud Build permissions
    c.run(f'gcloud projects add-iam-policy-binding "{project}" '
          f'--member "serviceAccount:{cloudbuild_sa}" '
          f'--role roles/iam.serviceAccountUser --quiet', hide=True)

    c.run(f'gcloud projects add-iam-policy-binding "{project}" '
          f'--member "serviceAccount:{cloudbuild_sa}" '
          f'--role roles/run.admin --quiet', hide=True)

    # Cloud SQL permissions (if using shared instance)
    if sql_project != project:
        log_info(f"Setting up Cloud SQL permissions on {sql_project}...")
        c.run(f'gcloud projects add-iam-policy-binding "{sql_project}" '
              f'--member "serviceAccount:{cloudrun_sa}" '
              f'--role roles/cloudsql.client --quiet', hide=True)

        c.run(f'gcloud projects add-iam-policy-binding "{sql_project}" '
              f'--member "serviceAccount:{cloudbuild_sa}" '
              f'--role roles/cloudsql.client --quiet', hide=True)


def setup_create_database(c, project, sql_instance, sql_project):
    """Create database and user on Cloud SQL."""
    log_info(f"Creating database on {sql_instance}...")

    # Create database
    c.run(f'gcloud sql databases create "{project}" '
          f'--instance "{sql_instance}" '
          f'--project "{sql_project}"', warn=True, hide=True)

    # Create user with random password
    log_info("Creating database user...")
    password = generate_password()
    result = c.run(f'gcloud sql users create "{project}" '
                   f'--instance "{sql_instance}" '
                   f'--project "{sql_project}" '
                   f'--password "{password}"', warn=True, hide=True)

    if result.failed:
        log_warn("User already exists, you may need to reset the password")
        # Generate new password anyway for secrets
        password = generate_password()
        c.run(f'gcloud sql users set-password "{project}" '
              f'--instance "{sql_instance}" '
              f'--project "{sql_project}" '
              f'--password "{password}"', warn=True, hide=True)

    return password


def setup_create_bucket(c, project, bucket_name, region):
    """Create Cloud Storage bucket."""
    log_info(f"Creating storage bucket: {bucket_name}...")
    c.run(f'gsutil mb -l "{region}" -p "{project}" "gs://{bucket_name}"',
          warn=True, hide=True)

    # Set CORS using temp file
    log_info("Setting CORS configuration...")
    import tempfile
    import json
    cors_config = [{"origin": ["*"], "responseHeader": ["Content-Type"], "method": ["GET", "HEAD"], "maxAgeSeconds": 3600}]
    with tempfile.NamedTemporaryFile(mode='w', suffix='.json', delete=False) as f:
        json.dump(cors_config, f)
        cors_file = f.name
    try:
        c.run(f'gsutil cors set "{cors_file}" gs://{bucket_name}', warn=True)
    finally:
        os.unlink(cors_file)


def setup_create_secrets(c, project, secrets_name, bucket_name, db_password,
                         sql_project, region, sql_instance, cloudrun_sa, cloudbuild_sa):
    """Create secrets in Secret Manager."""
    log_info("Creating secrets in Secret Manager...")

    secret_key = generate_password(50)
    database_url = f"postgres://{project}:{db_password}@//cloudsql/{sql_project}:{region}:{sql_instance}/{project}"

    secrets_content = f'''DATABASE_URL="{database_url}"
GS_BUCKET_NAME="{bucket_name}"
SECRET_KEY="{secret_key}"
DEBUG="False"
ALLOWED_HOSTS=".run.app"
CORS_ALLOWED_ORIGINS=""
'''

    # Write to temp file
    import tempfile
    with tempfile.NamedTemporaryFile(mode='w', suffix='.env', delete=False) as f:
        f.write(secrets_content)
        temp_file = f.name

    try:
        # Try to create secret
        result = c.run(f'gcloud secrets create "{secrets_name}" '
                       f'--data-file="{temp_file}" '
                       f'--project "{project}"', warn=True, hide=True)

        if result.failed:
            # Secret exists, add new version
            c.run(f'gcloud secrets versions add "{secrets_name}" '
                  f'--data-file="{temp_file}" '
                  f'--project "{project}"', hide=True)
    finally:
        os.unlink(temp_file)

    # Grant secret access
    log_info("Granting secret access...")
    c.run(f'gcloud secrets add-iam-policy-binding "{secrets_name}" '
          f'--member "serviceAccount:{cloudrun_sa}" '
          f'--role roles/secretmanager.secretAccessor '
          f'--project "{project}" --quiet', hide=True)

    c.run(f'gcloud secrets add-iam-policy-binding "{secrets_name}" '
          f'--member "serviceAccount:{cloudbuild_sa}" '
          f'--role roles/secretmanager.secretAccessor '
          f'--project "{project}" --quiet', hide=True)


@task(name="setup-apis")
def setup_apis(c, project=None):
    """Enable required GCP APIs for an existing project."""
    project = project or GCP_PROJECT_ID
    setup_enable_apis(c, project)


@task(name="setup-iam")
def setup_iam(c, project=None):
    """Set up IAM permissions for an existing project."""
    project = project or GCP_PROJECT_ID
    cloudrun_sa, cloudbuild_sa = setup_get_service_accounts(c, project)
    setup_iam_permissions(c, project, cloudrun_sa, cloudbuild_sa, CLOUD_SQL_PROJECT)


@task(name="setup-bucket")
def setup_bucket(c, project=None, bucket=None):
    """Create Cloud Storage bucket for an existing project."""
    project = project or GCP_PROJECT_ID
    bucket = bucket or project
    setup_create_bucket(c, project, bucket, GCP_REGION)


@task(name="setup-database")
def setup_database(c, project=None):
    """Create database on Cloud SQL for an existing project."""
    project = project or GCP_PROJECT_ID
    password = setup_create_database(c, project, CLOUD_SQL_INSTANCE, CLOUD_SQL_PROJECT)
    print(f"\nDatabase password: {password}")
    print("Save this password - you'll need it for your secrets!")

