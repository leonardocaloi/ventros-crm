package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// MessageEnrichmentEntity representa o enriquecimento de uma mensagem com mídia
// Armazena o resultado do processamento de IA (transcrição, OCR, parsing, etc)
type MessageEnrichmentEntity struct {
	ID             uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	MessageID      uuid.UUID      `gorm:"type:uuid;not null;index:idx_enrichments_message"`
	MessageGroupID uuid.UUID      `gorm:"type:uuid;not null;index:idx_enrichments_group"`
	ContentType    string         `gorm:"type:varchar(50);not null;index:idx_enrichments_content_type"` // audio, voice, image, video, document
	Provider       string         `gorm:"type:varchar(50);not null"`                                    // whisper, deepgram, vision, llamaparse, ffmpeg
	MediaURL       string         `gorm:"type:text;not null"`
	Status         string         `gorm:"type:varchar(50);not null;default:'pending';index:idx_enrichments_status"` // pending, processing, completed, failed
	ExtractedText  *string        `gorm:"type:text"`                                                                 // Texto extraído (transcrição, OCR, parsing)
	Metadata       datatypes.JSON `gorm:"type:jsonb"`                                                                // Metadados do provider (segments, objects, etc)
	ProcessingTime *int           `gorm:"column:processing_time_ms"`                                                 // Tempo de processamento em milliseconds
	Error          *string        `gorm:"type:text"`                                                                 // Mensagem de erro (se falhou)
	Context        *string        `gorm:"type:varchar(50)"`                                                          // Contexto de processamento (chat_message, profile_picture, etc)
	CreatedAt      time.Time      `gorm:"not null;default:now();index:idx_enrichments_created"`
	ProcessedAt    *time.Time

	// Relacionamentos
	Message      MessageEntity      `gorm:"foreignKey:MessageID;constraint:OnDelete:CASCADE"`
	MessageGroup MessageGroupEntity `gorm:"foreignKey:MessageGroupID;constraint:OnDelete:CASCADE"`
}

func (MessageEnrichmentEntity) TableName() string {
	return "message_enrichments"
}
