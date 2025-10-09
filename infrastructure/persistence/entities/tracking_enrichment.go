package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TrackingEnrichmentEntity armazena dados enriquecidos de tracking (Meta Ads API, etc)
type TrackingEnrichmentEntity struct {
	ID         uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TrackingID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_enrichments_tracking_id"`
	TenantID   string    `gorm:"not null;index:idx_enrichments_tenant_id"`

	// Origem do enriquecimento
	Source string `gorm:"not null;index:idx_enrichments_source"` // meta_ads, google_ads, etc

	// Dados enriquecidos da Meta Ads API
	AdAccountID   string `gorm:"column:ad_account_id"`
	AdAccountName string `gorm:"column:ad_account_name"`
	CampaignID    string `gorm:"column:campaign_id"`
	CampaignName  string `gorm:"column:campaign_name"`
	AdSetID       string `gorm:"column:adset_id"`
	AdSetName     string `gorm:"column:adset_name"`
	AdID          string `gorm:"column:ad_id"`
	AdName        string `gorm:"column:ad_name"`
	AdCreativeID  string `gorm:"column:ad_creative_id"`

	// Informações do criativo
	CreativeType   string `gorm:"column:creative_type"`   // image, video, carousel, collection
	CreativeFormat string `gorm:"column:creative_format"` // stories, feed, reels
	CreativeBody   string `gorm:"type:text;column:creative_body"`
	CreativeTitle  string `gorm:"column:creative_title"`
	CreativeURL    string `gorm:"column:creative_url"`

	// Targeting & Audience
	TargetingData string `gorm:"type:jsonb;column:targeting_data"` // Dados de segmentação/público
	AudienceName  string `gorm:"column:audience_name"`

	// Métricas (snapshot no momento do enriquecimento)
	Impressions int64   `gorm:"column:impressions"`
	Clicks      int64   `gorm:"column:clicks"`
	Spend       float64 `gorm:"column:spend"`
	CTR         float64 `gorm:"column:ctr"` // Click-through rate
	CPC         float64 `gorm:"column:cpc"` // Cost per click

	// Raw data completo da API
	RawAPIData string `gorm:"type:jsonb;column:raw_api_data"`

	// Metadados do enriquecimento
	EnrichedAt     time.Time `gorm:"not null;index:idx_enrichments_enriched_at"`
	EnrichmentType string    `gorm:"column:enrichment_type"` // automatic, manual, scheduled
	APIVersion     string    `gorm:"column:api_version"`     // Ex: v18.0

	CreatedAt time.Time      `gorm:"not null"`
	UpdatedAt time.Time      `gorm:"not null"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (TrackingEnrichmentEntity) TableName() string {
	return "tracking_enrichments"
}
