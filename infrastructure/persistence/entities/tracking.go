package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TrackingEntity representa um tracking de conversão no banco de dados
type TrackingEntity struct {
	ID        uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	ContactID uuid.UUID  `gorm:"type:uuid;not null;index:idx_trackings_contact_id"`
	SessionID *uuid.UUID `gorm:"type:uuid;index:idx_trackings_session_id"`
	TenantID  string     `gorm:"not null;index:idx_trackings_tenant_id"`
	ProjectID uuid.UUID  `gorm:"type:uuid;not null;index:idx_trackings_project_id"`

	// Ad Tracking
	Source   string `gorm:"not null;index:idx_trackings_source"`    // meta_ads, google_ads, etc
	Platform string `gorm:"not null;index:idx_trackings_platform"`  // instagram, facebook, etc
	Campaign string `gorm:"index:idx_trackings_campaign"`           // Nome/ID da campanha
	AdID     string `gorm:"column:ad_id;index:idx_trackings_ad_id"` // ID do anúncio
	AdURL    string `gorm:"column:ad_url"`                          // URL do anúncio/post

	// Click & Conversion Tracking
	ClickID        string `gorm:"column:click_id;uniqueIndex:idx_trackings_click_id"` // CTWA click ID (único)
	ConversionData string `gorm:"type:text"`                                          // Dados encriptados

	// UTM Parameters
	UTMSource   string `gorm:"column:utm_source"`
	UTMMedium   string `gorm:"column:utm_medium"`
	UTMCampaign string `gorm:"column:utm_campaign"`
	UTMTerm     string `gorm:"column:utm_term"`
	UTMContent  string `gorm:"column:utm_content"`

	// Metadata
	Metadata map[string]interface{} `gorm:"type:jsonb"`

	CreatedAt time.Time      `gorm:"not null;index:idx_trackings_created_at"`
	UpdatedAt time.Time      `gorm:"not null"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (TrackingEntity) TableName() string {
	return "trackings"
}
