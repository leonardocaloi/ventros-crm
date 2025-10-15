package chat

import (
	"context"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	domainchat "github.com/ventros/crm/internal/domain/crm/chat"
)

// MockChatRepository is a mock implementation of domainchat.Repository
type MockChatRepository struct {
	mock.Mock
}

func (m *MockChatRepository) Create(ctx context.Context, chat *domainchat.Chat) error {
	args := m.Called(ctx, chat)
	return args.Error(0)
}

func (m *MockChatRepository) FindByID(ctx context.Context, id uuid.UUID) (*domainchat.Chat, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domainchat.Chat), args.Error(1)
}

func (m *MockChatRepository) FindByExternalID(ctx context.Context, externalID string) (*domainchat.Chat, error) {
	args := m.Called(ctx, externalID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domainchat.Chat), args.Error(1)
}

func (m *MockChatRepository) FindByProject(ctx context.Context, projectID uuid.UUID) ([]*domainchat.Chat, error) {
	args := m.Called(ctx, projectID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domainchat.Chat), args.Error(1)
}

func (m *MockChatRepository) FindByTenant(ctx context.Context, tenantID string) ([]*domainchat.Chat, error) {
	args := m.Called(ctx, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domainchat.Chat), args.Error(1)
}

func (m *MockChatRepository) FindByContact(ctx context.Context, contactID uuid.UUID) ([]*domainchat.Chat, error) {
	args := m.Called(ctx, contactID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domainchat.Chat), args.Error(1)
}

func (m *MockChatRepository) FindActiveByProject(ctx context.Context, projectID uuid.UUID) ([]*domainchat.Chat, error) {
	args := m.Called(ctx, projectID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domainchat.Chat), args.Error(1)
}

func (m *MockChatRepository) FindIndividualByContact(ctx context.Context, contactID uuid.UUID, projectID uuid.UUID) (*domainchat.Chat, error) {
	args := m.Called(ctx, contactID, projectID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domainchat.Chat), args.Error(1)
}

func (m *MockChatRepository) Update(ctx context.Context, chat *domainchat.Chat) error {
	args := m.Called(ctx, chat)
	return args.Error(0)
}

func (m *MockChatRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockChatRepository) SearchBySubject(ctx context.Context, tenantID string, subject string) ([]*domainchat.Chat, error) {
	args := m.Called(ctx, tenantID, subject)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domainchat.Chat), args.Error(1)
}

// MockEventBus is a mock implementation of EventBus
type MockEventBus struct {
	mock.Mock
}

func (m *MockEventBus) Publish(ctx context.Context, event domainchat.DomainEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}
