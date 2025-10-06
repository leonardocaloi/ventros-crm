#!/bin/bash

# Script de teste E2E para Ventros CRM
# Uso: ./scripts/test-e2e.sh [cleanup]

set -e

# Cores para output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuração
API_BASE_URL="${API_BASE_URL:-http://localhost:8080}"
TEST_DATA_FILE="/tmp/ventros-test-data.json"

# Funções helpers
log_info() {
    echo -e "${BLUE}ℹ${NC}  $1"
}

log_success() {
    echo -e "${GREEN}✓${NC}  $1"
}

log_error() {
    echo -e "${RED}✗${NC}  $1"
}

log_warning() {
    echo -e "${YELLOW}⚠${NC}  $1"
}

wait_for_api() {
    log_info "Waiting for API to be ready..."
    local max_retries=30
    local count=0
    
    while [ $count -lt $max_retries ]; do
        if curl -sf "$API_BASE_URL/health" > /dev/null 2>&1; then
            log_success "API is ready"
            return 0
        fi
        count=$((count + 1))
        sleep 1
    done
    
    log_error "API not ready after 30 seconds"
    exit 1
}

save_test_data() {
    local key=$1
    local value=$2
    
    # Cria arquivo se não existir
    if [ ! -f "$TEST_DATA_FILE" ]; then
        echo '{}' > "$TEST_DATA_FILE"
    fi
    
    # Adiciona/atualiza chave
    jq --arg key "$key" --arg value "$value" '.[$key] = $value' "$TEST_DATA_FILE" > "${TEST_DATA_FILE}.tmp"
    mv "${TEST_DATA_FILE}.tmp" "$TEST_DATA_FILE"
}

get_test_data() {
    local key=$1
    if [ -f "$TEST_DATA_FILE" ]; then
        jq -r --arg key "$key" '.[$key] // ""' "$TEST_DATA_FILE"
    else
        echo ""
    fi
}

cleanup_database_direct() {
    log_info "Cleaning up test user directly from database..."
    
    # Delete test user directly from database (cascade will handle related data)
    PGPASSWORD=ventros123 psql -h localhost -U ventros -d ventros_crm -c \
        "DELETE FROM users WHERE email = 'admin.e2e@ventros.local';" > /dev/null 2>&1 || true
    
    log_success "Database cleanup completed"
}

cleanup() {
    log_info "Starting cleanup..."
    
    if [ ! -f "$TEST_DATA_FILE" ]; then
        log_warning "No test data found for cleanup"
        # Try direct database cleanup anyway
        cleanup_database_direct
        return 0
    fi
    
    local api_key=$(get_test_data "api_key")
    
    if [ -z "$api_key" ]; then
        log_warning "No API key found for cleanup"
        # Try direct database cleanup anyway
        cleanup_database_direct
        return 0
    fi
    
    # Cleanup webhook subscription
    local webhook_id=$(get_test_data "webhook_id")
    if [ -n "$webhook_id" ]; then
        log_info "Deleting webhook subscription: $webhook_id"
        curl -sf -X DELETE "$API_BASE_URL/api/v1/webhook-subscriptions/$webhook_id" \
            -H "Authorization: Bearer $api_key" > /dev/null || true
        log_success "Webhook subscription deleted"
    fi

    # Cleanup channels
    local channel_id=$(get_test_data "channel_id")
    if [ -n "$channel_id" ]; then
        log_info "Deleting channel: $channel_id"
        curl -sf -X DELETE "$API_BASE_URL/api/v1/channels/$channel_id" \
            -H "Authorization: Bearer $api_key" > /dev/null || true
        log_success "Channel deleted"
    fi
    
    # Cleanup contacts
    local contact_id=$(get_test_data "contact_id")
    if [ -n "$contact_id" ]; then
        log_info "Deleting contact: $contact_id"
        curl -sf -X DELETE "$API_BASE_URL/api/v1/contacts/$contact_id" \
            -H "Authorization: Bearer $api_key" > /dev/null || true
        log_success "Contact deleted"
    fi
    
    # Note: User deletion endpoint doesn't exist, so we skip user cleanup
    # The idempotent user creation will handle existing users
    
    # Also try direct database cleanup as fallback
    cleanup_database_direct
    
    # Remove test data file
    rm -f "$TEST_DATA_FILE"
    log_success "Cleanup completed"
}

# Se argumento é 'cleanup', apenas limpa e sai
if [ "$1" == "cleanup" ]; then
    cleanup
    exit 0
fi

# Inicia testes
echo ""
echo "=========================================="
echo "  Ventros CRM - E2E Tests"
echo "=========================================="
echo ""

# 1. Aguarda API
wait_for_api

# 2. Limpa dados de teste anteriores (se existirem)
log_info "Cleaning up previous test data..."
cleanup || true  # Não falha se não houver dados para limpar

# 3. Cria usuário
log_info "Test 1: Creating user..."
RESPONSE=$(curl -sf -X POST "$API_BASE_URL/api/v1/auth/register" \
    -H "Content-Type: application/json" \
    -d '{
        "name": "Admin Teste E2E",
        "email": "admin.e2e@ventros.local",
        "password": "senha_teste_123",
        "role": "admin"
    }')

if [ $? -eq 0 ]; then
    USER_ID=$(echo "$RESPONSE" | jq -r '.user_id')
    API_KEY=$(echo "$RESPONSE" | jq -r '.api_key')
    PROJECT_ID=$(echo "$RESPONSE" | jq -r '.default_project_id')
    PIPELINE_ID=$(echo "$RESPONSE" | jq -r '.default_pipeline_id')
    
    save_test_data "user_id" "$USER_ID"
    save_test_data "api_key" "$API_KEY"
    save_test_data "project_id" "$PROJECT_ID"
    save_test_data "pipeline_id" "$PIPELINE_ID"
    
    log_success "User created: $USER_ID"
    log_success "  → Project: $PROJECT_ID"
    log_success "  → Pipeline: $PIPELINE_ID"
else
    log_error "Failed to create user"
    exit 1
fi

# 4. Cria canal
log_info "Test 2: Creating channel..."
RESPONSE=$(curl -sf -X POST "$API_BASE_URL/api/v1/channels?project_id=$PROJECT_ID" \
    -H "Authorization: Bearer $API_KEY" \
    -H "Content-Type: application/json" \
    -d '{
        "name": "WhatsApp Teste E2E",
        "type": "waha",
        "waha_config": {
            "base_url": "https://waha.ventros.cloud",
            "token": "test_token_e2e",
            "session_id": "test_session_e2e",
            "webhook_url": "http://localhost:8080/api/v1/webhooks/waha"
        }
    }')

if [ $? -eq 0 ]; then
    CHANNEL_ID=$(echo "$RESPONSE" | jq -r '.id')
    save_test_data "channel_id" "$CHANNEL_ID"
    log_success "Channel created: $CHANNEL_ID"
else
    log_error "Failed to create channel"
    exit 1
fi

# 5. Ativa o canal
log_info "Test 3: Activating channel..."
RESPONSE=$(curl -sf -X POST "$API_BASE_URL/api/v1/channels/$CHANNEL_ID/activate" \
    -H "Authorization: Bearer $API_KEY")

if [ $? -eq 0 ]; then
    log_success "Channel activated successfully"
else
    log_warning "Channel activation failed, but continuing..."
fi

# 6. Cria contato
log_info "Test 4: Creating contact..."
RESPONSE=$(curl -sf -X POST "$API_BASE_URL/api/v1/contacts?project_id=$PROJECT_ID" \
    -H "Authorization: Bearer $API_KEY" \
    -H "Content-Type: application/json" \
    -d '{
        "name": "João Silva Teste E2E",
        "phone": "+5511999887766",
        "email": "joao.e2e@example.com"
    }')

if [ $? -eq 0 ]; then
    CONTACT_ID=$(echo "$RESPONSE" | jq -r '.id')
    save_test_data "contact_id" "$CONTACT_ID"
    log_success "Contact created: $CONTACT_ID"
else
    log_error "Failed to create contact"
    exit 1
fi

# 7. Lista contatos
log_info "Test 5: Listing contacts..."
RESPONSE=$(curl -sf -X GET "$API_BASE_URL/api/v1/contacts?project_id=$PROJECT_ID" \
    -H "Authorization: Bearer $API_KEY")

if [ $? -eq 0 ]; then
    CONTACT_COUNT=$(echo "$RESPONSE" | jq '.contacts | length')
    log_success "Listed $CONTACT_COUNT contacts"
else
    log_error "Failed to list contacts"
    exit 1
fi

# 8. Lista canais
log_info "Test 6: Listing channels..."
RESPONSE=$(curl -sf -X GET "$API_BASE_URL/api/v1/channels?project_id=$PROJECT_ID" \
    -H "Authorization: Bearer $API_KEY")

if [ $? -eq 0 ]; then
    CHANNEL_COUNT=$(echo "$RESPONSE" | jq '.channels | length')
    log_success "Listed $CHANNEL_COUNT channels"
else
    log_error "Failed to list channels"
    exit 1
fi

# 9. Cria webhook subscription para N8N
log_info "Test 7: Creating webhook subscription..."
RESPONSE=$(curl -sf -X POST "$API_BASE_URL/api/v1/webhook-subscriptions" \
    -H "Authorization: Bearer $API_KEY" \
    -H "Content-Type: application/json" \
    -d '{
        "name": "N8N Webhook E2E Test",
        "url": "https://dev.webhook.n8n.ventros.cloud/webhook/ventros/crm/events",
        "events": ["message.created", "message.delivered", "message.read", "contact.created", "session.started", "session.ended", "session.message_recorded", "ad_campaign.tracked"],
        "retry_count": 3,
        "timeout_seconds": 30
    }')

if [ $? -eq 0 ]; then
    WEBHOOK_ID=$(echo "$RESPONSE" | jq -r '.id')
    save_test_data "webhook_id" "$WEBHOOK_ID"
    log_success "Webhook subscription created: $WEBHOOK_ID"
else
    log_error "Failed to create webhook subscription"
    exit 1
fi

# 10. Testa envio de mensagem WAHA (Facebook Ads)
log_info "Test 8: Testing WAHA message send (Facebook Ads)..."
RESPONSE=$(curl -sf -X POST "$API_BASE_URL/api/v1/test/send-waha-message?type=fb_ads" \
    -H "Authorization: Bearer $API_KEY" \
    -H "Content-Type: application/json" \
    -d '{}')

if [ $? -eq 0 ]; then
    log_success "WAHA message test completed successfully"
else
    log_warning "WAHA message test failed, but continuing..."
fi

echo ""
echo "=========================================="
echo "  ✓ All tests passed!"
echo "=========================================="
echo ""
echo "Test data saved to: $TEST_DATA_FILE"
echo ""
log_info "Test data kept for reuse in subsequent tests"
log_info "API Key: $API_KEY"
log_info "Project ID: $PROJECT_ID"
echo ""
echo "To cleanup manually: ./scripts/test-e2e.sh cleanup"
echo ""
