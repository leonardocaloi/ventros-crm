package persistence

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/caloi/ventros-crm/infrastructure/persistence/entities"
	"github.com/caloi/ventros-crm/internal/domain/automation/broadcast"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// GormBroadcastRepository implements broadcast.Repository using GORM
type GormBroadcastRepository struct {
	db *gorm.DB
}

// NewGormBroadcastRepository creates a new instance
func NewGormBroadcastRepository(db *gorm.DB) broadcast.Repository {
	return &GormBroadcastRepository{db: db}
}

// Save saves or updates a broadcast
func (r *GormBroadcastRepository) Save(b *broadcast.Broadcast) error {
	entity, err := r.toEntity(b)
	if err != nil {
		return fmt.Errorf("failed to convert broadcast to entity: %w", err)
	}

	// Check if exists
	var existing entities.BroadcastEntity
	err = r.db.Where("id = ?", entity.ID).First(&existing).Error

	if err == nil {
		// Update
		return r.db.Model(&existing).Updates(entity).Error
	} else if errors.Is(err, gorm.ErrRecordNotFound) {
		// Insert
		return r.db.Create(entity).Error
	}

	return err
}

// FindByID finds a broadcast by ID
func (r *GormBroadcastRepository) FindByID(id uuid.UUID) (*broadcast.Broadcast, error) {
	var entity entities.BroadcastEntity
	err := r.db.Where("id = ?", id).First(&entity).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return r.toDomain(&entity)
}

// FindByTenantID finds all broadcasts for a tenant
func (r *GormBroadcastRepository) FindByTenantID(tenantID string) ([]*broadcast.Broadcast, error) {
	var entities []entities.BroadcastEntity
	err := r.db.Where("tenant_id = ?", tenantID).Order("created_at DESC").Find(&entities).Error
	if err != nil {
		return nil, err
	}

	return r.toDomainSlice(entities)
}

// FindScheduledReady finds broadcasts scheduled and ready to start
func (r *GormBroadcastRepository) FindScheduledReady() ([]*broadcast.Broadcast, error) {
	var entities []entities.BroadcastEntity
	err := r.db.Where("status = ? AND scheduled_for <= NOW()", string(broadcast.BroadcastStatusScheduled)).
		Order("scheduled_for ASC").
		Find(&entities).Error
	if err != nil {
		return nil, err
	}

	return r.toDomainSlice(entities)
}

// FindByStatus finds broadcasts by status
func (r *GormBroadcastRepository) FindByStatus(status broadcast.BroadcastStatus) ([]*broadcast.Broadcast, error) {
	var entities []entities.BroadcastEntity
	err := r.db.Where("status = ?", string(status)).Order("created_at DESC").Find(&entities).Error
	if err != nil {
		return nil, err
	}

	return r.toDomainSlice(entities)
}

// Delete deletes a broadcast
func (r *GormBroadcastRepository) Delete(id uuid.UUID) error {
	result := r.db.Delete(&entities.BroadcastEntity{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// Converters

func (r *GormBroadcastRepository) toEntity(b *broadcast.Broadcast) (*entities.BroadcastEntity, error) {
	if b == nil {
		return nil, errors.New("broadcast cannot be nil")
	}

	// Serialize message template
	templateJSON, err := json.Marshal(b.MessageTemplate())
	if err != nil {
		return nil, fmt.Errorf("failed to marshal message template: %w", err)
	}

	return &entities.BroadcastEntity{
		ID:              b.ID(),
		TenantID:        b.TenantID(),
		Name:            b.Name(),
		ListID:          b.ListID(),
		MessageTemplate: templateJSON,
		Status:          string(b.Status()),
		ScheduledFor:    b.ScheduledFor(),
		StartedAt:       b.StartedAt(),
		CompletedAt:     b.CompletedAt(),
		TotalContacts:   b.TotalContacts(),
		SentCount:       b.SentCount(),
		FailedCount:     b.FailedCount(),
		PendingCount:    b.PendingCount(),
		RateLimit:       b.RateLimit(),
		CreatedAt:       b.CreatedAt(),
		UpdatedAt:       b.UpdatedAt(),
	}, nil
}

func (r *GormBroadcastRepository) toDomain(entity *entities.BroadcastEntity) (*broadcast.Broadcast, error) {
	if entity == nil {
		return nil, errors.New("entity cannot be nil")
	}

	// Deserialize message template
	var messageTemplate broadcast.MessageTemplate
	if len(entity.MessageTemplate) > 0 {
		if err := json.Unmarshal(entity.MessageTemplate, &messageTemplate); err != nil {
			return nil, fmt.Errorf("failed to unmarshal message template: %w", err)
		}
	}

	// Reconstruct domain model
	return broadcast.ReconstructBroadcast(
		entity.ID,
		entity.TenantID,
		entity.Name,
		entity.ListID,
		messageTemplate,
		broadcast.BroadcastStatus(entity.Status),
		entity.ScheduledFor,
		entity.StartedAt,
		entity.CompletedAt,
		entity.TotalContacts,
		entity.SentCount,
		entity.FailedCount,
		entity.PendingCount,
		entity.RateLimit,
		entity.CreatedAt,
		entity.UpdatedAt,
	), nil
}

func (r *GormBroadcastRepository) toDomainSlice(entities []entities.BroadcastEntity) ([]*broadcast.Broadcast, error) {
	broadcasts := make([]*broadcast.Broadcast, 0, len(entities))
	for _, entity := range entities {
		b, err := r.toDomain(&entity)
		if err != nil {
			return nil, err
		}
		broadcasts = append(broadcasts, b)
	}
	return broadcasts, nil
}

// GormBroadcastExecutionRepository implements broadcast.ExecutionRepository using GORM
type GormBroadcastExecutionRepository struct {
	db *gorm.DB
}

// NewGormBroadcastExecutionRepository creates a new instance
func NewGormBroadcastExecutionRepository(db *gorm.DB) broadcast.ExecutionRepository {
	return &GormBroadcastExecutionRepository{db: db}
}

// Save saves or updates an execution
func (r *GormBroadcastExecutionRepository) Save(exec *broadcast.BroadcastExecution) error {
	entity := r.executionToEntity(exec)

	// Check if exists
	var existing entities.BroadcastExecutionEntity
	err := r.db.Where("id = ?", entity.ID).First(&existing).Error

	if err == nil {
		// Update
		return r.db.Model(&existing).Updates(entity).Error
	} else if errors.Is(err, gorm.ErrRecordNotFound) {
		// Insert
		return r.db.Create(entity).Error
	}

	return err
}

// SaveBatch saves multiple executions in a batch
func (r *GormBroadcastExecutionRepository) SaveBatch(executions []*broadcast.BroadcastExecution) error {
	entities := make([]entities.BroadcastExecutionEntity, len(executions))
	for i, exec := range executions {
		entities[i] = *r.executionToEntity(exec)
	}

	return r.db.Create(&entities).Error
}

// FindByID finds an execution by ID
func (r *GormBroadcastExecutionRepository) FindByID(id uuid.UUID) (*broadcast.BroadcastExecution, error) {
	var entity entities.BroadcastExecutionEntity
	err := r.db.Where("id = ?", id).First(&entity).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return r.executionToDomain(&entity), nil
}

// FindByBroadcastID finds all executions for a broadcast
func (r *GormBroadcastExecutionRepository) FindByBroadcastID(broadcastID uuid.UUID) ([]*broadcast.BroadcastExecution, error) {
	var entities []entities.BroadcastExecutionEntity
	err := r.db.Where("broadcast_id = ?", broadcastID).Order("created_at ASC").Find(&entities).Error
	if err != nil {
		return nil, err
	}

	return r.executionToDomainSlice(entities), nil
}

// FindPendingByBroadcastID finds pending executions for a broadcast
func (r *GormBroadcastExecutionRepository) FindPendingByBroadcastID(broadcastID uuid.UUID) ([]*broadcast.BroadcastExecution, error) {
	var entities []entities.BroadcastExecutionEntity
	err := r.db.Where("broadcast_id = ? AND status = ?", broadcastID, string(broadcast.ExecutionStatusPending)).
		Order("created_at ASC").
		Find(&entities).Error
	if err != nil {
		return nil, err
	}

	return r.executionToDomainSlice(entities), nil
}

// Delete deletes an execution
func (r *GormBroadcastExecutionRepository) Delete(id uuid.UUID) error {
	result := r.db.Delete(&entities.BroadcastExecutionEntity{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// Execution converters

func (r *GormBroadcastExecutionRepository) executionToEntity(exec *broadcast.BroadcastExecution) *entities.BroadcastExecutionEntity {
	return &entities.BroadcastExecutionEntity{
		ID:          exec.ID(),
		BroadcastID: exec.BroadcastID(),
		ContactID:   exec.ContactID(),
		Status:      string(exec.Status()),
		MessageID:   exec.MessageID(),
		Error:       exec.Error(),
		SentAt:      exec.SentAt(),
		CreatedAt:   exec.CreatedAt(),
		UpdatedAt:   exec.UpdatedAt(),
	}
}

func (r *GormBroadcastExecutionRepository) executionToDomain(entity *entities.BroadcastExecutionEntity) *broadcast.BroadcastExecution {
	return broadcast.ReconstructBroadcastExecution(
		entity.ID,
		entity.BroadcastID,
		entity.ContactID,
		broadcast.ExecutionStatus(entity.Status),
		entity.MessageID,
		entity.Error,
		entity.SentAt,
		entity.CreatedAt,
		entity.UpdatedAt,
	)
}

func (r *GormBroadcastExecutionRepository) executionToDomainSlice(entities []entities.BroadcastExecutionEntity) []*broadcast.BroadcastExecution {
	executions := make([]*broadcast.BroadcastExecution, len(entities))
	for i, entity := range entities {
		executions[i] = r.executionToDomain(&entity)
	}
	return executions
}
