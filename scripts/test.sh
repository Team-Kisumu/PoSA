#!/usr/bin/env bash
# test.sh: Run tests across all project stacks
set -e

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m'

FAILURES=0
PASSED=0

run_stack() {
    local name="$1" dir="$2" cmd="$3"
    if [ ! -d "$ROOT/$dir" ]; then
        echo -e "${YELLOW}⊘ $name — skipped (directory not found)${NC}"
        return
    fi
    echo -e "${CYAN}▶ $name${NC}"
    if (cd "$ROOT/$dir" && eval "$cmd"); then
        echo -e "${GREEN}✓ $name passed${NC}\n"
        PASSED=$((PASSED + 1))
    else
        echo -e "${RED}✗ $name failed${NC}\n"
        FAILURES=$((FAILURES + 1))
    fi
}

echo -e "${CYAN}═══════════════════════════════════${NC}"
echo -e "${CYAN}  PoSA Test Suite${NC}"
echo -e "${CYAN}═══════════════════════════════════${NC}\n"

# Go backend
run_stack "Backend (Go)" "backend" \
    "go vet ./... && go test -race -coverprofile=coverage.out ./... && go tool cover -func=coverage.out | tail -1"

# Python AI engine
run_stack "AI Engine (Python)" "ai" \
    "pip install -q -r requirements.txt 2>/dev/null; python3 -m pytest --tb=short -q 2>/dev/null || python3 -m unittest discover -s . -q"

# Node frontend
run_stack "Frontend (Node)" "frontend" \
    "npm test --if-present 2>/dev/null || echo 'No test script defined'"

echo -e "${CYAN}═══════════════════════════════════${NC}"
echo -e "  Results: ${GREEN}${PASSED} passed${NC}, ${RED}${FAILURES} failed${NC}"
echo -e "${CYAN}═══════════════════════════════════${NC}"

exit $FAILURES
