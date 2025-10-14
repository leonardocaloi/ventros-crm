package chat

import (
	"time"

	"github.com/google/uuid"
	"github.com/ventros/crm/internal/domain/core/shared"
)

type Chat struct {
	id            uuid.UUID
	version       int // Optimistic locking - prevents lost updates
	projectID     uuid.UUID
	tenantID      string
	chatType      ChatType
	externalID    *string // ID externo do chat no canal (ex: grupo WhatsApp @g.us)
	subject       *string
	description   *string
	participants  []Participant
	status        ChatStatus
	metadata      map[string]interface{}
	lastMessageAt *time.Time
	createdAt     time.Time
	updatedAt     time.Time

	events []shared.DomainEvent
}

// NewIndividualChat creates a new individual (1-on-1) chat
func NewIndividualChat(
	projectID uuid.UUID,
	tenantID string,
	contactID uuid.UUID,
) (*Chat, error) {
	if projectID == uuid.Nil {
		return nil, ErrProjectIDRequired
	}
	if tenantID == "" {
		return nil, ErrTenantIDRequired
	}
	if contactID == uuid.Nil {
		return nil, ErrContactIDRequired
	}

	now := time.Now()
	chat := &Chat{
		id:        uuid.New(),
		version:   1, // Start with version 1 for new aggregates
		projectID: projectID,
		tenantID:  tenantID,
		chatType:  ChatTypeIndividual,
		participants: []Participant{
			{
				ID:       contactID,
				Type:     ParticipantTypeContact,
				JoinedAt: now,
				IsAdmin:  false,
			},
		},
		status:    ChatStatusActive,
		metadata:  make(map[string]interface{}),
		createdAt: now,
		updatedAt: now,
		events:    []shared.DomainEvent{},
	}

	chat.addEvent(NewChatCreatedEvent(chat.id, chat.chatType, chat.projectID))

	return chat, nil
}

// NewGroupChat creates a new group chat (WhatsApp group, Telegram group)
func NewGroupChat(
	projectID uuid.UUID,
	tenantID string,
	subject string,
	creatorID uuid.UUID,
	externalID *string,
) (*Chat, error) {
	if projectID == uuid.Nil {
		return nil, ErrProjectIDRequired
	}
	if tenantID == "" {
		return nil, ErrTenantIDRequired
	}
	if subject == "" {
		return nil, ErrSubjectRequired
	}
	if creatorID == uuid.Nil {
		return nil, ErrCreatorIDRequired
	}

	now := time.Now()
	chat := &Chat{
		id:         uuid.New(),
		version:    1, // Start with version 1 for new aggregates
		projectID:  projectID,
		tenantID:   tenantID,
		chatType:   ChatTypeGroup,
		externalID: externalID,
		subject:    &subject,
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
		events:    []shared.DomainEvent{},
	}

	chat.addEvent(NewChatCreatedEvent(chat.id, chat.chatType, chat.projectID))

	return chat, nil
}

// NewChannelChat creates a new channel chat (Telegram channel, WhatsApp Business broadcast)
func NewChannelChat(
	projectID uuid.UUID,
	tenantID string,
	subject string,
) (*Chat, error) {
	if projectID == uuid.Nil {
		return nil, ErrProjectIDRequired
	}
	if tenantID == "" {
		return nil, ErrTenantIDRequired
	}
	if subject == "" {
		return nil, ErrSubjectRequired
	}

	now := time.Now()
	chat := &Chat{
		id:           uuid.New(),
		version:      1, // Start with version 1 for new aggregates
		projectID:    projectID,
		tenantID:     tenantID,
		chatType:     ChatTypeChannel,
		subject:      &subject,
		participants: []Participant{}, // Channels may have no explicit participants
		status:       ChatStatusActive,
		metadata:     make(map[string]interface{}),
		createdAt:    now,
		updatedAt:    now,
		events:       []shared.DomainEvent{},
	}

	chat.addEvent(NewChatCreatedEvent(chat.id, chat.chatType, chat.projectID))

	return chat, nil
}

// ReconstructChat reconstructs a chat from persistence (for repository)
func ReconstructChat(
	id uuid.UUID,
	version int, // Optimistic locking version
	projectID uuid.UUID,
	tenantID string,
	chatType ChatType,
	externalID *string,
	subject *string,
	description *string,
	participants []Participant,
	status ChatStatus,
	metadata map[string]interface{},
	lastMessageAt *time.Time,
	createdAt time.Time,
	updatedAt time.Time,
) *Chat {
	if version == 0 {
		version = 1 // Default to version 1 (backwards compatibility)
	}
	if metadata == nil {
		metadata = make(map[string]interface{})
	}

	return &Chat{
		id:            id,
		version:       version,
		projectID:     projectID,
		tenantID:      tenantID,
		chatType:      chatType,
		externalID:    externalID,
		subject:       subject,
		description:   description,
		participants:  participants,
		status:        status,
		metadata:      metadata,
		lastMessageAt: lastMessageAt,
		createdAt:     createdAt,
		updatedAt:     updatedAt,
		events:        []shared.DomainEvent{},
	}
}

// AddParticipant adds a contact or agent to the chat
func (c *Chat) AddParticipant(participantID uuid.UUID, participantType ParticipantType) error {
	if c.status == ChatStatusClosed {
		return ErrChatClosed
	}

	// Check if already participant
	for _, p := range c.participants {
		if p.ID == participantID {
			return ErrParticipantAlreadyExists
		}
	}

	// Individual chats can only have 1 contact (+ optional agent)
	if c.chatType == ChatTypeIndividual {
		if participantType == ParticipantTypeContact && c.hasContactParticipant() {
			return ErrIndividualChatLimitReached
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

// RemoveParticipant removes a participant from the chat
func (c *Chat) RemoveParticipant(participantID uuid.UUID) error {
	if c.chatType == ChatTypeIndividual {
		return ErrCannotRemoveFromIndividual
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
		return ErrParticipantNotFound
	}

	c.participants = newParticipants
	c.updatedAt = time.Now()

	c.addEvent(NewParticipantRemovedEvent(c.id, participantID))

	return nil
}

// Archive archives the chat
func (c *Chat) Archive() {
	c.status = ChatStatusArchived
	c.updatedAt = time.Now()

	c.addEvent(NewChatArchivedEvent(c.id))
}

// Unarchive unarchives the chat
func (c *Chat) Unarchive() {
	c.status = ChatStatusActive
	c.updatedAt = time.Now()

	c.addEvent(NewChatUnarchivedEvent(c.id))
}

// Close permanently closes the chat
func (c *Chat) Close() {
	c.status = ChatStatusClosed
	c.updatedAt = time.Now()

	c.addEvent(NewChatClosedEvent(c.id))
}

// UpdateLastMessageAt updates the last message timestamp
func (c *Chat) UpdateLastMessageAt(timestamp time.Time) {
	c.lastMessageAt = &timestamp
	c.updatedAt = time.Now()
}

// UpdateSubject updates the group/channel subject
func (c *Chat) UpdateSubject(subject string) error {
	if c.chatType == ChatTypeIndividual {
		return ErrIndividualChatNoSubject
	}
	c.subject = &subject
	c.updatedAt = time.Now()

	c.addEvent(NewChatSubjectUpdatedEvent(c.id, subject))

	return nil
}

// UpdateDescription updates the group/channel description
func (c *Chat) UpdateDescription(description string) error {
	if c.chatType == ChatTypeIndividual {
		return ErrIndividualChatNoSubject
	}
	c.description = &description
	c.updatedAt = time.Now()

	c.addEvent(NewChatDescriptionUpdatedEvent(c.id, description))

	return nil
}

// IsParticipant checks if the given ID is a participant in the chat
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

// PromoteToAdmin promotes a participant to admin (for groups)
func (c *Chat) PromoteToAdmin(participantID uuid.UUID) error {
	if c.chatType != ChatTypeGroup {
		return ErrIndividualChatNoSubject // Reutilizando erro, deveria ser ErrNotAGroup
	}

	for i, p := range c.participants {
		if p.ID == participantID {
			c.participants[i].IsAdmin = true
			c.updatedAt = time.Now()
			c.addEvent(NewParticipantPromotedEvent(c.id, participantID))
			return nil
		}
	}

	return ErrParticipantNotFound
}

// DemoteFromAdmin removes admin privileges from a participant (for groups)
func (c *Chat) DemoteFromAdmin(participantID uuid.UUID) error {
	if c.chatType != ChatTypeGroup {
		return ErrIndividualChatNoSubject // Reutilizando erro
	}

	for i, p := range c.participants {
		if p.ID == participantID {
			c.participants[i].IsAdmin = false
			c.updatedAt = time.Now()
			c.addEvent(NewParticipantDemotedEvent(c.id, participantID))
			return nil
		}
	}

	return ErrParticipantNotFound
}

// UpdateExternalID updates the external ID (WhatsApp group ID, etc)
func (c *Chat) UpdateExternalID(externalID string) {
	c.externalID = &externalID
	c.updatedAt = time.Now()
}

// IsGroup checks if chat is a group
func (c *Chat) IsGroup() bool {
	return c.chatType == ChatTypeGroup
}

// Getters
func (c *Chat) ID() uuid.UUID        { return c.id }
func (c *Chat) Version() int         { return c.version }
func (c *Chat) ProjectID() uuid.UUID { return c.projectID }
func (c *Chat) TenantID() string     { return c.tenantID }
func (c *Chat) ChatType() ChatType   { return c.chatType }
func (c *Chat) ExternalID() *string  { return c.externalID }
func (c *Chat) Subject() *string     { return c.subject }
func (c *Chat) Description() *string { return c.description }
func (c *Chat) Status() ChatStatus   { return c.status }
func (c *Chat) CreatedAt() time.Time { return c.createdAt }
func (c *Chat) UpdatedAt() time.Time { return c.updatedAt }
func (c *Chat) LastMessageAt() *time.Time {
	return c.lastMessageAt
}

func (c *Chat) Participants() []Participant {
	// Return copy to prevent external modification
	return append([]Participant{}, c.participants...)
}

func (c *Chat) Metadata() map[string]interface{} {
	// Return copy to prevent external modification
	copy := make(map[string]interface{})
	for k, v := range c.metadata {
		copy[k] = v
	}
	return copy
}

// DomainEvents returns the domain events
func (c *Chat) DomainEvents() []shared.DomainEvent {
	return append([]shared.DomainEvent{}, c.events...)
}

// ClearEvents clears the domain events
func (c *Chat) ClearEvents() {
	c.events = []shared.DomainEvent{}
}

// addEvent adds a domain event
func (c *Chat) addEvent(event shared.DomainEvent) {
	c.events = append(c.events, event)
}

// Label Management Methods

// GetLabelIDs returns the list of label IDs associated with this chat
func (c *Chat) GetLabelIDs() []string {
	if c.metadata == nil {
		return []string{}
	}

	labelIDsRaw, ok := c.metadata["label_ids"]
	if !ok {
		return []string{}
	}

	// Handle different serialization formats
	switch v := labelIDsRaw.(type) {
	case []string:
		return v
	case []interface{}:
		labelIDs := make([]string, 0, len(v))
		for _, item := range v {
			if str, ok := item.(string); ok {
				labelIDs = append(labelIDs, str)
			}
		}
		return labelIDs
	default:
		return []string{}
	}
}

// SetLabelIDs sets the label IDs for this chat
func (c *Chat) SetLabelIDs(labelIDs []string) {
	if c.metadata == nil {
		c.metadata = make(map[string]interface{})
	}

	c.metadata["label_ids"] = labelIDs
	c.updatedAt = time.Now()
}

// AddLabel adds a label to the chat
func (c *Chat) AddLabel(labelID string) error {
	if labelID == "" {
		return ErrLabelIDRequired
	}

	labelIDs := c.GetLabelIDs()

	// Check if label already exists
	for _, id := range labelIDs {
		if id == labelID {
			return nil // Already has this label, no-op
		}
	}

	labelIDs = append(labelIDs, labelID)
	c.SetLabelIDs(labelIDs)

	c.addEvent(NewChatLabelAddedEvent(c.id, labelID))

	return nil
}

// RemoveLabel removes a label from the chat
func (c *Chat) RemoveLabel(labelID string) error {
	if labelID == "" {
		return ErrLabelIDRequired
	}

	labelIDs := c.GetLabelIDs()
	found := false
	newLabelIDs := make([]string, 0, len(labelIDs))

	for _, id := range labelIDs {
		if id != labelID {
			newLabelIDs = append(newLabelIDs, id)
		} else {
			found = true
		}
	}

	if !found {
		return ErrLabelNotFound
	}

	c.SetLabelIDs(newLabelIDs)

	c.addEvent(NewChatLabelRemovedEvent(c.id, labelID))

	return nil
}

// HasLabel checks if the chat has a specific label
func (c *Chat) HasLabel(labelID string) bool {
	labelIDs := c.GetLabelIDs()
	for _, id := range labelIDs {
		if id == labelID {
			return true
		}
	}
	return false
}

// ClearLabels removes all labels from the chat
func (c *Chat) ClearLabels() {
	c.SetLabelIDs([]string{})
}

// GetLabelCount returns the number of labels on this chat
func (c *Chat) GetLabelCount() int {
	return len(c.GetLabelIDs())
}

// Compile-time check that Chat implements AggregateRoot interface
var _ shared.AggregateRoot = (*Chat)(nil)
