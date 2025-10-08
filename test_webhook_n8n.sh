#!/bin/bash

# Script para testar webhook N8N com todos os eventos de domínio
# Uso: ./test_webhook_n8n.sh

set -e

API_URL="${API_BASE_URL:-http://localhost:8080}"
WEBHOOK_URL="https://dev.webhook.n8n.ventros.cloud/webhook/ventros/crm/events/all"

echo "🧪 Configurando Webhook N8N para Eventos de Domínio"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "🌐 API URL: $API_URL"
echo "🔗 Webhook URL: $WEBHOOK_URL"
echo ""

# 1. Setup ambiente de teste
echo "1️⃣ Configurando ambiente de teste..."
SETUP_RESPONSE=$(curl -s -X POST "$API_URL/api/v1/test/setup?webhook_url=$WEBHOOK_URL&api_base_url=$API_URL")

# Extrair dados do response
USER_ID=$(echo $SETUP_RESPONSE | jq -r '.data.user_id')
PROJECT_ID=$(echo $SETUP_RESPONSE | jq -r '.data.project_id')
CHANNEL_ID=$(echo $SETUP_RESPONSE | jq -r '.data.channel_id')
CHANNEL_WEBHOOK_URL=$(echo $SETUP_RESPONSE | jq -r '.data.channel_webhook_url')
WEBHOOK_ID=$(echo $SETUP_RESPONSE | jq -r '.data.webhook_id')
API_KEY=$(echo $SETUP_RESPONSE | jq -r '.data.api_key')

echo "✅ User: $USER_ID"
echo "✅ Project: $PROJECT_ID"
echo "✅ Channel: $CHANNEL_ID"
echo "✅ Webhook Subscription: $WEBHOOK_ID"
echo "✅ API Key: ${API_KEY:0:20}..."
echo ""

# 2. Atualizar webhook com TODOS os eventos de domínio e tracking
echo "2️⃣ Atualizando webhook com todos os eventos de domínio..."

# Eventos de domínio (excluindo WAHA pois são internos)
DOMAIN_EVENTS='[
  "contact.created",
  "contact.updated",
  "contact.deleted",
  "contact.merged",
  "contact.enriched",
  "session.started",
  "session.ended",
  "session.message_recorded",
  "session.agent_assigned",
  "session.resolved",
  "session.escalated",
  "session.summarized",
  "session.abandoned",
  "message.created",
  "message.delivered",
  "message.read",
  "message.failed",
  "tracking.message.meta_ads",
  "pipeline.created",
  "pipeline.updated",
  "pipeline.activated",
  "pipeline.deactivated",
  "status.created",
  "status.updated",
  "contact.status_changed",
  "contact.entered_pipeline",
  "contact.exited_pipeline"
]'

UPDATE_PAYLOAD=$(cat <<EOF
{
  "name": "Webhook N8N - Todos Eventos",
  "url": "$WEBHOOK_URL",
  "events": $DOMAIN_EVENTS,
  "active": true,
  "retry_count": 3,
  "timeout_seconds": 30
}
EOF
)

UPDATE_RESPONSE=$(curl -s -X PUT "$API_URL/api/v1/webhook-subscriptions/$WEBHOOK_ID" \
  -H "Authorization: Bearer $API_KEY" \
  -H "Content-Type: application/json" \
  -d "$UPDATE_PAYLOAD")

echo "✅ Webhook atualizado!"
echo ""

# 3. Verificar eventos configurados
echo "3️⃣ Verificando eventos configurados..."
WEBHOOK_INFO=$(curl -s -X GET "$API_URL/api/v1/webhook-subscriptions/$WEBHOOK_ID" \
  -H "Authorization: Bearer $API_KEY")

echo "📋 Eventos ativos:"
echo $WEBHOOK_INFO | jq -r '.webhook.events[] | "   ✓ \(.)"'
echo ""

# 4. Enviar mensagem de teste para gerar eventos
echo "4️⃣ Enviando mensagem de teste para gerar eventos..."
TEST_MESSAGE='{"id":"evt_test_n8n","timestamp":1696598400000,"event":"message","session":"test-session-waha","payload":{"id":"msg_test_001","from":"5511999999999@c.us","fromMe":false,"body":"Teste de webhook N8N - Todos os eventos de domínio configurados!","_data":{"Info":{"PushName":"Cliente Teste"}}}}'

WEBHOOK_RESPONSE=$(curl -s -X POST "$CHANNEL_WEBHOOK_URL" \
  -H "Content-Type: application/json" \
  -d "$TEST_MESSAGE")

echo "✅ Mensagem enviada!"
echo ""

# 5. Verificar canal
echo "5️⃣ Verificando canal..."
CHANNEL_INFO=$(curl -s -X GET "$API_URL/api/v1/channels/$CHANNEL_ID" \
  -H "Authorization: Bearer $API_KEY")

echo $CHANNEL_INFO | jq '.channel | {id, name, type, webhook_url, webhook_active, messages_received}'
echo ""

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "✅ Webhook N8N configurado com sucesso!"
echo ""
echo "📤 Eventos que serão enviados para N8N:"
echo "   • Contatos: created, updated, deleted, merged, enriched"
echo "   • Sessões: started, ended, message_recorded, agent_assigned, resolved, escalated, summarized, abandoned"
echo "   • Mensagens: created, delivered, read, failed"
echo "   • Tracking: tracking.message.meta_ads (Meta Ads: FB/Instagram)"
echo "   • Pipelines: created, updated, activated, deactivated"
echo "   • Status: created, updated, contact.status_changed, contact.entered_pipeline, contact.exited_pipeline"
echo ""
echo "🔗 Webhook URL: $WEBHOOK_URL"
echo "📋 Webhook ID: $WEBHOOK_ID"
echo "🔑 API Key: $API_KEY"
echo ""
echo "💡 Para testar, envie mensagens para o canal ou use:"
echo "   curl -X POST \"$CHANNEL_WEBHOOK_URL\" \\"
echo "     -H \"Content-Type: application/json\" \\"
echo "     -d @events_waha/message_text.json"
echo ""
echo "🌐 Verifique os eventos em: $WEBHOOK_URL"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
