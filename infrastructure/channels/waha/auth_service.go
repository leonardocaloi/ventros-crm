package waha

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"

	"go.uber.org/zap"
)

// AuthService handles WAHA authentication operations
//
// Responsibilities:
// - Get QR code for WhatsApp pairing
// - Request authentication code (for phone number)
type AuthService struct {
	client *Client
	logger *zap.Logger
}

// NewAuthService creates a new auth service
func NewAuthService(client *Client, logger *zap.Logger) *AuthService {
	return &AuthService{
		client: client,
		logger: logger,
	}
}

// QRCodeFormat represents the format of QR code to return
type QRCodeFormat string

const (
	// QRCodeFormatImage returns QR code as PNG image
	QRCodeFormatImage QRCodeFormat = "image"

	// QRCodeFormatRaw returns QR code as raw text
	QRCodeFormatRaw QRCodeFormat = "raw"
)

// GetQRCode gets QR code for pairing WhatsApp
//
// Returns:
// - image: PNG image (Content-Type: image/png)
// - raw: Text string
//
// Usage:
//
//	// Get QR code as image (PNG)
//	qrImage, err := auth.GetQRCode(ctx, "default", QRCodeFormatImage)
//
//	// Get QR code as text
//	qrText, err := auth.GetQRCode(ctx, "default", QRCodeFormatRaw)
func (a *AuthService) GetQRCode(ctx context.Context, sessionName string, format QRCodeFormat) ([]byte, error) {
	a.logger.Info("Getting QR code for session",
		zap.String("session_name", sessionName),
		zap.String("format", string(format)))

	path := fmt.Sprintf("/api/%s/auth/qr?format=%s", sessionName, format)

	resp, err := a.client.Get(ctx, path)
	if err != nil {
		a.logger.Error("Failed to get QR code",
			zap.String("session_name", sessionName),
			zap.Error(err))
		return nil, fmt.Errorf("failed to get QR code: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read QR code response: %w", err)
	}

	// Check status code
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
	}

	a.logger.Info("QR code retrieved successfully",
		zap.String("session_name", sessionName),
		zap.String("format", string(format)),
		zap.Int("size_bytes", len(body)))

	return body, nil
}

// GetQRCodeAsBase64 gets QR code as base64-encoded string
//
// This is useful for:
// - Storing QR code in database
// - Sending QR code in JSON responses
// - Embedding QR code in HTML (data URLs)
//
// Usage:
//
//	qrBase64, err := auth.GetQRCodeAsBase64(ctx, "default", QRCodeFormatImage)
//	// Use in HTML: <img src="data:image/png;base64,{qrBase64}" />
func (a *AuthService) GetQRCodeAsBase64(ctx context.Context, sessionName string, format QRCodeFormat) (string, error) {
	qrData, err := a.GetQRCode(ctx, sessionName, format)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(qrData), nil
}

// AuthCodeRequest represents a request to send authentication code
type AuthCodeRequest struct {
	PhoneNumber string  `json:"phoneNumber"`      // e.g., "12132132130"
	Method      *string `json:"method,omitempty"` // "sms" or "voice"
}

// RequestAuthCode requests authentication code via SMS or voice
//
// This is an alternative to QR code scanning.
// Use this when you want to authenticate via phone number instead of scanning QR.
//
// Example:
//
//	req := AuthCodeRequest{
//	    PhoneNumber: "5511999999999",
//	    Method: ptr("sms"), // or "voice"
//	}
//	err := auth.RequestAuthCode(ctx, "default", req)
func (a *AuthService) RequestAuthCode(ctx context.Context, sessionName string, req AuthCodeRequest) error {
	a.logger.Info("Requesting auth code",
		zap.String("session_name", sessionName),
		zap.String("phone_number", req.PhoneNumber))

	path := fmt.Sprintf("/api/%s/auth/request-code", sessionName)

	resp, err := a.client.Post(ctx, path, req)
	if err != nil {
		a.logger.Error("Failed to request auth code",
			zap.String("session_name", sessionName),
			zap.Error(err))
		return fmt.Errorf("failed to request auth code: %w", err)
	}

	// Check if response is 201 Created
	if resp.StatusCode != 201 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
	}

	a.logger.Info("Auth code requested successfully",
		zap.String("session_name", sessionName),
		zap.String("phone_number", req.PhoneNumber))

	return nil
}

// GetMe gets information about the authenticated account
//
// Returns nil if session is not authenticated yet
func (a *AuthService) GetMe(ctx context.Context, sessionName string) (*SessionMe, error) {
	path := fmt.Sprintf("/api/sessions/%s/me", sessionName)

	resp, err := a.client.Get(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("failed to get me: %w", err)
	}

	var me SessionMe
	if err := a.client.ParseResponse(resp, &me); err != nil {
		return nil, fmt.Errorf("failed to parse me response: %w", err)
	}

	return &me, nil
}

// IsAuthenticated checks if session is authenticated
func (a *AuthService) IsAuthenticated(ctx context.Context, sessionName string) (bool, error) {
	me, err := a.GetMe(ctx, sessionName)
	if err != nil {
		// If error, assume not authenticated
		return false, nil
	}

	// Authenticated if we have an ID
	return me != nil && me.ID != "", nil
}
