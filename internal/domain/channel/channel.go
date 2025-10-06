package channel

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// ChannelType representa os tipos de canal
type ChannelType string

const (
	TypeWAHA      ChannelType = "waha"
	TypeWhatsApp  ChannelType = "whatsapp"
	TypeTelegram  ChannelType = "telegram"
	TypeMessenger ChannelType = "messenger"
	TypeInstagram ChannelType = "instagram"
)

// ChannelStatus representa o status do canal
type ChannelStatus string

const (
	StatusActive      ChannelStatus = "active"
	StatusInactive    ChannelStatus = "inactive"
	StatusConnecting  ChannelStatus = "connecting"
	StatusDisconnected ChannelStatus = "disconnected"
	StatusError       ChannelStatus = "error"
)

// Channel representa um canal de comunicação
type Channel struct {
	ID         uuid.UUID
	UserID     uuid.UUID
	ProjectID  uuid.UUID
	TenantID   string
	Name       string
	Type       ChannelType
	Status     ChannelStatus
	ExternalID string // ID externo do canal (ex: session_id do WAHA, phone_number_id do WhatsApp)
	Config     map[string]interface{}
	
	// Estatísticas
	MessagesReceived int
	MessagesSent     int
	LastMessageAt    *time.Time
	LastErrorAt      *time.Time
	LastError        string
	
	CreatedAt time.Time
	UpdatedAt time.Time
}

// WAHAAuth representa a autenticação do WAHA
type WAHAAuth struct {
	APIKey string `json:"api_key"` // Chave da API para autenticação (obtida de variável de ambiente)
	Token  string `json:"token"`   // Token de acesso (se diferente da API key)
}

// WAHASessionStatus representa o status da sessão WAHA
type WAHASessionStatus string

const (
	WAHASessionStatusStarting     WAHASessionStatus = "STARTING"     // Iniciando sessão
	WAHASessionStatusScanQR       WAHASessionStatus = "SCAN_QR_CODE" // Aguardando scan do QR code
	WAHASessionStatusWorking      WAHASessionStatus = "WORKING"      // Sessão ativa e funcionando
	WAHASessionStatusFailed       WAHASessionStatus = "FAILED"       // Falha na sessão
	WAHASessionStatusStopped      WAHASessionStatus = "STOPPED"      // Sessão parada
	WAHASessionStatusUnauthorized WAHASessionStatus = "UNAUTHORIZED" // Não autorizado
)

// WAHASessionEvent representa um evento de status da sessão WAHA
type WAHASessionEvent struct {
	SessionID string            `json:"session"`
	Status    WAHASessionStatus `json:"status"`
	QRCode    string            `json:"qr,omitempty"`    // QR code quando status é SCAN_QR_CODE
	Message   string            `json:"message,omitempty"` // Mensagem adicional
	Timestamp int64             `json:"timestamp"`
}

// WAHAConfig representa a configuração específica do WAHA
type WAHAConfig struct {
	BaseURL    string   `json:"base_url"`
	Auth       WAHAAuth `json:"auth"`
	SessionID  string   `json:"session_id"`  // Equivale ao ExternalID do canal
	WebhookURL string   `json:"webhook_url"`
}

// WhatsAppConfig representa a configuração do WhatsApp Business API
type WhatsAppConfig struct {
	AccessToken     string `json:"access_token"`
	PhoneNumberID   string `json:"phone_number_id"` // Equivale ao ExternalID do canal
	BusinessID      string `json:"business_id"`
	WebhookURL      string `json:"webhook_url"`
	VerifyToken     string `json:"verify_token"`
}

// TelegramConfig representa a configuração do Telegram Bot
type TelegramConfig struct {
	BotToken   string `json:"bot_token"`   // Equivale à autenticação
	BotID      string `json:"bot_id"`      // Equivale ao ExternalID do canal
	WebhookURL string `json:"webhook_url"`
}

// NewChannel cria um novo canal
func NewChannel(userID, projectID uuid.UUID, tenantID, name string, channelType ChannelType) (*Channel, error) {
	if name == "" {
		return nil, fmt.Errorf("channel name is required")
	}
	
	if !isValidChannelType(channelType) {
		return nil, fmt.Errorf("invalid channel type: %s", channelType)
	}
	
	now := time.Now()
	return &Channel{
		ID:        uuid.New(),
		UserID:    userID,
		ProjectID: projectID,
		TenantID:  tenantID,
		Name:      name,
		Type:      channelType,
		Status:    StatusInactive,
		Config:    make(map[string]interface{}),
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// NewWAHAChannel cria um novo canal WAHA
func NewWAHAChannel(userID, projectID uuid.UUID, tenantID, name string, config WAHAConfig) (*Channel, error) {
	channel, err := NewChannel(userID, projectID, tenantID, name, TypeWAHA)
	if err != nil {
		return nil, err
	}
	
	// Define o ExternalID como o SessionID do WAHA
	channel.ExternalID = config.SessionID
	
	if err := channel.SetWAHAConfig(config); err != nil {
		return nil, err
	}
	
	return channel, nil
}

// NewWhatsAppChannel cria um novo canal WhatsApp Business API
func NewWhatsAppChannel(userID, projectID uuid.UUID, tenantID, name string, config WhatsAppConfig) (*Channel, error) {
	channel, err := NewChannel(userID, projectID, tenantID, name, TypeWhatsApp)
	if err != nil {
		return nil, err
	}
	
	// Define o ExternalID como o PhoneNumberID do WhatsApp
	channel.ExternalID = config.PhoneNumberID
	
	if err := channel.SetWhatsAppConfig(config); err != nil {
		return nil, err
	}
	
	return channel, nil
}

// NewTelegramChannel cria um novo canal Telegram Bot
func NewTelegramChannel(userID, projectID uuid.UUID, tenantID, name string, config TelegramConfig) (*Channel, error) {
	channel, err := NewChannel(userID, projectID, tenantID, name, TypeTelegram)
	if err != nil {
		return nil, err
	}
	
	// Define o ExternalID como o BotID do Telegram
	channel.ExternalID = config.BotID
	
	if err := channel.SetTelegramConfig(config); err != nil {
		return nil, err
	}
	
	return channel, nil
}

// SetWAHAConfig configura o canal para WAHA
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
	
	c.Config["base_url"] = config.BaseURL
	c.Config["auth_api_key"] = config.Auth.APIKey
	c.Config["auth_token"] = config.Auth.Token
	c.Config["session_id"] = config.SessionID
	c.Config["webhook_url"] = config.WebhookURL
	
	// Atualiza o ExternalID com o SessionID
	c.ExternalID = config.SessionID
	c.UpdatedAt = time.Now()
	
	return nil
}

// SetWhatsAppConfig configura o canal para WhatsApp Business API
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
	
	// Atualiza o ExternalID com o PhoneNumberID
	c.ExternalID = config.PhoneNumberID
	c.UpdatedAt = time.Now()
	
	return nil
}

// SetTelegramConfig configura o canal para Telegram Bot
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
	
	// Atualiza o ExternalID com o BotID
	c.ExternalID = config.BotID
	c.UpdatedAt = time.Now()
	
	return nil
}

// GetWAHAConfig retorna a configuração WAHA
func (c *Channel) GetWAHAConfig() (*WAHAConfig, error) {
	if c.Type != TypeWAHA {
		return nil, fmt.Errorf("channel is not WAHA type")
	}
	
	config := &WAHAConfig{
		Auth: WAHAAuth{},
	}
	
	if baseURL, ok := c.Config["base_url"].(string); ok {
		config.BaseURL = baseURL
	}
	
	if apiKey, ok := c.Config["auth_api_key"].(string); ok {
		config.Auth.APIKey = apiKey
	}
	
	if token, ok := c.Config["auth_token"].(string); ok {
		config.Auth.Token = token
	}
	
	if sessionID, ok := c.Config["session_id"].(string); ok {
		config.SessionID = sessionID
	}
	
	if webhookURL, ok := c.Config["webhook_url"].(string); ok {
		config.WebhookURL = webhookURL
	}
	
	return config, nil
}

// Activate ativa o canal
func (c *Channel) Activate() {
	c.Status = StatusActive
	c.UpdatedAt = time.Now()
}

// Deactivate desativa o canal
func (c *Channel) Deactivate() {
	c.Status = StatusInactive
	c.UpdatedAt = time.Now()
}

// SetConnecting define o status como conectando
func (c *Channel) SetConnecting() {
	c.Status = StatusConnecting
	c.UpdatedAt = time.Now()
}

// SetError define o status como erro
func (c *Channel) SetError(errorMsg string) {
	c.Status = StatusError
	c.LastError = errorMsg
	c.LastErrorAt = &time.Time{}
	*c.LastErrorAt = time.Now()
	c.UpdatedAt = time.Now()
}

// IncrementMessagesReceived incrementa contador de mensagens recebidas
func (c *Channel) IncrementMessagesReceived() {
	c.MessagesReceived++
	now := time.Now()
	c.LastMessageAt = &now
	c.UpdatedAt = time.Now()
}

// IncrementMessagesSent incrementa contador de mensagens enviadas
func (c *Channel) IncrementMessagesSent() {
	c.MessagesSent++
	c.UpdatedAt = time.Now()
}

// IsActive verifica se o canal está ativo
func (c *Channel) IsActive() bool {
	return c.Status == StatusActive
}

// IsWAHA verifica se é canal WAHA
func (c *Channel) IsWAHA() bool {
	return c.Type == TypeWAHA
}

// GetWAHASessionStatus retorna o status atual da sessão WAHA
func (c *Channel) GetWAHASessionStatus() WAHASessionStatus {
	if c.Type != TypeWAHA {
		return ""
	}
	
	if status, ok := c.Config["session_status"].(string); ok {
		return WAHASessionStatus(status)
	}
	
	return WAHASessionStatusStarting
}

// SetWAHASessionStatus atualiza o status da sessão WAHA
func (c *Channel) SetWAHASessionStatus(status WAHASessionStatus) {
	if c.Type != TypeWAHA {
		return
	}
	
	c.Config["session_status"] = string(status)
	c.UpdatedAt = time.Now()
	
	// Se o status for WORKING, ativa o canal
	if status == WAHASessionStatusWorking {
		c.Activate()
	} else if status == WAHASessionStatusFailed || status == WAHASessionStatusStopped {
		c.Deactivate()
	}
}

// SetWAHAQRCode armazena o QR code da sessão WAHA
func (c *Channel) SetWAHAQRCode(qrCode string) {
	if c.Type != TypeWAHA {
		return
	}
	
	c.Config["qr_code"] = qrCode
	c.Config["qr_generated_at"] = time.Now().Unix()
	c.UpdatedAt = time.Now()
}

// GetWAHAQRCode retorna o QR code da sessão WAHA
func (c *Channel) GetWAHAQRCode() string {
	if c.Type != TypeWAHA {
		return ""
	}
	
	if qrCode, ok := c.Config["qr_code"].(string); ok {
		return qrCode
	}
	
	return ""
}

// IsWAHAQRCodeValid verifica se o QR code ainda é válido (não expirou)
func (c *Channel) IsWAHAQRCodeValid() bool {
	if c.Type != TypeWAHA {
		return false
	}
	
	// Se não tem QR code, não é válido
	if c.GetWAHAQRCode() == "" {
		return false
	}
	
	// Se já está conectado (WORKING), QR code não é mais necessário
	if c.GetWAHASessionStatus() == WAHASessionStatusWorking {
		return false
	}
	
	// QR codes da WAHA expiram rapidamente (30-60 segundos típico do WhatsApp)
	if generatedAt, ok := c.Config["qr_generated_at"].(int64); ok {
		expirationTime := time.Unix(generatedAt, 0).Add(45 * time.Second) // 45s é mais realista
		return time.Now().Before(expirationTime)
	}
	
	return false
}

// ClearWAHAQRCode limpa o QR code (quando expira ou sessão conecta)
func (c *Channel) ClearWAHAQRCode() {
	if c.Type != TypeWAHA {
		return
	}
	
	delete(c.Config, "qr_code")
	delete(c.Config, "qr_generated_at")
	c.UpdatedAt = time.Now()
}

// NeedsNewQRCode verifica se precisa gerar um novo QR code
func (c *Channel) NeedsNewQRCode() bool {
	if c.Type != TypeWAHA {
		return false
	}
	
	status := c.GetWAHASessionStatus()
	
	// Só precisa de QR code se está aguardando scan e não tem QR válido
	return status == WAHASessionStatusScanQR && !c.IsWAHAQRCodeValid()
}

// UpdateWAHAQRCode atualiza o QR code quando a WAHA envia um novo
// Este método deve ser chamado toda vez que receber um evento session.status com SCAN_QR_CODE
func (c *Channel) UpdateWAHAQRCode(qrCode string) {
	if c.Type != TypeWAHA {
		return
	}
	
	// Incrementa contador de QR codes gerados para tracking
	if count, ok := c.Config["qr_count"].(int); ok {
		c.Config["qr_count"] = count + 1
	} else {
		c.Config["qr_count"] = 1
	}
	
	// Atualiza o QR code e timestamp
	c.Config["qr_code"] = qrCode
	c.Config["qr_generated_at"] = time.Now().Unix()
	c.Config["qr_last_updated"] = time.Now().Format(time.RFC3339)
	
	// Define status como SCAN_QR_CODE
	c.SetWAHASessionStatus(WAHASessionStatusScanQR)
	
	c.UpdatedAt = time.Now()
}

// LogQRCodeToConsole imprime o QR code no console para debug/teste
func (c *Channel) LogQRCodeToConsole() {
	if c.Type != TypeWAHA {
		return
	}
	
	qrCode := c.GetWAHAQRCode()
	if qrCode == "" {
		fmt.Printf("🔍 [WAHA QR] Canal %s (%s): Nenhum QR code disponível\n", c.Name, c.ExternalID)
		return
	}
	
	count := c.GetWAHAQRCodeCount()
	status := c.GetWAHASessionStatus()
	
	fmt.Printf("\n" + strings.Repeat("=", 80) + "\n")
	fmt.Printf("📱 [WAHA QR CODE] Canal: %s | Session: %s | Status: %s | QR #%d\n", 
		c.Name, c.ExternalID, string(status), count)
	fmt.Printf("🕒 Gerado em: %s\n", time.Unix(c.Config["qr_generated_at"].(int64), 0).Format("15:04:05"))
	fmt.Printf("⏰ Expira em: %s\n", time.Unix(c.Config["qr_generated_at"].(int64), 0).Add(45*time.Second).Format("15:04:05"))
	fmt.Printf("📋 QR Code:\n%s\n", qrCode)
	fmt.Printf(strings.Repeat("=", 80) + "\n\n")
}

// GetWAHAQRCodeCount retorna quantos QR codes já foram gerados para esta sessão
func (c *Channel) GetWAHAQRCodeCount() int {
	if c.Type != TypeWAHA {
		return 0
	}
	
	if count, ok := c.Config["qr_count"].(int); ok {
		return count
	}
	
	return 0
}

// isValidChannelType valida o tipo de canal
func isValidChannelType(channelType ChannelType) bool {
	validTypes := []ChannelType{
		TypeWAHA, TypeWhatsApp, TypeTelegram, TypeMessenger, TypeInstagram,
	}
	
	for _, validType := range validTypes {
		if channelType == validType {
			return true
		}
	}
	
	return false
}

// Repository interface para persistência de canais
type Repository interface {
	Create(channel *Channel) error
	GetByID(id uuid.UUID) (*Channel, error)
	GetByUserID(userID uuid.UUID) ([]*Channel, error)
	GetByProjectID(projectID uuid.UUID) ([]*Channel, error)
	GetByExternalID(externalID string) (*Channel, error) // Buscar por ExternalID (ex: session_id do WAHA)
	Update(channel *Channel) error
	Delete(id uuid.UUID) error
	GetActiveWAHAChannels() ([]*Channel, error)
}
