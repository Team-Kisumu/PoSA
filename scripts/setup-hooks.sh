#!/usr/bin/env bash
# setup-hooks.sh: Install project git hooks
set -e

HOOK_DIR="$(git rev-parse --show-toplevel)/scripts/hooks"
GIT_HOOK_DIR="$(git rev-parse --git-dir)/hooks"

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

if [ ! -d "$HOOK_DIR" ]; then
    echo "Hook source directory not found: $HOOK_DIR"
    exit 1
fi

echo -e "${YELLOW}Installing git hooks...${NC}"

for hook in "$HOOK_DIR"/*; do
    name=$(basename "$hook")
    chmod +x "$hook"
    ln -sf "$hook" "$GIT_HOOK_DIR/$name"
    echo -e "  ${GREEN}✓${NC} $name"
done

echo -e "${GREEN}All hooks installed.${NC}"
