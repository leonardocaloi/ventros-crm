# Chat Aggregate

**Last Updated**: 2025-10-10
**Status**: ❌ NOT IMPLEMENTED YET - Design Document
**Lines of Code**: 0 (to be implemented)
**Test Coverage**: N/A

---

## Overview

- **Purpose**: Provide context and grouping for messages in conversations
- **Location**: `internal/domain/chat/` (to be created)
- **Entity**: `infrastructure/persistence/entities/chat.go` (to be created)
- **Repository**: `infrastructure/persistence/gorm_chat_repository.go` (to be created)
- **Aggregate Root**: `Chat`

**Business Problem**:
The Chat aggregate is **CRITICAL** for providing proper context to messages. Currently, messages only have `contactID` and `sessionID`, but lack a **chat context**. This creates problems for:
- **Group conversations** - Multiple participants in WhatsApp groups
- **Channel conversations** - Telegram channels, WhatsApp Business broadcasts
- **Individual conversations** - 1-on-1 chats with context preservation
- **Message history** - Grouping messages by conversation thread
- **Participant management** - Who's in the conversation (contacts + agents)
- **Chat metadata** - Subject, description, mute status, archived status
- **Multi-channel support** - Same contact across different channels needs separate chats

---

## Why Chat is CRITICAL

### Current Problem

**Messages currently have NO chat context**:

```go
// internal/domain/message/message.go (CURRENT)
type Message struct {
    id         uuid.UUID
    contactID  uuid.UUID   // Who sent/received
    sessionID  *uuid.UUID  // Optional - active session
    channelID  uuid.UUID   // Which channel
    agentID    *uuid.UUID  // Optional - which agent
    // ... no chatID!
}
```

**Issues**:
1. ❌ No way to group messages by conversation
2. ❌ No support for WhatsApp groups (multiple participants)
3. ❌ No way to distinguish multiple conversations with same contact on different channels
4. ❌ No chat-level metadata (subject, description, archived, muted)
5. ❌ Imported historical messages have no context

### Solution: Chat Aggregate

```go
// internal/domain/chat/chat.go (PROPOSED)
type Chat struct {
    id           uuid.UUID
    projectID    uuid.UUID
    tenantID     string
    chatType     ChatType      // individual, group, channel
    subject      *string       // Optional (group name, channel name)
    description  *string       // Optional
    participants []Participant // Contacts + Agents
    status       ChatStatus    // active, archived, closed
    metadata     map[string]interface{}
    lastMessageAt *time.Time
    createdAt    time.Time
    updatedAt    time.Time
}
```

**Benefits**:
1. ✅ Messages have proper conversation context
2. ✅ Support for groups/channels (multiple participants)
3. ✅ Multiple conversations with same contact (different chats)
4. ✅ Chat-level operations (archive, mute, close)
5. ✅ Historical context preserved

---

## Domain Model

### Aggregate Root: Chat

```go
package chat

import (
    "errors"
    "time"

    "github.com/google/uuid"
)

type Chat struct {
    id            uuid.UUID
    projectID     uuid.UUID
    tenantID      string
    chatType      ChatType
    subject       *string       // Group name, channel name (optional for individual)
    description   *string       // Group/channel description
    participants  []Participant // All participants (contacts + agents)
    status        ChatStatus
    metadata      map[string]interface{}
    lastMessageAt *time.Time
    createdAt     time.Time
    updatedAt     time.Time

    events []DomainEvent
}

// NewIndividualChat creates 1-on-1 chat
func NewIndividualChat(
    projectID uuid.UUID,
    tenantID string,
    contactID uuid.UUID,
) (*Chat, error) {
    if projectID == uuid.Nil {
        return nil, errors.New("projectID cannot be nil")
    }
    if tenantID == "" {
        return nil, errors.New("tenantID cannot be empty")
    }
    if contactID == uuid.Nil {
        return nil, errors.New("contactID cannot be nil")
    }

    now := time.Now()
    chat := &Chat{
        id:        uuid.New(),
        projectID: projectID,
        tenantID:  tenantID,
        chatType:  ChatTypeIndividual,
        participants: []Participant{
            {
                ID:         contactID,
                Type:       ParticipantTypeContact,
                JoinedAt:   now,
                IsAdmin:    false,
            },
        },
        status:    ChatStatusActive,
        metadata:  make(map[string]interface{}),
        createdAt: now,
        updatedAt: now,
        events:    []DomainEvent{},
    }

    chat.addEvent(NewChatCreatedEvent(chat.id, chat.chatType, chat.projectID))

    return chat, nil
}

// NewGroupChat creates group chat (WhatsApp group, Telegram group)
func NewGroupChat(
    projectID uuid.UUID,
    tenantID string,
    subject string,
    creatorID uuid.UUID,
) (*Chat, error) {
    if projectID == uuid.Nil {
        return nil, errors.New("projectID cannot be nil")
    }
    if tenantID == "" {
        return nil, errors.New("tenantID cannot be empty")
    }
    if subject == "" {
        return nil, errors.New("group subject cannot be empty")
    }
    if creatorID == uuid.Nil {
        return nil, errors.New("creatorID cannot be nil")
    }

    now := time.Now()
    chat := &Chat{
        id:        uuid.New(),
        projectID: projectID,
        tenantID:  tenantID,
        chatType:  ChatTypeGroup,
        subject:   &subject,
        participants: []Participant{
            {
                ID:       creatorID,
                Type:     ParticipantTypeContact,
                JoinedAt: now,
                IsAdmin:  true, // Creator is admin
            },
        },
        status:    ChatStatusActive,
        metadata:  make(map[string]interface{}),
        createdAt: now,
        updatedAt: now,
        events:    []DomainEvent{},
    }

    chat.addEvent(NewChatCreatedEvent(chat.id, chat.chatType, chat.projectID))

    return chat, nil
}

// NewChannelChat creates channel (Telegram channel, WhatsApp Business broadcast)
func NewChannelChat(
    projectID uuid.UUID,
    tenantID string,
    subject string,
) (*Chat, error) {
    if projectID == uuid.Nil {
        return nil, errors.New("projectID cannot be nil")
    }
    if tenantID == "" {
        return nil, errors.New("tenantID cannot be empty")
    }
    if subject == "" {
        return nil, errors.New("channel subject cannot be empty")
    }

    now := time.Now()
    chat := &Chat{
        id:           uuid.UUID{},
        projectID:    projectID,
        tenantID:     tenantID,
        chatType:     ChatTypeChannel,
        subject:      &subject,
        participants: []Participant{}, // Channels may have no explicit participants
        status:       ChatStatusActive,
        metadata:     make(map[string]interface{}),
        createdAt:    now,
        updatedAt:    now,
        events:       []DomainEvent{},
    }

    chat.addEvent(NewChatCreatedEvent(chat.id, chat.chatType, chat.projectID))

    return chat, nil
}

// AddParticipant adds contact or agent to chat
func (c *Chat) AddParticipant(participantID uuid.UUID, participantType ParticipantType) error {
    if c.status == ChatStatusClosed {
        return errors.New("cannot add participant to closed chat")
    }

    // Check if already participant
    for _, p := range c.participants {
        if p.ID == participantID {
            return errors.New("participant already in chat")
        }
    }

    // Individual chats can only have 1 contact (+ optional agent)
    if c.chatType == ChatTypeIndividual {
        if participantType == ParticipantTypeContact && c.hasContactParticipant() {
            return errors.New("individual chat can only have one contact")
        }
    }

    now := time.Now()
    participant := Participant{
        ID:       participantID,
        Type:     participantType,
        JoinedAt: now,
        IsAdmin:  false,
    }

    c.participants = append(c.participants, participant)
    c.updatedAt = now

    c.addEvent(NewParticipantAddedEvent(c.id, participantID, participantType))

    return nil
}

// RemoveParticipant removes participant from chat
func (c *Chat) RemoveParticipant(participantID uuid.UUID) error {
    if c.chatType == ChatTypeIndividual {
        return errors.New("cannot remove participant from individual chat")
    }

    found := false
    newParticipants := []Participant{}
    for _, p := range c.participants {
        if p.ID != participantID {
            newParticipants = append(newParticipants, p)
        } else {
            found = true
        }
    }

    if !found {
        return errors.New("participant not in chat")
    }

    c.participants = newParticipants
    c.updatedAt = time.Now()

    c.addEvent(NewParticipantRemovedEvent(c.id, participantID))

    return nil
}

// Archive archives chat
func (c *Chat) Archive() {
    c.status = ChatStatusArchived
    c.updatedAt = time.Now()

    c.addEvent(NewChatArchivedEvent(c.id))
}

// Unarchive unarchives chat
func (c *Chat) Unarchive() {
    c.status = ChatStatusActive
    c.updatedAt = time.Now()

    c.addEvent(NewChatUnarchivedEvent(c.id))
}

// Close closes chat (permanent)
func (c *Chat) Close() {
    c.status = ChatStatusClosed
    c.updatedAt = time.Now()

    c.addEvent(NewChatClosedEvent(c.id))
}

// UpdateLastMessageAt updates last message timestamp
func (c *Chat) UpdateLastMessageAt(timestamp time.Time) {
    c.lastMessageAt = &timestamp
    c.updatedAt = time.Now()
}

// UpdateSubject updates group/channel subject
func (c *Chat) UpdateSubject(subject string) error {
    if c.chatType == ChatTypeIndividual {
        return errors.New("individual chats don't have subjects")
    }
    c.subject = &subject
    c.updatedAt = time.Now()
    return nil
}

// UpdateDescription updates group/channel description
func (c *Chat) UpdateDescription(description string) error {
    if c.chatType == ChatTypeIndividual {
        return errors.New("individual chats don't have descriptions")
    }
    c.description = &description
    c.updatedAt = time.Now()
    return nil
}

// IsParticipant checks if ID is a participant
func (c *Chat) IsParticipant(participantID uuid.UUID) bool {
    for _, p := range c.participants {
        if p.ID == participantID {
            return true
        }
    }
    return false
}

// GetContactParticipants returns only contact participants
func (c *Chat) GetContactParticipants() []Participant {
    contacts := []Participant{}
    for _, p := range c.participants {
        if p.Type == ParticipantTypeContact {
            contacts = append(contacts, p)
        }
    }
    return contacts
}

// GetAgentParticipants returns only agent participants
func (c *Chat) GetAgentParticipants() []Participant {
    agents := []Participant{}
    for _, p := range c.participants {
        if p.Type == ParticipantTypeAgent {
            agents = append(agents, p)
        }
    }
    return agents
}

// hasContactParticipant checks if chat has any contact participant
func (c *Chat) hasContactParticipant() bool {
    for _, p := range c.participants {
        if p.Type == ParticipantTypeContact {
            return true
        }
    }
    return false
}

// Getters
func (c *Chat) ID() uuid.UUID                   { return c.id }
func (c *Chat) ProjectID() uuid.UUID            { return c.projectID }
func (c *Chat) TenantID() string                { return c.tenantID }
func (c *Chat) ChatType() ChatType              { return c.chatType }
func (c *Chat) Subject() *string                { return c.subject }
func (c *Chat) Description() *string            { return c.description }
func (c *Chat) Participants() []Participant     { return append([]Participant{}, c.participants...) }
func (c *Chat) Status() ChatStatus              { return c.status }
func (c *Chat) Metadata() map[string]interface{} { return c.metadata }
func (c *Chat) LastMessageAt() *time.Time       { return c.lastMessageAt }
func (c *Chat) CreatedAt() time.Time            { return c.createdAt }
func (c *Chat) UpdatedAt() time.Time            { return c.updatedAt }

func (c *Chat) DomainEvents() []DomainEvent {
    return append([]DomainEvent{}, c.events...)
}

func (c *Chat) ClearEvents() {
    c.events = []DomainEvent{}
}

func (c *Chat) addEvent(event DomainEvent) {
    c.events = append(c.events, event)
}
```

---

### Value Objects

#### ChatType

```go
package chat

type ChatType string

const (
    ChatTypeIndividual ChatType = "individual" // 1-on-1 chat
    ChatTypeGroup      ChatType = "group"      // WhatsApp group, Telegram group
    ChatTypeChannel    ChatType = "channel"    // Telegram channel, WhatsApp Business broadcast
)

func (ct ChatType) IsValid() bool {
    switch ct {
    case ChatTypeIndividual, ChatTypeGroup, ChatTypeChannel:
        return true
    default:
        return false
    }
}

func (ct ChatType) String() string {
    return string(ct)
}
```

#### ChatStatus

```go
package chat

type ChatStatus string

const (
    ChatStatusActive   ChatStatus = "active"   // Active conversation
    ChatStatusArchived ChatStatus = "archived" // Archived (hidden but can be reopened)
    ChatStatusClosed   ChatStatus = "closed"   // Closed (permanent - historical only)
)

func (cs ChatStatus) IsValid() bool {
    switch cs {
    case ChatStatusActive, ChatStatusArchived, ChatStatusClosed:
        return true
    default:
        return false
    }
}

func (cs ChatStatus) String() string {
    return string(cs)
}
```

#### Participant

```go
package chat

import (
    "time"

    "github.com/google/uuid"
)

type Participant struct {
    ID       uuid.UUID       // Contact ID or Agent ID
    Type     ParticipantType // contact or agent
    JoinedAt time.Time       // When joined the chat
    LeftAt   *time.Time      // When left (for groups/channels)
    IsAdmin  bool            // Is admin/moderator (for groups)
}

type ParticipantType string

const (
    ParticipantTypeContact ParticipantType = "contact"
    ParticipantTypeAgent   ParticipantType = "agent"
)

func (pt ParticipantType) IsValid() bool {
    switch pt {
    case ParticipantTypeContact, ParticipantTypeAgent:
        return true
    default:
        return false
    }
}
```

---

### Business Invariants

1. **Chat must belong to project**
   - `projectID` and `tenantID` required

2. **Chat type determines structure**
   - **Individual**: Exactly 1 contact + 0 or more agents
   - **Group**: 1 or more participants (contacts + agents)
   - **Channel**: 0 or more participants (broadcast)

3. **Subject required for groups/channels**
   - Individual chats: subject = null
   - Group/Channel: subject required

4. **Status lifecycle**
   - Created as `active`
   - Can be `archived` (reversible)
   - Can be `closed` (permanent - read-only)

5. **Participants**
   - Cannot add duplicate participants
   - Cannot remove participants from individual chats
   - Individual chats limited to 1 contact participant

---

## Events Emitted

| Event | When | Purpose |
|-------|------|---------|
| `chat.created` | New chat created | Initialize chat resources |
| `chat.participant_added` | Participant added to group/channel | Track membership |
| `chat.participant_removed` | Participant removed from group/channel | Update membership |
| `chat.archived` | Chat archived | Hide from active list |
| `chat.unarchived` | Chat brought back to active | Show in active list |
| `chat.closed` | Chat permanently closed | Mark as historical |
| `chat.subject_updated` | Group/channel name changed | Update UI |
| `chat.description_updated` | Group/channel description changed | Update metadata |
| `chat.last_message_updated` | New message in chat | Update timestamp |

---

## Repository Interface

```go
package chat

import (
    "context"

    "github.com/google/uuid"
)

type Repository interface {
    // Create new chat
    Create(ctx context.Context, chat *Chat) error

    // Find by ID
    FindByID(ctx context.Context, id uuid.UUID) (*Chat, error)

    // Find all chats for project
    FindByProject(ctx context.Context, projectID uuid.UUID) ([]*Chat, error)

    // Find chats by tenant
    FindByTenant(ctx context.Context, tenantID string) ([]*Chat, error)

    // Find chats for contact (across all chats where contact is participant)
    FindByContact(ctx context.Context, contactID uuid.UUID) ([]*Chat, error)

    // Find active chats (not archived, not closed)
    FindActiveByProject(ctx context.Context, projectID uuid.UUID) ([]*Chat, error)

    // Find individual chat between contact (for direct messages)
    FindIndividualByContact(ctx context.Context, contactID uuid.UUID, projectID uuid.UUID) (*Chat, error)

    // Update chat
    Update(ctx context.Context, chat *Chat) error

    // Delete chat (soft delete)
    Delete(ctx context.Context, id uuid.UUID) error

    // Search chats by subject
    SearchBySubject(ctx context.Context, tenantID string, subject string) ([]*Chat, error)
}
```

---

## Commands (CQRS)

### ❌ To Be Implemented

1. **CreateIndividualChatCommand** - Create 1-on-1 chat
2. **CreateGroupChatCommand** - Create group chat
3. **CreateChannelChatCommand** - Create channel/broadcast
4. **AddParticipantCommand** - Add participant to group/channel
5. **RemoveParticipantCommand** - Remove participant from group/channel
6. **ArchiveChatCommand** - Archive chat
7. **UnarchiveChatCommand** - Unarchive chat
8. **CloseChatCommand** - Permanently close chat
9. **UpdateChatSubjectCommand** - Update group/channel name
10. **UpdateChatDescriptionCommand** - Update group/channel description

---

## Use Cases

### ❌ To Be Implemented

1. **CreateIndividualChatUseCase** - Create or get existing individual chat
   - Check if chat already exists (same contact + project)
   - Create if doesn't exist
   - Return existing chat if found

2. **CreateGroupChatUseCase** - Create group chat
   - Validate group name
   - Create chat with creator as admin
   - Emit chat.created event

3. **AddParticipantToChatUseCase** - Add participant to group/channel
   - Validate chat type (not individual)
   - Check participant not already in chat
   - Add participant
   - Emit participant_added event

4. **RemoveParticipantFromChatUseCase** - Remove participant from group/channel
   - Validate chat type (not individual)
   - Validate participant exists
   - Remove participant
   - Emit participant_removed event

5. **ArchiveChatUseCase** - Archive chat
   - Validate chat is active
   - Archive chat
   - Emit chat.archived event

6. **CloseChatUseCase** - Close chat permanently
   - Validate chat not already closed
   - Close chat
   - Emit chat.closed event

7. **SearchChatsUseCase** - Search chats by subject/participants

8. **ListChatMessagesUseCase** - List all messages for chat
   - Paginated
   - Ordered by timestamp

---

## Integration with Message Aggregate

### Proposed Changes to Message

```go
// internal/domain/message/message.go (UPDATED)

type Message struct {
    id        uuid.UUID
    chatID    uuid.UUID   // NEW - Required link to chat
    contactID uuid.UUID
    sessionID *uuid.UUID  // Optional - can be nil for historical messages
    agentID   *uuid.UUID  // Optional - can be nil for system messages
    // ... rest of fields
}

// NewMessage updated to require chatID
func NewMessage(
    chatID uuid.UUID,      // NEW - Required
    contactID uuid.UUID,
    projectID uuid.UUID,
    customerID uuid.UUID,
    contentType ContentType,
    fromMe bool,
) (*Message, error) {
    if chatID == uuid.Nil {
        return nil, errors.New("chatID cannot be nil")
    }
    // ... rest of validation
}

// AssignAgent - allows assigning agent after message creation
// Useful when importing historical messages
func (m *Message) AssignAgent(agentID uuid.UUID) {
    m.agentID = &agentID
    m.addEvent(NewMessageAgentAssignedEvent(m, agentID))
}

// MarkAsSystem - marks message as system-generated
func (m *Message) MarkAsSystem() {
    m.agentID = nil // nil means "system"
    m.addEvent(NewMessageMarkedAsSystemEvent(m))
}
```

---

## Real-World Usage

### Scenario 1: Individual Chat (1-on-1 WhatsApp)

```go
// Contact sends first message via WhatsApp
contactID := uuid.MustParse("contact-uuid")
projectID := uuid.MustParse("project-uuid")

// 1. Check if individual chat exists
chat, err := chatRepo.FindIndividualByContact(ctx, contactID, projectID)
if err != nil || chat == nil {
    // 2. Create individual chat
    chat, _ = chat.NewIndividualChat(projectID, "tenant-123", contactID)
    chatRepo.Create(ctx, chat)
}

// 3. Create message in chat
message, _ := message.NewMessage(
    chat.ID(),      // chatID
    contactID,
    projectID,
    customerID,
    message.ContentTypeText,
    false,          // inbound
)
message.SetText("Hello! I need help.")

// 4. Update chat last message timestamp
chat.UpdateLastMessageAt(message.Timestamp())
chatRepo.Update(ctx, chat)
```

### Scenario 2: Group Chat (WhatsApp Group)

```go
// Create group chat
groupChat, _ := chat.NewGroupChat(
    projectID,
    "tenant-123",
    "Customer Support Team",  // Group name
    creatorContactID,
)
chatRepo.Create(ctx, groupChat)

// Add participants (contacts)
groupChat.AddParticipant(contact2ID, chat.ParticipantTypeContact)
groupChat.AddParticipant(contact3ID, chat.ParticipantTypeContact)

// Add agent to group
groupChat.AddParticipant(agentID, chat.ParticipantTypeAgent)
chatRepo.Update(ctx, groupChat)

// Message in group
message, _ := message.NewMessage(
    groupChat.ID(),  // chatID
    contact2ID,      // who sent
    projectID,
    customerID,
    message.ContentTypeText,
    false,
)
message.SetText("When will my order arrive?")
message.AssignAgent(agentID)  // Assign to specific agent
```

### Scenario 3: Channel/Broadcast (WhatsApp Business)

```go
// Create channel/broadcast
channel, _ := chat.NewChannelChat(
    projectID,
    "tenant-123",
    "Product Updates Channel",
)
chatRepo.Create(ctx, channel)

// Send broadcast message (from business to all subscribers)
message, _ := message.NewMessage(
    channel.ID(),     // chatID
    systemContactID,  // system
    projectID,
    customerID,
    message.ContentTypeText,
    true,             // outbound
)
message.SetText("🎉 New product launch!")
message.MarkAsSystem()  // System message (no agent)
```

### Scenario 4: Archive Old Chat

```go
// User archives chat
chat, _ := chatRepo.FindByID(ctx, chatID)
chat.Archive()
chatRepo.Update(ctx, chat)

// Later: unarchive
chat.Unarchive()
chatRepo.Update(ctx, chat)
```

### Scenario 5: Historical Message Import

```go
// Import old WhatsApp messages (no agent initially)
chat, _ := chatRepo.FindIndividualByContact(ctx, contactID, projectID)

for _, wahaMsg := range historicalMessages {
    message, _ := message.NewMessage(
        chat.ID(),
        contactID,
        projectID,
        customerID,
        message.ContentTypeText,
        false,
    )
    message.SetText(wahaMsg.Body)
    // No agent assigned initially (historical import)
    messageRepo.Create(ctx, message)
}

// Later: agent reviews and takes ownership
message, _ := messageRepo.FindByID(ctx, messageID)
message.AssignAgent(agentID)
messageRepo.Update(ctx, message)
```

---

## Database Schema

### Chats Table

```sql
-- 000043_create_chats.up.sql
CREATE TABLE chats (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    tenant_id TEXT NOT NULL,
    chat_type TEXT NOT NULL CHECK (chat_type IN ('individual', 'group', 'channel')),
    subject TEXT,
    description TEXT,
    participants JSONB NOT NULL DEFAULT '[]',
    status TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'archived', 'closed')),
    metadata JSONB NOT NULL DEFAULT '{}',
    last_message_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_chats_project ON chats(project_id);
CREATE INDEX idx_chats_tenant ON chats(tenant_id);
CREATE INDEX idx_chats_status ON chats(status);
CREATE INDEX idx_chats_type ON chats(chat_type);
CREATE INDEX idx_chats_last_message ON chats(last_message_at DESC);

-- GIN index for participants search
CREATE INDEX idx_chats_participants ON chats USING gin(participants jsonb_path_ops);

-- RLS (Row-Level Security)
ALTER TABLE chats ENABLE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation ON chats
    FOR ALL
    USING (tenant_id = current_setting('app.current_tenant_id')::text);
```

### Migration: Add chat_id to messages

```sql
-- 000044_add_chat_id_to_messages.up.sql

-- Step 1: Add chat_id column (nullable initially)
ALTER TABLE messages ADD COLUMN chat_id UUID REFERENCES chats(id) ON DELETE CASCADE;

-- Step 2: Create index
CREATE INDEX idx_messages_chat ON messages(chat_id);

-- Step 3: Data migration (create individual chats for existing messages)
-- This is a critical step - map existing messages to chats

DO $$
DECLARE
    msg RECORD;
    new_chat_id UUID;
BEGIN
    -- For each unique contact+project combination, create individual chat
    FOR msg IN
        SELECT DISTINCT contact_id, project_id, tenant_id
        FROM messages
        WHERE chat_id IS NULL
    LOOP
        -- Create individual chat
        INSERT INTO chats (project_id, tenant_id, chat_type, participants, status)
        VALUES (
            msg.project_id,
            msg.tenant_id,
            'individual',
            jsonb_build_array(
                jsonb_build_object(
                    'id', msg.contact_id,
                    'type', 'contact',
                    'joined_at', NOW(),
                    'is_admin', false
                )
            ),
            'active'
        )
        RETURNING id INTO new_chat_id;

        -- Update all messages for this contact+project
        UPDATE messages
        SET chat_id = new_chat_id
        WHERE contact_id = msg.contact_id
          AND project_id = msg.project_id
          AND chat_id IS NULL;
    END LOOP;
END $$;

-- Step 4: Make chat_id NOT NULL (after migration)
ALTER TABLE messages ALTER COLUMN chat_id SET NOT NULL;
```

```sql
-- 000044_add_chat_id_to_messages.down.sql
ALTER TABLE messages DROP COLUMN chat_id;
```

---

## API Examples

### Create Individual Chat

```http
POST /api/v1/chats
Authorization: Bearer {token}
Content-Type: application/json

{
  "type": "individual",
  "contact_id": "contact-uuid",
  "project_id": "project-uuid"
}

Response (201 Created):
{
  "id": "chat-uuid",
  "project_id": "project-uuid",
  "tenant_id": "tenant-123",
  "chat_type": "individual",
  "participants": [
    {
      "id": "contact-uuid",
      "type": "contact",
      "joined_at": "2025-10-10T10:00:00Z",
      "is_admin": false
    }
  ],
  "status": "active",
  "created_at": "2025-10-10T10:00:00Z",
  "updated_at": "2025-10-10T10:00:00Z"
}
```

### Create Group Chat

```http
POST /api/v1/chats
Authorization: Bearer {token}
Content-Type: application/json

{
  "type": "group",
  "subject": "Customer Support Team",
  "description": "General support group",
  "creator_id": "contact-uuid",
  "project_id": "project-uuid"
}

Response (201 Created):
{
  "id": "chat-uuid",
  "project_id": "project-uuid",
  "tenant_id": "tenant-123",
  "chat_type": "group",
  "subject": "Customer Support Team",
  "description": "General support group",
  "participants": [
    {
      "id": "contact-uuid",
      "type": "contact",
      "joined_at": "2025-10-10T10:00:00Z",
      "is_admin": true
    }
  ],
  "status": "active",
  "created_at": "2025-10-10T10:00:00Z"
}
```

### Add Participant to Group

```http
POST /api/v1/chats/{chat_id}/participants
Authorization: Bearer {token}
Content-Type: application/json

{
  "participant_id": "contact-2-uuid",
  "participant_type": "contact"
}

Response (200 OK):
{
  "chat_id": "chat-uuid",
  "participant_id": "contact-2-uuid",
  "participant_type": "contact",
  "joined_at": "2025-10-10T10:05:00Z"
}
```

### Remove Participant from Group

```http
DELETE /api/v1/chats/{chat_id}/participants/{participant_id}
Authorization: Bearer {token}

Response (204 No Content)
```

### Archive Chat

```http
POST /api/v1/chats/{chat_id}/archive
Authorization: Bearer {token}

Response (200 OK):
{
  "chat_id": "chat-uuid",
  "status": "archived",
  "updated_at": "2025-10-10T10:10:00Z"
}
```

### List Chats

```http
GET /api/v1/chats?project_id={project_id}&status=active&limit=50
Authorization: Bearer {token}

Response (200 OK):
{
  "chats": [
    {
      "id": "chat-uuid-1",
      "chat_type": "individual",
      "participants": [...],
      "status": "active",
      "last_message_at": "2025-10-10T09:50:00Z"
    },
    {
      "id": "chat-uuid-2",
      "chat_type": "group",
      "subject": "Team Chat",
      "participants": [...],
      "status": "active",
      "last_message_at": "2025-10-10T09:55:00Z"
    }
  ],
  "total": 2
}
```

### List Messages in Chat

```http
GET /api/v1/chats/{chat_id}/messages?limit=100&offset=0
Authorization: Bearer {token}

Response (200 OK):
{
  "chat_id": "chat-uuid",
  "messages": [
    {
      "id": "message-uuid-1",
      "chat_id": "chat-uuid",
      "contact_id": "contact-uuid",
      "agent_id": null,
      "text": "Hello!",
      "timestamp": "2025-10-10T09:50:00Z",
      "from_me": false
    },
    {
      "id": "message-uuid-2",
      "chat_id": "chat-uuid",
      "contact_id": "contact-uuid",
      "agent_id": "agent-uuid",
      "text": "How can I help you?",
      "timestamp": "2025-10-10T09:51:00Z",
      "from_me": true
    }
  ],
  "total": 2
}
```

---

## Performance Considerations

### Indexes

```sql
-- Chats table
CREATE INDEX idx_chats_project ON chats(project_id);
CREATE INDEX idx_chats_tenant ON chats(tenant_id);
CREATE INDEX idx_chats_status ON chats(status);
CREATE INDEX idx_chats_type ON chats(chat_type);
CREATE INDEX idx_chats_last_message ON chats(last_message_at DESC);

-- GIN index for participant search
CREATE INDEX idx_chats_participants ON chats USING gin(participants jsonb_path_ops);

-- Composite indexes
CREATE INDEX idx_chats_tenant_status ON chats(tenant_id, status);
CREATE INDEX idx_chats_project_type ON chats(project_id, chat_type);
```

### Caching Strategy

```go
// Cache chat by ID (10 min TTL)
cacheKey := fmt.Sprintf("chat:%s", chatID)
chat, err := cache.Get(cacheKey)

// Cache individual chat by contact (5 min TTL)
cacheKey := fmt.Sprintf("chat:individual:contact:%s:project:%s", contactID, projectID)
chat, err := cache.Get(cacheKey)

// Cache chat participants (5 min TTL)
cacheKey := fmt.Sprintf("chat:%s:participants", chatID)
participants, err := cache.Get(cacheKey)
```

### Query Optimization

```go
// Find active chats with last message in last 7 days
// Uses idx_chats_last_message for performance
SELECT *
FROM chats
WHERE status = 'active'
  AND last_message_at > NOW() - INTERVAL '7 days'
ORDER BY last_message_at DESC
LIMIT 50;

// Find chats where contact is participant
// Uses GIN index on participants
SELECT *
FROM chats
WHERE participants @> '[{"id": "contact-uuid"}]'::jsonb;
```

---

## Implementation Roadmap

### Phase 1: Domain Model (2-3 days)
- [ ] Create `internal/domain/chat/` directory
- [ ] Implement Chat aggregate (`chat.go`)
- [ ] Implement value objects (`chat_type.go`, `chat_status.go`, `participant.go`)
- [ ] Implement domain events (`events.go`)
- [ ] Implement errors (`errors.go`)
- [ ] Implement repository interface (`repository.go`)
- [ ] Write comprehensive unit tests (`chat_test.go`)

### Phase 2: Infrastructure (2-3 days)
- [ ] Create GORM entity (`infrastructure/persistence/entities/chat.go`)
- [ ] Implement GORM repository (`infrastructure/persistence/gorm_chat_repository.go`)
- [ ] Create database migration (create_chats table)
- [ ] Create database migration (add chat_id to messages)
- [ ] Run data migration (create individual chats for existing messages)
- [ ] Add repository tests

### Phase 3: Application Layer (2-3 days)
- [ ] Implement use cases (`internal/application/chat/`)
  - `create_individual_chat.go`
  - `create_group_chat.go`
  - `add_participant.go`
  - `remove_participant.go`
  - `archive_chat.go`
  - `list_chats.go`
- [ ] Implement DTOs (`internal/application/chat/dto.go`)
- [ ] Write use case tests

### Phase 4: HTTP Layer (1-2 days)
- [ ] Implement HTTP handler (`infrastructure/http/handlers/chat_handler.go`)
- [ ] Add routes (`infrastructure/http/routes/routes.go`)
- [ ] Add Swagger documentation
- [ ] Test API endpoints

### Phase 5: Integration (1-2 days)
- [ ] Update Message aggregate to require `chatID`
- [ ] Update message creation flows to get/create chat
- [ ] Update WAHA webhook handler to handle group messages
- [ ] Update message repository queries to include chat_id
- [ ] Integration tests

### Phase 6: Advanced Features (2-3 days)
- [ ] Chat search (by subject, participants)
- [ ] Chat analytics (message count, active participants)
- [ ] Chat export (PDF, CSV)
- [ ] Mute/unmute chat notifications
- [ ] Pin/unpin chats

---

## Migration Strategy

### For Existing Messages

```go
// internal/application/chat/migrate_existing_messages.go

type MigrateExistingMessagesUseCase struct {
    chatRepo    chat.Repository
    messageRepo message.Repository
}

func (uc *MigrateExistingMessagesUseCase) Execute(ctx context.Context) error {
    // 1. Get all unique contact+project combinations from messages
    combinations, err := uc.messageRepo.GetUniqueContactProjectCombinations(ctx)
    if err != nil {
        return err
    }

    // 2. For each combination, create individual chat
    for _, combo := range combinations {
        // Check if chat already exists
        existingChat, _ := uc.chatRepo.FindIndividualByContact(ctx, combo.ContactID, combo.ProjectID)
        if existingChat != nil {
            continue // Skip if already exists
        }

        // Create individual chat
        newChat, err := chat.NewIndividualChat(combo.ProjectID, combo.TenantID, combo.ContactID)
        if err != nil {
            return fmt.Errorf("failed to create chat: %w", err)
        }

        // Save chat
        if err := uc.chatRepo.Create(ctx, newChat); err != nil {
            return fmt.Errorf("failed to save chat: %w", err)
        }

        // Update all messages for this contact+project
        if err := uc.messageRepo.UpdateChatIDForContactAndProject(ctx, newChat.ID(), combo.ContactID, combo.ProjectID); err != nil {
            return fmt.Errorf("failed to update messages: %w", err)
        }
    }

    return nil
}
```

---

## References

**DDD Resources**:
- [Domain-Driven Design - Chat Aggregate](https://stackoverflow.com/questions/57650529/should-user-chat-message-be-an-aggregate)
- [DDD Group Modeling](https://stackoverflow.com/questions/21755154/how-to-model-users-and-groups-in-ddd)
- [DDD Saga Orchestration - Chat Example](https://github.com/vladovsiychuk/saga-orchestrator-ddd-chat)

**System Design**:
- [WhatsApp System Design](https://www.hellointerview.com/learn/system-design/problem-breakdowns/whatsapp)
- [Chat Application Architecture](https://medium.com/@m.romaniiuk/system-design-chat-application-1d6fbf21b372)

---

**Next**: [Note Aggregate](note_aggregate.md) →
**Previous**: [Webhook Aggregate](webhook_aggregate.md) ←

---

## Summary

✅ **Chat Aggregate Design**:
1. **Three types** - Individual (1-on-1), Group (WhatsApp groups), Channel (broadcasts)
2. **Participant management** - Contacts + Agents with roles (admin/member)
3. **Status lifecycle** - Active, Archived, Closed
4. **Message context** - All messages belong to a chat
5. **Multi-channel support** - Same contact can have multiple chats on different channels

❌ **Implementation Status**: **NOT IMPLEMENTED YET** - Complete design document ready for implementation.

**Why CRITICAL**: Currently messages have NO chat context, which prevents:
- WhatsApp group support
- Multiple conversations with same contact
- Chat-level operations (archive, mute)
- Historical message import with context
- Proper participant tracking

**Next Steps**: Implement Phase 1 (Domain Model) → Phase 2 (Infrastructure) → Phase 3 (Application Layer).

**Estimated Effort**: 10-15 days (full implementation + testing + migration).
