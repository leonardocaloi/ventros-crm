package dtos

import (
	"time"

	"github.com/google/uuid"
)

// PipelineStatusDTO - DTO para status no pipeline
type PipelineStatusDTO struct {
	PipelineID   uuid.UUID  `json:"pipeline_id"`
	PipelineName string     `json:"pipeline_name"`
	StatusID     uuid.UUID  `json:"status_id"`
	StatusName   string     `json:"status_name"`
	StatusType   string     `json:"status_type"`
	Color        *string    `json:"color,omitempty"`
	EnteredAt    time.Time  `json:"entered_at"`
	Duration     *int64     `json:"duration_seconds,omitempty"`
}

// AgentSummaryDTO - DTO resumido para agente
type AgentSummaryDTO struct {
	ID     uuid.UUID `json:"id"`
	Name   string    `json:"name"`
	Email  string    `json:"email"`
	Role   string    `json:"role"`
	Active bool      `json:"active"`
}
