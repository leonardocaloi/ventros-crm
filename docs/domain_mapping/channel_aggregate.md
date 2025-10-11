# Channel Aggregate

**Last Updated**: 2025-10-10
**Status**: ✅ Complete and Production-Ready
**Lines of Code**: ~650
**Test Coverage**: Partial

---

## Overview

- **Purpose**: Manages communication channels (WhatsApp, Telegram, etc.)
- **Location**: `internal/domain/channel/`
- **Entity**: `infrastructure/persistence/entities/channel.go`
- **Repository**: `infrastructure/persistence/gorm_channel_repository.go`
- **Aggregate Root**: `Channel`

**Business Problem**:
The Channel aggregate represents **external communication platforms** that connect to the CRM. Channels are the entry/exit points for messages - they receive incoming messages from customers and send outbound messages. Critical for:
- **Multi-channel support** - WhatsApp, Telegram, Facebook Messenger, Instagram
- **Message routing** - Route incoming messages to correct pipelines/agents
- **Webhook management** - Auto-configure webhooks in external platforms
- **Session management** - Maintain connection state (QR codes, tokens)
- **History import** - Bulk import existing conversations
- **AI integration** - Enable AI agents per channel

---

## Domain Model

### Aggregate Root: Channel

```go
type Channel struct {
    id                   uuid.UUID
    userID               uuid.UUID
    projectID            uuid.UUID
    tenantID             string
    name                 string
    channelType          ChannelType    // waha, whatsapp, telegram, messenger, instagram
    status               ChannelStatus  // active, inactive, connecting, disconnected, error
    externalID           string         // External platform ID
    config               map[string]interface{}  // Channel-specific config

    // Webhook management
    webhookID            string
    webhookURL           string
    webhookConfiguredAt  *time.Time
    webhookActive        bool

    // Pipeline integration
    pipelineID                   *uuid.UUID  // Default pipeline for this channel
    defaultSessionTimeoutMinutes int         // Override global timeout

    // AI features
    aiEnabled            bool  // Enable AI responses
    aiAgentsEnabled      bool  // Enable AI agent assignment

    // Metrics
    messagesReceived     int
    messagesSent         int
    lastMessageAt        *time.Time
    lastErrorAt          *time.Time
    lastError            string

    createdAt            time.Time
    updatedAt            time.Time
}
```

### Value Objects

#### 1. ChannelType

```go
type ChannelType string
const (
    ChannelTypeWAHA       ChannelType = "waha"       // WAHA (WhatsApp wrapper)
    ChannelTypeWhatsApp   ChannelType = "whatsapp"   // WhatsApp Business API
    ChannelTypeTelegram   ChannelType = "telegram"   // Telegram Bot API
    ChannelTypeMessenger  ChannelType = "messenger"  // Facebook Messenger
    ChannelTypeInstagram  ChannelType = "instagram"  // Instagram DMs
)
```

#### 2. ChannelStatus

```go
type ChannelStatus string
const (
    ChannelStatusActive       ChannelStatus = "active"        // Connected and working
    ChannelStatusInactive     ChannelStatus = "inactive"      // Disabled by user
    ChannelStatusConnecting   ChannelStatus = "connecting"    // Initial setup
    ChannelStatusDisconnected ChannelStatus = "disconnected"  // Lost connection
    ChannelStatusError        ChannelStatus = "error"         // Configuration error
)
```

#### 3. WAHA-Specific Types

```go
// WAHA Session Status (from external API)
type WAHASessionStatus string
const (
    WAHASessionStatusStarting     WAHASessionStatus = "STARTING"
    WAHASessionStatusScanQR       WAHASessionStatus = "SCAN_QR_CODE"
    WAHASessionStatusWorking      WAHASessionStatus = "WORKING"
    WAHASessionStatusFailed       WAHASessionStatus = "FAILED"
    WAHASessionStatusStopped      WAHASessionStatus = "STOPPED"
    WAHASessionStatusUnauthorized WAHASessionStatus = "UNAUTHORIZED"
)

// WAHA Import Strategy
type WAHAImportStrategy string
const (
    WAHAImportNone    WAHAImportStrategy = "none"      // Don't import history
    WAHAImportNewOnly WAHAImportStrategy = "new_only"  // Import only new messages
    WAHAImportAll     WAHAImportStrategy = "all"       // Import all messages
)

// WAHA Config (stored in Channel.config)
type WAHAConfig struct {
    BaseURL         string             // WAHA server URL
    Auth            WAHAAuth           // API credentials
    SessionID       string             // WAHA session identifier
    WebhookURL      string             // CRM webhook endpoint
    QRCode          string             // Current QR code (if scanning)
    QRCodeExpiresAt *time.Time         // QR code expiration
    SessionStatus   WAHASessionStatus  // Current session state
    ImportStrategy  WAHAImportStrategy // History import strategy
    ImportCompleted bool               // History import finished
}
```

### Business Invariants

1. **Channel must belong to Project**
   - `projectID` and `tenantID` required
   - `name` required
   - `channelType` required

2. **Channel type determines config structure**
   - WAHA requires: `baseURL`, `sessionID`, `auth`
   - WhatsApp requires: `phoneNumberID`, `accessToken`
   - Telegram requires: `botToken`

3. **Webhook lifecycle**
   - Webhook must be configured before channel can be activated
   - Webhook URL is auto-generated per channel
   - Webhook can be reconfigured if external platform changes

4. **Pipeline association**
   - Channel can have default pipeline for all incoming messages
   - If no pipeline, messages route to default project pipeline
   - Pipeline can be changed at any time

5. **QR Code lifecycle (WAHA only)**
   - QR codes expire after ~60 seconds
   - New QR code generated on each connection attempt
   - QR code cleared after successful scan

---

## Events Emitted

| Event | When | Purpose |
|-------|------|---------|
| `channel.created` | New channel created | Initialize webhook |
| `channel.updated` | Channel modified | Sync configuration |
| `channel.activated` | Channel enabled | Start message processing |
| `channel.deactivated` | Channel disabled | Stop message processing |
| `channel.deleted` | Channel removed | Cleanup webhooks |
| `channel.pipeline.associated` | Pipeline linked | Route messages to pipeline |
| `channel.pipeline.disassociated` | Pipeline unlinked | Stop pipeline routing |

---

## Repository Interface

```go
type Repository interface {
    Save(ctx context.Context, channel *Channel) error
    FindByID(ctx context.Context, id uuid.UUID) (*Channel, error)
    FindByExternalID(ctx context.Context, externalID string) (*Channel, error)
    FindByProject(ctx context.Context, projectID uuid.UUID) ([]*Channel, error)
    FindActiveByProject(ctx context.Context, projectID uuid.UUID) ([]*Channel, error)
    Delete(ctx context.Context, id uuid.UUID) error

    // Advanced queries
    FindByTenantWithFilters(ctx context.Context, filters ChannelFilters) ([]*Channel, int64, error)
    SearchByText(ctx context.Context, tenantID string, searchText string, limit int, offset int) ([]*Channel, int64, error)
}
```

---

## Commands (CQRS)

### ✅ Implemented

1. **CreateChannelCommand**
2. **ActivateChannelCommand**
3. **DeactivateChannelCommand**
4. **ConfigureWebhookCommand**
5. **AssociatePipelineCommand**
6. **ActivateWAHAChannelCommand** - WAHA-specific activation
7. **ImportWAHAHistoryCommand** - Bulk import messages

### ❌ Suggested

- **RefreshWAHASessionCommand** - Refresh QR code
- **TestChannelConnectionCommand** - Verify connectivity
- **BulkImportMessagesCommand** - Generic history import
- **SyncChannelConfigCommand** - Sync with external platform
- **GenerateChannelReportCommand** - Usage analytics

---

## Use Cases

### ✅ Implemented

1. **CreateChannelUseCase** - Create new channel (via handler)
2. **ActivateWAHAChannelUseCase** - Start WAHA session with QR code
3. **ImportWAHAHistoryUseCase** - Import message history via Temporal workflow

### ❌ Suggested

4. **ProcessInboundMessageUseCase** - Route incoming messages
5. **SendOutboundMessageUseCase** - Send via channel
6. **RefreshChannelStatusUseCase** - Check connection health
7. **CalculateChannelMetricsUseCase** - Generate analytics
8. **RotateWebhookSecretUseCase** - Security refresh

---

## WAHA Integration Details

### WAHA Channel Lifecycle

```
1. Create Channel
   → POST /api/v1/channels
   → Channel created with status=connecting

2. Activate WAHA Session
   → POST /api/v1/channels/{id}/activate-waha
   → WAHA API called to start session
   → QR code generated
   → Status = SCAN_QR_CODE

3. User Scans QR Code
   → WAHA webhook notifies CRM
   → Session status = WORKING
   → Channel status = active

4. Import History (Optional)
   → POST /api/v1/channels/{id}/import-history
   → Temporal workflow started
   → Messages imported in background

5. Process Messages
   → Incoming webhooks from WAHA
   → Route to sessions/contacts
   → Send outbound messages via WAHA API
```

### QR Code Management

```go
// Generate new QR code
channel.UpdateWAHAQRCode(qrCode string)
channel.SetWAHAQRCodeExpiration(time.Now().Add(60 * time.Second))

// Check if QR code is valid
if channel.IsWAHAQRCodeValid() {
    // Display QR code to user
    channel.LogQRCodeToConsole()
} else {
    // Need to refresh
    if channel.NeedsNewQRCode() {
        // Request new QR code from WAHA
    }
}

// QR code expires after 60 seconds
func (c *Channel) IsWAHAQRCodeValid() bool {
    if c.wahaQRCodeExpiresAt == nil {
        return false
    }
    return time.Now().Before(*c.wahaQRCodeExpiresAt)
}
```

### History Import Workflow

```go
// Temporal workflow for async import
type WAHAHistoryImportWorkflowInput struct {
    ChannelID string
    SessionID string
    Strategy  string  // "recent", "new_only", "all"
    Limit     int     // Messages per batch
    ProjectID string
    TenantID  string
    UserID    string
}

// Workflow steps:
1. Validate channel exists and is WAHA type
2. Fetch messages from WAHA API in batches
3. For each message:
   a. Find or create Contact
   b. Find or create Session
   c. Create Message
   d. Emit events
4. Update channel.ImportCompleted = true
5. Return import summary
```

---

## Channel Type Configurations

### 1. WAHA (WhatsApp via WAHA)

```go
config := map[string]interface{}{
    "baseURL":    "https://waha.example.com",
    "sessionID":  "my-whatsapp",
    "auth": map[string]string{
        "type":   "basic",
        "username": "admin",
        "password": "secret",
    },
    "webhookURL":      "https://crm.example.com/webhooks/waha/{channelID}",
    "qrCode":          "data:image/png;base64,...",
    "qrCodeExpiresAt": "2025-10-10T15:30:00Z",
    "sessionStatus":   "WORKING",
    "importStrategy":  "new_only",
    "importCompleted": true,
}
```

**Use Cases**: Small businesses, personal WhatsApp accounts

### 2. WhatsApp Business API

```go
config := map[string]interface{}{
    "phoneNumberID":    "123456789",
    "businessAccountID": "987654321",
    "accessToken":      "EAAxxxxxxxxxxxxx",
    "webhookToken":     "random-secret-token",
    "apiVersion":       "v17.0",
}
```

**Use Cases**: Large businesses, verified business accounts

### 3. Telegram Bot

```go
config := map[string]interface{}{
    "botToken":    "123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11",
    "botUsername": "@MyCompanyBot",
    "webhookURL":  "https://crm.example.com/webhooks/telegram/{channelID}",
}
```

**Use Cases**: Telegram communities, support bots

### 4. Facebook Messenger

```go
config := map[string]interface{}{
    "pageID":       "123456789012345",
    "pageAccessToken": "EAAxxxxxxxxxxxxx",
    "appSecret":    "app-secret-key",
    "webhookToken": "verify-token",
}
```

**Use Cases**: Facebook business pages

### 5. Instagram Direct

```go
config := map[string]interface{}{
    "instagramAccountID": "123456789",
    "pageAccessToken":    "EAAxxxxxxxxxxxxx",
    "webhookToken":       "verify-token",
}
```

**Use Cases**: Instagram business accounts

---

## Webhook Management

### Automatic Webhook Configuration

```go
// Webhook URL pattern
webhookURL := fmt.Sprintf("https://crm.example.com/webhooks/%s/%s",
    channelType, channelID)

// Webhook verification
webhookSecret := generateSecureToken()

// Store in channel
channel.SetWebhookURL(webhookURL)
channel.MarkWebhookConfigured()
```

### Webhook Endpoints

| Channel Type | Webhook Path | Verification |
|-------------|-------------|--------------|
| WAHA | `/webhooks/waha/{channelID}` | None (internal) |
| WhatsApp | `/webhooks/whatsapp/{channelID}` | Token verification |
| Telegram | `/webhooks/telegram/{channelID}` | Bot token |
| Messenger | `/webhooks/messenger/{channelID}` | App secret |
| Instagram | `/webhooks/instagram/{channelID}` | App secret |

---

## Performance Metrics

### Tracked Metrics

```go
type ChannelMetrics struct {
    MessagesReceived  int       // Total inbound messages
    MessagesSent      int       // Total outbound messages
    LastMessageAt     time.Time // Last activity
    LastErrorAt       time.Time // Last error timestamp
    LastError         string    // Error message
}

// Update on message received
channel.IncrementMessagesReceived()
channel.RecordLastMessageAt(time.Now())

// Update on error
channel.RecordError(errorMessage)
```

### Suggested Metrics (Not Implemented)

```go
type AdvancedMetrics struct {
    AverageResponseTimeMs  int       // Outbound message latency
    WebhookSuccessRate     float64   // Webhook reliability
    MessageDeliveryRate    float64   // % of messages delivered
    SessionsCreated        int       // New conversations
    ActiveSessionsCount    int       // Concurrent sessions
    ErrorRate              float64   // Errors per 1000 messages
    UptimePercentage       float64   // Availability
}
```

---

## API Examples

### Create WAHA Channel

```http
POST /api/v1/channels
{
  "name": "Main WhatsApp",
  "channel_type": "waha",
  "config": {
    "baseURL": "https://waha.example.com",
    "sessionID": "main-whatsapp",
    "auth": {
      "type": "basic",
      "username": "admin",
      "password": "secret"
    },
    "importStrategy": "new_only"
  }
}
```

### Activate WAHA Channel (Get QR Code)

```http
POST /api/v1/channels/{id}/activate-waha
{
  "session_id": "main-whatsapp"
}

Response:
{
  "success": true,
  "qr_code": "data:image/png;base64,iVBORw0KGgoAAAANS...",
  "qr_code_expires_at": "2025-10-10T15:30:00Z",
  "session_status": "SCAN_QR_CODE",
  "message": "Scan QR code with WhatsApp mobile app"
}
```

### Import Message History

```http
POST /api/v1/channels/{id}/import-history
{
  "strategy": "new_only",
  "limit": 100
}

Response:
{
  "success": true,
  "workflow_id": "waha-import-abc123",
  "message": "History import started in background"
}
```

### Configure Webhook

```http
POST /api/v1/channels/{id}/configure-webhook
{
  "webhook_url": "https://custom.domain.com/webhook"
}

Response:
{
  "success": true,
  "webhook_url": "https://crm.example.com/webhooks/waha/abc-123",
  "webhook_configured": true
}
```

### Get Webhook Info

```http
GET /api/v1/channels/{id}/webhook-info

Response:
{
  "webhook_url": "https://crm.example.com/webhooks/waha/abc-123",
  "webhook_active": true,
  "webhook_configured_at": "2025-10-10T10:00:00Z",
  "last_webhook_received": "2025-10-10T14:30:00Z"
}
```

### List Channels with Filters

```http
GET /api/v1/channels?type=waha&status=active&page=1&limit=20

Response:
{
  "channels": [
    {
      "id": "abc-123",
      "name": "Main WhatsApp",
      "channel_type": "waha",
      "status": "active",
      "messages_received": 1247,
      "messages_sent": 892,
      "last_message_at": "2025-10-10T14:30:00Z"
    }
  ],
  "total": 3,
  "page": 1,
  "limit": 20
}
```

### Activate Channel

```http
POST /api/v1/channels/{id}/activate

Response:
{
  "success": true,
  "channel_id": "abc-123",
  "status": "active",
  "activated_at": "2025-10-10T15:00:00Z"
}
```

### Deactivate Channel

```http
POST /api/v1/channels/{id}/deactivate

Response:
{
  "success": true,
  "channel_id": "abc-123",
  "status": "inactive",
  "deactivated_at": "2025-10-10T15:05:00Z"
}
```

---

## Real-World Usage

### WAHA Channel Setup Flow

```
1. User creates WAHA channel
   → POST /api/v1/channels
   → Channel created with status=connecting

2. User activates channel to get QR code
   → POST /api/v1/channels/{id}/activate-waha
   → QR code displayed in UI (valid for 60 seconds)

3. User scans QR code with WhatsApp mobile
   → WAHA webhook notifies CRM
   → Channel status → active
   → Session status → WORKING

4. User optionally imports message history
   → POST /api/v1/channels/{id}/import-history
   → Temporal workflow runs in background
   → Existing messages imported as Message entities

5. Channel ready to process messages
   → Incoming webhooks create Contact/Session/Message
   → Outbound messages sent via WAHA API
```

### Message Routing Logic

```
Incoming Message from Channel:
1. Webhook received at /webhooks/{type}/{channelID}
2. Parse webhook payload
3. Find Channel by ID
4. Extract sender identifier (phone, telegram_id, etc)
5. Find or create Contact
6. Find or create Session (or reopen if timed out)
7. Create Message entity
8. Route to Pipeline (if configured)
9. Route to Agent (if assigned)
10. Trigger Automations (if rules match)
11. Send outbound response (if automated)
```

---

## Performance Considerations

### Indexes

```sql
-- Channels
CREATE INDEX idx_channels_project ON channels(project_id);
CREATE INDEX idx_channels_type ON channels(channel_type);
CREATE INDEX idx_channels_status ON channels(status);
CREATE INDEX idx_channels_external ON channels(external_id);
CREATE INDEX idx_channels_active ON channels(project_id, status)
    WHERE status = 'active';

-- Webhook lookups
CREATE INDEX idx_channels_webhook ON channels(webhook_id);
```

### Caching Strategy

```go
// Cache active channels per project (5 min TTL)
cacheKey := fmt.Sprintf("channels:project:%s:active", projectID)
channels, err := cache.Get(cacheKey)

// Cache channel config (1 min TTL)
configKey := fmt.Sprintf("channel:%s:config", channelID)
config, err := cache.Get(configKey)
```

### Webhook Processing

```go
// Process webhooks asynchronously
go func() {
    // 1. Parse webhook
    // 2. Create entities
    // 3. Emit events
    // 4. Send response (if needed)
}()

// Respond immediately (within 5 seconds for WhatsApp)
return http.StatusOK
```

---

## Relationships

### Channel → Pipeline (Many-to-One, Optional)

```go
// Set default pipeline for channel
channel.AssociatePipeline(pipelineID)

// All messages from this channel route to this pipeline
// (unless contact already has different pipeline)

// Remove pipeline association
channel.DisassociatePipeline()
```

### Channel → Message (One-to-Many)

```go
// Find all messages for channel
messages, _ := messageRepo.FindByChannel(ctx, channelID)

// Count messages
count := channel.GetMessagesReceived() + channel.GetMessagesSent()
```

### Channel → Session (One-to-Many)

```go
// Find all sessions for channel
sessions, _ := sessionRepo.FindByChannel(ctx, channelID)
```

---

## References

- [Channel Domain](../../internal/domain/channel/)
- [Channel Events](../../internal/domain/channel/events.go)
- [Channel Repository](../../infrastructure/persistence/gorm_channel_repository.go)
- [Channel Handler](../../infrastructure/http/handlers/channel_handler.go)
- [WAHA Client](../../infrastructure/channels/waha/client.go)
- [WAHA Webhook Handler](../../infrastructure/http/handlers/waha_webhook_handler.go)

---

**Next**: [ChannelType Aggregate](channel_type_aggregate.md) →
**Previous**: [Agent Aggregate](agent_aggregate.md) ←

---

## Summary

✅ **Channel Aggregate Features**:
1. Multi-channel support (WAHA, WhatsApp, Telegram, Messenger, Instagram)
2. Webhook auto-configuration and management
3. WAHA-specific QR code authentication
4. Message history import via Temporal workflows
5. Pipeline association for message routing
6. AI integration capabilities
7. Comprehensive metrics tracking

The Channel aggregate is the **gateway** for all external communication in the Ventros CRM system.
