package agent

import (
	"context"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/ventros/crm/internal/domain/core/shared"
	"github.com/ventros/crm/internal/domain/crm/agent"
)

// ========== Shared Mocks for agent package tests ==========

type MockAgentRepository struct {
	mock.Mock
}

func (m *MockAgentRepository) Save(ctx context.Context, a *agent.Agent) error {
	args := m.Called(ctx, a)
	return args.Error(0)
}

func (m *MockAgentRepository) FindByID(ctx context.Context, id uuid.UUID) (*agent.Agent, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*agent.Agent), args.Error(1)
}

func (m *MockAgentRepository) FindByEmail(ctx context.Context, tenantID, email string) (*agent.Agent, error) {
	args := m.Called(ctx, tenantID, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*agent.Agent), args.Error(1)
}

func (m *MockAgentRepository) FindByTenant(ctx context.Context, tenantID string) ([]*agent.Agent, error) {
	args := m.Called(ctx, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*agent.Agent), args.Error(1)
}

func (m *MockAgentRepository) FindActiveByTenant(ctx context.Context, tenantID string) ([]*agent.Agent, error) {
	args := m.Called(ctx, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*agent.Agent), args.Error(1)
}

func (m *MockAgentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockAgentRepository) FindByTenantWithFilters(ctx context.Context, filters agent.AgentFilters) ([]*agent.Agent, int64, error) {
	args := m.Called(ctx, filters)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*agent.Agent), args.Get(1).(int64), args.Error(2)
}

func (m *MockAgentRepository) SearchByText(ctx context.Context, tenantID string, searchText string, limit int, offset int) ([]*agent.Agent, int64, error) {
	args := m.Called(ctx, tenantID, searchText, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*agent.Agent), args.Get(1).(int64), args.Error(2)
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
