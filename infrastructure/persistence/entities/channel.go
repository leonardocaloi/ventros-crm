package entities

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// ChannelType representa os tipos de canal disponíveis
type ChannelType string

const (
	ChannelTypeWAHA      ChannelType = "waha"
	ChannelTypeWhatsApp  ChannelType = "whatsapp"
	ChannelTypeTelegram  ChannelType = "telegram"
	ChannelTypeMessenger ChannelType = "messenger"
	ChannelTypeInstagram ChannelType = "instagram"
)

// ChannelStatus representa o status do canal
type ChannelStatus string

const (
	ChannelStatusActive      ChannelStatus = "active"
	ChannelStatusInactive    ChannelStatus = "inactive"
	ChannelStatusConnecting  ChannelStatus = "connecting"
	ChannelStatusDisconnected ChannelStatus = "disconnected"
	ChannelStatusError       ChannelStatus = "error"
)

// ChannelEntity representa um canal de comunicação no banco de dados
type ChannelEntity struct {
	ID          uuid.UUID              `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID      uuid.UUID              `gorm:"type:uuid;not null;index"`
	ProjectID   uuid.UUID              `gorm:"type:uuid;not null;index"`
	TenantID    string                 `gorm:"not null;index"`
	Name        string                 `gorm:"not null"`
	Type        ChannelType            `gorm:"not null;index"`
	Status      ChannelStatus          `gorm:"default:'inactive';index"`
	
	// ExternalID: ID do canal na plataforma externa
	// - WAHA: session_id (ex: "ask-dermato-imersao")
	// - Telegram: bot_id
	// - Messenger: page_id
	// - WhatsApp Business: phone_number_id
	ExternalID string `gorm:"index"` // Indexed para buscas rápidas por session
	
	// Config: Configurações específicas de cada tipo de canal em JSONB
	// Exemplo WAHA: {"base_url": "...", "token": "...", "webhook_url": "..."}
	// Exemplo Telegram: {"bot_token": "...", "webhook_url": "..."}
	Config datatypes.JSON `gorm:"type:jsonb"`
	
	// Estatísticas genéricas (aplicam a todos os tipos)
	MessagesReceived int `gorm:"default:0"`
	MessagesSent     int `gorm:"default:0"`
	LastMessageAt    *time.Time
	LastErrorAt      *time.Time
	LastError        string
	
	// Timestamps
	CreatedAt time.Time      `gorm:"autoCreateTime"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `gorm:"index"`

	// Relacionamentos
	User    UserEntity    `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	Project ProjectEntity `gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE"`
}

func (ChannelEntity) TableName() string {
	return "channels"
}

// IsWAHA verifica se o canal é do tipo WAHA
func (c *ChannelEntity) IsWAHA() bool {
	return c.Type == ChannelTypeWAHA
}

// IsActive verifica se o canal está ativo
func (c *ChannelEntity) IsActive() bool {
	return c.Status == ChannelStatusActive
}

// GetWAHAConfig retorna a configuração WAHA do canal (do Config JSONB)
func (c *ChannelEntity) GetWAHAConfig() map[string]string {
	if !c.IsWAHA() {
		return nil
	}
	
	config := make(map[string]string)
	var jsonConfig map[string]interface{}
	
	if len(c.Config) > 0 {
		if err := json.Unmarshal(c.Config, &jsonConfig); err == nil {
			// Extrai campos do JSONB
			if baseURL, ok := jsonConfig["base_url"].(string); ok {
				config["base_url"] = baseURL
			}
			if token, ok := jsonConfig["token"].(string); ok {
				config["token"] = token
			}
			if sessionID, ok := jsonConfig["session_id"].(string); ok {
				config["session_id"] = sessionID
			}
			if webhookURL, ok := jsonConfig["webhook_url"].(string); ok {
				config["webhook_url"] = webhookURL
			}
		}
	}
	
	return config
}
