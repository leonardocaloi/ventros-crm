package waha

import (
	"context"
	"fmt"

	"github.com/caloi/ventros-crm/internal/domain/crm/channel"
	"go.uber.org/zap"
)

// ProfileManagerAdapter implements domain.ProfileManager for WAHA
//
// This adapter manages the channel's own profile (not other contacts).
// Used by both channel types:
// - TypeWAHA (manual mode)
// - TypeWhatsAppBusiness (auto mode)
type ProfileManagerAdapter struct {
	client      *Client
	sessionName string
	logger      *zap.Logger
}

// NewProfileManagerAdapter creates a new profile manager adapter
func NewProfileManagerAdapter(client *Client, sessionName string, logger *zap.Logger) *ProfileManagerAdapter {
	return &ProfileManagerAdapter{
		client:      client,
		sessionName: sessionName,
		logger:      logger,
	}
}

// GetProfile gets current profile information
func (a *ProfileManagerAdapter) GetProfile(ctx context.Context) (*channel.Profile, error) {
	a.logger.Info("Getting profile",
		zap.String("session_name", a.sessionName))

	path := fmt.Sprintf("/api/%s/profile/about", a.sessionName)

	resp, err := a.client.Get(ctx, path)
	if err != nil {
		a.logger.Error("Failed to get profile",
			zap.String("session_name", a.sessionName),
			zap.Error(err))
		return nil, fmt.Errorf("failed to get profile: %w", err)
	}

	var wahaProfile WAHAProfile
	if err := a.client.ParseResponse(resp, &wahaProfile); err != nil {
		return nil, fmt.Errorf("failed to parse profile response: %w", err)
	}

	profile := &channel.Profile{
		ID:      wahaProfile.ID,
		Name:    wahaProfile.Name,
		Status:  wahaProfile.Status,
		Picture: wahaProfile.Picture,
	}

	a.logger.Info("Successfully retrieved profile",
		zap.String("session_name", a.sessionName),
		zap.String("profile_id", profile.ID))

	return profile, nil
}

// SetProfileName sets profile display name
func (a *ProfileManagerAdapter) SetProfileName(ctx context.Context, name string) error {
	a.logger.Info("Setting profile name",
		zap.String("session_name", a.sessionName),
		zap.String("name", name))

	path := fmt.Sprintf("/api/%s/profile/name", a.sessionName)

	payload := map[string]string{
		"name": name,
	}

	resp, err := a.client.Post(ctx, path, payload)
	if err != nil {
		a.logger.Error("Failed to set profile name",
			zap.String("session_name", a.sessionName),
			zap.Error(err))
		return fmt.Errorf("failed to set profile name: %w", err)
	}

	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	a.logger.Info("Profile name set successfully",
		zap.String("session_name", a.sessionName),
		zap.String("name", name))

	return nil
}

// SetProfileStatus sets profile status (About)
func (a *ProfileManagerAdapter) SetProfileStatus(ctx context.Context, status string) error {
	a.logger.Info("Setting profile status",
		zap.String("session_name", a.sessionName),
		zap.String("status", status))

	path := fmt.Sprintf("/api/%s/profile/status", a.sessionName)

	payload := map[string]string{
		"status": status,
	}

	resp, err := a.client.Post(ctx, path, payload)
	if err != nil {
		a.logger.Error("Failed to set profile status",
			zap.String("session_name", a.sessionName),
			zap.Error(err))
		return fmt.Errorf("failed to set profile status: %w", err)
	}

	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	a.logger.Info("Profile status set successfully",
		zap.String("session_name", a.sessionName),
		zap.String("status", status))

	return nil
}

// SetProfilePicture sets profile picture
//
// file can be:
// - URL (https://...)
// - Base64 data (data:image/jpeg;base64,...)
func (a *ProfileManagerAdapter) SetProfilePicture(ctx context.Context, file channel.ProfilePictureFile) error {
	a.logger.Info("Setting profile picture",
		zap.String("session_name", a.sessionName),
		zap.String("mimetype", file.Mimetype),
		zap.String("filename", file.Filename))

	path := fmt.Sprintf("/api/%s/profile/picture", a.sessionName)

	payload := WAHAProfilePictureRequest{
		File: WAHAProfilePictureFile{
			Mimetype: file.Mimetype,
			Filename: file.Filename,
			URL:      file.URL,
		},
	}

	resp, err := a.client.Post(ctx, path, payload)
	if err != nil {
		a.logger.Error("Failed to set profile picture",
			zap.String("session_name", a.sessionName),
			zap.Error(err))
		return fmt.Errorf("failed to set profile picture: %w", err)
	}

	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	a.logger.Info("Profile picture set successfully",
		zap.String("session_name", a.sessionName))

	return nil
}

// DeleteProfilePicture deletes current profile picture
func (a *ProfileManagerAdapter) DeleteProfilePicture(ctx context.Context) error {
	a.logger.Info("Deleting profile picture",
		zap.String("session_name", a.sessionName))

	path := fmt.Sprintf("/api/%s/profile/picture", a.sessionName)

	resp, err := a.client.Delete(ctx, path)
	if err != nil {
		a.logger.Error("Failed to delete profile picture",
			zap.String("session_name", a.sessionName),
			zap.Error(err))
		return fmt.Errorf("failed to delete profile picture: %w", err)
	}

	if resp.StatusCode != 200 && resp.StatusCode != 204 {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	a.logger.Info("Profile picture deleted successfully",
		zap.String("session_name", a.sessionName))

	return nil
}

// WAHAProfile represents WAHA profile response
type WAHAProfile struct {
	ID      string  `json:"id"`      // e.g., "11111111111@c.us"
	Name    string  `json:"name"`    // Display name
	Status  *string `json:"status"`  // About text
	Picture *string `json:"picture"` // Profile picture URL
}

// WAHAProfilePictureRequest represents profile picture upload request
type WAHAProfilePictureRequest struct {
	File WAHAProfilePictureFile `json:"file"`
}

// WAHAProfilePictureFile represents a file to be uploaded
type WAHAProfilePictureFile struct {
	Mimetype string `json:"mimetype"` // e.g., "image/jpeg"
	Filename string `json:"filename"` // e.g., "profile.jpg"
	URL      string `json:"url"`      // URL or data URL
}
