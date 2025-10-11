package queries

import (
	"context"

	"github.com/caloi/ventros-crm/internal/domain/core/shared"
	"github.com/caloi/ventros-crm/internal/domain/crm/contact"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// GetContactStatsQuery query to get contact statistics
type GetContactStatsQuery struct {
	ContactID uuid.UUID
	TenantID  shared.TenantID
}

// GetContactStatsResponse response for contact statistics
type GetContactStatsResponse struct {
	ContactID              string                 `json:"contact_id"`
	TotalMessages          int64                  `json:"total_messages"`
	TotalSessions          int64                  `json:"total_sessions"`
	ActiveSessions         int64                  `json:"active_sessions"`
	AverageSessionDuration string                 `json:"average_session_duration"`
	FirstContactAt         string                 `json:"first_contact_at"`
	LastContactAt          string                 `json:"last_contact_at"`
	PipelineStatuses       map[string]string      `json:"pipeline_statuses"`
	ConversionEvents       int64                  `json:"conversion_events"`
	TrackingCount          int64                  `json:"tracking_count"`
	CustomMetrics          map[string]interface{} `json:"custom_metrics,omitempty"`
}

// GetContactStatsQueryHandler handles GetContactStatsQuery
type GetContactStatsQueryHandler struct {
	contactRepo contact.Repository
	logger      *zap.Logger
}

// NewGetContactStatsQueryHandler creates a new GetContactStatsQueryHandler
func NewGetContactStatsQueryHandler(contactRepo contact.Repository, logger *zap.Logger) *GetContactStatsQueryHandler {
	return &GetContactStatsQueryHandler{
		contactRepo: contactRepo,
		logger:      logger,
	}
}

// Handle executes the GetContactStatsQuery
func (h *GetContactStatsQueryHandler) Handle(ctx context.Context, query GetContactStatsQuery) (*GetContactStatsResponse, error) {
	// Get contact to verify existence and tenant
	c, err := h.contactRepo.FindByID(ctx, query.ContactID)
	if err != nil {
		h.logger.Error("Failed to get contact", zap.Error(err))
		return nil, err
	}

	// Verify tenant
	if c.TenantID() != query.TenantID.String() {
		h.logger.Warn("Contact not found for tenant",
			zap.String("contact_id", query.ContactID.String()),
			zap.String("tenant_id", query.TenantID.String()))
		return nil, contact.ErrContactNotFound
	}

	// TODO: Implement aggregated queries in repository
	// For now, return basic stats from contact entity
	return &GetContactStatsResponse{
		ContactID:      c.ID().String(),
		TotalMessages:  0, // TODO: Query from message repository
		TotalSessions:  0, // TODO: Query from session repository
		ActiveSessions: 0, // TODO: Query from session repository
		FirstContactAt: c.CreatedAt().Format("2006-01-02T15:04:05Z07:00"),
		LastContactAt:  c.UpdatedAt().Format("2006-01-02T15:04:05Z07:00"),
		// PipelineStatuses: c.PipelineStatuses, // TODO: Map pipeline statuses
		ConversionEvents: 0, // TODO: Query from tracking repository
		TrackingCount:    0, // TODO: Query from tracking repository
	}, nil
}
