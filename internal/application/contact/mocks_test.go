package contact

import (
	"context"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/ventros/crm/internal/domain/crm/contact"
	"github.com/ventros/crm/internal/domain/crm/pipeline"
)

// ========== Shared Mocks for contact package tests ==========

type MockContactRepository struct {
	mock.Mock
}

func (m *MockContactRepository) Save(ctx context.Context, c *contact.Contact) error {
	args := m.Called(ctx, c)
	return args.Error(0)
}

func (m *MockContactRepository) FindByID(ctx context.Context, id uuid.UUID) (*contact.Contact, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*contact.Contact), args.Error(1)
}

func (m *MockContactRepository) FindByPhone(ctx context.Context, projectID uuid.UUID, phone string) (*contact.Contact, error) {
	args := m.Called(ctx, projectID, phone)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*contact.Contact), args.Error(1)
}

func (m *MockContactRepository) FindByEmail(ctx context.Context, projectID uuid.UUID, email string) (*contact.Contact, error) {
	args := m.Called(ctx, projectID, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*contact.Contact), args.Error(1)
}

func (m *MockContactRepository) FindByExternalID(ctx context.Context, projectID uuid.UUID, externalID string) (*contact.Contact, error) {
	args := m.Called(ctx, projectID, externalID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*contact.Contact), args.Error(1)
}

func (m *MockContactRepository) FindByProject(ctx context.Context, projectID uuid.UUID, limit, offset int) ([]*contact.Contact, error) {
	args := m.Called(ctx, projectID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*contact.Contact), args.Error(1)
}

func (m *MockContactRepository) CountByProject(ctx context.Context, projectID uuid.UUID) (int, error) {
	args := m.Called(ctx, projectID)
	return args.Int(0), args.Error(1)
}

func (m *MockContactRepository) FindByTenantWithFilters(ctx context.Context, tenantID string, filters contact.ContactFilters, page, limit int, sortBy, sortDir string) ([]*contact.Contact, int64, error) {
	args := m.Called(ctx, tenantID, filters, page, limit, sortBy, sortDir)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*contact.Contact), args.Get(1).(int64), args.Error(2)
}

func (m *MockContactRepository) SearchByText(ctx context.Context, tenantID string, searchText string, limit int) ([]*contact.Contact, error) {
	args := m.Called(ctx, tenantID, searchText, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*contact.Contact), args.Error(1)
}

func (m *MockContactRepository) SaveCustomFields(ctx context.Context, contactID uuid.UUID, fields map[string]string) error {
	args := m.Called(ctx, contactID, fields)
	return args.Error(0)
}

func (m *MockContactRepository) FindByCustomField(ctx context.Context, tenantID, key, value string) (*contact.Contact, error) {
	args := m.Called(ctx, tenantID, key, value)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*contact.Contact), args.Error(1)
}

func (m *MockContactRepository) GetCustomFields(ctx context.Context, contactID uuid.UUID) (map[string]string, error) {
	args := m.Called(ctx, contactID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]string), args.Error(1)
}

type MockPipelineRepository struct {
	mock.Mock
}

func (m *MockPipelineRepository) SavePipeline(ctx context.Context, p *pipeline.Pipeline) error {
	args := m.Called(ctx, p)
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
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*pipeline.Pipeline), args.Error(1)
}

func (m *MockPipelineRepository) FindPipelinesByTenant(ctx context.Context, tenantID string) ([]*pipeline.Pipeline, error) {
	args := m.Called(ctx, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*pipeline.Pipeline), args.Error(1)
}

func (m *MockPipelineRepository) FindActivePipelinesByProject(ctx context.Context, projectID uuid.UUID) ([]*pipeline.Pipeline, error) {
	args := m.Called(ctx, projectID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*pipeline.Pipeline), args.Error(1)
}

func (m *MockPipelineRepository) DeletePipeline(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockPipelineRepository) SaveStatus(ctx context.Context, s *pipeline.Status) error {
	args := m.Called(ctx, s)
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
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*pipeline.Status), args.Error(1)
}

func (m *MockPipelineRepository) FindActiveStatusesByPipeline(ctx context.Context, pipelineID uuid.UUID) ([]*pipeline.Status, error) {
	args := m.Called(ctx, pipelineID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
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
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]uuid.UUID), args.Error(1)
}

func (m *MockPipelineRepository) GetContactStatusHistory(ctx context.Context, contactID, pipelineID uuid.UUID) ([]*pipeline.ContactStatusHistory, error) {
	args := m.Called(ctx, contactID, pipelineID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*pipeline.ContactStatusHistory), args.Error(1)
}

func (m *MockPipelineRepository) FindByTenantWithFilters(ctx context.Context, filters pipeline.PipelineFilters) ([]*pipeline.Pipeline, int64, error) {
	args := m.Called(ctx, filters)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*pipeline.Pipeline), args.Get(1).(int64), args.Error(2)
}

func (m *MockPipelineRepository) SearchByText(ctx context.Context, tenantID string, searchText string, limit int, offset int) ([]*pipeline.Pipeline, int64, error) {
	args := m.Called(ctx, tenantID, searchText, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*pipeline.Pipeline), args.Get(1).(int64), args.Error(2)
}

func (m *MockPipelineRepository) SaveCustomField(ctx context.Context, field *pipeline.PipelineCustomField) error {
	args := m.Called(ctx, field)
	return args.Error(0)
}

func (m *MockPipelineRepository) FindCustomFieldByID(ctx context.Context, id uuid.UUID) (*pipeline.PipelineCustomField, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pipeline.PipelineCustomField), args.Error(1)
}

func (m *MockPipelineRepository) FindCustomFieldByKey(ctx context.Context, pipelineID uuid.UUID, key string) (*pipeline.PipelineCustomField, error) {
	args := m.Called(ctx, pipelineID, key)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pipeline.PipelineCustomField), args.Error(1)
}

func (m *MockPipelineRepository) FindCustomFieldsByPipeline(ctx context.Context, pipelineID uuid.UUID) ([]*pipeline.PipelineCustomField, error) {
	args := m.Called(ctx, pipelineID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*pipeline.PipelineCustomField), args.Error(1)
}

func (m *MockPipelineRepository) DeleteCustomField(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockPipelineRepository) DeleteCustomFieldByKey(ctx context.Context, pipelineID uuid.UUID, key string) error {
	args := m.Called(ctx, pipelineID, key)
	return args.Error(0)
}

type MockEventBus struct {
	mock.Mock
}

func (m *MockEventBus) Publish(ctx context.Context, event contact.DomainEvent) error {
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
