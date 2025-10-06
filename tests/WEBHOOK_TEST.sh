#!/bin/bash

# Teste completo de webhooks por canal
# Execute após autenticar e ter um canal criado

BASE_URL="http://localhost:8080"
TOKEN="your-auth-token-here"
CHANNEL_ID="your-channel-id-here"

echo "🔧 TESTE DE WEBHOOKS POR CANAL"
echo "================================"
echo ""

# 1. Obter URL do webhook
echo "1️⃣ Obtendo URL do webhook..."
curl -X GET "$BASE_URL/api/v1/channels/$CHANNEL_ID/webhook-url" \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-Base-URL: https://api.ventros.com" | jq

echo -e "\n"

# 2. Obter informações do webhook
echo "2️⃣ Obtendo informações do webhook..."
curl -X GET "$BASE_URL/api/v1/channels/$CHANNEL_ID/webhook-info" \
  -H "Authorization: Bearer $TOKEN" | jq

echo -e "\n"

# 3. Configurar webhook automaticamente (WAHA)
echo "3️⃣ Configurando webhook automaticamente..."
curl -X POST "$BASE_URL/api/v1/channels/$CHANNEL_ID/configure-webhook" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "base_url": "https://api.ventros.com"
  }' | jq

echo -e "\n"

# 4. Verificar canal atualizado
echo "4️⃣ Verificando canal atualizado..."
curl -X GET "$BASE_URL/api/v1/channels/$CHANNEL_ID" \
  -H "Authorization: Bearer $TOKEN" | jq '.channel | {
    id,
    name,
    type,
    webhook_url,
    webhook_configured_at,
    webhook_active
  }'

echo -e "\n"

# 5. Simular webhook recebendo mensagem
echo "5️⃣ Simulando mensagem no webhook..."
WEBHOOK_URL=$(curl -s -X GET "$BASE_URL/api/v1/channels/$CHANNEL_ID/webhook-url" \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-Base-URL: http://localhost:8080" | jq -r '.webhook_url')

echo "Webhook URL: $WEBHOOK_URL"

curl -X POST "$WEBHOOK_URL" \
  -H "Content-Type: application/json" \
  -d '{
    "id": "evt_test_001",
    "timestamp": 1696598400000,
    "event": "message",
    "session": "test_session",
    "payload": {
      "id": "msg_001",
      "from": "5511999999999@c.us",
      "fromMe": false,
      "body": "Teste de webhook por canal!",
      "_data": {
        "Info": {
          "PushName": "Teste User"
        },
        "Message": {
          "conversation": "Teste de webhook por canal!"
        }
      }
    }
  }' | jq

echo -e "\n✅ Teste completo!"
