package channel

import (
	"time"

	"github.com/google/uuid"
	"github.com/ventros/crm/internal/domain/core/shared"
)

type DomainEvent = shared.DomainEvent

type ChannelCreatedEvent struct {
	shared.BaseEvent
	ChannelID  uuid.UUID
	ProjectID  uuid.UUID
	TenantID   string
	Name       string
	Type       ChannelType
	ExternalID string
	CreatedAt  time.Time
}

func NewChannelCreatedEvent(channelID, projectID uuid.UUID, tenantID, name string, channelType ChannelType, externalID string) ChannelCreatedEvent {
	return ChannelCreatedEvent{
		BaseEvent:  shared.NewBaseEvent("channel.created", time.Now()),
		ChannelID:  channelID,
		ProjectID:  projectID,
		TenantID:   tenantID,
		Name:       name,
		Type:       channelType,
		ExternalID: externalID,
		CreatedAt:  time.Now(),
	}
}

// ChannelActivationRequestedEvent é publicado quando um canal solicita ativação
// Este evento inicia o processo assíncrono de ativação
type ChannelActivationRequestedEvent struct {
	shared.BaseEvent
	ChannelID   uuid.UUID
	ChannelType ChannelType
	RequestedAt time.Time
}

func NewChannelActivationRequestedEvent(channelID uuid.UUID, channelType ChannelType) ChannelActivationRequestedEvent {
	return ChannelActivationRequestedEvent{
		BaseEvent:   shared.NewBaseEvent("channel.activation.requested", time.Now()),
		ChannelID:   channelID,
		ChannelType: channelType,
		RequestedAt: time.Now(),
	}
}

// ChannelActivatedEvent é publicado quando o canal foi ativado com sucesso
type ChannelActivatedEvent struct {
	shared.BaseEvent
	ChannelID   uuid.UUID
	ActivatedAt time.Time
}

func NewChannelActivatedEvent(channelID uuid.UUID) ChannelActivatedEvent {
	return ChannelActivatedEvent{
		BaseEvent:   shared.NewBaseEvent("channel.activated", time.Now()),
		ChannelID:   channelID,
		ActivatedAt: time.Now(),
	}
}

// ChannelActivationFailedEvent é publicado quando a ativação falha
// Trigger para compensação (voltar status para inactive)
type ChannelActivationFailedEvent struct {
	shared.BaseEvent
	ChannelID uuid.UUID
	Reason    string
	FailedAt  time.Time
}

func NewChannelActivationFailedEvent(channelID uuid.UUID, reason string) ChannelActivationFailedEvent {
	return ChannelActivationFailedEvent{
		BaseEvent: shared.NewBaseEvent("channel.activation.failed", time.Now()),
		ChannelID: channelID,
		Reason:    reason,
		FailedAt:  time.Now(),
	}
}

type ChannelDeactivatedEvent struct {
	shared.BaseEvent
	ChannelID     uuid.UUID
	DeactivatedAt time.Time
}

func NewChannelDeactivatedEvent(channelID uuid.UUID) ChannelDeactivatedEvent {
	return ChannelDeactivatedEvent{
		BaseEvent:     shared.NewBaseEvent("channel.deactivated", time.Now()),
		ChannelID:     channelID,
		DeactivatedAt: time.Now(),
	}
}

type ChannelDeletedEvent struct {
	shared.BaseEvent
	ChannelID uuid.UUID
	DeletedAt time.Time
}

func NewChannelDeletedEvent(channelID uuid.UUID) ChannelDeletedEvent {
	return ChannelDeletedEvent{
		BaseEvent: shared.NewBaseEvent("channel.deleted", time.Now()),
		ChannelID: channelID,
		DeletedAt: time.Now(),
	}
}

type ChannelPipelineAssociatedEvent struct {
	shared.BaseEvent
	ChannelID    uuid.UUID
	PipelineID   uuid.UUID
	AssociatedAt time.Time
}

func NewChannelPipelineAssociatedEvent(channelID, pipelineID uuid.UUID) ChannelPipelineAssociatedEvent {
	return ChannelPipelineAssociatedEvent{
		BaseEvent:    shared.NewBaseEvent("channel.pipeline.associated", time.Now()),
		ChannelID:    channelID,
		PipelineID:   pipelineID,
		AssociatedAt: time.Now(),
	}
}

type ChannelPipelineDisassociatedEvent struct {
	shared.BaseEvent
	ChannelID       uuid.UUID
	PipelineID      uuid.UUID
	DisassociatedAt time.Time
}

func NewChannelPipelineDisassociatedEvent(channelID, pipelineID uuid.UUID) ChannelPipelineDisassociatedEvent {
	return ChannelPipelineDisassociatedEvent{
		BaseEvent:       shared.NewBaseEvent("channel.pipeline.disassociated", time.Now()),
		ChannelID:       channelID,
		PipelineID:      pipelineID,
		DisassociatedAt: time.Now(),
	}
}

// Label Events

type ChannelLabelUpsertedEvent struct {
	shared.BaseEvent
	ChannelID uuid.UUID
	LabelID   string
	LabelName string
	Timestamp time.Time
}

func NewChannelLabelUpsertedEvent(channelID uuid.UUID, labelID, labelName string) ChannelLabelUpsertedEvent {
	return ChannelLabelUpsertedEvent{
		BaseEvent: shared.NewBaseEvent("channel.label.upserted", time.Now()),
		ChannelID: channelID,
		LabelID:   labelID,
		LabelName: labelName,
		Timestamp: time.Now(),
	}
}

type ChannelLabelDeletedEvent struct {
	shared.BaseEvent
	ChannelID uuid.UUID
	LabelID   string
	Timestamp time.Time
}

func NewChannelLabelDeletedEvent(channelID uuid.UUID, labelID string) ChannelLabelDeletedEvent {
	return ChannelLabelDeletedEvent{
		BaseEvent: shared.NewBaseEvent("channel.label.deleted", time.Now()),
		ChannelID: channelID,
		LabelID:   labelID,
		Timestamp: time.Now(),
	}
}

// History Import Events

type ChannelHistoryImportEnabled struct {
	shared.BaseEvent
	ChannelID uuid.UUID
	AgentID   uuid.UUID
	Timestamp time.Time
}

func NewChannelHistoryImportEnabled(channelID, agentID uuid.UUID) ChannelHistoryImportEnabled {
	return ChannelHistoryImportEnabled{
		BaseEvent: shared.NewBaseEvent("channel.history_import.enabled", time.Now()),
		ChannelID: channelID,
		AgentID:   agentID,
		Timestamp: time.Now(),
	}
}

// ChannelHistoryImportRequestedEvent é publicado quando um canal solicita importação de histórico
// Este evento inicia o processo assíncrono de importação
type ChannelHistoryImportRequestedEvent struct {
	shared.BaseEvent
	ChannelID     uuid.UUID   `json:"channel_id"`
	ChannelType   ChannelType `json:"channel_type"`
	CorrelationID string      `json:"correlation_id"` // Para tracking (Saga Pattern)
	Strategy      string      `json:"strategy"`       // "time_range", "full", "recent"
	TimeRangeDays int         `json:"time_range_days"`
	Limit         int         `json:"limit"`
	RequestedAt   time.Time   `json:"requested_at"`
}

func NewChannelHistoryImportRequestedEvent(channelID uuid.UUID, channelType ChannelType, correlationID, strategy string, timeRangeDays, limit int) ChannelHistoryImportRequestedEvent {
	return ChannelHistoryImportRequestedEvent{
		BaseEvent:     shared.NewBaseEvent("channel.history_import.requested", time.Now()),
		ChannelID:     channelID,
		ChannelType:   channelType,
		CorrelationID: correlationID,
		Strategy:      strategy,
		TimeRangeDays: timeRangeDays,
		Limit:         limit,
		RequestedAt:   time.Now(),
	}
}

type ChannelHistoryImportStarted struct {
	shared.BaseEvent
	ChannelID uuid.UUID
	Timestamp time.Time
}

func NewChannelHistoryImportStarted(channelID uuid.UUID) ChannelHistoryImportStarted {
	return ChannelHistoryImportStarted{
		BaseEvent: shared.NewBaseEvent("channel.history_import.started", time.Now()),
		ChannelID: channelID,
		Timestamp: time.Now(),
	}
}

type ChannelHistoryImportCompleted struct {
	shared.BaseEvent
	ChannelID uuid.UUID
	Stats     HistoryImportStats
	Timestamp time.Time
}

func NewChannelHistoryImportCompleted(channelID uuid.UUID, stats HistoryImportStats) ChannelHistoryImportCompleted {
	return ChannelHistoryImportCompleted{
		BaseEvent: shared.NewBaseEvent("channel.history_import.completed", time.Now()),
		ChannelID: channelID,
		Stats:     stats,
		Timestamp: time.Now(),
	}
}

type ChannelHistoryImportFailed struct {
	shared.BaseEvent
	ChannelID uuid.UUID
	Reason    string
	Timestamp time.Time
}

func NewChannelHistoryImportFailed(channelID uuid.UUID, reason string) ChannelHistoryImportFailed {
	return ChannelHistoryImportFailed{
		BaseEvent: shared.NewBaseEvent("channel.history_import.failed", time.Now()),
		ChannelID: channelID,
		Reason:    reason,
		Timestamp: time.Now(),
	}
}
