package channel

import (
	"context"
	"fmt"

	"github.com/caloi/ventros-crm/infrastructure/channels/waha"
	"github.com/caloi/ventros-crm/internal/domain/channel"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// GetAllContactsCommand contains the data to fetch all contacts
type GetAllContactsCommand struct {
	ChannelID uuid.UUID
	SortBy    string // "id" or "name"
	SortOrder string // "asc" or "desc"
	Limit     int
	Offset    int
}

// GetAllContactsResult returns the list of contacts
type GetAllContactsResult struct {
	Contacts []channel.Contact `json:"contacts"`
	Total    int               `json:"total"`
}

// GetAllContactsUseCase implements the use case to get all contacts from a channel
type GetAllContactsUseCase struct {
	channelRepo channel.Repository
	logger      *zap.Logger
}

// NewGetAllContactsUseCase creates a new instance
func NewGetAllContactsUseCase(
	channelRepo channel.Repository,
	logger *zap.Logger,
) *GetAllContactsUseCase {
	return &GetAllContactsUseCase{
		channelRepo: channelRepo,
		logger:      logger,
	}
}

// Execute executes the use case
func (uc *GetAllContactsUseCase) Execute(ctx context.Context, cmd GetAllContactsCommand) (*GetAllContactsResult, error) {
	// Validation
	if cmd.ChannelID == uuid.Nil {
		return nil, fmt.Errorf("channelID is required")
	}

	// Get channel from repository
	ch, err := uc.channelRepo.GetByID(cmd.ChannelID)
	if err != nil {
		return nil, fmt.Errorf("failed to get channel: %w", err)
	}

	// Check if channel supports contact operations (must be WAHA-based)
	if !ch.IsWAHABased() {
		return nil, fmt.Errorf("channel type %s does not support contact operations", ch.Type)
	}

	// Create contact provider based on channel config
	provider, err := uc.createContactProvider(ch)
	if err != nil {
		return nil, fmt.Errorf("failed to create contact provider: %w", err)
	}

	// Set defaults
	sortBy := cmd.SortBy
	if sortBy == "" {
		sortBy = "name"
	}
	sortOrder := cmd.SortOrder
	if sortOrder == "" {
		sortOrder = "asc"
	}
	limit := cmd.Limit
	if limit <= 0 {
		limit = 100 // default limit
	}

	// Get all contacts
	contacts, err := provider.GetAllContacts(ctx, sortBy, sortOrder, limit, cmd.Offset)
	if err != nil {
		uc.logger.Error("Failed to get all contacts",
			zap.String("channel_id", cmd.ChannelID.String()),
			zap.Error(err))
		return nil, fmt.Errorf("failed to get all contacts: %w", err)
	}

	uc.logger.Info("Successfully retrieved contacts",
		zap.String("channel_id", cmd.ChannelID.String()),
		zap.Int("count", len(contacts)))

	return &GetAllContactsResult{
		Contacts: contacts,
		Total:    len(contacts),
	}, nil
}

// createContactProvider creates a ContactProvider based on channel configuration
func (uc *GetAllContactsUseCase) createContactProvider(ch *channel.Channel) (channel.ContactProvider, error) {
	// Get WAHA config (works for both TypeWAHA and TypeWhatsAppBusiness)
	wahaConfig, err := ch.GetWAHAConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get WAHA config: %w", err)
	}

	// Get auth token (try APIKey first, then Token)
	authToken := wahaConfig.Auth.APIKey
	if authToken == "" {
		authToken = wahaConfig.Auth.Token
	}

	// Create WAHA client
	wahaClient := waha.NewClient(wahaConfig.BaseURL, authToken)

	// Create contact provider adapter
	sessionName := wahaConfig.SessionID
	if sessionName == "" {
		sessionName = ch.ExternalID
	}

	provider := waha.NewContactProviderAdapter(wahaClient, sessionName, uc.logger)

	return provider, nil
}
