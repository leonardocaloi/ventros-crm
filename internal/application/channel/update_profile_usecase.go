package channel

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/ventros/crm/infrastructure/channels/waha"
	"github.com/ventros/crm/internal/domain/crm/channel"
	"go.uber.org/zap"
)

// UpdateProfileCommand contains the data to update channel profile
type UpdateProfileCommand struct {
	ChannelID     uuid.UUID
	Name          *string                     // Optional: Update profile name
	Status        *string                     // Optional: Update profile status (About)
	Picture       *channel.ProfilePictureFile // Optional: Update profile picture
	DeletePicture bool                        // If true, delete profile picture
}

// UpdateProfileResult returns the updated profile
type UpdateProfileResult struct {
	Profile *channel.Profile `json:"profile"`
}

// UpdateProfileUseCase implements the use case to update channel's own profile
type UpdateProfileUseCase struct {
	channelRepo channel.Repository
	logger      *zap.Logger
}

// NewUpdateProfileUseCase creates a new instance
func NewUpdateProfileUseCase(
	channelRepo channel.Repository,
	logger *zap.Logger,
) *UpdateProfileUseCase {
	return &UpdateProfileUseCase{
		channelRepo: channelRepo,
		logger:      logger,
	}
}

// Execute executes the use case
func (uc *UpdateProfileUseCase) Execute(ctx context.Context, cmd UpdateProfileCommand) (*UpdateProfileResult, error) {
	// Validation
	if cmd.ChannelID == uuid.Nil {
		return nil, fmt.Errorf("channelID is required")
	}

	// At least one field must be provided
	if cmd.Name == nil && cmd.Status == nil && cmd.Picture == nil && !cmd.DeletePicture {
		return nil, fmt.Errorf("at least one field must be provided to update")
	}

	// Get channel from repository
	ch, err := uc.channelRepo.GetByID(cmd.ChannelID)
	if err != nil {
		return nil, fmt.Errorf("failed to get channel: %w", err)
	}

	// Check if channel supports profile operations (must be WAHA-based)
	if !ch.IsWAHABased() {
		return nil, fmt.Errorf("channel type %s does not support profile operations", ch.Type)
	}

	// Create profile manager based on channel config
	manager, err := uc.createProfileManager(ch)
	if err != nil {
		return nil, fmt.Errorf("failed to create profile manager: %w", err)
	}

	// Update profile name if provided
	if cmd.Name != nil {
		if err := manager.SetProfileName(ctx, *cmd.Name); err != nil {
			uc.logger.Error("Failed to update profile name",
				zap.String("channel_id", cmd.ChannelID.String()),
				zap.Error(err))
			return nil, fmt.Errorf("failed to update profile name: %w", err)
		}
		uc.logger.Info("Profile name updated",
			zap.String("channel_id", cmd.ChannelID.String()),
			zap.String("name", *cmd.Name))
	}

	// Update profile status if provided
	if cmd.Status != nil {
		if err := manager.SetProfileStatus(ctx, *cmd.Status); err != nil {
			uc.logger.Error("Failed to update profile status",
				zap.String("channel_id", cmd.ChannelID.String()),
				zap.Error(err))
			return nil, fmt.Errorf("failed to update profile status: %w", err)
		}
		uc.logger.Info("Profile status updated",
			zap.String("channel_id", cmd.ChannelID.String()),
			zap.String("status", *cmd.Status))
	}

	// Update profile picture if provided
	if cmd.Picture != nil {
		if err := manager.SetProfilePicture(ctx, *cmd.Picture); err != nil {
			uc.logger.Error("Failed to update profile picture",
				zap.String("channel_id", cmd.ChannelID.String()),
				zap.Error(err))
			return nil, fmt.Errorf("failed to update profile picture: %w", err)
		}
		uc.logger.Info("Profile picture updated",
			zap.String("channel_id", cmd.ChannelID.String()))
	}

	// Delete profile picture if requested
	if cmd.DeletePicture {
		if err := manager.DeleteProfilePicture(ctx); err != nil {
			uc.logger.Error("Failed to delete profile picture",
				zap.String("channel_id", cmd.ChannelID.String()),
				zap.Error(err))
			return nil, fmt.Errorf("failed to delete profile picture: %w", err)
		}
		uc.logger.Info("Profile picture deleted",
			zap.String("channel_id", cmd.ChannelID.String()))
	}

	// Get updated profile
	profile, err := manager.GetProfile(ctx)
	if err != nil {
		uc.logger.Error("Failed to get updated profile",
			zap.String("channel_id", cmd.ChannelID.String()),
			zap.Error(err))
		return nil, fmt.Errorf("failed to get updated profile: %w", err)
	}

	uc.logger.Info("Profile updated successfully",
		zap.String("channel_id", cmd.ChannelID.String()))

	return &UpdateProfileResult{
		Profile: profile,
	}, nil
}

// createProfileManager creates a ProfileManager based on channel configuration
func (uc *UpdateProfileUseCase) createProfileManager(ch *channel.Channel) (channel.ProfileManager, error) {
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

	// Create profile manager adapter
	sessionName := wahaConfig.SessionID
	if sessionName == "" {
		sessionName = ch.ExternalID
	}

	manager := waha.NewProfileManagerAdapter(wahaClient, sessionName, uc.logger)

	return manager, nil
}
