package tracking

import (
	"context"
	"errors"
	"testing"

	"github.com/caloi/ventros-crm/internal/domain/crm/tracking"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

// ========== Tests for GetContactTrackingsUseCase ==========

func TestNewGetContactTrackingsUseCase(t *testing.T) {
	// Arrange
	repo := new(MockTrackingRepository)
	logger := zaptest.NewLogger(t)

	// Act
	useCase := NewGetContactTrackingsUseCase(repo, logger)

	// Assert
	assert.NotNil(t, useCase)
	assert.Equal(t, repo, useCase.repo)
	assert.Equal(t, logger, useCase.logger)
}

func TestGetContactTrackingsUseCase_Execute_Success_SingleTracking(t *testing.T) {
	// Arrange
	ctx := context.Background()
	repo := new(MockTrackingRepository)
	logger := zaptest.NewLogger(t)

	useCase := NewGetContactTrackingsUseCase(repo, logger)

	contactID := uuid.New()
	trackingID := uuid.New()
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

	trackings := []*tracking.Tracking{sampleTracking}

	// Mock repository FindByContactID
	repo.On("FindByContactID", ctx, contactID).Return(trackings, nil)

	// Act
	result, err := useCase.Execute(ctx, contactID)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 1)
	assert.Equal(t, trackingID, result[0].ID)
	assert.Equal(t, contactID, result[0].ContactID)
	assert.Equal(t, &sessionID, result[0].SessionID)
	assert.Equal(t, projectID, result[0].ProjectID)
	assert.Equal(t, tenantID, result[0].TenantID)
	assert.Equal(t, string(tracking.SourceMetaAds), result[0].Source)
	assert.Equal(t, string(tracking.PlatformInstagram), result[0].Platform)
	assert.Equal(t, "summer-sale-2025", result[0].Campaign)
	assert.Equal(t, "ad-12345", result[0].AdID)
	assert.Equal(t, "https://example.com/ad", result[0].AdURL)
	assert.Equal(t, "click-67890", result[0].ClickID)
	assert.Equal(t, `{"conversion_id": "conv-123"}`, result[0].ConversionData)
	assert.Equal(t, "google", result[0].UTMSource)
	assert.Equal(t, "cpc", result[0].UTMMedium)
	assert.Equal(t, "summer-sale", result[0].UTMCampaign)
	assert.Equal(t, "shoes", result[0].UTMTerm)
	assert.Equal(t, "banner-top", result[0].UTMContent)
	assert.NotNil(t, result[0].Metadata)
	assert.Equal(t, "value1", result[0].Metadata["key1"])
	assert.Equal(t, 123, result[0].Metadata["key2"])

	repo.AssertExpectations(t)
}

func TestGetContactTrackingsUseCase_Execute_Success_MultipleTrackings(t *testing.T) {
	// Arrange
	ctx := context.Background()
	repo := new(MockTrackingRepository)
	logger := zaptest.NewLogger(t)

	useCase := NewGetContactTrackingsUseCase(repo, logger)

	contactID := uuid.New()
	projectID := uuid.New()
	tenantID := "tenant-123"

	// Create multiple tracking entities from different sources
	tracking1 := tracking.ReconstructTracking(
		uuid.New(),
		contactID,
		nil,
		tenantID,
		projectID,
		tracking.SourceMetaAds,
		tracking.PlatformInstagram,
		"campaign-1",
		"ad-1",
		"https://example.com/ad1",
		"click-1",
		"",
		"facebook",
		"social",
		"campaign-1",
		"",
		"",
		map[string]interface{}{"source": "meta"},
		testTime(),
		testTime(),
	)

	tracking2 := tracking.ReconstructTracking(
		uuid.New(),
		contactID,
		nil,
		tenantID,
		projectID,
		tracking.SourceGoogleAds,
		tracking.PlatformGoogle,
		"campaign-2",
		"ad-2",
		"https://example.com/ad2",
		"click-2",
		"",
		"google",
		"cpc",
		"campaign-2",
		"keywords",
		"",
		map[string]interface{}{"source": "google"},
		testTime(),
		testTime(),
	)

	tracking3 := tracking.ReconstructTracking(
		uuid.New(),
		contactID,
		nil,
		tenantID,
		projectID,
		tracking.SourceOrganic,
		tracking.PlatformWhatsApp,
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
		map[string]interface{}{"source": "organic"},
		testTime(),
		testTime(),
	)

	trackings := []*tracking.Tracking{tracking1, tracking2, tracking3}

	// Mock repository FindByContactID
	repo.On("FindByContactID", ctx, contactID).Return(trackings, nil)

	// Act
	result, err := useCase.Execute(ctx, contactID)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 3)

	// Verify first tracking (Meta Ads)
	assert.Equal(t, contactID, result[0].ContactID)
	assert.Equal(t, string(tracking.SourceMetaAds), result[0].Source)
	assert.Equal(t, string(tracking.PlatformInstagram), result[0].Platform)
	assert.Equal(t, "campaign-1", result[0].Campaign)
	assert.Equal(t, "ad-1", result[0].AdID)

	// Verify second tracking (Google Ads)
	assert.Equal(t, contactID, result[1].ContactID)
	assert.Equal(t, string(tracking.SourceGoogleAds), result[1].Source)
	assert.Equal(t, string(tracking.PlatformGoogle), result[1].Platform)
	assert.Equal(t, "campaign-2", result[1].Campaign)
	assert.Equal(t, "ad-2", result[1].AdID)

	// Verify third tracking (Organic)
	assert.Equal(t, contactID, result[2].ContactID)
	assert.Equal(t, string(tracking.SourceOrganic), result[2].Source)
	assert.Equal(t, string(tracking.PlatformWhatsApp), result[2].Platform)
	assert.Empty(t, result[2].Campaign)

	repo.AssertExpectations(t)
}

func TestGetContactTrackingsUseCase_Execute_Success_EmptyList(t *testing.T) {
	// Arrange
	ctx := context.Background()
	repo := new(MockTrackingRepository)
	logger := zaptest.NewLogger(t)

	useCase := NewGetContactTrackingsUseCase(repo, logger)

	contactID := uuid.New()

	// Mock repository FindByContactID - return empty list
	trackings := []*tracking.Tracking{}
	repo.On("FindByContactID", ctx, contactID).Return(trackings, nil)

	// Act
	result, err := useCase.Execute(ctx, contactID)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 0)
	assert.Empty(t, result)

	repo.AssertExpectations(t)
}

func TestGetContactTrackingsUseCase_Execute_RepositoryError(t *testing.T) {
	// Arrange
	ctx := context.Background()
	repo := new(MockTrackingRepository)
	logger := zaptest.NewLogger(t)

	useCase := NewGetContactTrackingsUseCase(repo, logger)

	contactID := uuid.New()
	expectedError := errors.New("database connection error")

	// Mock repository FindByContactID - return error
	repo.On("FindByContactID", ctx, contactID).Return(nil, expectedError)

	// Act
	result, err := useCase.Execute(ctx, contactID)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to fetch trackings")
	assert.ErrorIs(t, err, expectedError)

	repo.AssertExpectations(t)
}

func TestGetContactTrackingsUseCase_Execute_ValidatesDTOListConversion(t *testing.T) {
	// Arrange
	ctx := context.Background()
	repo := new(MockTrackingRepository)
	logger := zaptest.NewLogger(t)

	useCase := NewGetContactTrackingsUseCase(repo, logger)

	contactID := uuid.New()
	projectID := uuid.New()
	tenantID := "tenant-789"

	// Create trackings with specific values to verify conversion
	tracking1 := tracking.ReconstructTracking(
		uuid.New(),
		contactID,
		nil,
		tenantID,
		projectID,
		tracking.SourceMetaAds,
		tracking.PlatformInstagram,
		"test-campaign-1",
		"test-ad-id-1",
		"https://test.com/ad1",
		"test-click-id-1",
		"test-conversion-data-1",
		"test-utm-source-1",
		"test-utm-medium-1",
		"test-utm-campaign-1",
		"test-utm-term-1",
		"test-utm-content-1",
		map[string]interface{}{
			"test_key_1": "test_value_1",
		},
		testTime(),
		testTime(),
	)

	tracking2 := tracking.ReconstructTracking(
		uuid.New(),
		contactID,
		nil,
		tenantID,
		projectID,
		tracking.SourceGoogleAds,
		tracking.PlatformGoogle,
		"test-campaign-2",
		"test-ad-id-2",
		"https://test.com/ad2",
		"test-click-id-2",
		"test-conversion-data-2",
		"test-utm-source-2",
		"test-utm-medium-2",
		"test-utm-campaign-2",
		"test-utm-term-2",
		"test-utm-content-2",
		map[string]interface{}{
			"test_key_2": "test_value_2",
		},
		testTime(),
		testTime(),
	)

	trackings := []*tracking.Tracking{tracking1, tracking2}

	// Mock repository FindByContactID
	repo.On("FindByContactID", ctx, contactID).Return(trackings, nil)

	// Act
	result, err := useCase.Execute(ctx, contactID)

	// Assert - Verify ToDTOList() was called correctly by checking all fields
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 2)

	// Verify first tracking conversion
	assert.Equal(t, tracking1.ID(), result[0].ID)
	assert.Equal(t, tracking1.ContactID(), result[0].ContactID)
	assert.Equal(t, tracking1.SessionID(), result[0].SessionID)
	assert.Equal(t, tracking1.TenantID(), result[0].TenantID)
	assert.Equal(t, tracking1.ProjectID(), result[0].ProjectID)
	assert.Equal(t, string(tracking1.Source()), result[0].Source)
	assert.Equal(t, string(tracking1.Platform()), result[0].Platform)
	assert.Equal(t, tracking1.Campaign(), result[0].Campaign)
	assert.Equal(t, tracking1.AdID(), result[0].AdID)
	assert.Equal(t, tracking1.AdURL(), result[0].AdURL)
	assert.Equal(t, tracking1.ClickID(), result[0].ClickID)
	assert.Equal(t, tracking1.ConversionData(), result[0].ConversionData)
	assert.Equal(t, tracking1.UTMSource(), result[0].UTMSource)
	assert.Equal(t, tracking1.UTMMedium(), result[0].UTMMedium)
	assert.Equal(t, tracking1.UTMCampaign(), result[0].UTMCampaign)
	assert.Equal(t, tracking1.UTMTerm(), result[0].UTMTerm)
	assert.Equal(t, tracking1.UTMContent(), result[0].UTMContent)
	assert.Equal(t, tracking1.Metadata(), result[0].Metadata)
	assert.Equal(t, tracking1.CreatedAt(), result[0].CreatedAt)
	assert.Equal(t, tracking1.UpdatedAt(), result[0].UpdatedAt)

	// Verify second tracking conversion
	assert.Equal(t, tracking2.ID(), result[1].ID)
	assert.Equal(t, tracking2.ContactID(), result[1].ContactID)
	assert.Equal(t, tracking2.SessionID(), result[1].SessionID)
	assert.Equal(t, tracking2.TenantID(), result[1].TenantID)
	assert.Equal(t, tracking2.ProjectID(), result[1].ProjectID)
	assert.Equal(t, string(tracking2.Source()), result[1].Source)
	assert.Equal(t, string(tracking2.Platform()), result[1].Platform)
	assert.Equal(t, tracking2.Campaign(), result[1].Campaign)
	assert.Equal(t, tracking2.AdID(), result[1].AdID)
	assert.Equal(t, tracking2.AdURL(), result[1].AdURL)
	assert.Equal(t, tracking2.ClickID(), result[1].ClickID)
	assert.Equal(t, tracking2.ConversionData(), result[1].ConversionData)
	assert.Equal(t, tracking2.UTMSource(), result[1].UTMSource)
	assert.Equal(t, tracking2.UTMMedium(), result[1].UTMMedium)
	assert.Equal(t, tracking2.UTMCampaign(), result[1].UTMCampaign)
	assert.Equal(t, tracking2.UTMTerm(), result[1].UTMTerm)
	assert.Equal(t, tracking2.UTMContent(), result[1].UTMContent)
	assert.Equal(t, tracking2.Metadata(), result[1].Metadata)
	assert.Equal(t, tracking2.CreatedAt(), result[1].CreatedAt)
	assert.Equal(t, tracking2.UpdatedAt(), result[1].UpdatedAt)

	repo.AssertExpectations(t)
}

func TestGetContactTrackingsUseCase_Execute_MixedSources(t *testing.T) {
	// Arrange
	ctx := context.Background()
	repo := new(MockTrackingRepository)
	logger := zaptest.NewLogger(t)

	useCase := NewGetContactTrackingsUseCase(repo, logger)

	contactID := uuid.New()
	projectID := uuid.New()
	tenantID := "tenant-123"

	// Create trackings from all different sources
	allSources := []struct {
		source   tracking.Source
		platform tracking.Platform
	}{
		{tracking.SourceMetaAds, tracking.PlatformInstagram},
		{tracking.SourceGoogleAds, tracking.PlatformGoogle},
		{tracking.SourceTikTokAds, tracking.PlatformTikTok},
		{tracking.SourceLinkedIn, tracking.PlatformLinkedIn},
		{tracking.SourceOrganic, tracking.PlatformWhatsApp},
		{tracking.SourceDirect, tracking.PlatformFacebook},
		{tracking.SourceReferral, tracking.PlatformOther},
		{tracking.SourceOther, tracking.PlatformOther},
	}

	trackings := make([]*tracking.Tracking, 0, len(allSources))
	for _, src := range allSources {
		t := tracking.ReconstructTracking(
			uuid.New(),
			contactID,
			nil,
			tenantID,
			projectID,
			src.source,
			src.platform,
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
		trackings = append(trackings, t)
	}

	// Mock repository FindByContactID
	repo.On("FindByContactID", ctx, contactID).Return(trackings, nil)

	// Act
	result, err := useCase.Execute(ctx, contactID)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, len(allSources))

	// Verify all sources are present
	for i, src := range allSources {
		assert.Equal(t, string(src.source), result[i].Source)
		assert.Equal(t, string(src.platform), result[i].Platform)
		assert.Equal(t, contactID, result[i].ContactID)
	}

	repo.AssertExpectations(t)
}

func TestGetContactTrackingsUseCase_Execute_WithAndWithoutSessionID(t *testing.T) {
	// Arrange
	ctx := context.Background()
	repo := new(MockTrackingRepository)
	logger := zaptest.NewLogger(t)

	useCase := NewGetContactTrackingsUseCase(repo, logger)

	contactID := uuid.New()
	sessionID := uuid.New()
	projectID := uuid.New()
	tenantID := "tenant-123"

	// Create tracking with session ID
	trackingWithSession := tracking.ReconstructTracking(
		uuid.New(),
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

	// Create tracking without session ID
	trackingWithoutSession := tracking.ReconstructTracking(
		uuid.New(),
		contactID,
		nil,
		tenantID,
		projectID,
		tracking.SourceGoogleAds,
		tracking.PlatformGoogle,
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

	trackings := []*tracking.Tracking{trackingWithSession, trackingWithoutSession}

	// Mock repository FindByContactID
	repo.On("FindByContactID", ctx, contactID).Return(trackings, nil)

	// Act
	result, err := useCase.Execute(ctx, contactID)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 2)

	// Verify first tracking has session ID
	assert.NotNil(t, result[0].SessionID)
	assert.Equal(t, sessionID, *result[0].SessionID)

	// Verify second tracking has no session ID
	assert.Nil(t, result[1].SessionID)

	repo.AssertExpectations(t)
}

func TestGetContactTrackingsUseCase_Execute_NilUUID(t *testing.T) {
	// Arrange
	ctx := context.Background()
	repo := new(MockTrackingRepository)
	logger := zaptest.NewLogger(t)

	useCase := NewGetContactTrackingsUseCase(repo, logger)

	nilUUID := uuid.Nil

	// Mock repository FindByContactID - return empty list for nil UUID
	trackings := []*tracking.Tracking{}
	repo.On("FindByContactID", ctx, nilUUID).Return(trackings, nil)

	// Act
	result, err := useCase.Execute(ctx, nilUUID)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Empty(t, result)

	repo.AssertExpectations(t)
}

func TestGetContactTrackingsUseCase_Execute_LargeNumberOfTrackings(t *testing.T) {
	// Arrange
	ctx := context.Background()
	repo := new(MockTrackingRepository)
	logger := zaptest.NewLogger(t)

	useCase := NewGetContactTrackingsUseCase(repo, logger)

	contactID := uuid.New()
	projectID := uuid.New()
	tenantID := "tenant-123"

	// Create 100 tracking entities
	trackings := make([]*tracking.Tracking, 100)
	for i := 0; i < 100; i++ {
		trackings[i] = tracking.ReconstructTracking(
			uuid.New(),
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
	}

	// Mock repository FindByContactID
	repo.On("FindByContactID", ctx, contactID).Return(trackings, nil)

	// Act
	result, err := useCase.Execute(ctx, contactID)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 100)

	// Verify all trackings have the same contact ID
	for _, dto := range result {
		assert.Equal(t, contactID, dto.ContactID)
	}

	repo.AssertExpectations(t)
}

func TestGetContactTrackingsUseCase_Execute_ComplexMetadata(t *testing.T) {
	// Arrange
	ctx := context.Background()
	repo := new(MockTrackingRepository)
	logger := zaptest.NewLogger(t)

	useCase := NewGetContactTrackingsUseCase(repo, logger)

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
		uuid.New(),
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

	trackings := []*tracking.Tracking{sampleTracking}

	// Mock repository FindByContactID
	repo.On("FindByContactID", ctx, contactID).Return(trackings, nil)

	// Act
	result, err := useCase.Execute(ctx, contactID)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 1)
	assert.NotNil(t, result[0].Metadata)
	assert.Equal(t, "test", result[0].Metadata["string_value"])
	assert.Equal(t, 123, result[0].Metadata["int_value"])
	assert.Equal(t, 456.78, result[0].Metadata["float_value"])
	assert.Equal(t, true, result[0].Metadata["bool_value"])

	repo.AssertExpectations(t)
}

func TestGetContactTrackingsUseCase_Execute_DatabaseTimeout(t *testing.T) {
	// Arrange
	ctx := context.Background()
	repo := new(MockTrackingRepository)
	logger := zaptest.NewLogger(t)

	useCase := NewGetContactTrackingsUseCase(repo, logger)

	contactID := uuid.New()
	timeoutError := errors.New("database query timeout")

	// Mock repository FindByContactID - return timeout error
	repo.On("FindByContactID", ctx, contactID).Return(nil, timeoutError)

	// Act
	result, err := useCase.Execute(ctx, contactID)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to fetch trackings")
	assert.ErrorIs(t, err, timeoutError)

	repo.AssertExpectations(t)
}

func TestGetContactTrackingsUseCase_Execute_ConnectionError(t *testing.T) {
	// Arrange
	ctx := context.Background()
	repo := new(MockTrackingRepository)
	logger := zaptest.NewLogger(t)

	useCase := NewGetContactTrackingsUseCase(repo, logger)

	contactID := uuid.New()
	connectionError := errors.New("connection refused")

	// Mock repository FindByContactID - return connection error
	repo.On("FindByContactID", ctx, contactID).Return(nil, connectionError)

	// Act
	result, err := useCase.Execute(ctx, contactID)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to fetch trackings")
	assert.ErrorIs(t, err, connectionError)

	repo.AssertExpectations(t)
}

func TestGetContactTrackingsUseCase_Execute_MinimalFieldsMultipleTrackings(t *testing.T) {
	// Arrange
	ctx := context.Background()
	repo := new(MockTrackingRepository)
	logger := zaptest.NewLogger(t)

	useCase := NewGetContactTrackingsUseCase(repo, logger)

	contactID := uuid.New()
	projectID := uuid.New()
	tenantID := "tenant-123"

	// Create multiple tracking entities with minimal fields
	tracking1 := tracking.ReconstructTracking(
		uuid.New(),
		contactID,
		nil,
		tenantID,
		projectID,
		tracking.SourceOrganic,
		tracking.PlatformWhatsApp,
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
		map[string]interface{}{},
		testTime(),
		testTime(),
	)

	tracking2 := tracking.ReconstructTracking(
		uuid.New(),
		contactID,
		nil,
		tenantID,
		projectID,
		tracking.SourceDirect,
		tracking.PlatformFacebook,
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
		map[string]interface{}{},
		testTime(),
		testTime(),
	)

	trackings := []*tracking.Tracking{tracking1, tracking2}

	// Mock repository FindByContactID
	repo.On("FindByContactID", ctx, contactID).Return(trackings, nil)

	// Act
	result, err := useCase.Execute(ctx, contactID)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 2)

	// Verify all optional fields are empty
	for _, dto := range result {
		assert.Empty(t, dto.Campaign)
		assert.Empty(t, dto.AdID)
		assert.Empty(t, dto.AdURL)
		assert.Empty(t, dto.ClickID)
		assert.Empty(t, dto.ConversionData)
		assert.Empty(t, dto.UTMSource)
		assert.Empty(t, dto.UTMMedium)
		assert.Empty(t, dto.UTMCampaign)
		assert.Empty(t, dto.UTMTerm)
		assert.Empty(t, dto.UTMContent)
		assert.Nil(t, dto.SessionID)
	}

	repo.AssertExpectations(t)
}
