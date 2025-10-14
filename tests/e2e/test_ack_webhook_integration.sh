#!/usr/bin/env bash
set -e

# ============================================================================
# Teste de Integração: ACKs + Webhooks
# ============================================================================
# Este script testa o fluxo completo de ACKs com notificações de webhook
#
# Fluxo:
# 1. Limpa ambiente e reinicia API
# 2. Cria usuário, projeto, canal e contato
# 3. Subscreve eventos no webhook.site
# 4. Envia mensagem de vídeo
# 5. Simula ACKs progressivos (1 → 2 → 3 → 4)
# 6. Verifica se webhooks foram disparados
# ============================================================================

# Cores para output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color
BOLD='\033[1m'

# URL do webhook.site (passada como argumento ou default)
WEBHOOK_URL="${1:-https://webhook.site/9bc9e1ce-9fe0-497b-bbdd-70034a76043a}"

# Configurações
API_URL="http://localhost:8080"
DB_HOST="localhost"
DB_PORT="5432"
DB_NAME="ventros_crm"
DB_USER="postgres"
DB_PASSWORD="postgres"

# Carrega variáveis de ambiente
if [ -f .env ]; then
    source .env
fi

WAHA_SESSION_ID="${WAHA_DEFAULT_SESSION_ID_TEST:-test-session}"
TEST_PHONE="${TEST_PHONE_NUMBER:-5544970444747}"

echo ""
echo -e "${BOLD}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BOLD}  ACK + Webhook Integration Test${NC}"
echo -e "${BOLD}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""
echo -e "${BLUE}→ Webhook URL: ${WEBHOOK_URL}${NC}"
echo ""

# ============================================================================
# Função auxiliar: pretty print JSON
# ============================================================================
print_json() {
    if command -v jq &> /dev/null; then
        echo "$1" | jq '.'
    else
        echo "$1"
    fi
}

# ============================================================================
# Função auxiliar: query PostgreSQL
# ============================================================================
query_db() {
    PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -t -A -c "$1"
}

# ============================================================================
# STEP 1: Limpar processos duplicados da API
# ============================================================================
echo -e "${YELLOW}→ Killing duplicate API processes...${NC}"

# Mata processos na porta 8080
lsof -ti:8080 | xargs kill -9 2>/dev/null || true

# Mata processos make api
pkill -f "make api" 2>/dev/null || true

sleep 2
echo -e "${GREEN}✓ Processes cleaned${NC}"
echo ""

# ============================================================================
# STEP 2: Reset completo do ambiente
# ============================================================================
echo -e "${YELLOW}→ Resetting environment...${NC}"
make reset-full > /dev/null 2>&1
sleep 3
echo -e "${GREEN}✓ Environment reset${NC}"
echo ""

# ============================================================================
# STEP 3: Iniciar API em background
# ============================================================================
echo -e "${YELLOW}→ Starting API...${NC}"
make api > /tmp/ventros-api.log 2>&1 &
API_PID=$!

# Aguardar API estar pronta (até 30s)
MAX_WAIT=30
WAITED=0
while [ $WAITED -lt $MAX_WAIT ]; do
    if curl -s http://localhost:8080/health > /dev/null 2>&1; then
        echo -e "${GREEN}✓ API is ready (PID: $API_PID)${NC}"
        break
    fi
    sleep 1
    WAITED=$((WAITED + 1))
    echo -ne "${YELLOW}  Waiting for API... ${WAITED}s${NC}\r"
done

if [ $WAITED -eq $MAX_WAIT ]; then
    echo -e "${RED}✗ API failed to start within ${MAX_WAIT}s${NC}"
    exit 1
fi

echo ""

# ============================================================================
# STEP 4: Setup - Criar usuário, projeto, canal e contato
# ============================================================================
echo -e "${BOLD}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BOLD}  Setup: Creating test environment${NC}"
echo -e "${BOLD}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""

# 4.1 - Registrar usuário
echo -e "${YELLOW}→ Registering user...${NC}"
REGISTER_RESPONSE=$(curl -s -X POST "$API_URL/api/v1/auth/register" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "SecurePass123!",
    "first_name": "Test",
    "last_name": "User"
  }')

TOKEN=$(echo "$REGISTER_RESPONSE" | jq -r '.token // .access_token // empty')
USER_ID=$(echo "$REGISTER_RESPONSE" | jq -r '.user.id // .user_id // empty')
PROJECT_ID=$(echo "$REGISTER_RESPONSE" | jq -r '.project.id // .project_id // empty')
TENANT_ID=$(echo "$REGISTER_RESPONSE" | jq -r '.tenant_id // .customer_id // empty')

if [ -z "$TOKEN" ] || [ "$TOKEN" = "null" ]; then
    echo -e "${RED}✗ Failed to register user${NC}"
    print_json "$REGISTER_RESPONSE"
    exit 1
fi

echo -e "${GREEN}✓ User registered${NC}"
echo -e "  User ID: ${USER_ID}"
echo -e "  Project ID: ${PROJECT_ID}"
echo -e "  Tenant ID: ${TENANT_ID}"
echo ""

# 4.2 - Criar canal WAHA
echo -e "${YELLOW}→ Creating WAHA channel...${NC}"
CHANNEL_RESPONSE=$(curl -s -X POST "$API_URL/api/v1/crm/channels" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test WhatsApp Channel",
    "type": "waha",
    "session_id": "'"$WAHA_SESSION_ID"'",
    "webhook_id": "'"$(uuidgen | tr '[:upper:]' '[:lower:]')"'",
    "phone_number": "'"$TEST_PHONE"'"
  }')

CHANNEL_ID=$(echo "$CHANNEL_RESPONSE" | jq -r '.id // .channel_id // empty')

if [ -z "$CHANNEL_ID" ] || [ "$CHANNEL_ID" = "null" ]; then
    echo -e "${RED}✗ Failed to create channel${NC}"
    print_json "$CHANNEL_RESPONSE"
    exit 1
fi

echo -e "${GREEN}✓ Channel created${NC}"
echo -e "  Channel ID: ${CHANNEL_ID}"
echo ""

# Buscar webhook_id do banco (se não retornou na API)
WEBHOOK_ID=$(query_db "SELECT webhook_id FROM channels WHERE id = '$CHANNEL_ID';")
echo -e "  Webhook ID: ${WEBHOOK_ID}"
echo ""

# 4.3 - Ativar canal
echo -e "${YELLOW}→ Activating channel...${NC}"
sleep 2
ACTIVATE_RESPONSE=$(curl -s -X POST "$API_URL/api/v1/crm/channels/$CHANNEL_ID/activate" \
  -H "Authorization: Bearer $TOKEN")

echo -e "${GREEN}✓ Channel activated${NC}"
echo ""

# 4.4 - Criar contato
echo -e "${YELLOW}→ Creating contact...${NC}"
CONTACT_RESPONSE=$(curl -s -X POST "$API_URL/api/v1/crm/contacts" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test Contact",
    "phone": "'"$TEST_PHONE"'",
    "whatsapp_identifiers": {
      "phone_number": "'"$TEST_PHONE"'",
      "waid": "'"$TEST_PHONE"'@c.us",
      "chat_id": "'"$TEST_PHONE"'@c.us"
    }
  }')

CONTACT_ID=$(echo "$CONTACT_RESPONSE" | jq -r '.id // .contact_id // empty')

if [ -z "$CONTACT_ID" ] || [ "$CONTACT_ID" = "null" ]; then
    echo -e "${RED}✗ Failed to create contact${NC}"
    print_json "$CONTACT_RESPONSE"
    exit 1
fi

echo -e "${GREEN}✓ Contact created${NC}"
echo -e "  Contact ID: ${CONTACT_ID}"
echo ""

# ============================================================================
# STEP 5: Subscrever eventos no webhook.site
# ============================================================================
echo -e "${BOLD}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BOLD}  Subscribing to webhooks${NC}"
echo -e "${BOLD}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""

echo -e "${YELLOW}→ Creating webhook subscription...${NC}"
WEBHOOK_SUB_RESPONSE=$(curl -s -X POST "$API_URL/api/v1/webhook-subscriptions" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Webhook.site Integration Test",
    "url": "'"$WEBHOOK_URL"'",
    "events": [
      "message.sent",
      "message.delivered",
      "message.read",
      "message.played",
      "message.failed",
      "message.received"
    ],
    "retry_count": 3,
    "timeout_seconds": 30
  }')

WEBHOOK_SUB_ID=$(echo "$WEBHOOK_SUB_RESPONSE" | jq -r '.id // empty')

if [ -z "$WEBHOOK_SUB_ID" ] || [ "$WEBHOOK_SUB_ID" = "null" ]; then
    echo -e "${RED}✗ Failed to create webhook subscription${NC}"
    print_json "$WEBHOOK_SUB_RESPONSE"
    exit 1
fi

echo -e "${GREEN}✓ Webhook subscription created${NC}"
echo -e "  Subscription ID: ${WEBHOOK_SUB_ID}"
echo -e "  Subscribed events: message.sent, message.delivered, message.read, message.played"
echo -e "  Target URL: ${WEBHOOK_URL}"
echo ""

# ============================================================================
# STEP 6: Enviar mensagem de vídeo
# ============================================================================
echo -e "${BOLD}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BOLD}  Sending test message${NC}"
echo -e "${BOLD}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""

echo -e "${YELLOW}→ Sending video message...${NC}"
MESSAGE_RESPONSE=$(curl -s -X POST "$API_URL/api/v1/crm/messages" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "contact_id": "'"$CONTACT_ID"'",
    "channel_id": "'"$CHANNEL_ID"'",
    "content_type": "video",
    "media_url": "https://example.com/test-video.mp4",
    "media_mimetype": "video/mp4"
  }')

MESSAGE_ID=$(echo "$MESSAGE_RESPONSE" | jq -r '.id // .message_id // empty')

if [ -z "$MESSAGE_ID" ] || [ "$MESSAGE_ID" = "null" ]; then
    echo -e "${RED}✗ Failed to send message${NC}"
    print_json "$MESSAGE_RESPONSE"
    exit 1
fi

echo -e "${GREEN}✓ Message sent${NC}"
echo -e "  Message ID: ${MESSAGE_ID}"
echo ""

# Buscar channel_message_id (WAMID) do banco
sleep 2
CHANNEL_MESSAGE_ID=$(query_db "SELECT channel_message_id FROM messages WHERE id = '$MESSAGE_ID';")
echo -e "  WAMID: ${CHANNEL_MESSAGE_ID}"
echo ""

# Aguardar webhook de message.sent
echo -e "${YELLOW}→ Waiting for message.sent webhook (3s)...${NC}"
sleep 3
echo -e "${GREEN}✓ Check webhook.site for message.sent event${NC}"
echo ""

# ============================================================================
# STEP 7: Simular ACKs progressivos
# ============================================================================
echo -e "${BOLD}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BOLD}  Simulating ACKs${NC}"
echo -e "${BOLD}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""

# Helper: Enviar ACK webhook
send_ack() {
    local ack_value=$1
    local ack_name=$2
    local timestamp=$(date -u +"%Y-%m-%dT%H:%M:%S.%3NZ")

    echo -e "${YELLOW}→ Simulating ACK ${ack_value} (${ack_name})...${NC}"

    ACK_RESPONSE=$(curl -s -X POST "$API_URL/api/v1/webhooks/waha/$WEBHOOK_ID" \
      -H "Content-Type: application/json" \
      -d '{
        "event": "message.ack",
        "session": "'"$WAHA_SESSION_ID"'",
        "payload": {
          "id": "'"$CHANNEL_MESSAGE_ID"'",
          "ack": '"$ack_value"',
          "ackName": "'"$ack_name"'",
          "from": "'"$TEST_PHONE"'@c.us",
          "timestamp": '"$(date +%s)"'
        }
      }')

    ACK_STATUS=$(echo "$ACK_RESPONSE" | jq -r '.status // "unknown"')

    if [ "$ACK_STATUS" = "queued" ] || [ "$ACK_STATUS" = "accepted" ]; then
        echo -e "${GREEN}✓ ACK ${ack_value} webhook sent${NC}"
    else
        echo -e "${YELLOW}⚠ ACK ${ack_value} response: ${ACK_STATUS}${NC}"
    fi

    # Aguardar processamento
    sleep 3

    # Verificar status no banco
    CURRENT_STATUS=$(query_db "SELECT status FROM messages WHERE id = '$MESSAGE_ID';")
    echo -e "${GREEN}✓ Message status: ${CURRENT_STATUS}${NC}"

    echo ""
}

# ACK 1: SERVER (sent)
send_ack 1 "SERVER"
echo -e "${BLUE}→ Check webhook.site for message.sent event${NC}"
echo ""

# ACK 2: DEVICE (delivered)
send_ack 2 "DEVICE"
echo -e "${BLUE}→ Check webhook.site for message.delivered event${NC}"
echo ""

# ACK 3: READ (read)
send_ack 3 "READ"
echo -e "${BLUE}→ Check webhook.site for message.read event${NC}"
echo ""

# ACK 4: PLAYED (played - SOMENTE voice/audio)
send_ack 4 "PLAYED"
echo -e "${BLUE}→ Check webhook.site for message.played event${NC}"
echo ""

# ============================================================================
# STEP 8: Verificar resultados finais
# ============================================================================
echo -e "${BOLD}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BOLD}  Final Results${NC}"
echo -e "${BOLD}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""

FINAL_STATUS=$(query_db "SELECT status, delivered_at IS NOT NULL as delivered, read_at IS NOT NULL as read, played_at IS NOT NULL as played FROM messages WHERE id = '$MESSAGE_ID';")

echo -e "${GREEN}✓ Message final state:${NC}"
echo "$FINAL_STATUS" | sed 's/|/ | /g'
echo ""

echo -e "${BOLD}${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BOLD}${GREEN}  ✓ Test Complete!${NC}"
echo -e "${BOLD}${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""
echo -e "${BLUE}→ Open webhook.site to see all received events:${NC}"
echo -e "  ${WEBHOOK_URL}"
echo ""
echo -e "${YELLOW}Expected webhooks:${NC}"
echo -e "  1. message.sent      (ACK 1)"
echo -e "  2. message.delivered (ACK 2)"
echo -e "  3. message.read      (ACK 3)"
echo -e "  4. message.played    (ACK 4)"
echo ""
echo -e "${YELLOW}Note: Check webhook.site for HTTP 200 responses confirming delivery${NC}"
echo ""

# Opcional: Listar webhooks recebidos do banco
echo -e "${YELLOW}→ Checking webhook delivery logs...${NC}"
WEBHOOK_TRIGGERS=$(query_db "
    SELECT
        triggered_at,
        success,
        http_status
    FROM webhook_triggers
    WHERE webhook_subscription_id = '$WEBHOOK_SUB_ID'
    ORDER BY triggered_at DESC
    LIMIT 10;
")

if [ ! -z "$WEBHOOK_TRIGGERS" ]; then
    echo "$WEBHOOK_TRIGGERS" | sed 's/|/ | /g'
else
    echo -e "${YELLOW}  No webhook triggers found yet (async processing)${NC}"
fi

echo ""
echo -e "${GREEN}✓ API running on PID ${API_PID}${NC}"
echo -e "${YELLOW}→ Press Ctrl+C to stop the API and exit${NC}"
echo ""

# Manter API rodando
wait $API_PID
