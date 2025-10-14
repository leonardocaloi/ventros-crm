#!/bin/bash

# Complete ACK test with webhook subscriptions and event tracking
# Tests: ACK updates + Domain events + Webhook notifications

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
CYAN='\033[0;36m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
RESET='\033[0m'

API_BASE="http://localhost:8080/api/v1"
WEBHOOK_SITE="https://webhook.site/c98be537-02d2-48b7-aa38-15b680caffa3"
TIMESTAMP=$(date +%s)

# Source .env
if [ -f .env ]; then
    source .env
fi

WAHA_SESSION_ID="${WAHA_DEFAULT_SESSION_ID_TEST:-test-session}"
TEST_PHONE="${TEST_PHONE_NUMBER:-5544970444747}"

# Database config
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_NAME="${DB_NAME:-ventros_crm}"
DB_USER="${DB_USER:-postgres}"
DB_PASSWORD="${DB_PASSWORD:-postgres}"

echo ""
echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${RESET}"
echo -e "${CYAN}  Complete ACK + Webhook Test${RESET}"
echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${RESET}"
echo ""
echo "This test will:"
echo "  1. Create test environment (user, project, channel, contact)"
echo "  2. Subscribe webhooks for ALL events"
echo "  3. Create mock message in database"
echo "  4. Send ACK webhooks (2 → 3 → 4)"
echo "  5. Track all events sent to webhook.site"
echo ""
echo -e "${BLUE}Webhook URL: $WEBHOOK_SITE${RESET}"
echo ""

# ===== SETUP =====
echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${RESET}"
echo -e "${CYAN}  Step 1: Setup Test Environment${RESET}"
echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${RESET}"
echo ""

# Register user
echo -e "${YELLOW}→ Registering user...${RESET}"
REGISTER_RESPONSE=$(curl -s -X POST "$API_BASE/auth/register" \
  -H "Content-Type: application/json" \
  -d "{
    \"name\": \"ACK Webhook Test $TIMESTAMP\",
    \"email\": \"ack-webhook-$TIMESTAMP@test.com\",
    \"password\": \"Test@123456\"
  }")

USER_ID=$(echo $REGISTER_RESPONSE | jq -r '.user_id')
PROJECT_ID=$(echo $REGISTER_RESPONSE | jq -r '.default_project_id')
TOKEN=$(echo $REGISTER_RESPONSE | jq -r '.api_key')

if [ "$USER_ID" == "null" ]; then
    echo -e "${RED}✗ Failed to register user${RESET}"
    exit 1
fi

echo -e "${GREEN}✓ User registered${RESET}"
echo "  User ID: $USER_ID"
echo "  Project ID: $PROJECT_ID"

# Create WAHA channel
echo ""
echo -e "${YELLOW}→ Creating WAHA channel...${RESET}"
CHANNEL_RESPONSE=$(curl -s -X POST "$API_BASE/crm/channels?project_id=$PROJECT_ID" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d "{
    \"name\": \"ACK Webhook Channel\",
    \"type\": \"waha\",
    \"waha_config\": {
      \"base_url\": \"https://waha.ventros.cloud\",
      \"api_key\": \"mock-key\",
      \"session_id\": \"$WAHA_SESSION_ID\"
    }
  }")

CHANNEL_ID=$(echo $CHANNEL_RESPONSE | jq -r '.id')

if [ "$CHANNEL_ID" == "null" ]; then
    echo -e "${RED}✗ Failed to create channel${RESET}"
    exit 1
fi

# Get webhook_id from database
WEBHOOK_ID=$(PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -t -A -c \
    "SELECT webhook_id FROM channels WHERE id = '$CHANNEL_ID';")

echo -e "${GREEN}✓ Channel created${RESET}"
echo "  Channel ID: $CHANNEL_ID"
echo "  Webhook ID: $WEBHOOK_ID"

# Create contact
echo ""
echo -e "${YELLOW}→ Creating contact...${RESET}"
CONTACT_RESPONSE=$(curl -s -X POST "$API_BASE/contacts?project_id=$PROJECT_ID" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d "{
    \"name\": \"ACK Webhook Contact\",
    \"phone\": \"$TEST_PHONE\"
  }")

CONTACT_ID=$(echo $CONTACT_RESPONSE | jq -r '.id')

if [ "$CONTACT_ID" == "null" ]; then
    echo -e "${RED}✗ Failed to create contact${RESET}"
    exit 1
fi

echo -e "${GREEN}✓ Contact created${RESET}"
echo "  Contact ID: $CONTACT_ID"

# ===== SUBSCRIBE WEBHOOKS =====
echo ""
echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${RESET}"
echo -e "${CYAN}  Step 2: Subscribe Webhooks for Events${RESET}"
echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${RESET}"
echo ""

# List of events to subscribe
EVENTS=(
    "message.created"
    "message.delivered"
    "message.read"
    "message.played"
    "message.failed"
    "message.status_changed"
)

echo -e "${YELLOW}→ Subscribing to events...${RESET}"
for EVENT in "${EVENTS[@]}"; do
    echo "  Subscribing to: $EVENT"

    SUBSCRIBE_RESPONSE=$(curl -s -X POST "$API_BASE/webhooks/subscriptions?project_id=$PROJECT_ID" \
      -H "Content-Type: application/json" \
      -H "Authorization: Bearer $TOKEN" \
      -d "{
        \"url\": \"$WEBHOOK_SITE\",
        \"events\": [\"$EVENT\"],
        \"description\": \"Test subscription for $EVENT\",
        \"active\": true
      }")

    SUBSCRIPTION_ID=$(echo $SUBSCRIBE_RESPONSE | jq -r '.id // .subscription_id')

    if [ "$SUBSCRIPTION_ID" != "null" ] && [ -n "$SUBSCRIPTION_ID" ]; then
        echo -e "    ${GREEN}✓ Subscribed${RESET} (ID: $SUBSCRIPTION_ID)"
    else
        echo -e "    ${YELLOW}⚠ Subscription may have failed${RESET} (response: $(echo $SUBSCRIBE_RESPONSE | jq -c '.'))"
    fi
done

echo ""
echo -e "${GREEN}✓ Webhook subscriptions configured${RESET}"
echo ""
echo -e "${BLUE}📡 All events will be sent to: $WEBHOOK_SITE${RESET}"

# ===== CREATE MOCK MESSAGE =====
echo ""
echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${RESET}"
echo -e "${CYAN}  Step 3: Create Mock Message${RESET}"
echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${RESET}"
echo ""

# Generate UUIDs and WAMID
MESSAGE_ID=$(uuidgen | tr '[:upper:]' '[:lower:]')
CHANNEL_MESSAGE_ID="wamid.test$(date +%s)$(( RANDOM % 1000 ))"

echo "Creating mock video message:"
echo "  Message ID: $MESSAGE_ID"
echo "  Channel Message ID (WAMID): $CHANNEL_MESSAGE_ID"
echo ""

# Insert mock message
PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "
INSERT INTO messages (
    id,
    tenant_id,
    user_id,
    channel_id,
    contact_id,
    project_id,
    from_me,
    content_type,
    text,
    status,
    channel_message_id,
    timestamp,
    created_at,
    updated_at
) VALUES (
    '$MESSAGE_ID'::uuid,
    '$PROJECT_ID',
    '$USER_ID'::uuid,
    '$CHANNEL_ID'::uuid,
    '$CONTACT_ID'::uuid,
    '$PROJECT_ID'::uuid,
    true,
    'video',
    '🎥 Testing ACK webhooks - watch webhook.site!',
    'sent',
    '$CHANNEL_MESSAGE_ID',
    NOW(),
    NOW(),
    NOW()
);
" > /dev/null

echo -e "${GREEN}✓ Mock message created${RESET}"

# ===== SEND ACKs =====
echo ""
echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${RESET}"
echo -e "${CYAN}  Step 4: Send ACK Webhooks${RESET}"
echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${RESET}"
echo ""

# Function to send ACK
send_ack() {
    local ACK_NUM=$1
    local ACK_NAME=$2
    local EXPECTED_STATUS=$3
    local EVENT_EXPECTED=$4

    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${RESET}"
    echo -e "${BLUE}  ACK $ACK_NUM ($ACK_NAME)${RESET}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${RESET}"
    echo ""

    # Build payload
    ACK_PAYLOAD=$(cat <<EOF
{
  "event": "message.ack",
  "session": "$WAHA_SESSION_ID",
  "payload": {
    "id": "$CHANNEL_MESSAGE_ID",
    "ack": $ACK_NUM,
    "ackName": "$ACK_NAME",
    "from": "${TEST_PHONE}@c.us"
  }
}
EOF
)

    echo -e "${YELLOW}→ Sending ACK webhook...${RESET}"

    # Send webhook
    WEBHOOK_RESPONSE=$(curl -s -X POST "$API_BASE/webhooks/$WEBHOOK_ID" \
      -H "Content-Type: application/json" \
      -d "$ACK_PAYLOAD")

    WEBHOOK_STATUS=$(echo $WEBHOOK_RESPONSE | jq -r '.status')

    if [ "$WEBHOOK_STATUS" != "queued" ]; then
        echo -e "${RED}✗ Webhook failed${RESET}"
        echo "Response: $WEBHOOK_RESPONSE"
        return 1
    fi

    echo -e "${GREEN}✓ ACK webhook sent and queued${RESET}"

    # Wait for processing
    echo -e "${YELLOW}→ Waiting 3s for processing...${RESET}"
    sleep 3

    # Check database
    NEW_STATUS=$(PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -t -A -c \
        "SELECT status FROM messages WHERE id = '$MESSAGE_ID';")

    echo ""
    echo "Database Status: $NEW_STATUS"

    if [ "$NEW_STATUS" == "$EXPECTED_STATUS" ]; then
        echo -e "${GREEN}✓ Status updated to '$EXPECTED_STATUS'${RESET}"
    else
        echo -e "${RED}✗ Expected '$EXPECTED_STATUS', got '$NEW_STATUS'${RESET}"
        return 1
    fi

    echo ""
    echo -e "${BLUE}📡 Expected domain event: $EVENT_EXPECTED${RESET}"
    echo -e "${BLUE}   Check webhook.site for the event!${RESET}"
    echo ""
}

# Send ACKs progressively
send_ack 2 "DEVICE" "delivered" "message.delivered"
send_ack 3 "READ" "read" "message.read"
send_ack 4 "PLAYED" "played" "message.played"

# ===== FINAL STATUS =====
echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${RESET}"
echo -e "${CYAN}  Final Status${RESET}"
echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${RESET}"
echo ""

PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c \
    "SELECT
        id,
        status,
        delivered_at,
        read_at,
        played_at
    FROM messages
    WHERE id = '$MESSAGE_ID';"

echo ""
echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${RESET}"
echo -e "${GREEN}  ✓ Test Complete!${RESET}"
echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${RESET}"
echo ""
echo "Summary:"
echo "  - Message ID: $MESSAGE_ID"
echo "  - WAMID: $CHANNEL_MESSAGE_ID"
echo "  - ACKs sent: DEVICE (2) → READ (3) → PLAYED (4)"
echo "  - Status progression: sent → delivered → read → played"
echo ""
echo -e "${BLUE}📊 Check Webhook Results:${RESET}"
echo -e "${BLUE}   $WEBHOOK_SITE${RESET}"
echo ""
echo "You should see 3 webhook calls:"
echo "  1. message.delivered event"
echo "  2. message.read event"
echo "  3. message.played event"
echo ""
echo "Each webhook should return HTTP 200 from webhook.site"
echo ""
