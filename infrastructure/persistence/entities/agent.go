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
	AgentTypeVirtual AgentType = "virtual" // Agente virtual (representa pessoa do passado)
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
	ID        uuid.UUID   `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	ProjectID uuid.UUID   `gorm:"type:uuid;not null;index:idx_agents_project"`
	UserID    *uuid.UUID  `gorm:"type:uuid;index:idx_agents_user"` // Null para agentes não-humanos
	TenantID  string      `gorm:"not null;index:idx_agents_tenant;index:idx_agents_tenant_type,priority:1;index:idx_agents_tenant_active,priority:1;index:idx_agents_tenant_status,priority:1"`
	Name      string      `gorm:"not null;index:idx_agents_name"`
	Email     string      `gorm:"index:idx_agents_email"`
	Type      AgentType   `gorm:"type:text;not null;index:idx_agents_type;index:idx_agents_tenant_type,priority:2;default:'human'"` // human, ai, bot, channel
	Status    AgentStatus `gorm:"type:text;not null;index:idx_agents_status;index:idx_agents_tenant_status,priority:2;default:'offline'"`
	Active    bool        `gorm:"default:true;index:idx_agents_active;index:idx_agents_tenant_active,priority:2"`

	// Configurações específicas por tipo
	Config map[string]interface{} `gorm:"type:jsonb;index:idx_agents_config,type:gin"` // Para AI: provider, model, api_key, etc

	// Virtual agent metadata (for historical representation)
	VirtualMetadata map[string]interface{} `gorm:"type:jsonb;index:idx_agents_virtual_metadata,type:gin"`

	// Métricas
	SessionsHandled   int        `gorm:"default:0"`
	AverageResponseMs int        `gorm:"default:0"` // Tempo médio de resposta em ms
	LastActivityAt    *time.Time `gorm:"index:idx_agents_last_activity"`

	CreatedAt time.Time      `gorm:"autoCreateTime;index:idx_agents_created"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime;index:idx_agents_updated"`
	DeletedAt gorm.DeletedAt `gorm:"index:idx_agents_deleted"`

	// Relacionamentos
	Project ProjectEntity `gorm:"foreignKey:ProjectID"`
	User    *UserEntity   `gorm:"foreignKey:UserID"`
}

func (AgentEntity) TableName() string {
	return "agents"
}
