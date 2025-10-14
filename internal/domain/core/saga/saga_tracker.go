package saga

import (
	"context"
	"fmt"
	"time"

	"github.com/ventros/crm/internal/domain/core/outbox"
)

// SagaTracker provides utilities for tracking and monitoring Saga executions via Outbox.
type SagaTracker struct {
	outboxRepo outbox.Repository
}

// NewSagaTracker cria uma nova instância do tracker.
func NewSagaTracker(outboxRepo outbox.Repository) *SagaTracker {
	return &SagaTracker{
		outboxRepo: outboxRepo,
	}
}

// SagaExecution representa o estado de uma execução de Saga.
type SagaExecution struct {
	CorrelationID  string
	SagaType       string
	Status         string // "in_progress", "completed", "failed"
	TotalSteps     int
	CompletedSteps int
	FailedSteps    int
	StartedAt      time.Time
	CompletedAt    *time.Time
	Duration       time.Duration
	Events         []*outbox.OutboxEvent
}

// TrackSaga recupera o status completo de uma Saga execution.
func (t *SagaTracker) TrackSaga(ctx context.Context, correlationID string) (*SagaExecution, error) {
	// Get all events for this saga
	events, err := t.outboxRepo.GetSagaEvents(ctx, correlationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get saga events: %w", err)
	}

	if len(events) == 0 {
		return nil, fmt.Errorf("saga not found: %s", correlationID)
	}

	// Get aggregated status
	status, total, completed, err := t.outboxRepo.GetSagaStatus(ctx, correlationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get saga status: %w", err)
	}

	// Count failed steps
	failed := 0
	for _, event := range events {
		if event.Status == outbox.StatusFailed {
			failed++
		}
	}

	// Calculate duration
	startedAt := events[0].CreatedAt
	var completedAt *time.Time
	var duration time.Duration

	if status == "completed" || status == "failed" {
		// Find the last processed or failed event
		for i := len(events) - 1; i >= 0; i-- {
			if events[i].ProcessedAt != nil {
				completedAt = events[i].ProcessedAt
				duration = completedAt.Sub(startedAt)
				break
			}
		}
	} else {
		duration = time.Since(startedAt)
	}

	// Extract saga type from first event metadata
	sagaType := ""
	if events[0].Metadata != nil {
		if st, ok := events[0].Metadata["saga_type"].(string); ok {
			sagaType = st
		}
	}

	return &SagaExecution{
		CorrelationID:  correlationID,
		SagaType:       sagaType,
		Status:         status,
		TotalSteps:     total,
		CompletedSteps: completed,
		FailedSteps:    failed,
		StartedAt:      startedAt,
		CompletedAt:    completedAt,
		Duration:       duration,
		Events:         events,
	}, nil
}

// IsCompleted verifica se uma Saga foi completada com sucesso.
func (t *SagaTracker) IsCompleted(ctx context.Context, correlationID string) (bool, error) {
	status, total, completed, err := t.outboxRepo.GetSagaStatus(ctx, correlationID)
	if err != nil {
		return false, err
	}

	return status == "completed" && completed == total, nil
}

// IsFailed verifica se uma Saga falhou.
func (t *SagaTracker) IsFailed(ctx context.Context, correlationID string) (bool, error) {
	status, _, _, err := t.outboxRepo.GetSagaStatus(ctx, correlationID)
	if err != nil {
		return false, err
	}

	return status == "failed", nil
}

// GetFailedSteps retorna os eventos que falharam em uma Saga.
func (t *SagaTracker) GetFailedSteps(ctx context.Context, correlationID string) ([]*outbox.OutboxEvent, error) {
	events, err := t.outboxRepo.GetSagaEvents(ctx, correlationID)
	if err != nil {
		return nil, err
	}

	failedSteps := make([]*outbox.OutboxEvent, 0)
	for _, event := range events {
		if event.Status == outbox.StatusFailed {
			failedSteps = append(failedSteps, event)
		}
	}

	return failedSteps, nil
}

// GetExecutionTimeline retorna uma timeline dos eventos da Saga ordenados por tempo.
func (t *SagaTracker) GetExecutionTimeline(ctx context.Context, correlationID string) ([]TimelineEntry, error) {
	events, err := t.outboxRepo.GetSagaEvents(ctx, correlationID)
	if err != nil {
		return nil, err
	}

	timeline := make([]TimelineEntry, len(events))
	for i, event := range events {
		sagaStep := ""
		stepNumber := 0

		if event.Metadata != nil {
			if step, ok := event.Metadata["saga_step"].(string); ok {
				sagaStep = step
			}
			if num, ok := event.Metadata["step_number"].(float64); ok {
				stepNumber = int(num)
			}
		}

		timeline[i] = TimelineEntry{
			StepNumber:  stepNumber,
			SagaStep:    sagaStep,
			EventType:   event.EventType,
			Status:      string(event.Status),
			Timestamp:   event.CreatedAt,
			ProcessedAt: event.ProcessedAt,
			Error:       event.LastError,
		}
	}

	return timeline, nil
}

// TimelineEntry representa uma entrada na timeline de execução da Saga.
type TimelineEntry struct {
	StepNumber  int
	SagaStep    string
	EventType   string
	Status      string
	Timestamp   time.Time
	ProcessedAt *time.Time
	Error       *string
}

// WaitForCompletion aguarda a conclusão de uma Saga com timeout.
func (t *SagaTracker) WaitForCompletion(ctx context.Context, correlationID string, timeout time.Duration, pollInterval time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for saga completion: %s", correlationID)

		case <-ticker.C:
			status, _, _, err := t.outboxRepo.GetSagaStatus(ctx, correlationID)
			if err != nil {
				return fmt.Errorf("failed to check saga status: %w", err)
			}

			if status == "completed" {
				return nil
			}

			if status == "failed" {
				return fmt.Errorf("saga failed: %s", correlationID)
			}
		}
	}
}
