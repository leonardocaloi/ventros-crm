package contact_test

import (
	"testing"

	"github.com/caloi/ventros-crm/internal/domain/contact"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAdConversionTrackedEvent(t *testing.T) {
	// Arrange
	contactID := uuid.New()
	sessionID := uuid.New()
	tenantID := "tenant_123"
	
	trackingData := map[string]string{
		"conversion_source": "ctwa_ad",
		"conversion_app":    "instagram",
		"ad_source_type":    "ad",
		"ad_source_id":      "120232639087330785",
		"ad_source_app":     "instagram",
		"ad_source_url":     "https://www.instagram.com/p/DOqYjmWDOCx/",
		"ctwa_clid":         "Afcg5DA4aj8pfp0faj5HIKtKi2vOUGt4agVezOhRfZA0",
		"conversion_data":   "QWZkU0JBdl9lOUwtVk5sTzJy...",
		"external_source":   "FB_Ads",
		"external_medium":   "unavailable",
	}
	
	// Act
	event := contact.NewAdConversionTrackedEvent(contactID, sessionID, tenantID, trackingData)
	
	// Assert
	assert.Equal(t, contactID, event.ContactID)
	assert.Equal(t, sessionID, event.SessionID)
	assert.Equal(t, tenantID, event.TenantID)
	assert.Equal(t, "ctwa_ad", event.ConversionSource)
	assert.Equal(t, "instagram", event.ConversionApp)
	assert.Equal(t, "ad", event.AdSourceType)
	assert.Equal(t, "120232639087330785", event.AdSourceID)
	assert.Equal(t, "instagram", event.AdSourceApp)
	assert.Equal(t, "https://www.instagram.com/p/DOqYjmWDOCx/", event.AdSourceURL)
	assert.NotEmpty(t, event.CTWAClickID)
	assert.Equal(t, "FB_Ads", event.ExternalSource)
	assert.False(t, event.TrackedAt.IsZero())
}

func TestAdConversionTrackedEvent_EventName(t *testing.T) {
	// Arrange
	event := contact.NewAdConversionTrackedEvent(
		uuid.New(),
		uuid.New(),
		"tenant_123",
		map[string]string{},
	)
	
	// Act
	eventName := event.EventName()
	
	// Assert
	assert.Equal(t, "ad_campaign.tracked", eventName)
}

func TestAdConversionTrackedEvent_OccurredAt(t *testing.T) {
	// Arrange
	event := contact.NewAdConversionTrackedEvent(
		uuid.New(),
		uuid.New(),
		"tenant_123",
		map[string]string{},
	)
	
	// Act
	occurredAt := event.OccurredAt()
	
	// Assert
	assert.False(t, occurredAt.IsZero())
	assert.Equal(t, event.TrackedAt, occurredAt)
}

func TestAdConversionTrackedEvent_ToContactEventPayload(t *testing.T) {
	// Arrange
	trackingData := map[string]string{
		"conversion_source": "ctwa_ad",
		"conversion_app":    "instagram",
		"ad_source_type":    "ad",
		"ad_source_id":      "120232639087330785",
		"ad_source_app":     "instagram",
		"ad_source_url":     "https://www.instagram.com/p/DOqYjmWDOCx/",
		"ctwa_clid":         "Afcg5DA4aj8pfp0faj5HIKtKi2vOUGt4agVezOhRfZA0",
	}
	
	event := contact.NewAdConversionTrackedEvent(
		uuid.New(),
		uuid.New(),
		"tenant_123",
		trackingData,
	)
	
	// Act
	payload := event.ToContactEventPayload()
	
	// Assert
	require.NotEmpty(t, payload)
	assert.Equal(t, "ctwa_ad", payload["conversion_source"])
	assert.Equal(t, "instagram", payload["conversion_app"])
	assert.Equal(t, "ad", payload["ad_source_type"])
	assert.Equal(t, "120232639087330785", payload["ad_source_id"])
	assert.Equal(t, "instagram", payload["ad_source_app"])
	assert.Equal(t, "https://www.instagram.com/p/DOqYjmWDOCx/", payload["ad_source_url"])
	assert.NotEmpty(t, payload["ctwa_click_id"])
}

func TestAdConversionTrackedEvent_ToContactEventPayload_EmptyFields(t *testing.T) {
	// Arrange - evento com dados vazios
	event := contact.NewAdConversionTrackedEvent(
		uuid.New(),
		uuid.New(),
		"tenant_123",
		map[string]string{},
	)
	
	// Act
	payload := event.ToContactEventPayload()
	
	// Assert - não deve incluir campos vazios
	assert.Empty(t, payload)
}

func TestAdConversionTrackedEvent_GetTitle(t *testing.T) {
	tests := []struct {
		name             string
		trackingData     map[string]string
		expectedTitle    string
	}{
		{
			name: "Instagram ad",
			trackingData: map[string]string{
				"ad_source_app": "instagram",
			},
			expectedTitle: "Message from instagram ad",
		},
		{
			name: "Facebook ad",
			trackingData: map[string]string{
				"ad_source_app": "facebook",
			},
			expectedTitle: "Message from facebook ad",
		},
		{
			name:          "No app specified",
			trackingData:  map[string]string{},
			expectedTitle: "Message from ad",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			event := contact.NewAdConversionTrackedEvent(
				uuid.New(),
				uuid.New(),
				"tenant_123",
				tt.trackingData,
			)
			
			// Act
			title := event.GetTitle()
			
			// Assert
			assert.Equal(t, tt.expectedTitle, title)
		})
	}
}

func TestAdConversionTrackedEvent_GetDescription(t *testing.T) {
	tests := []struct {
		name                string
		trackingData        map[string]string
		expectedDescription string
	}{
		{
			name: "With ad ID",
			trackingData: map[string]string{
				"ad_source_id": "120232639087330785",
			},
			expectedDescription: "Contact came from ad campaign (ID: 120232639087330785)",
		},
		{
			name:                "Without ad ID",
			trackingData:        map[string]string{},
			expectedDescription: "Contact came from ad campaign",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			event := contact.NewAdConversionTrackedEvent(
				uuid.New(),
				uuid.New(),
				"tenant_123",
				tt.trackingData,
			)
			
			// Act
			description := event.GetDescription()
			
			// Assert
			assert.Equal(t, tt.expectedDescription, description)
		})
	}
}

func TestAdConversionTrackedEvent_FullScenario(t *testing.T) {
	// Arrange - Simula cenário completo de mensagem vindo de anúncio do Instagram
	contactID := uuid.New()
	sessionID := uuid.New()
	tenantID := "ask-dermatologia"
	
	trackingData := map[string]string{
		"conversion_source": "ctwa_ad",
		"conversion_app":    "instagram",
		"ad_source_type":    "ad",
		"ad_source_id":      "120232639087330785",
		"ad_source_app":     "instagram",
		"ad_source_url":     "https://www.instagram.com/p/DOqYjmWDOCx/",
		"ctwa_clid":         "Afcg5DA4aj8pfp0faj5HIKtKi2vOUGt4agVezOhRfZA0MLUab3uUE98YGiB3KDFGsjR7Ohm5nIHEHugiymilThQM5gEEr_XAu5sHLZX5GGcnc9v9y3ImoTXetXaKP640XR222lch",
		"external_source":   "FB_Ads",
		"external_medium":   "unavailable",
	}
	
	// Act
	event := contact.NewAdConversionTrackedEvent(contactID, sessionID, tenantID, trackingData)
	
	// Assert - Verifica todos os dados necessários para tracking de conversão
	assert.Equal(t, "ad_campaign.tracked", event.EventName())
	assert.Equal(t, "ctwa_ad", event.ConversionSource)
	assert.Equal(t, "instagram", event.ConversionApp)
	assert.NotEmpty(t, event.CTWAClickID, "CTWA Click ID é crucial para rastreamento de conversão")
	assert.Equal(t, "120232639087330785", event.AdSourceID, "Ad ID necessário para ROI tracking")
	
	// Assert - Payload pode ser usado para criar ContactEvent
	payload := event.ToContactEventPayload()
	require.NotEmpty(t, payload)
	assert.Contains(t, payload, "ctwa_click_id")
	assert.Contains(t, payload, "ad_source_id")
	
	// Assert - Títulos legíveis para UI
	assert.Equal(t, "Message from instagram ad", event.GetTitle())
	assert.Contains(t, event.GetDescription(), "120232639087330785")
}
