package session

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ventros/crm/internal/domain/crm/channel"
	"github.com/ventros/crm/internal/domain/crm/pipeline"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockChannelRepository mocks the channel repository
type MockChannelRepository struct {
	mock.Mock
}

func (m *MockChannelRepository) Create(ch *channel.Channel) error {
	args := m.Called(ch)
	return args.Error(0)
}

func (m *MockChannelRepository) GetByID(id uuid.UUID) (*channel.Channel, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*channel.Channel), args.Error(1)
}

func (m *MockChannelRepository) GetByUserID(userID uuid.UUID) ([]*channel.Channel, error) {
	args := m.Called(userID)
	return args.Get(0).([]*channel.Channel), args.Error(1)
}

func (m *MockChannelRepository) GetByProjectID(projectID uuid.UUID) ([]*channel.Channel, error) {
	args := m.Called(projectID)
	return args.Get(0).([]*channel.Channel), args.Error(1)
}

func (m *MockChannelRepository) GetByExternalID(externalID string) (*channel.Channel, error) {
	args := m.Called(externalID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*channel.Channel), args.Error(1)
}

func (m *MockChannelRepository) GetByWebhookID(webhookID string) (*channel.Channel, error) {
	args := m.Called(webhookID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*channel.Channel), args.Error(1)
}

func (m *MockChannelRepository) Update(ch *channel.Channel) error {
	args := m.Called(ch)
	return args.Error(0)
}

func (m *MockChannelRepository) Delete(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockChannelRepository) GetActiveWAHAChannels() ([]*channel.Channel, error) {
	args := m.Called()
	return args.Get(0).([]*channel.Channel), args.Error(1)
}

// MockPipelineRepository mocks the pipeline repository
type MockPipelineRepository struct {
	mock.Mock
}

func (m *MockPipelineRepository) SavePipeline(ctx context.Context, pipe *pipeline.Pipeline) error {
	args := m.Called(ctx, pipe)
	return args.Error(0)
}

func (m *MockPipelineRepository) FindPipelineByID(ctx context.Context, id uuid.UUID) (*pipeline.Pipeline, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pipeline.Pipeline), args.Error(1)
}

func (m *MockPipelineRepository) FindPipelinesByProject(ctx context.Context, projectID uuid.UUID) ([]*pipeline.Pipeline, error) {
	args := m.Called(ctx, projectID)
	return args.Get(0).([]*pipeline.Pipeline), args.Error(1)
}

func (m *MockPipelineRepository) FindPipelinesByTenant(ctx context.Context, tenantID string) ([]*pipeline.Pipeline, error) {
	args := m.Called(ctx, tenantID)
	return args.Get(0).([]*pipeline.Pipeline), args.Error(1)
}

func (m *MockPipelineRepository) FindActivePipelinesByProject(ctx context.Context, projectID uuid.UUID) ([]*pipeline.Pipeline, error) {
	args := m.Called(ctx, projectID)
	return args.Get(0).([]*pipeline.Pipeline), args.Error(1)
}

func (m *MockPipelineRepository) DeletePipeline(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockPipelineRepository) SaveStatus(ctx context.Context, status *pipeline.Status) error {
	args := m.Called(ctx, status)
	return args.Error(0)
}

func (m *MockPipelineRepository) FindStatusByID(ctx context.Context, id uuid.UUID) (*pipeline.Status, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pipeline.Status), args.Error(1)
}

func (m *MockPipelineRepository) FindStatusesByPipeline(ctx context.Context, pipelineID uuid.UUID) ([]*pipeline.Status, error) {
	args := m.Called(ctx, pipelineID)
	return args.Get(0).([]*pipeline.Status), args.Error(1)
}

func (m *MockPipelineRepository) FindActiveStatusesByPipeline(ctx context.Context, pipelineID uuid.UUID) ([]*pipeline.Status, error) {
	args := m.Called(ctx, pipelineID)
	return args.Get(0).([]*pipeline.Status), args.Error(1)
}

func (m *MockPipelineRepository) DeleteStatus(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockPipelineRepository) AddStatusToPipeline(ctx context.Context, pipelineID, statusID uuid.UUID) error {
	args := m.Called(ctx, pipelineID, statusID)
	return args.Error(0)
}

func (m *MockPipelineRepository) RemoveStatusFromPipeline(ctx context.Context, pipelineID, statusID uuid.UUID) error {
	args := m.Called(ctx, pipelineID, statusID)
	return args.Error(0)
}

func (m *MockPipelineRepository) GetPipelineWithStatuses(ctx context.Context, pipelineID uuid.UUID) (*pipeline.Pipeline, []*pipeline.Status, error) {
	args := m.Called(ctx, pipelineID)
	if args.Get(0) == nil {
		return nil, nil, args.Error(2)
	}
	return args.Get(0).(*pipeline.Pipeline), args.Get(1).([]*pipeline.Status), args.Error(2)
}

func (m *MockPipelineRepository) SetContactStatus(ctx context.Context, contactID, pipelineID, statusID uuid.UUID) error {
	args := m.Called(ctx, contactID, pipelineID, statusID)
	return args.Error(0)
}

func (m *MockPipelineRepository) GetContactStatus(ctx context.Context, contactID, pipelineID uuid.UUID) (*pipeline.Status, error) {
	args := m.Called(ctx, contactID, pipelineID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pipeline.Status), args.Error(1)
}

func (m *MockPipelineRepository) GetContactsByStatus(ctx context.Context, pipelineID, statusID uuid.UUID) ([]uuid.UUID, error) {
	args := m.Called(ctx, pipelineID, statusID)
	return args.Get(0).([]uuid.UUID), args.Error(1)
}

func (m *MockPipelineRepository) GetContactStatusHistory(ctx context.Context, contactID, pipelineID uuid.UUID) ([]*pipeline.ContactStatusHistory, error) {
	args := m.Called(ctx, contactID, pipelineID)
	return args.Get(0).([]*pipeline.ContactStatusHistory), args.Error(1)
}

func (m *MockPipelineRepository) FindByTenantWithFilters(ctx context.Context, filters pipeline.PipelineFilters) ([]*pipeline.Pipeline, int64, error) {
	args := m.Called(ctx, filters)
	return args.Get(0).([]*pipeline.Pipeline), args.Get(1).(int64), args.Error(2)
}

func (m *MockPipelineRepository) SearchByText(ctx context.Context, tenantID string, searchText string, limit int, offset int) ([]*pipeline.Pipeline, int64, error) {
	args := m.Called(ctx, tenantID, searchText, limit, offset)
	return args.Get(0).([]*pipeline.Pipeline), args.Get(1).(int64), args.Error(2)
}

// Test Helpers

func createTestChannel(t *testing.T, pipelineID *uuid.UUID, defaultTimeout int) *channel.Channel {
	t.Helper()
	ch, err := channel.NewChannel(
		uuid.New(),
		uuid.New(),
		"test-tenant",
		"Test Channel",
		channel.TypeWAHA,
	)
	require.NoError(t, err)
	ch.PipelineID = pipelineID
	ch.DefaultSessionTimeoutMinutes = defaultTimeout
	return ch
}

func createTestPipeline(t *testing.T, sessionTimeoutMinutes *int) *pipeline.Pipeline {
	t.Helper()
	pipe, err := pipeline.NewPipeline(
		uuid.New(),
		"test-tenant",
		"Test Pipeline",
	)
	require.NoError(t, err)

	if sessionTimeoutMinutes != nil {
		err = pipe.SetSessionTimeout(sessionTimeoutMinutes)
		require.NoError(t, err)
	}

	return pipe
}

// Tests

func TestNewSessionTimeoutResolver(t *testing.T) {
	// Arrange
	channelRepo := new(MockChannelRepository)
	pipelineRepo := new(MockPipelineRepository)

	// Act
	resolver := NewSessionTimeoutResolver(channelRepo, pipelineRepo)

	// Assert
	assert.NotNil(t, resolver)
	assert.Equal(t, 30*time.Minute, resolver.systemDefault)
	assert.NotNil(t, resolver.channelRepo)
	assert.NotNil(t, resolver.pipelineRepo)
}

func TestResolveForChannel_WithPipelineTimeout(t *testing.T) {
	// Arrange
	channelRepo := new(MockChannelRepository)
	pipelineRepo := new(MockPipelineRepository)
	resolver := NewSessionTimeoutResolver(channelRepo, pipelineRepo)
	ctx := context.Background()

	pipelineID := uuid.New()
	channelID := uuid.New()
	pipelineTimeout := 60

	ch := createTestChannel(t, &pipelineID, 45)
	pipe := createTestPipeline(t, &pipelineTimeout)

	channelRepo.On("GetByID", channelID).Return(ch, nil)
	pipelineRepo.On("FindPipelineByID", ctx, pipelineID).Return(pipe, nil)

	// Act
	timeout, returnedPipelineID, err := resolver.ResolveForChannel(ctx, channelID)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, 60*time.Minute, timeout)
	assert.NotNil(t, returnedPipelineID)
	assert.Equal(t, pipelineID, *returnedPipelineID)
	channelRepo.AssertExpectations(t)
	pipelineRepo.AssertExpectations(t)
}

func TestResolveForChannel_WithPipelineNotFound(t *testing.T) {
	// Arrange
	channelRepo := new(MockChannelRepository)
	pipelineRepo := new(MockPipelineRepository)
	resolver := NewSessionTimeoutResolver(channelRepo, pipelineRepo)
	ctx := context.Background()

	pipelineID := uuid.New()
	channelID := uuid.New()

	ch := createTestChannel(t, &pipelineID, 45)

	channelRepo.On("GetByID", channelID).Return(ch, nil)
	pipelineRepo.On("FindPipelineByID", ctx, pipelineID).Return(nil, errors.New("pipeline not found"))

	// Act
	timeout, returnedPipelineID, err := resolver.ResolveForChannel(ctx, channelID)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, 45*time.Minute, timeout) // Falls back to channel timeout
	assert.Nil(t, returnedPipelineID)
	channelRepo.AssertExpectations(t)
	pipelineRepo.AssertExpectations(t)
}

func TestResolveForChannel_WithPipelineTimeoutNil(t *testing.T) {
	// Arrange
	channelRepo := new(MockChannelRepository)
	pipelineRepo := new(MockPipelineRepository)
	resolver := NewSessionTimeoutResolver(channelRepo, pipelineRepo)
	ctx := context.Background()

	pipelineID := uuid.New()
	channelID := uuid.New()

	ch := createTestChannel(t, &pipelineID, 45)
	pipe := createTestPipeline(t, nil) // No timeout set

	channelRepo.On("GetByID", channelID).Return(ch, nil)
	pipelineRepo.On("FindPipelineByID", ctx, pipelineID).Return(pipe, nil)

	// Act
	timeout, returnedPipelineID, err := resolver.ResolveForChannel(ctx, channelID)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, 45*time.Minute, timeout) // Falls back to channel timeout
	assert.Nil(t, returnedPipelineID)
	channelRepo.AssertExpectations(t)
	pipelineRepo.AssertExpectations(t)
}

func TestResolveForChannel_WithChannelTimeout(t *testing.T) {
	// Arrange
	channelRepo := new(MockChannelRepository)
	pipelineRepo := new(MockPipelineRepository)
	resolver := NewSessionTimeoutResolver(channelRepo, pipelineRepo)
	ctx := context.Background()

	channelID := uuid.New()

	ch := createTestChannel(t, nil, 45) // No pipeline, has channel timeout

	channelRepo.On("GetByID", channelID).Return(ch, nil)

	// Act
	timeout, returnedPipelineID, err := resolver.ResolveForChannel(ctx, channelID)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, 45*time.Minute, timeout)
	assert.Nil(t, returnedPipelineID)
	channelRepo.AssertExpectations(t)
	pipelineRepo.AssertExpectations(t)
}

func TestResolveForChannel_ChannelNotFound(t *testing.T) {
	// Arrange
	channelRepo := new(MockChannelRepository)
	pipelineRepo := new(MockPipelineRepository)
	resolver := NewSessionTimeoutResolver(channelRepo, pipelineRepo)
	ctx := context.Background()

	channelID := uuid.New()

	channelRepo.On("GetByID", channelID).Return(nil, errors.New("channel not found"))

	// Act
	timeout, returnedPipelineID, err := resolver.ResolveForChannel(ctx, channelID)

	// Assert
	assert.NoError(t, err)                   // No error, just returns system default
	assert.Equal(t, 30*time.Minute, timeout) // System default
	assert.Nil(t, returnedPipelineID)
	channelRepo.AssertExpectations(t)
	pipelineRepo.AssertExpectations(t)
}

func TestResolveForChannel_WithNoTimeoutConfigured(t *testing.T) {
	// Arrange
	channelRepo := new(MockChannelRepository)
	pipelineRepo := new(MockPipelineRepository)
	resolver := NewSessionTimeoutResolver(channelRepo, pipelineRepo)
	ctx := context.Background()

	channelID := uuid.New()

	ch := createTestChannel(t, nil, 0) // No pipeline, no channel timeout

	channelRepo.On("GetByID", channelID).Return(ch, nil)

	// Act
	timeout, returnedPipelineID, err := resolver.ResolveForChannel(ctx, channelID)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, 30*time.Minute, timeout) // System default
	assert.Nil(t, returnedPipelineID)
	channelRepo.AssertExpectations(t)
	pipelineRepo.AssertExpectations(t)
}

func TestResolveForChannel_WithPipelineIDButNilUUID(t *testing.T) {
	// Arrange
	channelRepo := new(MockChannelRepository)
	pipelineRepo := new(MockPipelineRepository)
	resolver := NewSessionTimeoutResolver(channelRepo, pipelineRepo)
	ctx := context.Background()

	channelID := uuid.New()
	nilUUID := uuid.Nil

	ch := createTestChannel(t, &nilUUID, 45) // Pipeline ID set but is Nil UUID

	channelRepo.On("GetByID", channelID).Return(ch, nil)

	// Act
	timeout, returnedPipelineID, err := resolver.ResolveForChannel(ctx, channelID)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, 45*time.Minute, timeout) // Falls back to channel timeout
	assert.Nil(t, returnedPipelineID)
	channelRepo.AssertExpectations(t)
	pipelineRepo.AssertExpectations(t)
}

func TestResolveForContact(t *testing.T) {
	// Arrange
	channelRepo := new(MockChannelRepository)
	pipelineRepo := new(MockPipelineRepository)
	resolver := NewSessionTimeoutResolver(channelRepo, pipelineRepo)
	ctx := context.Background()

	contactID := uuid.New()
	channelID := uuid.New()

	ch := createTestChannel(t, nil, 45)

	channelRepo.On("GetByID", channelID).Return(ch, nil)

	// Act
	timeout, returnedPipelineID, err := resolver.ResolveForContact(ctx, contactID, channelID)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, 45*time.Minute, timeout)
	assert.Nil(t, returnedPipelineID)
	channelRepo.AssertExpectations(t)
	pipelineRepo.AssertExpectations(t)
}

func TestResolveWithFallback_UsesResolvedTimeout(t *testing.T) {
	// Arrange
	channelRepo := new(MockChannelRepository)
	pipelineRepo := new(MockPipelineRepository)
	resolver := NewSessionTimeoutResolver(channelRepo, pipelineRepo)
	ctx := context.Background()

	channelID := uuid.New()
	customFallback := 90 * time.Minute

	ch := createTestChannel(t, nil, 45)

	channelRepo.On("GetByID", channelID).Return(ch, nil)

	// Act
	timeout, returnedPipelineID, err := resolver.ResolveWithFallback(ctx, channelID, customFallback)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, 45*time.Minute, timeout) // Uses channel timeout, not fallback
	assert.Nil(t, returnedPipelineID)
	channelRepo.AssertExpectations(t)
	pipelineRepo.AssertExpectations(t)
}

func TestResolveWithFallback_UsesCustomFallback(t *testing.T) {
	// Arrange
	channelRepo := new(MockChannelRepository)
	pipelineRepo := new(MockPipelineRepository)
	resolver := NewSessionTimeoutResolver(channelRepo, pipelineRepo)
	ctx := context.Background()

	channelID := uuid.New()
	customFallback := 90 * time.Minute

	ch := createTestChannel(t, nil, 0) // No timeout configured

	channelRepo.On("GetByID", channelID).Return(ch, nil)

	// Act
	timeout, returnedPipelineID, err := resolver.ResolveWithFallback(ctx, channelID, customFallback)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, customFallback, timeout) // Uses custom fallback
	assert.Nil(t, returnedPipelineID)
	channelRepo.AssertExpectations(t)
	pipelineRepo.AssertExpectations(t)
}

func TestResolveWithFallback_WithError(t *testing.T) {
	// Arrange
	channelRepo := new(MockChannelRepository)
	pipelineRepo := new(MockPipelineRepository)
	resolver := NewSessionTimeoutResolver(channelRepo, pipelineRepo)
	ctx := context.Background()

	channelID := uuid.New()
	customFallback := 90 * time.Minute

	channelRepo.On("GetByID", channelID).Return(nil, errors.New("channel not found"))

	// Act
	timeout, returnedPipelineID, err := resolver.ResolveWithFallback(ctx, channelID, customFallback)

	// Assert
	assert.NoError(t, err)                   // ResolveForChannel returns no error for channel not found
	assert.Equal(t, customFallback, timeout) // Uses custom fallback since system default was returned
	assert.Nil(t, returnedPipelineID)
	channelRepo.AssertExpectations(t)
	pipelineRepo.AssertExpectations(t)
}

func TestGetEffectiveTimeout(t *testing.T) {
	// Arrange
	channelRepo := new(MockChannelRepository)
	pipelineRepo := new(MockPipelineRepository)
	resolver := NewSessionTimeoutResolver(channelRepo, pipelineRepo)
	ctx := context.Background()

	channelID := uuid.New()

	ch := createTestChannel(t, nil, 45)

	channelRepo.On("GetByID", channelID).Return(ch, nil)

	// Act
	timeout := resolver.GetEffectiveTimeout(ctx, channelID)

	// Assert
	assert.Equal(t, 45*time.Minute, timeout)
	channelRepo.AssertExpectations(t)
	pipelineRepo.AssertExpectations(t)
}

func TestGetEffectiveTimeout_WithError(t *testing.T) {
	// Arrange
	channelRepo := new(MockChannelRepository)
	pipelineRepo := new(MockPipelineRepository)
	resolver := NewSessionTimeoutResolver(channelRepo, pipelineRepo)
	ctx := context.Background()

	channelID := uuid.New()

	channelRepo.On("GetByID", channelID).Return(nil, errors.New("channel not found"))

	// Act
	timeout := resolver.GetEffectiveTimeout(ctx, channelID)

	// Assert
	assert.Equal(t, 30*time.Minute, timeout) // System default
	channelRepo.AssertExpectations(t)
	pipelineRepo.AssertExpectations(t)
}

func TestResolveWithDetails_FromPipeline(t *testing.T) {
	// Arrange
	channelRepo := new(MockChannelRepository)
	pipelineRepo := new(MockPipelineRepository)
	resolver := NewSessionTimeoutResolver(channelRepo, pipelineRepo)
	ctx := context.Background()

	pipelineID := uuid.New()
	channelID := uuid.New()
	pipelineTimeout := 60

	ch := createTestChannel(t, &pipelineID, 45)
	pipe := createTestPipeline(t, &pipelineTimeout)

	channelRepo.On("GetByID", channelID).Return(ch, nil)
	pipelineRepo.On("FindPipelineByID", ctx, pipelineID).Return(pipe, nil)

	// Act
	info, err := resolver.ResolveWithDetails(ctx, channelID)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, info)
	assert.Equal(t, 60*time.Minute, info.Duration)
	assert.Equal(t, TimeoutSourcePipeline, info.Source)
	assert.NotNil(t, info.PipelineID)
	assert.Equal(t, pipelineID, *info.PipelineID)
	assert.Equal(t, "Test Pipeline", info.PipelineName)
	assert.Equal(t, "Test Channel", info.ChannelName)
	channelRepo.AssertExpectations(t)
	pipelineRepo.AssertExpectations(t)
}

func TestResolveWithDetails_FromChannel(t *testing.T) {
	// Arrange
	channelRepo := new(MockChannelRepository)
	pipelineRepo := new(MockPipelineRepository)
	resolver := NewSessionTimeoutResolver(channelRepo, pipelineRepo)
	ctx := context.Background()

	channelID := uuid.New()

	ch := createTestChannel(t, nil, 45)

	channelRepo.On("GetByID", channelID).Return(ch, nil)

	// Act
	info, err := resolver.ResolveWithDetails(ctx, channelID)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, info)
	assert.Equal(t, 45*time.Minute, info.Duration)
	assert.Equal(t, TimeoutSourceChannel, info.Source)
	assert.Nil(t, info.PipelineID)
	assert.Equal(t, "", info.PipelineName)
	assert.Equal(t, "Test Channel", info.ChannelName)
	channelRepo.AssertExpectations(t)
	pipelineRepo.AssertExpectations(t)
}

func TestResolveWithDetails_FromSystemDefault(t *testing.T) {
	// Arrange
	channelRepo := new(MockChannelRepository)
	pipelineRepo := new(MockPipelineRepository)
	resolver := NewSessionTimeoutResolver(channelRepo, pipelineRepo)
	ctx := context.Background()

	channelID := uuid.New()

	ch := createTestChannel(t, nil, 0) // No timeout configured

	channelRepo.On("GetByID", channelID).Return(ch, nil)

	// Act
	info, err := resolver.ResolveWithDetails(ctx, channelID)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, info)
	assert.Equal(t, 30*time.Minute, info.Duration)
	assert.Equal(t, TimeoutSourceSystemDefault, info.Source)
	assert.Nil(t, info.PipelineID)
	assert.Equal(t, "", info.PipelineName)
	assert.Equal(t, "Test Channel", info.ChannelName)
	channelRepo.AssertExpectations(t)
	pipelineRepo.AssertExpectations(t)
}

func TestResolveWithDetails_ChannelNotFound(t *testing.T) {
	// Arrange
	channelRepo := new(MockChannelRepository)
	pipelineRepo := new(MockPipelineRepository)
	resolver := NewSessionTimeoutResolver(channelRepo, pipelineRepo)
	ctx := context.Background()

	channelID := uuid.New()

	channelRepo.On("GetByID", channelID).Return(nil, errors.New("channel not found"))

	// Act
	info, err := resolver.ResolveWithDetails(ctx, channelID)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, info)
	assert.Equal(t, 30*time.Minute, info.Duration)
	assert.Equal(t, TimeoutSourceSystemDefault, info.Source)
	assert.Nil(t, info.PipelineID)
	assert.Equal(t, "", info.PipelineName)
	assert.Equal(t, "", info.ChannelName) // Channel name is empty
	channelRepo.AssertExpectations(t)
	pipelineRepo.AssertExpectations(t)
}

func TestResolveWithDetails_PipelineNotFoundFallsBackToChannel(t *testing.T) {
	// Arrange
	channelRepo := new(MockChannelRepository)
	pipelineRepo := new(MockPipelineRepository)
	resolver := NewSessionTimeoutResolver(channelRepo, pipelineRepo)
	ctx := context.Background()

	pipelineID := uuid.New()
	channelID := uuid.New()

	ch := createTestChannel(t, &pipelineID, 45)

	channelRepo.On("GetByID", channelID).Return(ch, nil)
	pipelineRepo.On("FindPipelineByID", ctx, pipelineID).Return(nil, errors.New("pipeline not found"))

	// Act
	info, err := resolver.ResolveWithDetails(ctx, channelID)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, info)
	assert.Equal(t, 45*time.Minute, info.Duration)
	assert.Equal(t, TimeoutSourceChannel, info.Source)
	assert.Nil(t, info.PipelineID)
	assert.Equal(t, "", info.PipelineName)
	assert.Equal(t, "Test Channel", info.ChannelName)
	channelRepo.AssertExpectations(t)
	pipelineRepo.AssertExpectations(t)
}

func TestResolveWithDetails_PipelineTimeoutNilFallsBackToChannel(t *testing.T) {
	// Arrange
	channelRepo := new(MockChannelRepository)
	pipelineRepo := new(MockPipelineRepository)
	resolver := NewSessionTimeoutResolver(channelRepo, pipelineRepo)
	ctx := context.Background()

	pipelineID := uuid.New()
	channelID := uuid.New()

	ch := createTestChannel(t, &pipelineID, 45)
	pipe := createTestPipeline(t, nil) // No timeout set

	channelRepo.On("GetByID", channelID).Return(ch, nil)
	pipelineRepo.On("FindPipelineByID", ctx, pipelineID).Return(pipe, nil)

	// Act
	info, err := resolver.ResolveWithDetails(ctx, channelID)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, info)
	assert.Equal(t, 45*time.Minute, info.Duration)
	assert.Equal(t, TimeoutSourceChannel, info.Source)
	assert.Nil(t, info.PipelineID)
	assert.Equal(t, "", info.PipelineName)
	assert.Equal(t, "Test Channel", info.ChannelName)
	channelRepo.AssertExpectations(t)
	pipelineRepo.AssertExpectations(t)
}

func TestSetSystemDefault(t *testing.T) {
	// Arrange
	channelRepo := new(MockChannelRepository)
	pipelineRepo := new(MockPipelineRepository)
	resolver := NewSessionTimeoutResolver(channelRepo, pipelineRepo)

	// Act
	resolver.SetSystemDefault(60 * time.Minute)

	// Assert
	assert.Equal(t, 60*time.Minute, resolver.systemDefault)
}

func TestSetSystemDefault_AffectsResolution(t *testing.T) {
	// Arrange
	channelRepo := new(MockChannelRepository)
	pipelineRepo := new(MockPipelineRepository)
	resolver := NewSessionTimeoutResolver(channelRepo, pipelineRepo)
	ctx := context.Background()

	channelID := uuid.New()

	ch := createTestChannel(t, nil, 0) // No timeout configured

	channelRepo.On("GetByID", channelID).Return(ch, nil)

	// Act
	resolver.SetSystemDefault(60 * time.Minute)
	timeout, _, err := resolver.ResolveForChannel(ctx, channelID)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, 60*time.Minute, timeout) // Uses new system default
	channelRepo.AssertExpectations(t)
	pipelineRepo.AssertExpectations(t)
}

func TestTimeoutHierarchy_PipelineHasHighestPriority(t *testing.T) {
	// Arrange
	channelRepo := new(MockChannelRepository)
	pipelineRepo := new(MockPipelineRepository)
	resolver := NewSessionTimeoutResolver(channelRepo, pipelineRepo)
	resolver.SetSystemDefault(10 * time.Minute) // System default: 10 minutes
	ctx := context.Background()

	pipelineID := uuid.New()
	channelID := uuid.New()
	pipelineTimeout := 60

	ch := createTestChannel(t, &pipelineID, 45)     // Channel: 45 minutes
	pipe := createTestPipeline(t, &pipelineTimeout) // Pipeline: 60 minutes

	channelRepo.On("GetByID", channelID).Return(ch, nil)
	pipelineRepo.On("FindPipelineByID", ctx, pipelineID).Return(pipe, nil)

	// Act
	timeout, _, err := resolver.ResolveForChannel(ctx, channelID)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, 60*time.Minute, timeout) // Uses pipeline timeout (highest priority)
	channelRepo.AssertExpectations(t)
	pipelineRepo.AssertExpectations(t)
}

func TestTimeoutHierarchy_ChannelHasSecondPriority(t *testing.T) {
	// Arrange
	channelRepo := new(MockChannelRepository)
	pipelineRepo := new(MockPipelineRepository)
	resolver := NewSessionTimeoutResolver(channelRepo, pipelineRepo)
	resolver.SetSystemDefault(10 * time.Minute) // System default: 10 minutes
	ctx := context.Background()

	channelID := uuid.New()

	ch := createTestChannel(t, nil, 45) // Channel: 45 minutes, no pipeline

	channelRepo.On("GetByID", channelID).Return(ch, nil)

	// Act
	timeout, _, err := resolver.ResolveForChannel(ctx, channelID)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, 45*time.Minute, timeout) // Uses channel timeout (second priority)
	channelRepo.AssertExpectations(t)
	pipelineRepo.AssertExpectations(t)
}

func TestTimeoutHierarchy_SystemDefaultHasLowestPriority(t *testing.T) {
	// Arrange
	channelRepo := new(MockChannelRepository)
	pipelineRepo := new(MockPipelineRepository)
	resolver := NewSessionTimeoutResolver(channelRepo, pipelineRepo)
	resolver.SetSystemDefault(10 * time.Minute) // System default: 10 minutes
	ctx := context.Background()

	channelID := uuid.New()

	ch := createTestChannel(t, nil, 0) // No channel timeout, no pipeline

	channelRepo.On("GetByID", channelID).Return(ch, nil)

	// Act
	timeout, _, err := resolver.ResolveForChannel(ctx, channelID)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, 10*time.Minute, timeout) // Uses system default (lowest priority)
	channelRepo.AssertExpectations(t)
	pipelineRepo.AssertExpectations(t)
}

func TestEdgeCase_ZeroChannelTimeout(t *testing.T) {
	// Arrange
	channelRepo := new(MockChannelRepository)
	pipelineRepo := new(MockPipelineRepository)
	resolver := NewSessionTimeoutResolver(channelRepo, pipelineRepo)
	ctx := context.Background()

	channelID := uuid.New()

	ch := createTestChannel(t, nil, 0)

	channelRepo.On("GetByID", channelID).Return(ch, nil)

	// Act
	timeout, _, err := resolver.ResolveForChannel(ctx, channelID)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, 30*time.Minute, timeout) // Falls back to system default
	channelRepo.AssertExpectations(t)
	pipelineRepo.AssertExpectations(t)
}

func TestEdgeCase_NegativeChannelTimeout(t *testing.T) {
	// Arrange
	channelRepo := new(MockChannelRepository)
	pipelineRepo := new(MockPipelineRepository)
	resolver := NewSessionTimeoutResolver(channelRepo, pipelineRepo)
	ctx := context.Background()

	channelID := uuid.New()

	ch := createTestChannel(t, nil, -10) // Negative timeout

	channelRepo.On("GetByID", channelID).Return(ch, nil)

	// Act
	timeout, _, err := resolver.ResolveForChannel(ctx, channelID)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, 30*time.Minute, timeout) // Falls back to system default (negative treated as 0)
	channelRepo.AssertExpectations(t)
	pipelineRepo.AssertExpectations(t)
}

func TestConcurrency_MultipleGoroutines(t *testing.T) {
	// Arrange
	channelRepo := new(MockChannelRepository)
	pipelineRepo := new(MockPipelineRepository)
	resolver := NewSessionTimeoutResolver(channelRepo, pipelineRepo)
	ctx := context.Background()

	channelID := uuid.New()

	ch := createTestChannel(t, nil, 45)

	// Setup mock to handle multiple calls
	channelRepo.On("GetByID", channelID).Return(ch, nil).Times(10)

	// Act - call from multiple goroutines
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			timeout, _, err := resolver.ResolveForChannel(ctx, channelID)
			assert.NoError(t, err)
			assert.Equal(t, 45*time.Minute, timeout)
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Assert
	channelRepo.AssertExpectations(t)
	pipelineRepo.AssertExpectations(t)
}
