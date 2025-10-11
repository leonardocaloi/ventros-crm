package contact_list

import (
	"context"

	"github.com/caloi/ventros-crm/internal/domain/crm/contact_list"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// MockContactListRepository is a mock implementation of contact_list.Repository
type MockContactListRepository struct {
	mock.Mock
}

func (m *MockContactListRepository) Create(ctx context.Context, list *contact_list.ContactList) error {
	args := m.Called(ctx, list)
	return args.Error(0)
}

func (m *MockContactListRepository) Update(ctx context.Context, list *contact_list.ContactList) error {
	args := m.Called(ctx, list)
	return args.Error(0)
}

func (m *MockContactListRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockContactListRepository) FindByID(ctx context.Context, id uuid.UUID) (*contact_list.ContactList, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*contact_list.ContactList), args.Error(1)
}

func (m *MockContactListRepository) FindByProjectID(ctx context.Context, projectID uuid.UUID) ([]*contact_list.ContactList, error) {
	args := m.Called(ctx, projectID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*contact_list.ContactList), args.Error(1)
}

func (m *MockContactListRepository) FindByTenantID(ctx context.Context, tenantID string) ([]*contact_list.ContactList, error) {
	args := m.Called(ctx, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*contact_list.ContactList), args.Error(1)
}

func (m *MockContactListRepository) ListByProject(ctx context.Context, projectID uuid.UUID, limit, offset int) ([]*contact_list.ContactList, int, error) {
	args := m.Called(ctx, projectID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*contact_list.ContactList), args.Int(1), args.Error(2)
}

func (m *MockContactListRepository) GetContactsInList(ctx context.Context, listID uuid.UUID, limit, offset int) ([]uuid.UUID, int, error) {
	args := m.Called(ctx, listID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]uuid.UUID), args.Int(1), args.Error(2)
}

func (m *MockContactListRepository) RecalculateContactCount(ctx context.Context, listID uuid.UUID) (int, error) {
	args := m.Called(ctx, listID)
	return args.Int(0), args.Error(1)
}

func (m *MockContactListRepository) AddContactToStaticList(ctx context.Context, listID, contactID uuid.UUID) error {
	args := m.Called(ctx, listID, contactID)
	return args.Error(0)
}

func (m *MockContactListRepository) RemoveContactFromStaticList(ctx context.Context, listID, contactID uuid.UUID) error {
	args := m.Called(ctx, listID, contactID)
	return args.Error(0)
}

func (m *MockContactListRepository) IsContactInList(ctx context.Context, listID, contactID uuid.UUID) (bool, error) {
	args := m.Called(ctx, listID, contactID)
	return args.Bool(0), args.Error(1)
}
