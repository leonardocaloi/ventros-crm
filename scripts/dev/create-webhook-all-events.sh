#!/bin/bash

# Script para criar webhook que recebe TODOS os eventos EXCETO message.*
# URL: https://dev.webhook.n8n.ventros.cloud/webhook/ventros/crm/events/all

set -e

API_URL="${API_URL:-http://localhost:8080}"
WEBHOOK_URL="https://dev.webhook.n8n.ventros.cloud/webhook/ventros/crm/events/all"

echo "🚀 Criando webhook para TODOS os eventos (exceto message.*)"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# 1. Login
echo "1️⃣ Fazendo login..."
LOGIN_RESPONSE=$(curl -s -X POST "$API_URL/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123"
  }')

TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.api_key // .token')
USER_ID=$(echo "$LOGIN_RESPONSE" | jq -r '.user_id // .user.id')
PROJECT_ID=$(echo "$LOGIN_RESPONSE" | jq -r '.default_project_id // .user.projects[0].id')

if [ "$TOKEN" = "null" ] || [ -z "$TOKEN" ]; then
  echo "❌ Erro no login"
  echo "$LOGIN_RESPONSE" | jq '.'
  exit 1
fi

echo "✅ Login OK (User: $USER_ID, Project: $PROJECT_ID)"

# 2. Deletar webhook antigo se existir
echo ""
echo "2️⃣ Deletando webhooks antigos para esta URL..."
EXISTING_WEBHOOKS=$(curl -s "$API_URL/api/v1/webhook-subscriptions" \
  -H "Authorization: Bearer $TOKEN" | jq -r '.data[]? | select(.url == "'"$WEBHOOK_URL"'") | .id')

if [ -n "$EXISTING_WEBHOOKS" ]; then
  for WEBHOOK_ID in $EXISTING_WEBHOOKS; do
    echo "   Deletando webhook: $WEBHOOK_ID"
    curl -s -X DELETE "$API_URL/api/v1/webhook-subscriptions/$WEBHOOK_ID" \
      -H "Authorization: Bearer $TOKEN" > /dev/null
  done
  echo "✅ Webhooks antigos deletados"
else
  echo "ℹ️  Nenhum webhook antigo encontrado"
fi

# 3. Criar webhook com wildcards para TODOS os eventos EXCETO message.*
echo ""
echo "3️⃣ Criando webhook com wildcards..."

WEBHOOK_RESPONSE=$(curl -s -X POST "$API_URL/api/v1/webhook-subscriptions" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "N8N - All Events (Exceto Messages)",
    "url": "'"$WEBHOOK_URL"'",
    "events": [
      "contact.*",
      "session.*",
      "tracking.*",
      "note.*",
      "agent.*",
      "channel.*",
      "pipeline.*",
      "billing.*",
      "project.*",
      "webhook.*",
      "credential.*",
      "automation.*"
    ],
    "active": true,
    "secret": "",
    "retry_count": 3,
    "timeout_seconds": 30
  }')

WEBHOOK_ID=$(echo "$WEBHOOK_RESPONSE" | jq -r '.data.id // .id // empty')

if [ -z "$WEBHOOK_ID" ] || [ "$WEBHOOK_ID" = "null" ]; then
  echo "❌ Erro ao criar webhook"
  echo "$WEBHOOK_RESPONSE" | jq '.'
  exit 1
fi

echo "✅ Webhook criado: $WEBHOOK_ID"
echo ""
echo "📋 Eventos inscritos (wildcards):"
echo "   • contact.* (created, updated, status_changed, entered_pipeline, exited_pipeline, etc.)"
echo "   • session.* (created, closed, agent_assigned, resolved, escalated, etc.)"
echo "   • tracking.* (message.meta_ads, created, enriched)"
echo "   • note.* (added, updated, deleted, pinned)"
echo "   • agent.* (created, updated, activated, deactivated)"
echo "   • channel.* (created, activated, deactivated, deleted)"
echo "   • pipeline.* (created, updated, status.created, status.updated, activated, deactivated)"
echo "   • billing.* (account_created, limit_reached, etc.)"
echo "   • project.* (created, updated, deleted)"
echo "   • webhook.* (created, updated, deleted)"
echo "   • credential.* (created, updated, deleted)"
echo "   • automation.* (rule_created, rule_triggered, etc.)"
echo ""
echo "🚫 EXCLUÍDOS:"
echo "   • message.* (received, sent, delivered, read, failed)"
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "✅ Webhook configurado com sucesso!"
echo ""
echo "🔗 URL: $WEBHOOK_URL"
echo "🆔 ID: $WEBHOOK_ID"
echo "📊 Status: Ativo"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
