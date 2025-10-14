package persistence

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/ventros/crm/infrastructure/persistence/entities"
	"github.com/ventros/crm/internal/domain/core/shared"
	"gorm.io/gorm"
)

// GormUsageMeterRepository implementa o repositório de usage meters usando GORM
type GormUsageMeterRepository struct {
	db *gorm.DB
}

// NewGormUsageMeterRepository cria uma nova instância do repositório
func NewGormUsageMeterRepository(db *gorm.DB) *GormUsageMeterRepository {
	return &GormUsageMeterRepository{db: db}
}

// Create cria um novo usage meter
func (r *GormUsageMeterRepository) Create(ctx context.Context, meter *entities.UsageMeterEntity) error {
	return r.db.WithContext(ctx).Create(meter).Error
}

// FindByID busca um usage meter por ID
func (r *GormUsageMeterRepository) FindByID(ctx context.Context, id uuid.UUID) (*entities.UsageMeterEntity, error) {
	var meter entities.UsageMeterEntity
	err := r.db.WithContext(ctx).First(&meter, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &meter, nil
}

// FindByBillingAccountAndMetric busca um usage meter por billing account e métrica
func (r *GormUsageMeterRepository) FindByBillingAccountAndMetric(ctx context.Context, billingAccountID uuid.UUID, metricName string) (*entities.UsageMeterEntity, error) {
	var meter entities.UsageMeterEntity
	err := r.db.WithContext(ctx).
		Where("billing_account_id = ? AND metric_name = ?", billingAccountID, metricName).
		Order("created_at DESC").
		First(&meter).Error
	if err != nil {
		return nil, err
	}
	return &meter, nil
}

// FindByBillingAccountID busca todos os usage meters de uma billing account
func (r *GormUsageMeterRepository) FindByBillingAccountID(ctx context.Context, billingAccountID uuid.UUID) ([]*entities.UsageMeterEntity, error) {
	var meters []*entities.UsageMeterEntity
	err := r.db.WithContext(ctx).
		Where("billing_account_id = ?", billingAccountID).
		Order("created_at DESC").
		Find(&meters).Error
	return meters, err
}

// FindByStripeCustomerID busca todos os usage meters de um Stripe customer
func (r *GormUsageMeterRepository) FindByStripeCustomerID(ctx context.Context, stripeCustomerID string) ([]*entities.UsageMeterEntity, error) {
	var meters []*entities.UsageMeterEntity
	err := r.db.WithContext(ctx).
		Where("stripe_customer_id = ?", stripeCustomerID).
		Order("created_at DESC").
		Find(&meters).Error
	return meters, err
}

// FindPendingReports busca usage meters que precisam reportar ao Stripe
// (não reportados ou reportados há mais de 1 hora)
func (r *GormUsageMeterRepository) FindPendingReports(ctx context.Context, limit int) ([]*entities.UsageMeterEntity, error) {
	var meters []*entities.UsageMeterEntity
	oneHourAgo := time.Now().Add(-1 * time.Hour)

	query := r.db.WithContext(ctx).
		Where("last_reported_at IS NULL OR last_reported_at < ?", oneHourAgo).
		Where("quantity > 0").
		Order("last_reported_at ASC NULLS FIRST")

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Find(&meters).Error
	return meters, err
}

// FindByPeriod busca usage meters por período
func (r *GormUsageMeterRepository) FindByPeriod(ctx context.Context, periodStart, periodEnd time.Time) ([]*entities.UsageMeterEntity, error) {
	var meters []*entities.UsageMeterEntity
	err := r.db.WithContext(ctx).
		Where("period_start >= ? AND period_end <= ?", periodStart, periodEnd).
		Order("created_at DESC").
		Find(&meters).Error
	return meters, err
}

// Update atualiza um usage meter
func (r *GormUsageMeterRepository) Update(ctx context.Context, meter *entities.UsageMeterEntity) error {
	// Check if exists
	var existing entities.UsageMeterEntity
	err := r.db.WithContext(ctx).Where("id = ?", meter.ID).First(&existing).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Insert if not found
			return r.db.WithContext(ctx).Create(meter).Error
		}
		return err
	}

	// Update with optimistic locking
	result := r.db.WithContext(ctx).Model(&entities.UsageMeterEntity{}).
		Where("id = ? AND version = ?", meter.ID, existing.Version).
		Updates(map[string]interface{}{
			"version":            existing.Version + 1, // Increment version
			"billing_account_id": meter.BillingAccountID,
			"stripe_customer_id": meter.StripeCustomerID,
			"stripe_meter_id":    meter.StripeMeterID,
			"metric_name":        meter.MetricName,
			"event_name":         meter.EventName,
			"quantity":           meter.Quantity,
			"period_start":       meter.PeriodStart,
			"period_end":         meter.PeriodEnd,
			"last_reported_at":   meter.LastReportedAt,
			"metadata":           meter.Metadata,
			"updated_at":         meter.UpdatedAt,
		})

	if result.Error != nil {
		return result.Error
	}

	// Check optimistic locking - if 0 rows affected, version mismatch (concurrent update)
	if result.RowsAffected == 0 {
		return shared.NewOptimisticLockError(
			"UsageMeter",
			meter.ID.String(),
			existing.Version,
			meter.Version,
		)
	}

	return nil
}

// Delete deleta um usage meter (soft delete)
func (r *GormUsageMeterRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entities.UsageMeterEntity{}, "id = ?", id).Error
}
