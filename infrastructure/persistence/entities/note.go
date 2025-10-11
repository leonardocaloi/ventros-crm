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
	ContactID uuid.UUID  `gorm:"type:uuid;not null;index:idx_notes_contact;index:idx_notes_tenant_contact,priority:2"`
	SessionID *uuid.UUID `gorm:"type:uuid;index:idx_notes_session"`
	TenantID  string     `gorm:"not null;index:idx_notes_tenant;index:idx_notes_tenant_contact,priority:1;index:idx_notes_tenant_type,priority:1;index:idx_notes_tenant_priority,priority:1"`

	// Autoria
	AuthorID   uuid.UUID `gorm:"type:uuid;not null;index:idx_notes_author"`
	AuthorType string    `gorm:"not null;index:idx_notes_author_type"` // agent, system, user
	AuthorName string    `gorm:"not null"`

	// Conteúdo
	Content  string `gorm:"type:text;not null"`
	NoteType string `gorm:"not null;index:idx_notes_type;index:idx_notes_tenant_type,priority:2"`                          // general, automation, complaint, resolution, etc
	Priority string `gorm:"not null;index:idx_notes_priority;index:idx_notes_tenant_priority,priority:2;default:'normal'"` // low, normal, high, urgent

	// Visibilidade
	VisibleToClient bool `gorm:"default:false;index:idx_notes_visible"`
	Pinned          bool `gorm:"default:false;index:idx_notes_pinned"`

	// Metadata
	Tags        pq.StringArray `gorm:"type:text[];index:idx_notes_tags,type:gin"`
	Mentions    []byte         `gorm:"type:jsonb;index:idx_notes_mentions,type:gin"` // Array de UUIDs de agentes mencionados
	Attachments pq.StringArray `gorm:"type:text[]"`                                  // URLs de anexos

	// Timestamps
	CreatedAt time.Time      `gorm:"autoCreateTime;index:idx_notes_created"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime;index:idx_notes_updated"`
	DeletedAt gorm.DeletedAt `gorm:"index:idx_notes_deleted"`

	// Relacionamentos
	Contact ContactEntity  `gorm:"foreignKey:ContactID"`
	Session *SessionEntity `gorm:"foreignKey:SessionID"`
}

func (NoteEntity) TableName() string {
	return "notes"
}
