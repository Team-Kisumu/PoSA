#!/usr/bin/env bash
# Start all PoSA services: backend, AI engine, and frontend.
# Run from project root: ./scripts/start.sh
#
# Stops all services on Ctrl+C.

set -e

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

# Track PIDs for cleanup.
PIDS=()

cleanup() {
    echo -e "\n${YELLOW}Stopping all services...${NC}"
    for pid in "${PIDS[@]}"; do
        kill "$pid" 2>/dev/null || true
    done
    wait 2>/dev/null
    echo -e "${GREEN}All services stopped${NC}"
}
trap cleanup EXIT INT TERM

# Load .env if present (skip lines with <placeholder> values or comments).
if [ -f "$ROOT/.env" ]; then
    while IFS= read -r line; do
        # Skip comments, empty lines, and lines with <placeholder> values.
        [[ -z "$line" || "$line" == \#* || "$line" == *"<"* ]] && continue
        export "$line"
    done < "$ROOT/.env"
    echo -e "${GREEN}Loaded .env${NC}"
fi

# --- Backend (Go) ---
echo -e "${YELLOW}Starting backend on :8080...${NC}"
cd "$ROOT/backend"
go run main.go &
PIDS+=($!)
sleep 1

# --- AI Engine (Python) ---
echo -e "${YELLOW}Starting AI engine on :8000...${NC}"
cd "$ROOT/ai"
if [ -d ".venv" ]; then
    source .venv/bin/activate
fi
PYTHONPATH="$ROOT" python evaluator.py &
PIDS+=($!)
sleep 1

# --- Frontend (Next.js) ---
echo -e "${YELLOW}Starting frontend on :3000...${NC}"
cd "$ROOT/frontend"
if [ ! -d "node_modules" ]; then
    echo -e "${YELLOW}Installing frontend dependencies...${NC}"
    npm install --silent
fi
npm run dev &
PIDS+=($!)

echo ""
echo -e "${GREEN}All services running:${NC}"
echo -e "  Backend:  http://localhost:8080"
echo -e "  AI Engine: http://localhost:8000"
echo -e "  Frontend: http://localhost:3000"
echo ""
echo -e "${YELLOW}Press Ctrl+C to stop all services${NC}"

# Wait for any child to exit.
wait
