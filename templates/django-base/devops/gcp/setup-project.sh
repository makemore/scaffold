#!/usr/bin/env bash
#
# GCP Project Setup Script for Django on Cloud Run
#
# This script creates all the GCP resources needed to deploy a Django app:
# - GCP Project (or uses existing)
# - Cloud Run, Cloud SQL, Secret Manager, Cloud Build APIs
# - Cloud Storage bucket for media files
# - Cloud SQL database (on shared instance)
# - Secret Manager secrets for Django settings
# - IAM permissions for Cloud Run and Cloud Build
#
# Usage:
#   ./setup-project.sh <project-id> [--staging]
#
# Example:
#   ./setup-project.sh myproject
#   ./setup-project.sh myproject-staging --staging
#

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

log_info() { echo -e "${GREEN}[INFO]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

# Check arguments
if [ "$#" -lt 1 ]; then
    echo "Usage: $0 <project-id> [--staging]"
    echo "Example: $0 myproject"
    exit 1
fi

PROJECT_ID=$1
IS_STAGING=false
if [ "$2" == "--staging" ]; then
    IS_STAGING=true
fi

# Configuration - EDIT THESE FOR YOUR ORGANIZATION
ORGANIZATION_ID="${GCP_ORGANIZATION_ID:-}"  # Optional: your GCP org ID
BILLING_ACCOUNT="${GCP_BILLING_ACCOUNT:-}"  # Required: your billing account
REGION="${GCP_REGION:-europe-west2}"
CLOUD_SQL_INSTANCE="${CLOUD_SQL_INSTANCE:-kryten}"
CLOUD_SQL_PROJECT="${CLOUD_SQL_PROJECT:-crimson-305210}"

# Derived values
GS_BUCKET_NAME="${PROJECT_ID}"
if [ "$IS_STAGING" = true ]; then
    SECRETS_NAME="application_settings_staging"
else
    SECRETS_NAME="application_settings"
fi

confirm_continue() {
    read -p "$1 (y/N)? " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
}

# Validate billing account
if [ -z "$BILLING_ACCOUNT" ]; then
    log_error "GCP_BILLING_ACCOUNT environment variable is required"
    echo "Set it with: export GCP_BILLING_ACCOUNT=XXXXXX-XXXXXX-XXXXXX"
    exit 1
fi

log_info "Setting up GCP project: $PROJECT_ID"
log_info "Region: $REGION"
log_info "Staging: $IS_STAGING"

# Create or select project
log_info "Creating/selecting project..."
if [ -n "$ORGANIZATION_ID" ]; then
    gcloud projects create "$PROJECT_ID" --organization "$ORGANIZATION_ID" 2>/dev/null || \
        log_warn "Project already exists or creation failed, continuing..."
else
    gcloud projects create "$PROJECT_ID" 2>/dev/null || \
        log_warn "Project already exists or creation failed, continuing..."
fi

# Link billing
log_info "Linking billing account..."
gcloud beta billing projects link "$PROJECT_ID" --billing-account "$BILLING_ACCOUNT" || \
    log_error "Failed to link billing account"

# Enable APIs
log_info "Enabling Cloud APIs (this may take a few minutes)..."
gcloud services --project "$PROJECT_ID" enable \
    run.googleapis.com \
    sql-component.googleapis.com \
    sqladmin.googleapis.com \
    compute.googleapis.com \
    cloudbuild.googleapis.com \
    secretmanager.googleapis.com \
    storage.googleapis.com

# Get service account emails
PROJECTNUM=$(gcloud projects describe "$PROJECT_ID" --format 'value(projectNumber)')
CLOUDRUN_SA="${PROJECTNUM}-compute@developer.gserviceaccount.com"
CLOUDBUILD_SA="${PROJECTNUM}@cloudbuild.gserviceaccount.com"

log_info "Cloud Run SA: $CLOUDRUN_SA"
log_info "Cloud Build SA: $CLOUDBUILD_SA"

# IAM permissions for Cloud Build
log_info "Setting up IAM permissions..."
gcloud projects add-iam-policy-binding "$PROJECT_ID" \
    --member "serviceAccount:${CLOUDBUILD_SA}" \
    --role roles/iam.serviceAccountUser --quiet

gcloud projects add-iam-policy-binding "$PROJECT_ID" \
    --member "serviceAccount:${CLOUDBUILD_SA}" \
    --role roles/run.admin --quiet

# Cloud SQL permissions (if using shared instance)
if [ "$CLOUD_SQL_PROJECT" != "$PROJECT_ID" ]; then
    log_info "Setting up Cloud SQL permissions on $CLOUD_SQL_PROJECT..."
    gcloud projects add-iam-policy-binding "$CLOUD_SQL_PROJECT" \
        --member "serviceAccount:${CLOUDRUN_SA}" \
        --role roles/cloudsql.client --quiet
    
    gcloud projects add-iam-policy-binding "$CLOUD_SQL_PROJECT" \
        --member "serviceAccount:${CLOUDBUILD_SA}" \
        --role roles/cloudsql.client --quiet
fi

# Create database
log_info "Creating database on $CLOUD_SQL_INSTANCE..."
gcloud sql databases create "$PROJECT_ID" \
    --instance "$CLOUD_SQL_INSTANCE" \
    --project "$CLOUD_SQL_PROJECT" 2>/dev/null || \
    log_warn "Database already exists, continuing..."

# Create database user with random password
log_info "Creating database user..."
PGPASS="$(LC_ALL=C tr -dc 'a-zA-Z0-9' < /dev/urandom | fold -w 30 | head -n 1)"
gcloud sql users create "$PROJECT_ID" \
    --instance "$CLOUD_SQL_INSTANCE" \
    --project "$CLOUD_SQL_PROJECT" \
    --password "$PGPASS" 2>/dev/null || \
    log_warn "User already exists, you may need to reset the password"

# Create storage bucket
log_info "Creating storage bucket: $GS_BUCKET_NAME..."
gsutil mb -l "$REGION" -p "$PROJECT_ID" "gs://${GS_BUCKET_NAME}" 2>/dev/null || \
    log_warn "Bucket already exists, continuing..."

# Set CORS on bucket
log_info "Setting CORS configuration..."
cat > /tmp/cors.json << 'EOF'
[
    {
        "origin": ["*"],
        "responseHeader": ["Content-Type"],
        "method": ["GET", "HEAD"],
        "maxAgeSeconds": 3600
    }
]
EOF
gsutil cors set /tmp/cors.json "gs://$GS_BUCKET_NAME"
rm /tmp/cors.json

# Create secrets
log_info "Creating secrets in Secret Manager..."
SECRET_KEY="$(LC_ALL=C tr -dc 'a-zA-Z0-9' < /dev/urandom | fold -w 50 | head -n 1)"
DATABASE_URL="postgres://${PROJECT_ID}:${PGPASS}@//cloudsql/${CLOUD_SQL_PROJECT}:${REGION}:${CLOUD_SQL_INSTANCE}/${PROJECT_ID}"

cat > /tmp/secrets.env << EOF
DATABASE_URL="${DATABASE_URL}"
GS_BUCKET_NAME="${GS_BUCKET_NAME}"
SECRET_KEY="${SECRET_KEY}"
DEBUG="False"
ALLOWED_HOSTS=".run.app"
CORS_ALLOWED_ORIGINS=""
EOF

gcloud secrets create "$SECRETS_NAME" \
    --data-file /tmp/secrets.env \
    --project "$PROJECT_ID" 2>/dev/null || \
    gcloud secrets versions add "$SECRETS_NAME" \
        --data-file /tmp/secrets.env \
        --project "$PROJECT_ID"

rm /tmp/secrets.env

# Grant secret access
log_info "Granting secret access..."
gcloud secrets add-iam-policy-binding "$SECRETS_NAME" \
    --member "serviceAccount:${CLOUDRUN_SA}" \
    --role roles/secretmanager.secretAccessor \
    --project "$PROJECT_ID" --quiet

gcloud secrets add-iam-policy-binding "$SECRETS_NAME" \
    --member "serviceAccount:${CLOUDBUILD_SA}" \
    --role roles/secretmanager.secretAccessor \
    --project "$PROJECT_ID" --quiet

# Summary
echo ""
log_info "=========================================="
log_info "GCP Project Setup Complete!"
log_info "=========================================="
echo ""
echo "Project ID:     $PROJECT_ID"
echo "Region:         $REGION"
echo "Database:       $PROJECT_ID on $CLOUD_SQL_INSTANCE"
echo "Storage Bucket: gs://$GS_BUCKET_NAME"
echo "Secrets:        $SECRETS_NAME"
echo ""
echo "Next steps:"
echo "  1. Update your .env file with the project settings"
echo "  2. Build and deploy: fab deploy --env=production"
echo "  3. Run migrations: fab migrate --env=production"
echo ""
log_info "Done!"
