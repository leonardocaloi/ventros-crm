package broadcast

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// BroadcastExecution rastreia envio para cada contato
type BroadcastExecution struct {
	id          uuid.UUID
	broadcastID uuid.UUID
	contactID   uuid.UUID
	status      ExecutionStatus
	messageID   *uuid.UUID // ID da mensagem enviada
	error       *string
	sentAt      *time.Time
	createdAt   time.Time
	updatedAt   time.Time
}

type ExecutionStatus string

const (
	ExecutionStatusPending ExecutionStatus = "pending"
	ExecutionStatusSending ExecutionStatus = "sending"
	ExecutionStatusSent    ExecutionStatus = "sent"
	ExecutionStatusFailed  ExecutionStatus = "failed"
	ExecutionStatusSkipped ExecutionStatus = "skipped"
)

// NewBroadcastExecution creates a new broadcast execution
func NewBroadcastExecution(broadcastID, contactID uuid.UUID) (*BroadcastExecution, error) {
	if broadcastID == uuid.Nil {
		return nil, errors.New("broadcastID cannot be empty")
	}
	if contactID == uuid.Nil {
		return nil, errors.New("contactID cannot be empty")
	}

	now := time.Now()
	return &BroadcastExecution{
		id:          uuid.New(),
		broadcastID: broadcastID,
		contactID:   contactID,
		status:      ExecutionStatusPending,
		createdAt:   now,
		updatedAt:   now,
	}, nil
}

// ReconstructBroadcastExecution reconstructs execution from persistence
func ReconstructBroadcastExecution(
	id, broadcastID, contactID uuid.UUID,
	status ExecutionStatus,
	messageID *uuid.UUID,
	errorMsg *string,
	sentAt *time.Time,
	createdAt, updatedAt time.Time,
) *BroadcastExecution {
	return &BroadcastExecution{
		id:          id,
		broadcastID: broadcastID,
		contactID:   contactID,
		status:      status,
		messageID:   messageID,
		error:       errorMsg,
		sentAt:      sentAt,
		createdAt:   createdAt,
		updatedAt:   updatedAt,
	}
}

// MarkSending marks the execution as sending
func (e *BroadcastExecution) MarkSending() {
	e.status = ExecutionStatusSending
	e.updatedAt = time.Now()
}

// MarkSent marks the execution as sent successfully
func (e *BroadcastExecution) MarkSent(messageID uuid.UUID) {
	now := time.Now()
	e.status = ExecutionStatusSent
	e.messageID = &messageID
	e.sentAt = &now
	e.updatedAt = now
}

// MarkFailed marks the execution as failed
func (e *BroadcastExecution) MarkFailed(errorMsg string) {
	e.status = ExecutionStatusFailed
	e.error = &errorMsg
	e.updatedAt = time.Now()
}

// MarkSkipped marks the execution as skipped
func (e *BroadcastExecution) MarkSkipped(reason string) {
	e.status = ExecutionStatusSkipped
	e.error = &reason
	e.updatedAt = time.Now()
}

// Getters
func (e *BroadcastExecution) ID() uuid.UUID           { return e.id }
func (e *BroadcastExecution) BroadcastID() uuid.UUID  { return e.broadcastID }
func (e *BroadcastExecution) ContactID() uuid.UUID    { return e.contactID }
func (e *BroadcastExecution) Status() ExecutionStatus { return e.status }
func (e *BroadcastExecution) MessageID() *uuid.UUID   { return e.messageID }
func (e *BroadcastExecution) Error() *string          { return e.error }
func (e *BroadcastExecution) SentAt() *time.Time      { return e.sentAt }
func (e *BroadcastExecution) CreatedAt() time.Time    { return e.createdAt }
func (e *BroadcastExecution) UpdatedAt() time.Time    { return e.updatedAt }

// ExecutionRepository interface
type ExecutionRepository interface {
	Save(execution *BroadcastExecution) error
	SaveBatch(executions []*BroadcastExecution) error
	FindByID(id uuid.UUID) (*BroadcastExecution, error)
	FindByBroadcastID(broadcastID uuid.UUID) ([]*BroadcastExecution, error)
	FindPendingByBroadcastID(broadcastID uuid.UUID) ([]*BroadcastExecution, error)
	Delete(id uuid.UUID) error
}
