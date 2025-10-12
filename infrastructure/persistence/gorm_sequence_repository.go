package persistence

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/caloi/ventros-crm/infrastructure/persistence/entities"
	"github.com/caloi/ventros-crm/internal/domain/automation/sequence"
	"github.com/caloi/ventros-crm/internal/domain/core/shared"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// GormSequenceRepository implements sequence.Repository using GORM
type GormSequenceRepository struct {
	db *gorm.DB
}

// NewGormSequenceRepository creates a new instance
func NewGormSequenceRepository(db *gorm.DB) sequence.Repository {
	return &GormSequenceRepository{db: db}
}

// Save saves or updates a sequence
func (r *GormSequenceRepository) Save(s *sequence.Sequence) error {
	entity, err := r.toEntity(s)
	if err != nil {
		return fmt.Errorf("failed to convert sequence to entity: %w", err)
	}

	// Check if exists
	var existing entities.SequenceEntity
	err = r.db.Where("id = ?", entity.ID).First(&existing).Error

	if err == nil {
		// Update - use transaction to update sequence and steps with optimistic locking
		return r.db.Transaction(func(tx *gorm.DB) error {
			// Update sequence with version check (optimistic locking)
			result := tx.Model(&entities.SequenceEntity{}).
				Where("id = ? AND version = ?", entity.ID, existing.Version).
				Updates(map[string]interface{}{
					"version":         existing.Version + 1, // Increment version
					"tenant_id":       entity.TenantID,
					"name":            entity.Name,
					"description":     entity.Description,
					"status":          entity.Status,
					"trigger_type":    entity.TriggerType,
					"trigger_data":    entity.TriggerData,
					"exit_on_reply":   entity.ExitOnReply,
					"total_enrolled":  entity.TotalEnrolled,
					"active_count":    entity.ActiveCount,
					"completed_count": entity.CompletedCount,
					"exited_count":    entity.ExitedCount,
					"updated_at":      entity.UpdatedAt,
				})

			if result.Error != nil {
				return result.Error
			}

			// Check optimistic locking - if 0 rows affected, version mismatch (concurrent update)
			if result.RowsAffected == 0 {
				return shared.NewOptimisticLockError(
					"Sequence",
					entity.ID.String(),
					existing.Version,
					entity.Version,
				)
			}

			// Delete existing steps
			if err := tx.Where("sequence_id = ?", entity.ID).Delete(&entities.SequenceStepEntity{}).Error; err != nil {
				return err
			}

			// Insert new steps
			steps := r.stepsToEntities(s.Steps(), s.ID())
			if len(steps) > 0 {
				if err := tx.Create(&steps).Error; err != nil {
					return err
				}
			}

			return nil
		})
	} else if errors.Is(err, gorm.ErrRecordNotFound) {
		// Insert - use transaction to insert sequence and steps
		return r.db.Transaction(func(tx *gorm.DB) error {
			// Create sequence
			if err := tx.Create(entity).Error; err != nil {
				return err
			}

			// Create steps
			steps := r.stepsToEntities(s.Steps(), s.ID())
			if len(steps) > 0 {
				if err := tx.Create(&steps).Error; err != nil {
					return err
				}
			}

			return nil
		})
	}

	return err
}

// FindByID finds a sequence by ID
func (r *GormSequenceRepository) FindByID(id uuid.UUID) (*sequence.Sequence, error) {
	var entity entities.SequenceEntity
	err := r.db.Where("id = ?", id).First(&entity).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	// Load steps
	var stepEntities []entities.SequenceStepEntity
	if err := r.db.Where("sequence_id = ?", id).Order("\"order\" ASC").Find(&stepEntities).Error; err != nil {
		return nil, err
	}

	return r.toDomain(&entity, stepEntities)
}

// FindByTenantID finds all sequences for a tenant
func (r *GormSequenceRepository) FindByTenantID(tenantID string) ([]*sequence.Sequence, error) {
	var entities []entities.SequenceEntity
	err := r.db.Where("tenant_id = ?", tenantID).Order("created_at DESC").Find(&entities).Error
	if err != nil {
		return nil, err
	}

	return r.toDomainSlice(entities)
}

// FindActiveByTriggerType finds active sequences by trigger type
func (r *GormSequenceRepository) FindActiveByTriggerType(triggerType sequence.TriggerType) ([]*sequence.Sequence, error) {
	var entities []entities.SequenceEntity
	err := r.db.Where("status = ? AND trigger_type = ?", string(sequence.SequenceStatusActive), string(triggerType)).
		Order("created_at DESC").
		Find(&entities).Error
	if err != nil {
		return nil, err
	}

	return r.toDomainSlice(entities)
}

// FindByStatus finds sequences by status
func (r *GormSequenceRepository) FindByStatus(status sequence.SequenceStatus) ([]*sequence.Sequence, error) {
	var entities []entities.SequenceEntity
	err := r.db.Where("status = ?", string(status)).Order("created_at DESC").Find(&entities).Error
	if err != nil {
		return nil, err
	}

	return r.toDomainSlice(entities)
}

// Delete deletes a sequence
func (r *GormSequenceRepository) Delete(id uuid.UUID) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Delete steps first
		if err := tx.Where("sequence_id = ?", id).Delete(&entities.SequenceStepEntity{}).Error; err != nil {
			return err
		}

		// Delete sequence
		result := tx.Delete(&entities.SequenceEntity{}, "id = ?", id)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}

		return nil
	})
}

// Converters

func (r *GormSequenceRepository) toEntity(s *sequence.Sequence) (*entities.SequenceEntity, error) {
	if s == nil {
		return nil, errors.New("sequence cannot be nil")
	}

	// Serialize trigger data
	triggerDataJSON, err := json.Marshal(s.TriggerData())
	if err != nil {
		return nil, fmt.Errorf("failed to marshal trigger data: %w", err)
	}

	return &entities.SequenceEntity{
		ID:             s.ID(),
		Version:        s.Version(),
		TenantID:       s.TenantID(),
		Name:           s.Name(),
		Description:    s.Description(),
		Status:         string(s.Status()),
		TriggerType:    string(s.TriggerType()),
		TriggerData:    triggerDataJSON,
		ExitOnReply:    s.ExitOnReply(),
		TotalEnrolled:  s.TotalEnrolled(),
		ActiveCount:    s.ActiveCount(),
		CompletedCount: s.CompletedCount(),
		ExitedCount:    s.ExitedCount(),
		CreatedAt:      s.CreatedAt(),
		UpdatedAt:      s.UpdatedAt(),
	}, nil
}

func (r *GormSequenceRepository) stepsToEntities(steps []sequence.SequenceStep, sequenceID uuid.UUID) []entities.SequenceStepEntity {
	stepEntities := make([]entities.SequenceStepEntity, len(steps))
	for i, step := range steps {
		templateJSON, _ := json.Marshal(step.MessageTemplate)
		conditionsJSON, _ := json.Marshal(step.Conditions)

		stepEntities[i] = entities.SequenceStepEntity{
			ID:              step.ID,
			SequenceID:      sequenceID,
			Order:           step.Order,
			Name:            step.Name,
			DelayAmount:     step.DelayAmount,
			DelayUnit:       string(step.DelayUnit),
			MessageTemplate: templateJSON,
			Conditions:      conditionsJSON,
			CreatedAt:       step.CreatedAt,
		}
	}
	return stepEntities
}

func (r *GormSequenceRepository) toDomain(entity *entities.SequenceEntity, stepEntities []entities.SequenceStepEntity) (*sequence.Sequence, error) {
	if entity == nil {
		return nil, errors.New("entity cannot be nil")
	}

	// Deserialize trigger data
	var triggerData map[string]interface{}
	if len(entity.TriggerData) > 0 {
		if err := json.Unmarshal(entity.TriggerData, &triggerData); err != nil {
			return nil, fmt.Errorf("failed to unmarshal trigger data: %w", err)
		}
	}

	// Convert steps
	steps := make([]sequence.SequenceStep, len(stepEntities))
	for i, stepEntity := range stepEntities {
		var messageTemplate sequence.MessageTemplate
		if len(stepEntity.MessageTemplate) > 0 {
			if err := json.Unmarshal(stepEntity.MessageTemplate, &messageTemplate); err != nil {
				return nil, fmt.Errorf("failed to unmarshal message template: %w", err)
			}
		}

		var conditions []sequence.StepCondition
		if len(stepEntity.Conditions) > 0 {
			if err := json.Unmarshal(stepEntity.Conditions, &conditions); err != nil {
				return nil, fmt.Errorf("failed to unmarshal conditions: %w", err)
			}
		}

		steps[i] = sequence.SequenceStep{
			ID:              stepEntity.ID,
			Order:           stepEntity.Order,
			Name:            stepEntity.Name,
			DelayAmount:     stepEntity.DelayAmount,
			DelayUnit:       sequence.DelayUnit(stepEntity.DelayUnit),
			MessageTemplate: messageTemplate,
			Conditions:      conditions,
			CreatedAt:       stepEntity.CreatedAt,
		}
	}

	// Reconstruct domain model
	return sequence.ReconstructSequence(
		entity.ID,
		entity.Version,
		entity.TenantID,
		entity.Name,
		entity.Description,
		sequence.SequenceStatus(entity.Status),
		steps,
		sequence.TriggerType(entity.TriggerType),
		triggerData,
		entity.ExitOnReply,
		entity.TotalEnrolled,
		entity.ActiveCount,
		entity.CompletedCount,
		entity.ExitedCount,
		entity.CreatedAt,
		entity.UpdatedAt,
	), nil
}

func (r *GormSequenceRepository) toDomainSlice(sequenceEntities []entities.SequenceEntity) ([]*sequence.Sequence, error) {
	sequences := make([]*sequence.Sequence, 0, len(sequenceEntities))
	for _, entity := range sequenceEntities {
		// Load steps for each sequence
		var stepEntities []entities.SequenceStepEntity
		if err := r.db.Where("sequence_id = ?", entity.ID).Order("\"order\" ASC").Find(&stepEntities).Error; err != nil {
			return nil, err
		}

		s, err := r.toDomain(&entity, stepEntities)
		if err != nil {
			return nil, err
		}
		sequences = append(sequences, s)
	}
	return sequences, nil
}

// GormSequenceEnrollmentRepository implements sequence.EnrollmentRepository using GORM
type GormSequenceEnrollmentRepository struct {
	db *gorm.DB
}

// NewGormSequenceEnrollmentRepository creates a new instance
func NewGormSequenceEnrollmentRepository(db *gorm.DB) sequence.EnrollmentRepository {
	return &GormSequenceEnrollmentRepository{db: db}
}

// Save saves or updates an enrollment
func (r *GormSequenceEnrollmentRepository) Save(e *sequence.SequenceEnrollment) error {
	entity := r.enrollmentToEntity(e)

	// Check if exists
	var existing entities.SequenceEnrollmentEntity
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

// FindByID finds an enrollment by ID
func (r *GormSequenceEnrollmentRepository) FindByID(id uuid.UUID) (*sequence.SequenceEnrollment, error) {
	var entity entities.SequenceEnrollmentEntity
	err := r.db.Where("id = ?", id).First(&entity).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return r.enrollmentToDomain(&entity), nil
}

// FindBySequenceID finds all enrollments for a sequence
func (r *GormSequenceEnrollmentRepository) FindBySequenceID(sequenceID uuid.UUID) ([]*sequence.SequenceEnrollment, error) {
	var entities []entities.SequenceEnrollmentEntity
	err := r.db.Where("sequence_id = ?", sequenceID).Order("enrolled_at DESC").Find(&entities).Error
	if err != nil {
		return nil, err
	}

	return r.enrollmentToDomainSlice(entities), nil
}

// FindByContactID finds all enrollments for a contact
func (r *GormSequenceEnrollmentRepository) FindByContactID(contactID uuid.UUID) ([]*sequence.SequenceEnrollment, error) {
	var entities []entities.SequenceEnrollmentEntity
	err := r.db.Where("contact_id = ?", contactID).Order("enrolled_at DESC").Find(&entities).Error
	if err != nil {
		return nil, err
	}

	return r.enrollmentToDomainSlice(entities), nil
}

// FindReadyForNextStep finds enrollments ready for the next step
func (r *GormSequenceEnrollmentRepository) FindReadyForNextStep() ([]*sequence.SequenceEnrollment, error) {
	var entities []entities.SequenceEnrollmentEntity
	err := r.db.Where("status = ? AND next_scheduled_at <= NOW()", string(sequence.EnrollmentStatusActive)).
		Order("next_scheduled_at ASC").
		Find(&entities).Error
	if err != nil {
		return nil, err
	}

	return r.enrollmentToDomainSlice(entities), nil
}

// FindActiveBySequenceAndContact finds an active enrollment for a sequence and contact
func (r *GormSequenceEnrollmentRepository) FindActiveBySequenceAndContact(sequenceID, contactID uuid.UUID) (*sequence.SequenceEnrollment, error) {
	var entity entities.SequenceEnrollmentEntity
	err := r.db.Where("sequence_id = ? AND contact_id = ? AND status = ?",
		sequenceID, contactID, string(sequence.EnrollmentStatusActive)).
		First(&entity).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return r.enrollmentToDomain(&entity), nil
}

// Delete deletes an enrollment
func (r *GormSequenceEnrollmentRepository) Delete(id uuid.UUID) error {
	result := r.db.Delete(&entities.SequenceEnrollmentEntity{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// Enrollment converters

func (r *GormSequenceEnrollmentRepository) enrollmentToEntity(e *sequence.SequenceEnrollment) *entities.SequenceEnrollmentEntity {
	return &entities.SequenceEnrollmentEntity{
		ID:               e.ID(),
		SequenceID:       e.SequenceID(),
		ContactID:        e.ContactID(),
		Status:           string(e.Status()),
		CurrentStepOrder: e.CurrentStepOrder(),
		NextScheduledAt:  e.NextScheduledAt(),
		ExitedAt:         e.ExitedAt(),
		ExitReason:       e.ExitReason(),
		CompletedAt:      e.CompletedAt(),
		EnrolledAt:       e.EnrolledAt(),
		UpdatedAt:        e.UpdatedAt(),
	}
}

func (r *GormSequenceEnrollmentRepository) enrollmentToDomain(entity *entities.SequenceEnrollmentEntity) *sequence.SequenceEnrollment {
	return sequence.ReconstructEnrollment(
		entity.ID,
		entity.SequenceID,
		entity.ContactID,
		sequence.EnrollmentStatus(entity.Status),
		entity.CurrentStepOrder,
		entity.NextScheduledAt,
		entity.ExitedAt,
		entity.ExitReason,
		entity.CompletedAt,
		entity.EnrolledAt,
		entity.UpdatedAt,
	)
}

func (r *GormSequenceEnrollmentRepository) enrollmentToDomainSlice(entities []entities.SequenceEnrollmentEntity) []*sequence.SequenceEnrollment {
	enrollments := make([]*sequence.SequenceEnrollment, len(entities))
	for i, entity := range entities {
		enrollments[i] = r.enrollmentToDomain(&entity)
	}
	return enrollments
}
