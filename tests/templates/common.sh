#!/bin/bash
# Common utilities for template testing

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Print colored status messages
info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Run a command and report success/failure
run_step() {
    local step_name="$1"
    shift
    
    info "Running: $step_name"
    if "$@"; then
        info "✓ $step_name passed"
        return 0
    else
        error "✗ $step_name failed"
        return 1
    fi
}

# Simple template variable substitution
# Usage: render_template "input_file" "output_file" "var1=value1" "var2=value2"
render_template() {
    local input="$1"
    local output="$2"
    shift 2
    
    cp "$input" "$output"
    
    for var in "$@"; do
        local key="${var%%=*}"
        local value="${var#*=}"
        # Use sed to replace {{ key }} with value
        if [[ "$OSTYPE" == "darwin"* ]]; then
            sed -i '' "s/{{ *$key *}}/$value/g" "$output"
        else
            sed -i "s/{{ *$key *}}/$value/g" "$output"
        fi
    done
}

# Render all files in a directory recursively
render_directory() {
    local src_dir="$1"
    local dest_dir="$2"
    shift 2
    local vars=("$@")
    
    # Copy directory structure
    cp -r "$src_dir" "$dest_dir"
    
    # Find all files and render them
    find "$dest_dir" -type f | while read -r file; do
        local temp_file="${file}.tmp"
        cp "$file" "$temp_file"
        
        for var in "${vars[@]}"; do
            local key="${var%%=*}"
            local value="${var#*=}"
            if [[ "$OSTYPE" == "darwin"* ]]; then
                sed -i '' "s/{{ *$key *}}/$value/g" "$temp_file"
            else
                sed -i "s/{{ *$key *}}/$value/g" "$temp_file"
            fi
        done
        
        mv "$temp_file" "$file"
    done
    
    # Rename directories with __project_slug__ pattern
    for var in "${vars[@]}"; do
        local key="${var%%=*}"
        local value="${var#*=}"
        
        # Find and rename directories
        find "$dest_dir" -type d -name "__${key}__" | while read -r dir; do
            local new_dir="${dir/__${key}__/$value}"
            mv "$dir" "$new_dir"
        done
    done
}

# Create a temporary directory for testing
create_test_dir() {
    local prefix="${1:-scaffold-test}"
    mktemp -d "/tmp/${prefix}.XXXXXX"
}

# Cleanup function
cleanup_test_dir() {
    local dir="$1"
    if [[ -d "$dir" && "$dir" == /tmp/* ]]; then
        rm -rf "$dir"
    fi
}

