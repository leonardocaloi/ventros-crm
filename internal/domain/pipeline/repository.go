package pipeline

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Repository define as operações de persistência para Pipeline e Status
type Repository interface {
	// Pipeline operations
	SavePipeline(ctx context.Context, pipeline *Pipeline) error
	FindPipelineByID(ctx context.Context, id uuid.UUID) (*Pipeline, error)
	FindPipelinesByProject(ctx context.Context, projectID uuid.UUID) ([]*Pipeline, error)
	FindPipelinesByTenant(ctx context.Context, tenantID string) ([]*Pipeline, error)
	FindActivePipelinesByProject(ctx context.Context, projectID uuid.UUID) ([]*Pipeline, error)
	DeletePipeline(ctx context.Context, id uuid.UUID) error

	// Status operations
	SaveStatus(ctx context.Context, status *Status) error
	FindStatusByID(ctx context.Context, id uuid.UUID) (*Status, error)
	FindStatusesByPipeline(ctx context.Context, pipelineID uuid.UUID) ([]*Status, error)
	FindActiveStatusesByPipeline(ctx context.Context, pipelineID uuid.UUID) ([]*Status, error)
	DeleteStatus(ctx context.Context, id uuid.UUID) error

	// Pipeline-Status relationships
	AddStatusToPipeline(ctx context.Context, pipelineID, statusID uuid.UUID) error
	RemoveStatusFromPipeline(ctx context.Context, pipelineID, statusID uuid.UUID) error
	GetPipelineWithStatuses(ctx context.Context, pipelineID uuid.UUID) (*Pipeline, []*Status, error)

	// Contact-Status relationships (for tracking contact status in pipelines)
	SetContactStatus(ctx context.Context, contactID, pipelineID, statusID uuid.UUID) error
	GetContactStatus(ctx context.Context, contactID, pipelineID uuid.UUID) (*Status, error)
	GetContactsByStatus(ctx context.Context, pipelineID, statusID uuid.UUID) ([]uuid.UUID, error)
	GetContactStatusHistory(ctx context.Context, contactID, pipelineID uuid.UUID) ([]*ContactStatusHistory, error)
}

// ContactStatusHistory representa o histórico de mudanças de status de um contato
type ContactStatusHistory struct {
	ID          uuid.UUID
	ContactID   uuid.UUID
	PipelineID  uuid.UUID
	StatusID    uuid.UUID
	StatusName  string
	ChangedAt   time.Time
	ChangedBy   *uuid.UUID
	Reason      string
	Duration    *time.Duration // Tempo que ficou neste status
}

// ContactPipelineStatus representa o status atual de um contato em um pipeline
type ContactPipelineStatus struct {
	ContactID   uuid.UUID
	PipelineID  uuid.UUID
	StatusID    uuid.UUID
	StatusName  string
	EnteredAt   time.Time
	UpdatedAt   time.Time
	UpdatedBy   *uuid.UUID
}
