package queries

import (
	"context"

	"github.com/caloi/ventros-crm/internal/domain/core/shared"
	"github.com/caloi/ventros-crm/internal/domain/crm/contact"
	"go.uber.org/zap"
)

// ListContactsQuery query to list contacts with filters, pagination, and sorting
type ListContactsQuery struct {
	TenantID shared.TenantID
	Filters  ContactFilters
	Page     int
	Limit    int
	SortBy   string
	SortDir  string
}

// ContactFilters filters for contact list query
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

// ListContactsResponse response for list contacts query
type ListContactsResponse struct {
	Contacts   []ContactDTO
	TotalCount int64
	Page       int
	Limit      int
	TotalPages int
}

// ContactDTO data transfer object for contact
type ContactDTO struct {
	ID             string                 `json:"id"`
	Name           string                 `json:"name"`
	Phone          string                 `json:"phone"`
	Email          string                 `json:"email,omitempty"`
	ProfilePicture string                 `json:"profile_picture,omitempty"`
	PipelineStatus map[string]string      `json:"pipeline_status,omitempty"`
	CustomFields   map[string]interface{} `json:"custom_fields,omitempty"`
	Tags           []string               `json:"tags,omitempty"`
	CreatedAt      string                 `json:"created_at"`
	UpdatedAt      string                 `json:"updated_at"`
	LastMessageAt  *string                `json:"last_message_at,omitempty"`
	LastSessionAt  *string                `json:"last_session_at,omitempty"`
	MessageCount   int                    `json:"message_count"`
	SessionCount   int                    `json:"session_count"`
}

// ListContactsQueryHandler handles ListContactsQuery
type ListContactsQueryHandler struct {
	contactRepo contact.Repository
	logger      *zap.Logger
}

// NewListContactsQueryHandler creates a new ListContactsQueryHandler
func NewListContactsQueryHandler(contactRepo contact.Repository, logger *zap.Logger) *ListContactsQueryHandler {
	return &ListContactsQueryHandler{
		contactRepo: contactRepo,
		logger:      logger,
	}
}

// Handle executes the ListContactsQuery
func (h *ListContactsQueryHandler) Handle(ctx context.Context, query ListContactsQuery) (*ListContactsResponse, error) {
	h.logger.Info("Listing contacts",
		zap.String("tenant_id", query.TenantID.String()),
		zap.Int("page", query.Page),
		zap.Int("limit", query.Limit))

	// Convert query filters to domain filters
	domainFilters := contact.ContactFilters{
		Name:           query.Filters.Name,
		Phone:          query.Filters.Phone,
		Email:          query.Filters.Email,
		PipelineID:     query.Filters.PipelineID,
		PipelineStatus: query.Filters.PipelineStatus,
		Tags:           query.Filters.Tags,
		CreatedAfter:   query.Filters.CreatedAfter,
		CreatedBefore:  query.Filters.CreatedBefore,
	}

	// Fetch contacts from repository
	contacts, totalCount, err := h.contactRepo.FindByTenantWithFilters(
		ctx,
		query.TenantID.String(),
		domainFilters,
		query.Page,
		query.Limit,
		query.SortBy,
		query.SortDir,
	)
	if err != nil {
		h.logger.Error("Failed to list contacts", zap.Error(err))
		return nil, err
	}

	// Convert domain entities to DTOs
	contactDTOs := make([]ContactDTO, len(contacts))
	for i, c := range contacts {
		dto := ContactDTO{
			ID:        c.ID().String(),
			Name:      c.Name(),
			Tags:      c.Tags(),
			CreatedAt: c.CreatedAt().Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt: c.UpdatedAt().Format("2006-01-02T15:04:05Z07:00"),
		}

		// Handle optional fields
		if phone := c.Phone(); phone != nil {
			dto.Phone = phone.String()
		}
		if email := c.Email(); email != nil {
			dto.Email = email.String()
		}
		if profilePicURL := c.ProfilePictureURL(); profilePicURL != nil {
			dto.ProfilePicture = *profilePicURL
		}

		// TODO: Add aggregated data (LastMessageAt, MessageCount, etc) from joins
		// This would require additional repository methods or database views

		contactDTOs[i] = dto
	}

	// Calculate pagination
	totalPages := int(totalCount) / query.Limit
	if int(totalCount)%query.Limit > 0 {
		totalPages++
	}

	return &ListContactsResponse{
		Contacts:   contactDTOs,
		TotalCount: totalCount,
		Page:       query.Page,
		Limit:      query.Limit,
		TotalPages: totalPages,
	}, nil
}
