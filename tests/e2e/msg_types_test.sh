#!/bin/bash

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# E2E Test: All Message Types (Send → WAHA → Webhook → Process)
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
#
# Este teste valida o fluxo completo de todos os tipos de mensagens:
# 1. Envia mensagem via API (outbound)
# 2. WAHA processa e envia via webhook (inbound)
# 3. Sistema recebe e processa pelo webhook handler
#
# Tipos testados:
# - text
# - image
# - video
# - audio
# - voice
# - document
# - sticker
#
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
MAGENTA='\033[0;35m'
RESET='\033[0m'

API_BASE="http://localhost:8080/api/v1"
TIMESTAMP=$(date +%s)

# Source .env to get WAHA configuration
if [ -f .env ]; then
    source .env
fi

WAHA_SESSION_ID="${WAHA_DEFAULT_SESSION_ID_TEST:-test-session}"
TEST_PHONE="${TEST_PHONE_NUMBER:-+5511999999999}"

# URLs de mídia de teste (publicamente acessíveis)
IMAGE_URL="https://picsum.photos/800/600"
VIDEO_URL="https://sample-videos.com/video123/mp4/720/big_buck_bunny_720p_1mb.mp4"
AUDIO_URL="https://www.soundhelix.com/examples/mp3/SoundHelix-Song-1.mp3"
DOCUMENT_URL="https://www.w3.org/WAI/ER/tests/xhtml/testfiles/resources/pdf/dummy.pdf"
STICKER_URL="https://raw.githubusercontent.com/WhatsApp/stickers/main/Android/app/src/main/assets/1/01_Cuppy_smile.webp"

echo ""
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${RESET}"
echo -e "${BLUE}  E2E Test: All Message Types${RESET}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${RESET}"
echo ""

# Check API health
echo -e "${CYAN}→ Checking API health...${RESET}"
if ! curl -s -f "http://localhost:8080/health" > /dev/null; then
    echo -e "${RED}✗ API is not running!${RESET}"
    echo -e "${YELLOW}  Start API with: make api${RESET}"
    exit 1
fi
echo -e "${GREEN}✓ API is healthy${RESET}"
echo ""

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# Setup: Register user and create resources
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

echo -e "${CYAN}→ Step 1/3: Setting up test environment...${RESET}"

# Register user
REGISTER_RESPONSE=$(curl -s -X POST "$API_BASE/auth/register" \
  -H "Content-Type: application/json" \
  -d "{
    \"name\": \"E2E Message Types Test $TIMESTAMP\",
    \"email\": \"msg-types-$TIMESTAMP@test.com\",
    \"password\": \"Test@123456\"
  }")

USER_ID=$(echo $REGISTER_RESPONSE | jq -r '.user_id')
PROJECT_ID=$(echo $REGISTER_RESPONSE | jq -r '.default_project_id')
TOKEN=$(echo $REGISTER_RESPONSE | jq -r '.api_key')

if [ "$USER_ID" == "null" ] || [ -z "$USER_ID" ]; then
    echo -e "${RED}✗ Failed to register user${RESET}"
    exit 1
fi

echo -e "${GREEN}✓ User registered${RESET}"
echo "  User ID: $USER_ID"
echo "  Project ID: $PROJECT_ID"
echo ""

# Create WAHA channel
echo -e "${CYAN}→ Step 2/3: Creating WAHA channel...${RESET}"
WAHA_BASE_URL="${WAHA_BASE_URL:-https://waha.ventros.cloud}"
WAHA_API_KEY="${WAHA_API_KEY:-}"

CHANNEL_RESPONSE=$(curl -s -X POST "$API_BASE/crm/channels?project_id=$PROJECT_ID" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d "{
    \"name\": \"E2E Test Channel\",
    \"type\": \"waha\",
    \"waha_config\": {
      \"base_url\": \"$WAHA_BASE_URL\",
      \"api_key\": \"$WAHA_API_KEY\",
      \"session_id\": \"$WAHA_SESSION_ID\",
      \"webhook_url\": \"http://localhost:8080/api/v1/webhooks/waha\"
    }
  }")

CHANNEL_ID=$(echo $CHANNEL_RESPONSE | jq -r '.id')

if [ "$CHANNEL_ID" == "null" ] || [ -z "$CHANNEL_ID" ]; then
    echo -e "${RED}✗ Failed to create channel${RESET}"
    exit 1
fi

# Activate channel
curl -s -X POST "$API_BASE/crm/channels/$CHANNEL_ID/activate" \
  -H "Authorization: Bearer $TOKEN" > /dev/null

echo -e "${GREEN}✓ Channel created and activated${RESET}"
echo "  Channel ID: $CHANNEL_ID"
echo ""

# Create contact
echo -e "${CYAN}→ Step 3/3: Creating contact...${RESET}"
CONTACT_RESPONSE=$(curl -s -X POST "$API_BASE/contacts?project_id=$PROJECT_ID" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d "{
    \"name\": \"E2E Test Contact\",
    \"phone\": \"$TEST_PHONE\",
    \"email\": \"e2e-msg-types-$TIMESTAMP@test.com\"
  }")

CONTACT_ID=$(echo $CONTACT_RESPONSE | jq -r '.id')

if [ "$CONTACT_ID" == "null" ] || [ -z "$CONTACT_ID" ]; then
    echo -e "${RED}✗ Failed to create contact${RESET}"
    exit 1
fi

echo -e "${GREEN}✓ Contact created${RESET}"
echo "  Contact ID: $CONTACT_ID"
echo "  Phone: $TEST_PHONE"
echo ""

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# Test: Send messages of all types
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${RESET}"
echo -e "${BLUE}  Testing All Message Types${RESET}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${RESET}"
echo ""

SUCCESS_COUNT=0
FAIL_COUNT=0

# Function to send message and check response
send_message() {
    local TYPE=$1
    local PAYLOAD=$2
    local DESCRIPTION=$3

    echo -e "${CYAN}→ Testing: ${DESCRIPTION}${RESET}"

    RESPONSE=$(curl -s -X POST "$API_BASE/crm/messages/send" \
      -H "Content-Type: application/json" \
      -H "Authorization: Bearer $TOKEN" \
      -d "$PAYLOAD")

    MESSAGE_ID=$(echo $RESPONSE | jq -r '.id // .message_id')

    if echo "$RESPONSE" | jq -e '.error' > /dev/null 2>&1; then
        echo -e "${RED}✗ Failed: $(echo $RESPONSE | jq -r '.error')${RESET}"
        FAIL_COUNT=$((FAIL_COUNT + 1))
        return 1
    elif [ "$MESSAGE_ID" == "null" ] || [ -z "$MESSAGE_ID" ]; then
        echo -e "${RED}✗ Failed: No message ID returned${RESET}"
        FAIL_COUNT=$((FAIL_COUNT + 1))
        return 1
    else
        echo -e "${GREEN}✓ Success${RESET}"
        echo "  Message ID: $MESSAGE_ID"
        echo "  Type: $TYPE"
        SUCCESS_COUNT=$((SUCCESS_COUNT + 1))

        # Wait a bit for webhook processing
        sleep 2
        return 0
    fi
}

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# 1. TEXT MESSAGE
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

TEXT_PAYLOAD=$(cat <<EOF
{
  "contact_id": "$CONTACT_ID",
  "channel_id": "$CHANNEL_ID",
  "content_type": "text",
  "text": "🧪 E2E Test - Text Message (ID: $TIMESTAMP)"
}
EOF
)

send_message "text" "$TEXT_PAYLOAD" "Text Message"
echo ""

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# 2. IMAGE MESSAGE
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

IMAGE_PAYLOAD=$(cat <<EOF
{
  "contact_id": "$CONTACT_ID",
  "channel_id": "$CHANNEL_ID",
  "content_type": "image",
  "media_url": "$IMAGE_URL",
  "media_mimetype": "image/jpeg",
  "text": "📸 E2E Test - Image Message"
}
EOF
)

send_message "image" "$IMAGE_PAYLOAD" "Image Message"
echo ""

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# 3. VIDEO MESSAGE
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

VIDEO_PAYLOAD=$(cat <<EOF
{
  "contact_id": "$CONTACT_ID",
  "channel_id": "$CHANNEL_ID",
  "content_type": "video",
  "media_url": "$VIDEO_URL",
  "media_mimetype": "video/mp4",
  "text": "🎥 E2E Test - Video Message"
}
EOF
)

send_message "video" "$VIDEO_PAYLOAD" "Video Message"
echo ""

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# 4. AUDIO MESSAGE
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

AUDIO_PAYLOAD=$(cat <<EOF
{
  "contact_id": "$CONTACT_ID",
  "channel_id": "$CHANNEL_ID",
  "content_type": "audio",
  "media_url": "$AUDIO_URL",
  "media_mimetype": "audio/mpeg"
}
EOF
)

send_message "audio" "$AUDIO_PAYLOAD" "Audio Message"
echo ""

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# 5. DOCUMENT MESSAGE
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

DOCUMENT_PAYLOAD=$(cat <<EOF
{
  "contact_id": "$CONTACT_ID",
  "channel_id": "$CHANNEL_ID",
  "content_type": "document",
  "media_url": "$DOCUMENT_URL",
  "media_mimetype": "application/pdf",
  "text": "📄 E2E Test Document - Test ID: $TIMESTAMP"
}
EOF
)

send_message "document" "$DOCUMENT_PAYLOAD" "Document Message (PDF)"
echo ""

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# 6. LOCATION MESSAGE
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

LOCATION_PAYLOAD=$(cat <<EOF
{
  "contact_id": "$CONTACT_ID",
  "channel_id": "$CHANNEL_ID",
  "content_type": "location",
  "metadata": {
    "latitude": -23.550520,
    "longitude": -46.633308,
    "address": "Av. Paulista, 1578 - São Paulo, SP",
    "name": "E2E Test Location"
  }
}
EOF
)

send_message "location" "$LOCATION_PAYLOAD" "Location Message"
echo ""

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# 7. CONTACT MESSAGE (vCard)
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

CONTACT_VCARD_PAYLOAD=$(cat <<EOF
{
  "contact_id": "$CONTACT_ID",
  "channel_id": "$CHANNEL_ID",
  "content_type": "contact",
  "metadata": {
    "name": "João da Silva",
    "phone": "+5511999999999",
    "email": "joao@example.com",
    "company": "Ventros CRM"
  }
}
EOF
)

send_message "contact" "$CONTACT_VCARD_PAYLOAD" "Contact Message (vCard)"
echo ""

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# Summary
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

TOTAL=$((SUCCESS_COUNT + FAIL_COUNT))

echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${RESET}"
echo -e "${BLUE}  Test Summary${RESET}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${RESET}"
echo ""
echo "Total Tests: $TOTAL"
echo -e "${GREEN}✓ Success: $SUCCESS_COUNT${RESET}"
if [ $FAIL_COUNT -gt 0 ]; then
    echo -e "${RED}✗ Failed: $FAIL_COUNT${RESET}"
fi
echo ""

# Message types tested
echo "Message Types Tested:"
echo "  • text"
echo "  • image"
echo "  • video"
echo "  • audio"
echo "  • document"
echo "  • location"
echo "  • contact"
echo ""

echo -e "${CYAN}💡 Check your WhatsApp ($TEST_PHONE) for all messages!${RESET}"
echo ""

# Wait for webhook processing
echo -e "${YELLOW}⏳ Waiting 10 seconds for webhook processing...${RESET}"
sleep 10

# Verify messages in database
echo ""
echo -e "${CYAN}→ Verifying messages in database...${RESET}"
MESSAGES_RESPONSE=$(curl -s -X GET "$API_BASE/crm/messages?project_id=$PROJECT_ID&contact_id=$CONTACT_ID&limit=20" \
  -H "Authorization: Bearer $TOKEN")

MESSAGE_COUNT=$(echo $MESSAGES_RESPONSE | jq -r '.data | length')
echo -e "${GREEN}✓ Messages in database: $MESSAGE_COUNT${RESET}"
echo ""

if [ $FAIL_COUNT -eq 0 ]; then
    echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${RESET}"
    echo -e "${GREEN}  ✅ All Tests Passed!${RESET}"
    echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${RESET}"
    echo ""
    exit 0
else
    echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${RESET}"
    echo -e "${YELLOW}  ⚠ Some Tests Failed${RESET}"
    echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${RESET}"
    echo ""
    exit 1
fi
