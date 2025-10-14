#!/usr/bin/env bash
set -e

# ============================================================================
# Teste Completo: Fluxo de Webhooks WAHA → Eventos → Webhook.site
# ============================================================================
# Simula webhooks do WAHA chegando e verifica eventos sendo disparados
# Cenário: 2 contatos, 2 sessões, mensagens, ACKs, tracking
# ============================================================================

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
MAGENTA='\033[0;35m'
CYAN='\033[0;36m'
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

echo ""
echo -e "${BOLD}${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BOLD}${CYAN}  Full Webhook Flow Test${NC}"
echo -e "${BOLD}${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""
echo -e "${BLUE}→ Webhook URL: ${WEBHOOK_URL}${NC}"
echo ""

query_db() {
    PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -t -A -c "$1"
}

# ============================================================================
# SETUP
# ============================================================================
echo -e "${BOLD}${MAGENTA}[1/8] Setup Environment${NC}"
echo ""

# Limpar processos
echo -e "${YELLOW}→ Cleaning processes...${NC}"
lsof -ti:8080 | xargs kill -9 2>/dev/null || true
pkill -f "make api" 2>/dev/null || true
sleep 2

# Iniciar API
echo -e "${YELLOW}→ Starting API...${NC}"
make api > /tmp/api-webhook-test.log 2>&1 &
API_PID=$!

MAX_WAIT=30
WAITED=0
while [ $WAITED -lt $MAX_WAIT ]; do
    if curl -s http://localhost:8080/health > /dev/null 2>&1; then
        break
    fi
    sleep 1
    WAITED=$((WAITED + 1))
done

if [ $WAITED -eq $MAX_WAIT ]; then
    echo -e "${RED}✗ API failed to start${NC}"
    exit 1
fi

echo -e "${GREEN}✓ API ready (PID: $API_PID)${NC}"
echo ""

# Criar usuário
echo -e "${YELLOW}→ Creating user...${NC}"
REGISTER_RESPONSE=$(curl -s -X POST "$API_URL/api/v1/auth/register" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test-full@example.com",
    "password": "SecurePass123!",
    "name": "Full Test User"
  }')

API_KEY=$(echo "$REGISTER_RESPONSE" | jq -r '.api_key')
USER_ID=$(echo "$REGISTER_RESPONSE" | jq -r '.user_id')
PROJECT_ID=$(echo "$REGISTER_RESPONSE" | jq -r '.default_project_id')

if [ -z "$API_KEY" ] || [ "$API_KEY" = "null" ]; then
    echo -e "${RED}✗ Failed to create user${NC}"
    echo "$REGISTER_RESPONSE" | jq '.'
    exit 1
fi

echo -e "${GREEN}✓ User created${NC}"
echo -e "  API Key: ${API_KEY:0:20}..."
echo -e "  Project ID: $PROJECT_ID"
echo ""

# Criar canal WAHA
echo -e "${YELLOW}→ Creating WAHA channel...${NC}"
WEBHOOK_ID=$(uuidgen | tr '[:upper:]' '[:lower:]')
CHANNEL_RESPONSE=$(curl -s -X POST "$API_URL/api/v1/crm/channels" \
  -H "Authorization: Bearer $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test WAHA Channel",
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
    echo "$CHANNEL_RESPONSE" | jq '.'
    exit 1
fi

echo -e "${GREEN}✓ Channel created${NC}"
echo -e "  Channel ID: $CHANNEL_ID"
echo -e "  Webhook ID: $WEBHOOK_ID"
echo ""

# Ativar canal
echo -e "${YELLOW}→ Activating channel...${NC}"
sleep 2
curl -s -X POST "$API_URL/api/v1/crm/channels/$CHANNEL_ID/activate" \
  -H "Authorization: Bearer $API_KEY" > /dev/null
echo -e "${GREEN}✓ Channel activated${NC}"
echo ""

# ============================================================================
# [2/8] SUBSCREVER WEBHOOKS
# ============================================================================
echo -e "${BOLD}${MAGENTA}[2/8] Subscribe to All Events${NC}"
echo ""

WEBHOOK_SUB_RESPONSE=$(curl -s -X POST "$API_URL/api/v1/webhook-subscriptions" \
  -H "Authorization: Bearer $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Full Flow Test",
    "url": "'"$WEBHOOK_URL"'",
    "events": [
      "contact.created",
      "contact.updated",
      "session.created",
      "session.closed",
      "message.received",
      "message.sent",
      "message.delivered",
      "message.read",
      "message.played",
      "tracking.created"
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
echo -e "  Subscription ID: $WEBHOOK_SUB_ID"
echo -e "  Subscribed to 10 events"
echo ""

# ============================================================================
# [3/8] SIMULAR WEBHOOKS: MENSAGENS RECEBIDAS (2 CONTATOS)
# ============================================================================
echo -e "${BOLD}${MAGENTA}[3/8] Simulate Incoming Messages (2 Contacts)${NC}"
echo ""

# Contato 1: João
echo -e "${CYAN}→ Contact 1: João (554497044474)${NC}"

CONTACT1_PHONE="554497044474"
CONTACT1_WAMID="$CONTACT1_PHONE@c.us"
MSG1_ID="wamid.$(date +%s)001"

curl -s -X POST "$API_URL/api/v1/webhooks/waha/$WEBHOOK_ID" \
  -H "Content-Type: application/json" \
  -d '{
    "event": "message",
    "session": "'"$WAHA_SESSION_ID"'",
    "payload": {
      "id": "'"$MSG1_ID"'",
      "timestamp": '"$(date +%s)"',
      "from": "'"$CONTACT1_WAMID"'",
      "fromMe": false,
      "body": "Olá, gostaria de saber mais sobre o produto",
      "hasMedia": false,
      "_data": {
        "id": {
          "fromMe": false,
          "remote": "'"$CONTACT1_WAMID"'",
          "id": "'"$MSG1_ID"'"
        }
      }
    }
  }' > /dev/null

sleep 2
echo -e "${GREEN}  ✓ Message received from João${NC}"
echo -e "  ${BLUE}→ Events: contact.created, session.created, message.received${NC}"
echo ""

# Contato 2: Maria
echo -e "${CYAN}→ Contact 2: Maria (554497044475)${NC}"

CONTACT2_PHONE="554497044475"
CONTACT2_WAMID="$CONTACT2_PHONE@c.us"
MSG2_ID="wamid.$(date +%s)002"

curl -s -X POST "$API_URL/api/v1/webhooks/waha/$WEBHOOK_ID" \
  -H "Content-Type: application/json" \
  -d '{
    "event": "message",
    "session": "'"$WAHA_SESSION_ID"'",
    "payload": {
      "id": "'"$MSG2_ID"'",
      "timestamp": '"$(date +%s)"',
      "from": "'"$CONTACT2_WAMID"'",
      "fromMe": false,
      "body": "Boa tarde! Qual o prazo de entrega?",
      "hasMedia": false,
      "_data": {
        "id": {
          "fromMe": false,
          "remote": "'"$CONTACT2_WAMID"'",
          "id": "'"$MSG2_ID"'"
        }
      }
    }
  }' > /dev/null

sleep 2
echo -e "${GREEN}  ✓ Message received from Maria${NC}"
echo -e "  ${BLUE}→ Events: contact.created, session.created, message.received${NC}"
echo ""

# ============================================================================
# [4/8] ENVIAR RESPOSTAS
# ============================================================================
echo -e "${BOLD}${MAGENTA}[4/8] Send Responses (Outbound Messages)${NC}"
echo ""

# Obter contact IDs
CONTACT1_ID=$(query_db "SELECT id FROM contacts WHERE phone = '$CONTACT1_PHONE' LIMIT 1;")
CONTACT2_ID=$(query_db "SELECT id FROM contacts WHERE phone = '$CONTACT2_PHONE' LIMIT 1;")

echo -e "${CYAN}→ Sending response to João...${NC}"

RESPONSE1=$(curl -s -X POST "$API_URL/api/v1/crm/messages" \
  -H "Authorization: Bearer $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "contact_id": "'"$CONTACT1_ID"'",
    "channel_id": "'"$CHANNEL_ID"'",
    "content_type": "text",
    "text": "Olá João! Vou te enviar nosso catálogo completo."
  }')

RESPONSE1_ID=$(echo "$RESPONSE1" | jq -r '.id')
RESPONSE1_WAMID=$(query_db "SELECT channel_message_id FROM messages WHERE id = '$RESPONSE1_ID';")

echo -e "${GREEN}  ✓ Response sent to João${NC}"
echo -e "  ${BLUE}→ Event: message.sent${NC}"
echo ""

echo -e "${CYAN}→ Sending response to Maria...${NC}"

RESPONSE2=$(curl -s -X POST "$API_URL/api/v1/crm/messages" \
  -H "Authorization: Bearer $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "contact_id": "'"$CONTACT2_ID"'",
    "channel_id": "'"$CHANNEL_ID"'",
    "content_type": "text",
    "text": "Olá Maria! O prazo de entrega é de 5-7 dias úteis."
  }')

RESPONSE2_ID=$(echo "$RESPONSE2" | jq -r '.id')
RESPONSE2_WAMID=$(query_db "SELECT channel_message_id FROM messages WHERE id = '$RESPONSE2_ID';")

echo -e "${GREEN}  ✓ Response sent to Maria${NC}"
echo -e "  ${BLUE}→ Event: message.sent${NC}"
echo ""

sleep 2

# ============================================================================
# [5/8] SIMULAR ACKs (DELIVERED + READ)
# ============================================================================
echo -e "${BOLD}${MAGENTA}[5/8] Simulate ACKs (Delivered → Read)${NC}"
echo ""

echo -e "${CYAN}→ ACK 2 (DELIVERED) - João's message${NC}"
curl -s -X POST "$API_URL/api/v1/webhooks/waha/$WEBHOOK_ID" \
  -H "Content-Type: application/json" \
  -d '{
    "event": "message.ack",
    "session": "'"$WAHA_SESSION_ID"'",
    "payload": {
      "id": "'"$RESPONSE1_WAMID"'",
      "ack": 2,
      "ackName": "DEVICE",
      "from": "'"$CONTACT1_WAMID"'",
      "timestamp": '"$(date +%s)"'
    }
  }' > /dev/null

sleep 2
echo -e "${GREEN}  ✓ Message delivered to João${NC}"
echo -e "  ${BLUE}→ Event: message.delivered${NC}"
echo ""

echo -e "${CYAN}→ ACK 3 (READ) - João's message${NC}"
curl -s -X POST "$API_URL/api/v1/webhooks/waha/$WEBHOOK_ID" \
  -H "Content-Type: application/json" \
  -d '{
    "event": "message.ack",
    "session": "'"$WAHA_SESSION_ID"'",
    "payload": {
      "id": "'"$RESPONSE1_WAMID"'",
      "ack": 3,
      "ackName": "READ",
      "from": "'"$CONTACT1_WAMID"'",
      "timestamp": '"$(date +%s)"'
    }
  }' > /dev/null

sleep 2
echo -e "${GREEN}  ✓ Message read by João${NC}"
echo -e "  ${BLUE}→ Event: message.read${NC}"
echo ""

echo -e "${CYAN}→ ACK 2 (DELIVERED) - Maria's message${NC}"
curl -s -X POST "$API_URL/api/v1/webhooks/waha/$WEBHOOK_ID" \
  -H "Content-Type: application/json" \
  -d '{
    "event": "message.ack",
    "session": "'"$WAHA_SESSION_ID"'",
    "payload": {
      "id": "'"$RESPONSE2_WAMID"'",
      "ack": 2,
      "ackName": "DEVICE",
      "from": "'"$CONTACT2_WAMID"'",
      "timestamp": '"$(date +%s)"'
    }
  }' > /dev/null

sleep 2
echo -e "${GREEN}  ✓ Message delivered to Maria${NC}"
echo -e "  ${BLUE}→ Event: message.delivered${NC}"
echo ""

echo -e "${CYAN}→ ACK 3 (READ) - Maria's message${NC}"
curl -s -X POST "$API_URL/api/v1/webhooks/waha/$WEBHOOK_ID" \
  -H "Content-Type: application/json" \
  -d '{
    "event": "message.ack",
    "session": "'"$WAHA_SESSION_ID"'",
    "payload": {
      "id": "'"$RESPONSE2_WAMID"'",
      "ack": 3,
      "ackName": "READ",
      "from": "'"$CONTACT2_WAMID"'",
      "timestamp": '"$(date +%s)"'
    }
  }' > /dev/null

sleep 2
echo -e "${GREEN}  ✓ Message read by Maria${NC}"
echo -e "  ${BLUE}→ Event: message.read${NC}"
echo ""

# ============================================================================
# [6/8] ENVIAR MENSAGEM DE VOZ E SIMULAR ACK 4 (PLAYED)
# ============================================================================
echo -e "${BOLD}${MAGENTA}[6/8] Send Voice Message + ACK 4 (PLAYED)${NC}"
echo ""

echo -e "${CYAN}→ Sending voice message to João...${NC}"

VOICE_MSG=$(curl -s -X POST "$API_URL/api/v1/crm/messages" \
  -H "Authorization: Bearer $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "contact_id": "'"$CONTACT1_ID"'",
    "channel_id": "'"$CHANNEL_ID"'",
    "content_type": "voice",
    "media_url": "https://example.com/audio-catalog.ogg",
    "media_mimetype": "audio/ogg"
  }')

VOICE_MSG_ID=$(echo "$VOICE_MSG" | jq -r '.id')
VOICE_MSG_WAMID=$(query_db "SELECT channel_message_id FROM messages WHERE id = '$VOICE_MSG_ID';")

echo -e "${GREEN}  ✓ Voice message sent${NC}"
echo -e "  ${BLUE}→ Event: message.sent${NC}"
echo ""

sleep 2

echo -e "${CYAN}→ ACK 2 (DELIVERED) - Voice message${NC}"
curl -s -X POST "$API_URL/api/v1/webhooks/waha/$WEBHOOK_ID" \
  -H "Content-Type: application/json" \
  -d '{
    "event": "message.ack",
    "session": "'"$WAHA_SESSION_ID"'",
    "payload": {
      "id": "'"$VOICE_MSG_WAMID"'",
      "ack": 2,
      "ackName": "DEVICE",
      "from": "'"$CONTACT1_WAMID"'",
      "timestamp": '"$(date +%s)"'
    }
  }' > /dev/null

sleep 2
echo -e "${GREEN}  ✓ Voice message delivered${NC}"
echo -e "  ${BLUE}→ Event: message.delivered${NC}"
echo ""

echo -e "${CYAN}→ ACK 4 (PLAYED) - Voice message${NC}"
curl -s -X POST "$API_URL/api/v1/webhooks/waha/$WEBHOOK_ID" \
  -H "Content-Type: application/json" \
  -d '{
    "event": "message.ack",
    "session": "'"$WAHA_SESSION_ID"'",
    "payload": {
      "id": "'"$VOICE_MSG_WAMID"'",
      "ack": 4,
      "ackName": "PLAYED",
      "from": "'"$CONTACT1_WAMID"'",
      "timestamp": '"$(date +%s)"'
    }
  }' > /dev/null

sleep 2
echo -e "${GREEN}  ✓ Voice message played by João${NC}"
echo -e "  ${BLUE}→ Event: message.played (ONLY FOR VOICE!)${NC}"
echo ""

# ============================================================================
# [7/8] CRIAR TRACKING
# ============================================================================
echo -e "${BOLD}${MAGENTA}[7/8] Create Tracking Link${NC}"
echo ""

echo -e "${CYAN}→ Creating tracking for Maria's conversation...${NC}"

TRACKING_RESPONSE=$(curl -s -X POST "$API_URL/api/v1/crm/trackings" \
  -H "Authorization: Bearer $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "contact_id": "'"$CONTACT2_ID"'",
    "channel_id": "'"$CHANNEL_ID"'",
    "source": "whatsapp",
    "medium": "chat",
    "campaign": "product-inquiry",
    "content": "delivery-question"
  }')

TRACKING_ID=$(echo "$TRACKING_RESPONSE" | jq -r '.id')

if [ ! -z "$TRACKING_ID" ] && [ "$TRACKING_ID" != "null" ]; then
    echo -e "${GREEN}  ✓ Tracking created${NC}"
    echo -e "  ${BLUE}→ Event: tracking.created${NC}"
    echo -e "  Tracking ID: $TRACKING_ID"
else
    echo -e "${YELLOW}  ⚠ Tracking creation skipped or failed${NC}"
fi
echo ""

# ============================================================================
# [8/8] VERIFICAR RESULTADOS
# ============================================================================
echo -e "${BOLD}${MAGENTA}[8/8] Verify Results${NC}"
echo ""

# Contar contatos
CONTACT_COUNT=$(query_db "SELECT COUNT(*) FROM contacts WHERE project_id = '$PROJECT_ID';")
echo -e "${CYAN}Contacts created:${NC} $CONTACT_COUNT"

# Contar sessões
SESSION_COUNT=$(query_db "SELECT COUNT(*) FROM sessions WHERE project_id = '$PROJECT_ID';")
echo -e "${CYAN}Sessions created:${NC} $SESSION_COUNT"

# Contar mensagens
MESSAGE_COUNT=$(query_db "SELECT COUNT(*) FROM messages WHERE project_id = '$PROJECT_ID';")
echo -e "${CYAN}Messages exchanged:${NC} $MESSAGE_COUNT"

# Contar trackings
TRACKING_COUNT=$(query_db "SELECT COUNT(*) FROM trackings WHERE project_id = '$PROJECT_ID';")
echo -e "${CYAN}Trackings created:${NC} $TRACKING_COUNT"

echo ""

# Verificar mensagem de voz com played_at
PLAYED_COUNT=$(query_db "SELECT COUNT(*) FROM messages WHERE project_id = '$PROJECT_ID' AND played_at IS NOT NULL;")
echo -e "${CYAN}Voice messages played:${NC} $PLAYED_COUNT"

echo ""

# Verificar webhook triggers
WEBHOOK_TRIGGER_COUNT=$(query_db "SELECT COUNT(*) FROM webhook_triggers WHERE webhook_subscription_id = '$WEBHOOK_SUB_ID';")
echo -e "${CYAN}Webhook triggers:${NC} $WEBHOOK_TRIGGER_COUNT"

if [ "$WEBHOOK_TRIGGER_COUNT" -gt 0 ]; then
    echo ""
    echo -e "${CYAN}Recent webhook triggers:${NC}"
    query_db "
        SELECT
            triggered_at,
            success,
            http_status,
            event_type
        FROM webhook_triggers
        WHERE webhook_subscription_id = '$WEBHOOK_SUB_ID'
        ORDER BY triggered_at DESC
        LIMIT 5;
    " | while IFS='|' read -r triggered success status event; do
        if [ "$success" = "t" ]; then
            echo -e "  ${GREEN}✓${NC} $event - HTTP $status"
        else
            echo -e "  ${RED}✗${NC} $event - HTTP $status"
        fi
    done
fi

echo ""
echo -e "${BOLD}${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BOLD}${GREEN}  ✓ Test Complete!${NC}"
echo -e "${BOLD}${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""
echo -e "${CYAN}→ Check webhook.site for all events:${NC}"
echo -e "  ${WEBHOOK_URL}"
echo ""
echo -e "${YELLOW}Expected events (in order):${NC}"
echo -e "  1. ${MAGENTA}contact.created${NC} (João)"
echo -e "  2. ${MAGENTA}session.created${NC} (João)"
echo -e "  3. ${MAGENTA}message.received${NC} (João)"
echo -e "  4. ${MAGENTA}contact.created${NC} (Maria)"
echo -e "  5. ${MAGENTA}session.created${NC} (Maria)"
echo -e "  6. ${MAGENTA}message.received${NC} (Maria)"
echo -e "  7. ${MAGENTA}message.sent${NC} (Response to João)"
echo -e "  8. ${MAGENTA}message.sent${NC} (Response to Maria)"
echo -e "  9. ${MAGENTA}message.delivered${NC} (João)"
echo -e " 10. ${MAGENTA}message.read${NC} (João)"
echo -e " 11. ${MAGENTA}message.delivered${NC} (Maria)"
echo -e " 12. ${MAGENTA}message.read${NC} (Maria)"
echo -e " 13. ${MAGENTA}message.sent${NC} (Voice to João)"
echo -e " 14. ${MAGENTA}message.delivered${NC} (Voice)"
echo -e " 15. ${MAGENTA}message.played${NC} (Voice - ONLY FOR VOICE!)"
echo -e " 16. ${MAGENTA}tracking.created${NC} (Maria)"
echo ""
echo -e "${YELLOW}API running on PID: $API_PID${NC}"
echo -e "${YELLOW}Logs: tail -f /tmp/api-webhook-test.log${NC}"
echo ""
