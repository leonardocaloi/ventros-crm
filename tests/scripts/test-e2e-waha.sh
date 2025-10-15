#!/bin/bash
set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
RESET='\033[0m'

# Usage
if [ "$#" -ne 1 ]; then
    echo -e "${RED}❌ Usage: $0 <waha_session_id>${RESET}"
    echo ""
    echo "Example:"
    echo "  $0 5511999999999"
    echo ""
    exit 1
fi

SESSION_ID="$1"

echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${RESET}"
echo -e "${BLUE}  🧪 WAHA E2E Test - Full Flow${RESET}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${RESET}"
echo ""
echo -e "${GREEN}WAHA Session ID:${RESET} $SESSION_ID"
echo -e "${GREEN}Product:${RESET} CRM (selected automatically by /api/v1/crm/* routes)"
echo ""

# Check if API is running
echo -e "${YELLOW}1️⃣  Checking if API is running...${RESET}"
if ! curl -s http://localhost:8080/health > /dev/null 2>&1; then
    echo -e "${RED}❌ API is not running!${RESET}"
    echo ""
    echo "Please start the API first:"
    echo -e "${BLUE}  make api${RESET}"
    echo ""
    exit 1
fi
echo -e "${GREEN}✅ API is ready${RESET}"
echo ""

# Run E2E tests
echo -e "${YELLOW}2️⃣  Running E2E tests with session: $SESSION_ID${RESET}"
echo ""

# Export session_id as environment variable
export WAHA_SESSION_ID="$SESSION_ID"
export API_BASE_URL="http://localhost:8080"

# Run Go test
go test -v -timeout 10m -run TestWAHAWebhookTestSuite ./tests/e2e/

echo ""
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${RESET}"
echo -e "${GREEN}✅ E2E Test completed!${RESET}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${RESET}"
