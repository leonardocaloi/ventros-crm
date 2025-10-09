package tracking

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrTrackingNotFound = errors.New("tracking not found")
	ErrInvalidContactID = errors.New("invalid contact ID")
	ErrInvalidSource    = errors.New("invalid tracking source")
)

// Source representa as fontes de tracking suportadas
type Source string

const (
	SourceMetaAds   Source = "meta_ads"   // Facebook/Instagram Ads
	SourceGoogleAds Source = "google_ads" // Google Ads
	SourceTikTokAds Source = "tiktok_ads" // TikTok Ads
	SourceLinkedIn  Source = "linkedin"   // LinkedIn Ads
	SourceOrganic   Source = "organic"    // Tráfego orgânico
	SourceDirect    Source = "direct"     // Acesso direto
	SourceReferral  Source = "referral"   // Referência
	SourceOther     Source = "other"      // Outros
)

// Platform representa as plataformas de origem
type Platform string

const (
	PlatformInstagram Platform = "instagram"
	PlatformFacebook  Platform = "facebook"
	PlatformGoogle    Platform = "google"
	PlatformTikTok    Platform = "tiktok"
	PlatformLinkedIn  Platform = "linkedin"
	PlatformWhatsApp  Platform = "whatsapp"
	PlatformOther     Platform = "other"
)

// Tracking é o Aggregate Root para tracking de conversões e atribuição.
// Rastreia origens de contatos vindos de anúncios, campanhas, etc.
type Tracking struct {
	id        uuid.UUID
	contactID uuid.UUID  // Referência ao contato
	sessionID *uuid.UUID // Referência à sessão (opcional)
	tenantID  string
	projectID uuid.UUID

	// Ad Tracking - Informações de origem
	source   Source   // Fonte principal: meta_ads, google_ads, etc
	platform Platform // Plataforma: instagram, facebook, etc
	campaign string   // Nome/ID da campanha
	adID     string   // ID do anúncio na plataforma
	adURL    string   // URL do anúncio/post

	// Click & Conversion Tracking
	clickID        string // Click-to-WhatsApp Click ID (CTWA)
	conversionData string // Dados encriptados da plataforma

	// UTM Parameters
	utmSource   string // utm_source
	utmMedium   string // utm_medium
	utmCampaign string // utm_campaign
	utmTerm     string // utm_term
	utmContent  string // utm_content

	// Metadata
	metadata map[string]interface{} // Dados adicionais específicos da fonte

	createdAt time.Time
	updatedAt time.Time

	// Domain Events
	events []DomainEvent
}

// NewTracking cria um novo tracking de conversão.
func NewTracking(
	contactID uuid.UUID,
	sessionID *uuid.UUID,
	tenantID string,
	projectID uuid.UUID,
	source Source,
	platform Platform,
) (*Tracking, error) {
	if contactID == uuid.Nil {
		return nil, ErrInvalidContactID
	}
	if tenantID == "" {
		return nil, errors.New("tenantID cannot be empty")
	}
	if projectID == uuid.Nil {
		return nil, errors.New("projectID cannot be nil")
	}
	if source == "" {
		return nil, ErrInvalidSource
	}

	now := time.Now()
	tracking := &Tracking{
		id:        uuid.New(),
		contactID: contactID,
		sessionID: sessionID,
		tenantID:  tenantID,
		projectID: projectID,
		source:    source,
		platform:  platform,
		metadata:  make(map[string]interface{}),
		createdAt: now,
		updatedAt: now,
		events:    []DomainEvent{},
	}

	tracking.addEvent(NewTrackingCreatedEvent(
		tracking.id,
		contactID,
		projectID,
		sessionID,
		tenantID,
		string(source),
		string(platform),
	))

	return tracking, nil
}

// ReconstructTracking reconstrói um tracking a partir de dados persistidos.
func ReconstructTracking(
	id uuid.UUID,
	contactID uuid.UUID,
	sessionID *uuid.UUID,
	tenantID string,
	projectID uuid.UUID,
	source Source,
	platform Platform,
	campaign string,
	adID string,
	adURL string,
	clickID string,
	conversionData string,
	utmSource string,
	utmMedium string,
	utmCampaign string,
	utmTerm string,
	utmContent string,
	metadata map[string]interface{},
	createdAt time.Time,
	updatedAt time.Time,
) *Tracking {
	return &Tracking{
		id:             id,
		contactID:      contactID,
		sessionID:      sessionID,
		tenantID:       tenantID,
		projectID:      projectID,
		source:         source,
		platform:       platform,
		campaign:       campaign,
		adID:           adID,
		adURL:          adURL,
		clickID:        clickID,
		conversionData: conversionData,
		utmSource:      utmSource,
		utmMedium:      utmMedium,
		utmCampaign:    utmCampaign,
		utmTerm:        utmTerm,
		utmContent:     utmContent,
		metadata:       metadata,
		createdAt:      createdAt,
		updatedAt:      updatedAt,
		events:         []DomainEvent{},
	}
}

// SetCampaign define a campanha do tracking.
func (t *Tracking) SetCampaign(campaign string) {
	if t.campaign != campaign {
		t.campaign = campaign
		t.updatedAt = time.Now()
	}
}

// SetAdInfo define informações do anúncio.
func (t *Tracking) SetAdInfo(adID, adURL string) {
	changed := false
	if t.adID != adID {
		t.adID = adID
		changed = true
	}
	if t.adURL != adURL {
		t.adURL = adURL
		changed = true
	}
	if changed {
		t.updatedAt = time.Now()
	}
}

// SetClickID define o click ID (CTWA).
func (t *Tracking) SetClickID(clickID string) {
	if t.clickID != clickID {
		t.clickID = clickID
		t.updatedAt = time.Now()
	}
}

// SetConversionData define os dados de conversão encriptados.
func (t *Tracking) SetConversionData(data string) {
	if t.conversionData != data {
		t.conversionData = data
		t.updatedAt = time.Now()
	}
}

// SetUTMParameters define os parâmetros UTM.
func (t *Tracking) SetUTMParameters(source, medium, campaign, term, content string) {
	changed := false
	if t.utmSource != source {
		t.utmSource = source
		changed = true
	}
	if t.utmMedium != medium {
		t.utmMedium = medium
		changed = true
	}
	if t.utmCampaign != campaign {
		t.utmCampaign = campaign
		changed = true
	}
	if t.utmTerm != term {
		t.utmTerm = term
		changed = true
	}
	if t.utmContent != content {
		t.utmContent = content
		changed = true
	}
	if changed {
		t.updatedAt = time.Now()
	}
}

// SetMetadata define metadados adicionais.
func (t *Tracking) SetMetadata(metadata map[string]interface{}) {
	t.metadata = metadata
	t.updatedAt = time.Now()
}

// AddMetadata adiciona um metadado específico.
func (t *Tracking) AddMetadata(key string, value interface{}) {
	if t.metadata == nil {
		t.metadata = make(map[string]interface{})
	}
	t.metadata[key] = value
	t.updatedAt = time.Now()
}

// Enrich enriquece o tracking com dados adicionais.
func (t *Tracking) Enrich(changes map[string]interface{}) {
	t.updatedAt = time.Now()

	t.addEvent(NewTrackingEnrichedEvent(
		t.id,
		t.contactID,
		changes,
	))
}

// Getters
func (t *Tracking) ID() uuid.UUID                    { return t.id }
func (t *Tracking) ContactID() uuid.UUID             { return t.contactID }
func (t *Tracking) SessionID() *uuid.UUID            { return t.sessionID }
func (t *Tracking) TenantID() string                 { return t.tenantID }
func (t *Tracking) ProjectID() uuid.UUID             { return t.projectID }
func (t *Tracking) Source() Source                   { return t.source }
func (t *Tracking) Platform() Platform               { return t.platform }
func (t *Tracking) Campaign() string                 { return t.campaign }
func (t *Tracking) AdID() string                     { return t.adID }
func (t *Tracking) AdURL() string                    { return t.adURL }
func (t *Tracking) ClickID() string                  { return t.clickID }
func (t *Tracking) ConversionData() string           { return t.conversionData }
func (t *Tracking) UTMSource() string                { return t.utmSource }
func (t *Tracking) UTMMedium() string                { return t.utmMedium }
func (t *Tracking) UTMCampaign() string              { return t.utmCampaign }
func (t *Tracking) UTMTerm() string                  { return t.utmTerm }
func (t *Tracking) UTMContent() string               { return t.utmContent }
func (t *Tracking) Metadata() map[string]interface{} { return t.metadata }
func (t *Tracking) CreatedAt() time.Time             { return t.createdAt }
func (t *Tracking) UpdatedAt() time.Time             { return t.updatedAt }

// Domain Events
func (t *Tracking) DomainEvents() []DomainEvent {
	return t.events
}

func (t *Tracking) ClearEvents() {
	t.events = []DomainEvent{}
}

func (t *Tracking) addEvent(event DomainEvent) {
	t.events = append(t.events, event)
}
