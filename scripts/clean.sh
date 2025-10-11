#!/bin/bash
# Ventros CRM - Clean Script
# Stops API + Infrastructure + Removes Data + Cleans Files
# NO CONFIRMATION - runs automatically

set -e  # Exit on error

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
RESET='\033[0m'

echo -e "${RED}⚠️  CLEANING: API + Infrastructure + Data + Files${RESET}"

# 1. Kill API processes
echo -ne "${BLUE}[1/4] Stopping API...${RESET} "
pkill -9 -f "./bin/api" 2>/dev/null || true
pkill -9 -f "go run cmd/api/main.go" 2>/dev/null || true
echo -e "${GREEN}✓${RESET}"

# 2. Stop and remove containers + volumes
echo -ne "${BLUE}[2/4] Removing infrastructure...${RESET} "
docker compose --env-file .deploy/container/.env -f .deploy/container/compose.api.yaml down -v > /dev/null 2>&1 || true
echo -e "${GREEN}✓${RESET}"

# 3. Remove generated files
echo -ne "${BLUE}[3/4] Cleaning files...${RESET} "
rm -f bin/api coverage*.out coverage.html 2>/dev/null || true
rm -rf vendor/ 2>/dev/null || true
echo -e "${GREEN}✓${RESET}"

# 4. Verify
echo -ne "${BLUE}[4/4] Verifying...${RESET} "
API_RUNNING=$(ps aux | grep -E "(./bin/api|go run cmd/api)" | grep -v grep | wc -l)
CONTAINERS=$(docker ps | grep ventros | wc -l)

if [ "$API_RUNNING" -eq 0 ] && [ "$CONTAINERS" -eq 0 ]; then
    echo -e "${GREEN}✓${RESET}"
    echo ""
    echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${RESET}"
    echo -e "${GREEN}✓ CLEAN COMPLETE${RESET}"
    echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${RESET}"
    echo ""
    echo -e "Next steps:"
    echo -e "  ${GREEN}make infra${RESET}  - Start infrastructure"
    echo -e "  ${GREEN}make dev${RESET}    - Start infra + API"
else
    echo -e "${YELLOW}⚠${RESET}"
    echo ""
    echo -e "${YELLOW}Warning: Some processes may still be running${RESET}"
    [ "$API_RUNNING" -gt 0 ] && echo "  • API processes: $API_RUNNING"
    [ "$CONTAINERS" -gt 0 ] && echo "  • Docker containers: $CONTAINERS"
fi
