package channel

import (
	"context"
	"fmt"
	"time"

	"github.com/caloi/ventros-crm/infrastructure/channels/waha"
	"github.com/caloi/ventros-crm/internal/domain/channel"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// GetQRCodeCommand contains the data to get QR code
type GetQRCodeCommand struct {
	ChannelID uuid.UUID
	Format    string // "image" (PNG) or "raw" (text)
}

// GetQRCodeResult returns the QR code
type GetQRCodeResult struct {
	QRCode    string    `json:"qr_code"`     // Base64-encoded if image, raw text if raw
	Format    string    `json:"format"`      // "image" or "raw"
	ExpiresAt time.Time `json:"expires_at"`  // QR code expiration time
}

// GetQRCodeUseCase implements the use case to get QR code for WhatsApp Business channels
//
// This use case is ONLY for TypeWhatsAppBusiness channels (auto mode).
// TypeWAHA channels (manual mode) do NOT support QR code generation.
type GetQRCodeUseCase struct {
	channelRepo channel.Repository
	logger      *zap.Logger
}

// NewGetQRCodeUseCase creates a new instance
func NewGetQRCodeUseCase(
	channelRepo channel.Repository,
	logger *zap.Logger,
) *GetQRCodeUseCase {
	return &GetQRCodeUseCase{
		channelRepo: channelRepo,
		logger:      logger,
	}
}

// Execute executes the use case
func (uc *GetQRCodeUseCase) Execute(ctx context.Context, cmd GetQRCodeCommand) (*GetQRCodeResult, error) {
	// Validation
	if cmd.ChannelID == uuid.Nil {
		return nil, fmt.Errorf("channelID is required")
	}

	// Get channel from repository
	ch, err := uc.channelRepo.GetByID(cmd.ChannelID)
	if err != nil {
		return nil, fmt.Errorf("failed to get channel: %w", err)
	}

	// IMPORTANT: Only WhatsApp Business channels (auto mode) support QR code
	if ch.Type != channel.TypeWhatsAppBusiness {
		return nil, fmt.Errorf("QR code is only available for whatsapp_business channels, got: %s", ch.Type)
	}

	// Check connection mode
	if ch.ConnectionMode != channel.ConnectionModeAuto {
		return nil, fmt.Errorf("QR code is only available for auto connection mode, got: %s", ch.ConnectionMode)
	}

	// Check channel status - must be connecting or active
	if ch.Status != channel.StatusConnecting && ch.Status != channel.StatusActive {
		return nil, fmt.Errorf("channel must be in connecting or active status to get QR code, got: %s", ch.Status)
	}

	// Set default format
	format := cmd.Format
	if format == "" {
		format = "image" // default to image (PNG)
	}

	// Validate format
	if format != "image" && format != "raw" {
		return nil, fmt.Errorf("invalid format: %s (must be 'image' or 'raw')", format)
	}

	// Create auth service
	authService, err := uc.createAuthService(ch)
	if err != nil {
		return nil, fmt.Errorf("failed to create auth service: %w", err)
	}

	// Get session name (ExternalID is the WAHA session name)
	sessionName := ch.ExternalID
	if sessionName == "" {
		return nil, fmt.Errorf("channel has no WAHA session configured")
	}

	// Get QR code as base64
	var qrFormat waha.QRCodeFormat
	if format == "image" {
		qrFormat = waha.QRCodeFormatImage
	} else {
		qrFormat = waha.QRCodeFormatRaw
	}

	qrCodeBase64, err := authService.GetQRCodeAsBase64(ctx, sessionName, qrFormat)
	if err != nil {
		uc.logger.Error("Failed to get QR code",
			zap.String("channel_id", cmd.ChannelID.String()),
			zap.String("session_name", sessionName),
			zap.Error(err))
		return nil, fmt.Errorf("failed to get QR code: %w", err)
	}

	// QR codes typically expire after 60 seconds
	expiresAt := time.Now().Add(60 * time.Second)

	uc.logger.Info("QR code retrieved successfully",
		zap.String("channel_id", cmd.ChannelID.String()),
		zap.String("session_name", sessionName),
		zap.String("format", format))

	return &GetQRCodeResult{
		QRCode:    qrCodeBase64,
		Format:    format,
		ExpiresAt: expiresAt,
	}, nil
}

// createAuthService creates an AuthService based on channel configuration
func (uc *GetQRCodeUseCase) createAuthService(ch *channel.Channel) (*waha.AuthService, error) {
	// Get WAHA config
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

	// Create auth service
	authService := waha.NewAuthService(wahaClient, uc.logger)

	return authService, nil
}
