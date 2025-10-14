package tracking

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ventros/crm/internal/domain/crm/tracking"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap/zaptest"
)

// ========== Tests for GetTrackingUseCase ==========

// Helper function to create test timestamps
func testTime() time.Time {
	return time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
}

func TestNewGetTrackingUseCase(t *testing.T) {
	// Arrange
	repo := new(MockTrackingRepository)
	logger := zaptest.NewLogger(t)

	// Act
	useCase := NewGetTrackingUseCase(repo, logger)

	// Assert
	assert.NotNil(t, useCase)
	assert.Equal(t, repo, useCase.repo)
	assert.Equal(t, logger, useCase.logger)
}

func TestGetTrackingUseCase_Execute_Success(t *testing.T) {
	// Arrange
	ctx := context.Background()
	repo := new(MockTrackingRepository)
	logger := zaptest.NewLogger(t)

	useCase := NewGetTrackingUseCase(repo, logger)

	trackingID := uuid.New()
	contactID := uuid.New()
	sessionID := uuid.New()
	projectID := uuid.New()
	tenantID := "tenant-123"

	// Create a sample tracking entity
	sampleTracking := tracking.ReconstructTracking(
		trackingID,
		contactID,
		&sessionID,
		tenantID,
		projectID,
		tracking.SourceMetaAds,
		tracking.PlatformInstagram,
		"summer-sale-2025",
		"ad-12345",
		"https://example.com/ad",
		"click-67890",
		`{"conversion_id": "conv-123"}`,
		"google",
		"cpc",
		"summer-sale",
		"shoes",
		"banner-top",
		map[string]interface{}{
			"key1": "value1",
			"key2": 123,
		},
		testTime(),
		testTime(),
	)

	// Mock repository FindByID
	repo.On("FindByID", ctx, trackingID).Return(sampleTracking, nil)

	// Act
	result, err := useCase.Execute(ctx, trackingID)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, trackingID, result.ID)
	assert.Equal(t, contactID, result.ContactID)
	assert.Equal(t, &sessionID, result.SessionID)
	assert.Equal(t, projectID, result.ProjectID)
	assert.Equal(t, tenantID, result.TenantID)
	assert.Equal(t, string(tracking.SourceMetaAds), result.Source)
	assert.Equal(t, string(tracking.PlatformInstagram), result.Platform)
	assert.Equal(t, "summer-sale-2025", result.Campaign)
	assert.Equal(t, "ad-12345", result.AdID)
	assert.Equal(t, "https://example.com/ad", result.AdURL)
	assert.Equal(t, "click-67890", result.ClickID)
	assert.Equal(t, `{"conversion_id": "conv-123"}`, result.ConversionData)
	assert.Equal(t, "google", result.UTMSource)
	assert.Equal(t, "cpc", result.UTMMedium)
	assert.Equal(t, "summer-sale", result.UTMCampaign)
	assert.Equal(t, "shoes", result.UTMTerm)
	assert.Equal(t, "banner-top", result.UTMContent)
	assert.NotNil(t, result.Metadata)
	assert.Equal(t, "value1", result.Metadata["key1"])
	assert.Equal(t, 123, result.Metadata["key2"])

	repo.AssertExpectations(t)
}

func TestGetTrackingUseCase_Execute_Success_MinimalFields(t *testing.T) {
	// Arrange
	ctx := context.Background()
	repo := new(MockTrackingRepository)
	logger := zaptest.NewLogger(t)

	useCase := NewGetTrackingUseCase(repo, logger)

	trackingID := uuid.New()
	contactID := uuid.New()
	projectID := uuid.New()
	tenantID := "tenant-123"

	// Create a sample tracking entity with minimal fields
	sampleTracking := tracking.ReconstructTracking(
		trackingID,
		contactID,
		nil, // no session ID
		tenantID,
		projectID,
		tracking.SourceOrganic,
		tracking.PlatformWhatsApp,
		"", // no campaign
		"", // no ad ID
		"", // no ad URL
		"", // no click ID
		"", // no conversion data
		"", // no UTM source
		"", // no UTM medium
		"", // no UTM campaign
		"", // no UTM term
		"", // no UTM content
		map[string]interface{}{}, // empty metadata
		testTime(),
		testTime(),
	)

	// Mock repository FindByID
	repo.On("FindByID", ctx, trackingID).Return(sampleTracking, nil)

	// Act
	result, err := useCase.Execute(ctx, trackingID)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, trackingID, result.ID)
	assert.Equal(t, contactID, result.ContactID)
	assert.Nil(t, result.SessionID)
	assert.Equal(t, projectID, result.ProjectID)
	assert.Equal(t, tenantID, result.TenantID)
	assert.Equal(t, string(tracking.SourceOrganic), result.Source)
	assert.Equal(t, string(tracking.PlatformWhatsApp), result.Platform)
	assert.Empty(t, result.Campaign)
	assert.Empty(t, result.AdID)
	assert.Empty(t, result.AdURL)
	assert.Empty(t, result.ClickID)
	assert.Empty(t, result.ConversionData)
	assert.Empty(t, result.UTMSource)
	assert.Empty(t, result.UTMMedium)
	assert.Empty(t, result.UTMCampaign)
	assert.Empty(t, result.UTMTerm)
	assert.Empty(t, result.UTMContent)

	repo.AssertExpectations(t)
}

func TestGetTrackingUseCase_Execute_TrackingNotFound(t *testing.T) {
	// Arrange
	ctx := context.Background()
	repo := new(MockTrackingRepository)
	logger := zaptest.NewLogger(t)

	useCase := NewGetTrackingUseCase(repo, logger)

	trackingID := uuid.New()

	// Mock repository FindByID - return ErrTrackingNotFound
	repo.On("FindByID", ctx, trackingID).Return(nil, tracking.ErrTrackingNotFound)

	// Act
	result, err := useCase.Execute(ctx, trackingID)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, tracking.ErrTrackingNotFound, err)

	repo.AssertExpectations(t)
}

func TestGetTrackingUseCase_Execute_RepositoryError(t *testing.T) {
	// Arrange
	ctx := context.Background()
	repo := new(MockTrackingRepository)
	logger := zaptest.NewLogger(t)

	useCase := NewGetTrackingUseCase(repo, logger)

	trackingID := uuid.New()
	expectedError := errors.New("database connection error")

	// Mock repository FindByID - return generic error
	repo.On("FindByID", ctx, trackingID).Return(nil, expectedError)

	// Act
	result, err := useCase.Execute(ctx, trackingID)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to fetch tracking")
	assert.ErrorIs(t, err, expectedError)

	repo.AssertExpectations(t)
}

func TestGetTrackingUseCase_Execute_ValidatesDTOConversion(t *testing.T) {
	// Arrange
	ctx := context.Background()
	repo := new(MockTrackingRepository)
	logger := zaptest.NewLogger(t)

	useCase := NewGetTrackingUseCase(repo, logger)

	trackingID := uuid.New()
	contactID := uuid.New()
	projectID := uuid.New()
	tenantID := "tenant-456"

	// Create a tracking with specific values to verify conversion
	sampleTracking := tracking.ReconstructTracking(
		trackingID,
		contactID,
		nil,
		tenantID,
		projectID,
		tracking.SourceGoogleAds,
		tracking.PlatformGoogle,
		"test-campaign",
		"test-ad-id",
		"https://test.com/ad",
		"test-click-id",
		"test-conversion-data",
		"test-utm-source",
		"test-utm-medium",
		"test-utm-campaign",
		"test-utm-term",
		"test-utm-content",
		map[string]interface{}{
			"test_key": "test_value",
		},
		testTime(),
		testTime(),
	)

	// Mock repository FindByID
	repo.On("FindByID", ctx, trackingID).Return(sampleTracking, nil)

	// Act
	result, err := useCase.Execute(ctx, trackingID)

	// Assert - Verify ToDTO() was called correctly by checking all fields
	assert.NoError(t, err)
	assert.NotNil(t, result)

	// Verify all domain entity getters were properly converted to DTO
	assert.Equal(t, sampleTracking.ID(), result.ID)
	assert.Equal(t, sampleTracking.ContactID(), result.ContactID)
	assert.Equal(t, sampleTracking.SessionID(), result.SessionID)
	assert.Equal(t, sampleTracking.TenantID(), result.TenantID)
	assert.Equal(t, sampleTracking.ProjectID(), result.ProjectID)
	assert.Equal(t, string(sampleTracking.Source()), result.Source)
	assert.Equal(t, string(sampleTracking.Platform()), result.Platform)
	assert.Equal(t, sampleTracking.Campaign(), result.Campaign)
	assert.Equal(t, sampleTracking.AdID(), result.AdID)
	assert.Equal(t, sampleTracking.AdURL(), result.AdURL)
	assert.Equal(t, sampleTracking.ClickID(), result.ClickID)
	assert.Equal(t, sampleTracking.ConversionData(), result.ConversionData)
	assert.Equal(t, sampleTracking.UTMSource(), result.UTMSource)
	assert.Equal(t, sampleTracking.UTMMedium(), result.UTMMedium)
	assert.Equal(t, sampleTracking.UTMCampaign(), result.UTMCampaign)
	assert.Equal(t, sampleTracking.UTMTerm(), result.UTMTerm)
	assert.Equal(t, sampleTracking.UTMContent(), result.UTMContent)
	assert.Equal(t, sampleTracking.Metadata(), result.Metadata)
	assert.Equal(t, sampleTracking.CreatedAt(), result.CreatedAt)
	assert.Equal(t, sampleTracking.UpdatedAt(), result.UpdatedAt)

	repo.AssertExpectations(t)
}

func TestGetTrackingUseCase_Execute_AllSources(t *testing.T) {
	testCases := []struct {
		name     string
		source   tracking.Source
		platform tracking.Platform
	}{
		{"MetaAds_Instagram", tracking.SourceMetaAds, tracking.PlatformInstagram},
		{"GoogleAds_Google", tracking.SourceGoogleAds, tracking.PlatformGoogle},
		{"TikTokAds_TikTok", tracking.SourceTikTokAds, tracking.PlatformTikTok},
		{"LinkedIn_LinkedIn", tracking.SourceLinkedIn, tracking.PlatformLinkedIn},
		{"Organic_WhatsApp", tracking.SourceOrganic, tracking.PlatformWhatsApp},
		{"Direct_Facebook", tracking.SourceDirect, tracking.PlatformFacebook},
		{"Referral_Other", tracking.SourceReferral, tracking.PlatformOther},
		{"Other_Other", tracking.SourceOther, tracking.PlatformOther},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			ctx := context.Background()
			repo := new(MockTrackingRepository)
			logger := zaptest.NewLogger(t)

			useCase := NewGetTrackingUseCase(repo, logger)

			trackingID := uuid.New()
			contactID := uuid.New()
			projectID := uuid.New()
			tenantID := "tenant-123"

			sampleTracking := tracking.ReconstructTracking(
				trackingID,
				contactID,
				nil,
				tenantID,
				projectID,
				tc.source,
				tc.platform,
				"",
				"",
				"",
				"",
				"",
				"",
				"",
				"",
				"",
				"",
				nil,
				testTime(),
				testTime(),
			)

			// Mock repository FindByID
			repo.On("FindByID", ctx, trackingID).Return(sampleTracking, nil)

			// Act
			result, err := useCase.Execute(ctx, trackingID)

			// Assert
			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, string(tc.source), result.Source)
			assert.Equal(t, string(tc.platform), result.Platform)

			repo.AssertExpectations(t)
		})
	}
}

func TestGetTrackingUseCase_Execute_NilContext(t *testing.T) {
	// Arrange
	repo := new(MockTrackingRepository)
	logger := zaptest.NewLogger(t)

	useCase := NewGetTrackingUseCase(repo, logger)

	trackingID := uuid.New()

	// Mock repository FindByID - this will be called with nil context
	repo.On("FindByID", mock.Anything, trackingID).Return(nil, errors.New("context is nil"))

	// Act
	result, err := useCase.Execute(nil, trackingID)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)

	repo.AssertExpectations(t)
}

func TestGetTrackingUseCase_Execute_NilUUID(t *testing.T) {
	// Arrange
	ctx := context.Background()
	repo := new(MockTrackingRepository)
	logger := zaptest.NewLogger(t)

	useCase := NewGetTrackingUseCase(repo, logger)

	nilUUID := uuid.Nil

	// Mock repository FindByID - return not found for nil UUID
	repo.On("FindByID", ctx, nilUUID).Return(nil, tracking.ErrTrackingNotFound)

	// Act
	result, err := useCase.Execute(ctx, nilUUID)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, tracking.ErrTrackingNotFound, err)

	repo.AssertExpectations(t)
}

func TestGetTrackingUseCase_Execute_ComplexMetadata(t *testing.T) {
	// Arrange
	ctx := context.Background()
	repo := new(MockTrackingRepository)
	logger := zaptest.NewLogger(t)

	useCase := NewGetTrackingUseCase(repo, logger)

	trackingID := uuid.New()
	contactID := uuid.New()
	projectID := uuid.New()
	tenantID := "tenant-123"

	complexMetadata := map[string]interface{}{
		"string_value":  "test",
		"int_value":     123,
		"float_value":   456.78,
		"bool_value":    true,
		"array_value":   []string{"item1", "item2", "item3"},
		"nested_object": map[string]interface{}{"nested_key": "nested_value"},
	}

	sampleTracking := tracking.ReconstructTracking(
		trackingID,
		contactID,
		nil,
		tenantID,
		projectID,
		tracking.SourceMetaAds,
		tracking.PlatformInstagram,
		"",
		"",
		"",
		"",
		"",
		"",
		"",
		"",
		"",
		"",
		complexMetadata,
		testTime(),
		testTime(),
	)

	// Mock repository FindByID
	repo.On("FindByID", ctx, trackingID).Return(sampleTracking, nil)

	// Act
	result, err := useCase.Execute(ctx, trackingID)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotNil(t, result.Metadata)
	assert.Equal(t, "test", result.Metadata["string_value"])
	assert.Equal(t, 123, result.Metadata["int_value"])
	assert.Equal(t, 456.78, result.Metadata["float_value"])
	assert.Equal(t, true, result.Metadata["bool_value"])

	repo.AssertExpectations(t)
}

func TestGetTrackingUseCase_Execute_WithSessionID(t *testing.T) {
	// Arrange
	ctx := context.Background()
	repo := new(MockTrackingRepository)
	logger := zaptest.NewLogger(t)

	useCase := NewGetTrackingUseCase(repo, logger)

	trackingID := uuid.New()
	contactID := uuid.New()
	sessionID := uuid.New()
	projectID := uuid.New()
	tenantID := "tenant-123"

	sampleTracking := tracking.ReconstructTracking(
		trackingID,
		contactID,
		&sessionID,
		tenantID,
		projectID,
		tracking.SourceMetaAds,
		tracking.PlatformInstagram,
		"",
		"",
		"",
		"",
		"",
		"",
		"",
		"",
		"",
		"",
		nil,
		testTime(),
		testTime(),
	)

	// Mock repository FindByID
	repo.On("FindByID", ctx, trackingID).Return(sampleTracking, nil)

	// Act
	result, err := useCase.Execute(ctx, trackingID)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotNil(t, result.SessionID)
	assert.Equal(t, sessionID, *result.SessionID)

	repo.AssertExpectations(t)
}

func TestGetTrackingUseCase_Execute_WithoutSessionID(t *testing.T) {
	// Arrange
	ctx := context.Background()
	repo := new(MockTrackingRepository)
	logger := zaptest.NewLogger(t)

	useCase := NewGetTrackingUseCase(repo, logger)

	trackingID := uuid.New()
	contactID := uuid.New()
	projectID := uuid.New()
	tenantID := "tenant-123"

	sampleTracking := tracking.ReconstructTracking(
		trackingID,
		contactID,
		nil,
		tenantID,
		projectID,
		tracking.SourceMetaAds,
		tracking.PlatformInstagram,
		"",
		"",
		"",
		"",
		"",
		"",
		"",
		"",
		"",
		"",
		nil,
		testTime(),
		testTime(),
	)

	// Mock repository FindByID
	repo.On("FindByID", ctx, trackingID).Return(sampleTracking, nil)

	// Act
	result, err := useCase.Execute(ctx, trackingID)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Nil(t, result.SessionID)

	repo.AssertExpectations(t)
}
