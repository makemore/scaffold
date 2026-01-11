# {{ project_name }}

{{ description }}

## Features

- Django REST Framework
- Django Allauth (email-only auth)
- dj-rest-auth (API endpoints)
- django-authtools (custom User model)
- S3 storage (optional)
- django-cloud-tasks (GCP Cloud Tasks queue)
- GCP Cloud Run deployment ready

## Quick Start

```bash
# Create virtual environment
python -m venv .venv
source .venv/bin/activate

# Install dependencies
pip install -r requirements.txt

# Run migrations
python manage.py migrate

# Start development server
python manage.py runserver
```

## Deployment

See `fabfile.py` for GCP Cloud Run deployment commands:

```bash
fab setup --project={{ project_slug }} --billing=YOUR-BILLING-ID
fab deploy
```

