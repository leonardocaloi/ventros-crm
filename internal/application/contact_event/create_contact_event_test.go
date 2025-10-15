package contact_event

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/ventros/crm/internal/domain/crm/contact_event"
)

// MockContactEventRepository is a mock implementation of contact_event.Repository
type MockContactEventRepository struct {
	mock.Mock
}

func (m *MockContactEventRepository) Save(ctx context.Context, event *contact_event.ContactEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockContactEventRepository) Update(ctx context.Context, event *contact_event.ContactEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockContactEventRepository) FindByID(ctx context.Context, id uuid.UUID) (*contact_event.ContactEvent, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*contact_event.ContactEvent), args.Error(1)
}

func (m *MockContactEventRepository) FindByContactID(ctx context.Context, contactID uuid.UUID, limit int, offset int) ([]*contact_event.ContactEvent, error) {
	args := m.Called(ctx, contactID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*contact_event.ContactEvent), args.Error(1)
}

func (m *MockContactEventRepository) FindByContactIDVisible(ctx context.Context, contactID uuid.UUID, visibleToClient bool, limit int, offset int) ([]*contact_event.ContactEvent, error) {
	args := m.Called(ctx, contactID, visibleToClient, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*contact_event.ContactEvent), args.Error(1)
}

func (m *MockContactEventRepository) FindBySessionID(ctx context.Context, sessionID uuid.UUID, limit int, offset int) ([]*contact_event.ContactEvent, error) {
	args := m.Called(ctx, sessionID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*contact_event.ContactEvent), args.Error(1)
}

func (m *MockContactEventRepository) FindUndeliveredRealtime(ctx context.Context, limit int) ([]*contact_event.ContactEvent, error) {
	args := m.Called(ctx, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*contact_event.ContactEvent), args.Error(1)
}

func (m *MockContactEventRepository) FindUndeliveredForContact(ctx context.Context, contactID uuid.UUID) ([]*contact_event.ContactEvent, error) {
	args := m.Called(ctx, contactID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*contact_event.ContactEvent), args.Error(1)
}

func (m *MockContactEventRepository) FindByTenantAndType(ctx context.Context, tenantID string, eventType string, since time.Time, limit int) ([]*contact_event.ContactEvent, error) {
	args := m.Called(ctx, tenantID, eventType, since, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*contact_event.ContactEvent), args.Error(1)
}

func (m *MockContactEventRepository) FindByCategory(ctx context.Context, tenantID string, category contact_event.Category, since time.Time, limit int) ([]*contact_event.ContactEvent, error) {
	args := m.Called(ctx, tenantID, category, since, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*contact_event.ContactEvent), args.Error(1)
}

func (m *MockContactEventRepository) FindExpired(ctx context.Context, before time.Time, limit int) ([]*contact_event.ContactEvent, error) {
	args := m.Called(ctx, before, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*contact_event.ContactEvent), args.Error(1)
}

func (m *MockContactEventRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockContactEventRepository) DeleteExpired(ctx context.Context, before time.Time) (int, error) {
	args := m.Called(ctx, before)
	return args.Int(0), args.Error(1)
}

func (m *MockContactEventRepository) CountByContact(ctx context.Context, contactID uuid.UUID) (int, error) {
	args := m.Called(ctx, contactID)
	return args.Int(0), args.Error(1)
}

// Test NewCreateContactEventUseCase
func TestNewCreateContactEventUseCase(t *testing.T) {
	// Arrange
	repo := new(MockContactEventRepository)

	// Act
	useCase := NewCreateContactEventUseCase(repo)

	// Assert
	assert.NotNil(t, useCase)
	assert.Equal(t, repo, useCase.repo)
}

// Test Execute - Success with minimal required fields
func TestCreateContactEventUseCase_Execute_Success_MinimalFields(t *testing.T) {
	// Arrange
	ctx := context.Background()
	repo := new(MockContactEventRepository)
	useCase := NewCreateContactEventUseCase(repo)

	contactID := uuid.New()
	tenantID := "tenant-123"
	eventType := "test_event"

	cmd := CreateContactEventCommand{
		ContactID: contactID,
		TenantID:  tenantID,
		EventType: eventType,
		Category:  contact_event.CategoryGeneral,
		Priority:  contact_event.PriorityNormal,
		Source:    contact_event.SourceSystem,
	}

	// Mock repository
	repo.On("Save", ctx, mock.AnythingOfType("*contact_event.ContactEvent")).Return(nil)

	// Act
	result, err := useCase.Execute(ctx, cmd)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEqual(t, uuid.Nil, result.ID())
	assert.Equal(t, contactID, result.ContactID())
	assert.Equal(t, tenantID, result.TenantID())
	assert.Equal(t, eventType, result.EventType())
	assert.Equal(t, contact_event.CategoryGeneral, result.Category())
	assert.Equal(t, contact_event.PriorityNormal, result.Priority())
	assert.Equal(t, contact_event.SourceSystem, result.Source())
	assert.Nil(t, result.SessionID())
	assert.Nil(t, result.Title())
	assert.Nil(t, result.Description())

	repo.AssertExpectations(t)
}

// Test Execute - Success with all optional fields
func TestCreateContactEventUseCase_Execute_Success_AllFields(t *testing.T) {
	// Arrange
	ctx := context.Background()
	repo := new(MockContactEventRepository)
	useCase := NewCreateContactEventUseCase(repo)

	contactID := uuid.New()
	sessionID := uuid.New()
	tenantID := "tenant-123"
	eventType := "test_event"
	title := "Test Title"
	description := "Test Description"
	triggeredBy := uuid.New()
	integrationSource := "test-integration"

	cmd := CreateContactEventCommand{
		ContactID:         contactID,
		SessionID:         &sessionID,
		TenantID:          tenantID,
		EventType:         eventType,
		Category:          contact_event.CategoryPipeline,
		Priority:          contact_event.PriorityHigh,
		Source:            contact_event.SourceAgent,
		Title:             &title,
		Description:       &description,
		Payload:           map[string]interface{}{"key1": "value1", "key2": 123},
		Metadata:          map[string]interface{}{"meta1": "metavalue1", "meta2": true},
		TriggeredBy:       &triggeredBy,
		IntegrationSource: &integrationSource,
		IsRealtime:        false,
		VisibleToClient:   false,
		VisibleToAgent:    true,
	}

	// Mock repository
	repo.On("Save", ctx, mock.AnythingOfType("*contact_event.ContactEvent")).Return(nil)

	// Act
	result, err := useCase.Execute(ctx, cmd)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, contactID, result.ContactID())
	assert.NotNil(t, result.SessionID())
	assert.Equal(t, sessionID, *result.SessionID())
	assert.Equal(t, tenantID, result.TenantID())
	assert.Equal(t, eventType, result.EventType())
	assert.Equal(t, contact_event.CategoryPipeline, result.Category())
	assert.Equal(t, contact_event.PriorityHigh, result.Priority())
	assert.Equal(t, contact_event.SourceAgent, result.Source())
	assert.NotNil(t, result.Title())
	assert.Equal(t, title, *result.Title())
	assert.NotNil(t, result.Description())
	assert.Equal(t, description, *result.Description())
	assert.NotNil(t, result.TriggeredBy())
	assert.Equal(t, triggeredBy, *result.TriggeredBy())
	assert.NotNil(t, result.IntegrationSource())
	assert.Equal(t, integrationSource, *result.IntegrationSource())

	// Check payload
	payload := result.Payload()
	assert.Equal(t, "value1", payload["key1"])
	assert.Equal(t, 123, payload["key2"])

	// Check metadata
	metadata := result.Metadata()
	assert.Equal(t, "metavalue1", metadata["meta1"])
	assert.Equal(t, true, metadata["meta2"])

	repo.AssertExpectations(t)
}

// Test Execute - Missing ContactID
func TestCreateContactEventUseCase_Execute_MissingContactID(t *testing.T) {
	// Arrange
	ctx := context.Background()
	repo := new(MockContactEventRepository)
	useCase := NewCreateContactEventUseCase(repo)

	cmd := CreateContactEventCommand{
		ContactID: uuid.Nil,
		TenantID:  "tenant-123",
		EventType: "test_event",
		Category:  contact_event.CategoryGeneral,
		Priority:  contact_event.PriorityNormal,
		Source:    contact_event.SourceSystem,
	}

	// Act
	result, err := useCase.Execute(ctx, cmd)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "contact_id is required")

	repo.AssertNotCalled(t, "Save")
}

// Test Execute - Missing TenantID
func TestCreateContactEventUseCase_Execute_MissingTenantID(t *testing.T) {
	// Arrange
	ctx := context.Background()
	repo := new(MockContactEventRepository)
	useCase := NewCreateContactEventUseCase(repo)

	cmd := CreateContactEventCommand{
		ContactID: uuid.New(),
		TenantID:  "",
		EventType: "test_event",
		Category:  contact_event.CategoryGeneral,
		Priority:  contact_event.PriorityNormal,
		Source:    contact_event.SourceSystem,
	}

	// Act
	result, err := useCase.Execute(ctx, cmd)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "tenant_id is required")

	repo.AssertNotCalled(t, "Save")
}

// Test Execute - Missing EventType
func TestCreateContactEventUseCase_Execute_MissingEventType(t *testing.T) {
	// Arrange
	ctx := context.Background()
	repo := new(MockContactEventRepository)
	useCase := NewCreateContactEventUseCase(repo)

	cmd := CreateContactEventCommand{
		ContactID: uuid.New(),
		TenantID:  "tenant-123",
		EventType: "",
		Category:  contact_event.CategoryGeneral,
		Priority:  contact_event.PriorityNormal,
		Source:    contact_event.SourceSystem,
	}

	// Act
	result, err := useCase.Execute(ctx, cmd)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "event_type is required")

	repo.AssertNotCalled(t, "Save")
}

// Test Execute - Invalid Category
func TestCreateContactEventUseCase_Execute_InvalidCategory(t *testing.T) {
	// Arrange
	ctx := context.Background()
	repo := new(MockContactEventRepository)
	useCase := NewCreateContactEventUseCase(repo)

	cmd := CreateContactEventCommand{
		ContactID: uuid.New(),
		TenantID:  "tenant-123",
		EventType: "test_event",
		Category:  contact_event.Category("invalid_category"),
		Priority:  contact_event.PriorityNormal,
		Source:    contact_event.SourceSystem,
	}

	// Act
	result, err := useCase.Execute(ctx, cmd)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to create contact event")

	repo.AssertNotCalled(t, "Save")
}

// Test Execute - Invalid Priority
func TestCreateContactEventUseCase_Execute_InvalidPriority(t *testing.T) {
	// Arrange
	ctx := context.Background()
	repo := new(MockContactEventRepository)
	useCase := NewCreateContactEventUseCase(repo)

	cmd := CreateContactEventCommand{
		ContactID: uuid.New(),
		TenantID:  "tenant-123",
		EventType: "test_event",
		Category:  contact_event.CategoryGeneral,
		Priority:  contact_event.Priority("invalid_priority"),
		Source:    contact_event.SourceSystem,
	}

	// Act
	result, err := useCase.Execute(ctx, cmd)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to create contact event")

	repo.AssertNotCalled(t, "Save")
}

// Test Execute - Invalid Source
func TestCreateContactEventUseCase_Execute_InvalidSource(t *testing.T) {
	// Arrange
	ctx := context.Background()
	repo := new(MockContactEventRepository)
	useCase := NewCreateContactEventUseCase(repo)

	cmd := CreateContactEventCommand{
		ContactID: uuid.New(),
		TenantID:  "tenant-123",
		EventType: "test_event",
		Category:  contact_event.CategoryGeneral,
		Priority:  contact_event.PriorityNormal,
		Source:    contact_event.Source("invalid_source"),
	}

	// Act
	result, err := useCase.Execute(ctx, cmd)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to create contact event")

	repo.AssertNotCalled(t, "Save")
}

// Test Execute - Invalid SessionID (nil UUID)
func TestCreateContactEventUseCase_Execute_InvalidSessionID(t *testing.T) {
	// Arrange
	ctx := context.Background()
	repo := new(MockContactEventRepository)
	useCase := NewCreateContactEventUseCase(repo)

	nilSessionID := uuid.Nil

	cmd := CreateContactEventCommand{
		ContactID: uuid.New(),
		SessionID: &nilSessionID,
		TenantID:  "tenant-123",
		EventType: "test_event",
		Category:  contact_event.CategoryGeneral,
		Priority:  contact_event.PriorityNormal,
		Source:    contact_event.SourceSystem,
	}

	// Act
	result, err := useCase.Execute(ctx, cmd)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to attach session")

	repo.AssertNotCalled(t, "Save")
}

// Test Execute - Invalid TriggeredBy (nil UUID)
func TestCreateContactEventUseCase_Execute_InvalidTriggeredBy(t *testing.T) {
	// Arrange
	ctx := context.Background()
	repo := new(MockContactEventRepository)
	useCase := NewCreateContactEventUseCase(repo)

	nilTriggeredBy := uuid.Nil

	cmd := CreateContactEventCommand{
		ContactID:   uuid.New(),
		TenantID:    "tenant-123",
		EventType:   "test_event",
		Category:    contact_event.CategoryGeneral,
		Priority:    contact_event.PriorityNormal,
		Source:      contact_event.SourceSystem,
		TriggeredBy: &nilTriggeredBy,
	}

	// Act
	result, err := useCase.Execute(ctx, cmd)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to set triggered_by")

	repo.AssertNotCalled(t, "Save")
}

// Test Execute - Repository Save Error
func TestCreateContactEventUseCase_Execute_RepositorySaveError(t *testing.T) {
	// Arrange
	ctx := context.Background()
	repo := new(MockContactEventRepository)
	useCase := NewCreateContactEventUseCase(repo)

	cmd := CreateContactEventCommand{
		ContactID: uuid.New(),
		TenantID:  "tenant-123",
		EventType: "test_event",
		Category:  contact_event.CategoryGeneral,
		Priority:  contact_event.PriorityNormal,
		Source:    contact_event.SourceSystem,
	}

	// Mock repository error
	saveError := errors.New("database connection failed")
	repo.On("Save", ctx, mock.AnythingOfType("*contact_event.ContactEvent")).Return(saveError)

	// Act
	result, err := useCase.Execute(ctx, cmd)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to save contact event")
	assert.Contains(t, err.Error(), "database connection failed")

	repo.AssertExpectations(t)
}

// Test Execute - Success with different categories
func TestCreateContactEventUseCase_Execute_Success_DifferentCategories(t *testing.T) {
	categories := []contact_event.Category{
		contact_event.CategoryGeneral,
		contact_event.CategoryStatus,
		contact_event.CategoryPipeline,
		contact_event.CategoryAssignment,
		contact_event.CategoryTag,
		contact_event.CategoryNote,
		contact_event.CategorySession,
		contact_event.CategoryCustomField,
		contact_event.CategorySystem,
		contact_event.CategoryNotification,
		contact_event.CategoryTracking,
	}

	for _, category := range categories {
		t.Run(string(category), func(t *testing.T) {
			// Arrange
			ctx := context.Background()
			repo := new(MockContactEventRepository)
			useCase := NewCreateContactEventUseCase(repo)

			cmd := CreateContactEventCommand{
				ContactID: uuid.New(),
				TenantID:  "tenant-123",
				EventType: "test_event",
				Category:  category,
				Priority:  contact_event.PriorityNormal,
				Source:    contact_event.SourceSystem,
			}

			repo.On("Save", ctx, mock.AnythingOfType("*contact_event.ContactEvent")).Return(nil)

			// Act
			result, err := useCase.Execute(ctx, cmd)

			// Assert
			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, category, result.Category())

			repo.AssertExpectations(t)
		})
	}
}

// Test Execute - Success with different priorities
func TestCreateContactEventUseCase_Execute_Success_DifferentPriorities(t *testing.T) {
	priorities := []contact_event.Priority{
		contact_event.PriorityLow,
		contact_event.PriorityNormal,
		contact_event.PriorityHigh,
		contact_event.PriorityUrgent,
	}

	for _, priority := range priorities {
		t.Run(string(priority), func(t *testing.T) {
			// Arrange
			ctx := context.Background()
			repo := new(MockContactEventRepository)
			useCase := NewCreateContactEventUseCase(repo)

			cmd := CreateContactEventCommand{
				ContactID: uuid.New(),
				TenantID:  "tenant-123",
				EventType: "test_event",
				Category:  contact_event.CategoryGeneral,
				Priority:  priority,
				Source:    contact_event.SourceSystem,
			}

			repo.On("Save", ctx, mock.AnythingOfType("*contact_event.ContactEvent")).Return(nil)

			// Act
			result, err := useCase.Execute(ctx, cmd)

			// Assert
			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, priority, result.Priority())

			repo.AssertExpectations(t)
		})
	}
}

// Test Execute - Success with different sources
func TestCreateContactEventUseCase_Execute_Success_DifferentSources(t *testing.T) {
	sources := []contact_event.Source{
		contact_event.SourceSystem,
		contact_event.SourceAgent,
		contact_event.SourceWebhook,
		contact_event.SourceWorkflow,
		contact_event.SourceAutomation,
		contact_event.SourceIntegration,
	}

	for _, source := range sources {
		t.Run(string(source), func(t *testing.T) {
			// Arrange
			ctx := context.Background()
			repo := new(MockContactEventRepository)
			useCase := NewCreateContactEventUseCase(repo)

			cmd := CreateContactEventCommand{
				ContactID: uuid.New(),
				TenantID:  "tenant-123",
				EventType: "test_event",
				Category:  contact_event.CategoryGeneral,
				Priority:  contact_event.PriorityNormal,
				Source:    source,
			}

			repo.On("Save", ctx, mock.AnythingOfType("*contact_event.ContactEvent")).Return(nil)

			// Act
			result, err := useCase.Execute(ctx, cmd)

			// Assert
			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, source, result.Source())

			repo.AssertExpectations(t)
		})
	}
}

// Test Execute - Empty Payload
func TestCreateContactEventUseCase_Execute_Success_EmptyPayload(t *testing.T) {
	// Arrange
	ctx := context.Background()
	repo := new(MockContactEventRepository)
	useCase := NewCreateContactEventUseCase(repo)

	cmd := CreateContactEventCommand{
		ContactID: uuid.New(),
		TenantID:  "tenant-123",
		EventType: "test_event",
		Category:  contact_event.CategoryGeneral,
		Priority:  contact_event.PriorityNormal,
		Source:    contact_event.SourceSystem,
		Payload:   map[string]interface{}{},
	}

	repo.On("Save", ctx, mock.AnythingOfType("*contact_event.ContactEvent")).Return(nil)

	// Act
	result, err := useCase.Execute(ctx, cmd)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Empty(t, result.Payload())

	repo.AssertExpectations(t)
}

// Test Execute - Empty Metadata
func TestCreateContactEventUseCase_Execute_Success_EmptyMetadata(t *testing.T) {
	// Arrange
	ctx := context.Background()
	repo := new(MockContactEventRepository)
	useCase := NewCreateContactEventUseCase(repo)

	cmd := CreateContactEventCommand{
		ContactID: uuid.New(),
		TenantID:  "tenant-123",
		EventType: "test_event",
		Category:  contact_event.CategoryGeneral,
		Priority:  contact_event.PriorityNormal,
		Source:    contact_event.SourceSystem,
		Metadata:  map[string]interface{}{},
	}

	repo.On("Save", ctx, mock.AnythingOfType("*contact_event.ContactEvent")).Return(nil)

	// Act
	result, err := useCase.Execute(ctx, cmd)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Empty(t, result.Metadata())

	repo.AssertExpectations(t)
}

// Test Execute - Nil Payload
func TestCreateContactEventUseCase_Execute_Success_NilPayload(t *testing.T) {
	// Arrange
	ctx := context.Background()
	repo := new(MockContactEventRepository)
	useCase := NewCreateContactEventUseCase(repo)

	cmd := CreateContactEventCommand{
		ContactID: uuid.New(),
		TenantID:  "tenant-123",
		EventType: "test_event",
		Category:  contact_event.CategoryGeneral,
		Priority:  contact_event.PriorityNormal,
		Source:    contact_event.SourceSystem,
		Payload:   nil,
	}

	repo.On("Save", ctx, mock.AnythingOfType("*contact_event.ContactEvent")).Return(nil)

	// Act
	result, err := useCase.Execute(ctx, cmd)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Empty(t, result.Payload())

	repo.AssertExpectations(t)
}

// Test Execute - Nil Metadata
func TestCreateContactEventUseCase_Execute_Success_NilMetadata(t *testing.T) {
	// Arrange
	ctx := context.Background()
	repo := new(MockContactEventRepository)
	useCase := NewCreateContactEventUseCase(repo)

	cmd := CreateContactEventCommand{
		ContactID: uuid.New(),
		TenantID:  "tenant-123",
		EventType: "test_event",
		Category:  contact_event.CategoryGeneral,
		Priority:  contact_event.PriorityNormal,
		Source:    contact_event.SourceSystem,
		Metadata:  nil,
	}

	repo.On("Save", ctx, mock.AnythingOfType("*contact_event.ContactEvent")).Return(nil)

	// Act
	result, err := useCase.Execute(ctx, cmd)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Empty(t, result.Metadata())

	repo.AssertExpectations(t)
}

// Test Execute - Visibility settings
func TestCreateContactEventUseCase_Execute_Success_VisibilitySettings(t *testing.T) {
	testCases := []struct {
		name            string
		visibleToClient bool
		visibleToAgent  bool
	}{
		{"both_visible", true, true},
		{"client_only", true, false},
		{"agent_only", false, true},
		{"none_visible", false, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			ctx := context.Background()
			repo := new(MockContactEventRepository)
			useCase := NewCreateContactEventUseCase(repo)

			cmd := CreateContactEventCommand{
				ContactID:       uuid.New(),
				TenantID:        "tenant-123",
				EventType:       "test_event",
				Category:        contact_event.CategoryGeneral,
				Priority:        contact_event.PriorityNormal,
				Source:          contact_event.SourceSystem,
				VisibleToClient: tc.visibleToClient,
				VisibleToAgent:  tc.visibleToAgent,
			}

			repo.On("Save", ctx, mock.AnythingOfType("*contact_event.ContactEvent")).Return(nil)

			// Act
			result, err := useCase.Execute(ctx, cmd)

			// Assert
			assert.NoError(t, err)
			assert.NotNil(t, result)

			repo.AssertExpectations(t)
		})
	}
}

// Test Execute - IsRealtime settings
func TestCreateContactEventUseCase_Execute_Success_RealtimeSettings(t *testing.T) {
	testCases := []struct {
		name       string
		isRealtime bool
	}{
		{"realtime_enabled", true},
		{"realtime_disabled", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			ctx := context.Background()
			repo := new(MockContactEventRepository)
			useCase := NewCreateContactEventUseCase(repo)

			cmd := CreateContactEventCommand{
				ContactID:  uuid.New(),
				TenantID:   "tenant-123",
				EventType:  "test_event",
				Category:   contact_event.CategoryGeneral,
				Priority:   contact_event.PriorityNormal,
				Source:     contact_event.SourceSystem,
				IsRealtime: tc.isRealtime,
			}

			repo.On("Save", ctx, mock.AnythingOfType("*contact_event.ContactEvent")).Return(nil)

			// Act
			result, err := useCase.Execute(ctx, cmd)

			// Assert
			assert.NoError(t, err)
			assert.NotNil(t, result)

			repo.AssertExpectations(t)
		})
	}
}

// Test Execute - Complex Payload with nested structures
func TestCreateContactEventUseCase_Execute_Success_ComplexPayload(t *testing.T) {
	// Arrange
	ctx := context.Background()
	repo := new(MockContactEventRepository)
	useCase := NewCreateContactEventUseCase(repo)

	cmd := CreateContactEventCommand{
		ContactID: uuid.New(),
		TenantID:  "tenant-123",
		EventType: "test_event",
		Category:  contact_event.CategoryGeneral,
		Priority:  contact_event.PriorityNormal,
		Source:    contact_event.SourceSystem,
		Payload: map[string]interface{}{
			"string":  "value",
			"number":  123,
			"boolean": true,
			"array":   []string{"item1", "item2", "item3"},
			"nested": map[string]interface{}{
				"key1": "nested_value1",
				"key2": 456,
			},
			"null": nil,
		},
	}

	repo.On("Save", ctx, mock.AnythingOfType("*contact_event.ContactEvent")).Return(nil)

	// Act
	result, err := useCase.Execute(ctx, cmd)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)

	payload := result.Payload()
	assert.Equal(t, "value", payload["string"])
	assert.Equal(t, 123, payload["number"])
	assert.Equal(t, true, payload["boolean"])
	assert.NotNil(t, payload["array"])
	assert.NotNil(t, payload["nested"])
	assert.Nil(t, payload["null"])

	repo.AssertExpectations(t)
}

// Test Execute - Context cancellation
func TestCreateContactEventUseCase_Execute_ContextCancellation(t *testing.T) {
	// Arrange
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	repo := new(MockContactEventRepository)
	useCase := NewCreateContactEventUseCase(repo)

	cmd := CreateContactEventCommand{
		ContactID: uuid.New(),
		TenantID:  "tenant-123",
		EventType: "test_event",
		Category:  contact_event.CategoryGeneral,
		Priority:  contact_event.PriorityNormal,
		Source:    contact_event.SourceSystem,
	}

	// Mock repository to return context error
	repo.On("Save", ctx, mock.AnythingOfType("*contact_event.ContactEvent")).
		Return(context.Canceled)

	// Act
	result, err := useCase.Execute(ctx, cmd)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)

	repo.AssertExpectations(t)
}

// Test Execute - Multiple payload fields
func TestCreateContactEventUseCase_Execute_Success_MultiplePayloadFields(t *testing.T) {
	// Arrange
	ctx := context.Background()
	repo := new(MockContactEventRepository)
	useCase := NewCreateContactEventUseCase(repo)

	cmd := CreateContactEventCommand{
		ContactID: uuid.New(),
		TenantID:  "tenant-123",
		EventType: "test_event",
		Category:  contact_event.CategoryGeneral,
		Priority:  contact_event.PriorityNormal,
		Source:    contact_event.SourceSystem,
		Payload: map[string]interface{}{
			"field1":  "value1",
			"field2":  "value2",
			"field3":  "value3",
			"field4":  123,
			"field5":  456.78,
			"field6":  true,
			"field7":  false,
			"field8":  nil,
			"field9":  []int{1, 2, 3},
			"field10": map[string]string{"nested": "value"},
		},
	}

	repo.On("Save", ctx, mock.AnythingOfType("*contact_event.ContactEvent")).Return(nil)

	// Act
	result, err := useCase.Execute(ctx, cmd)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)

	payload := result.Payload()
	assert.Len(t, payload, 10)
	assert.Equal(t, "value1", payload["field1"])
	assert.Equal(t, "value2", payload["field2"])
	assert.Equal(t, "value3", payload["field3"])

	repo.AssertExpectations(t)
}

// Test Execute - Multiple metadata fields
func TestCreateContactEventUseCase_Execute_Success_MultipleMetadataFields(t *testing.T) {
	// Arrange
	ctx := context.Background()
	repo := new(MockContactEventRepository)
	useCase := NewCreateContactEventUseCase(repo)

	cmd := CreateContactEventCommand{
		ContactID: uuid.New(),
		TenantID:  "tenant-123",
		EventType: "test_event",
		Category:  contact_event.CategoryGeneral,
		Priority:  contact_event.PriorityNormal,
		Source:    contact_event.SourceSystem,
		Metadata: map[string]interface{}{
			"user_agent": "Mozilla/5.0",
			"ip_address": "192.168.1.1",
			"timestamp":  1234567890,
			"version":    "1.0.0",
		},
	}

	repo.On("Save", ctx, mock.AnythingOfType("*contact_event.ContactEvent")).Return(nil)

	// Act
	result, err := useCase.Execute(ctx, cmd)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)

	metadata := result.Metadata()
	assert.Len(t, metadata, 4)
	assert.Equal(t, "Mozilla/5.0", metadata["user_agent"])
	assert.Equal(t, "192.168.1.1", metadata["ip_address"])

	repo.AssertExpectations(t)
}
