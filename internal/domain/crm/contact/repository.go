package contact

import (
	"context"

	"github.com/google/uuid"
)

// ContactFilters represents filters for contact queries
type ContactFilters struct {
	Name           string
	Phone          string
	Email          string
	PipelineID     string
	PipelineStatus string
	Tags           []string
	CreatedAfter   string
	CreatedBefore  string
}

type Repository interface {
	Save(ctx context.Context, contact *Contact) error
	FindByID(ctx context.Context, id uuid.UUID) (*Contact, error)
	FindByPhone(ctx context.Context, projectID uuid.UUID, phone string) (*Contact, error)
	FindByEmail(ctx context.Context, projectID uuid.UUID, email string) (*Contact, error)
	FindByExternalID(ctx context.Context, projectID uuid.UUID, externalID string) (*Contact, error)
	FindByProject(ctx context.Context, projectID uuid.UUID, limit, offset int) ([]*Contact, error)
	CountByProject(ctx context.Context, projectID uuid.UUID) (int, error)

	// Query methods for CQRS read models
	FindByTenantWithFilters(ctx context.Context, tenantID string, filters ContactFilters, page, limit int, sortBy, sortDir string) ([]*Contact, int64, error)
	SearchByText(ctx context.Context, tenantID string, searchText string, limit int) ([]*Contact, error)

	// Custom Fields management
	SaveCustomFields(ctx context.Context, contactID uuid.UUID, fields map[string]string) error
	FindByCustomField(ctx context.Context, tenantID, key, value string) (*Contact, error)
	GetCustomFields(ctx context.Context, contactID uuid.UUID) (map[string]string, error)
}
