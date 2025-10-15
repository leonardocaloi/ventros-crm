#!/bin/bash

# Script to test message sending with agent

API_BASE="http://localhost:8080/api/v1"
TIMESTAMP=$(date +%s)

echo "=== Testing System Agents Message Send ==="
echo ""

# 1. Register user
echo "1. Registering user..."
REGISTER_RESPONSE=$(curl -s -X POST "$API_BASE/auth/register" \
  -H "Content-Type: application/json" \
  -d "{
    \"name\": \"Test User $TIMESTAMP\",
    \"email\": \"test-$TIMESTAMP@example.com\",
    \"password\": \"Test@123456\"
  }")

echo "Register response: $REGISTER_RESPONSE"
USER_ID=$(echo $REGISTER_RESPONSE | jq -r '.user_id')
PROJECT_ID=$(echo $REGISTER_RESPONSE | jq -r '.default_project_id')
TOKEN=$(echo $REGISTER_RESPONSE | jq -r '.api_key')

echo "User ID: $USER_ID"
echo "Project ID: $PROJECT_ID"
echo "Token: $TOKEN"
echo ""

# 2. Create agent
echo "2. Creating agent for user..."
AGENT_RESPONSE=$(curl -s -X POST "$API_BASE/crm/agents?project_id=$PROJECT_ID" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d "{
    \"project_id\": \"$PROJECT_ID\",
    \"tenant_id\": \"default\",
    \"user_id\": \"$USER_ID\",
    \"name\": \"Test Agent $TIMESTAMP\",
    \"email\": \"agent-$TIMESTAMP@example.com\",
    \"type\": \"human\"
  }")

echo "Agent response: $AGENT_RESPONSE"
AGENT_ID=$(echo $AGENT_RESPONSE | jq -r '.id')
echo "Agent ID: $AGENT_ID"
echo ""

# 3. Create channel
echo "3. Creating WAHA channel..."
CHANNEL_RESPONSE=$(curl -s -X POST "$API_BASE/crm/channels?project_id=$PROJECT_ID" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d "{
    \"name\": \"Test WhatsApp Channel\",
    \"type\": \"waha\",
    \"waha_session_id\": \"guilherme-batilani-suporte\",
    \"config\": {
      \"session_id\": \"guilherme-batilani-suporte\",
      \"base_url\": \"https://waha.ventros.cloud\",
      \"api_key\": \"4bffec302d5f4312b8b73700da3ff3cb\"
    }
  }")

echo "Channel response: $CHANNEL_RESPONSE"
CHANNEL_ID=$(echo $CHANNEL_RESPONSE | jq -r '.id')
echo "Channel ID: $CHANNEL_ID"
echo ""

# 4. Activate channel
echo "4. Activating channel..."
ACTIVATE_RESPONSE=$(curl -s -X POST "$API_BASE/crm/channels/$CHANNEL_ID/activate" \
  -H "Authorization: Bearer $TOKEN")
echo "Activate response: $ACTIVATE_RESPONSE"
echo ""

# 5. Create contact (your WhatsApp number)
echo "5. Creating contact..."
CONTACT_RESPONSE=$(curl -s -X POST "$API_BASE/contacts?project_id=$PROJECT_ID" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d "{
    \"name\": \"Leonardo (Test)\",
    \"phone\": \"+554497044474\",
    \"email\": \"leonardo@test.com\"
  }")

echo "Contact response: $CONTACT_RESPONSE"
CONTACT_ID=$(echo $CONTACT_RESPONSE | jq -r '.id')
echo "Contact ID: $CONTACT_ID"
echo ""

# 6. Send message
echo "6. Sending message..."
MESSAGE_RESPONSE=$(curl -s -X POST "$API_BASE/crm/messages/send" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d "{
    \"channel_id\": \"$CHANNEL_ID\",
    \"contact_id\": \"$CONTACT_ID\",
    \"agent_id\": \"$AGENT_ID\",
    \"source\": \"manual\",
    \"content\": \"Olá! Esta é uma mensagem de teste do sistema com System Agents implementados! 🚀\"
  }")

echo "Message response: $MESSAGE_RESPONSE"
echo ""

echo "=== Test completed ==="
echo "Summary:"
echo "  User ID: $USER_ID"
echo "  Project ID: $PROJECT_ID"
echo "  Agent ID: $AGENT_ID"
echo "  Channel ID: $CHANNEL_ID"
echo "  Contact ID: $CONTACT_ID"
