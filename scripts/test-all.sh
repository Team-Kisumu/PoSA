#!/usr/bin/env bash
# Run all PoSA tests: unit tests (always) + integration tests (if services running).
# Suitable for CI and local development.
#
# Usage: ./scripts/test-all.sh

set -e

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

ERRORS=0

echo "PoSA Test Suite"
echo "============================================================"

# --- Backend unit tests ---
echo -e "\n${YELLOW}[1/4] Backend unit tests (Go)${NC}"
if [ -d "$ROOT/backend" ]; then
    cd "$ROOT/backend"
    go vet ./...
    go test -race -count=1 ./... || ERRORS=$((ERRORS + 1))
    echo -e "${GREEN}Backend tests complete${NC}"
else
    echo -e "${YELLOW}Skipped: no backend directory${NC}"
fi

# --- AI unit tests ---
echo -e "\n${YELLOW}[2/4] AI unit tests (Python)${NC}"
if [ -d "$ROOT/ai" ]; then
    cd "$ROOT"
    if [ -d "ai/.venv" ]; then
        source ai/.venv/bin/activate
    fi
    PYTHONPATH=. pytest ai/tests/ -q || ERRORS=$((ERRORS + 1))
    echo -e "${GREEN}AI tests complete${NC}"
else
    echo -e "${YELLOW}Skipped: no ai directory${NC}"
fi

# --- Frontend build + lint ---
echo -e "\n${YELLOW}[3/4] Frontend (Node)${NC}"
if [ -d "$ROOT/frontend" ] && [ -f "$ROOT/frontend/package.json" ]; then
    cd "$ROOT/frontend"
    npm run lint || ERRORS=$((ERRORS + 1))
    npm run build || ERRORS=$((ERRORS + 1))
    echo -e "${GREEN}Frontend tests complete${NC}"
else
    echo -e "${YELLOW}Skipped: no frontend directory${NC}"
fi

# --- E2E pipeline tests (only if services are running) ---
echo -e "\n${YELLOW}[4/4] E2E pipeline tests${NC}"
if curl -s http://localhost:8080/health > /dev/null 2>&1; then
    cd "$ROOT"
    if [ -d "ai/.venv" ]; then
        source ai/.venv/bin/activate
    fi
    PYTHONPATH=. python scripts/test_e2e_pipeline.py || ERRORS=$((ERRORS + 1))
    echo -e "${GREEN}E2E tests complete${NC}"
else
    echo -e "${YELLOW}Skipped: backend not running (start with ./scripts/start.sh)${NC}"
fi

# --- Results ---
echo ""
echo "============================================================"
if [ $ERRORS -eq 0 ]; then
    echo -e "${GREEN}All tests passed${NC}"
else
    echo -e "${RED}$ERRORS test suite(s) failed${NC}"
fi
echo "============================================================"
exit $ERRORS
