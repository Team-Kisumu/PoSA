#!/usr/bin/env bash
# test-watch.sh: Watch for file changes and re-run tests
set -e

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
CYAN='\033[0;36m'
YELLOW='\033[1;33m'
NC='\033[0m'

STACK="${1:-all}"

case "$STACK" in
    backend|go)
        DIR="$ROOT/backend"
        PATTERN="*.go"
        CMD="cd $DIR && go test -race -count=1 ./..."
        ;;
    ai|python)
        DIR="$ROOT/ai"
        PATTERN="*.py"
        CMD="cd $DIR && python3 -m pytest --tb=short -q 2>/dev/null || python3 -m unittest discover -q"
        ;;
    frontend|node)
        DIR="$ROOT/frontend"
        PATTERN="*.{js,jsx,ts,tsx}"
        CMD="cd $DIR && npm test --if-present"
        ;;
    all)
        echo -e "${CYAN}Watching all stacks — use Ctrl+C to stop${NC}"
        echo -e "${YELLOW}Tip: pass 'backend', 'ai', or 'frontend' to watch a single stack${NC}\n"
        DIR="$ROOT"
        PATTERN="*.{go,py,js,jsx,ts,tsx}"
        CMD="$ROOT/scripts/test.sh"
        ;;
    *)
        echo "Usage: $0 [backend|ai|frontend|all]"
        exit 1
        ;;
esac

if ! command -v inotifywait &>/dev/null; then
    echo -e "${YELLOW}inotifywait not found — falling back to polling (2s interval)${NC}\n"
    LAST_HASH=""
    while true; do
        HASH=$(find "$DIR" -name "$PATTERN" -newer "$0" 2>/dev/null | head -20 | md5sum)
        if [ "$HASH" != "$LAST_HASH" ]; then
            LAST_HASH="$HASH"
            echo -e "\n${CYAN}[$(date +%H:%M:%S)] Running tests...${NC}\n"
            eval "$CMD" || true
            echo -e "\n${CYAN}Waiting for changes...${NC}"
        fi
        sleep 2
    done
else
    echo -e "${CYAN}Watching $DIR for $PATTERN changes — Ctrl+C to stop${NC}\n"
    while true; do
        inotifywait -r -e modify,create,delete --include "$PATTERN" "$DIR" 2>/dev/null
        echo -e "\n${CYAN}[$(date +%H:%M:%S)] Running tests...${NC}\n"
        eval "$CMD" || true
        echo -e "\n${CYAN}Waiting for changes...${NC}"
    done
fi
