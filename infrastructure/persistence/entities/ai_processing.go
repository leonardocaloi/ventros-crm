package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// AIProcessingEntity representa um processamento de IA (resumo, análise, etc)
type AIProcessingEntity struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TenantID  string    `gorm:"not null;index"`
	ProjectID uuid.UUID `gorm:"type:uuid;not null;index"`

	// Relacionamento Genérico (Polymorphic)
	EntityType string    `gorm:"not null;index"`           // message, session, contact, profile_picture, etc
	EntityID   uuid.UUID `gorm:"type:uuid;not null;index"` // ID da entidade relacionada

	// Origem do processamento (opcional)
	SourceType *string    `gorm:"index"`           // event, webhook, manual, scheduled, etc
	SourceID   *uuid.UUID `gorm:"type:uuid;index"` // ID do evento/webhook/job que disparou

	// Relacionamentos opcionais (para queries e contexto)
	SessionID *uuid.UUID `gorm:"type:uuid;index"` // Sessão relacionada (se aplicável)
	ContactID *uuid.UUID `gorm:"type:uuid;index"` // Contato relacionado (se aplicável)
	ChannelID *uuid.UUID `gorm:"type:uuid;index"` // Canal relacionado (se aplicável)

	// Tipo de processamento
	ProcessingType string `gorm:"not null;index"`                   // transcription, ocr, sentiment, summary, profile_analysis, etc
	Status         string `gorm:"not null;index;default:'pending'"` // pending, processing, completed, failed

	// IA Provider
	Provider string `gorm:"not null"` // openai, anthropic, google, etc
	Model    string `gorm:"not null"` // gpt-4, claude-3-opus, gemini-pro, etc

	// Input/Output
	InputData  datatypes.JSON `gorm:"type:jsonb"` // Dados de entrada (mensagens, contexto, etc)
	OutputData datatypes.JSON `gorm:"type:jsonb"` // Resultado do processamento

	// Métricas
	TokensUsed       *int     `gorm:""` // Tokens consumidos
	ProcessingTimeMs *int     `gorm:""` // Tempo de processamento em ms
	Cost             *float64 `gorm:""` // Custo estimado em USD

	// Erro (se houver)
	ErrorMessage *string `gorm:"type:text"`
	RetryCount   int     `gorm:"default:0"`

	// Timestamps
	StartedAt   *time.Time     `gorm:""`
	CompletedAt *time.Time     `gorm:""`
	CreatedAt   time.Time      `gorm:"autoCreateTime"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime"`
	DeletedAt   gorm.DeletedAt `gorm:"index"`

	// Relacionamentos
	Session *SessionEntity `gorm:"foreignKey:SessionID"`
	Contact *ContactEntity `gorm:"foreignKey:ContactID"`
}

func (AIProcessingEntity) TableName() string {
	return "ai_processes"
}
