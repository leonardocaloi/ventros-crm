package tracking

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTracking_Valid(t *testing.T) {
	contactID := uuid.New()
	sessionID := uuid.New()
	tenantID := "tenant-123"
	projectID := uuid.New()

	tests := []struct {
		name     string
		source   Source
		platform Platform
	}{
		{
			name:     "Meta Ads tracking",
			source:   SourceMetaAds,
			platform: PlatformInstagram,
		},
		{
			name:     "Google Ads tracking",
			source:   SourceGoogleAds,
			platform: PlatformGoogle,
		},
		{
			name:     "TikTok Ads tracking",
			source:   SourceTikTokAds,
			platform: PlatformTikTok,
		},
		{
			name:     "Organic tracking",
			source:   SourceOrganic,
			platform: PlatformWhatsApp,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tracking, err := NewTracking(
				contactID,
				&sessionID,
				tenantID,
				projectID,
				tt.source,
				tt.platform,
			)

			require.NoError(t, err)
			assert.NotNil(t, tracking)
			assert.NotEqual(t, uuid.Nil, tracking.ID())
			assert.Equal(t, contactID, tracking.ContactID())
			assert.Equal(t, sessionID, *tracking.SessionID())
			assert.Equal(t, tenantID, tracking.TenantID())
			assert.Equal(t, projectID, tracking.ProjectID())
			assert.Equal(t, tt.source, tracking.Source())
			assert.Equal(t, tt.platform, tracking.Platform())
			assert.NotZero(t, tracking.CreatedAt())
			assert.NotZero(t, tracking.UpdatedAt())

			// Verify event was emitted
			events := tracking.DomainEvents()
			assert.Len(t, events, 1)
			createdEvent, ok := events[0].(TrackingCreatedEvent)
			assert.True(t, ok)
			assert.Equal(t, tracking.ID(), createdEvent.TrackingID)
			assert.Equal(t, contactID, createdEvent.ContactID)
		})
	}
}

func TestNewTracking_WithoutSession(t *testing.T) {
	contactID := uuid.New()
	tenantID := "tenant-123"
	projectID := uuid.New()

	tracking, err := NewTracking(
		contactID,
		nil, // No session
		tenantID,
		projectID,
		SourceDirect,
		PlatformOther,
	)

	require.NoError(t, err)
	assert.NotNil(t, tracking)
	assert.Nil(t, tracking.SessionID())
}

func TestNewTracking_Invalid(t *testing.T) {
	contactID := uuid.New()
	sessionID := uuid.New()
	tenantID := "tenant-123"
	projectID := uuid.New()

	tests := []struct {
		name      string
		contactID uuid.UUID
		sessionID *uuid.UUID
		tenantID  string
		projectID uuid.UUID
		source    Source
		platform  Platform
		expectErr error
	}{
		{
			name:      "nil contact ID",
			contactID: uuid.Nil,
			sessionID: &sessionID,
			tenantID:  tenantID,
			projectID: projectID,
			source:    SourceMetaAds,
			platform:  PlatformFacebook,
			expectErr: ErrInvalidContactID,
		},
		{
			name:      "empty tenant ID",
			contactID: contactID,
			sessionID: &sessionID,
			tenantID:  "",
			projectID: projectID,
			source:    SourceMetaAds,
			platform:  PlatformFacebook,
			expectErr: nil, // Contains "tenantID cannot be empty"
		},
		{
			name:      "nil project ID",
			contactID: contactID,
			sessionID: &sessionID,
			tenantID:  tenantID,
			projectID: uuid.Nil,
			source:    SourceMetaAds,
			platform:  PlatformFacebook,
			expectErr: nil, // Contains "projectID cannot be nil"
		},
		{
			name:      "empty source",
			contactID: contactID,
			sessionID: &sessionID,
			tenantID:  tenantID,
			projectID: projectID,
			source:    "",
			platform:  PlatformFacebook,
			expectErr: ErrInvalidSource,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tracking, err := NewTracking(
				tt.contactID,
				tt.sessionID,
				tt.tenantID,
				tt.projectID,
				tt.source,
				tt.platform,
			)

			assert.Error(t, err)
			assert.Nil(t, tracking)

			if tt.expectErr != nil {
				assert.Equal(t, tt.expectErr, err)
			}
		})
	}
}

func TestTracking_SetCampaign(t *testing.T) {
	tracking := createTestTracking(t)
	assert.Empty(t, tracking.Campaign())

	campaign := "Summer Sale 2024"
	tracking.SetCampaign(campaign)

	assert.Equal(t, campaign, tracking.Campaign())
}

func TestTracking_SetAdInfo(t *testing.T) {
	tracking := createTestTracking(t)
	assert.Empty(t, tracking.AdID())
	assert.Empty(t, tracking.AdURL())

	adID := "ad_123456"
	adURL := "https://www.instagram.com/p/abc123/"

	tracking.SetAdInfo(adID, adURL)

	assert.Equal(t, adID, tracking.AdID())
	assert.Equal(t, adURL, tracking.AdURL())
}

func TestTracking_SetClickID(t *testing.T) {
	tracking := createTestTracking(t)
	assert.Empty(t, tracking.ClickID())

	clickID := "ctwa_click_123"
	tracking.SetClickID(clickID)

	assert.Equal(t, clickID, tracking.ClickID())
}

func TestTracking_SetConversionData(t *testing.T) {
	tracking := createTestTracking(t)
	assert.Empty(t, tracking.ConversionData())

	conversionData := "encrypted_conversion_data_123"
	tracking.SetConversionData(conversionData)

	assert.Equal(t, conversionData, tracking.ConversionData())
}

func TestTracking_SetUTMParameters(t *testing.T) {
	tracking := createTestTracking(t)

	utmSource := "instagram"
	utmMedium := "paid_social"
	utmCampaign := "summer_sale"
	utmTerm := "shoes"
	utmContent := "carousel_ad"

	tracking.SetUTMParameters(utmSource, utmMedium, utmCampaign, utmTerm, utmContent)

	assert.Equal(t, utmSource, tracking.UTMSource())
	assert.Equal(t, utmMedium, tracking.UTMMedium())
	assert.Equal(t, utmCampaign, tracking.UTMCampaign())
	assert.Equal(t, utmTerm, tracking.UTMTerm())
	assert.Equal(t, utmContent, tracking.UTMContent())
}

func TestTracking_SetMetadata(t *testing.T) {
	tracking := createTestTracking(t)

	metadata := map[string]interface{}{
		"ad_set_id":    "adset_123",
		"creative_id":  "creative_456",
		"audience":     "lookalike_1",
		"budget":       1000.50,
		"is_retarget":  true,
	}

	tracking.SetMetadata(metadata)

	assert.Equal(t, metadata, tracking.Metadata())
	assert.Equal(t, "adset_123", tracking.Metadata()["ad_set_id"])
	assert.Equal(t, 1000.50, tracking.Metadata()["budget"])
}

func TestTracking_AddMetadata(t *testing.T) {
	tracking := createTestTracking(t)

	tracking.AddMetadata("key1", "value1")
	tracking.AddMetadata("key2", 123)
	tracking.AddMetadata("key3", true)

	metadata := tracking.Metadata()
	assert.Equal(t, "value1", metadata["key1"])
	assert.Equal(t, 123, metadata["key2"])
	assert.Equal(t, true, metadata["key3"])
}

func TestTracking_Enrich(t *testing.T) {
	tracking := createTestTracking(t)
	tracking.ClearEvents()

	changes := map[string]interface{}{
		"device_type":  "mobile",
		"browser":      "Chrome",
		"geo_location": "SP, Brazil",
	}

	tracking.Enrich(changes)

	// Verify event was emitted
	events := tracking.DomainEvents()
	assert.Len(t, events, 1)
	enrichedEvent, ok := events[0].(TrackingEnrichedEvent)
	assert.True(t, ok)
	assert.Equal(t, tracking.ID(), enrichedEvent.TrackingID)
	assert.Equal(t, tracking.ContactID(), enrichedEvent.ContactID)
	assert.Equal(t, changes, enrichedEvent.Changes)
}

func TestTracking_UpdatesTimestamp(t *testing.T) {
	tracking := createTestTracking(t)

	originalUpdatedAt := tracking.UpdatedAt()
	time.Sleep(10 * time.Millisecond) // Ensure time difference

	tests := []struct {
		name   string
		action func()
	}{
		{
			name: "SetCampaign updates timestamp",
			action: func() {
				tracking.SetCampaign("New Campaign")
			},
		},
		{
			name: "SetAdInfo updates timestamp",
			action: func() {
				tracking.SetAdInfo("new_ad", "new_url")
			},
		},
		{
			name: "SetClickID updates timestamp",
			action: func() {
				tracking.SetClickID("new_click_id")
			},
		},
		{
			name: "SetConversionData updates timestamp",
			action: func() {
				tracking.SetConversionData("new_data")
			},
		},
		{
			name: "SetUTMParameters updates timestamp",
			action: func() {
				tracking.SetUTMParameters("src", "med", "camp", "term", "cont")
			},
		},
		{
			name: "SetMetadata updates timestamp",
			action: func() {
				tracking.SetMetadata(map[string]interface{}{"key": "value"})
			},
		},
		{
			name: "AddMetadata updates timestamp",
			action: func() {
				tracking.AddMetadata("new_key", "new_value")
			},
		},
		{
			name: "Enrich updates timestamp",
			action: func() {
				tracking.Enrich(map[string]interface{}{"change": "value"})
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Record current timestamp
			beforeUpdate := tracking.UpdatedAt()
			time.Sleep(10 * time.Millisecond)

			// Execute action
			tt.action()

			// Verify timestamp was updated
			assert.True(t, tracking.UpdatedAt().After(beforeUpdate),
				"UpdatedAt should be after previous timestamp")
		})
	}

	assert.True(t, tracking.UpdatedAt().After(originalUpdatedAt))
}

func TestTracking_NoUpdateWhenValueUnchanged(t *testing.T) {
	tracking := createTestTracking(t)

	// Set initial values
	tracking.SetCampaign("Campaign 1")
	tracking.SetClickID("click_123")
	time.Sleep(10 * time.Millisecond)

	updatedAt := tracking.UpdatedAt()
	time.Sleep(10 * time.Millisecond)

	// Set same values again - should not update timestamp
	tracking.SetCampaign("Campaign 1")
	assert.Equal(t, updatedAt, tracking.UpdatedAt(), "UpdatedAt should not change when setting same campaign")

	tracking.SetClickID("click_123")
	assert.Equal(t, updatedAt, tracking.UpdatedAt(), "UpdatedAt should not change when setting same clickID")
}

func TestTracking_EventManagement(t *testing.T) {
	tracking := createTestTracking(t)

	t.Run("domain events are collected", func(t *testing.T) {
		events := tracking.DomainEvents()
		assert.Len(t, events, 1) // TrackingCreatedEvent
	})

	t.Run("clear events", func(t *testing.T) {
		tracking.ClearEvents()
		events := tracking.DomainEvents()
		assert.Len(t, events, 0)
	})
}

func TestReconstructTracking(t *testing.T) {
	id := uuid.New()
	contactID := uuid.New()
	sessionID := uuid.New()
	tenantID := "tenant-123"
	projectID := uuid.New()
	now := time.Now()

	metadata := map[string]interface{}{
		"custom_field": "custom_value",
	}

	tracking := ReconstructTracking(
		id,
		contactID,
		&sessionID,
		tenantID,
		projectID,
		SourceMetaAds,
		PlatformInstagram,
		"Summer Campaign",
		"ad_123",
		"https://instagram.com/p/123",
		"ctwa_click_456",
		"encrypted_data",
		"instagram",
		"paid_social",
		"summer_sale",
		"shoes",
		"carousel",
		metadata,
		now,
		now,
	)

	assert.NotNil(t, tracking)
	assert.Equal(t, id, tracking.ID())
	assert.Equal(t, contactID, tracking.ContactID())
	assert.Equal(t, sessionID, *tracking.SessionID())
	assert.Equal(t, tenantID, tracking.TenantID())
	assert.Equal(t, projectID, tracking.ProjectID())
	assert.Equal(t, SourceMetaAds, tracking.Source())
	assert.Equal(t, PlatformInstagram, tracking.Platform())
	assert.Equal(t, "Summer Campaign", tracking.Campaign())
	assert.Equal(t, "ad_123", tracking.AdID())
	assert.Equal(t, "https://instagram.com/p/123", tracking.AdURL())
	assert.Equal(t, "ctwa_click_456", tracking.ClickID())
	assert.Equal(t, "encrypted_data", tracking.ConversionData())
	assert.Equal(t, "instagram", tracking.UTMSource())
	assert.Equal(t, "paid_social", tracking.UTMMedium())
	assert.Equal(t, "summer_sale", tracking.UTMCampaign())
	assert.Equal(t, "shoes", tracking.UTMTerm())
	assert.Equal(t, "carousel", tracking.UTMContent())
	assert.Equal(t, metadata, tracking.Metadata())
	assert.Equal(t, now, tracking.CreatedAt())
	assert.Equal(t, now, tracking.UpdatedAt())

	// Events should be empty for reconstructed entities
	events := tracking.DomainEvents()
	assert.Len(t, events, 0)
}

func TestTracking_AllSources(t *testing.T) {
	sources := []Source{
		SourceMetaAds,
		SourceGoogleAds,
		SourceTikTokAds,
		SourceLinkedIn,
		SourceOrganic,
		SourceDirect,
		SourceReferral,
		SourceOther,
	}

	for _, source := range sources {
		t.Run(string(source), func(t *testing.T) {
			tracking, err := NewTracking(
				uuid.New(),
				nil,
				"tenant-123",
				uuid.New(),
				source,
				PlatformOther,
			)

			require.NoError(t, err)
			assert.Equal(t, source, tracking.Source())
		})
	}
}

func TestTracking_AllPlatforms(t *testing.T) {
	platforms := []Platform{
		PlatformInstagram,
		PlatformFacebook,
		PlatformGoogle,
		PlatformTikTok,
		PlatformLinkedIn,
		PlatformWhatsApp,
		PlatformOther,
	}

	for _, platform := range platforms {
		t.Run(string(platform), func(t *testing.T) {
			tracking, err := NewTracking(
				uuid.New(),
				nil,
				"tenant-123",
				uuid.New(),
				SourceOrganic,
				platform,
			)

			require.NoError(t, err)
			assert.Equal(t, platform, tracking.Platform())
		})
	}
}

// Helper functions
func createTestTracking(t *testing.T) *Tracking {
	contactID := uuid.New()
	sessionID := uuid.New()
	tenantID := "tenant-123"
	projectID := uuid.New()

	tracking, err := NewTracking(
		contactID,
		&sessionID,
		tenantID,
		projectID,
		SourceMetaAds,
		PlatformInstagram,
	)

	require.NoError(t, err)
	return tracking
}
