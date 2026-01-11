#!/bin/bash
# Test script for Next.js template
# 
# This script tests the Next.js template by:
# 1. Generating a project from the template
# 2. Installing dependencies
# 3. Running the build
# 4. Running lint
#
# Usage: ./test_nextjs.sh [--keep] [--no-docker]
#   --keep      Don't cleanup the test directory after running
#   --no-docker Run tests locally instead of in Docker

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
TEMPLATE_DIR="$REPO_ROOT/templates/nextjs-base"

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
PROJECT_NAME="Test Next.js App"
PROJECT_SLUG="test-nextjs-app"
DESCRIPTION="A test project for template validation"

info "Testing Next.js template"
info "Template directory: $TEMPLATE_DIR"

# Create test directory
TEST_DIR=$(create_test_dir "nextjs-test")
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
    "description=$DESCRIPTION"

# Remove scaffold.yaml and devcontainer from generated project
rm -f "$PROJECT_DIR/scaffold.yaml"
rm -rf "$PROJECT_DIR/.devcontainer"

info "Project generated at: $PROJECT_DIR"

# Step 2: Verify key files exist
info "Step 2: Verifying project structure"
run_step "Check package.json exists" test -f "$PROJECT_DIR/package.json"
run_step "Check next.config.ts exists" test -f "$PROJECT_DIR/next.config.ts"
run_step "Check layout.tsx exists" test -f "$PROJECT_DIR/src/app/layout.tsx"
run_step "Check page.tsx exists" test -f "$PROJECT_DIR/src/app/page.tsx"

# Step 3: Verify template variables were substituted
info "Step 3: Verifying template variable substitution"
if grep -q "{{ project_slug }}" "$PROJECT_DIR/package.json"; then
    error "Template variables not substituted in package.json"
    exit 1
fi
info "✓ Template variables substituted correctly"

if [ "$USE_DOCKER" = true ]; then
    # Run tests in Docker container
    info "Step 4: Running tests in Docker container"
    
    docker run --rm -v "$PROJECT_DIR:/app" -w /app node:22-slim bash -c "
        set -e
        echo 'Installing dependencies...'
        npm install --silent
        
        echo 'Running build...'
        npm run build
        
        echo 'Running lint...'
        npm run lint
        
        echo 'All tests passed!'
    "
else
    # Run tests locally
    info "Step 4: Running tests locally"
    
    cd "$PROJECT_DIR"
    
    run_step "Install dependencies" npm install --silent
    run_step "Build project" npm run build
    run_step "Run lint" npm run lint
fi

info "=========================================="
info "Next.js template tests PASSED"
info "=========================================="

if [ "$KEEP_DIR" = true ]; then
    info "Test directory kept at: $PROJECT_DIR"
fi

