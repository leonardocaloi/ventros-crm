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
	ChannelStatusActive       ChannelStatus = "active"
	ChannelStatusInactive     ChannelStatus = "inactive"
	ChannelStatusConnecting   ChannelStatus = "connecting"
	ChannelStatusDisconnected ChannelStatus = "disconnected"
	ChannelStatusError        ChannelStatus = "error"
)

// ChannelEntity representa um canal de comunicação no banco de dados
type ChannelEntity struct {
	ID        uuid.UUID     `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID    uuid.UUID     `gorm:"type:uuid;not null;index:idx_channels_user"`
	ProjectID uuid.UUID     `gorm:"type:uuid;not null;index:idx_channels_project;index:idx_channels_project_type,priority:1"`
	TenantID  string        `gorm:"not null;index:idx_channels_tenant;index:idx_channels_tenant_type,priority:1;index:idx_channels_tenant_status,priority:1"`
	Name      string        `gorm:"not null;index:idx_channels_name"`
	Type      ChannelType   `gorm:"not null;index:idx_channels_type;index:idx_channels_tenant_type,priority:2;index:idx_channels_project_type,priority:2"`
	Status    ChannelStatus `gorm:"default:'inactive';index:idx_channels_status;index:idx_channels_tenant_status,priority:2"`

	// ExternalID: ID do canal na plataforma externa
	// - WAHA: session_id (ex: "ask-dermato-imersao")
	// - Telegram: bot_id
	// - Messenger: page_id
	// - WhatsApp Business: phone_number_id
	ExternalID string `gorm:"index:idx_channels_external_id"` // Indexed para buscas rápidas por session

	// Config: Configurações específicas de cada tipo de canal em JSONB
	// Exemplo WAHA: {"base_url": "...", "token": "...", "webhook_url": "..."}
	// Exemplo Telegram: {"bot_token": "...", "webhook_url": "..."}
	// Exemplo AI: {"ai_process_image": true, "ai_process_video": false, ...}
	Config datatypes.JSON `gorm:"type:jsonb;index:idx_channels_config,type:gin"`

	WebhookID           string     `gorm:"uniqueIndex:idx_channels_webhook_id_unique"` // UUID único para webhook inbound (padrão indústria)
	WebhookURL          string     `gorm:"index:idx_channels_webhook_url"`       // URL do webhook configurada
	WebhookConfiguredAt *time.Time `gorm:"index:idx_channels_webhook_configured"` // Quando o webhook foi configurado
	WebhookActive       bool       `gorm:"default:false;index:idx_channels_webhook_active"` // Se o webhook está ativo

	// Pipeline Association
	PipelineID *uuid.UUID `gorm:"type:uuid;index:idx_channels_pipeline"` // Pipeline associado (opcional)

	// Session Timeout Override (NULL = inherit from project)
	SessionTimeoutMinutes *int `gorm:"index:idx_channels_timeout"` // Override do timeout do projeto (NULL = herda do projeto)

	// AI Features (mantido para compatibilidade, mas processamento vai para message_enriched)
	AIEnabled       bool `gorm:"default:false;index:idx_channels_ai_enabled"` // Canal Inteligente - habilita processamento
	AIAgentsEnabled bool `gorm:"default:false;index:idx_channels_ai_agents"` // Agentes IA - permite respostas automáticas
	AllowGroups     bool `gorm:"default:false;index:idx_channels_allow_groups"` // Se o canal aceita mensagens de grupos WhatsApp
	TrackingEnabled bool `gorm:"default:false;index:idx_channels_tracking_enabled"` // Se o canal rastreia origem das mensagens

	// Message Debouncer Configuration
	DebounceTimeoutMs int `gorm:"not null;default:15000"` // Timeout do debouncer em milissegundos (default: 15s)

	// Estatísticas genéricas (aplicam a todos os tipos)
	MessagesReceived int        `gorm:"default:0"`
	MessagesSent     int        `gorm:"default:0"`
	LastMessageAt    *time.Time `gorm:"index:idx_channels_last_message"`
	LastErrorAt      *time.Time `gorm:"index:idx_channels_last_error"`
	LastError        string

	// Timestamps
	CreatedAt time.Time      `gorm:"autoCreateTime;index:idx_channels_created"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime;index:idx_channels_updated"`
	DeletedAt gorm.DeletedAt `gorm:"index:idx_channels_deleted"`

	// Relacionamentos
	User     UserEntity      `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	Project  ProjectEntity   `gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE"`
	Pipeline *PipelineEntity `gorm:"foreignKey:PipelineID;constraint:OnDelete:SET NULL"`
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
