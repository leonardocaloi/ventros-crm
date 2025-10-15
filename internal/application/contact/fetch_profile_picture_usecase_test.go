package contact

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/ventros/crm/internal/domain/crm/contact"
	"go.uber.org/zap"
)

// ========== Mock for WahaProfileService ==========

type MockWahaProfileService struct {
	mock.Mock
}

func (m *MockWahaProfileService) FetchAndUpdateContactProfilePicture(ctx context.Context, phone, session string) (string, error) {
	args := m.Called(ctx, phone, session)
	return args.String(0), args.Error(1)
}

// ========== Test Helpers ==========

func createTestContactForProfilePicture(projectID uuid.UUID, tenantID string) *contact.Contact {
	c, _ := contact.NewContact(projectID, tenantID, "John Doe")
	return c
}

func setupFetchProfilePictureTest(t *testing.T) (
	*MockContactRepository,
	*MockWahaProfileService,
	*MockEventBus,
	*zap.Logger,
	*FetchProfilePictureUseCase,
) {
	contactRepo := new(MockContactRepository)
	wahaService := new(MockWahaProfileService)
	eventBus := new(MockEventBus)
	logger := zap.NewNop()

	useCase := NewFetchProfilePictureUseCase(contactRepo, wahaService, eventBus, logger)

	return contactRepo, wahaService, eventBus, logger, useCase
}

// ========== Constructor Tests ==========

func TestNewFetchProfilePictureUseCase(t *testing.T) {
	// Arrange
	contactRepo := new(MockContactRepository)
	wahaService := new(MockWahaProfileService)
	eventBus := new(MockEventBus)
	logger := zap.NewNop()

	// Act
	useCase := NewFetchProfilePictureUseCase(contactRepo, wahaService, eventBus, logger)

	// Assert
	assert.NotNil(t, useCase)
	assert.Equal(t, contactRepo, useCase.contactRepo)
	assert.Equal(t, wahaService, useCase.wahaService)
	assert.Equal(t, eventBus, useCase.eventBus)
	assert.Equal(t, logger, useCase.logger)
}

// ========== Success Scenarios ==========

func TestFetchProfilePictureUseCase_Execute_Success(t *testing.T) {
	// Arrange
	ctx := context.Background()
	contactRepo, wahaService, eventBus, _, useCase := setupFetchProfilePictureTest(t)

	contactID := uuid.New()
	projectID := uuid.New()
	tenantID := "tenant-1"
	phone := "+5511999999999"
	session := "default"
	profilePictureURL := "https://example.com/profile.jpg"

	testContact := createTestContactForProfilePicture(projectID, tenantID)

	cmd := FetchProfilePictureCommand{
		ContactID: contactID,
		Phone:     phone,
		Session:   session,
	}

	// Mock expectations
	contactRepo.On("FindByID", ctx, contactID).Return(testContact, nil)
	wahaService.On("FetchAndUpdateContactProfilePicture", ctx, phone, session).Return(profilePictureURL, nil)
	contactRepo.On("Save", ctx, testContact).Return(nil)
	eventBus.On("Publish", ctx, mock.MatchedBy(func(event contact.DomainEvent) bool {
		if evt, ok := event.(contact.ContactProfilePictureUpdatedEvent); ok {
			return evt.ContactID == contactID &&
				evt.TenantID == tenantID &&
				evt.ProfilePictureURL == profilePictureURL &&
				!evt.FetchedAt.IsZero()
		}
		return false
	})).Return(nil)

	// Act
	err := useCase.Execute(ctx, cmd)

	// Assert
	assert.NoError(t, err)
	contactRepo.AssertExpectations(t)
	wahaService.AssertExpectations(t)
	eventBus.AssertExpectations(t)
}

func TestFetchProfilePictureUseCase_Execute_SuccessWithEmptyProfilePicture(t *testing.T) {
	// Arrange
	ctx := context.Background()
	contactRepo, wahaService, eventBus, _, useCase := setupFetchProfilePictureTest(t)

	contactID := uuid.New()
	projectID := uuid.New()
	tenantID := "tenant-1"
	phone := "+5511999999999"
	session := "default"

	testContact := createTestContactForProfilePicture(projectID, tenantID)

	cmd := FetchProfilePictureCommand{
		ContactID: contactID,
		Phone:     phone,
		Session:   session,
	}

	// Mock expectations - WAHA returns empty string (no profile picture)
	contactRepo.On("FindByID", ctx, contactID).Return(testContact, nil)
	wahaService.On("FetchAndUpdateContactProfilePicture", ctx, phone, session).Return("", nil)

	// Act
	err := useCase.Execute(ctx, cmd)

	// Assert
	assert.NoError(t, err)
	contactRepo.AssertExpectations(t)
	wahaService.AssertExpectations(t)
	// Save and Publish should NOT be called when profile picture is empty
	contactRepo.AssertNotCalled(t, "Save")
	eventBus.AssertNotCalled(t, "Publish")
}

func TestFetchProfilePictureUseCase_Execute_SuccessEventPublishFailureDoesNotReturnError(t *testing.T) {
	// Arrange
	ctx := context.Background()
	contactRepo, wahaService, eventBus, _, useCase := setupFetchProfilePictureTest(t)

	contactID := uuid.New()
	projectID := uuid.New()
	tenantID := "tenant-1"
	phone := "+5511999999999"
	session := "default"
	profilePictureURL := "https://example.com/profile.jpg"

	testContact := createTestContactForProfilePicture(projectID, tenantID)

	cmd := FetchProfilePictureCommand{
		ContactID: contactID,
		Phone:     phone,
		Session:   session,
	}

	// Mock expectations
	contactRepo.On("FindByID", ctx, contactID).Return(testContact, nil)
	wahaService.On("FetchAndUpdateContactProfilePicture", ctx, phone, session).Return(profilePictureURL, nil)
	contactRepo.On("Save", ctx, testContact).Return(nil)
	eventBus.On("Publish", ctx, mock.AnythingOfType("contact.ContactProfilePictureUpdatedEvent")).Return(errors.New("event bus error"))

	// Act
	err := useCase.Execute(ctx, cmd)

	// Assert - should NOT return error even if event publish fails
	assert.NoError(t, err)
	contactRepo.AssertExpectations(t)
	wahaService.AssertExpectations(t)
	eventBus.AssertExpectations(t)
}

// ========== Validation Scenarios ==========

func TestFetchProfilePictureUseCase_Execute_MissingContactID(t *testing.T) {
	// Arrange
	ctx := context.Background()
	contactRepo, wahaService, eventBus, _, useCase := setupFetchProfilePictureTest(t)

	cmd := FetchProfilePictureCommand{
		ContactID: uuid.Nil,
		Phone:     "+5511999999999",
		Session:   "default",
	}

	// Act
	err := useCase.Execute(ctx, cmd)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ContactID cannot be nil")

	contactRepo.AssertNotCalled(t, "FindByID")
	wahaService.AssertNotCalled(t, "FetchAndUpdateContactProfilePicture")
	eventBus.AssertNotCalled(t, "Publish")
}

func TestFetchProfilePictureUseCase_Execute_MissingPhone(t *testing.T) {
	// Arrange
	ctx := context.Background()
	contactRepo, wahaService, eventBus, _, useCase := setupFetchProfilePictureTest(t)

	cmd := FetchProfilePictureCommand{
		ContactID: uuid.New(),
		Phone:     "",
		Session:   "default",
	}

	// Act
	err := useCase.Execute(ctx, cmd)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Phone cannot be empty")

	contactRepo.AssertNotCalled(t, "FindByID")
	wahaService.AssertNotCalled(t, "FetchAndUpdateContactProfilePicture")
	eventBus.AssertNotCalled(t, "Publish")
}

func TestFetchProfilePictureUseCase_Execute_MissingSession(t *testing.T) {
	// Arrange
	ctx := context.Background()
	contactRepo, wahaService, eventBus, _, useCase := setupFetchProfilePictureTest(t)

	cmd := FetchProfilePictureCommand{
		ContactID: uuid.New(),
		Phone:     "+5511999999999",
		Session:   "",
	}

	// Act
	err := useCase.Execute(ctx, cmd)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Session cannot be empty")

	contactRepo.AssertNotCalled(t, "FindByID")
	wahaService.AssertNotCalled(t, "FetchAndUpdateContactProfilePicture")
	eventBus.AssertNotCalled(t, "Publish")
}

// ========== Error Scenarios ==========

func TestFetchProfilePictureUseCase_Execute_ContactNotFound(t *testing.T) {
	// Arrange
	ctx := context.Background()
	contactRepo, wahaService, eventBus, _, useCase := setupFetchProfilePictureTest(t)

	contactID := uuid.New()
	phone := "+5511999999999"
	session := "default"

	cmd := FetchProfilePictureCommand{
		ContactID: contactID,
		Phone:     phone,
		Session:   session,
	}

	// Mock expectations
	contactRepo.On("FindByID", ctx, contactID).Return(nil, errors.New("contact not found"))

	// Act
	err := useCase.Execute(ctx, cmd)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to find contact")

	contactRepo.AssertExpectations(t)
	wahaService.AssertNotCalled(t, "FetchAndUpdateContactProfilePicture")
	eventBus.AssertNotCalled(t, "Publish")
}

func TestFetchProfilePictureUseCase_Execute_ContactRepositoryError(t *testing.T) {
	// Arrange
	ctx := context.Background()
	contactRepo, wahaService, eventBus, _, useCase := setupFetchProfilePictureTest(t)

	contactID := uuid.New()
	phone := "+5511999999999"
	session := "default"

	cmd := FetchProfilePictureCommand{
		ContactID: contactID,
		Phone:     phone,
		Session:   session,
	}

	// Mock expectations
	dbError := errors.New("database connection error")
	contactRepo.On("FindByID", ctx, contactID).Return(nil, dbError)

	// Act
	err := useCase.Execute(ctx, cmd)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to find contact")

	contactRepo.AssertExpectations(t)
	wahaService.AssertNotCalled(t, "FetchAndUpdateContactProfilePicture")
	eventBus.AssertNotCalled(t, "Publish")
}

func TestFetchProfilePictureUseCase_Execute_WahaServiceError(t *testing.T) {
	// Arrange
	ctx := context.Background()
	contactRepo, wahaService, eventBus, _, useCase := setupFetchProfilePictureTest(t)

	contactID := uuid.New()
	projectID := uuid.New()
	tenantID := "tenant-1"
	phone := "+5511999999999"
	session := "default"

	testContact := createTestContactForProfilePicture(projectID, tenantID)

	cmd := FetchProfilePictureCommand{
		ContactID: contactID,
		Phone:     phone,
		Session:   session,
	}

	// Mock expectations
	wahaError := errors.New("WAHA API error")
	contactRepo.On("FindByID", ctx, contactID).Return(testContact, nil)
	wahaService.On("FetchAndUpdateContactProfilePicture", ctx, phone, session).Return("", wahaError)

	// Act
	err := useCase.Execute(ctx, cmd)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to fetch profile picture")

	contactRepo.AssertExpectations(t)
	wahaService.AssertExpectations(t)
	contactRepo.AssertNotCalled(t, "Save")
	eventBus.AssertNotCalled(t, "Publish")
}

func TestFetchProfilePictureUseCase_Execute_ContactSaveError(t *testing.T) {
	// Arrange
	ctx := context.Background()
	contactRepo, wahaService, eventBus, _, useCase := setupFetchProfilePictureTest(t)

	contactID := uuid.New()
	projectID := uuid.New()
	tenantID := "tenant-1"
	phone := "+5511999999999"
	session := "default"
	profilePictureURL := "https://example.com/profile.jpg"

	testContact := createTestContactForProfilePicture(projectID, tenantID)

	cmd := FetchProfilePictureCommand{
		ContactID: contactID,
		Phone:     phone,
		Session:   session,
	}

	// Mock expectations
	saveError := errors.New("database save error")
	contactRepo.On("FindByID", ctx, contactID).Return(testContact, nil)
	wahaService.On("FetchAndUpdateContactProfilePicture", ctx, phone, session).Return(profilePictureURL, nil)
	contactRepo.On("Save", ctx, testContact).Return(saveError)

	// Act
	err := useCase.Execute(ctx, cmd)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to update contact")

	contactRepo.AssertExpectations(t)
	wahaService.AssertExpectations(t)
	eventBus.AssertNotCalled(t, "Publish")
}

// ========== Edge Cases ==========

func TestFetchProfilePictureUseCase_Execute_WhitespacePhone(t *testing.T) {
	// Arrange
	ctx := context.Background()
	contactRepo, wahaService, eventBus, _, useCase := setupFetchProfilePictureTest(t)

	cmd := FetchProfilePictureCommand{
		ContactID: uuid.New(),
		Phone:     "   ",
		Session:   "default",
	}

	// Act
	err := useCase.Execute(ctx, cmd)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Phone cannot be empty")

	contactRepo.AssertNotCalled(t, "FindByID")
	wahaService.AssertNotCalled(t, "FetchAndUpdateContactProfilePicture")
	eventBus.AssertNotCalled(t, "Publish")
}

func TestFetchProfilePictureUseCase_Execute_WhitespaceSession(t *testing.T) {
	// Arrange
	ctx := context.Background()
	contactRepo, wahaService, eventBus, _, useCase := setupFetchProfilePictureTest(t)

	cmd := FetchProfilePictureCommand{
		ContactID: uuid.New(),
		Phone:     "+5511999999999",
		Session:   "   ",
	}

	// Act
	err := useCase.Execute(ctx, cmd)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Session cannot be empty")

	contactRepo.AssertNotCalled(t, "FindByID")
	wahaService.AssertNotCalled(t, "FetchAndUpdateContactProfilePicture")
	eventBus.AssertNotCalled(t, "Publish")
}

func TestFetchProfilePictureUseCase_Execute_ContextCancelled(t *testing.T) {
	// Arrange
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	contactRepo, wahaService, _, _, useCase := setupFetchProfilePictureTest(t)

	contactID := uuid.New()
	projectID := uuid.New()
	tenantID := "tenant-1"
	phone := "+5511999999999"
	session := "default"

	testContact := createTestContactForProfilePicture(projectID, tenantID)

	cmd := FetchProfilePictureCommand{
		ContactID: contactID,
		Phone:     phone,
		Session:   session,
	}

	// Mock expectations
	contactRepo.On("FindByID", ctx, contactID).Return(testContact, nil)
	wahaService.On("FetchAndUpdateContactProfilePicture", ctx, phone, session).Return("", context.Canceled)

	// Act
	err := useCase.Execute(ctx, cmd)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to fetch profile picture")

	contactRepo.AssertExpectations(t)
	wahaService.AssertExpectations(t)
}

func TestFetchProfilePictureUseCase_Execute_VeryLongProfilePictureURL(t *testing.T) {
	// Arrange
	ctx := context.Background()
	contactRepo, wahaService, eventBus, _, useCase := setupFetchProfilePictureTest(t)

	contactID := uuid.New()
	projectID := uuid.New()
	tenantID := "tenant-1"
	phone := "+5511999999999"
	session := "default"

	// Create a very long URL (2000+ characters)
	veryLongURL := "https://example.com/profile/"
	for i := 0; i < 200; i++ {
		veryLongURL += "very-long-path-segment/"
	}
	veryLongURL += "image.jpg"

	testContact := createTestContactForProfilePicture(projectID, tenantID)

	cmd := FetchProfilePictureCommand{
		ContactID: contactID,
		Phone:     phone,
		Session:   session,
	}

	// Mock expectations
	contactRepo.On("FindByID", ctx, contactID).Return(testContact, nil)
	wahaService.On("FetchAndUpdateContactProfilePicture", ctx, phone, session).Return(veryLongURL, nil)
	contactRepo.On("Save", ctx, testContact).Return(nil)
	eventBus.On("Publish", ctx, mock.AnythingOfType("contact.ContactProfilePictureUpdatedEvent")).Return(nil)

	// Act
	err := useCase.Execute(ctx, cmd)

	// Assert
	assert.NoError(t, err)
	contactRepo.AssertExpectations(t)
	wahaService.AssertExpectations(t)
	eventBus.AssertExpectations(t)
}

func TestFetchProfilePictureUseCase_Execute_SpecialCharactersInPhone(t *testing.T) {
	// Arrange
	ctx := context.Background()
	contactRepo, wahaService, eventBus, _, useCase := setupFetchProfilePictureTest(t)

	contactID := uuid.New()
	projectID := uuid.New()
	tenantID := "tenant-1"
	phone := "+55 (11) 99999-9999" // Phone with special characters
	session := "default"
	profilePictureURL := "https://example.com/profile.jpg"

	testContact := createTestContactForProfilePicture(projectID, tenantID)

	cmd := FetchProfilePictureCommand{
		ContactID: contactID,
		Phone:     phone,
		Session:   session,
	}

	// Mock expectations
	contactRepo.On("FindByID", ctx, contactID).Return(testContact, nil)
	wahaService.On("FetchAndUpdateContactProfilePicture", ctx, phone, session).Return(profilePictureURL, nil)
	contactRepo.On("Save", ctx, testContact).Return(nil)
	eventBus.On("Publish", ctx, mock.AnythingOfType("contact.ContactProfilePictureUpdatedEvent")).Return(nil)

	// Act
	err := useCase.Execute(ctx, cmd)

	// Assert
	assert.NoError(t, err)
	contactRepo.AssertExpectations(t)
	wahaService.AssertExpectations(t)
	eventBus.AssertExpectations(t)
}

func TestFetchProfilePictureUseCase_Execute_DifferentSessionNames(t *testing.T) {
	testCases := []struct {
		name    string
		session string
	}{
		{"default session", "default"},
		{"custom session", "my-whatsapp-session"},
		{"numeric session", "session123"},
		{"with hyphens", "session-prod-01"},
		{"with underscores", "session_prod_01"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			ctx := context.Background()
			contactRepo, wahaService, eventBus, _, useCase := setupFetchProfilePictureTest(t)

			contactID := uuid.New()
			projectID := uuid.New()
			tenantID := "tenant-1"
			phone := "+5511999999999"
			profilePictureURL := "https://example.com/profile.jpg"

			testContact := createTestContactForProfilePicture(projectID, tenantID)

			cmd := FetchProfilePictureCommand{
				ContactID: contactID,
				Phone:     phone,
				Session:   tc.session,
			}

			// Mock expectations
			contactRepo.On("FindByID", ctx, contactID).Return(testContact, nil)
			wahaService.On("FetchAndUpdateContactProfilePicture", ctx, phone, tc.session).Return(profilePictureURL, nil)
			contactRepo.On("Save", ctx, testContact).Return(nil)
			eventBus.On("Publish", ctx, mock.AnythingOfType("contact.ContactProfilePictureUpdatedEvent")).Return(nil)

			// Act
			err := useCase.Execute(ctx, cmd)

			// Assert
			assert.NoError(t, err)
			contactRepo.AssertExpectations(t)
			wahaService.AssertExpectations(t)
			eventBus.AssertExpectations(t)
		})
	}
}

// ========== Event Publishing Tests ==========

func TestFetchProfilePictureUseCase_Execute_EventContainsCorrectData(t *testing.T) {
	// Arrange
	ctx := context.Background()
	contactRepo, wahaService, eventBus, _, useCase := setupFetchProfilePictureTest(t)

	contactID := uuid.New()
	projectID := uuid.New()
	tenantID := "tenant-1"
	phone := "+5511999999999"
	session := "default"
	profilePictureURL := "https://example.com/profile.jpg"

	testContact := createTestContactForProfilePicture(projectID, tenantID)

	cmd := FetchProfilePictureCommand{
		ContactID: contactID,
		Phone:     phone,
		Session:   session,
	}

	// Mock expectations
	contactRepo.On("FindByID", ctx, contactID).Return(testContact, nil)
	wahaService.On("FetchAndUpdateContactProfilePicture", ctx, phone, session).Return(profilePictureURL, nil)
	contactRepo.On("Save", ctx, testContact).Return(nil)

	var capturedEvent contact.ContactProfilePictureUpdatedEvent
	eventBus.On("Publish", ctx, mock.MatchedBy(func(event contact.DomainEvent) bool {
		if evt, ok := event.(contact.ContactProfilePictureUpdatedEvent); ok {
			capturedEvent = evt
			return true
		}
		return false
	})).Return(nil)

	// Act
	err := useCase.Execute(ctx, cmd)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, contactID, capturedEvent.ContactID)
	assert.Equal(t, tenantID, capturedEvent.TenantID)
	assert.Equal(t, profilePictureURL, capturedEvent.ProfilePictureURL)
	assert.WithinDuration(t, time.Now(), capturedEvent.FetchedAt, 5*time.Second)

	contactRepo.AssertExpectations(t)
	wahaService.AssertExpectations(t)
	eventBus.AssertExpectations(t)
}

func TestFetchProfilePictureUseCase_Execute_MultipleEventPublishErrorsAreLogged(t *testing.T) {
	// Arrange
	ctx := context.Background()
	contactRepo, wahaService, eventBus, _, useCase := setupFetchProfilePictureTest(t)

	contactID := uuid.New()
	projectID := uuid.New()
	tenantID := "tenant-1"
	phone := "+5511999999999"
	session := "default"
	profilePictureURL := "https://example.com/profile.jpg"

	testContact := createTestContactForProfilePicture(projectID, tenantID)

	cmd := FetchProfilePictureCommand{
		ContactID: contactID,
		Phone:     phone,
		Session:   session,
	}

	// Mock expectations
	contactRepo.On("FindByID", ctx, contactID).Return(testContact, nil)
	wahaService.On("FetchAndUpdateContactProfilePicture", ctx, phone, session).Return(profilePictureURL, nil)
	contactRepo.On("Save", ctx, testContact).Return(nil)
	eventBus.On("Publish", ctx, mock.AnythingOfType("contact.ContactProfilePictureUpdatedEvent")).Return(errors.New("RabbitMQ connection lost"))

	// Act
	err := useCase.Execute(ctx, cmd)

	// Assert - should still succeed even if event publishing fails
	assert.NoError(t, err)
	contactRepo.AssertExpectations(t)
	wahaService.AssertExpectations(t)
	eventBus.AssertExpectations(t)
}

// ========== Command Validation Tests ==========

func TestFetchProfilePictureCommand_AllFields(t *testing.T) {
	// Arrange
	contactID := uuid.New()
	phone := "+5511999999999"
	session := "default"

	// Act
	cmd := FetchProfilePictureCommand{
		ContactID: contactID,
		Phone:     phone,
		Session:   session,
	}

	// Assert
	assert.Equal(t, contactID, cmd.ContactID)
	assert.Equal(t, phone, cmd.Phone)
	assert.Equal(t, session, cmd.Session)
}

// ========== Integration-like Tests ==========

func TestFetchProfilePictureUseCase_Execute_SequentialCalls(t *testing.T) {
	// Test that multiple sequential calls work correctly
	ctx := context.Background()
	contactRepo, wahaService, eventBus, _, useCase := setupFetchProfilePictureTest(t)

	contactID1 := uuid.New()
	contactID2 := uuid.New()
	projectID := uuid.New()
	tenantID := "tenant-1"

	testContact1 := createTestContactForProfilePicture(projectID, tenantID)
	testContact2 := createTestContactForProfilePicture(projectID, tenantID)

	// First call
	contactRepo.On("FindByID", ctx, contactID1).Return(testContact1, nil).Once()
	wahaService.On("FetchAndUpdateContactProfilePicture", ctx, "+5511111111111", "default").Return("https://example.com/1.jpg", nil).Once()
	contactRepo.On("Save", ctx, testContact1).Return(nil).Once()
	eventBus.On("Publish", ctx, mock.AnythingOfType("contact.ContactProfilePictureUpdatedEvent")).Return(nil).Once()

	err1 := useCase.Execute(ctx, FetchProfilePictureCommand{
		ContactID: contactID1,
		Phone:     "+5511111111111",
		Session:   "default",
	})

	// Second call
	contactRepo.On("FindByID", ctx, contactID2).Return(testContact2, nil).Once()
	wahaService.On("FetchAndUpdateContactProfilePicture", ctx, "+5522222222222", "default").Return("https://example.com/2.jpg", nil).Once()
	contactRepo.On("Save", ctx, testContact2).Return(nil).Once()
	eventBus.On("Publish", ctx, mock.AnythingOfType("contact.ContactProfilePictureUpdatedEvent")).Return(nil).Once()

	err2 := useCase.Execute(ctx, FetchProfilePictureCommand{
		ContactID: contactID2,
		Phone:     "+5522222222222",
		Session:   "default",
	})

	// Assert
	assert.NoError(t, err1)
	assert.NoError(t, err2)
	contactRepo.AssertExpectations(t)
	wahaService.AssertExpectations(t)
	eventBus.AssertExpectations(t)
}

func TestFetchProfilePictureUseCase_Execute_UpdateExistingProfilePicture(t *testing.T) {
	// Test updating a contact that already has a profile picture
	ctx := context.Background()
	contactRepo, wahaService, eventBus, _, useCase := setupFetchProfilePictureTest(t)

	contactID := uuid.New()
	projectID := uuid.New()
	tenantID := "tenant-1"
	phone := "+5511999999999"
	session := "default"
	oldProfilePictureURL := "https://example.com/old-profile.jpg"
	newProfilePictureURL := "https://example.com/new-profile.jpg"

	testContact := createTestContactForProfilePicture(projectID, tenantID)
	// Set existing profile picture
	testContact.SetProfilePicture(oldProfilePictureURL)

	cmd := FetchProfilePictureCommand{
		ContactID: contactID,
		Phone:     phone,
		Session:   session,
	}

	// Mock expectations
	contactRepo.On("FindByID", ctx, contactID).Return(testContact, nil)
	wahaService.On("FetchAndUpdateContactProfilePicture", ctx, phone, session).Return(newProfilePictureURL, nil)
	contactRepo.On("Save", ctx, testContact).Return(nil)
	eventBus.On("Publish", ctx, mock.MatchedBy(func(event contact.DomainEvent) bool {
		if evt, ok := event.(contact.ContactProfilePictureUpdatedEvent); ok {
			return evt.ProfilePictureURL == newProfilePictureURL
		}
		return false
	})).Return(nil)

	// Act
	err := useCase.Execute(ctx, cmd)

	// Assert
	assert.NoError(t, err)
	contactRepo.AssertExpectations(t)
	wahaService.AssertExpectations(t)
	eventBus.AssertExpectations(t)
}
