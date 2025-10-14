package queries

import (
	"context"
	"strings"

	"github.com/ventros/crm/internal/domain/core/shared"
	"github.com/ventros/crm/internal/domain/crm/contact"
	"go.uber.org/zap"
)

// SearchContactsQuery query to search contacts by text
type SearchContactsQuery struct {
	TenantID   shared.TenantID
	SearchText string
	Limit      int
}

// SearchContactsResponse response for search contacts query
type SearchContactsResponse struct {
	Contacts []ContactSearchResultDTO `json:"contacts"`
	Count    int                      `json:"count"`
}

// ContactSearchResultDTO search result for contact
type ContactSearchResultDTO struct {
	ID             string  `json:"id"`
	Name           string  `json:"name"`
	Phone          string  `json:"phone"`
	Email          string  `json:"email,omitempty"`
	ProfilePicture string  `json:"profile_picture,omitempty"`
	MatchScore     float64 `json:"match_score"`
	MatchField     string  `json:"match_field"`
}

// SearchContactsQueryHandler handles SearchContactsQuery
type SearchContactsQueryHandler struct {
	contactRepo contact.Repository
	logger      *zap.Logger
}

// NewSearchContactsQueryHandler creates a new SearchContactsQueryHandler
func NewSearchContactsQueryHandler(contactRepo contact.Repository, logger *zap.Logger) *SearchContactsQueryHandler {
	return &SearchContactsQueryHandler{
		contactRepo: contactRepo,
		logger:      logger,
	}
}

// Handle executes the SearchContactsQuery
func (h *SearchContactsQueryHandler) Handle(ctx context.Context, query SearchContactsQuery) (*SearchContactsResponse, error) {
	// Normalize search text
	searchText := strings.ToLower(strings.TrimSpace(query.SearchText))
	if searchText == "" {
		return &SearchContactsResponse{
			Contacts: []ContactSearchResultDTO{},
			Count:    0,
		}, nil
	}

	h.logger.Info("Searching contacts",
		zap.String("tenant_id", query.TenantID.String()),
		zap.String("search_text", searchText),
		zap.Int("limit", query.Limit))

	// Search contacts using repository
	contacts, err := h.contactRepo.SearchByText(
		ctx,
		query.TenantID.String(),
		searchText,
		query.Limit,
	)
	if err != nil {
		h.logger.Error("Failed to search contacts", zap.Error(err))
		return nil, err
	}

	// Convert domain entities to DTOs with match scoring
	results := make([]ContactSearchResultDTO, len(contacts))
	for i, c := range contacts {
		dto := ContactSearchResultDTO{
			ID:         c.ID().String(),
			Name:       c.Name(),
			MatchScore: 1.0, // Repository already ordered by relevance
		}

		// Handle optional fields
		if phone := c.Phone(); phone != nil {
			phoneStr := phone.String()
			dto.Phone = phoneStr

			// Determine match field based on what was matched
			if strings.Contains(strings.ToLower(c.Name()), searchText) {
				dto.MatchField = "name"
				dto.MatchScore = 1.5
			} else if strings.Contains(phoneStr, searchText) {
				dto.MatchField = "phone"
				dto.MatchScore = 1.3
			}
		}

		if email := c.Email(); email != nil {
			emailStr := email.String()
			dto.Email = emailStr
			if dto.MatchField == "" && strings.Contains(strings.ToLower(emailStr), searchText) {
				dto.MatchField = "email"
				dto.MatchScore = 1.2
			}
		}

		if profilePicURL := c.ProfilePictureURL(); profilePicURL != nil {
			dto.ProfilePicture = *profilePicURL
		}

		// Default match field if none set
		if dto.MatchField == "" {
			dto.MatchField = "name"
		}

		results[i] = dto
	}

	return &SearchContactsResponse{
		Contacts: results,
		Count:    len(results),
	}, nil
}
