#!/bin/bash

# Quick test with SMALL media files to debug why video/audio/document don't arrive

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
CYAN='\033[0;36m'
RESET='\033[0m'

API_BASE="http://localhost:8080/api/v1"
TIMESTAMP=$(date +%s)

# Source .env
if [ -f .env ]; then
    source .env
fi

WAHA_SESSION_ID="${WAHA_DEFAULT_SESSION_ID_TEST:-test-session}"
TEST_PHONE="${TEST_PHONE_NUMBER:-+5511999999999}"

# ⚠️ SMALLER TEST URLS - Fast to download/process - Use widely accessible CDNs
IMAGE_URL="https://picsum.photos/800/600"           # Works - we tested this before
VIDEO_URL="https://commondatastorage.googleapis.com/gtv-videos-bucket/sample/ForBiggerBlazes.mp4"  # 15s Google sample
AUDIO_URL="http://www.kozco.com/tech/piano2-CoolEdit.mp3"  # MP3 from kozco (very reliable, 25KB)
DOCUMENT_URL="https://www.w3.org/WAI/ER/tests/xhtml/testfiles/resources/pdf/dummy.pdf"  # W3C test PDF

echo ""
echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${RESET}"
echo -e "${CYAN}  Quick Media Test - Small Files${RESET}"
echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${RESET}"
echo ""

# Setup (reuse existing user/channel/contact from previous test or create new ones)
echo -e "${CYAN}→ Setting up test environment...${RESET}"

# Register user
REGISTER_RESPONSE=$(curl -s -X POST "$API_BASE/auth/register" \
  -H "Content-Type: application/json" \
  -d "{
    \"name\": \"Small Media Test $TIMESTAMP\",
    \"email\": \"small-media-$TIMESTAMP@test.com\",
    \"password\": \"Test@123456\"
  }")

USER_ID=$(echo $REGISTER_RESPONSE | jq -r '.user_id')
PROJECT_ID=$(echo $REGISTER_RESPONSE | jq -r '.default_project_id')
TOKEN=$(echo $REGISTER_RESPONSE | jq -r '.api_key')

if [ "$USER_ID" == "null" ]; then
    echo -e "${RED}✗ Failed to register user${RESET}"
    exit 1
fi

# Create WAHA channel
WAHA_BASE_URL="${WAHA_BASE_URL:-https://waha.ventros.cloud}"
WAHA_API_KEY="${WAHA_API_KEY:-}"

CHANNEL_RESPONSE=$(curl -s -X POST "$API_BASE/crm/channels?project_id=$PROJECT_ID" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d "{
    \"name\": \"Small Media Test Channel\",
    \"type\": \"waha\",
    \"waha_config\": {
      \"base_url\": \"$WAHA_BASE_URL\",
      \"api_key\": \"$WAHA_API_KEY\",
      \"session_id\": \"$WAHA_SESSION_ID\",
      \"webhook_url\": \"http://localhost:8080/api/v1/webhooks/waha\"
    }
  }")

CHANNEL_ID=$(echo $CHANNEL_RESPONSE | jq -r '.id')

if [ "$CHANNEL_ID" == "null" ]; then
    echo -e "${RED}✗ Failed to create channel${RESET}"
    exit 1
fi

# Activate channel
curl -s -X POST "$API_BASE/crm/channels/$CHANNEL_ID/activate" \
  -H "Authorization: Bearer $TOKEN" > /dev/null

# Create contact
CONTACT_RESPONSE=$(curl -s -X POST "$API_BASE/contacts?project_id=$PROJECT_ID" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d "{
    \"name\": \"Small Media Test Contact\",
    \"phone\": \"$TEST_PHONE\",
    \"email\": \"small-media-$TIMESTAMP@test.com\"
  }")

CONTACT_ID=$(echo $CONTACT_RESPONSE | jq -r '.id')

if [ "$CONTACT_ID" == "null" ]; then
    echo -e "${RED}✗ Failed to create contact${RESET}"
    exit 1
fi

echo -e "${GREEN}✓ Setup complete${RESET}"
echo "  Channel ID: $CHANNEL_ID"
echo "  Contact ID: $CONTACT_ID"
echo ""

# Function to send and test
send_message() {
    local TYPE=$1
    local PAYLOAD=$2
    local DESC=$3

    echo -e "${CYAN}→ Testing: $DESC${RESET}"
    echo "  Sending to WAHA..."

    RESPONSE=$(curl -s -w "\nHTTP_STATUS:%{http_code}" -X POST "$API_BASE/crm/messages/send" \
      -H "Content-Type: application/json" \
      -H "Authorization: Bearer $TOKEN" \
      -d "$PAYLOAD")

    HTTP_STATUS=$(echo "$RESPONSE" | grep "HTTP_STATUS:" | cut -d: -f2)
    BODY=$(echo "$RESPONSE" | sed '/HTTP_STATUS:/d')

    MESSAGE_ID=$(echo $BODY | jq -r '.id // .message_id')

    if [ "$HTTP_STATUS" != "202" ] && [ "$HTTP_STATUS" != "200" ]; then
        echo -e "${RED}✗ HTTP $HTTP_STATUS: $(echo $BODY | jq -r '.error // .message')${RESET}"
        return 1
    elif [ "$MESSAGE_ID" == "null" ] || [ -z "$MESSAGE_ID" ]; then
        echo -e "${RED}✗ No message ID returned${RESET}"
        echo "  Response: $BODY"
        return 1
    else
        echo -e "${GREEN}✓ Sent successfully (ID: $MESSAGE_ID)${RESET}"
        return 0
    fi
}

echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${RESET}"
echo -e "${CYAN}  Testing Small Media Files${RESET}"
echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${RESET}"
echo ""

# 1. Image (baseline - we know this works)
IMAGE_PAYLOAD=$(cat <<EOF
{
  "contact_id": "$CONTACT_ID",
  "channel_id": "$CHANNEL_ID",
  "content_type": "image",
  "media_url": "$IMAGE_URL",
  "text": "📸 Small image test",
  "metadata": {
    "mimetype": "image/jpeg"
  }
}
EOF
)

send_message "image" "$IMAGE_PAYLOAD" "Image (300x200, ~30KB)"
sleep 3
echo ""

# 2. Video (10 seconds, 1MB)
VIDEO_PAYLOAD=$(cat <<EOF
{
  "contact_id": "$CONTACT_ID",
  "channel_id": "$CHANNEL_ID",
  "content_type": "video",
  "media_url": "$VIDEO_URL",
  "text": "🎥 Small video test (10s)",
  "metadata": {
    "mimetype": "video/mp4"
  }
}
EOF
)

send_message "video" "$VIDEO_PAYLOAD" "Video (10s, 1MB)"
sleep 5
echo ""

# 3. Audio (MP3)
AUDIO_PAYLOAD=$(cat <<EOF
{
  "contact_id": "$CONTACT_ID",
  "channel_id": "$CHANNEL_ID",
  "content_type": "audio",
  "media_url": "$AUDIO_URL",
  "metadata": {
    "mimetype": "audio/mpeg"
  }
}
EOF
)

send_message "audio" "$AUDIO_PAYLOAD" "Audio (MP3 format)"
sleep 3
echo ""

# 4. Document (tiny PDF)
DOCUMENT_PAYLOAD=$(cat <<EOF
{
  "contact_id": "$CONTACT_ID",
  "channel_id": "$CHANNEL_ID",
  "content_type": "document",
  "media_url": "$DOCUMENT_URL",
  "text": "📄 Small PDF test",
  "metadata": {
    "mimetype": "application/pdf"
  }
}
EOF
)

send_message "document" "$DOCUMENT_PAYLOAD" "Document (PDF, ~30KB)"
sleep 3
echo ""

echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${RESET}"
echo -e "${GREEN}✓ All messages sent to WAHA${RESET}"
echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${RESET}"
echo ""
echo -e "${CYAN}📱 Check your WhatsApp ($TEST_PHONE) to see which messages actually arrived!${RESET}"
echo ""
