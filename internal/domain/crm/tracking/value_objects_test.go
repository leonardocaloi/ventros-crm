package tracking

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUTMStandard_Validate(t *testing.T) {
	tests := []struct {
		name    string
		utm     UTMStandard
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid meta utm",
			utm: UTMStandard{
				SourcePlatform: UTMPlatformMeta,
				Source:         string(MetaInstagram),
				Medium:         MediumPaidSocial,
				Campaign:       "summer-sale",
			},
			wantErr: false,
		},
		{
			name: "valid google utm",
			utm: UTMStandard{
				SourcePlatform: UTMPlatformGoogle,
				Source:         string(GoogleSearch),
				Medium:         MediumPaidSearch,
				Campaign:       "brand-keywords",
			},
			wantErr: false,
		},
		{
			name: "missing source_platform",
			utm: UTMStandard{
				Source:   string(MetaInstagram),
				Medium:   MediumPaidSocial,
				Campaign: "test",
			},
			wantErr: true,
			errMsg:  "utm_source_platform is required",
		},
		{
			name: "missing source",
			utm: UTMStandard{
				SourcePlatform: UTMPlatformMeta,
				Medium:         MediumPaidSocial,
				Campaign:       "test",
			},
			wantErr: true,
			errMsg:  "utm_source is required",
		},
		{
			name: "missing medium",
			utm: UTMStandard{
				SourcePlatform: UTMPlatformMeta,
				Source:         string(MetaInstagram),
				Campaign:       "test",
			},
			wantErr: true,
			errMsg:  "utm_medium is required",
		},
		{
			name: "missing campaign",
			utm: UTMStandard{
				SourcePlatform: UTMPlatformMeta,
				Source:         string(MetaInstagram),
				Medium:         MediumPaidSocial,
			},
			wantErr: true,
			errMsg:  "utm_campaign is required",
		},
		{
			name: "invalid medium for meta",
			utm: UTMStandard{
				SourcePlatform: UTMPlatformMeta,
				Source:         string(MetaInstagram),
				Medium:         MediumEmail,
				Campaign:       "test",
			},
			wantErr: true,
			errMsg:  "should use paid-social or organic-social medium",
		},
		{
			name: "invalid medium for google",
			utm: UTMStandard{
				SourcePlatform: UTMPlatformGoogle,
				Source:         string(GoogleSearch),
				Medium:         MediumPaidSocial,
				Campaign:       "test",
			},
			wantErr: true,
			errMsg:  "should use search, display or video mediums",
		},
		{
			name: "invalid medium for mkt-direto",
			utm: UTMStandard{
				SourcePlatform: PlatformMktDireto,
				Source:         string(MktDisparo),
				Medium:         MediumPaidSocial,
				Campaign:       "test",
			},
			wantErr: true,
			errMsg:  "should use email, sms or whatsapp medium",
		},
		{
			name: "invalid medium for offline",
			utm: UTMStandard{
				SourcePlatform: UTMPlatformOffline,
				Source:         string(OfflineTV),
				Medium:         MediumPaidSocial,
				Campaign:       "test",
			},
			wantErr: true,
			errMsg:  "should use offline medium",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.utm.Validate()

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidatePlatformMediumCompatibility(t *testing.T) {
	tests := []struct {
		name     string
		platform UTMSourcePlatform
		medium   UTMMedium
		wantErr  bool
	}{
		// Meta
		{"meta + paid-social", UTMPlatformMeta, MediumPaidSocial, false},
		{"meta + organic-social", UTMPlatformMeta, MediumOrganicSocial, false},
		{"meta + email", UTMPlatformMeta, MediumEmail, true},

		// Google
		{"google + paid-search", UTMPlatformGoogle, MediumPaidSearch, false},
		{"google + organic-search", UTMPlatformGoogle, MediumOrganicSearch, false},
		{"google + display", UTMPlatformGoogle, MediumDisplay, false},
		{"google + video", UTMPlatformGoogle, MediumVideo, false},
		{"google + paid-social", UTMPlatformGoogle, MediumPaidSocial, true},

		// TikTok
		{"tiktok + paid-social", UTMPlatformTikTok, MediumPaidSocial, false},
		{"tiktok + email", UTMPlatformTikTok, MediumEmail, true},

		// LinkedIn
		{"linkedin + paid-social", UTMPlatformLinkedIn, MediumPaidSocial, false},
		{"linkedin + display", UTMPlatformLinkedIn, MediumDisplay, true},

		// Mkt Direto
		{"mkt-direto + email", PlatformMktDireto, MediumEmail, false},
		{"mkt-direto + sms", PlatformMktDireto, MediumSMS, false},
		{"mkt-direto + whatsapp", PlatformMktDireto, MediumWhatsApp, false},
		{"mkt-direto + paid-social", PlatformMktDireto, MediumPaidSocial, true},

		// Offline
		{"offline + offline", UTMPlatformOffline, MediumOffline, false},
		{"offline + email", UTMPlatformOffline, MediumEmail, true},

		// Other (no validation)
		{"other + any medium", UTMPlatformOther, MediumOther, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			utm := &UTMStandard{
				SourcePlatform: tt.platform,
				Source:         "test",
				Medium:         tt.medium,
				Campaign:       "test",
			}

			err := utm.validatePlatformMediumCompatibility()

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetValidSourcesForPlatform(t *testing.T) {
	tests := []struct {
		name            string
		platform        UTMSourcePlatform
		expectedSources []string
		minCount        int
	}{
		{
			name:            "meta sources",
			platform:        UTMPlatformMeta,
			expectedSources: []string{"facebook", "instagram", "messenger", "audience-network"},
			minCount:        4,
		},
		{
			name:            "google sources",
			platform:        UTMPlatformGoogle,
			expectedSources: []string{"search", "display", "youtube", "gmail"},
			minCount:        4,
		},
		{
			name:            "mkt-direto sources",
			platform:        PlatformMktDireto,
			expectedSources: []string{"influencer", "disparo", "affiliate"},
			minCount:        3,
		},
		{
			name:            "offline sources",
			platform:        UTMPlatformOffline,
			expectedSources: []string{"tv", "impresso", "outdoor", "evento"},
			minCount:        4,
		},
		{
			name:            "tiktok sources",
			platform:        UTMPlatformTikTok,
			expectedSources: []string{"tiktok"},
			minCount:        1,
		},
		{
			name:            "linkedin sources",
			platform:        UTMPlatformLinkedIn,
			expectedSources: []string{"linkedin"},
			minCount:        1,
		},
		{
			name:            "unknown platform",
			platform:        "unknown",
			expectedSources: []string{},
			minCount:        0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sources := GetValidSourcesForPlatform(tt.platform)

			assert.Len(t, sources, tt.minCount)

			// Check that all expected sources are present
			for _, expectedSource := range tt.expectedSources {
				assert.Contains(t, sources, expectedSource)
			}
		})
	}
}

func TestGetValidMediumsForPlatform(t *testing.T) {
	tests := []struct {
		name     string
		platform UTMSourcePlatform
		expected []UTMMedium
	}{
		{
			name:     "meta mediums",
			platform: UTMPlatformMeta,
			expected: []UTMMedium{MediumPaidSocial, MediumOrganicSocial},
		},
		{
			name:     "google mediums",
			platform: UTMPlatformGoogle,
			expected: []UTMMedium{MediumPaidSearch, MediumOrganicSearch, MediumDisplay, MediumVideo},
		},
		{
			name:     "mkt-direto mediums",
			platform: PlatformMktDireto,
			expected: []UTMMedium{MediumEmail, MediumSMS, MediumWhatsApp},
		},
		{
			name:     "offline mediums",
			platform: UTMPlatformOffline,
			expected: []UTMMedium{MediumOffline},
		},
		{
			name:     "tiktok mediums",
			platform: UTMPlatformTikTok,
			expected: []UTMMedium{MediumPaidSocial, MediumOrganicSocial},
		},
		{
			name:     "linkedin mediums",
			platform: UTMPlatformLinkedIn,
			expected: []UTMMedium{MediumPaidSocial, MediumOrganicSocial},
		},
		{
			name:     "unknown platform",
			platform: "unknown",
			expected: []UTMMedium{MediumOther},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mediums := GetValidMediumsForPlatform(tt.platform)

			assert.Equal(t, len(tt.expected), len(mediums))

			for _, expectedMedium := range tt.expected {
				assert.Contains(t, mediums, expectedMedium)
			}
		})
	}
}

func TestIsValidSource(t *testing.T) {
	tests := []struct {
		name     string
		platform UTMSourcePlatform
		source   string
		expected bool
	}{
		{"meta + instagram", UTMPlatformMeta, "instagram", true},
		{"meta + facebook", UTMPlatformMeta, "facebook", true},
		{"meta + invalid", UTMPlatformMeta, "invalid", false},
		{"google + search", UTMPlatformGoogle, "search", true},
		{"google + youtube", UTMPlatformGoogle, "youtube", true},
		{"google + invalid", UTMPlatformGoogle, "invalid", false},
		{"mkt-direto + influencer", PlatformMktDireto, "influencer", true},
		{"mkt-direto + invalid", PlatformMktDireto, "invalid", false},
		{"offline + tv", UTMPlatformOffline, "tv", true},
		{"offline + invalid", UTMPlatformOffline, "invalid", false},
		{"tiktok + tiktok", UTMPlatformTikTok, "tiktok", true},
		{"tiktok + invalid", UTMPlatformTikTok, "invalid", false},
		{"linkedin + linkedin", UTMPlatformLinkedIn, "linkedin", true},
		{"linkedin + invalid", UTMPlatformLinkedIn, "invalid", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidSource(tt.platform, tt.source)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsValidMedium(t *testing.T) {
	tests := []struct {
		name     string
		platform UTMSourcePlatform
		medium   UTMMedium
		expected bool
	}{
		{"meta + paid-social", UTMPlatformMeta, MediumPaidSocial, true},
		{"meta + organic-social", UTMPlatformMeta, MediumOrganicSocial, true},
		{"meta + email", UTMPlatformMeta, MediumEmail, false},
		{"google + paid-search", UTMPlatformGoogle, MediumPaidSearch, true},
		{"google + display", UTMPlatformGoogle, MediumDisplay, true},
		{"google + video", UTMPlatformGoogle, MediumVideo, true},
		{"google + email", UTMPlatformGoogle, MediumEmail, false},
		{"mkt-direto + email", PlatformMktDireto, MediumEmail, true},
		{"mkt-direto + sms", PlatformMktDireto, MediumSMS, true},
		{"mkt-direto + whatsapp", PlatformMktDireto, MediumWhatsApp, true},
		{"mkt-direto + paid-social", PlatformMktDireto, MediumPaidSocial, false},
		{"offline + offline", UTMPlatformOffline, MediumOffline, true},
		{"offline + email", UTMPlatformOffline, MediumEmail, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidMedium(tt.platform, tt.medium)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestUTMPlatformConstants(t *testing.T) {
	// Verify platform constants
	assert.Equal(t, UTMSourcePlatform("mkt-direto"), PlatformMktDireto)
	assert.Equal(t, UTMSourcePlatform("meta"), UTMPlatformMeta)
	assert.Equal(t, UTMSourcePlatform("google"), UTMPlatformGoogle)
	assert.Equal(t, UTMSourcePlatform("tiktok"), UTMPlatformTikTok)
	assert.Equal(t, UTMSourcePlatform("linkedin"), UTMPlatformLinkedIn)
	assert.Equal(t, UTMSourcePlatform("offline"), UTMPlatformOffline)
	assert.Equal(t, UTMSourcePlatform("other"), UTMPlatformOther)
}

func TestUTMMediumConstants(t *testing.T) {
	// Verify medium constants
	assert.Equal(t, UTMMedium("paid-social"), MediumPaidSocial)
	assert.Equal(t, UTMMedium("organic-social"), MediumOrganicSocial)
	assert.Equal(t, UTMMedium("paid-search"), MediumPaidSearch)
	assert.Equal(t, UTMMedium("organic"), MediumOrganicSearch)
	assert.Equal(t, UTMMedium("display"), MediumDisplay)
	assert.Equal(t, UTMMedium("video"), MediumVideo)
	assert.Equal(t, UTMMedium("email"), MediumEmail)
	assert.Equal(t, UTMMedium("sms"), MediumSMS)
	assert.Equal(t, UTMMedium("whatsapp"), MediumWhatsApp)
	assert.Equal(t, UTMMedium("direct"), MediumDirect)
	assert.Equal(t, UTMMedium("offline"), MediumOffline)
	assert.Equal(t, UTMMedium("referral"), MediumReferral)
	assert.Equal(t, UTMMedium("other"), MediumOther)
}

func TestUTMStandard_WithOptionalFields(t *testing.T) {
	utm := UTMStandard{
		SourcePlatform:  UTMPlatformMeta,
		Source:          string(MetaInstagram),
		Medium:          MediumPaidSocial,
		Campaign:        "summer-sale",
		MarketingTactic: TacticRemarketing,
		Term:            "dermatologia",
		Content:         "video-acne-treatment",
		CreativeFormat:  FormatVideo,
	}

	err := utm.Validate()
	assert.NoError(t, err)

	// Optional fields should be preserved
	assert.Equal(t, TacticRemarketing, utm.MarketingTactic)
	assert.Equal(t, "dermatologia", utm.Term)
	assert.Equal(t, "video-acne-treatment", utm.Content)
	assert.Equal(t, FormatVideo, utm.CreativeFormat)
}

func TestNewTrackingBuilder(t *testing.T) {
	builder := NewTrackingBuilder()

	assert.NotNil(t, builder)
	assert.NotNil(t, builder.utm)
	assert.NotNil(t, builder.metadata)
	assert.NotNil(t, builder.errors)
	assert.Len(t, builder.errors, 0)
}

func TestTrackingBuilder_WithContact(t *testing.T) {
	builder := NewTrackingBuilder()

	result := builder.WithContact("contact-123", "tenant-123", "project-123")

	assert.Equal(t, builder, result) // Fluent interface
	assert.Equal(t, "contact-123", builder.contactID)
	assert.Equal(t, "tenant-123", builder.tenantID)
	assert.Equal(t, "project-123", builder.projectID)
}

func TestTrackingBuilder_WithSession(t *testing.T) {
	builder := NewTrackingBuilder()

	result := builder.WithSession("session-123")

	assert.Equal(t, builder, result)
	assert.NotNil(t, builder.sessionID)
	assert.Equal(t, "session-123", *builder.sessionID)
}

func TestTrackingBuilder_WithSourcePlatform(t *testing.T) {
	tests := []struct {
		name     string
		platform UTMSourcePlatform
		wantErr  bool
	}{
		{"valid - meta", UTMPlatformMeta, false},
		{"valid - google", UTMPlatformGoogle, false},
		{"valid - tiktok", UTMPlatformTikTok, false},
		{"valid - linkedin", UTMPlatformLinkedIn, false},
		{"valid - mkt-direto", PlatformMktDireto, false},
		{"valid - offline", UTMPlatformOffline, false},
		{"valid - other", UTMPlatformOther, false},
		{"invalid platform", "invalid", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewTrackingBuilder()
			result := builder.WithSourcePlatform(tt.platform)

			assert.Equal(t, builder, result)
			assert.Equal(t, tt.platform, builder.utm.SourcePlatform)

			if tt.wantErr {
				assert.NotEmpty(t, builder.errors)
			} else {
				assert.Empty(t, builder.errors)
			}
		})
	}
}

func TestTrackingBuilder_WithSource(t *testing.T) {
	t.Run("valid source for meta", func(t *testing.T) {
		builder := NewTrackingBuilder().
			WithSourcePlatform(UTMPlatformMeta).
			WithSource(string(MetaInstagram))

		assert.Equal(t, string(MetaInstagram), builder.utm.Source)
		assert.Empty(t, builder.errors)
	})

	t.Run("invalid source for meta", func(t *testing.T) {
		builder := NewTrackingBuilder().
			WithSourcePlatform(UTMPlatformMeta).
			WithSource("invalid-source")

		assert.NotEmpty(t, builder.errors)
	})

	t.Run("source before platform", func(t *testing.T) {
		builder := NewTrackingBuilder().
			WithSource(string(MetaInstagram))

		assert.NotEmpty(t, builder.errors)
		assert.Contains(t, builder.errors[0].Error(), "must set source_platform before source")
	})
}

func TestTrackingBuilder_WithMedium(t *testing.T) {
	t.Run("valid medium for meta", func(t *testing.T) {
		builder := NewTrackingBuilder().
			WithSourcePlatform(UTMPlatformMeta).
			WithMedium(MediumPaidSocial)

		assert.Equal(t, MediumPaidSocial, builder.utm.Medium)
		assert.Empty(t, builder.errors)
	})

	t.Run("invalid medium for meta", func(t *testing.T) {
		builder := NewTrackingBuilder().
			WithSourcePlatform(UTMPlatformMeta).
			WithMedium(MediumEmail)

		assert.NotEmpty(t, builder.errors)
	})

	t.Run("medium before platform", func(t *testing.T) {
		builder := NewTrackingBuilder().
			WithMedium(MediumPaidSocial)

		assert.NotEmpty(t, builder.errors)
		assert.Contains(t, builder.errors[0].Error(), "must set source_platform before medium")
	})
}

func TestTrackingBuilder_WithCampaign(t *testing.T) {
	t.Run("valid campaign", func(t *testing.T) {
		builder := NewTrackingBuilder().
			WithCampaign("summer-sale-2025")

		assert.Equal(t, "summer-sale-2025", builder.utm.Campaign)
		assert.Empty(t, builder.errors)
	})

	t.Run("empty campaign", func(t *testing.T) {
		builder := NewTrackingBuilder().
			WithCampaign("")

		assert.NotEmpty(t, builder.errors)
		assert.Contains(t, builder.errors[0].Error(), "campaign cannot be empty")
	})
}

func TestTrackingBuilder_OptionalFields(t *testing.T) {
	builder := NewTrackingBuilder().
		WithMarketingTactic(TacticRemarketing).
		WithTerm("dermatologia").
		WithContent("video-acne").
		WithCreativeFormat(FormatVideo)

	assert.Equal(t, TacticRemarketing, builder.utm.MarketingTactic)
	assert.Equal(t, "dermatologia", builder.utm.Term)
	assert.Equal(t, "video-acne", builder.utm.Content)
	assert.Equal(t, FormatVideo, builder.utm.CreativeFormat)
}

func TestTrackingBuilder_WithAdID(t *testing.T) {
	t.Run("adID sets content if empty", func(t *testing.T) {
		builder := NewTrackingBuilder().
			WithAdID("12345")

		assert.Equal(t, "12345", builder.adID)
		assert.Equal(t, "ad_id_12345", builder.utm.Content)
	})

	t.Run("adID does not override existing content", func(t *testing.T) {
		builder := NewTrackingBuilder().
			WithContent("existing-content").
			WithAdID("12345")

		assert.Equal(t, "12345", builder.adID)
		assert.Equal(t, "existing-content", builder.utm.Content)
	})
}

func TestTrackingBuilder_WithClickID(t *testing.T) {
	builder := NewTrackingBuilder().
		WithClickID("click-123")

	assert.Equal(t, "click-123", builder.clickID)
}

func TestTrackingBuilder_WithMetadata(t *testing.T) {
	builder := NewTrackingBuilder().
		WithMetadata("key1", "value1").
		WithMetadata("key2", 123)

	assert.Equal(t, "value1", builder.metadata["key1"])
	assert.Equal(t, 123, builder.metadata["key2"])
}

func TestTrackingBuilder_Validate(t *testing.T) {
	t.Run("valid builder", func(t *testing.T) {
		builder := NewTrackingBuilder().
			WithContact("contact-123", "tenant-123", "project-123").
			WithSourcePlatform(UTMPlatformMeta).
			WithSource(string(MetaInstagram)).
			WithMedium(MediumPaidSocial).
			WithCampaign("test-campaign")

		err := builder.Validate()
		assert.NoError(t, err)
	})

	t.Run("missing contactID", func(t *testing.T) {
		builder := NewTrackingBuilder().
			WithContact("", "tenant-123", "project-123").
			WithSourcePlatform(UTMPlatformMeta).
			WithSource(string(MetaInstagram)).
			WithMedium(MediumPaidSocial).
			WithCampaign("test-campaign")

		err := builder.Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "contactID is required")
	})

	t.Run("missing tenantID", func(t *testing.T) {
		builder := NewTrackingBuilder().
			WithContact("contact-123", "", "project-123").
			WithSourcePlatform(UTMPlatformMeta).
			WithSource(string(MetaInstagram)).
			WithMedium(MediumPaidSocial).
			WithCampaign("test-campaign")

		err := builder.Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "tenantID is required")
	})

	t.Run("missing projectID", func(t *testing.T) {
		builder := NewTrackingBuilder().
			WithContact("contact-123", "tenant-123", "").
			WithSourcePlatform(UTMPlatformMeta).
			WithSource(string(MetaInstagram)).
			WithMedium(MediumPaidSocial).
			WithCampaign("test-campaign")

		err := builder.Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "projectID is required")
	})

	t.Run("multiple validation errors", func(t *testing.T) {
		builder := NewTrackingBuilder().
			WithCampaign("") // Empty campaign

		err := builder.Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "validation errors")
	})
}

func TestTrackingBuilder_Build(t *testing.T) {
	t.Run("successful build", func(t *testing.T) {
		builder := NewTrackingBuilder().
			WithContact("contact-123", "tenant-123", "project-123").
			WithSession("session-123").
			WithSourcePlatform(UTMPlatformMeta).
			WithSource(string(MetaInstagram)).
			WithMedium(MediumPaidSocial).
			WithCampaign("summer-sale").
			WithAdID("ad-123").
			WithClickID("click-456").
			WithMetadata("custom_key", "custom_value")

		utm, metadata, err := builder.Build()
		require.NoError(t, err)
		require.NotNil(t, utm)
		require.NotNil(t, metadata)

		// Check UTM
		assert.Equal(t, UTMPlatformMeta, utm.SourcePlatform)
		assert.Equal(t, string(MetaInstagram), utm.Source)
		assert.Equal(t, MediumPaidSocial, utm.Medium)
		assert.Equal(t, "summer-sale", utm.Campaign)

		// Check metadata
		assert.Equal(t, "contact-123", metadata["contact_id"])
		assert.Equal(t, "tenant-123", metadata["tenant_id"])
		assert.Equal(t, "project-123", metadata["project_id"])
		assert.Equal(t, "session-123", metadata["session_id"])
		assert.Equal(t, "ad-123", metadata["ad_id"])
		assert.Equal(t, "click-456", metadata["click_id"])
		assert.Equal(t, "custom_value", metadata["custom_key"])
	})

	t.Run("build without session", func(t *testing.T) {
		builder := NewTrackingBuilder().
			WithContact("contact-123", "tenant-123", "project-123").
			WithSourcePlatform(UTMPlatformMeta).
			WithSource(string(MetaInstagram)).
			WithMedium(MediumPaidSocial).
			WithCampaign("test")

		_, metadata, err := builder.Build()
		require.NoError(t, err)

		_, hasSession := metadata["session_id"]
		assert.False(t, hasSession)
	})

	t.Run("build fails validation", func(t *testing.T) {
		builder := NewTrackingBuilder()

		utm, metadata, err := builder.Build()
		require.Error(t, err)
		assert.Nil(t, utm)
		assert.Nil(t, metadata)
	})
}

func TestTrackingBuilder_BuildURL(t *testing.T) {
	t.Run("successful URL build", func(t *testing.T) {
		builder := NewTrackingBuilder().
			WithContact("contact-123", "tenant-123", "project-123").
			WithSourcePlatform(UTMPlatformMeta).
			WithSource(string(MetaInstagram)).
			WithMedium(MediumPaidSocial).
			WithCampaign("summer-sale").
			WithTerm("dermatologia").
			WithContent("video-1").
			WithMarketingTactic(TacticProspecting).
			WithCreativeFormat(FormatVideo)

		url, err := builder.BuildURL("https://example.com/landing")
		require.NoError(t, err)

		assert.Contains(t, url, "https://example.com/landing?")
		assert.Contains(t, url, "utm_source_platform=meta")
		assert.Contains(t, url, "utm_source=instagram")
		assert.Contains(t, url, "utm_medium=paid-social")
		assert.Contains(t, url, "utm_campaign=summer-sale")
		assert.Contains(t, url, "utm_term=dermatologia")
		assert.Contains(t, url, "utm_content=video-1")
		assert.Contains(t, url, "utm_marketing_tactic=prospecting")
		assert.Contains(t, url, "utm_creative_format=video")
	})

	t.Run("URL with existing query params", func(t *testing.T) {
		builder := NewTrackingBuilder().
			WithContact("contact-123", "tenant-123", "project-123").
			WithSourcePlatform(UTMPlatformMeta).
			WithSource(string(MetaInstagram)).
			WithMedium(MediumPaidSocial).
			WithCampaign("test")

		url, err := builder.BuildURL("https://example.com/page?existing=param")
		require.NoError(t, err)

		assert.Contains(t, url, "existing=param")
		assert.Contains(t, url, "&utm_source_platform=meta")
	})

	t.Run("empty base URL", func(t *testing.T) {
		builder := NewTrackingBuilder().
			WithContact("contact-123", "tenant-123", "project-123").
			WithSourcePlatform(UTMPlatformMeta).
			WithSource(string(MetaInstagram)).
			WithMedium(MediumPaidSocial).
			WithCampaign("test")

		url, err := builder.BuildURL("")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "base URL cannot be empty")
		assert.Empty(t, url)
	})

	t.Run("validation fails before building URL", func(t *testing.T) {
		builder := NewTrackingBuilder()

		url, err := builder.BuildURL("https://example.com")
		require.Error(t, err)
		assert.Empty(t, url)
	})
}

func TestContains(t *testing.T) {
	tests := []struct {
		s      string
		substr string
		want   bool
	}{
		{"https://example.com?param=value", "?", true},
		{"https://example.com", "?", false},
		{"hello world", "world", true},
		{"hello world", "xyz", false},
		{"", "test", false},
		{"test", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.s+" contains "+tt.substr, func(t *testing.T) {
			result := contains(tt.s, tt.substr)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestFindSubstring(t *testing.T) {
	tests := []struct {
		s      string
		substr string
		want   bool
	}{
		{"hello world", "world", true},
		{"hello world", "hello", true},
		{"hello world", "lo wo", true},
		{"hello world", "xyz", false},
		{"test", "testing", false},
	}

	for _, tt := range tests {
		t.Run(tt.s+" contains "+tt.substr, func(t *testing.T) {
			result := findSubstring(tt.s, tt.substr)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestTrackingBuilder_FluentInterface(t *testing.T) {
	// Test that all methods return the builder for method chaining
	builder := NewTrackingBuilder().
		WithContact("contact", "tenant", "project").
		WithSession("session").
		WithSourcePlatform(UTMPlatformGoogle).
		WithSource(string(GoogleSearch)).
		WithMedium(MediumPaidSearch).
		WithCampaign("campaign").
		WithMarketingTactic(TacticProspecting).
		WithTerm("term").
		WithContent("content").
		WithCreativeFormat(FormatBannerEstatico).
		WithAdID("ad").
		WithClickID("click").
		WithMetadata("key", "value")

	assert.NotNil(t, builder)
	err := builder.Validate()
	assert.NoError(t, err)
}
