#!/bin/bash

# Quick script to send ACK webhook for an existing message
# Usage: ./send_ack_webhook.sh <channel_message_id> <ack_number> [webhook_id]

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
CYAN='\033[0;36m'
YELLOW='\033[1;33m'
RESET='\033[0m'

API_BASE="http://localhost:8080/api/v1"

# Source .env
if [ -f .env ]; then
    source .env
fi

WAHA_SESSION_ID="${WAHA_DEFAULT_SESSION_ID_TEST:-test-session}"
TEST_PHONE="${TEST_PHONE_NUMBER:-5544970444747}"

# Parse arguments
CHANNEL_MESSAGE_ID=$1
ACK_NUM=$2
WEBHOOK_ID=$3

if [ -z "$CHANNEL_MESSAGE_ID" ] || [ -z "$ACK_NUM" ]; then
    echo "Usage: $0 <channel_message_id> <ack_number> [webhook_id]"
    echo ""
    echo "Arguments:"
    echo "  channel_message_id  - The WAMID (e.g., wamid.HBgNNTU0NDk3...)"
    echo "  ack_number          - ACK number (1=SERVER, 2=DEVICE, 3=READ, 4=PLAYED)"
    echo "  webhook_id          - Optional webhook ID (if not provided, will try to get from last channel)"
    echo ""
    echo "Examples:"
    echo "  $0 'wamid.HBgN...' 1                    # Send ACK 1 (SERVER)"
    echo "  $0 'wamid.HBgN...' 2 webhook-uuid-123   # Send ACK 2 (DEVICE) to specific webhook"
    echo ""
    echo "ACK Numbers:"
    echo "  -1 = ERROR   (failed)"
    echo "   0 = PENDING (queued)"
    echo "   1 = SERVER  (sent)"
    echo "   2 = DEVICE  (delivered ✓✓)"
    echo "   3 = READ    (read ✓✓ blue)"
    echo "   4 = PLAYED  (media played/viewed ▶️)"
    exit 1
fi

# Map ACK number to name
case $ACK_NUM in
    -1) ACK_NAME="ERROR" ;;
    0)  ACK_NAME="PENDING" ;;
    1)  ACK_NAME="SERVER" ;;
    2)  ACK_NAME="DEVICE" ;;
    3)  ACK_NAME="READ" ;;
    4)  ACK_NAME="PLAYED" ;;
    *)
        echo -e "${RED}✗ Invalid ACK number: $ACK_NUM${RESET}"
        echo "Valid values: -1, 0, 1, 2, 3, 4"
        exit 1
        ;;
esac

echo ""
echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${RESET}"
echo -e "${CYAN}  Send ACK Webhook${RESET}"
echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${RESET}"
echo ""
echo "Parameters:"
echo "  Channel Message ID: $CHANNEL_MESSAGE_ID"
echo "  ACK Number: $ACK_NUM ($ACK_NAME)"

# If webhook_id not provided, try to get from database
if [ -z "$WEBHOOK_ID" ]; then
    echo ""
    echo -e "${YELLOW}→ Webhook ID not provided, querying database...${RESET}"

    # Try to get webhook_id from the most recent WAHA channel
    DB_HOST="${DB_HOST:-localhost}"
    DB_PORT="${DB_PORT:-5432}"
    DB_NAME="${DB_NAME:-ventros_crm}"
    DB_USER="${DB_USER:-postgres}"
    DB_PASSWORD="${DB_PASSWORD:-postgres}"

    WEBHOOK_ID=$(PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -t -A -c \
        "SELECT webhook_id FROM channels WHERE type = 'waha' AND deleted_at IS NULL ORDER BY created_at DESC LIMIT 1;")

    if [ -z "$WEBHOOK_ID" ] || [ "$WEBHOOK_ID" == "" ]; then
        echo -e "${RED}✗ Could not find webhook_id in database${RESET}"
        echo "Please provide webhook_id as third argument"
        exit 1
    fi

    echo -e "${GREEN}✓ Found webhook_id: $WEBHOOK_ID${RESET}"
fi

echo "  Webhook ID: $WEBHOOK_ID"
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

echo "Webhook Payload:"
echo "$ACK_PAYLOAD" | jq '.'
echo ""

# Send webhook
echo -e "${CYAN}→ Sending webhook to $API_BASE/webhooks/$WEBHOOK_ID${RESET}"
WEBHOOK_RESPONSE=$(curl -s -w "\nHTTP_STATUS:%{http_code}" -X POST "$API_BASE/webhooks/$WEBHOOK_ID" \
  -H "Content-Type: application/json" \
  -d "$ACK_PAYLOAD")

HTTP_STATUS=$(echo "$WEBHOOK_RESPONSE" | grep "HTTP_STATUS:" | cut -d: -f2)
BODY=$(echo "$WEBHOOK_RESPONSE" | sed '/HTTP_STATUS:/d')

echo ""
echo "Response (HTTP $HTTP_STATUS):"
echo "$BODY" | jq '.'
echo ""

if [ "$HTTP_STATUS" == "200" ]; then
    echo -e "${GREEN}✓ ACK webhook sent successfully!${RESET}"
    echo ""
    echo "The ACK has been queued for processing."
    echo "Wait 1-2 seconds and check the message status in the database:"
    echo ""
    echo "  SELECT id, status, delivered_at, read_at, played_at"
    echo "  FROM messages"
    echo "  WHERE channel_message_id = '$CHANNEL_MESSAGE_ID';"
    echo ""
else
    echo -e "${RED}✗ Webhook failed (HTTP $HTTP_STATUS)${RESET}"
    exit 1
fi
