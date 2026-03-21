#!/usr/bin/env bash
# test-backend.sh: Go backend test runner with coverage
set -e

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
BACKEND="$ROOT/backend"
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m'

VERBOSE=""
COVER_HTML=false

for arg in "$@"; do
    case "$arg" in
        -v|--verbose) VERBOSE="-v" ;;
        --html) COVER_HTML=true ;;
        -h|--help)
            echo "Usage: $0 [OPTIONS]"
            echo "  -v, --verbose   Verbose test output"
            echo "  --html          Generate HTML coverage report"
            echo "  -h, --help      Show this help"
            exit 0 ;;
    esac
done

cd "$BACKEND"

echo -e "${CYAN}▶ Vet${NC}"
go vet ./...
echo -e "${GREEN}✓ No issues${NC}\n"

echo -e "${CYAN}▶ Test${NC}"
go test -race -count=1 -coverprofile=coverage.out $VERBOSE ./...
echo ""

echo -e "${CYAN}▶ Coverage${NC}"
go tool cover -func=coverage.out
echo ""

if $COVER_HTML; then
    go tool cover -html=coverage.out -o coverage.html
    echo -e "${GREEN}✓ HTML report: backend/coverage.html${NC}"
fi

TOTAL=$(go tool cover -func=coverage.out | grep total | awk '{print $3}')
echo -e "${CYAN}Total coverage: ${GREEN}${TOTAL}${NC}"
