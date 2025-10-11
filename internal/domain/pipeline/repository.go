package pipeline

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// PipelineFilters represents filtering options for pipeline queries
type PipelineFilters struct {
	TenantID  string
	ProjectID *uuid.UUID
	Active    *bool
	Color     *string
	Limit     int
	Offset    int
	SortBy    string // position, name, created_at
	SortOrder string // asc, desc
}

type Repository interface {
	SavePipeline(ctx context.Context, pipeline *Pipeline) error
	FindPipelineByID(ctx context.Context, id uuid.UUID) (*Pipeline, error)
	FindPipelinesByProject(ctx context.Context, projectID uuid.UUID) ([]*Pipeline, error)
	FindPipelinesByTenant(ctx context.Context, tenantID string) ([]*Pipeline, error)
	FindActivePipelinesByProject(ctx context.Context, projectID uuid.UUID) ([]*Pipeline, error)
	DeletePipeline(ctx context.Context, id uuid.UUID) error

	SaveStatus(ctx context.Context, status *Status) error
	FindStatusByID(ctx context.Context, id uuid.UUID) (*Status, error)
	FindStatusesByPipeline(ctx context.Context, pipelineID uuid.UUID) ([]*Status, error)
	FindActiveStatusesByPipeline(ctx context.Context, pipelineID uuid.UUID) ([]*Status, error)
	DeleteStatus(ctx context.Context, id uuid.UUID) error

	AddStatusToPipeline(ctx context.Context, pipelineID, statusID uuid.UUID) error
	RemoveStatusFromPipeline(ctx context.Context, pipelineID, statusID uuid.UUID) error
	GetPipelineWithStatuses(ctx context.Context, pipelineID uuid.UUID) (*Pipeline, []*Status, error)

	SetContactStatus(ctx context.Context, contactID, pipelineID, statusID uuid.UUID) error
	GetContactStatus(ctx context.Context, contactID, pipelineID uuid.UUID) (*Status, error)
	GetContactsByStatus(ctx context.Context, pipelineID, statusID uuid.UUID) ([]uuid.UUID, error)
	GetContactStatusHistory(ctx context.Context, contactID, pipelineID uuid.UUID) ([]*ContactStatusHistory, error)

	// Advanced query methods
	FindByTenantWithFilters(ctx context.Context, filters PipelineFilters) ([]*Pipeline, int64, error)

	SearchByText(ctx context.Context, tenantID string, searchText string, limit int, offset int) ([]*Pipeline, int64, error)
}

type ContactStatusHistory struct {
	ID         uuid.UUID
	ContactID  uuid.UUID
	PipelineID uuid.UUID
	StatusID   uuid.UUID
	StatusName string
	ChangedAt  time.Time
	ChangedBy  *uuid.UUID
	Reason     string
	Duration   *time.Duration
}

type ContactPipelineStatus struct {
	ContactID  uuid.UUID
	PipelineID uuid.UUID
	StatusID   uuid.UUID
	StatusName string
	EnteredAt  time.Time
	UpdatedAt  time.Time
	UpdatedBy  *uuid.UUID
}
