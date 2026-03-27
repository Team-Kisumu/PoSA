#!/usr/bin/env bash
# End-to-end test for the PoSA storage layer.
#
# Tests:
# 1. IPFS client with a mock server (upload + retrieve roundtrip)
# 2. Filecoin client config validation
# 3. Factory backend selection
#
# Run from project root:
#   ./scripts/test_e2e_storage.sh

set -e

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo "PoSA Storage Layer - End-to-End Test"
echo "============================================================"

cd backend

echo -e "\n${YELLOW}[1] Running storage unit tests...${NC}"
go test -race -count=1 -v ./storage/ 2>&1
echo -e "${GREEN}✓ Storage unit tests passed${NC}"

echo -e "\n${YELLOW}[2] Running full backend test suite...${NC}"
go test -race -count=1 ./... 2>&1
echo -e "${GREEN}✓ Full backend suite passed${NC}"

echo -e "\n${YELLOW}[3] Verifying go-synapse dependency...${NC}"
if grep -q "go-synapse" go.mod; then
    echo -e "${GREEN}✓ go-synapse found in go.mod${NC}"
else
    echo -e "${RED}✗ go-synapse not found in go.mod${NC}"
    exit 1
fi

echo -e "\n${YELLOW}[4] Verifying build with storage package...${NC}"
go build ./... 2>&1
echo -e "${GREEN}✓ Build successful${NC}"

echo -e "\n============================================================"
echo -e "${GREEN}All storage e2e tests passed${NC}"
echo "============================================================"
