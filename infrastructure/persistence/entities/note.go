package entities

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

// NoteEntity representa a entidade Note no banco de dados
type NoteEntity struct {
	ID        uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	ContactID uuid.UUID  `gorm:"type:uuid;not null;index"`
	SessionID *uuid.UUID `gorm:"type:uuid;index"`
	TenantID  string     `gorm:"not null;index"`

	// Autoria
	AuthorID   uuid.UUID `gorm:"type:uuid;not null;index"`
	AuthorType string    `gorm:"not null;index"` // agent, system, user
	AuthorName string    `gorm:"not null"`

	// Conteúdo
	Content  string `gorm:"type:text;not null"`
	NoteType string `gorm:"not null;index"`                  // general, automation, complaint, resolution, etc
	Priority string `gorm:"not null;index;default:'normal'"` // low, normal, high, urgent

	// Visibilidade
	VisibleToClient bool `gorm:"default:false;index"`
	Pinned          bool `gorm:"default:false;index"`

	// Metadata
	Tags        pq.StringArray `gorm:"type:text[]"`
	Mentions    []byte         `gorm:"type:jsonb"`  // Array de UUIDs de agentes mencionados
	Attachments pq.StringArray `gorm:"type:text[]"` // URLs de anexos

	// Timestamps
	CreatedAt time.Time      `gorm:"autoCreateTime"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `gorm:"index"`

	// Relacionamentos
	Contact ContactEntity  `gorm:"foreignKey:ContactID"`
	Session *SessionEntity `gorm:"foreignKey:SessionID"`
}

func (NoteEntity) TableName() string {
	return "notes"
}
