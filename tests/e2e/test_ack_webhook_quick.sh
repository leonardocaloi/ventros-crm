#!/usr/bin/env bash
set -e

# ============================================================================
# Teste Rápido: ACKs + Webhooks (SEM RESET)
# ============================================================================
# Este script testa webhooks SEM fazer reset completo (mais rápido)
# ============================================================================

# Cores
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'
BOLD='\033[1m'

WEBHOOK_URL="${1:-https://webhook.site/9bc9e1ce-9fe0-497b-bbdd-70034a76043a}"

API_URL="http://localhost:8080"
DB_HOST="localhost"
DB_PORT="5432"
DB_NAME="ventros_crm"
DB_USER="postgres"
DB_PASSWORD="postgres"

if [ -f .env ]; then
    source .env
fi

WAHA_SESSION_ID="${WAHA_DEFAULT_SESSION_ID_TEST:-test-session}"
TEST_PHONE="${TEST_PHONE_NUMBER:-5544970444747}"

echo ""
echo -e "${BOLD}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BOLD}  Quick ACK + Webhook Test${NC}"
echo -e "${BOLD}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""
echo -e "${BLUE}→ Webhook URL: ${WEBHOOK_URL}${NC}"
echo ""

query_db() {
    PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -t -A -c "$1"
}

# Matar processos duplicados
echo -e "${YELLOW}→ Cleaning up duplicate processes...${NC}"
lsof -ti:8080 | xargs kill -9 2>/dev/null || true
pkill -f "make api" 2>/dev/null || true
sleep 2
echo -e "${GREEN}✓ Processes cleaned${NC}"
echo ""

# Iniciar API
echo -e "${YELLOW}→ Starting API...${NC}"
make api > /tmp/ventros-api-quick.log 2>&1 &
API_PID=$!

MAX_WAIT=30
WAITED=0
while [ $WAITED -lt $MAX_WAIT ]; do
    if curl -s http://localhost:8080/health > /dev/null 2>&1; then
        echo -e "${GREEN}✓ API ready (PID: $API_PID)${NC}"
        break
    fi
    sleep 1
    WAITED=$((WAITED + 1))
done

if [ $WAITED -eq $MAX_WAIT ]; then
    echo -e "${RED}✗ API failed to start${NC}"
    exit 1
fi
echo ""

# Registrar usuário
echo -e "${YELLOW}→ Creating test user...${NC}"
REGISTER_RESPONSE=$(curl -s -X POST "$API_URL/api/v1/auth/register" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "webhook-test@example.com",
    "password": "SecurePass123!",
    "name": "Webhook Test"
  }')

TOKEN=$(echo "$REGISTER_RESPONSE" | jq -r '.token // .access_token // .api_key')
PROJECT_ID=$(echo "$REGISTER_RESPONSE" | jq -r '.project.id // .project_id // .default_project_id')
USER_ID=$(echo "$REGISTER_RESPONSE" | jq -r '.user.id // .user_id')

if [ -z "$TOKEN" ] || [ "$TOKEN" = "null" ]; then
    echo -e "${RED}✗ Failed to register user${NC}"
    echo "$REGISTER_RESPONSE" | jq '.'
    exit 1
fi

echo -e "${GREEN}✓ User created${NC}"
echo -e "  User ID: $USER_ID"
echo -e "  Project ID: $PROJECT_ID"
echo ""

# Criar canal
echo -e "${YELLOW}→ Creating channel...${NC}"
WEBHOOK_ID=$(uuidgen | tr '[:upper:]' '[:lower:]')
CHANNEL_RESPONSE=$(curl -s -X POST "$API_URL/api/v1/crm/channels" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Webhook Test Channel",
    "type": "waha",
    "waha_config": {
      "base_url": "http://localhost:3000",
      "api_key": "test-api-key",
      "session_id": "'"$WAHA_SESSION_ID"'",
      "webhook_url": "'"$API_URL/api/v1/webhooks/waha/$WEBHOOK_ID"'",
      "import_strategy": "none"
    }
  }')

CHANNEL_ID=$(echo "$CHANNEL_RESPONSE" | jq -r '.id')

if [ -z "$CHANNEL_ID" ] || [ "$CHANNEL_ID" = "null" ]; then
    echo -e "${RED}✗ Failed to create channel${NC}"
    exit 1
fi

echo -e "${GREEN}✓ Channel created: $CHANNEL_ID${NC}"
echo ""

# Ativar canal
echo -e "${YELLOW}→ Activating channel...${NC}"
sleep 2
curl -s -X POST "$API_URL/api/v1/crm/channels/$CHANNEL_ID/activate" \
  -H "Authorization: Bearer $TOKEN" > /dev/null
echo -e "${GREEN}✓ Channel activated${NC}"
echo ""

# Criar contato
echo -e "${YELLOW}→ Creating contact...${NC}"
CONTACT_RESPONSE=$(curl -s -X POST "$API_URL/api/v1/crm/contacts" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Webhook Test Contact",
    "phone": "'"$TEST_PHONE"'",
    "whatsapp_identifiers": {
      "phone_number": "'"$TEST_PHONE"'",
      "waid": "'"$TEST_PHONE"'@c.us",
      "chat_id": "'"$TEST_PHONE"'@c.us"
    }
  }')

CONTACT_ID=$(echo "$CONTACT_RESPONSE" | jq -r '.id')
echo -e "${GREEN}✓ Contact created: $CONTACT_ID${NC}"
echo ""

# Subscrever webhooks
echo -e "${BOLD}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BOLD}  Creating Webhook Subscription${NC}"
echo -e "${BOLD}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""

WEBHOOK_SUB_RESPONSE=$(curl -s -X POST "$API_URL/api/v1/webhook-subscriptions" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Webhook.site Quick Test",
    "url": "'"$WEBHOOK_URL"'",
    "events": [
      "message.sent",
      "message.delivered",
      "message.read",
      "message.played",
      "message.failed"
    ],
    "retry_count": 3,
    "timeout_seconds": 30
  }')

WEBHOOK_SUB_ID=$(echo "$WEBHOOK_SUB_RESPONSE" | jq -r '.id')

if [ -z "$WEBHOOK_SUB_ID" ] || [ "$WEBHOOK_SUB_ID" = "null" ]; then
    echo -e "${RED}✗ Failed to create webhook subscription${NC}"
    echo "$WEBHOOK_SUB_RESPONSE" | jq '.'
    exit 1
fi

echo -e "${GREEN}✓ Webhook subscription created${NC}"
echo -e "  ID: $WEBHOOK_SUB_ID"
echo -e "  URL: $WEBHOOK_URL"
echo -e "  Events: message.sent, delivered, read, played"
echo ""

# Enviar mensagem
echo -e "${BOLD}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BOLD}  Sending Test Message${NC}"
echo -e "${BOLD}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""

MESSAGE_RESPONSE=$(curl -s -X POST "$API_URL/api/v1/crm/messages" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "contact_id": "'"$CONTACT_ID"'",
    "channel_id": "'"$CHANNEL_ID"'",
    "content_type": "voice",
    "media_url": "https://example.com/test-voice.ogg",
    "media_mimetype": "audio/ogg"
  }')

MESSAGE_ID=$(echo "$MESSAGE_RESPONSE" | jq -r '.id')
echo -e "${GREEN}✓ Voice message sent: $MESSAGE_ID${NC}"
sleep 2

WEBHOOK_ID=$(query_db "SELECT webhook_id FROM channels WHERE id = '$CHANNEL_ID';")
CHANNEL_MESSAGE_ID=$(query_db "SELECT channel_message_id FROM messages WHERE id = '$MESSAGE_ID';")

echo -e "  Webhook ID: $WEBHOOK_ID"
echo -e "  WAMID: $CHANNEL_MESSAGE_ID"
echo ""

echo -e "${YELLOW}→ Waiting for message.sent webhook (3s)...${NC}"
sleep 3
echo ""

# Simular ACKs
echo -e "${BOLD}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BOLD}  Simulating ACK Events${NC}"
echo -e "${BOLD}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""

send_ack() {
    local ack=$1
    local name=$2

    echo -e "${YELLOW}→ Sending ACK $ack ($name)...${NC}"

    curl -s -X POST "$API_URL/api/v1/webhooks/waha/$WEBHOOK_ID" \
      -H "Content-Type: application/json" \
      -d '{
        "event": "message.ack",
        "session": "'"$WAHA_SESSION_ID"'",
        "payload": {
          "id": "'"$CHANNEL_MESSAGE_ID"'",
          "ack": '"$ack"',
          "ackName": "'"$name"'",
          "from": "'"$TEST_PHONE"'@c.us",
          "timestamp": '"$(date +%s)"'
        }
      }' > /dev/null

    sleep 3

    STATUS=$(query_db "SELECT status FROM messages WHERE id = '$MESSAGE_ID';")
    echo -e "${GREEN}✓ ACK $ack processed - Status: $STATUS${NC}"
    echo ""
}

send_ack 1 "SERVER"
send_ack 2 "DEVICE"
send_ack 3 "READ"
send_ack 4 "PLAYED"

# Resultados
echo -e "${BOLD}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BOLD}  Final Results${NC}"
echo -e "${BOLD}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""

FINAL=$(query_db "
    SELECT
        status,
        delivered_at IS NOT NULL as delivered,
        read_at IS NOT NULL as read,
        played_at IS NOT NULL as played
    FROM messages
    WHERE id = '$MESSAGE_ID';
")

echo -e "${GREEN}Message Status:${NC}"
echo "$FINAL"
echo ""

echo -e "${BOLD}${GREEN}✓ Test Complete!${NC}"
echo ""
echo -e "${BLUE}→ Check webhook.site:${NC}"
echo -e "  ${WEBHOOK_URL}"
echo ""
echo -e "${YELLOW}Expected webhooks:${NC}"
echo -e "  1. message.sent"
echo -e "  2. message.delivered"
echo -e "  3. message.read"
echo -e "  4. message.played"
echo ""

# Verificar webhook triggers
echo -e "${YELLOW}Webhook Triggers:${NC}"
TRIGGERS=$(query_db "
    SELECT
        triggered_at,
        success,
        http_status
    FROM webhook_triggers
    WHERE webhook_subscription_id = '$WEBHOOK_SUB_ID'
    ORDER BY triggered_at DESC;
")

if [ ! -z "$TRIGGERS" ]; then
    echo "$TRIGGERS"
else
    echo -e "${YELLOW}  No triggers logged yet (check async processing)${NC}"
fi

echo ""
echo -e "${YELLOW}API PID: $API_PID${NC}"
echo -e "${YELLOW}Press Ctrl+C to stop${NC}"
wait $API_PID
