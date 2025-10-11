package waha

import (
	"context"
	"fmt"

	"go.uber.org/zap"
)

// SessionManager manages WAHA session lifecycle
//
// Responsibilities:
// - Create new sessions in WAHA
// - Start/Stop/Restart sessions
// - Delete sessions
// - Get session status
//
// Used for "auto" connection mode where system manages WAHA sessions
type SessionManager struct {
	client *Client
	logger *zap.Logger
}

// NewSessionManager creates a new session manager
func NewSessionManager(client *Client, logger *zap.Logger) *SessionManager {
	return &SessionManager{
		client: client,
		logger: logger,
	}
}

// SessionConfig represents WAHA session configuration
type SessionConfig struct {
	Name   string               `json:"name"`
	Start  bool                 `json:"start"`
	Config SessionConfigOptions `json:"config"`
}

// SessionConfigOptions represents session configuration options
type SessionConfigOptions struct {
	Metadata map[string]string      `json:"metadata,omitempty"`
	Proxy    *string                `json:"proxy,omitempty"`
	Debug    bool                   `json:"debug,omitempty"`
	Webhooks []SessionWebhookConfig `json:"webhooks"`
}

// SessionWebhookConfig represents webhook configuration for a session
type SessionWebhookConfig struct {
	URL           string            `json:"url"`
	Events        []string          `json:"events"`
	HMAC          *string           `json:"hmac,omitempty"`
	Retries       *int              `json:"retries,omitempty"`
	CustomHeaders map[string]string `json:"customHeaders,omitempty"`
}

// SessionResponse represents a WAHA session response
type SessionResponse struct {
	Name   string               `json:"name"`
	Status string               `json:"status"` // STOPPED, STARTING, SCAN_QR_CODE, WORKING, FAILED
	Me     *SessionMe           `json:"me,omitempty"`
	Config SessionConfigOptions `json:"config"`
}

// SessionMe represents authenticated user info
type SessionMe struct {
	ID       string `json:"id"`  // 11111111111@c.us
	LID      string `json:"lid"` // 123123@lid
	PushName string `json:"pushName"`
}

// CreateSession creates a new session in WAHA
//
// This is used for "auto" connection mode where system manages sessions.
// After creation, the session will emit QR code events that need to be handled.
//
// Example:
//
//	config := SessionConfig{
//	    Name: "channel-123",
//	    Start: true,
//	    Config: SessionConfigOptions{
//	        Webhooks: []WebhookConfig{{
//	            URL: "https://api.crm.ventros.cloud/webhooks/waha",
//	            Events: []string{"message", "session.status"},
//	        }},
//	    },
//	}
//	session, err := sm.CreateSession(ctx, config)
func (sm *SessionManager) CreateSession(ctx context.Context, config SessionConfig) (*SessionResponse, error) {
	sm.logger.Info("Creating WAHA session",
		zap.String("session_name", config.Name),
		zap.Bool("start", config.Start))

	resp, err := sm.client.Post(ctx, "/api/sessions", config)
	if err != nil {
		sm.logger.Error("Failed to create WAHA session",
			zap.String("session_name", config.Name),
			zap.Error(err))
		return nil, fmt.Errorf("failed to create WAHA session: %w", err)
	}

	var session SessionResponse
	if err := sm.client.ParseResponse(resp, &session); err != nil {
		return nil, fmt.Errorf("failed to parse session response: %w", err)
	}

	sm.logger.Info("WAHA session created successfully",
		zap.String("session_name", session.Name),
		zap.String("status", session.Status))

	return &session, nil
}

// GetSession gets session information
func (sm *SessionManager) GetSession(ctx context.Context, sessionName string) (*SessionResponse, error) {
	path := fmt.Sprintf("/api/sessions/%s", sessionName)

	resp, err := sm.client.Get(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	var session SessionResponse
	if err := sm.client.ParseResponse(resp, &session); err != nil {
		return nil, fmt.Errorf("failed to parse session response: %w", err)
	}

	return &session, nil
}

// StartSession starts a session
//
// The session must exist. This is idempotent.
func (sm *SessionManager) StartSession(ctx context.Context, sessionName string) (*SessionResponse, error) {
	sm.logger.Info("Starting WAHA session", zap.String("session_name", sessionName))

	path := fmt.Sprintf("/api/sessions/%s/start", sessionName)
	resp, err := sm.client.Post(ctx, path, nil)
	if err != nil {
		sm.logger.Error("Failed to start WAHA session",
			zap.String("session_name", sessionName),
			zap.Error(err))
		return nil, fmt.Errorf("failed to start session: %w", err)
	}

	var session SessionResponse
	if err := sm.client.ParseResponse(resp, &session); err != nil {
		return nil, fmt.Errorf("failed to parse session response: %w", err)
	}

	sm.logger.Info("WAHA session started",
		zap.String("session_name", sessionName),
		zap.String("status", session.Status))

	return &session, nil
}

// StopSession stops a session
//
// This is idempotent.
func (sm *SessionManager) StopSession(ctx context.Context, sessionName string) (*SessionResponse, error) {
	sm.logger.Info("Stopping WAHA session", zap.String("session_name", sessionName))

	path := fmt.Sprintf("/api/sessions/%s/stop", sessionName)
	resp, err := sm.client.Post(ctx, path, nil)
	if err != nil {
		sm.logger.Error("Failed to stop WAHA session",
			zap.String("session_name", sessionName),
			zap.Error(err))
		return nil, fmt.Errorf("failed to stop session: %w", err)
	}

	var session SessionResponse
	if err := sm.client.ParseResponse(resp, &session); err != nil {
		return nil, fmt.Errorf("failed to parse session response: %w", err)
	}

	sm.logger.Info("WAHA session stopped",
		zap.String("session_name", sessionName),
		zap.String("status", session.Status))

	return &session, nil
}

// RestartSession restarts a session
func (sm *SessionManager) RestartSession(ctx context.Context, sessionName string) (*SessionResponse, error) {
	sm.logger.Info("Restarting WAHA session", zap.String("session_name", sessionName))

	path := fmt.Sprintf("/api/sessions/%s/restart", sessionName)
	resp, err := sm.client.Post(ctx, path, nil)
	if err != nil {
		sm.logger.Error("Failed to restart WAHA session",
			zap.String("session_name", sessionName),
			zap.Error(err))
		return nil, fmt.Errorf("failed to restart session: %w", err)
	}

	var session SessionResponse
	if err := sm.client.ParseResponse(resp, &session); err != nil {
		return nil, fmt.Errorf("failed to parse session response: %w", err)
	}

	sm.logger.Info("WAHA session restarted",
		zap.String("session_name", sessionName),
		zap.String("status", session.Status))

	return &session, nil
}

// LogoutSession logs out from the session
//
// Restarts the session if it was not STOPPED
func (sm *SessionManager) LogoutSession(ctx context.Context, sessionName string) (*SessionResponse, error) {
	sm.logger.Info("Logging out WAHA session", zap.String("session_name", sessionName))

	path := fmt.Sprintf("/api/sessions/%s/logout", sessionName)
	resp, err := sm.client.Post(ctx, path, nil)
	if err != nil {
		sm.logger.Error("Failed to logout WAHA session",
			zap.String("session_name", sessionName),
			zap.Error(err))
		return nil, fmt.Errorf("failed to logout session: %w", err)
	}

	var session SessionResponse
	if err := sm.client.ParseResponse(resp, &session); err != nil {
		return nil, fmt.Errorf("failed to parse session response: %w", err)
	}

	sm.logger.Info("WAHA session logged out",
		zap.String("session_name", sessionName),
		zap.String("status", session.Status))

	return &session, nil
}

// DeleteSession deletes a session
//
// Stop and logout as well. This is idempotent.
func (sm *SessionManager) DeleteSession(ctx context.Context, sessionName string) error {
	sm.logger.Info("Deleting WAHA session", zap.String("session_name", sessionName))

	path := fmt.Sprintf("/api/sessions/%s", sessionName)
	resp, err := sm.client.Delete(ctx, path)
	if err != nil {
		sm.logger.Error("Failed to delete WAHA session",
			zap.String("session_name", sessionName),
			zap.Error(err))
		return fmt.Errorf("failed to delete session: %w", err)
	}

	// Check if response is 200 OK
	if resp.StatusCode != 200 {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	sm.logger.Info("WAHA session deleted successfully",
		zap.String("session_name", sessionName))

	return nil
}

// ListSessions lists all sessions
func (sm *SessionManager) ListSessions(ctx context.Context, includeAll bool) ([]SessionResponse, error) {
	path := "/api/sessions"
	if includeAll {
		path += "?all=true"
	}

	resp, err := sm.client.Get(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("failed to list sessions: %w", err)
	}

	var sessions []SessionResponse
	if err := sm.client.ParseResponse(resp, &sessions); err != nil {
		return nil, fmt.Errorf("failed to parse sessions response: %w", err)
	}

	return sessions, nil
}

// UpdateSession updates session configuration
func (sm *SessionManager) UpdateSession(ctx context.Context, sessionName string, config SessionConfigOptions) (*SessionResponse, error) {
	sm.logger.Info("Updating WAHA session", zap.String("session_name", sessionName))

	path := fmt.Sprintf("/api/sessions/%s", sessionName)
	payload := map[string]interface{}{
		"config": config,
	}

	resp, err := sm.client.Put(ctx, path, payload)
	if err != nil {
		sm.logger.Error("Failed to update WAHA session",
			zap.String("session_name", sessionName),
			zap.Error(err))
		return nil, fmt.Errorf("failed to update session: %w", err)
	}

	var session SessionResponse
	if err := sm.client.ParseResponse(resp, &session); err != nil {
		return nil, fmt.Errorf("failed to parse session response: %w", err)
	}

	sm.logger.Info("WAHA session updated successfully",
		zap.String("session_name", sessionName))

	return &session, nil
}
