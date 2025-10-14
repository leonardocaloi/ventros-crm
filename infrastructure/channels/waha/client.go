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
			Timeout: 5 * time.Minute, // 5 minutos para permitir envio de vídeos grandes
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
	FromMe    bool   `json:"fromMe"`
	To        string `json:"to"`
	Body      string `json:"body"`
	Type      string `json:"type"`
	MimeType  string `json:"mimeType,omitempty"`
	MediaURL  string `json:"mediaUrl,omitempty"`
	HasMedia  bool   `json:"hasMedia"`
	Ack       int    `json:"ack,omitempty"`
	AckName   string `json:"ackName,omitempty"`
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

// SendTextRequest representa uma requisição para enviar mensagem de texto
type SendTextRequest struct {
	ChatID                 string  `json:"chatId"`
	Text                   string  `json:"text"`
	ReplyTo                *string `json:"reply_to,omitempty"`
	LinkPreview            *bool   `json:"linkPreview,omitempty"`
	LinkPreviewHighQuality *bool   `json:"linkPreviewHighQuality,omitempty"`
}

// SendFileRequest representa uma requisição para enviar arquivo/mídia
type SendFileRequest struct {
	ChatID  string      `json:"chatId"`
	File    FilePayload `json:"file"`
	ReplyTo *string     `json:"reply_to,omitempty"`
	Caption *string     `json:"caption,omitempty"`
}

// FilePayload representa o arquivo a ser enviado
type FilePayload struct {
	Mimetype string `json:"mimetype"`
	Filename string `json:"filename,omitempty"`
	URL      string `json:"url,omitempty"`  // URL do arquivo
	Data     string `json:"data,omitempty"` // Base64 data
}

// SendVoiceRequest representa uma requisição para enviar áudio/voz
type SendVoiceRequest struct {
	ChatID  string      `json:"chatId"`
	File    FilePayload `json:"file"`
	ReplyTo *string     `json:"reply_to,omitempty"`
	Convert *bool       `json:"convert,omitempty"` // Converter para formato WhatsApp
}

// SendVideoRequest representa uma requisição para enviar vídeo
type SendVideoRequest struct {
	ChatID  string      `json:"chatId"`
	File    FilePayload `json:"file"`
	ReplyTo *string     `json:"reply_to,omitempty"`
	Caption *string     `json:"caption,omitempty"`
	AsNote  *bool       `json:"asNote,omitempty"`  // Enviar como video note
	Convert *bool       `json:"convert,omitempty"` // Converter para formato WhatsApp
}

// SendLocationRequest representa uma requisição para enviar localização
type SendLocationRequest struct {
	ChatID    string  `json:"chatId"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Title     *string `json:"title,omitempty"`
}

// SendContactRequest representa uma requisição para enviar contato
type SendContactRequest struct {
	ChatID   string           `json:"chatId"`
	Contacts []ContactPayload `json:"contacts"` // WAHA espera array de contacts
}

// ContactPayload representa um contato individual
type ContactPayload struct {
	VCard string `json:"vcard"`
}

// SendMessage envia uma mensagem de texto (legacy method - mantido para compatibilidade)
func (c *WAHAClient) SendMessage(ctx context.Context, sessionID string, req SendMessageRequest) (*SendMessageResponse, error) {
	textReq := SendTextRequest{
		ChatID: req.ChatID,
		Text:   req.Text,
	}
	return c.SendText(ctx, sessionID, textReq)
}

// SendText envia uma mensagem de texto
func (c *WAHAClient) SendText(ctx context.Context, sessionID string, req SendTextRequest) (*SendMessageResponse, error) {
	url := fmt.Sprintf("%s/api/sendText", c.baseURL)

	// Adiciona session ao request
	payload := map[string]interface{}{
		"chatId":  req.ChatID,
		"text":    req.Text,
		"session": sessionID,
	}
	if req.ReplyTo != nil {
		payload["reply_to"] = *req.ReplyTo
	}
	if req.LinkPreview != nil {
		payload["linkPreview"] = *req.LinkPreview
	}
	if req.LinkPreviewHighQuality != nil {
		payload["linkPreviewHighQuality"] = *req.LinkPreviewHighQuality
	}

	jsonData, err := json.Marshal(payload)
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

	c.logger.Debug("Text message sent via WAHA",
		zap.String("session_id", sessionID),
		zap.String("chat_id", req.ChatID),
		zap.String("message_id", response.ID))

	return &response, nil
}

// SendImage envia uma imagem
func (c *WAHAClient) SendImage(ctx context.Context, sessionID string, req SendFileRequest) (*SendMessageResponse, error) {
	return c.sendFile(ctx, sessionID, "/api/sendImage", req)
}

// SendFile envia um arquivo/documento
func (c *WAHAClient) SendFile(ctx context.Context, sessionID string, req SendFileRequest) (*SendMessageResponse, error) {
	return c.sendFile(ctx, sessionID, "/api/sendFile", req)
}

// SendVoice envia um áudio/voz
func (c *WAHAClient) SendVoice(ctx context.Context, sessionID string, req SendVoiceRequest) (*SendMessageResponse, error) {
	url := fmt.Sprintf("%s/api/sendVoice", c.baseURL)

	payload := map[string]interface{}{
		"chatId":  req.ChatID,
		"file":    req.File,
		"session": sessionID,
	}
	if req.ReplyTo != nil {
		payload["reply_to"] = *req.ReplyTo
	}
	if req.Convert != nil {
		payload["convert"] = *req.Convert
	}

	return c.makeRequest(ctx, url, payload)
}

// SendVideo envia um vídeo
func (c *WAHAClient) SendVideo(ctx context.Context, sessionID string, req SendVideoRequest) (*SendMessageResponse, error) {
	url := fmt.Sprintf("%s/api/sendVideo", c.baseURL)

	payload := map[string]interface{}{
		"chatId":  req.ChatID,
		"file":    req.File,
		"session": sessionID,
	}
	if req.ReplyTo != nil {
		payload["reply_to"] = *req.ReplyTo
	}
	if req.Caption != nil {
		payload["caption"] = *req.Caption
	}
	if req.AsNote != nil {
		payload["asNote"] = *req.AsNote
	}
	if req.Convert != nil {
		payload["convert"] = *req.Convert
	}

	return c.makeRequest(ctx, url, payload)
}

// SendLocation envia uma localização
func (c *WAHAClient) SendLocation(ctx context.Context, sessionID string, req SendLocationRequest) (*SendMessageResponse, error) {
	url := fmt.Sprintf("%s/api/sendLocation", c.baseURL)

	payload := map[string]interface{}{
		"chatId":    req.ChatID,
		"latitude":  req.Latitude,
		"longitude": req.Longitude,
		"session":   sessionID,
	}
	if req.Title != nil {
		payload["title"] = *req.Title
	}

	return c.makeRequest(ctx, url, payload)
}

// SendContact envia um contato
func (c *WAHAClient) SendContact(ctx context.Context, sessionID string, req SendContactRequest) (*SendMessageResponse, error) {
	url := fmt.Sprintf("%s/api/sendContactVcard", c.baseURL)

	payload := map[string]interface{}{
		"chatId":   req.ChatID,
		"contacts": req.Contacts, // Array de contacts como WAHA espera
		"session":  sessionID,
	}

	return c.makeRequest(ctx, url, payload)
}

// sendFile é um helper para enviar arquivos (imagem, documento)
func (c *WAHAClient) sendFile(ctx context.Context, sessionID, endpoint string, req SendFileRequest) (*SendMessageResponse, error) {
	url := fmt.Sprintf("%s%s", c.baseURL, endpoint)

	payload := map[string]interface{}{
		"chatId":  req.ChatID,
		"file":    req.File,
		"session": sessionID,
	}
	if req.ReplyTo != nil {
		payload["reply_to"] = *req.ReplyTo
	}
	if req.Caption != nil {
		payload["caption"] = *req.Caption
	}

	return c.makeRequest(ctx, url, payload)
}

// makeRequest é um helper para fazer requests HTTP à API WAHA
func (c *WAHAClient) makeRequest(ctx context.Context, url string, payload interface{}) (*SendMessageResponse, error) {
	jsonData, err := json.Marshal(payload)
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

	c.logger.Debug("Message sent via WAHA", zap.String("url", url), zap.String("message_id", response.ID))

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
		req.Header.Set("X-Api-Key", c.token)
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
		"session.status", // Importante para detectar quando sessão fica ativa
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

// ChatOverview representa um chat na visão geral
type ChatOverview struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	Picture     string          `json:"picture"`
	LastMessage *MessagePayload `json:"lastMessage"`
}

// GetChatsOverview retorna visão geral de todos os chats
func (c *WAHAClient) GetChatsOverview(ctx context.Context, sessionID string, limit, offset int) ([]ChatOverview, error) {
	url := fmt.Sprintf("%s/api/%s/chats/overview?limit=%d&offset=%d", c.baseURL, sessionID, limit, offset)

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

	var chats []ChatOverview
	if err := json.NewDecoder(resp.Body).Decode(&chats); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return chats, nil
}

// GetChatMessages retorna mensagens de um chat específico
func (c *WAHAClient) GetChatMessages(ctx context.Context, sessionID, chatID string, limit int, downloadMedia bool) ([]MessagePayload, error) {
	return c.GetChatMessagesWithFilter(ctx, sessionID, chatID, limit, downloadMedia, 0, 0)
}

// GetChatMessagesWithFilter busca mensagens com filtro de timestamp
// timestampGte: mensagens >= este timestamp Unix (0 = sem filtro)
// timestampLte: mensagens <= este timestamp Unix (0 = sem filtro)
func (c *WAHAClient) GetChatMessagesWithFilter(ctx context.Context, sessionID, chatID string, limit int, downloadMedia bool, timestampGte, timestampLte int64) ([]MessagePayload, error) {
	url := fmt.Sprintf("%s/api/%s/chats/%s/messages?limit=%d&downloadMedia=%t",
		c.baseURL, sessionID, chatID, limit, downloadMedia)

	// Adicionar filtros de timestamp se especificados
	if timestampGte > 0 {
		url += fmt.Sprintf("&filter.timestamp.gte=%d", timestampGte)
	}
	if timestampLte > 0 {
		url += fmt.Sprintf("&filter.timestamp.lte=%d", timestampLte)
	}

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

	var messages []MessagePayload
	if err := json.NewDecoder(resp.Body).Decode(&messages); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return messages, nil
}

// HealthCheck verifica se a sessão está ativa e funcionando
func (c *WAHAClient) HealthCheck(ctx context.Context, sessionID string) (bool, string, error) {
	sessionInfo, err := c.GetSession(ctx, sessionID)
	if err != nil {
		return false, "error", fmt.Errorf("failed to get session: %w", err)
	}

	// Status WORKING significa que está ativo e conectado
	isHealthy := sessionInfo.Status == "WORKING"
	return isHealthy, sessionInfo.Status, nil
}

// CreateSession cria uma nova sessão WAHA
func (c *WAHAClient) CreateSession(ctx context.Context, sessionID string, config map[string]interface{}) error {
	url := fmt.Sprintf("%s/api/sessions", c.baseURL)

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

	c.logger.Info("WAHA session created", zap.String("session_id", sessionID))
	return nil
}
