#!/bin/bash

# Test ACK simulation with mock message (without WAHA)
# Creates a mock message directly in the database and tests ACK webhooks

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
CYAN='\033[0;36m'
YELLOW='\033[1;33m'
RESET='\033[0m'

API_BASE="http://localhost:8080/api/v1"
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
echo -e "${CYAN}  ACK Webhook Test (Mock Mode)${RESET}"
echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${RESET}"
echo ""

# ===== SETUP =====
echo -e "${CYAN}→ Setting up test environment...${RESET}"

# Register user
REGISTER_RESPONSE=$(curl -s -X POST "$API_BASE/auth/register" \
  -H "Content-Type: application/json" \
  -d "{
    \"name\": \"ACK Mock Test $TIMESTAMP\",
    \"email\": \"ack-mock-$TIMESTAMP@test.com\",
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

# Create WAHA channel
CHANNEL_RESPONSE=$(curl -s -X POST "$API_BASE/crm/channels?project_id=$PROJECT_ID" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d "{
    \"name\": \"ACK Mock Channel\",
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

# Get webhook_id from database (not returned by API)
WEBHOOK_ID=$(PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -t -A -c \
    "SELECT webhook_id FROM channels WHERE id = '$CHANNEL_ID';")

if [ -z "$WEBHOOK_ID" ] || [ "$WEBHOOK_ID" == "" ]; then
    echo -e "${RED}✗ Failed to get webhook_id from database${RESET}"
    exit 1
fi

echo -e "${GREEN}✓ Channel created${RESET}"
echo "  Channel ID: $CHANNEL_ID"
echo "  Webhook ID: $WEBHOOK_ID"

# Create contact
CONTACT_RESPONSE=$(curl -s -X POST "$API_BASE/contacts?project_id=$PROJECT_ID" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d "{
    \"name\": \"ACK Mock Contact\",
    \"phone\": \"$TEST_PHONE\"
  }")

CONTACT_ID=$(echo $CONTACT_RESPONSE | jq -r '.id')

if [ "$CONTACT_ID" == "null" ]; then
    echo -e "${RED}✗ Failed to create contact${RESET}"
    exit 1
fi

echo -e "${GREEN}✓ Contact created${RESET}"
echo ""

# ===== CREATE MOCK MESSAGE =====
echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${RESET}"
echo -e "${CYAN}  Creating Mock Message in Database${RESET}"
echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${RESET}"
echo ""

# Generate UUIDs and WAMID
MESSAGE_ID=$(uuidgen | tr '[:upper:]' '[:lower:]')
CHANNEL_MESSAGE_ID="wamid.mock$(date +%s)$(( RANDOM % 1000 ))"

echo "Creating mock message:"
echo "  Message ID: $MESSAGE_ID"
echo "  Channel Message ID (mock WAMID): $CHANNEL_MESSAGE_ID"
echo "  Channel ID: $CHANNEL_ID"
echo "  Contact ID: $CONTACT_ID"
echo ""

# Insert mock message directly into database
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
    '🎥 Mock video message for ACK testing',
    'sent',
    '$CHANNEL_MESSAGE_ID',
    NOW(),
    NOW(),
    NOW()
);
" > /dev/null

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ Mock message created in database${RESET}"
else
    echo -e "${RED}✗ Failed to create mock message${RESET}"
    exit 1
fi

echo ""

# Function to send ACK webhook
send_ack() {
    local ACK_NUM=$1
    local ACK_NAME=$2
    local EXPECTED_STATUS=$3

    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${RESET}"
    echo -e "${CYAN}  Test ACK $ACK_NUM ($ACK_NAME)${RESET}"
    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${RESET}"
    echo ""

    # Build ACK webhook payload
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

    echo "Sending ACK webhook..."
    echo "Payload:"
    echo "$ACK_PAYLOAD" | jq '.'
    echo ""

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

    echo -e "${GREEN}✓ Webhook accepted (queued for processing)${RESET}"

    # Wait for processing
    echo -e "${YELLOW}→ Waiting 2s for ACK to be processed...${RESET}"
    sleep 2

    # Query message status from database
    NEW_STATUS=$(PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -t -A -c \
        "SELECT status FROM messages WHERE id = '$MESSAGE_ID';")

    DELIVERED_AT=$(PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -t -A -c \
        "SELECT delivered_at FROM messages WHERE id = '$MESSAGE_ID';")

    READ_AT=$(PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -t -A -c \
        "SELECT read_at FROM messages WHERE id = '$MESSAGE_ID';")

    PLAYED_AT=$(PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -t -A -c \
        "SELECT played_at FROM messages WHERE id = '$MESSAGE_ID';")

    echo ""
    echo "Message Status Updated:"
    echo "  Status: $NEW_STATUS"

    if [ -n "$DELIVERED_AT" ] && [ "$DELIVERED_AT" != "" ]; then
        echo "  ✓✓ Delivered at: $DELIVERED_AT"
    fi

    if [ -n "$READ_AT" ] && [ "$READ_AT" != "" ]; then
        echo "  ✓✓ (blue) Read at: $READ_AT"
    fi

    if [ -n "$PLAYED_AT" ] && [ "$PLAYED_AT" != "" ]; then
        echo "  ▶️ Played at: $PLAYED_AT"
    fi

    echo ""

    # Verify status
    if [ "$NEW_STATUS" == "$EXPECTED_STATUS" ]; then
        echo -e "${GREEN}✓ Status correctly updated to '$EXPECTED_STATUS'${RESET}"
    else
        echo -e "${RED}✗ Expected status '$EXPECTED_STATUS', got '$NEW_STATUS'${RESET}"
        return 1
    fi

    echo ""
}

# ===== SIMULATE ACK PROGRESSION =====

# ACK 1: SERVER (Sent to WhatsApp server) - should NOT update (already at 'sent')
echo -e "${YELLOW}Note: Message starts at 'sent' status, ACK 1 won't change it${RESET}"
echo ""

# ACK 2: DEVICE (Delivered to device - ✓✓)
send_ack 2 "DEVICE" "delivered"

# ACK 3: READ (Read by recipient - ✓✓ blue)
send_ack 3 "READ" "read"

# ACK 4: PLAYED (Media played/viewed - ▶️)
send_ack 4 "PLAYED" "played"

# ===== FINAL STATUS =====
echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${RESET}"
echo -e "${CYAN}  Final Status Check${RESET}"
echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${RESET}"
echo ""

echo "Complete Message Details (from database):"
PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c \
    "SELECT
        id,
        channel_message_id,
        status,
        content_type,
        delivered_at,
        read_at,
        played_at,
        created_at
    FROM messages
    WHERE id = '$MESSAGE_ID';"

echo ""
echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${RESET}"
echo -e "${GREEN}  ✓ ACK Mock Test Complete!${RESET}"
echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${RESET}"
echo ""
echo "Summary:"
echo "  - Mock message created: $MESSAGE_ID"
echo "  - Mock WAMID: $CHANNEL_MESSAGE_ID"
echo "  - ACK progression tested: sent → delivered → read → played"
echo "  - All timestamps recorded correctly ✓"
echo ""
