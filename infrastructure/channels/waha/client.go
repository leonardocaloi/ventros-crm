package waha

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"go.uber.org/zap"
)

// WAHAClient é o cliente para interagir com a API WAHA
type WAHAClient struct {
	baseURL    string
	token      string
	httpClient *http.Client
	logger     *zap.Logger
}

// NewWAHAClient cria um novo cliente WAHA
func NewWAHAClient(baseURL, token string, logger *zap.Logger) *WAHAClient {
	return &WAHAClient{
		baseURL: baseURL,
		token:   token,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger: logger,
	}
}

// NewWAHAClientFromEnv cria um novo cliente WAHA usando variáveis de ambiente
func NewWAHAClientFromEnv(logger *zap.Logger) *WAHAClient {
	// A API key vem de variável de ambiente
	baseURL := getEnvOrDefault("WAHA_BASE_URL", "http://localhost:3000")
	token := getEnvOrDefault("WAHA_API_KEY", "")
	
	return NewWAHAClient(baseURL, token, logger)
}

// getEnvOrDefault retorna valor da variável de ambiente ou valor padrão
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// SessionInfo representa informações da sessão
type SessionInfo struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Config struct {
		Webhooks []WebhookConfig `json:"webhooks"`
	} `json:"config"`
}

// WebhookConfig representa configuração de webhook
type WebhookConfig struct {
	URL    string   `json:"url"`
	Events []string `json:"events"`
}

// SendMessageRequest representa uma requisição para enviar mensagem
type SendMessageRequest struct {
	ChatID string `json:"chatId"`
	Text   string `json:"text"`
}

// SendMessageResponse representa a resposta do envio de mensagem
type SendMessageResponse struct {
	ID        string `json:"id"`
	Timestamp int64  `json:"timestamp"`
	From      string `json:"from"`
	To        string `json:"to"`
	Body      string `json:"body"`
}

// WAHAWebhookEvent representa um evento recebido via webhook
type WAHAWebhookEvent struct {
	Event   string      `json:"event"`
	Session string      `json:"session"`
	Payload interface{} `json:"payload"`
}

// MessagePayload representa o payload de uma mensagem
type MessagePayload struct {
	ID        string `json:"id"`
	Timestamp int64  `json:"timestamp"`
	From      string `json:"from"`
	To        string `json:"to"`
	Body      string `json:"body"`
	Type      string `json:"type"`
	MimeType  string `json:"mimeType,omitempty"`
	MediaURL  string `json:"mediaUrl,omitempty"`
}

// GetSessions retorna todas as sessões
func (c *WAHAClient) GetSessions(ctx context.Context) ([]SessionInfo, error) {
	url := fmt.Sprintf("%s/api/sessions", c.baseURL)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	c.setAuthHeaders(req)
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}
	
	var sessions []SessionInfo
	if err := json.NewDecoder(resp.Body).Decode(&sessions); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	
	return sessions, nil
}

// GetSession retorna informações de uma sessão específica
func (c *WAHAClient) GetSession(ctx context.Context, sessionID string) (*SessionInfo, error) {
	url := fmt.Sprintf("%s/api/sessions/%s", c.baseURL, sessionID)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	c.setAuthHeaders(req)
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}
	
	var session SessionInfo
	if err := json.NewDecoder(resp.Body).Decode(&session); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	
	return &session, nil
}

// StartSession inicia uma nova sessão
func (c *WAHAClient) StartSession(ctx context.Context, sessionID string, config map[string]interface{}) error {
	url := fmt.Sprintf("%s/api/sessions/%s/start", c.baseURL, sessionID)
	
	payload := map[string]interface{}{
		"name":   sessionID,
		"config": config,
	}
	
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}
	
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	
	c.setAuthHeaders(req)
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}
	
	c.logger.Info("WAHA session started", zap.String("session_id", sessionID))
	return nil
}

// StopSession para uma sessão
func (c *WAHAClient) StopSession(ctx context.Context, sessionID string) error {
	url := fmt.Sprintf("%s/api/sessions/%s/stop", c.baseURL, sessionID)
	
	req, err := http.NewRequestWithContext(ctx, "POST", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	
	c.setAuthHeaders(req)
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}
	
	c.logger.Info("WAHA session stopped", zap.String("session_id", sessionID))
	return nil
}

// SetWebhook configura webhook para uma sessão
func (c *WAHAClient) SetWebhook(ctx context.Context, sessionID, webhookURL string, events []string) error {
	url := fmt.Sprintf("%s/api/sessions/%s/config", c.baseURL, sessionID)
	
	payload := map[string]interface{}{
		"webhooks": []WebhookConfig{
			{
				URL:    webhookURL,
				Events: events,
			},
		},
	}
	
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}
	
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	
	c.setAuthHeaders(req)
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}
	
	c.logger.Info("WAHA webhook configured", 
		zap.String("session_id", sessionID),
		zap.String("webhook_url", webhookURL),
		zap.Strings("events", events))
	
	return nil
}

// SendMessage envia uma mensagem de texto
func (c *WAHAClient) SendMessage(ctx context.Context, sessionID string, req SendMessageRequest) (*SendMessageResponse, error) {
	url := fmt.Sprintf("%s/api/sessions/%s/messages/text", c.baseURL, sessionID)
	
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}
	
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	c.setAuthHeaders(httpReq)
	httpReq.Header.Set("Content-Type", "application/json")
	
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}
	
	var response SendMessageResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	
	c.logger.Debug("Message sent via WAHA", 
		zap.String("session_id", sessionID),
		zap.String("chat_id", req.ChatID),
		zap.String("message_id", response.ID))
	
	return &response, nil
}

// GetQRCode obtém o QR code para autenticação
func (c *WAHAClient) GetQRCode(ctx context.Context, sessionID string) ([]byte, error) {
	url := fmt.Sprintf("%s/api/sessions/%s/auth/qr", c.baseURL, sessionID)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	c.setAuthHeaders(req)
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}
	
	qrData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read QR code: %w", err)
	}
	
	return qrData, nil
}

// setAuthHeaders define os headers de autenticação
func (c *WAHAClient) setAuthHeaders(req *http.Request) {
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
}

// ParseWebhookEvent parseia um evento de webhook
func ParseWebhookEvent(body []byte) (*WAHAWebhookEvent, error) {
	var event WAHAWebhookEvent
	if err := json.Unmarshal(body, &event); err != nil {
		return nil, fmt.Errorf("failed to parse webhook event: %w", err)
	}
	
	return &event, nil
}

// GetDefaultWebhookEvents retorna os eventos padrão para configurar no webhook
func GetDefaultWebhookEvents() []string {
	return []string{
		"message",
		"message.any",
		"message.ack",
		"message.reaction",
		"message.edited",
		"call.received",
		"call.accepted",
		"call.rejected",
		"label.upsert",
		"label.deleted",
		"label.chat.added",
		"label.chat.deleted",
		"group.v2.join",
		"group.v2.leave",
		"group.v2.update",
		"group.v2.participants",
	}
}

