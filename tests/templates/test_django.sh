#!/bin/bash
# Test script for Django template
# 
# This script tests the Django template by:
# 1. Generating a project from the template
# 2. Installing dependencies
# 3. Running migrations
# 4. Running the test suite
#
# Usage: ./test_django.sh [--keep] [--no-docker]
#   --keep      Don't cleanup the test directory after running
#   --no-docker Run tests locally instead of in Docker

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
TEMPLATE_DIR="$REPO_ROOT/templates/django-base"

source "$SCRIPT_DIR/common.sh"

# Parse arguments
KEEP_DIR=false
USE_DOCKER=true

for arg in "$@"; do
    case $arg in
        --keep)
            KEEP_DIR=true
            ;;
        --no-docker)
            USE_DOCKER=false
            ;;
    esac
done

# Test configuration
PROJECT_NAME="Test Django Project"
PROJECT_SLUG="test_django_project"
DESCRIPTION="A test project for template validation"
GCP_PROJECT="test-project"
GCP_REGION="us-central1"

info "Testing Django template"
info "Template directory: $TEMPLATE_DIR"

# Create test directory
TEST_DIR=$(create_test_dir "django-test")
PROJECT_DIR="$TEST_DIR/$PROJECT_SLUG"
info "Test directory: $TEST_DIR"

# Cleanup on exit (unless --keep is specified)
if [ "$KEEP_DIR" = false ]; then
    trap "cleanup_test_dir '$TEST_DIR'" EXIT
fi

# Step 1: Generate project from template
info "Step 1: Generating project from template"
render_directory "$TEMPLATE_DIR" "$PROJECT_DIR" \
    "project_name=$PROJECT_NAME" \
    "project_slug=$PROJECT_SLUG" \
    "description=$DESCRIPTION" \
    "gcp_project=$GCP_PROJECT" \
    "gcp_region=$GCP_REGION"

# Remove scaffold.yaml and devcontainer from generated project
rm -f "$PROJECT_DIR/scaffold.yaml"
rm -rf "$PROJECT_DIR/.devcontainer"

info "Project generated at: $PROJECT_DIR"

# Step 2: Verify key files exist
info "Step 2: Verifying project structure"
run_step "Check manage.py exists" test -f "$PROJECT_DIR/manage.py"
run_step "Check settings exist" test -f "$PROJECT_DIR/$PROJECT_SLUG/settings/base.py"
run_step "Check requirements.txt exists" test -f "$PROJECT_DIR/requirements.txt"

# Step 3: Verify template variables were substituted
info "Step 3: Verifying template variable substitution"
if grep -q "{{ project_slug }}" "$PROJECT_DIR/manage.py"; then
    error "Template variables not substituted in manage.py"
    exit 1
fi
info "✓ Template variables substituted correctly"

if [ "$USE_DOCKER" = true ]; then
    # Run tests in Docker container
    info "Step 4: Running tests in Docker container"
    
    docker run --rm -v "$PROJECT_DIR:/app" -w /app -e USE_SQLITE=true python:3.12-slim bash -c "
        set -e
        echo 'Installing dependencies...'
        pip install -q -r requirements.txt

        echo 'Running migrations...'
        python manage.py migrate --run-syncdb

        echo 'Running Django checks...'
        python manage.py check

        echo 'Running tests...'
        python manage.py test --verbosity=2

        echo 'All tests passed!'
    "
else
    # Run tests locally
    info "Step 4: Running tests locally"
    
    cd "$PROJECT_DIR"

    # Find Python 3.12+ (required for django-cloud-tasks)
    PYTHON_CMD=""
    for cmd in python3.12 python3.13 python3; do
        if command -v "$cmd" &> /dev/null; then
            version=$("$cmd" -c "import sys; print(f'{sys.version_info.major}.{sys.version_info.minor}')")
            major=$(echo "$version" | cut -d. -f1)
            minor=$(echo "$version" | cut -d. -f2)
            if [ "$major" -ge 3 ] && [ "$minor" -ge 12 ]; then
                PYTHON_CMD="$cmd"
                break
            fi
        fi
    done

    if [ -z "$PYTHON_CMD" ]; then
        error "Python 3.12+ is required but not found"
        exit 1
    fi

    info "Using Python: $PYTHON_CMD ($($PYTHON_CMD --version))"

    # Create virtual environment
    $PYTHON_CMD -m venv .venv
    source .venv/bin/activate

    run_step "Install dependencies" pip install -q -r requirements.txt

    # Use SQLite for testing (no PostgreSQL required)
    export USE_SQLITE=true

    run_step "Run migrations" python manage.py migrate --run-syncdb
    run_step "Django system check" python manage.py check
    run_step "Run tests" python manage.py test --verbosity=2

    deactivate
fi

info "=========================================="
info "Django template tests PASSED"
info "=========================================="

if [ "$KEEP_DIR" = true ]; then
    info "Test directory kept at: $PROJECT_DIR"
fi

