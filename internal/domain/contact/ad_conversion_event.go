package contact

import (
	"time"

	"github.com/google/uuid"
)

// AdConversionTrackedEvent é emitido quando uma mensagem de anúncio (FB/Instagram) é recebida.
// Permite rastreamento de ROI e atribuição de conversões.
type AdConversionTrackedEvent struct {
	ContactID uuid.UUID
	SessionID uuid.UUID
	TenantID  string
	
	// Campaign information
	ConversionSource string // "ctwa_ad", "facebook_ads", etc
	ConversionApp    string // "instagram", "facebook"
	
	// Ad details
	AdSourceType string // "ad"
	AdSourceID   string // Ad ID on platform
	AdSourceApp  string // Platform name
	AdSourceURL  string // Link to ad/post
	
	// Click tracking
	CTWAClickID string // Click-to-WhatsApp Click ID for conversion tracking
	
	// Additional tracking
	ConversionData    string // Encrypted payload from platform
	ExternalSource    string // "FB_Ads"
	ExternalMedium    string // "unavailable"
	
	TrackedAt time.Time
}

func (e AdConversionTrackedEvent) EventName() string     { return "ad_campaign.tracked" }
func (e AdConversionTrackedEvent) OccurredAt() time.Time { return e.TrackedAt }

// NewAdConversionTrackedEvent cria um novo evento de rastreamento de conversão de anúncio.
func NewAdConversionTrackedEvent(
	contactID uuid.UUID,
	sessionID uuid.UUID,
	tenantID string,
	trackingData map[string]string,
) AdConversionTrackedEvent {
	return AdConversionTrackedEvent{
		ContactID:         contactID,
		SessionID:         sessionID,
		TenantID:          tenantID,
		ConversionSource:  trackingData["conversion_source"],
		ConversionApp:     trackingData["conversion_app"],
		AdSourceType:      trackingData["ad_source_type"],
		AdSourceID:        trackingData["ad_source_id"],
		AdSourceApp:       trackingData["ad_source_app"],
		AdSourceURL:       trackingData["ad_source_url"],
		CTWAClickID:       trackingData["ctwa_clid"],
		ConversionData:    trackingData["conversion_data"],
		ExternalSource:    trackingData["external_source"],
		ExternalMedium:    trackingData["external_medium"],
		TrackedAt:         time.Now(),
	}
}

// ToContactEventPayload converte o evento para payload de ContactEvent.
func (e AdConversionTrackedEvent) ToContactEventPayload() map[string]interface{} {
	payload := make(map[string]interface{})
	
	if e.ConversionSource != "" {
		payload["conversion_source"] = e.ConversionSource
	}
	if e.ConversionApp != "" {
		payload["conversion_app"] = e.ConversionApp
	}
	if e.AdSourceType != "" {
		payload["ad_source_type"] = e.AdSourceType
	}
	if e.AdSourceID != "" {
		payload["ad_source_id"] = e.AdSourceID
	}
	if e.AdSourceApp != "" {
		payload["ad_source_app"] = e.AdSourceApp
	}
	if e.AdSourceURL != "" {
		payload["ad_source_url"] = e.AdSourceURL
	}
	if e.CTWAClickID != "" {
		payload["ctwa_click_id"] = e.CTWAClickID
	}
	
	return payload
}

// GetTitle retorna um título legível para o evento.
func (e AdConversionTrackedEvent) GetTitle() string {
	if e.AdSourceApp != "" {
		return "Message from " + e.AdSourceApp + " ad"
	}
	return "Message from ad"
}

// GetDescription retorna uma descrição legível para o evento.
func (e AdConversionTrackedEvent) GetDescription() string {
	if e.AdSourceID != "" {
		return "Contact came from ad campaign (ID: " + e.AdSourceID + ")"
	}
	return "Contact came from ad campaign"
}
