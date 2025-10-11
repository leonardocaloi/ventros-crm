# Message Aggregate

**Last Updated**: 2025-10-10
**Status**: ✅ Complete and Production-Ready
**Lines of Code**: ~1,500 (with tests)
**Test Coverage**: Partial (8+ tests passing for value objects)

---

## Overview

- **Purpose**: Represents individual chat messages in conversations
- **Location**: `internal/domain/message/`
- **Entity**: `infrastructure/persistence/entities/message_entity.go`
- **Repository**: `infrastructure/persistence/gorm_message_repository.go`
- **Aggregate Root**: `Message`

**Business Problem**:
The Message aggregate represents **individual messages** exchanged between contacts and the CRM system (via human agents or bots). Messages are the fundamental building blocks of conversations and can contain:
- **Text content** - Regular chat messages
- **Media attachments** - Images, videos, audio, documents
- **System notifications** - Session start/end, assignments
- **Location sharing** - GPS coordinates
- **Contact cards** - Shared contact information

Messages are critical for:
- **Conversation history** - Store all interactions
- **Delivery tracking** - Monitor message status (sent → delivered → read)
- **AI processing** - Analyze media content (transcription, OCR, vision)
- **Compliance** - Audit trail for regulatory requirements
- **Analytics** - Response times, engagement metrics

---

## Domain Model

### Aggregate Root: Message

```go
type Message struct {
    id               uuid.UUID
    timestamp        time.Time   // When message was sent/received
    customerID       uuid.UUID   // Multi-tenant: customer/tenant
    projectID        uuid.UUID   // Multi-tenant: project
    channelTypeID    *int        // Channel type (1=WhatsApp, 2=Email, etc)
    fromMe           bool        // Direction: true=outbound, false=inbound
    channelID        uuid.UUID   // Channel used
    contactID        uuid.UUID   // Required: belongs to contact
    sessionID        *uuid.UUID  // Optional: belongs to session (null for first message)

    // Content
    contentType      ContentType // text, image, video, audio, etc
    text             *string     // Text content (required for text messages)
    mediaURL         *string     // URL to media file (required for media messages)
    mediaMimetype    *string     // MIME type (e.g., "image/jpeg", "video/mp4")

    // External references
    channelMessageID *string     // External message ID (e.g., WhatsApp message ID)
    replyToID        *uuid.UUID  // Reply to another message (threading)

    // Status tracking
    status           Status      // queued, sent, delivered, read, failed
    deliveredAt      *time.Time  // When message was delivered
    readAt           *time.Time  // When message was read

    // Metadata
    language         *string                 // Detected language
    agentID          *uuid.UUID              // Agent who sent (for outbound messages)
    metadata         map[string]interface{}  // Custom metadata (JSON)

    // Event sourcing
    events []DomainEvent
}
```

### Value Objects

#### 1. ContentType (types.go:5)
```go
type ContentType string

const (
    ContentTypeText     ContentType = "text"
    ContentTypeImage    ContentType = "image"
    ContentTypeVideo    ContentType = "video"
    ContentTypeAudio    ContentType = "audio"
    ContentTypeVoice    ContentType = "voice"     // Voice notes
    ContentTypeDocument ContentType = "document"
    ContentTypeLocation ContentType = "location"
    ContentTypeContact  ContentType = "contact"   // Shared contact card
    ContentTypeSticker  ContentType = "sticker"
    ContentTypeSystem   ContentType = "system"    // System notifications
)

func (ct ContentType) IsText() bool
func (ct ContentType) IsMedia() bool
func (ct ContentType) IsSystem() bool
func (ct ContentType) RequiresURL() bool
```

**Invariants**:
- Must be one of the defined types
- Media types require `mediaURL` and `mediaMimetype`
- Text type requires `text` field
- System messages don't require content

#### 2. Status (types.go:57)
```go
type Status string

const (
    StatusQueued    Status = "queued"    // Waiting to be sent
    StatusSent      Status = "sent"      // Sent but not delivered
    StatusDelivered Status = "delivered" // Delivered to device
    StatusRead      Status = "read"      // Read by contact
    StatusFailed    Status = "failed"    // Failed to send
)
```

**Invariants**:
- Status transitions: queued → sent → delivered → read
- Failed status is terminal (no further transitions)
- Timestamps must correspond to status

#### 3. MediaURL (value_objects.go)
```go
type MediaURL struct {
    Value string
}

func NewMediaURL(value string) (MediaURL, error) {
    // Validates URL format (must be http:// or https://)
}
```

**Invariants**:
- Must be valid HTTP/HTTPS URL
- Required for media content types
- Immutable after creation

### Business Invariants

1. **Message must belong to a Contact**
   - `contactID` cannot be nil
   - `projectID` and `customerID` cannot be nil (multi-tenancy)

2. **Content type determines required fields**
   - Text messages require `text` field
   - Media messages require `mediaURL` and `mediaMimetype`
   - System messages don't require content

3. **Direction determines behavior**
   - `fromMe=true` - Outbound message (from agent/bot)
   - `fromMe=false` - Inbound message (from contact)
   - Outbound messages can have `agentID`

4. **Status lifecycle**
   - Can only transition forward (cannot go back from delivered to sent)
   - `MarkAsDelivered()` sets delivered status and timestamp
   - `MarkAsRead()` sets read status and timestamp
   - `MarkAsFailed()` sets failed status (terminal)

5. **Session assignment**
   - First message from contact creates new session
   - Subsequent messages assigned to active session
   - Session can be null during message creation

6. **AI processing is automatic**
   - Media messages can request AI processing
   - `RequestAIProcessing()` emits events for AI workers
   - Based on channel configuration (enable/disable per media type)

---

## Events Emitted

The Message aggregate emits **9 domain events**:

| Event | When | Purpose |
|-------|------|---------|
| `message.created` | New message created | Initialize message, trigger routing |
| `message.delivered` | Message delivered | Update UI, track delivery metrics |
| `message.read` | Message read by recipient | Update UI, track engagement |
| `message.failed` | Message failed to send | Alert agent, retry logic |
| `message.ai.process_image_requested` | Image received | Trigger AI vision analysis |
| `message.ai.process_video_requested` | Video received | Trigger video analysis |
| `message.ai.process_audio_requested` | Audio received | Trigger audio transcription |
| `message.ai.process_voice_requested` | Voice note received | Trigger voice transcription |
| `tracking.message.meta_ads` | Message with UTM tracking | Attribute to ad campaign |

### Event Examples

```go
// MessageCreatedEvent
type MessageCreatedEvent struct {
    MessageID uuid.UUID
    ContactID uuid.UUID
    FromMe    bool
    CreatedAt time.Time
}

// MessageDeliveredEvent
type MessageDeliveredEvent struct {
    MessageID   uuid.UUID
    DeliveredAt time.Time
}

// MessageReadEvent
type MessageReadEvent struct {
    MessageID uuid.UUID
    ReadAt    time.Time
}

// AIProcessImageRequestedEvent
type AIProcessImageRequestedEvent struct {
    MessageID   uuid.UUID
    ChannelID   uuid.UUID
    ContactID   uuid.UUID
    SessionID   uuid.UUID
    ImageURL    string
    MimeType    string
    RequestedAt time.Time
}

// Publishing
msg, _ := NewMessage(contactID, projectID, customerID, ContentTypeText, false)
// Event automatically added to msg.events[]
eventBus.Publish(msg.DomainEvents()...)
```

---

## Repository Interface

```go
// internal/domain/message/repository.go
type Repository interface {
    Save(ctx context.Context, message *Message) error

    FindByID(ctx context.Context, id uuid.UUID) (*Message, error)

    FindBySession(ctx context.Context, sessionID uuid.UUID, limit, offset int) ([]*Message, error)

    FindByContact(ctx context.Context, contactID uuid.UUID, limit, offset int) ([]*Message, error)

    FindByChannelMessageID(ctx context.Context, channelMessageID string) (*Message, error)

    CountBySession(ctx context.Context, sessionID uuid.UUID) (int, error)

    // Advanced query methods
    FindByTenantWithFilters(ctx context.Context, filters MessageFilters) ([]*Message, int64, error)

    SearchByText(ctx context.Context, tenantID string, searchText string, limit int, offset int) ([]*Message, int64, error)
}

// MessageFilters - powerful filtering
type MessageFilters struct {
    TenantID        string
    ContactID       *uuid.UUID
    SessionID       *uuid.UUID
    ChannelID       *uuid.UUID
    ProjectID       *uuid.UUID
    ChannelTypeID   *int
    FromMe          *bool
    ContentType     *string
    Status          *string
    AgentID         *uuid.UUID
    TimestampAfter  *time.Time
    TimestampBefore *time.Time
    HasMedia        *bool
    Limit           int
    Offset          int
    SortBy          string // timestamp, created_at
    SortOrder       string // asc, desc
}
```

**Implementation**: `infrastructure/persistence/gorm_message_repository.go`

**Key Methods**:
- `FindBySession()` - Get conversation history
- `FindByChannelMessageID()` - Map external message IDs to internal IDs
- `FindByContact()` - Get all messages for contact (across sessions)
- `FindByTenantWithFilters()` - Advanced filtering for analytics
- `SearchByText()` - Full-text search across message content

---

## Commands (CQRS)

**Status**: ✅ Implemented in `internal/application/commands/message/`

### ✅ Implemented

#### 1. SendMessageCommand (`internal/application/commands/message/send_message.go`)
```go
type SendMessageCommand struct {
    ContactID   uuid.UUID
    ChannelID   uuid.UUID
    ContentType ContentType
    Text        *string
    MediaURL    *string
    ReplyToID   *uuid.UUID
    Metadata    map[string]interface{}
    TenantID    string
    ProjectID   uuid.UUID
    CustomerID  uuid.UUID
    AgentID     *uuid.UUID
}

// Sends outbound message to contact via channel
// Flow:
// 1. Validate channel exists and is active
// 2. Create Message aggregate
// 3. Save to database
// 4. Dispatch to channel (WAHA, Email, SMS)
// 5. Publish message.created event
```

**Handler**: `SendMessageHandler` - `internal/application/commands/message/send_message_handler.go`

#### 2. ConfirmMessageDeliveryCommand (`internal/application/commands/message/confirm_delivery.go`)
```go
type ConfirmMessageDeliveryCommand struct {
    MessageID     uuid.UUID
    ExternalID    string  // WhatsApp message ID, etc
    Status        string  // "delivered", "read", "failed"
    DeliveredAt   *time.Time
    ReadAt        *time.Time
    FailureReason *string
}

// Updates message status based on external notification
// Called by: Webhook handlers (WAHA, Twilio, etc)
```

**Handler**: `ConfirmMessageDeliveryHandler` - `internal/application/commands/message/confirm_delivery_handler.go`

### ❌ Suggested (Not Implemented)

- **EditMessageCommand** - Edit sent message (if channel supports)
- **DeleteMessageCommand** - Delete/recall message (if channel supports)
- **TranslateMessageCommand** - Translate message to different language
- **ForwardMessageCommand** - Forward message to another contact
- **ScheduleMessageCommand** - Schedule message for future delivery
- **BulkSendMessageCommand** - Send same message to multiple contacts

---

## Queries (CQRS)

**Status**: ✅ Implemented in `internal/application/queries/`

### ✅ Implemented

#### 1. ListMessagesQuery (`internal/application/queries/list_messages_query.go`)
```go
type ListMessagesQuery struct {
    TenantID    shared.TenantID
    ContactID   *uuid.UUID
    SessionID   *uuid.UUID
    ChannelID   *uuid.UUID
    FromMe      *bool
    ContentType *string
    Status      *string
    Page        int
    Limit       int
    SortBy      string
    SortDir     string
}

// Returns: ListMessagesResponse with pagination
// Use cases:
// - Get conversation history
// - Filter by direction (inbound/outbound)
// - Filter by media type
// - Track delivery status
```

#### 2. SearchMessagesQuery (`internal/application/queries/search_messages_query.go`)
```go
type SearchMessagesQuery struct {
    TenantID   shared.TenantID
    SearchText string
    Limit      int
}

// Full-text search across message text
// Uses PostgreSQL ILIKE for case-insensitive search
// Returns: Matching messages with excerpts
```

#### 3. GetMessageByID
```go
// Implemented in handler: message_handler.go:233
// Returns: Full message details
```

#### 4. GetMessagesBySession
```go
// Implemented in handler: message_handler.go:322
// Returns: All messages in a session (conversation thread)
```

### ❌ Suggested (Not Implemented)

- **GetMediaMessagesByContactQuery** - All media from specific contact
- **GetUnreadMessagesQuery** - Messages not yet read by agent
- **GetFailedMessagesQuery** - Failed messages for retry
- **GetMessageStatsQuery** - Aggregate stats (avg response time, etc)
- **GetTopKeywordsQuery** - Most common words/phrases

---

## Use Cases

### ✅ Implemented

#### 1. ProcessInboundMessageUseCase (`internal/application/message/process_inbound_message.go`)
```go
// Processes incoming messages from channels (WhatsApp, etc)
// Flow:
// 1. Receive message from WAHA webhook
// 2. Find or create contact
// 3. Find or create session
// 4. Create Message aggregate
// 5. Assign to session
// 6. Check if AI processing needed
// 7. Publish message.created event
// 8. Trigger automation rules

// Called by: WAHA webhook handler
```

#### 2. SendMessageUseCase (`internal/application/commands/message/send_message_handler.go`)
```go
// Sends outbound message to contact
// Flow:
// 1. Validate channel is active
// 2. Create Message aggregate
// 3. Save to database
// 4. Dispatch to channel adapter (WAHA, Email, SMS)
// 5. Update status based on result
// 6. Publish events

// Called by: API handler, Automation rules, Bot responses
```

#### 3. ConfirmDeliveryUseCase (`internal/application/commands/message/confirm_delivery_handler.go`)
```go
// Updates message delivery status
// Flow:
// 1. Find message by external ID
// 2. Update status (delivered/read/failed)
// 3. Set timestamps
// 4. Publish status event
// 5. Update session metrics

// Called by: Channel webhooks (delivery receipts)
```

### ❌ Suggested (Not Implemented)

#### 4. TranscribeVoiceMessageUseCase
**Purpose**: Transcribe voice notes and audio messages
**Trigger**: AI worker picks up `message.ai.process_voice_requested` event
**Events**: `message.transcribed`
**External Dependencies**: OpenAI Whisper API, AssemblyAI
**Process**:
1. Download audio file from URL
2. Send to transcription service
3. Store transcript in message metadata
4. Optionally translate to different language

#### 5. AnalyzeImageMessageUseCase
**Purpose**: Extract text and analyze images using AI vision
**Trigger**: AI worker picks up `message.ai.process_image_requested` event
**Events**: `message.image_analyzed`
**External Dependencies**: OpenAI GPT-4V, Google Cloud Vision
**Process**:
1. Download image from URL
2. Send to vision API
3. Extract: text (OCR), objects, faces, sentiment
4. Store analysis in message metadata
5. Trigger automation based on content (e.g., "receipt detected")

#### 6. DetectLanguageUseCase
**Purpose**: Auto-detect message language
**Trigger**: New text message created
**Process**:
1. Analyze text with language detection library
2. Set message.language field
3. Enable auto-translation if needed

#### 7. BulkImportMessagesUseCase
**Purpose**: Import historical messages from other systems
**Trigger**: Admin initiates import
**Process**:
1. Parse CSV/JSON file
2. Create Message aggregates
3. Assign to contacts and sessions
4. Preserve original timestamps
5. Don't send to channels (historical only)

---

## Use Cases Cheat Sheet

| Use Case | Status | Complexity | Priority |
|----------|--------|-----------|----------|
| ProcessInboundMessage | ✅ Done | High | Critical |
| SendMessage | ✅ Done | Medium | Critical |
| ConfirmDelivery | ✅ Done | Low | Critical |
| TranscribeVoice | ❌ TODO | High | High |
| AnalyzeImage | ❌ TODO | High | High |
| DetectLanguage | ❌ TODO | Medium | Medium |
| EditMessage | ❌ TODO | Low | Low |
| DeleteMessage | ❌ TODO | Low | Low |
| ScheduleMessage | ❌ TODO | Medium | Medium |
| BulkImport | ❌ TODO | High | Low |
| TranslateMessage | ❌ TODO | Medium | Medium |

---

## Relationships

### Belongs To (N:1)
- **Contact**: Every message belongs to a contact
- **Session**: Message can belong to a session (null for first message)
- **Channel**: Message sent/received via a channel
- **Project**: Multi-tenancy
- **Agent**: Outbound messages can have an agent

### Has Many (1:N)
- **None** - Messages are leaf entities

### References
- **ReplyTo**: Message can reply to another message (threading)
- **ContactEvent**: Message creation logged in contact timeline

---

## Performance Considerations

### Indexes (PostgreSQL)

```sql
-- Primary key
CREATE INDEX idx_messages_id ON messages(id);

-- Multi-tenancy (CRITICAL for all queries)
CREATE INDEX idx_messages_tenant ON messages(customer_id);

-- Conversation history (CRITICAL for UI)
CREATE INDEX idx_messages_session ON messages(session_id, timestamp DESC)
    WHERE session_id IS NOT NULL;

-- Contact history
CREATE INDEX idx_messages_contact ON messages(contact_id, timestamp DESC);

-- External message ID lookup (CRITICAL for webhooks)
CREATE UNIQUE INDEX idx_messages_channel_msg_id ON messages(channel_message_id)
    WHERE channel_message_id IS NOT NULL;

-- Channel-specific queries
CREATE INDEX idx_messages_channel ON messages(channel_id, timestamp DESC);

-- Status tracking (for failed messages retry)
CREATE INDEX idx_messages_status ON messages(status, timestamp)
    WHERE status IN ('queued', 'failed');

-- Media messages
CREATE INDEX idx_messages_media ON messages(content_type, timestamp DESC)
    WHERE content_type IN ('image', 'video', 'audio', 'voice', 'document');

-- Full-text search (PostgreSQL GIN index)
CREATE INDEX idx_messages_text_search ON messages USING gin(to_tsvector('english', coalesce(text, '')));

-- Agent performance tracking
CREATE INDEX idx_messages_agent ON messages(agent_id, timestamp DESC)
    WHERE agent_id IS NOT NULL;
```

### Caching Strategy (Redis)

**Current**: ❌ NOT IMPLEMENTED

**Suggested**:
```go
// Cache keys
message:by_id:{uuid}                     TTL: 10min
message:session:{sessionID}:latest       TTL: 5min  // Last message in session
message:contact:{contactID}:unread_count TTL: 1min

// Invalidation
- On message create: Increment unread count cache
- On message read: Decrement unread count cache
- On message update: Delete message cache
```

**Impact**: 40-60% reduction in database queries for conversation views

### Message Storage Optimization

**Current**: All messages in `messages` table

**Suggested for scale**:
1. **Hot/Cold storage partitioning**
   ```sql
   -- Partition by month for recent messages (hot)
   CREATE TABLE messages_2025_10 PARTITION OF messages
       FOR VALUES FROM ('2025-10-01') TO ('2025-11-01');

   -- Archive old messages (cold storage) after 6 months
   ```

2. **Media URL caching**
   - Cache pre-signed URLs for media (avoid database lookup)
   - Expire after 1 hour
   - Regenerate on demand

3. **Text compression**
   - Compress long messages (>1KB)
   - Store compressed in `text_compressed` column
   - Decompress on read

---

## Testing

### Unit Tests (`value_objects_test.go`)
✅ **8+ tests passing**

Test Coverage:
```
TestNewMediaURL_Valid                   ✅
TestNewMediaURL_Invalid                 ✅
TestContentType_IsValid                 ✅
TestContentType_IsText                  ✅
TestContentType_IsMedia                 ✅
TestContentType_IsSystem                ✅
TestStatus_String                       ✅
TestParseContentType                    ✅
```

### Integration Tests
Location: `infrastructure/persistence/gorm_message_repository_test.go`
⚠️ Tests need expansion

### E2E Tests
Location: `tests/e2e/message_send_test.go`
✅ Basic send message test exists

---

## Suggested Improvements

### 1. Add Message Reactions
```go
// Allow contacts/agents to react to messages (like, love, etc)
type Reaction struct {
    Emoji     string    // "👍", "❤️", etc
    UserID    uuid.UUID
    Timestamp time.Time
}

func (m *Message) AddReaction(emoji string, userID uuid.UUID)
func (m *Message) RemoveReaction(emoji string, userID uuid.UUID)
```

### 2. Implement Message Threading
```go
// Better support for threaded conversations
func (m *Message) GetThread() ([]*Message, error)
func (m *Message) GetReplies() ([]*Message, error)
```

### 3. Add Message Encryption (E2E)
```go
// End-to-end encryption for sensitive content
type EncryptedMessage struct {
    EncryptedText []byte
    KeyID         string
    Algorithm     string
}

func (m *Message) Encrypt(publicKey []byte) error
func (m *Message) Decrypt(privateKey []byte) error
```

### 4. Implement Rich Media Support
```go
// Better support for rich media (carousels, buttons, etc)
type RichContent struct {
    Type    string // "carousel", "buttons", "list"
    Items   []RichItem
    Actions []Action
}

func (m *Message) SetRichContent(content RichContent)
```

### 5. Add AI-Generated Summaries
```go
// Automatically summarize long messages
func (m *Message) GenerateSummary(maxLength int) (string, error)
```

### 6. Implement Message Quality Score
```go
// Score messages for quality (spam detection, sentiment, etc)
type QualityScore struct {
    IsSpam      bool
    Sentiment   float64 // -1.0 to 1.0
    Profanity   bool
    Urgency     float64 // 0.0 to 1.0
    Clarity     float64 // 0.0 to 1.0
}

func (m *Message) CalculateQualityScore() QualityScore
```

---

## API Examples

### Send Message (Text)
```http
POST /api/v1/messages/send
Authorization: Bearer <token>
Content-Type: application/json

{
  "contact_id": "550e8400-e29b-41d4-a716-446655440000",
  "channel_id": "660e8400-e29b-41d4-a716-446655440001",
  "content_type": "text",
  "text": "Hello! How can I help you today?"
}
```

**Response**:
```json
{
  "message_id": "770e8400-e29b-41d4-a716-446655440002",
  "external_id": "wamid.HBgNNTU1MTIzNDU2Nzg5FQIAERgSMEE2NjdCNzE5OTEzRDhFMkYzAA==",
  "status": "sent",
  "sent_at": "2025-10-10T10:30:00Z"
}
```

### Send Message (Image)
```http
POST /api/v1/messages/send
Authorization: Bearer <token>
Content-Type: application/json

{
  "contact_id": "550e8400-e29b-41d4-a716-446655440000",
  "channel_id": "660e8400-e29b-41d4-a716-446655440001",
  "content_type": "image",
  "media_url": "https://example.com/image.jpg",
  "text": "Check out this product!"
}
```

### Confirm Delivery
```http
POST /api/v1/messages/confirm-delivery
Authorization: Bearer <token>
Content-Type: application/json

{
  "external_id": "wamid.HBgNNTU1MTIzNDU2Nzg5FQIAERgSMEE2NjdCNzE5OTEzRDhFMkYzAA==",
  "status": "delivered",
  "delivered_at": "2025-10-10T10:30:15Z"
}
```

### List Messages (Advanced)
```http
GET /api/v1/crm/messages/advanced?session_id=550e8400-e29b-41d4-a716-446655440000&page=1&limit=50&sort_by=timestamp&sort_dir=asc
Authorization: Bearer <token>
```

### Search Messages
```http
GET /api/v1/crm/messages/search?q=refund request&limit=20
Authorization: Bearer <token>
```

---

## Real-World Usage Patterns

### Pattern 1: Inbound Message Processing (WhatsApp)
```
1. WAHA sends webhook to /webhooks/waha
2. WAHAWebhookHandler receives message
3. Calls ProcessInboundMessageUseCase
4. UseCase:
   a. Find/create contact by phone
   b. Find active session or create new
   c. Create Message aggregate
   d. Assign to session
   e. Check AI processing (if image/voice)
   f. Publish message.created event
5. ContactEventConsumer picks up event
6. Updates session metrics (message count, last activity)
7. Triggers automation rules (if any)
8. Bot/Agent responds
```

### Pattern 2: Outbound Message (Agent Reply)
```
1. Agent types message in UI
2. Frontend calls POST /api/v1/messages/send
3. SendMessageHandler:
   a. Validate channel is active
   b. Create Message aggregate (fromMe=true)
   c. Save to database
   d. Call WAHAMessageSender.SendMessage()
   e. WAHA API sends to WhatsApp
   f. Publish message.created event
4. WAHA receives delivery receipt
5. Calls POST /api/v1/messages/confirm-delivery
6. Message status updated to "delivered"
7. UI updates in real-time (WebSocket)
```

### Pattern 3: Voice Message with AI Transcription
```
1. Contact sends voice note via WhatsApp
2. WAHA webhook delivers message
3. ProcessInboundMessage creates Message (contentType=voice)
4. Checks channel config: processVoice=true
5. Calls message.RequestAIProcessing()
6. Publishes message.ai.process_voice_requested event
7. AI Worker picks up event
8. Downloads audio file
9. Sends to OpenAI Whisper API
10. Receives transcript
11. Updates message.metadata with transcript
12. Bot can now respond to transcript content
```

---

## Integration Points

### WAHA (WhatsApp)
- **Inbound**: `infrastructure/http/handlers/waha_webhook_handler.go`
- **Outbound**: `infrastructure/channels/waha/client.go`
- **Message Types**: text, image, video, audio, voice, document, location, contact
- **Status Updates**: delivered, read, failed

### Message Debouncer
- **Purpose**: Prevent duplicate inbound messages
- **Location**: `infrastructure/messaging/message_debouncer.go`
- **TTL**: 10 minutes
- **Key**: `{channelMessageID}`

### AI Processing
- **Events**: `message.ai.process_*_requested`
- **Workers**: Background jobs process media
- **Services**: OpenAI, Google Cloud Vision, AssemblyAI

---

## References

- [Message Domain](../../internal/domain/message/)
- [Message Repository](../../infrastructure/persistence/gorm_message_repository.go)
- [Message Handler](../../infrastructure/http/handlers/message_handler.go)
- [Send Message Command](../../internal/application/commands/message/send_message.go)
- [Confirm Delivery Command](../../internal/application/commands/message/confirm_delivery.go)
- [Process Inbound Message](../../internal/application/message/process_inbound_message.go)
- [WAHA Webhook Handler](../../infrastructure/http/handlers/waha_webhook_handler.go)
- [WAHA Client](../../infrastructure/channels/waha/client.go)
- [List Messages Query](../../internal/application/queries/list_messages_query.go)
- [Search Messages Query](../../internal/application/queries/search_messages_query.go)

---

**Next**: [Pipeline Aggregate](pipeline_aggregate.md) →
**Previous**: [Session Aggregate](session_aggregate.md) ←
