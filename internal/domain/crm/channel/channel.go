package channel

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

type ChannelType string

const (
	// TypeWAHA - Manual WAHA connection (user provides base_url, token, session_id)
	// Generic for users who already have WAHA running anywhere
	TypeWAHA ChannelType = "waha"

	// TypeWhatsAppBusiness - Auto-managed WhatsApp (system creates WAHA session, returns QR code)
	// Hides WAHA from end-user, just "Connect WhatsApp"
	TypeWhatsAppBusiness ChannelType = "whatsapp_business"

	// Cloud API types
	TypeWhatsApp  ChannelType = "whatsapp"  // WhatsApp Cloud API
	TypeMessenger ChannelType = "messenger" // Facebook Messenger
	TypeInstagram ChannelType = "instagram" // Instagram DM

	// Other messaging platforms
	TypeTelegram  ChannelType = "telegram"   // Telegram Bot
	TypeWeChat    ChannelType = "wechat"     // WeChat Official Account
	TypeTwilioSMS ChannelType = "twilio_sms" // Twilio SMS

	// Web channels
	TypeWebForm ChannelType = "web_form" // Web Form / Webhook
)

type ChannelStatus string

const (
	StatusActive       ChannelStatus = "active"
	StatusInactive     ChannelStatus = "inactive"
	StatusConnecting   ChannelStatus = "connecting"
	StatusDisconnected ChannelStatus = "disconnected"
	StatusError        ChannelStatus = "error"
)

type Channel struct {
	ID         uuid.UUID
	UserID     uuid.UUID
	ProjectID  uuid.UUID
	TenantID   string
	Name       string
	Type       ChannelType
	Status     ChannelStatus
	ExternalID string
	Config     map[string]interface{}

	// Connection mode for WAHA channels (manual or auto)
	ConnectionMode ConnectionMode

	WebhookID           string
	WebhookURL          string
	WebhookConfiguredAt *time.Time
	WebhookActive       bool

	PipelineID                   *uuid.UUID
	DefaultSessionTimeoutMinutes int

	AIEnabled       bool
	AIAgentsEnabled bool
	AllowGroups     bool // Se o canal aceita mensagens de grupos WhatsApp
	TrackingEnabled bool // Se o canal rastreia origem das mensagens (tracking básico)

	// DebounceTimeoutMs define o timeout do debouncer em milissegundos
	// Usado para agrupar mensagens sequenciais (especialmente com mídia)
	// Default: 15000ms (15 segundos)
	// Se 0, usa o default de 15s
	DebounceTimeoutMs int

	MessagesReceived int
	MessagesSent     int
	LastMessageAt    *time.Time
	LastErrorAt      *time.Time
	LastError        string

	CreatedAt time.Time
	UpdatedAt time.Time

	events []DomainEvent
}

type WAHAAuth struct {
	APIKey string `json:"api_key"`
	Token  string `json:"token"`
}

type WAHASessionStatus string

const (
	WAHASessionStatusStarting     WAHASessionStatus = "STARTING"
	WAHASessionStatusScanQR       WAHASessionStatus = "SCAN_QR_CODE"
	WAHASessionStatusWorking      WAHASessionStatus = "WORKING"
	WAHASessionStatusFailed       WAHASessionStatus = "FAILED"
	WAHASessionStatusStopped      WAHASessionStatus = "STOPPED"
	WAHASessionStatusUnauthorized WAHASessionStatus = "UNAUTHORIZED"
)

type WAHASessionEvent struct {
	SessionID string            `json:"session"`
	Status    WAHASessionStatus `json:"status"`
	QRCode    string            `json:"qr,omitempty"`
	Message   string            `json:"message,omitempty"`
	Timestamp int64             `json:"timestamp"`
}

type WAHAImportStrategy string

const (
	WAHAImportNone    WAHAImportStrategy = "none"
	WAHAImportNewOnly WAHAImportStrategy = "new_only"
	WAHAImportAll     WAHAImportStrategy = "all"
)

// AIContentType representa os tipos de conteúdo suportados para processamento de IA
type AIContentType string

const (
	AIContentTypeText     AIContentType = "text"     // Texto - Claude/GPT
	AIContentTypeAudio    AIContentType = "audio"    // Áudio - Whisper/Deepgram
	AIContentTypeImage    AIContentType = "image"    // Imagem - Gemini Vision/GPT-4V
	AIContentTypeVideo    AIContentType = "video"    // Vídeo - Extração + processamento
	AIContentTypeDocument AIContentType = "document" // PDF/Docs - LlamaParse
	AIContentTypeVoice    AIContentType = "voice"    // Áudio de voz (PTT) - processamento prioritário
)

// AIProcessingConfig configura como o canal processa cada tipo de conteúdo
type AIProcessingConfig struct {
	Enabled          bool    `json:"enabled"`           // Se IA está habilitada para este tipo
	Provider         string  `json:"provider"`          // openai, anthropic, google, deepgram, llamaparse
	Model            string  `json:"model"`             // gpt-4, claude-3-opus, gemini-pro, whisper-1
	Priority         int     `json:"priority"`          // Prioridade (1-10, onde 10 é máxima)
	DebounceMs       int     `json:"debounce_ms"`       // Debounce em milissegundos
	MaxSizeBytes     int64   `json:"max_size_bytes"`    // Tamanho máximo do arquivo
	SplitLongAudio   bool    `json:"split_long_audio"`  // Se deve quebrar áudios longos
	SilenceThreshold float64 `json:"silence_threshold"` // Threshold para detecção de silêncio (0-1)
}

type WAHAConfig struct {
	BaseURL         string             `json:"base_url"`
	Auth            WAHAAuth           `json:"auth"`
	SessionID       string             `json:"session_id"`
	WebhookURL      string             `json:"webhook_url"`
	ImportStrategy  WAHAImportStrategy `json:"import_strategy"`
	ImportCompleted bool               `json:"import_completed"`
}

type WhatsAppConfig struct {
	AccessToken   string `json:"access_token"`
	PhoneNumberID string `json:"phone_number_id"`
	BusinessID    string `json:"business_id"`
	WebhookURL    string `json:"webhook_url"`
	VerifyToken   string `json:"verify_token"`
}

type TelegramConfig struct {
	BotToken   string `json:"bot_token"`
	BotID      string `json:"bot_id"`
	WebhookURL string `json:"webhook_url"`
}

func NewChannel(userID, projectID uuid.UUID, tenantID, name string, channelType ChannelType) (*Channel, error) {
	if name == "" {
		return nil, fmt.Errorf("channel name is required")
	}

	if !isValidChannelType(channelType) {
		return nil, fmt.Errorf("invalid channel type: %s", channelType)
	}

	now := time.Now()
	ch := &Channel{
		ID:        uuid.New(),
		UserID:    userID,
		ProjectID: projectID,
		TenantID:  tenantID,
		Name:      name,
		Type:      channelType,
		Status:    StatusInactive,
		WebhookID: uuid.New().String(),
		Config:    make(map[string]interface{}),
		CreatedAt: now,
		UpdatedAt: now,
		events:    []DomainEvent{},
	}

	ch.addEvent(ChannelCreatedEvent{
		ChannelID:  ch.ID,
		ProjectID:  projectID,
		TenantID:   tenantID,
		Name:       name,
		Type:       channelType,
		ExternalID: ch.ExternalID,
		CreatedAt:  now,
	})

	return ch, nil
}

func NewWAHAChannel(userID, projectID uuid.UUID, tenantID, name string, config WAHAConfig) (*Channel, error) {
	channel, err := NewChannel(userID, projectID, tenantID, name, TypeWAHA)
	if err != nil {
		return nil, err
	}

	// Default to manual mode if not specified
	channel.ConnectionMode = ConnectionModeManual
	channel.ExternalID = config.SessionID

	if err := channel.SetWAHAConfig(config); err != nil {
		return nil, err
	}

	return channel, nil
}

// NewWhatsAppBusinessChannel creates a new WhatsApp Business channel (auto mode)
//
// This is the user-friendly way to connect WhatsApp:
// - System creates and manages WAHA session internally
// - Returns QR code for user to scan
// - User doesn't need to know about WAHA
// - Frontend just shows "Connect WhatsApp" -> QR code appears
//
// Example usage:
//
//	channel := NewWhatsAppBusinessChannel(userID, projectID, tenantID, "My WhatsApp")
//	// System will create WAHA session and return QR code
func NewWhatsAppBusinessChannel(userID, projectID uuid.UUID, tenantID, name string) (*Channel, error) {
	channel, err := NewChannel(userID, projectID, tenantID, name, TypeWhatsAppBusiness)
	if err != nil {
		return nil, err
	}

	// Auto mode - system manages WAHA session internally
	channel.ConnectionMode = ConnectionModeAuto
	channel.Status = StatusConnecting
	// ExternalID will be set after WAHA session creation

	return channel, nil
}

func NewWhatsAppChannel(userID, projectID uuid.UUID, tenantID, name string, config WhatsAppConfig) (*Channel, error) {
	channel, err := NewChannel(userID, projectID, tenantID, name, TypeWhatsApp)
	if err != nil {
		return nil, err
	}

	channel.ExternalID = config.PhoneNumberID

	if err := channel.SetWhatsAppConfig(config); err != nil {
		return nil, err
	}

	return channel, nil
}

func NewTelegramChannel(userID, projectID uuid.UUID, tenantID, name string, config TelegramConfig) (*Channel, error) {
	channel, err := NewChannel(userID, projectID, tenantID, name, TypeTelegram)
	if err != nil {
		return nil, err
	}

	channel.ExternalID = config.BotID

	if err := channel.SetTelegramConfig(config); err != nil {
		return nil, err
	}

	return channel, nil
}

func (c *Channel) SetWAHAConfig(config WAHAConfig) error {
	if c.Type != TypeWAHA {
		return fmt.Errorf("channel is not WAHA type")
	}

	if config.BaseURL == "" {
		return fmt.Errorf("WAHA base URL is required")
	}

	if config.Auth.APIKey == "" && config.Auth.Token == "" {
		return fmt.Errorf("WAHA authentication (API key or token) is required")
	}

	if config.ImportStrategy == "" {
		config.ImportStrategy = WAHAImportNone
	}

	c.Config["base_url"] = config.BaseURL
	c.Config["auth"] = map[string]interface{}{
		"api_key": config.Auth.APIKey,
		"token":   config.Auth.Token,
	}
	c.Config["session_id"] = config.SessionID
	c.Config["webhook_url"] = config.WebhookURL
	c.Config["import_strategy"] = string(config.ImportStrategy)
	c.Config["import_completed"] = config.ImportCompleted

	c.ExternalID = config.SessionID
	c.UpdatedAt = time.Now()

	return nil
}

func (c *Channel) SetWhatsAppConfig(config WhatsAppConfig) error {
	if c.Type != TypeWhatsApp {
		return fmt.Errorf("channel is not WhatsApp type")
	}

	if config.AccessToken == "" {
		return fmt.Errorf("WhatsApp access token is required")
	}

	if config.PhoneNumberID == "" {
		return fmt.Errorf("WhatsApp phone number ID is required")
	}

	c.Config["access_token"] = config.AccessToken
	c.Config["phone_number_id"] = config.PhoneNumberID
	c.Config["business_id"] = config.BusinessID
	c.Config["webhook_url"] = config.WebhookURL
	c.Config["verify_token"] = config.VerifyToken

	c.ExternalID = config.PhoneNumberID
	c.UpdatedAt = time.Now()

	return nil
}

func (c *Channel) SetTelegramConfig(config TelegramConfig) error {
	if c.Type != TypeTelegram {
		return fmt.Errorf("channel is not Telegram type")
	}

	if config.BotToken == "" {
		return fmt.Errorf("Telegram bot token is required")
	}

	if config.BotID == "" {
		return fmt.Errorf("Telegram bot ID is required")
	}

	c.Config["bot_token"] = config.BotToken
	c.Config["bot_id"] = config.BotID
	c.Config["webhook_url"] = config.WebhookURL

	c.ExternalID = config.BotID
	c.UpdatedAt = time.Now()

	return nil
}

func (c *Channel) GetWAHAConfig() (*WAHAConfig, error) {
	if c.Type != TypeWAHA {
		return nil, fmt.Errorf("channel is not WAHA type")
	}

	config := &WAHAConfig{
		Auth:           WAHAAuth{},
		ImportStrategy: WAHAImportNone,
	}

	if baseURL, ok := c.Config["base_url"].(string); ok {
		config.BaseURL = baseURL
	}

	if auth, ok := c.Config["auth"].(map[string]interface{}); ok {
		if apiKey, ok := auth["api_key"].(string); ok {
			config.Auth.APIKey = apiKey
		}
		if token, ok := auth["token"].(string); ok {
			config.Auth.Token = token
		}
	}

	if sessionID, ok := c.Config["session_id"].(string); ok {
		config.SessionID = sessionID
	}

	if webhookURL, ok := c.Config["webhook_url"].(string); ok {
		config.WebhookURL = webhookURL
	}

	if importStrategy, ok := c.Config["import_strategy"].(string); ok {
		config.ImportStrategy = WAHAImportStrategy(importStrategy)
	}

	if importCompleted, ok := c.Config["import_completed"].(bool); ok {
		config.ImportCompleted = importCompleted
	}

	return config, nil
}

func (c *Channel) Activate() {
	c.Status = StatusActive
	now := time.Now()
	c.UpdatedAt = now

	c.addEvent(ChannelActivatedEvent{
		ChannelID:   c.ID,
		ActivatedAt: now,
	})
}

func (c *Channel) Deactivate() {
	c.Status = StatusInactive
	now := time.Now()
	c.UpdatedAt = now

	c.addEvent(ChannelDeactivatedEvent{
		ChannelID:     c.ID,
		DeactivatedAt: now,
	})
}

func (c *Channel) SetConnecting() {
	c.Status = StatusConnecting
	c.UpdatedAt = time.Now()
}

func (c *Channel) SetError(errorMsg string) {
	c.Status = StatusError
	c.LastError = errorMsg
	c.LastErrorAt = &time.Time{}
	*c.LastErrorAt = time.Now()
	c.UpdatedAt = time.Now()
}

func (c *Channel) IncrementMessagesReceived() {
	c.MessagesReceived++
	now := time.Now()
	c.LastMessageAt = &now
	c.UpdatedAt = time.Now()
}

func (c *Channel) IncrementMessagesSent() {
	c.MessagesSent++
	c.UpdatedAt = time.Now()
}

func (c *Channel) IsActive() bool {
	return c.Status == StatusActive
}

func (c *Channel) IsWAHA() bool {
	return c.Type == TypeWAHA
}

func (c *Channel) GetWAHASessionStatus() WAHASessionStatus {
	if c.Type != TypeWAHA {
		return ""
	}

	if status, ok := c.Config["session_status"].(string); ok {
		return WAHASessionStatus(status)
	}

	return WAHASessionStatusStarting
}

func (c *Channel) SetWAHASessionStatus(status WAHASessionStatus) {
	if c.Type != TypeWAHA {
		return
	}

	c.Config["session_status"] = string(status)
	c.UpdatedAt = time.Now()

	if status == WAHASessionStatusWorking {
		c.Activate()
	} else if status == WAHASessionStatusFailed || status == WAHASessionStatusStopped {
		c.Deactivate()
	}
}

func (c *Channel) SetWAHAQRCode(qrCode string) {
	if c.Type != TypeWAHA {
		return
	}

	c.Config["qr_code"] = qrCode
	c.Config["qr_generated_at"] = time.Now().Unix()
	c.UpdatedAt = time.Now()
}

func (c *Channel) GetWAHAQRCode() string {
	if c.Type != TypeWAHA {
		return ""
	}

	if qrCode, ok := c.Config["qr_code"].(string); ok {
		return qrCode
	}

	return ""
}

func (c *Channel) IsWAHAQRCodeValid() bool {
	if c.Type != TypeWAHA {
		return false
	}

	if c.GetWAHAQRCode() == "" {
		return false
	}

	if c.GetWAHASessionStatus() == WAHASessionStatusWorking {
		return false
	}

	if generatedAt, ok := c.Config["qr_generated_at"].(int64); ok {
		expirationTime := time.Unix(generatedAt, 0).Add(45 * time.Second)
		return time.Now().Before(expirationTime)
	}

	return false
}

func (c *Channel) ClearWAHAQRCode() {
	if c.Type != TypeWAHA {
		return
	}

	delete(c.Config, "qr_code")
	delete(c.Config, "qr_generated_at")
	c.UpdatedAt = time.Now()
}

func (c *Channel) NeedsNewQRCode() bool {
	if c.Type != TypeWAHA {
		return false
	}

	status := c.GetWAHASessionStatus()

	return status == WAHASessionStatusScanQR && !c.IsWAHAQRCodeValid()
}

func (c *Channel) UpdateWAHAQRCode(qrCode string) {
	if c.Type != TypeWAHA {
		return
	}

	if count, ok := c.Config["qr_count"].(int); ok {
		c.Config["qr_count"] = count + 1
	} else {
		c.Config["qr_count"] = 1
	}

	c.Config["qr_code"] = qrCode
	c.Config["qr_generated_at"] = time.Now().Unix()
	c.Config["qr_last_updated"] = time.Now().Format(time.RFC3339)

	c.SetWAHASessionStatus(WAHASessionStatusScanQR)

	c.UpdatedAt = time.Now()
}

func (c *Channel) LogQRCodeToConsole() {
	if c.Type != TypeWAHA {
		return
	}

	qrCode := c.GetWAHAQRCode()
	if qrCode == "" {
		fmt.Printf("🔍 [WAHA QR] Channel %s (%s): No QR code available\n", c.Name, c.ExternalID)
		return
	}

	count := c.GetWAHAQRCodeCount()
	status := c.GetWAHASessionStatus()

	separator := strings.Repeat("=", 80)
	fmt.Printf("\n%s\n", separator)
	fmt.Printf("📱 [WAHA QR CODE] Channel: %s | Session: %s | Status: %s | QR #%d\n",
		c.Name, c.ExternalID, string(status), count)
	fmt.Printf("🕒 Generated at: %s\n", time.Unix(c.Config["qr_generated_at"].(int64), 0).Format("15:04:05"))
	fmt.Printf("⏰ Expires at: %s\n", time.Unix(c.Config["qr_generated_at"].(int64), 0).Add(45*time.Second).Format("15:04:05"))
	fmt.Printf("📋 QR Code:\n%s\n", qrCode)
	fmt.Printf("%s\n\n", separator)
}

func (c *Channel) GetWAHAQRCodeCount() int {
	if c.Type != TypeWAHA {
		return 0
	}

	if count, ok := c.Config["qr_count"].(int); ok {
		return count
	}

	return 0
}

func (c *Channel) SetWAHAImportCompleted() {
	if c.Type != TypeWAHA {
		return
	}

	c.Config["import_completed"] = true
	c.Config["import_completed_at"] = time.Now().Format(time.RFC3339)
	c.UpdatedAt = time.Now()
}

func (c *Channel) IsWAHAImportCompleted() bool {
	if c.Type != TypeWAHA {
		return false
	}

	if completed, ok := c.Config["import_completed"].(bool); ok {
		return completed
	}

	return false
}

func (c *Channel) GetWAHAImportStrategy() WAHAImportStrategy {
	if c.Type != TypeWAHA {
		return WAHAImportNone
	}

	if strategy, ok := c.Config["import_strategy"].(string); ok {
		return WAHAImportStrategy(strategy)
	}

	return WAHAImportNone
}

func (c *Channel) NeedsHistoryImport() bool {
	if c.Type != TypeWAHA {
		return false
	}

	strategy := c.GetWAHAImportStrategy()
	return !c.IsWAHAImportCompleted() && strategy != WAHAImportNone
}

func isValidChannelType(channelType ChannelType) bool {
	validTypes := []ChannelType{
		TypeWAHA, TypeWhatsAppBusiness, TypeWhatsApp, TypeTelegram,
		TypeMessenger, TypeInstagram, TypeWeChat, TypeTwilioSMS, TypeWebForm,
	}

	for _, validType := range validTypes {
		if channelType == validType {
			return true
		}
	}

	return false
}

// IsWAHABased returns true if channel uses WAHA infrastructure
// (either manual WAHA or WhatsApp Business which uses WAHA internally)
func (c *Channel) IsWAHABased() bool {
	return c.Type == TypeWAHA || c.Type == TypeWhatsAppBusiness
}

func (c *Channel) DomainEvents() []DomainEvent {
	return c.events
}

func (c *Channel) ClearEvents() {
	c.events = []DomainEvent{}
}

func (c *Channel) addEvent(event DomainEvent) {
	c.events = append(c.events, event)
}

func (c *Channel) ShouldProcessAI() bool {
	return c.AIEnabled
}

func (c *Channel) AssociatePipeline(pipelineID uuid.UUID) error {
	if pipelineID == uuid.Nil {
		return fmt.Errorf("pipeline ID cannot be nil")
	}

	c.PipelineID = &pipelineID
	c.UpdatedAt = time.Now()

	c.addEvent(ChannelPipelineAssociatedEvent{
		ChannelID:    c.ID,
		PipelineID:   pipelineID,
		AssociatedAt: time.Now(),
	})

	return nil
}

func (c *Channel) DisassociatePipeline() {
	if c.PipelineID == nil {
		return
	}

	oldPipelineID := *c.PipelineID
	c.PipelineID = nil
	c.UpdatedAt = time.Now()

	c.addEvent(ChannelPipelineDisassociatedEvent{
		ChannelID:       c.ID,
		PipelineID:      oldPipelineID,
		DisassociatedAt: time.Now(),
	})
}

func (c *Channel) HasPipeline() bool {
	return c.PipelineID != nil && *c.PipelineID != uuid.Nil
}

func (c *Channel) SetDefaultTimeout(minutes int) error {
	if minutes <= 0 {
		return fmt.Errorf("timeout must be greater than 0")
	}
	if minutes > 1440 {
		return fmt.Errorf("timeout cannot exceed 1440 minutes (24 hours)")
	}

	c.DefaultSessionTimeoutMinutes = minutes
	c.UpdatedAt = time.Now()

	return nil
}

// EnableGroups habilita o processamento de mensagens de grupos
func (c *Channel) EnableGroups() {
	c.AllowGroups = true
	c.UpdatedAt = time.Now()
}

// DisableGroups desabilita o processamento de mensagens de grupos
func (c *Channel) DisableGroups() {
	c.AllowGroups = false
	c.UpdatedAt = time.Now()
}

// ShouldProcessGroups retorna se o canal deve processar mensagens de grupos
func (c *Channel) ShouldProcessGroups() bool {
	return c.AllowGroups
}

// EnableTracking habilita o rastreamento de origem das mensagens
func (c *Channel) EnableTracking() {
	c.TrackingEnabled = true
	c.UpdatedAt = time.Now()
}

// DisableTracking desabilita o rastreamento de origem das mensagens
func (c *Channel) DisableTracking() {
	c.TrackingEnabled = false
	c.UpdatedAt = time.Now()
}

// ShouldTrackMessages retorna se o canal deve rastrear origem das mensagens
func (c *Channel) ShouldTrackMessages() bool {
	return c.TrackingEnabled
}

// SetDebounceTimeout define o timeout do debouncer em milissegundos
func (c *Channel) SetDebounceTimeout(timeoutMs int) error {
	if timeoutMs < 0 {
		return fmt.Errorf("debounce timeout cannot be negative")
	}
	if timeoutMs > 300000 { // Max 5 minutes
		return fmt.Errorf("debounce timeout cannot exceed 300000ms (5 minutes)")
	}

	c.DebounceTimeoutMs = timeoutMs
	c.UpdatedAt = time.Now()
	return nil
}

// GetDebounceTimeout retorna o timeout do debouncer em milissegundos
// Se não configurado (0), retorna o default de 15 segundos
func (c *Channel) GetDebounceTimeout() int {
	if c.DebounceTimeoutMs <= 0 {
		return 15000 // Default: 15 segundos
	}
	return c.DebounceTimeoutMs
}

// GetDebounceDuration retorna o timeout do debouncer como time.Duration
func (c *Channel) GetDebounceDuration() time.Duration {
	return time.Duration(c.GetDebounceTimeout()) * time.Millisecond
}

// SetAIProcessingConfig configura o processamento de IA para um tipo de conteúdo
func (c *Channel) SetAIProcessingConfig(contentType AIContentType, config AIProcessingConfig) {
	if c.Config == nil {
		c.Config = make(map[string]interface{})
	}

	if c.Config["ai_processing"] == nil {
		c.Config["ai_processing"] = make(map[string]interface{})
	}

	aiProcessing := c.Config["ai_processing"].(map[string]interface{})
	aiProcessing[string(contentType)] = config
	c.UpdatedAt = time.Now()
}

// GetAIProcessingConfig retorna a configuração de IA para um tipo de conteúdo
func (c *Channel) GetAIProcessingConfig(contentType AIContentType) *AIProcessingConfig {
	if c.Config == nil {
		return nil
	}

	aiProcessing, ok := c.Config["ai_processing"].(map[string]interface{})
	if !ok {
		return nil
	}

	config, ok := aiProcessing[string(contentType)]
	if !ok {
		return nil
	}

	// Convert map to struct (simplified)
	if configMap, ok := config.(map[string]interface{}); ok {
		return &AIProcessingConfig{
			Enabled:          getBool(configMap, "enabled"),
			Provider:         getString(configMap, "provider"),
			Model:            getString(configMap, "model"),
			Priority:         getInt(configMap, "priority"),
			DebounceMs:       getInt(configMap, "debounce_ms"),
			MaxSizeBytes:     getInt64(configMap, "max_size_bytes"),
			SplitLongAudio:   getBool(configMap, "split_long_audio"),
			SilenceThreshold: getFloat64(configMap, "silence_threshold"),
		}
	}

	return nil
}

// ShouldProcessAIContent verifica se deve processar IA para um tipo de conteúdo
func (c *Channel) ShouldProcessAIContent(contentType AIContentType) bool {
	if !c.AIEnabled {
		return false
	}

	config := c.GetAIProcessingConfig(contentType)
	if config == nil {
		return false
	}

	return config.Enabled
}

// GetDefaultAIConfig retorna configurações padrão para cada tipo de conteúdo
func GetDefaultAIConfig(contentType AIContentType) AIProcessingConfig {
	configs := map[AIContentType]AIProcessingConfig{
		AIContentTypeText: {
			Enabled:      true,
			Provider:     "anthropic",
			Model:        "claude-3-5-sonnet-20241022",
			Priority:     5,
			DebounceMs:   1000,
			MaxSizeBytes: 1024 * 1024, // 1MB
		},
		AIContentTypeAudio: {
			Enabled:          true,
			Provider:         "openai",
			Model:            "whisper-1",
			Priority:         8,
			DebounceMs:       500,
			MaxSizeBytes:     25 * 1024 * 1024, // 25MB
			SplitLongAudio:   true,
			SilenceThreshold: 0.3,
		},
		AIContentTypeVoice: {
			Enabled:          true,
			Provider:         "openai",
			Model:            "whisper-1",
			Priority:         10, // Máxima prioridade para PTT
			DebounceMs:       100,
			MaxSizeBytes:     25 * 1024 * 1024,
			SplitLongAudio:   false, // PTT geralmente são curtos
			SilenceThreshold: 0.3,
		},
		AIContentTypeImage: {
			Enabled:      true,
			Provider:     "google",
			Model:        "gemini-1.5-pro",
			Priority:     7,
			DebounceMs:   1000,
			MaxSizeBytes: 10 * 1024 * 1024, // 10MB
		},
		AIContentTypeVideo: {
			Enabled:      false, // Desabilitado por padrão (processamento pesado)
			Provider:     "openai",
			Model:        "gpt-4-vision",
			Priority:     3,
			DebounceMs:   5000,
			MaxSizeBytes: 100 * 1024 * 1024, // 100MB
		},
		AIContentTypeDocument: {
			Enabled:      true,
			Provider:     "llamaparse",
			Model:        "default",
			Priority:     6,
			DebounceMs:   2000,
			MaxSizeBytes: 50 * 1024 * 1024, // 50MB
		},
	}

	if config, ok := configs[contentType]; ok {
		return config
	}

	// Fallback default
	return AIProcessingConfig{
		Enabled:      false,
		Priority:     1,
		DebounceMs:   1000,
		MaxSizeBytes: 10 * 1024 * 1024,
	}
}

// Helper functions for type conversion
func getBool(m map[string]interface{}, key string) bool {
	if v, ok := m[key]; ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return false
}

func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func getInt(m map[string]interface{}, key string) int {
	if v, ok := m[key]; ok {
		if i, ok := v.(int); ok {
			return i
		}
		if f, ok := v.(float64); ok {
			return int(f)
		}
	}
	return 0
}

func getInt64(m map[string]interface{}, key string) int64 {
	if v, ok := m[key]; ok {
		if i, ok := v.(int64); ok {
			return i
		}
		if f, ok := v.(float64); ok {
			return int64(f)
		}
	}
	return 0
}

func getFloat64(m map[string]interface{}, key string) float64 {
	if v, ok := m[key]; ok {
		if f, ok := v.(float64); ok {
			return f
		}
		if i, ok := v.(int); ok {
			return float64(i)
		}
	}
	return 0.0
}

// Label Management Methods

// GetLabels returns all labels configured for this channel
func (c *Channel) GetLabels() *LabelCollection {
	if c.Config == nil {
		return NewLabelCollection()
	}

	labelsData, ok := c.Config["labels"]
	if !ok {
		return NewLabelCollection()
	}

	// Handle different serialization formats
	switch v := labelsData.(type) {
	case []interface{}:
		// Slice format from JSON
		labels := make([]*Label, 0, len(v))
		for _, item := range v {
			if labelMap, ok := item.(map[string]interface{}); ok {
				label := &Label{
					ID:       getString(labelMap, "id"),
					Name:     getString(labelMap, "name"),
					Color:    getInt(labelMap, "color"),
					ColorHex: getString(labelMap, "colorHex"),
				}
				labels = append(labels, label)
			}
		}
		return ReconstructLabelCollection(labels)
	case []*Label:
		// Direct slice of labels
		return ReconstructLabelCollection(v)
	default:
		return NewLabelCollection()
	}
}

// SetLabels sets the labels for this channel
func (c *Channel) SetLabels(labels *LabelCollection) {
	if c.Config == nil {
		c.Config = make(map[string]interface{})
	}

	c.Config["labels"] = labels.ToSlice()
	c.UpdatedAt = time.Now()
}

// AddLabel adds or updates a label in the channel
func (c *Channel) AddLabel(label *Label) error {
	if !c.IsWAHABased() {
		return fmt.Errorf("labels are only supported for WAHA-based channels")
	}

	labels := c.GetLabels()
	labels.Add(label)
	c.SetLabels(labels)

	c.addEvent(ChannelLabelUpsertedEvent{
		ChannelID: c.ID,
		LabelID:   label.ID,
		LabelName: label.Name,
		Timestamp: time.Now(),
	})

	return nil
}

// RemoveLabel removes a label from the channel
func (c *Channel) RemoveLabel(labelID string) error {
	if !c.IsWAHABased() {
		return fmt.Errorf("labels are only supported for WAHA-based channels")
	}

	labels := c.GetLabels()
	if !labels.Has(labelID) {
		return fmt.Errorf("label not found: %s", labelID)
	}

	labels.Remove(labelID)
	c.SetLabels(labels)

	c.addEvent(ChannelLabelDeletedEvent{
		ChannelID: c.ID,
		LabelID:   labelID,
		Timestamp: time.Now(),
	})

	return nil
}

// GetLabel retrieves a specific label by ID
func (c *Channel) GetLabel(labelID string) (*Label, error) {
	labels := c.GetLabels()
	label, exists := labels.Get(labelID)
	if !exists {
		return nil, fmt.Errorf("label not found: %s", labelID)
	}
	return label, nil
}

// HasLabel checks if a label exists in the channel
func (c *Channel) HasLabel(labelID string) bool {
	labels := c.GetLabels()
	return labels.Has(labelID)
}

// GetLabelCount returns the number of labels in the channel
func (c *Channel) GetLabelCount() int {
	labels := c.GetLabels()
	return labels.Count()
}

type Repository interface {
	Create(channel *Channel) error
	GetByID(id uuid.UUID) (*Channel, error)
	GetByUserID(userID uuid.UUID) ([]*Channel, error)
	GetByProjectID(projectID uuid.UUID) ([]*Channel, error)
	GetByExternalID(externalID string) (*Channel, error)
	GetByWebhookID(webhookID string) (*Channel, error)
	Update(channel *Channel) error
	Delete(id uuid.UUID) error
	GetActiveWAHAChannels() ([]*Channel, error)
}
