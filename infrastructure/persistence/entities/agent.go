package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AgentType define os tipos de agentes no sistema
type AgentType string

const (
	AgentTypeHuman   AgentType = "human"   // Agente humano (atendente/admin)
	AgentTypeAI      AgentType = "ai"      // Agente de IA (externo via provider)
	AgentTypeBot     AgentType = "bot"     // Bot/automação (interno)
	AgentTypeChannel AgentType = "channel" // Canal/dispositivo
)

// AgentStatus define os status possíveis de um agente
type AgentStatus string

const (
	AgentStatusAvailable AgentStatus = "available" // Disponível para atender
	AgentStatusBusy      AgentStatus = "busy"      // Ocupado atendendo
	AgentStatusAway      AgentStatus = "away"      // Ausente
	AgentStatusOffline   AgentStatus = "offline"   // Offline/desconectado
)

// AgentEntity representa a entidade Agent no banco de dados
type AgentEntity struct {
	ID        uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	ProjectID uuid.UUID      `gorm:"type:uuid;not null;index"`
	UserID    *uuid.UUID     `gorm:"type:uuid;index"` // Null para agentes não-humanos
	TenantID  string         `gorm:"not null;index"`
	Name      string         `gorm:"not null"`
	Email     string         `gorm:""`
	Type      AgentType      `gorm:"type:text;not null;index;default:'human'"` // human, ai, bot, channel
	Status    AgentStatus    `gorm:"type:text;not null;index;default:'offline'"`
	Active    bool           `gorm:"default:true;index"`
	
	// Configurações específicas por tipo
	Config map[string]interface{} `gorm:"type:jsonb"` // Para AI: provider, model, api_key, etc
	
	// Métricas
	SessionsHandled   int        `gorm:"default:0"`
	AverageResponseMs int        `gorm:"default:0"` // Tempo médio de resposta em ms
	LastActivityAt    *time.Time `gorm:""`
	
	CreatedAt time.Time      `gorm:"autoCreateTime"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
	
	// Relacionamentos
	Project ProjectEntity `gorm:"foreignKey:ProjectID"`
	User    *UserEntity   `gorm:"foreignKey:UserID"`
}

func (AgentEntity) TableName() string {
	return "agents"
}
