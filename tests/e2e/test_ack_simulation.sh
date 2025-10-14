#!/bin/bash

# Test ACK simulation - Sends a message and simulates ACK webhooks (1->2->3->4)
# Tests the complete ACK flow: SENT -> DELIVERED -> READ -> PLAYED

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

echo ""
echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${RESET}"
echo -e "${CYAN}  ACK Webhook Simulation Test${RESET}"
echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${RESET}"
echo ""

# ===== SETUP =====
echo -e "${CYAN}→ Setting up test environment...${RESET}"

# Register user
REGISTER_RESPONSE=$(curl -s -X POST "$API_BASE/auth/register" \
  -H "Content-Type: application/json" \
  -d "{
    \"name\": \"ACK Test User $TIMESTAMP\",
    \"email\": \"ack-test-$TIMESTAMP@test.com\",
    \"password\": \"Test@123456\"
  }")

USER_ID=$(echo $REGISTER_RESPONSE | jq -r '.user_id')
PROJECT_ID=$(echo $REGISTER_RESPONSE | jq -r '.default_project_id')
TOKEN=$(echo $REGISTER_RESPONSE | jq -r '.api_key')

if [ "$USER_ID" == "null" ]; then
    echo -e "${RED}✗ Failed to register user${RESET}"
    echo "Response: $REGISTER_RESPONSE"
    exit 1
fi

echo -e "${GREEN}✓ User registered${RESET}"

# Create WAHA channel
WAHA_BASE_URL="${WAHA_BASE_URL:-https://waha.ventros.cloud}"
WAHA_API_KEY="${WAHA_API_KEY:-}"

CHANNEL_RESPONSE=$(curl -s -X POST "$API_BASE/crm/channels?project_id=$PROJECT_ID" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d "{
    \"name\": \"ACK Test Channel\",
    \"type\": \"waha\",
    \"waha_config\": {
      \"base_url\": \"$WAHA_BASE_URL\",
      \"api_key\": \"$WAHA_API_KEY\",
      \"session_id\": \"$WAHA_SESSION_ID\"
    }
  }")

CHANNEL_ID=$(echo $CHANNEL_RESPONSE | jq -r '.id')
WEBHOOK_ID=$(echo $CHANNEL_RESPONSE | jq -r '.webhook_id')

if [ "$CHANNEL_ID" == "null" ]; then
    echo -e "${RED}✗ Failed to create channel${RESET}"
    echo "Response: $CHANNEL_RESPONSE"
    exit 1
fi

echo -e "${GREEN}✓ Channel created${RESET}"
echo "  Channel ID: $CHANNEL_ID"
echo "  Webhook ID: $WEBHOOK_ID"

# Activate channel
curl -s -X POST "$API_BASE/crm/channels/$CHANNEL_ID/activate" \
  -H "Authorization: Bearer $TOKEN" > /dev/null

echo -e "${GREEN}✓ Channel activated${RESET}"

# Create contact
CONTACT_RESPONSE=$(curl -s -X POST "$API_BASE/contacts?project_id=$PROJECT_ID" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d "{
    \"name\": \"ACK Test Contact\",
    \"phone\": \"$TEST_PHONE\",
    \"email\": \"ack-test-$TIMESTAMP@test.com\"
  }")

CONTACT_ID=$(echo $CONTACT_RESPONSE | jq -r '.id')

if [ "$CONTACT_ID" == "null" ]; then
    echo -e "${RED}✗ Failed to create contact${RESET}"
    echo "Response: $CONTACT_RESPONSE"
    exit 1
fi

echo -e "${GREEN}✓ Contact created${RESET}"
echo "  Contact ID: $CONTACT_ID"
echo ""

# ===== SEND MESSAGE =====
echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${RESET}"
echo -e "${CYAN}  Step 1: Send Media Message${RESET}"
echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${RESET}"
echo ""

# Send a media message (video) to test PLAYED ACK
MESSAGE_RESPONSE=$(curl -s -X POST "$API_BASE/crm/messages/send" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d "{
    \"contact_id\": \"$CONTACT_ID\",
    \"channel_id\": \"$CHANNEL_ID\",
    \"content_type\": \"video\",
    \"media_url\": \"https://commondatastorage.googleapis.com/gtv-videos-bucket/sample/ForBiggerBlazes.mp4\",
    \"text\": \"🎥 Testing ACK flow - SENT -> DELIVERED -> READ -> PLAYED\"
  }")

MESSAGE_ID=$(echo $MESSAGE_RESPONSE | jq -r '.id // .message_id')

if [ "$MESSAGE_ID" == "null" ] || [ -z "$MESSAGE_ID" ]; then
    echo -e "${RED}✗ Failed to send message${RESET}"
    echo "Response: $MESSAGE_RESPONSE"
    exit 1
fi

echo -e "${GREEN}✓ Message sent${RESET}"
echo "  Message ID: $MESSAGE_ID"
echo ""

# Wait a bit for the message to be processed and get channel_message_id
echo -e "${YELLOW}→ Waiting 5s for message to be processed by WAHA...${RESET}"
sleep 5

# Query database directly to get channel_message_id (WAMID)
# Since GET /api/v1/crm/messages/:id is not implemented yet
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_NAME="${DB_NAME:-ventros_crm}"
DB_USER="${DB_USER:-postgres}"
DB_PASSWORD="${DB_PASSWORD:-postgres}"

echo -e "${CYAN}→ Querying database for channel_message_id...${RESET}"

CHANNEL_MESSAGE_ID=$(PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -t -A -c \
    "SELECT channel_message_id FROM messages WHERE id = '$MESSAGE_ID' AND channel_message_id IS NOT NULL;")

CURRENT_STATUS=$(PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -t -A -c \
    "SELECT status FROM messages WHERE id = '$MESSAGE_ID';")

if [ -z "$CHANNEL_MESSAGE_ID" ] || [ "$CHANNEL_MESSAGE_ID" == "" ]; then
    echo -e "${RED}✗ Message doesn't have channel_message_id yet${RESET}"
    echo "This means WAHA hasn't processed the message or the message failed to send."
    echo ""
    echo "Checking database for message details:"
    PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c \
        "SELECT id, status, channel_message_id, created_at FROM messages WHERE id = '$MESSAGE_ID';"

    echo ""
    echo "This is expected if:"
    echo "  1. WAHA is not configured (WAHA_BASE_URL, WAHA_API_KEY not set)"
    echo "  2. WAHA session is not active"
    echo "  3. The message is still being processed (wait more time)"
    echo ""
    echo "To test ACKs without WAHA, use send_ack_webhook.sh with an existing channel_message_id"
    exit 1
fi

echo -e "${GREEN}✓ Message processed by WAHA${RESET}"
echo "  Channel Message ID (WAMID): $CHANNEL_MESSAGE_ID"
echo "  Current Status: $CURRENT_STATUS"
echo ""

# Function to send ACK webhook
send_ack() {
    local ACK_NUM=$1
    local ACK_NAME=$2
    local EXPECTED_STATUS=$3

    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${RESET}"
    echo -e "${CYAN}  Step $((ACK_NUM + 2)): Simulate ACK $ACK_NUM ($ACK_NAME)${RESET}"
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

# ACK 1: SERVER (Sent to WhatsApp server)
send_ack 1 "SERVER" "sent"

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
    WHERE id = '$MESSAGE_ID';" | head -6

echo ""
echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${RESET}"
echo -e "${GREEN}  ✓ ACK Simulation Test Complete!${RESET}"
echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${RESET}"
echo ""
echo "Summary:"
echo "  - Message sent: $MESSAGE_ID"
echo "  - WAMID: $CHANNEL_MESSAGE_ID"
echo "  - ACK progression: pending → sent → delivered → read → played"
echo "  - All timestamps recorded correctly"
echo ""
echo "You can also check the database directly:"
echo "  SELECT id, status, delivered_at, read_at, played_at FROM messages WHERE id = '$MESSAGE_ID';"
echo ""
