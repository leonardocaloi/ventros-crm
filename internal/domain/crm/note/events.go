package note

import (
	"time"

	"github.com/google/uuid"
	"github.com/ventros/crm/internal/domain/core/shared"
)

type NoteAddedEvent struct {
	shared.BaseEvent
	NoteID     uuid.UUID
	ContactID  uuid.UUID
	SessionID  *uuid.UUID
	TenantID   string
	AuthorID   uuid.UUID
	AuthorType AuthorType
	AuthorName string
	Content    string
	NoteType   NoteType
	Priority   Priority
}

func NewNoteAddedEvent(noteID, contactID uuid.UUID, sessionID *uuid.UUID, tenantID string, authorID uuid.UUID, authorType AuthorType, authorName, content string, noteType NoteType, priority Priority) NoteAddedEvent {
	return NoteAddedEvent{
		BaseEvent:  shared.NewBaseEvent("note.added", time.Now()),
		NoteID:     noteID,
		ContactID:  contactID,
		SessionID:  sessionID,
		TenantID:   tenantID,
		AuthorID:   authorID,
		AuthorType: authorType,
		AuthorName: authorName,
		Content:    content,
		NoteType:   noteType,
		Priority:   priority,
	}
}

func (e NoteAddedEvent) EventType() string {
	return "note.added"
}

func (e NoteAddedEvent) AggregateID() uuid.UUID {
	return e.NoteID
}

type NoteUpdatedEvent struct {
	shared.BaseEvent
	NoteID     uuid.UUID
	ContactID  uuid.UUID
	TenantID   string
	UpdatedBy  uuid.UUID
	OldContent string
	NewContent string
}

func NewNoteUpdatedEvent(noteID, contactID uuid.UUID, tenantID string, updatedBy uuid.UUID, oldContent, newContent string) NoteUpdatedEvent {
	return NoteUpdatedEvent{
		BaseEvent:  shared.NewBaseEvent("note.updated", time.Now()),
		NoteID:     noteID,
		ContactID:  contactID,
		TenantID:   tenantID,
		UpdatedBy:  updatedBy,
		OldContent: oldContent,
		NewContent: newContent,
	}
}

func (e NoteUpdatedEvent) EventType() string {
	return "note.updated"
}

func (e NoteUpdatedEvent) AggregateID() uuid.UUID {
	return e.NoteID
}

type NoteDeletedEvent struct {
	shared.BaseEvent
	NoteID    uuid.UUID
	ContactID uuid.UUID
	TenantID  string
	DeletedBy uuid.UUID
}

func NewNoteDeletedEvent(noteID, contactID uuid.UUID, tenantID string, deletedBy uuid.UUID) NoteDeletedEvent {
	return NoteDeletedEvent{
		BaseEvent: shared.NewBaseEvent("note.deleted", time.Now()),
		NoteID:    noteID,
		ContactID: contactID,
		TenantID:  tenantID,
		DeletedBy: deletedBy,
	}
}

func (e NoteDeletedEvent) EventType() string {
	return "note.deleted"
}

func (e NoteDeletedEvent) AggregateID() uuid.UUID {
	return e.NoteID
}

type NotePinnedEvent struct {
	shared.BaseEvent
	NoteID    uuid.UUID
	ContactID uuid.UUID
	TenantID  string
	PinnedBy  uuid.UUID
}

func NewNotePinnedEvent(noteID, contactID uuid.UUID, tenantID string, pinnedBy uuid.UUID) NotePinnedEvent {
	return NotePinnedEvent{
		BaseEvent: shared.NewBaseEvent("note.pinned", time.Now()),
		NoteID:    noteID,
		ContactID: contactID,
		TenantID:  tenantID,
		PinnedBy:  pinnedBy,
	}
}

func (e NotePinnedEvent) EventType() string {
	return "note.pinned"
}

func (e NotePinnedEvent) AggregateID() uuid.UUID {
	return e.NoteID
}
