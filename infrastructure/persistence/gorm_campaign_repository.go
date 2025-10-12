package persistence

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/caloi/ventros-crm/infrastructure/persistence/entities"
	"github.com/caloi/ventros-crm/internal/domain/automation/campaign"
	"github.com/caloi/ventros-crm/internal/domain/core/shared"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// GormCampaignRepository implements campaign.Repository using GORM
type GormCampaignRepository struct {
	db *gorm.DB
}

// NewGormCampaignRepository creates a new instance
func NewGormCampaignRepository(db *gorm.DB) campaign.Repository {
	return &GormCampaignRepository{db: db}
}

// Save saves or updates a campaign with optimistic locking
func (r *GormCampaignRepository) Save(c *campaign.Campaign) error {
	entity, err := r.toEntity(c)
	if err != nil {
		return fmt.Errorf("failed to convert campaign to entity: %w", err)
	}

	// Check if exists
	var existing entities.CampaignEntity
	err = r.db.Where("id = ?", entity.ID).First(&existing).Error

	if err == nil {
		// UPDATE with optimistic locking - use transaction
		return r.db.Transaction(func(tx *gorm.DB) error {
			// Update campaign with version check (optimistic locking)
			result := tx.Model(&entities.CampaignEntity{}).
				Where("id = ? AND version = ?", entity.ID, existing.Version).
				Updates(map[string]interface{}{
					"version":           existing.Version + 1, // Increment version
					"tenant_id":         entity.TenantID,
					"name":              entity.Name,
					"description":       entity.Description,
					"status":            entity.Status,
					"goal_type":         entity.GoalType,
					"goal_value":        entity.GoalValue,
					"contacts_reached":  entity.ContactsReached,
					"conversions_count": entity.ConversionsCount,
					"start_date":        entity.StartDate,
					"end_date":          entity.EndDate,
					"updated_at":        entity.UpdatedAt,
				})

			if result.Error != nil {
				return result.Error
			}

			// Check optimistic locking - if 0 rows affected, version mismatch (concurrent update)
			if result.RowsAffected == 0 {
				return shared.NewOptimisticLockError(
					"Campaign",
					entity.ID.String(),
					existing.Version,
					entity.Version,
				)
			}

			// Delete existing steps
			if err := tx.Where("campaign_id = ?", entity.ID).Delete(&entities.CampaignStepEntity{}).Error; err != nil {
				return err
			}

			// Insert new steps
			steps := r.stepsToEntities(c.Steps(), c.ID())
			if len(steps) > 0 {
				if err := tx.Create(&steps).Error; err != nil {
					return err
				}
			}

			return nil
		})
	} else if errors.Is(err, gorm.ErrRecordNotFound) {
		// INSERT - use transaction to insert campaign and steps
		return r.db.Transaction(func(tx *gorm.DB) error {
			// Create campaign
			if err := tx.Create(entity).Error; err != nil {
				return err
			}

			// Create steps
			steps := r.stepsToEntities(c.Steps(), c.ID())
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

// FindByID finds a campaign by ID
func (r *GormCampaignRepository) FindByID(id uuid.UUID) (*campaign.Campaign, error) {
	var entity entities.CampaignEntity
	err := r.db.Where("id = ?", id).First(&entity).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	// Load steps
	var stepEntities []entities.CampaignStepEntity
	if err := r.db.Where("campaign_id = ?", id).Order("\"order\" ASC").Find(&stepEntities).Error; err != nil {
		return nil, err
	}

	return r.toDomain(&entity, stepEntities)
}

// FindByTenantID finds all campaigns for a tenant
func (r *GormCampaignRepository) FindByTenantID(tenantID string) ([]*campaign.Campaign, error) {
	var campaignEntities []entities.CampaignEntity
	err := r.db.Where("tenant_id = ?", tenantID).Order("created_at DESC").Find(&campaignEntities).Error
	if err != nil {
		return nil, err
	}

	return r.toDomainSlice(campaignEntities)
}

// FindActiveByStatus finds campaigns by status
func (r *GormCampaignRepository) FindActiveByStatus(status campaign.CampaignStatus) ([]*campaign.Campaign, error) {
	var campaignEntities []entities.CampaignEntity
	err := r.db.Where("status = ?", string(status)).Order("created_at DESC").Find(&campaignEntities).Error
	if err != nil {
		return nil, err
	}

	return r.toDomainSlice(campaignEntities)
}

// FindScheduled finds campaigns scheduled to start
func (r *GormCampaignRepository) FindScheduled() ([]*campaign.Campaign, error) {
	var campaignEntities []entities.CampaignEntity
	err := r.db.Where("status = ? AND start_date <= NOW()", string(campaign.CampaignStatusScheduled)).
		Order("start_date ASC").
		Find(&campaignEntities).Error
	if err != nil {
		return nil, err
	}

	return r.toDomainSlice(campaignEntities)
}

// Delete deletes a campaign
func (r *GormCampaignRepository) Delete(id uuid.UUID) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Delete steps first
		if err := tx.Where("campaign_id = ?", id).Delete(&entities.CampaignStepEntity{}).Error; err != nil {
			return err
		}

		// Delete campaign
		result := tx.Delete(&entities.CampaignEntity{}, "id = ?", id)
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

func (r *GormCampaignRepository) toEntity(c *campaign.Campaign) (*entities.CampaignEntity, error) {
	if c == nil {
		return nil, errors.New("campaign cannot be nil")
	}

	return &entities.CampaignEntity{
		ID:               c.ID(),
		Version:          c.Version(),
		TenantID:         c.TenantID(),
		Name:             c.Name(),
		Description:      c.Description(),
		Status:           string(c.Status()),
		GoalType:         string(c.GoalType()),
		GoalValue:        c.GoalValue(),
		ContactsReached:  c.ContactsReached(),
		ConversionsCount: c.ConversionsCount(),
		StartDate:        c.StartDate(),
		EndDate:          c.EndDate(),
		CreatedAt:        c.CreatedAt(),
		UpdatedAt:        c.UpdatedAt(),
	}, nil
}

func (r *GormCampaignRepository) stepsToEntities(steps []campaign.CampaignStep, campaignID uuid.UUID) []entities.CampaignStepEntity {
	stepEntities := make([]entities.CampaignStepEntity, len(steps))
	for i, step := range steps {
		configJSON, _ := json.Marshal(step.Config)
		conditionsJSON, _ := json.Marshal(step.Conditions)

		stepEntities[i] = entities.CampaignStepEntity{
			ID:         step.ID,
			CampaignID: campaignID,
			Order:      step.Order,
			Name:       step.Name,
			Type:       string(step.Type),
			Config:     configJSON,
			Conditions: conditionsJSON,
			CreatedAt:  step.CreatedAt,
		}
	}
	return stepEntities
}

func (r *GormCampaignRepository) toDomain(entity *entities.CampaignEntity, stepEntities []entities.CampaignStepEntity) (*campaign.Campaign, error) {
	if entity == nil {
		return nil, errors.New("entity cannot be nil")
	}

	// Convert steps
	steps := make([]campaign.CampaignStep, len(stepEntities))
	for i, stepEntity := range stepEntities {
		var config campaign.StepConfig
		if len(stepEntity.Config) > 0 {
			if err := json.Unmarshal(stepEntity.Config, &config); err != nil {
				return nil, fmt.Errorf("failed to unmarshal step config: %w", err)
			}
		}

		var conditions []campaign.StepCondition
		if len(stepEntity.Conditions) > 0 {
			if err := json.Unmarshal(stepEntity.Conditions, &conditions); err != nil {
				return nil, fmt.Errorf("failed to unmarshal conditions: %w", err)
			}
		}

		steps[i] = campaign.CampaignStep{
			ID:         stepEntity.ID,
			Order:      stepEntity.Order,
			Name:       stepEntity.Name,
			Type:       campaign.StepType(stepEntity.Type),
			Config:     config,
			Conditions: conditions,
			CreatedAt:  stepEntity.CreatedAt,
		}
	}

	// Reconstruct domain model
	return campaign.ReconstructCampaign(
		entity.ID,
		entity.Version,
		entity.TenantID,
		entity.Name,
		entity.Description,
		campaign.CampaignStatus(entity.Status),
		steps,
		campaign.GoalType(entity.GoalType),
		entity.GoalValue,
		entity.ContactsReached,
		entity.ConversionsCount,
		entity.StartDate,
		entity.EndDate,
		entity.CreatedAt,
		entity.UpdatedAt,
	), nil
}

func (r *GormCampaignRepository) toDomainSlice(campaignEntities []entities.CampaignEntity) ([]*campaign.Campaign, error) {
	campaigns := make([]*campaign.Campaign, 0, len(campaignEntities))
	for _, entity := range campaignEntities {
		// Load steps for each campaign
		var stepEntities []entities.CampaignStepEntity
		if err := r.db.Where("campaign_id = ?", entity.ID).Order("\"order\" ASC").Find(&stepEntities).Error; err != nil {
			return nil, err
		}

		c, err := r.toDomain(&entity, stepEntities)
		if err != nil {
			return nil, err
		}
		campaigns = append(campaigns, c)
	}
	return campaigns, nil
}

// GormCampaignEnrollmentRepository implements campaign.EnrollmentRepository using GORM
type GormCampaignEnrollmentRepository struct {
	db *gorm.DB
}

// NewGormCampaignEnrollmentRepository creates a new instance
func NewGormCampaignEnrollmentRepository(db *gorm.DB) campaign.EnrollmentRepository {
	return &GormCampaignEnrollmentRepository{db: db}
}

// Save saves or updates an enrollment
func (r *GormCampaignEnrollmentRepository) Save(e *campaign.CampaignEnrollment) error {
	entity := r.enrollmentToEntity(e)

	// Check if exists
	var existing entities.CampaignEnrollmentEntity
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
func (r *GormCampaignEnrollmentRepository) FindByID(id uuid.UUID) (*campaign.CampaignEnrollment, error) {
	var entity entities.CampaignEnrollmentEntity
	err := r.db.Where("id = ?", id).First(&entity).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return r.enrollmentToDomain(&entity), nil
}

// FindByCampaignID finds all enrollments for a campaign
func (r *GormCampaignEnrollmentRepository) FindByCampaignID(campaignID uuid.UUID) ([]*campaign.CampaignEnrollment, error) {
	var enrollmentEntities []entities.CampaignEnrollmentEntity
	err := r.db.Where("campaign_id = ?", campaignID).Order("enrolled_at DESC").Find(&enrollmentEntities).Error
	if err != nil {
		return nil, err
	}

	return r.enrollmentToDomainSlice(enrollmentEntities), nil
}

// FindByContactID finds all enrollments for a contact
func (r *GormCampaignEnrollmentRepository) FindByContactID(contactID uuid.UUID) ([]*campaign.CampaignEnrollment, error) {
	var enrollmentEntities []entities.CampaignEnrollmentEntity
	err := r.db.Where("contact_id = ?", contactID).Order("enrolled_at DESC").Find(&enrollmentEntities).Error
	if err != nil {
		return nil, err
	}

	return r.enrollmentToDomainSlice(enrollmentEntities), nil
}

// FindReadyForNextStep finds enrollments ready for the next step
func (r *GormCampaignEnrollmentRepository) FindReadyForNextStep() ([]*campaign.CampaignEnrollment, error) {
	var enrollmentEntities []entities.CampaignEnrollmentEntity
	err := r.db.Where("status = ? AND next_scheduled_at <= NOW()", string(campaign.EnrollmentStatusActive)).
		Order("next_scheduled_at ASC").
		Find(&enrollmentEntities).Error
	if err != nil {
		return nil, err
	}

	return r.enrollmentToDomainSlice(enrollmentEntities), nil
}

// FindActiveByCampaignAndContact finds an active enrollment for a campaign and contact
func (r *GormCampaignEnrollmentRepository) FindActiveByCampaignAndContact(campaignID, contactID uuid.UUID) (*campaign.CampaignEnrollment, error) {
	var entity entities.CampaignEnrollmentEntity
	err := r.db.Where("campaign_id = ? AND contact_id = ? AND status = ?",
		campaignID, contactID, string(campaign.EnrollmentStatusActive)).
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
func (r *GormCampaignEnrollmentRepository) Delete(id uuid.UUID) error {
	result := r.db.Delete(&entities.CampaignEnrollmentEntity{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// Enrollment converters

func (r *GormCampaignEnrollmentRepository) enrollmentToEntity(e *campaign.CampaignEnrollment) *entities.CampaignEnrollmentEntity {
	return &entities.CampaignEnrollmentEntity{
		ID:               e.ID(),
		CampaignID:       e.CampaignID(),
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

func (r *GormCampaignEnrollmentRepository) enrollmentToDomain(entity *entities.CampaignEnrollmentEntity) *campaign.CampaignEnrollment {
	return campaign.ReconstructEnrollment(
		entity.ID,
		entity.CampaignID,
		entity.ContactID,
		campaign.EnrollmentStatus(entity.Status),
		entity.CurrentStepOrder,
		entity.NextScheduledAt,
		entity.ExitedAt,
		entity.ExitReason,
		entity.CompletedAt,
		entity.EnrolledAt,
		entity.UpdatedAt,
	)
}

func (r *GormCampaignEnrollmentRepository) enrollmentToDomainSlice(enrollmentEntities []entities.CampaignEnrollmentEntity) []*campaign.CampaignEnrollment {
	enrollments := make([]*campaign.CampaignEnrollment, len(enrollmentEntities))
	for i, entity := range enrollmentEntities {
		enrollments[i] = r.enrollmentToDomain(&entity)
	}
	return enrollments
}
