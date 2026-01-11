"""
Django settings for Google Cloud Run production deployment.

This settings file:
- Reads secrets from Google Secret Manager
- Uses WhiteNoise for static files
- Uses Google Cloud Storage for media files
- Configures CORS for API access

Required environment variables:
- DJANGO_SETTINGS_MODULE={{ project_slug }}.settings.cloud_production
- GCP_PROJECT_ID (optional, auto-detected on Cloud Run)

Required secrets in Secret Manager (application_settings):
- DATABASE_URL
- SECRET_KEY
- GS_BUCKET_NAME
- ALLOWED_HOSTS (comma-separated)
- CORS_ALLOWED_ORIGINS (comma-separated, optional)
"""

import io
import os
from google.cloud import secretmanager

from .base import *

# =============================================================================
# Load secrets from Google Secret Manager
# =============================================================================

def get_secret(secret_id: str, project_id: str = None) -> str:
    """Fetch a secret from Google Secret Manager."""
    if project_id is None:
        project_id = os.environ.get("GCP_PROJECT_ID") or os.environ.get("GOOGLE_CLOUD_PROJECT")
    
    client = secretmanager.SecretManagerServiceClient()
    name = f"projects/{project_id}/secrets/{secret_id}/versions/latest"
    response = client.access_secret_version(request={"name": name})
    return response.payload.data.decode("UTF-8")


# Load application settings from Secret Manager
try:
    import environ
    env = environ.Env()
    
    secret_payload = get_secret("application_settings")
    env.read_env(io.StringIO(secret_payload))
except Exception as e:
    import logging
    logging.warning(f"Could not load secrets from Secret Manager: {e}")
    env = environ.Env()

# =============================================================================
# Core Settings
# =============================================================================

DEBUG = env.bool("DEBUG", default=False)
SECRET_KEY = env("SECRET_KEY")

# Parse ALLOWED_HOSTS from comma-separated string
ALLOWED_HOSTS = [h.strip() for h in env("ALLOWED_HOSTS", default="").split(",") if h.strip()]

# =============================================================================
# Database - Cloud SQL via Unix socket
# =============================================================================

DATABASES = {
    "default": env.db("DATABASE_URL")
}

# =============================================================================
# Static Files - WhiteNoise
# =============================================================================

MIDDLEWARE.insert(1, "whitenoise.middleware.WhiteNoiseMiddleware")

STATIC_URL = "/static/"
STATIC_ROOT = BASE_DIR / "staticfiles"

STORAGES = {
    "default": {
        "BACKEND": "storages.backends.gcloud.GoogleCloudStorage",
        "OPTIONS": {
            "bucket_name": env("GS_BUCKET_NAME"),
        },
    },
    "staticfiles": {
        "BACKEND": "whitenoise.storage.CompressedManifestStaticFilesStorage",
    },
}

# =============================================================================
# Media Files - Google Cloud Storage
# =============================================================================

GS_BUCKET_NAME = env("GS_BUCKET_NAME")
GS_DEFAULT_ACL = "publicRead"
GS_QUERYSTRING_AUTH = False

# =============================================================================
# CORS Configuration
# =============================================================================

CORS_ALLOWED_ORIGINS = [
    o.strip() for o in env("CORS_ALLOWED_ORIGINS", default="").split(",") if o.strip()
]
CORS_ALLOW_CREDENTIALS = True

if CORS_ALLOWED_ORIGINS:
    INSTALLED_APPS = ["corsheaders"] + list(INSTALLED_APPS)
    MIDDLEWARE.insert(0, "corsheaders.middleware.CorsMiddleware")

# =============================================================================
# Security Settings
# =============================================================================

SECURE_SSL_REDIRECT = True
SECURE_PROXY_SSL_HEADER = ("HTTP_X_FORWARDED_PROTO", "https")
SESSION_COOKIE_SECURE = True
CSRF_COOKIE_SECURE = True

# =============================================================================
# Logging
# =============================================================================

LOGGING = {
    "version": 1,
    "disable_existing_loggers": False,
    "formatters": {
        "json": {
            "format": '{"time": "%(asctime)s", "level": "%(levelname)s", "name": "%(name)s", "message": "%(message)s"}',
        },
    },
    "handlers": {
        "console": {
            "class": "logging.StreamHandler",
            "formatter": "json",
        },
    },
    "root": {
        "handlers": ["console"],
        "level": env("LOG_LEVEL", default="INFO"),
    },
    "loggers": {
        "django": {
            "handlers": ["console"],
            "level": env("LOG_LEVEL", default="INFO"),
            "propagate": False,
        },
    },
}

# =============================================================================
# Cloud Tasks Configuration (Production)
# =============================================================================
# Disable eager mode - use real Cloud Tasks infrastructure

DJANGO_CLOUD_TASKS['eager'] = False
DJANGO_CLOUD_TASKS['project_id'] = os.environ.get("GCP_PROJECT_ID") or os.environ.get("GOOGLE_CLOUD_PROJECT")
DJANGO_CLOUD_TASKS['location'] = env("GCP_LOCATION", default="europe-west2")
