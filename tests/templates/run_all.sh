#!/bin/bash
# Run all template tests
#
# Usage: ./run_all.sh [--no-docker] [--keep]

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m'

# Parse arguments
ARGS=""
for arg in "$@"; do
    ARGS="$ARGS $arg"
done

echo "=========================================="
echo "Running all template tests"
echo "=========================================="
echo ""

FAILED=0

# Test Django template
echo "Testing Django template..."
if bash "$SCRIPT_DIR/test_django.sh" $ARGS; then
    echo -e "${GREEN}✓ Django template tests passed${NC}"
else
    echo -e "${RED}✗ Django template tests failed${NC}"
    FAILED=1
fi

echo ""

# Test Next.js template
echo "Testing Next.js template..."
if bash "$SCRIPT_DIR/test_nextjs.sh" $ARGS; then
    echo -e "${GREEN}✓ Next.js template tests passed${NC}"
else
    echo -e "${RED}✗ Next.js template tests failed${NC}"
    FAILED=1
fi

echo ""
echo "=========================================="
if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}All template tests PASSED${NC}"
else
    echo -e "${RED}Some template tests FAILED${NC}"
    exit 1
fi
echo "=========================================="

