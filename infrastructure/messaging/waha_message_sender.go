package messaging

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	messageports "github.com/caloi/ventros-crm/internal/application/message"
	"github.com/google/uuid"
)

// WAHAMessageSender implementa o envio de mensagens via WAHA
// Seguindo Single Responsibility Principle (SRP)
type WAHAMessageSender struct {
	httpClient   *http.Client
	baseURL      string
	apiKey       string
	sessionName  string
	capabilities *messageports.ChannelCapabilities
}

// NewWAHAMessageSender cria uma nova instância do WAHA sender
func NewWAHAMessageSender(baseURL, apiKey, sessionName string) *WAHAMessageSender {
	return &WAHAMessageSender{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL:     baseURL,
		apiKey:      apiKey,
		sessionName: sessionName,
		capabilities: &messageports.ChannelCapabilities{
			SupportedTypes: []messageports.MessageType{
				messageports.MessageTypeText,
				messageports.MessageTypeImage,
				messageports.MessageTypeAudio,
				messageports.MessageTypeVideo,
				messageports.MessageTypeDocument,
				messageports.MessageTypeLocation,
				messageports.MessageTypeContact,
			},
			MaxContentLength: 4096,
			MaxMediaSize:     64 * 1024 * 1024, // 64MB
			SupportedFormats: []string{
				"image/jpeg", "image/png", "image/gif", "image/webp",
				"audio/mpeg", "audio/ogg", "audio/wav",
				"video/mp4", "video/webm",
				"application/pdf", "application/msword",
			},
			Features: map[string]bool{
				"read_receipts":    true,
				"typing_indicator": true,
				"location_sharing": true,
				"contact_sharing":  true,
				"media_upload":     true,
				"voice_messages":   true,
			},
			RateLimits: map[string]int{
				"messages_per_minute": 60,
				"messages_per_hour":   1000,
			},
		},
	}
}

// WAHATextMessage representa uma mensagem de texto para WAHA
type WAHATextMessage struct {
	ChatID string `json:"chatId"`
	Text   string `json:"text"`
}

// WAHAMediaMessage representa uma mensagem de mídia para WAHA
type WAHAMediaMessage struct {
	ChatID   string `json:"chatId"`
	File     string `json:"file"`
	Caption  string `json:"caption,omitempty"`
	Filename string `json:"filename,omitempty"`
}

// WAHALocationMessage representa uma mensagem de localização para WAHA
type WAHALocationMessage struct {
	ChatID    string  `json:"chatId"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Title     string  `json:"title,omitempty"`
}

// WAHAContactMessage representa uma mensagem de contato para WAHA
type WAHAContactMessage struct {
	ChatID string `json:"chatId"`
	VCard  string `json:"vcard"`
}

// WAHAResponse representa a resposta da API WAHA
type WAHAResponse struct {
	ID        string                 `json:"id"`
	Timestamp int64                  `json:"timestamp"`
	Status    string                 `json:"status"`
	Error     *string                `json:"error,omitempty"`
	Data      map[string]interface{} `json:"data,omitempty"`
}

// SendMessage implementa o envio de mensagem via WAHA
func (w *WAHAMessageSender) SendMessage(ctx context.Context, message *messageports.OutboundMessage) (*messageports.SendResult, error) {
	startTime := time.Now()

	// Validar mensagem
	if err := w.ValidateMessage(message); err != nil {
		return nil, fmt.Errorf("message validation failed: %w", err)
	}

	// Construir chat ID (formato WhatsApp: número@c.us)
	chatID := w.buildChatID(message.ContactID)

	var response *WAHAResponse
	var err error

	// Enviar baseado no tipo de mensagem
	switch message.Type {
	case messageports.MessageTypeText:
		response, err = w.sendTextMessage(ctx, chatID, message)
	case messageports.MessageTypeImage, messageports.MessageTypeAudio, messageports.MessageTypeVideo, messageports.MessageTypeDocument:
		response, err = w.sendMediaMessage(ctx, chatID, message)
	case messageports.MessageTypeLocation:
		response, err = w.sendLocationMessage(ctx, chatID, message)
	case messageports.MessageTypeContact:
		response, err = w.sendContactMessage(ctx, chatID, message)
	default:
		return &messageports.SendResult{
			MessageID:  message.ID,
			Status:     "failed",
			Error:      wahaStringPtr(fmt.Sprintf("unsupported message type: %s", message.Type)),
			RetryCount: 0,
		}, fmt.Errorf("unsupported message type: %s", message.Type)
	}

	if err != nil {
		return &messageports.SendResult{
			MessageID:  message.ID,
			Status:     "failed",
			Error:      wahaStringPtr(err.Error()),
			RetryCount: 0,
		}, err
	}

	// Construir resultado
	result := &messageports.SendResult{
		MessageID:   message.ID,
		ExternalID:  &response.ID,
		Status:      "sent",
		DeliveredAt: wahaTimePtr(time.Unix(response.Timestamp/1000, 0)),
		RetryCount:  0,
		Metadata: map[string]interface{}{
			"waha_response": response.Data,
			"send_duration": time.Since(startTime).Milliseconds(),
		},
	}

	if response.Error != nil {
		result.Status = "failed"
		result.Error = response.Error
	}

	return result, nil
}

// SendBulkMessages implementa o envio em lote
func (w *WAHAMessageSender) SendBulkMessages(ctx context.Context, messages []*messageports.OutboundMessage) ([]*messageports.SendResult, error) {
	results := make([]*messageports.SendResult, len(messages))

	for i, message := range messages {
		result, err := w.SendMessage(ctx, message)
		if err != nil {
			results[i] = &messageports.SendResult{
				MessageID:  message.ID,
				Status:     "failed",
				Error:      wahaStringPtr(err.Error()),
				RetryCount: 0,
			}
		} else {
			results[i] = result
		}

		// Rate limiting - aguardar entre mensagens
		if i < len(messages)-1 {
			time.Sleep(1 * time.Second)
		}
	}

	return results, nil
}

// GetSupportedTypes retorna os tipos de mensagem suportados
func (w *WAHAMessageSender) GetSupportedTypes() []messageports.MessageType {
	return w.capabilities.SupportedTypes
}

// ValidateMessage valida uma mensagem antes do envio
func (w *WAHAMessageSender) ValidateMessage(message *messageports.OutboundMessage) error {
	// Verificar tipo suportado
	supported := false
	for _, supportedType := range w.capabilities.SupportedTypes {
		if message.Type == supportedType {
			supported = true
			break
		}
	}
	if !supported {
		return fmt.Errorf("message type %s not supported", message.Type)
	}

	// Verificar tamanho do conteúdo
	if len(message.Content) > w.capabilities.MaxContentLength {
		return fmt.Errorf("content length exceeds maximum of %d characters", w.capabilities.MaxContentLength)
	}

	// Validações específicas por tipo
	switch message.Type {
	case messageports.MessageTypeText:
		if message.Content == "" {
			return fmt.Errorf("text message content cannot be empty")
		}
	case messageports.MessageTypeImage, messageports.MessageTypeAudio, messageports.MessageTypeVideo, messageports.MessageTypeDocument:
		if message.MediaURL == nil {
			return fmt.Errorf("media message requires media_url")
		}
	case messageports.MessageTypeLocation:
		if message.Metadata == nil {
			return fmt.Errorf("location message requires metadata with latitude and longitude")
		}
		if _, ok := message.Metadata["latitude"]; !ok {
			return fmt.Errorf("location message requires latitude in metadata")
		}
		if _, ok := message.Metadata["longitude"]; !ok {
			return fmt.Errorf("location message requires longitude in metadata")
		}
	case messageports.MessageTypeContact:
		if message.Metadata == nil || message.Metadata["vcard"] == nil {
			return fmt.Errorf("contact message requires vcard in metadata")
		}
	}

	return nil
}

// GetChannelType retorna o tipo do canal
func (w *WAHAMessageSender) GetChannelType() string {
	return "waha"
}

// IsChannelSupported verifica se o canal é suportado
func (w *WAHAMessageSender) IsChannelSupported(channelID uuid.UUID) bool {
	// TODO: Implementar verificação real com repository
	return true
}

// GetChannelCapabilities retorna as capacidades do canal
func (w *WAHAMessageSender) GetChannelCapabilities(channelID uuid.UUID) (*messageports.ChannelCapabilities, error) {
	return w.capabilities, nil
}

// sendTextMessage envia uma mensagem de texto
func (w *WAHAMessageSender) sendTextMessage(ctx context.Context, chatID string, message *messageports.OutboundMessage) (*WAHAResponse, error) {
	payload := WAHATextMessage{
		ChatID: chatID,
		Text:   message.Content,
	}

	return w.makeAPICall(ctx, "POST", "/api/sendText", payload)
}

// sendMediaMessage envia uma mensagem de mídia
func (w *WAHAMessageSender) sendMediaMessage(ctx context.Context, chatID string, message *messageports.OutboundMessage) (*WAHAResponse, error) {
	if message.MediaURL == nil {
		return nil, fmt.Errorf("media URL is required for media messages")
	}

	payload := WAHAMediaMessage{
		ChatID:  chatID,
		File:    *message.MediaURL,
		Caption: message.Content,
	}

	var endpoint string
	switch message.Type {
	case messageports.MessageTypeImage:
		endpoint = "/api/sendImage"
	case messageports.MessageTypeAudio:
		endpoint = "/api/sendVoice"
	case messageports.MessageTypeVideo:
		endpoint = "/api/sendVideo"
	case messageports.MessageTypeDocument:
		endpoint = "/api/sendDocument"
		if message.Metadata != nil && message.Metadata["filename"] != nil {
			if filename, ok := message.Metadata["filename"].(string); ok {
				payload.Filename = filename
			}
		}
	default:
		return nil, fmt.Errorf("unsupported media type: %s", message.Type)
	}

	return w.makeAPICall(ctx, "POST", endpoint, payload)
}

// sendLocationMessage envia uma mensagem de localização
func (w *WAHAMessageSender) sendLocationMessage(ctx context.Context, chatID string, message *messageports.OutboundMessage) (*WAHAResponse, error) {
	lat, ok := message.Metadata["latitude"].(float64)
	if !ok {
		return nil, fmt.Errorf("invalid latitude in metadata")
	}

	lng, ok := message.Metadata["longitude"].(float64)
	if !ok {
		return nil, fmt.Errorf("invalid longitude in metadata")
	}

	payload := WAHALocationMessage{
		ChatID:    chatID,
		Latitude:  lat,
		Longitude: lng,
		Title:     message.Content,
	}

	return w.makeAPICall(ctx, "POST", "/api/sendLocation", payload)
}

// sendContactMessage envia uma mensagem de contato
func (w *WAHAMessageSender) sendContactMessage(ctx context.Context, chatID string, message *messageports.OutboundMessage) (*WAHAResponse, error) {
	vcard, ok := message.Metadata["vcard"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid vcard in metadata")
	}

	payload := WAHAContactMessage{
		ChatID: chatID,
		VCard:  vcard,
	}

	return w.makeAPICall(ctx, "POST", "/api/sendContactVcard", payload)
}

// makeAPICall faz uma chamada para a API WAHA
func (w *WAHAMessageSender) makeAPICall(ctx context.Context, method, endpoint string, payload interface{}) (*WAHAResponse, error) {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, method, w.baseURL+endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if w.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+w.apiKey)
	}

	resp, err := w.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	var wahaResp WAHAResponse
	if err := json.Unmarshal(body, &wahaResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &wahaResp, nil
}

// buildChatID constrói o chat ID do WhatsApp a partir do contact ID
func (w *WAHAMessageSender) buildChatID(contactID uuid.UUID) string {
	// TODO: Implementar lookup real do número do contato
	// Por enquanto, usar o UUID como placeholder
	return contactID.String() + "@c.us"
}

// Helper functions
func wahaStringPtr(s string) *string {
	return &s
}

func wahaTimePtr(t time.Time) *time.Time {
	return &t
}
