# Django Starter - Dockerfile for Google Cloud Run
#
# Build:
#   docker build -t django_starter .
#
# Run locally:
#   docker run -p 8080:8080 -e DJANGO_SETTINGS_MODULE=django_starter.settings.dev django_starter

FROM python:3.11-slim

# Set environment variables
ENV PYTHONDONTWRITEBYTECODE=1
ENV PYTHONUNBUFFERED=1
ENV PORT=8080

# Set work directory
WORKDIR /app

# Install system dependencies
RUN apt-get update && apt-get install -y --no-install-recommends \
    build-essential \
    libpq-dev \
    && rm -rf /var/lib/apt/lists/*

# Install Python dependencies
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Copy project files
COPY . .

# Collect static files (for WhiteNoise)
RUN python manage.py collectstatic --noinput --settings=django_starter.settings.base || true

# Create non-root user for security
RUN useradd -m -u 1000 appuser && chown -R appuser:appuser /app
USER appuser

# Expose port
EXPOSE 8080

# Run gunicorn
CMD exec gunicorn --bind :$PORT --workers 2 --threads 4 --timeout 60 django_starter.wsgi:application
