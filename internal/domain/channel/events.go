package channel

import (
	"time"

	"github.com/google/uuid"
)

// DomainEvent é a interface base para eventos de domínio.
type DomainEvent interface {
	EventName() string
	OccurredAt() time.Time
}

// ChannelCreatedEvent - Canal criado no sistema.
type ChannelCreatedEvent struct {
	ChannelID  uuid.UUID
	ProjectID  uuid.UUID
	TenantID   string
	Name       string
	Type       ChannelType
	ExternalID string
	CreatedAt  time.Time
}

func (e ChannelCreatedEvent) EventName() string     { return "channel.created" }
func (e ChannelCreatedEvent) OccurredAt() time.Time { return e.CreatedAt }

// ChannelActivatedEvent - Canal ativado.
type ChannelActivatedEvent struct {
	ChannelID   uuid.UUID
	ActivatedAt time.Time
}

func (e ChannelActivatedEvent) EventName() string     { return "channel.activated" }
func (e ChannelActivatedEvent) OccurredAt() time.Time { return e.ActivatedAt }

// ChannelDeactivatedEvent - Canal desativado.
type ChannelDeactivatedEvent struct {
	ChannelID     uuid.UUID
	DeactivatedAt time.Time
}

func (e ChannelDeactivatedEvent) EventName() string     { return "channel.deactivated" }
func (e ChannelDeactivatedEvent) OccurredAt() time.Time { return e.DeactivatedAt }

// ChannelDeletedEvent - Canal deletado.
type ChannelDeletedEvent struct {
	ChannelID uuid.UUID
	DeletedAt time.Time
}

func (e ChannelDeletedEvent) EventName() string     { return "channel.deleted" }
func (e ChannelDeletedEvent) OccurredAt() time.Time { return e.DeletedAt }

// ChannelPipelineAssociatedEvent - Pipeline associado ao canal.
type ChannelPipelineAssociatedEvent struct {
	ChannelID    uuid.UUID
	PipelineID   uuid.UUID
	AssociatedAt time.Time
}

func (e ChannelPipelineAssociatedEvent) EventName() string     { return "channel.pipeline.associated" }
func (e ChannelPipelineAssociatedEvent) OccurredAt() time.Time { return e.AssociatedAt }

// ChannelPipelineDisassociatedEvent - Pipeline desassociado do canal.
type ChannelPipelineDisassociatedEvent struct {
	ChannelID       uuid.UUID
	PipelineID      uuid.UUID
	DisassociatedAt time.Time
}

func (e ChannelPipelineDisassociatedEvent) EventName() string {
	return "channel.pipeline.disassociated"
}
func (e ChannelPipelineDisassociatedEvent) OccurredAt() time.Time { return e.DisassociatedAt }
