package channel_type

import (
	"context"
	"fmt"

	"github.com/caloi/ventros-crm/internal/domain/crm/channel_type"
)

// GetChannelTypeUseCase handles channel type retrieval
type GetChannelTypeUseCase struct {
	channelTypeRepo channel_type.Repository
}

// NewGetChannelTypeUseCase creates a new instance
func NewGetChannelTypeUseCase(channelTypeRepo channel_type.Repository) *GetChannelTypeUseCase {
	return &GetChannelTypeUseCase{
		channelTypeRepo: channelTypeRepo,
	}
}

// GetChannelTypeRequest represents the request to get a channel type
type GetChannelTypeRequest struct {
	ID int `json:"id" validate:"required,min=1"`
}

// GetChannelTypeByNameRequest represents the request to get a channel type by name
type GetChannelTypeByNameRequest struct {
	Name string `json:"name" validate:"required,min=2"`
}

// GetChannelTypeResponse represents the response with channel type details
type GetChannelTypeResponse struct {
	ID            int                    `json:"id"`
	Name          string                 `json:"name"`
	Description   string                 `json:"description"`
	Provider      string                 `json:"provider"`
	Configuration map[string]interface{} `json:"configuration"`
	IsActive      bool                   `json:"is_active"`
	CreatedAt     string                 `json:"created_at"`
	UpdatedAt     string                 `json:"updated_at"`
}

// Execute retrieves a channel type by ID
func (uc *GetChannelTypeUseCase) Execute(ctx context.Context, req GetChannelTypeRequest) (*GetChannelTypeResponse, error) {
	// Find channel type
	foundChannelType, err := uc.channelTypeRepo.FindByID(ctx, req.ID)
	if err != nil {
		if err == channel_type.ErrChannelTypeNotFound {
			return nil, fmt.Errorf("channel type not found")
		}
		return nil, fmt.Errorf("failed to find channel type: %w", err)
	}

	// Return response
	return &GetChannelTypeResponse{
		ID:            foundChannelType.ID(),
		Name:          foundChannelType.Name(),
		Description:   foundChannelType.Description(),
		Provider:      foundChannelType.Provider(),
		Configuration: foundChannelType.Configuration(),
		IsActive:      foundChannelType.IsActive(),
		CreatedAt:     foundChannelType.CreatedAt().Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:     foundChannelType.UpdatedAt().Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}

// GetChannelTypeByNameUseCase handles channel type retrieval by name
type GetChannelTypeByNameUseCase struct {
	channelTypeRepo channel_type.Repository
}

// NewGetChannelTypeByNameUseCase creates a new instance
func NewGetChannelTypeByNameUseCase(channelTypeRepo channel_type.Repository) *GetChannelTypeByNameUseCase {
	return &GetChannelTypeByNameUseCase{
		channelTypeRepo: channelTypeRepo,
	}
}

// Execute retrieves a channel type by name
func (uc *GetChannelTypeByNameUseCase) Execute(ctx context.Context, req GetChannelTypeByNameRequest) (*GetChannelTypeResponse, error) {
	// Find channel type
	foundChannelType, err := uc.channelTypeRepo.FindByName(ctx, req.Name)
	if err != nil {
		if err == channel_type.ErrChannelTypeNotFound {
			return nil, fmt.Errorf("channel type '%s' not found", req.Name)
		}
		return nil, fmt.Errorf("failed to find channel type: %w", err)
	}

	// Return response
	return &GetChannelTypeResponse{
		ID:            foundChannelType.ID(),
		Name:          foundChannelType.Name(),
		Description:   foundChannelType.Description(),
		Provider:      foundChannelType.Provider(),
		Configuration: foundChannelType.Configuration(),
		IsActive:      foundChannelType.IsActive(),
		CreatedAt:     foundChannelType.CreatedAt().Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:     foundChannelType.UpdatedAt().Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}

// ListChannelTypesUseCase handles listing channel types
type ListChannelTypesUseCase struct {
	channelTypeRepo channel_type.Repository
}

// NewListChannelTypesUseCase creates a new instance
func NewListChannelTypesUseCase(channelTypeRepo channel_type.Repository) *ListChannelTypesUseCase {
	return &ListChannelTypesUseCase{
		channelTypeRepo: channelTypeRepo,
	}
}

// ListChannelTypesRequest represents the request to list channel types
type ListChannelTypesRequest struct {
	ActiveOnly bool `json:"active_only"`
}

// ListChannelTypesResponse represents the response with list of channel types
type ListChannelTypesResponse struct {
	ChannelTypes []GetChannelTypeResponse `json:"channel_types"`
	Total        int                      `json:"total"`
}

// Execute lists channel types
func (uc *ListChannelTypesUseCase) Execute(ctx context.Context, req ListChannelTypesRequest) (*ListChannelTypesResponse, error) {
	// Find channel types
	var channelTypes []*channel_type.ChannelType
	var err error

	if req.ActiveOnly {
		channelTypes, err = uc.channelTypeRepo.FindActive(ctx)
	} else {
		channelTypes, err = uc.channelTypeRepo.FindAll(ctx)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to find channel types: %w", err)
	}

	// Convert to response format
	channelTypeResponses := make([]GetChannelTypeResponse, len(channelTypes))
	for i, ct := range channelTypes {
		channelTypeResponses[i] = GetChannelTypeResponse{
			ID:            ct.ID(),
			Name:          ct.Name(),
			Description:   ct.Description(),
			Provider:      ct.Provider(),
			Configuration: ct.Configuration(),
			IsActive:      ct.IsActive(),
			CreatedAt:     ct.CreatedAt().Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:     ct.UpdatedAt().Format("2006-01-02T15:04:05Z07:00"),
		}
	}

	return &ListChannelTypesResponse{
		ChannelTypes: channelTypeResponses,
		Total:        len(channelTypeResponses),
	}, nil
}

// GetAvailableChannelTypesUseCase returns predefined channel types
type GetAvailableChannelTypesUseCase struct{}

// NewGetAvailableChannelTypesUseCase creates a new instance
func NewGetAvailableChannelTypesUseCase() *GetAvailableChannelTypesUseCase {
	return &GetAvailableChannelTypesUseCase{}
}

// AvailableChannelType represents a predefined channel type
type AvailableChannelType struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Description string `json:"description"`
	Provider    string `json:"provider"`
}

// GetAvailableChannelTypesResponse represents available channel types
type GetAvailableChannelTypesResponse struct {
	ChannelTypes []AvailableChannelType `json:"channel_types"`
}

// Execute returns all available channel types
func (uc *GetAvailableChannelTypesUseCase) Execute(ctx context.Context) (*GetAvailableChannelTypesResponse, error) {
	availableTypes := []AvailableChannelType{
		{
			ID:          channel_type.WAHA,
			Name:        "waha",
			DisplayName: "WhatsApp (WAHA)",
			Description: "WhatsApp HTTP API - Multi-device support",
			Provider:    "waha",
		},
		{
			ID:          channel_type.WhatsApp,
			Name:        "whatsapp",
			DisplayName: "WhatsApp Business",
			Description: "Official WhatsApp Business API",
			Provider:    "meta",
		},
		{
			ID:          channel_type.DirectIG,
			Name:        "direct_ig",
			DisplayName: "Instagram Direct",
			Description: "Instagram Direct Messages",
			Provider:    "meta",
		},
		{
			ID:          channel_type.Messenger,
			Name:        "messenger",
			DisplayName: "Facebook Messenger",
			Description: "Facebook Messenger Platform",
			Provider:    "meta",
		},
		{
			ID:          channel_type.Telegram,
			Name:        "telegram",
			DisplayName: "Telegram",
			Description: "Telegram Bot API",
			Provider:    "telegram",
		},
	}

	return &GetAvailableChannelTypesResponse{
		ChannelTypes: availableTypes,
	}, nil
}
