package tracking

import (
	"context"
	"errors"
	"testing"

	"github.com/ventros/crm/internal/domain/core/shared"
	"github.com/ventros/crm/internal/domain/crm/tracking"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// ========== Mocks ==========

type MockTrackingRepository struct {
	mock.Mock
}

func (m *MockTrackingRepository) Create(ctx context.Context, t *tracking.Tracking) error {
	args := m.Called(ctx, t)
	return args.Error(0)
}

func (m *MockTrackingRepository) FindByID(ctx context.Context, id uuid.UUID) (*tracking.Tracking, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*tracking.Tracking), args.Error(1)
}

func (m *MockTrackingRepository) FindByContactID(ctx context.Context, contactID uuid.UUID) ([]*tracking.Tracking, error) {
	args := m.Called(ctx, contactID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*tracking.Tracking), args.Error(1)
}

func (m *MockTrackingRepository) FindBySessionID(ctx context.Context, sessionID uuid.UUID) ([]*tracking.Tracking, error) {
	args := m.Called(ctx, sessionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*tracking.Tracking), args.Error(1)
}

func (m *MockTrackingRepository) FindByProjectID(ctx context.Context, projectID uuid.UUID, limit, offset int) ([]*tracking.Tracking, error) {
	args := m.Called(ctx, projectID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*tracking.Tracking), args.Error(1)
}

func (m *MockTrackingRepository) FindBySource(ctx context.Context, projectID uuid.UUID, source tracking.Source, limit, offset int) ([]*tracking.Tracking, error) {
	args := m.Called(ctx, projectID, source, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*tracking.Tracking), args.Error(1)
}

func (m *MockTrackingRepository) FindByCampaign(ctx context.Context, projectID uuid.UUID, campaign string, limit, offset int) ([]*tracking.Tracking, error) {
	args := m.Called(ctx, projectID, campaign, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*tracking.Tracking), args.Error(1)
}

func (m *MockTrackingRepository) FindByClickID(ctx context.Context, clickID string) (*tracking.Tracking, error) {
	args := m.Called(ctx, clickID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*tracking.Tracking), args.Error(1)
}

func (m *MockTrackingRepository) Update(ctx context.Context, t *tracking.Tracking) error {
	args := m.Called(ctx, t)
	return args.Error(0)
}

func (m *MockTrackingRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

type MockEventBus struct {
	mock.Mock
}

func (m *MockEventBus) Publish(ctx context.Context, event shared.DomainEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

type MockTransactionManager struct {
	mock.Mock
}

func (m *MockTransactionManager) ExecuteInTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	args := m.Called(ctx, fn)
	if len(args) > 0 && args.Get(0) != nil {
		// If the mock was configured to return an error, return it
		if err, ok := args.Get(0).(error); ok {
			return err
		}
	}
	// Otherwise, execute the function directly (simulating successful transaction)
	return fn(ctx)
}

// SimpleTransactionManager is a test transaction manager that just executes the function
type SimpleTransactionManager struct{}

func (m *SimpleTransactionManager) ExecuteInTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}

// MockTransactionManagerWithRollback is a transaction manager that tracks rollback
type MockTransactionManagerWithRollback struct {
	rolledBack bool
}

func (m *MockTransactionManagerWithRollback) ExecuteInTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	err := fn(ctx)
	if err != nil {
		m.rolledBack = true
	}
	return err
}

// ========== Tests ==========

func TestNewCreateTrackingUseCase(t *testing.T) {
	// Arrange
	repo := new(MockTrackingRepository)
	eventBus := new(MockEventBus)
	logger := zap.NewNop()
	txManager := &SimpleTransactionManager{}

	// Act
	useCase := NewCreateTrackingUseCase(repo, eventBus, logger, txManager)

	// Assert
	assert.NotNil(t, useCase)
	assert.Equal(t, repo, useCase.repo)
	assert.Equal(t, eventBus, useCase.eventBus)
	assert.Equal(t, logger, useCase.logger)
	assert.Equal(t, txManager, useCase.txManager)
}

func TestCreateTrackingUseCase_Execute_Success_MinimalFields(t *testing.T) {
	// Arrange
	ctx := context.Background()
	repo := new(MockTrackingRepository)
	eventBus := new(MockEventBus)
	logger := zap.NewNop()
	txManager := &SimpleTransactionManager{}

	useCase := NewCreateTrackingUseCase(repo, eventBus, logger, txManager)

	contactID := uuid.New()
	projectID := uuid.New()
	tenantID := "tenant-123"

	dto := CreateTrackingDTO{
		ContactID: contactID,
		SessionID: nil,
		TenantID:  tenantID,
		ProjectID: projectID,
		Source:    string(tracking.SourceMetaAds),
		Platform:  string(tracking.PlatformInstagram),
	}

	// Mock repository create
	repo.On("Create", ctx, mock.AnythingOfType("*tracking.Tracking")).Return(nil)

	// Mock event bus publish (TrackingCreatedEvent)
	eventBus.On("Publish", ctx, mock.AnythingOfType("tracking.TrackingCreatedEvent")).Return(nil)

	// Act
	result, err := useCase.Execute(ctx, dto)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEqual(t, uuid.Nil, result.ID)
	assert.Equal(t, contactID, result.ContactID)
	assert.Equal(t, projectID, result.ProjectID)
	assert.Equal(t, tenantID, result.TenantID)
	assert.Equal(t, string(tracking.SourceMetaAds), result.Source)
	assert.Equal(t, string(tracking.PlatformInstagram), result.Platform)

	repo.AssertExpectations(t)
	eventBus.AssertExpectations(t)
}

func TestCreateTrackingUseCase_Execute_Success_AllFields(t *testing.T) {
	// Arrange
	ctx := context.Background()
	repo := new(MockTrackingRepository)
	eventBus := new(MockEventBus)
	logger := zap.NewNop()
	txManager := &SimpleTransactionManager{}

	useCase := NewCreateTrackingUseCase(repo, eventBus, logger, txManager)

	contactID := uuid.New()
	sessionID := uuid.New()
	projectID := uuid.New()
	tenantID := "tenant-123"

	dto := CreateTrackingDTO{
		ContactID:      contactID,
		SessionID:      &sessionID,
		TenantID:       tenantID,
		ProjectID:      projectID,
		Source:         string(tracking.SourceGoogleAds),
		Platform:       string(tracking.PlatformGoogle),
		Campaign:       "summer-sale-2025",
		AdID:           "ad-12345",
		AdURL:          "https://example.com/ad",
		ClickID:        "click-67890",
		ConversionData: `{"conversion_id": "conv-123"}`,
		UTMSource:      "google",
		UTMMedium:      "cpc",
		UTMCampaign:    "summer-sale",
		UTMTerm:        "shoes",
		UTMContent:     "banner-top",
		Metadata: map[string]interface{}{
			"key1": "value1",
			"key2": 123,
		},
	}

	// Mock repository create
	repo.On("Create", ctx, mock.AnythingOfType("*tracking.Tracking")).Return(nil)

	// Mock event bus publish (TrackingCreatedEvent)
	eventBus.On("Publish", ctx, mock.AnythingOfType("tracking.TrackingCreatedEvent")).Return(nil)

	// Act
	result, err := useCase.Execute(ctx, dto)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEqual(t, uuid.Nil, result.ID)
	assert.Equal(t, contactID, result.ContactID)
	assert.Equal(t, &sessionID, result.SessionID)
	assert.Equal(t, projectID, result.ProjectID)
	assert.Equal(t, tenantID, result.TenantID)
	assert.Equal(t, string(tracking.SourceGoogleAds), result.Source)
	assert.Equal(t, string(tracking.PlatformGoogle), result.Platform)
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
	eventBus.AssertExpectations(t)
}

func TestCreateTrackingUseCase_Execute_Success_AllSources(t *testing.T) {
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
			eventBus := new(MockEventBus)
			logger := zap.NewNop()
			txManager := &SimpleTransactionManager{}

			useCase := NewCreateTrackingUseCase(repo, eventBus, logger, txManager)

			contactID := uuid.New()
			projectID := uuid.New()
			tenantID := "tenant-123"

			dto := CreateTrackingDTO{
				ContactID: contactID,
				TenantID:  tenantID,
				ProjectID: projectID,
				Source:    string(tc.source),
				Platform:  string(tc.platform),
			}

			// Mock repository create
			repo.On("Create", ctx, mock.AnythingOfType("*tracking.Tracking")).Return(nil)

			// Mock event bus publish
			eventBus.On("Publish", ctx, mock.AnythingOfType("tracking.TrackingCreatedEvent")).Return(nil)

			// Act
			result, err := useCase.Execute(ctx, dto)

			// Assert
			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, string(tc.source), result.Source)
			assert.Equal(t, string(tc.platform), result.Platform)

			repo.AssertExpectations(t)
			eventBus.AssertExpectations(t)
		})
	}
}

func TestCreateTrackingUseCase_Execute_InvalidSource(t *testing.T) {
	// Arrange
	ctx := context.Background()
	repo := new(MockTrackingRepository)
	eventBus := new(MockEventBus)
	logger := zap.NewNop()
	txManager := &SimpleTransactionManager{}

	useCase := NewCreateTrackingUseCase(repo, eventBus, logger, txManager)

	contactID := uuid.New()
	projectID := uuid.New()
	tenantID := "tenant-123"

	dto := CreateTrackingDTO{
		ContactID: contactID,
		TenantID:  tenantID,
		ProjectID: projectID,
		Source:    "invalid_source",
		Platform:  string(tracking.PlatformInstagram),
	}

	// Act
	result, err := useCase.Execute(ctx, dto)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "invalid source")

	repo.AssertNotCalled(t, "Create")
	eventBus.AssertNotCalled(t, "Publish")
}

func TestCreateTrackingUseCase_Execute_InvalidPlatform(t *testing.T) {
	// Arrange
	ctx := context.Background()
	repo := new(MockTrackingRepository)
	eventBus := new(MockEventBus)
	logger := zap.NewNop()
	txManager := &SimpleTransactionManager{}

	useCase := NewCreateTrackingUseCase(repo, eventBus, logger, txManager)

	contactID := uuid.New()
	projectID := uuid.New()
	tenantID := "tenant-123"

	dto := CreateTrackingDTO{
		ContactID: contactID,
		TenantID:  tenantID,
		ProjectID: projectID,
		Source:    string(tracking.SourceMetaAds),
		Platform:  "invalid_platform",
	}

	// Act
	result, err := useCase.Execute(ctx, dto)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "invalid platform")

	repo.AssertNotCalled(t, "Create")
	eventBus.AssertNotCalled(t, "Publish")
}

func TestCreateTrackingUseCase_Execute_MissingContactID(t *testing.T) {
	// Arrange
	ctx := context.Background()
	repo := new(MockTrackingRepository)
	eventBus := new(MockEventBus)
	logger := zap.NewNop()
	txManager := &SimpleTransactionManager{}

	useCase := NewCreateTrackingUseCase(repo, eventBus, logger, txManager)

	projectID := uuid.New()
	tenantID := "tenant-123"

	dto := CreateTrackingDTO{
		ContactID: uuid.Nil,
		TenantID:  tenantID,
		ProjectID: projectID,
		Source:    string(tracking.SourceMetaAds),
		Platform:  string(tracking.PlatformInstagram),
	}

	// Act
	result, err := useCase.Execute(ctx, dto)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to create tracking")

	repo.AssertNotCalled(t, "Create")
	eventBus.AssertNotCalled(t, "Publish")
}

func TestCreateTrackingUseCase_Execute_MissingTenantID(t *testing.T) {
	// Arrange
	ctx := context.Background()
	repo := new(MockTrackingRepository)
	eventBus := new(MockEventBus)
	logger := zap.NewNop()
	txManager := &SimpleTransactionManager{}

	useCase := NewCreateTrackingUseCase(repo, eventBus, logger, txManager)

	contactID := uuid.New()
	projectID := uuid.New()

	dto := CreateTrackingDTO{
		ContactID: contactID,
		TenantID:  "",
		ProjectID: projectID,
		Source:    string(tracking.SourceMetaAds),
		Platform:  string(tracking.PlatformInstagram),
	}

	// Act
	result, err := useCase.Execute(ctx, dto)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to create tracking")

	repo.AssertNotCalled(t, "Create")
	eventBus.AssertNotCalled(t, "Publish")
}

func TestCreateTrackingUseCase_Execute_MissingProjectID(t *testing.T) {
	// Arrange
	ctx := context.Background()
	repo := new(MockTrackingRepository)
	eventBus := new(MockEventBus)
	logger := zap.NewNop()
	txManager := &SimpleTransactionManager{}

	useCase := NewCreateTrackingUseCase(repo, eventBus, logger, txManager)

	contactID := uuid.New()
	tenantID := "tenant-123"

	dto := CreateTrackingDTO{
		ContactID: contactID,
		TenantID:  tenantID,
		ProjectID: uuid.Nil,
		Source:    string(tracking.SourceMetaAds),
		Platform:  string(tracking.PlatformInstagram),
	}

	// Act
	result, err := useCase.Execute(ctx, dto)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to create tracking")

	repo.AssertNotCalled(t, "Create")
	eventBus.AssertNotCalled(t, "Publish")
}

func TestCreateTrackingUseCase_Execute_RepositoryCreateError(t *testing.T) {
	// Arrange
	ctx := context.Background()
	repo := new(MockTrackingRepository)
	eventBus := new(MockEventBus)
	logger := zap.NewNop()
	txManager := &SimpleTransactionManager{}

	useCase := NewCreateTrackingUseCase(repo, eventBus, logger, txManager)

	contactID := uuid.New()
	projectID := uuid.New()
	tenantID := "tenant-123"

	dto := CreateTrackingDTO{
		ContactID: contactID,
		TenantID:  tenantID,
		ProjectID: projectID,
		Source:    string(tracking.SourceMetaAds),
		Platform:  string(tracking.PlatformInstagram),
	}

	// Mock repository create error
	createError := errors.New("database connection error")
	repo.On("Create", ctx, mock.AnythingOfType("*tracking.Tracking")).Return(createError)

	// Act
	result, err := useCase.Execute(ctx, dto)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to save tracking")

	repo.AssertExpectations(t)
	eventBus.AssertNotCalled(t, "Publish")
}

func TestCreateTrackingUseCase_Execute_EventBusPublishError(t *testing.T) {
	// Arrange
	ctx := context.Background()
	repo := new(MockTrackingRepository)
	eventBus := new(MockEventBus)
	logger := zap.NewNop()
	txManager := &SimpleTransactionManager{}

	useCase := NewCreateTrackingUseCase(repo, eventBus, logger, txManager)

	contactID := uuid.New()
	projectID := uuid.New()
	tenantID := "tenant-123"

	dto := CreateTrackingDTO{
		ContactID: contactID,
		TenantID:  tenantID,
		ProjectID: projectID,
		Source:    string(tracking.SourceMetaAds),
		Platform:  string(tracking.PlatformInstagram),
	}

	// Mock repository create
	repo.On("Create", ctx, mock.AnythingOfType("*tracking.Tracking")).Return(nil)

	// Mock event bus publish error
	publishError := errors.New("event bus unavailable")
	eventBus.On("Publish", ctx, mock.AnythingOfType("tracking.TrackingCreatedEvent")).Return(publishError)

	// Act
	result, err := useCase.Execute(ctx, dto)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to publish event")

	repo.AssertExpectations(t)
	eventBus.AssertExpectations(t)
}

func TestCreateTrackingUseCase_Execute_TransactionRollbackOnCreateError(t *testing.T) {
	// Arrange
	ctx := context.Background()
	repo := new(MockTrackingRepository)
	eventBus := new(MockEventBus)
	logger := zap.NewNop()
	txManager := &MockTransactionManagerWithRollback{
		rolledBack: false,
	}

	useCase := NewCreateTrackingUseCase(repo, eventBus, logger, txManager)

	contactID := uuid.New()
	projectID := uuid.New()
	tenantID := "tenant-123"

	dto := CreateTrackingDTO{
		ContactID: contactID,
		TenantID:  tenantID,
		ProjectID: projectID,
		Source:    string(tracking.SourceMetaAds),
		Platform:  string(tracking.PlatformInstagram),
	}

	// Mock repository create error
	createError := errors.New("database error")
	repo.On("Create", ctx, mock.AnythingOfType("*tracking.Tracking")).Return(createError)

	// Act
	result, err := useCase.Execute(ctx, dto)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, txManager.rolledBack, "transaction should have been rolled back")

	repo.AssertExpectations(t)
	eventBus.AssertNotCalled(t, "Publish")
}

func TestCreateTrackingUseCase_Execute_TransactionRollbackOnPublishError(t *testing.T) {
	// Arrange
	ctx := context.Background()
	repo := new(MockTrackingRepository)
	eventBus := new(MockEventBus)
	logger := zap.NewNop()
	txManager := &MockTransactionManagerWithRollback{
		rolledBack: false,
	}

	useCase := NewCreateTrackingUseCase(repo, eventBus, logger, txManager)

	contactID := uuid.New()
	projectID := uuid.New()
	tenantID := "tenant-123"

	dto := CreateTrackingDTO{
		ContactID: contactID,
		TenantID:  tenantID,
		ProjectID: projectID,
		Source:    string(tracking.SourceMetaAds),
		Platform:  string(tracking.PlatformInstagram),
	}

	// Mock repository create
	repo.On("Create", ctx, mock.AnythingOfType("*tracking.Tracking")).Return(nil)

	// Mock event bus publish error
	publishError := errors.New("event bus error")
	eventBus.On("Publish", ctx, mock.AnythingOfType("tracking.TrackingCreatedEvent")).Return(publishError)

	// Act
	result, err := useCase.Execute(ctx, dto)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, txManager.rolledBack, "transaction should have been rolled back")

	repo.AssertExpectations(t)
	eventBus.AssertExpectations(t)
}

func TestCreateTrackingUseCase_Execute_EventsClearedOnSuccess(t *testing.T) {
	// Arrange
	ctx := context.Background()
	repo := new(MockTrackingRepository)
	eventBus := new(MockEventBus)
	logger := zap.NewNop()
	txManager := &SimpleTransactionManager{}

	useCase := NewCreateTrackingUseCase(repo, eventBus, logger, txManager)

	contactID := uuid.New()
	projectID := uuid.New()
	tenantID := "tenant-123"

	dto := CreateTrackingDTO{
		ContactID: contactID,
		TenantID:  tenantID,
		ProjectID: projectID,
		Source:    string(tracking.SourceMetaAds),
		Platform:  string(tracking.PlatformInstagram),
	}

	// Mock repository create
	repo.On("Create", ctx, mock.AnythingOfType("*tracking.Tracking")).Return(nil)

	// Mock event bus publish (TrackingCreatedEvent)
	eventBus.On("Publish", ctx, mock.AnythingOfType("tracking.TrackingCreatedEvent")).Return(nil)

	// Act
	result, err := useCase.Execute(ctx, dto)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)

	// Verify events were cleared after successful execution
	// Note: We can't directly check the tracking entity here because ClearEvents()
	// is called after the transaction completes. This test verifies the flow executes without error.

	repo.AssertExpectations(t)
	eventBus.AssertExpectations(t)
}

func TestCreateTrackingUseCase_Execute_PartialUTMParameters(t *testing.T) {
	testCases := []struct {
		name        string
		utmSource   string
		utmMedium   string
		utmCampaign string
		utmTerm     string
		utmContent  string
		shouldSet   bool
	}{
		{
			name:        "Only UTMSource",
			utmSource:   "google",
			utmMedium:   "",
			utmCampaign: "",
			utmTerm:     "",
			utmContent:  "",
			shouldSet:   true,
		},
		{
			name:        "Only UTMMedium",
			utmSource:   "",
			utmMedium:   "cpc",
			utmCampaign: "",
			utmTerm:     "",
			utmContent:  "",
			shouldSet:   true,
		},
		{
			name:        "Only UTMCampaign",
			utmSource:   "",
			utmMedium:   "",
			utmCampaign: "summer-sale",
			utmTerm:     "",
			utmContent:  "",
			shouldSet:   true,
		},
		{
			name:        "All Empty",
			utmSource:   "",
			utmMedium:   "",
			utmCampaign: "",
			utmTerm:     "",
			utmContent:  "",
			shouldSet:   false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			ctx := context.Background()
			repo := new(MockTrackingRepository)
			eventBus := new(MockEventBus)
			logger := zap.NewNop()
			txManager := &SimpleTransactionManager{}

			useCase := NewCreateTrackingUseCase(repo, eventBus, logger, txManager)

			contactID := uuid.New()
			projectID := uuid.New()
			tenantID := "tenant-123"

			dto := CreateTrackingDTO{
				ContactID:   contactID,
				TenantID:    tenantID,
				ProjectID:   projectID,
				Source:      string(tracking.SourceMetaAds),
				Platform:    string(tracking.PlatformInstagram),
				UTMSource:   tc.utmSource,
				UTMMedium:   tc.utmMedium,
				UTMCampaign: tc.utmCampaign,
				UTMTerm:     tc.utmTerm,
				UTMContent:  tc.utmContent,
			}

			// Mock repository create
			repo.On("Create", ctx, mock.AnythingOfType("*tracking.Tracking")).Return(nil)

			// Mock event bus publish
			eventBus.On("Publish", ctx, mock.AnythingOfType("tracking.TrackingCreatedEvent")).Return(nil)

			// Act
			result, err := useCase.Execute(ctx, dto)

			// Assert
			assert.NoError(t, err)
			assert.NotNil(t, result)

			if tc.shouldSet {
				// At least one of the primary UTM parameters should be set
				hasUTM := result.UTMSource != "" || result.UTMMedium != "" || result.UTMCampaign != ""
				assert.True(t, hasUTM, "at least one primary UTM parameter should be set")
			}

			repo.AssertExpectations(t)
			eventBus.AssertExpectations(t)
		})
	}
}

func TestCreateTrackingUseCase_Execute_EmptyOptionalFields(t *testing.T) {
	// Arrange
	ctx := context.Background()
	repo := new(MockTrackingRepository)
	eventBus := new(MockEventBus)
	logger := zap.NewNop()
	txManager := &SimpleTransactionManager{}

	useCase := NewCreateTrackingUseCase(repo, eventBus, logger, txManager)

	contactID := uuid.New()
	projectID := uuid.New()
	tenantID := "tenant-123"

	dto := CreateTrackingDTO{
		ContactID:      contactID,
		TenantID:       tenantID,
		ProjectID:      projectID,
		Source:         string(tracking.SourceMetaAds),
		Platform:       string(tracking.PlatformInstagram),
		Campaign:       "",
		AdID:           "",
		AdURL:          "",
		ClickID:        "",
		ConversionData: "",
		UTMSource:      "",
		UTMMedium:      "",
		UTMCampaign:    "",
		UTMTerm:        "",
		UTMContent:     "",
		Metadata:       nil,
	}

	// Mock repository create
	repo.On("Create", ctx, mock.AnythingOfType("*tracking.Tracking")).Return(nil)

	// Mock event bus publish (TrackingCreatedEvent)
	eventBus.On("Publish", ctx, mock.AnythingOfType("tracking.TrackingCreatedEvent")).Return(nil)

	// Act
	result, err := useCase.Execute(ctx, dto)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "", result.Campaign)
	assert.Equal(t, "", result.AdID)
	assert.Equal(t, "", result.AdURL)
	assert.Equal(t, "", result.ClickID)
	assert.Equal(t, "", result.ConversionData)
	assert.Equal(t, "", result.UTMSource)
	assert.Equal(t, "", result.UTMMedium)
	assert.Equal(t, "", result.UTMCampaign)
	assert.Equal(t, "", result.UTMTerm)
	assert.Equal(t, "", result.UTMContent)

	repo.AssertExpectations(t)
	eventBus.AssertExpectations(t)
}

func TestCreateTrackingUseCase_Execute_EmptyMetadata(t *testing.T) {
	// Arrange
	ctx := context.Background()
	repo := new(MockTrackingRepository)
	eventBus := new(MockEventBus)
	logger := zap.NewNop()
	txManager := &SimpleTransactionManager{}

	useCase := NewCreateTrackingUseCase(repo, eventBus, logger, txManager)

	contactID := uuid.New()
	projectID := uuid.New()
	tenantID := "tenant-123"

	dto := CreateTrackingDTO{
		ContactID: contactID,
		TenantID:  tenantID,
		ProjectID: projectID,
		Source:    string(tracking.SourceMetaAds),
		Platform:  string(tracking.PlatformInstagram),
		Metadata:  map[string]interface{}{},
	}

	// Mock repository create
	repo.On("Create", ctx, mock.AnythingOfType("*tracking.Tracking")).Return(nil)

	// Mock event bus publish (TrackingCreatedEvent)
	eventBus.On("Publish", ctx, mock.AnythingOfType("tracking.TrackingCreatedEvent")).Return(nil)

	// Act
	result, err := useCase.Execute(ctx, dto)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)

	repo.AssertExpectations(t)
	eventBus.AssertExpectations(t)
}

func TestIsValidSource(t *testing.T) {
	testCases := []struct {
		name     string
		source   tracking.Source
		expected bool
	}{
		{"Valid_MetaAds", tracking.SourceMetaAds, true},
		{"Valid_GoogleAds", tracking.SourceGoogleAds, true},
		{"Valid_TikTokAds", tracking.SourceTikTokAds, true},
		{"Valid_LinkedIn", tracking.SourceLinkedIn, true},
		{"Valid_Organic", tracking.SourceOrganic, true},
		{"Valid_Direct", tracking.SourceDirect, true},
		{"Valid_Referral", tracking.SourceReferral, true},
		{"Valid_Other", tracking.SourceOther, true},
		{"Invalid_Empty", tracking.Source(""), false},
		{"Invalid_Unknown", tracking.Source("unknown"), false},
		{"Invalid_Random", tracking.Source("random_source"), false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			result := isValidSource(tc.source)

			// Assert
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestIsValidPlatform(t *testing.T) {
	testCases := []struct {
		name     string
		platform tracking.Platform
		expected bool
	}{
		{"Valid_Instagram", tracking.PlatformInstagram, true},
		{"Valid_Facebook", tracking.PlatformFacebook, true},
		{"Valid_Google", tracking.PlatformGoogle, true},
		{"Valid_TikTok", tracking.PlatformTikTok, true},
		{"Valid_LinkedIn", tracking.PlatformLinkedIn, true},
		{"Valid_WhatsApp", tracking.PlatformWhatsApp, true},
		{"Valid_Other", tracking.PlatformOther, true},
		{"Invalid_Empty", tracking.Platform(""), false},
		{"Invalid_Unknown", tracking.Platform("unknown"), false},
		{"Invalid_Random", tracking.Platform("random_platform"), false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			result := isValidPlatform(tc.platform)

			// Assert
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestCreateTrackingUseCase_Execute_ComplexMetadata(t *testing.T) {
	// Arrange
	ctx := context.Background()
	repo := new(MockTrackingRepository)
	eventBus := new(MockEventBus)
	logger := zap.NewNop()
	txManager := &SimpleTransactionManager{}

	useCase := NewCreateTrackingUseCase(repo, eventBus, logger, txManager)

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

	dto := CreateTrackingDTO{
		ContactID: contactID,
		TenantID:  tenantID,
		ProjectID: projectID,
		Source:    string(tracking.SourceMetaAds),
		Platform:  string(tracking.PlatformInstagram),
		Metadata:  complexMetadata,
	}

	// Mock repository create
	repo.On("Create", ctx, mock.AnythingOfType("*tracking.Tracking")).Return(nil)

	// Mock event bus publish (TrackingCreatedEvent)
	eventBus.On("Publish", ctx, mock.AnythingOfType("tracking.TrackingCreatedEvent")).Return(nil)

	// Act
	result, err := useCase.Execute(ctx, dto)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotNil(t, result.Metadata)
	assert.Equal(t, "test", result.Metadata["string_value"])
	assert.Equal(t, 123, result.Metadata["int_value"])
	assert.Equal(t, 456.78, result.Metadata["float_value"])
	assert.Equal(t, true, result.Metadata["bool_value"])

	repo.AssertExpectations(t)
	eventBus.AssertExpectations(t)
}

func TestCreateTrackingUseCase_Execute_WithSessionID(t *testing.T) {
	// Arrange
	ctx := context.Background()
	repo := new(MockTrackingRepository)
	eventBus := new(MockEventBus)
	logger := zap.NewNop()
	txManager := &SimpleTransactionManager{}

	useCase := NewCreateTrackingUseCase(repo, eventBus, logger, txManager)

	contactID := uuid.New()
	sessionID := uuid.New()
	projectID := uuid.New()
	tenantID := "tenant-123"

	dto := CreateTrackingDTO{
		ContactID: contactID,
		SessionID: &sessionID,
		TenantID:  tenantID,
		ProjectID: projectID,
		Source:    string(tracking.SourceMetaAds),
		Platform:  string(tracking.PlatformInstagram),
	}

	// Mock repository create
	repo.On("Create", ctx, mock.AnythingOfType("*tracking.Tracking")).Return(nil)

	// Mock event bus publish (TrackingCreatedEvent)
	eventBus.On("Publish", ctx, mock.AnythingOfType("tracking.TrackingCreatedEvent")).Return(nil)

	// Act
	result, err := useCase.Execute(ctx, dto)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotNil(t, result.SessionID)
	assert.Equal(t, sessionID, *result.SessionID)

	repo.AssertExpectations(t)
	eventBus.AssertExpectations(t)
}

func TestCreateTrackingUseCase_Execute_WithoutSessionID(t *testing.T) {
	// Arrange
	ctx := context.Background()
	repo := new(MockTrackingRepository)
	eventBus := new(MockEventBus)
	logger := zap.NewNop()
	txManager := &SimpleTransactionManager{}

	useCase := NewCreateTrackingUseCase(repo, eventBus, logger, txManager)

	contactID := uuid.New()
	projectID := uuid.New()
	tenantID := "tenant-123"

	dto := CreateTrackingDTO{
		ContactID: contactID,
		SessionID: nil,
		TenantID:  tenantID,
		ProjectID: projectID,
		Source:    string(tracking.SourceMetaAds),
		Platform:  string(tracking.PlatformInstagram),
	}

	// Mock repository create
	repo.On("Create", ctx, mock.AnythingOfType("*tracking.Tracking")).Return(nil)

	// Mock event bus publish (TrackingCreatedEvent)
	eventBus.On("Publish", ctx, mock.AnythingOfType("tracking.TrackingCreatedEvent")).Return(nil)

	// Act
	result, err := useCase.Execute(ctx, dto)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Nil(t, result.SessionID)

	repo.AssertExpectations(t)
	eventBus.AssertExpectations(t)
}
