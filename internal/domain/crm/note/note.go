package note

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type Note struct {
	id        uuid.UUID
	contactID uuid.UUID
	sessionID *uuid.UUID
	tenantID  string

	authorID   uuid.UUID
	authorType AuthorType
	authorName string

	content  string
	noteType NoteType
	priority Priority

	visibleToClient bool
	pinned          bool

	tags        []string
	mentions    []uuid.UUID
	attachments []string

	createdAt time.Time
	updatedAt time.Time
	deletedAt *time.Time

	events []DomainEvent
}

type AuthorType string

const (
	AuthorTypeAgent  AuthorType = "agent"
	AuthorTypeSystem AuthorType = "system"
	AuthorTypeUser   AuthorType = "user"
)

type NoteType string

const (
	NoteTypeGeneral         NoteType = "general"
	NoteTypeFollowUp        NoteType = "follow_up"
	NoteTypeComplaint       NoteType = "complaint"
	NoteTypeResolution      NoteType = "resolution"
	NoteTypeEscalation      NoteType = "escalation"
	NoteTypeInternal        NoteType = "internal"
	NoteTypeCustomer        NoteType = "customer"
	NoteTypeSessionSummary  NoteType = "session_summary"
	NoteTypeSessionHandoff  NoteType = "session_handoff"
	NoteTypeSessionFeedback NoteType = "session_feedback"
	NoteTypeAdConversion    NoteType = "ad_conversion"
	NoteTypeAdCampaign      NoteType = "ad_campaign"
	NoteTypeAdAttribution   NoteType = "ad_attribution"
	NoteTypeTrackingInsight NoteType = "tracking_insight"
)

type Priority string

const (
	PriorityLow    Priority = "low"
	PriorityNormal Priority = "normal"
	PriorityHigh   Priority = "high"
	PriorityUrgent Priority = "urgent"
)

var (
	ErrEmptyContent   = errors.New("note content cannot be empty")
	ErrInvalidContact = errors.New("invalid contact ID")
	ErrInvalidAuthor  = errors.New("invalid author ID")
	ErrNoteNotFound   = errors.New("note not found")
)

func NewNote(
	contactID uuid.UUID,
	tenantID string,
	authorID uuid.UUID,
	authorType AuthorType,
	authorName string,
	content string,
	noteType NoteType,
) (*Note, error) {
	if contactID == uuid.Nil {
		return nil, ErrInvalidContact
	}
	if authorID == uuid.Nil {
		return nil, ErrInvalidAuthor
	}
	if content == "" {
		return nil, ErrEmptyContent
	}

	now := time.Now()
	note := &Note{
		id:              uuid.New(),
		contactID:       contactID,
		tenantID:        tenantID,
		authorID:        authorID,
		authorType:      authorType,
		authorName:      authorName,
		content:         content,
		noteType:        noteType,
		priority:        PriorityNormal,
		visibleToClient: false,
		pinned:          false,
		tags:            []string{},
		mentions:        []uuid.UUID{},
		attachments:     []string{},
		createdAt:       now,
		updatedAt:       now,
		events:          []DomainEvent{},
	}

	note.addEvent(NewNoteAddedEvent(note.id, contactID, note.sessionID, tenantID, authorID, authorType, authorName, content, noteType, note.priority))

	return note, nil
}

func ReconstructNote(
	id uuid.UUID,
	contactID uuid.UUID,
	sessionID *uuid.UUID,
	tenantID string,
	authorID uuid.UUID,
	authorType AuthorType,
	authorName string,
	content string,
	noteType NoteType,
	priority Priority,
	visibleToClient bool,
	pinned bool,
	tags []string,
	mentions []uuid.UUID,
	attachments []string,
	createdAt time.Time,
	updatedAt time.Time,
	deletedAt *time.Time,
) *Note {
	if tags == nil {
		tags = []string{}
	}
	if mentions == nil {
		mentions = []uuid.UUID{}
	}
	if attachments == nil {
		attachments = []string{}
	}

	return &Note{
		id:              id,
		contactID:       contactID,
		sessionID:       sessionID,
		tenantID:        tenantID,
		authorID:        authorID,
		authorType:      authorType,
		authorName:      authorName,
		content:         content,
		noteType:        noteType,
		priority:        priority,
		visibleToClient: visibleToClient,
		pinned:          pinned,
		tags:            tags,
		mentions:        mentions,
		attachments:     attachments,
		createdAt:       createdAt,
		updatedAt:       updatedAt,
		deletedAt:       deletedAt,
		events:          []DomainEvent{},
	}
}

func (n *Note) AttachToSession(sessionID uuid.UUID) {
	n.sessionID = &sessionID
	n.updatedAt = time.Now()
}

func (n *Note) UpdateContent(content string, updatedBy uuid.UUID) error {
	if content == "" {
		return ErrEmptyContent
	}

	oldContent := n.content
	n.content = content
	n.updatedAt = time.Now()

	n.addEvent(NewNoteUpdatedEvent(n.id, n.contactID, n.tenantID, updatedBy, oldContent, content))

	return nil
}

func (n *Note) SetPriority(priority Priority) {
	n.priority = priority
	n.updatedAt = time.Now()
}

func (n *Note) SetVisibility(visible bool) {
	n.visibleToClient = visible
	n.updatedAt = time.Now()
}

func (n *Note) Pin(pinnedBy uuid.UUID) {
	if !n.pinned {
		n.pinned = true
		n.updatedAt = time.Now()

		n.addEvent(NewNotePinnedEvent(n.id, n.contactID, n.tenantID, pinnedBy))
	}
}

func (n *Note) Unpin() {
	n.pinned = false
	n.updatedAt = time.Now()
}

func (n *Note) AddTag(tag string) {
	if tag != "" {
		n.tags = append(n.tags, tag)
		n.updatedAt = time.Now()
	}
}

func (n *Note) RemoveTag(tag string) {
	for i, t := range n.tags {
		if t == tag {
			n.tags = append(n.tags[:i], n.tags[i+1:]...)
			n.updatedAt = time.Now()
			break
		}
	}
}

func (n *Note) MentionAgent(agentID uuid.UUID) {
	if agentID != uuid.Nil {
		n.mentions = append(n.mentions, agentID)
		n.updatedAt = time.Now()
	}
}

func (n *Note) AddAttachment(url string) {
	if url != "" {
		n.attachments = append(n.attachments, url)
		n.updatedAt = time.Now()
	}
}

func (n *Note) Delete(deletedBy uuid.UUID) {
	if n.deletedAt == nil {
		now := time.Now()
		n.deletedAt = &now
		n.updatedAt = now

		n.addEvent(NewNoteDeletedEvent(n.id, n.contactID, n.tenantID, deletedBy))
	}
}

func (n *Note) IsDeleted() bool {
	return n.deletedAt != nil
}

func (n *Note) ID() uuid.UUID          { return n.id }
func (n *Note) ContactID() uuid.UUID   { return n.contactID }
func (n *Note) SessionID() *uuid.UUID  { return n.sessionID }
func (n *Note) TenantID() string       { return n.tenantID }
func (n *Note) AuthorID() uuid.UUID    { return n.authorID }
func (n *Note) AuthorType() AuthorType { return n.authorType }
func (n *Note) AuthorName() string     { return n.authorName }
func (n *Note) Content() string        { return n.content }
func (n *Note) NoteType() NoteType     { return n.noteType }
func (n *Note) Priority() Priority     { return n.priority }
func (n *Note) VisibleToClient() bool  { return n.visibleToClient }
func (n *Note) Pinned() bool           { return n.pinned }
func (n *Note) Tags() []string         { return append([]string{}, n.tags...) }
func (n *Note) Mentions() []uuid.UUID  { return append([]uuid.UUID{}, n.mentions...) }
func (n *Note) Attachments() []string  { return append([]string{}, n.attachments...) }
func (n *Note) CreatedAt() time.Time   { return n.createdAt }
func (n *Note) UpdatedAt() time.Time   { return n.updatedAt }
func (n *Note) DeletedAt() *time.Time  { return n.deletedAt }

func (n *Note) DomainEvents() []DomainEvent {
	return append([]DomainEvent{}, n.events...)
}

func (n *Note) ClearEvents() {
	n.events = []DomainEvent{}
}

func (n *Note) addEvent(event DomainEvent) {
	n.events = append(n.events, event)
}

type DomainEvent interface {
	EventName() string
	EventID() uuid.UUID
	EventVersion() string
	EventType() string
	AggregateID() uuid.UUID
	OccurredAt() time.Time
}
