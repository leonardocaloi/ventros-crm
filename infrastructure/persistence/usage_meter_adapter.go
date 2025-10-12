package persistence

import (
	"context"
	"encoding/json"
	"time"

	"github.com/caloi/ventros-crm/infrastructure/persistence/entities"
	"github.com/caloi/ventros-crm/internal/domain/core/billing"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UsageMeterRepositoryAdapter adapta o GormUsageMeterRepository para implementar billing.UsageMeterRepository
type UsageMeterRepositoryAdapter struct {
	gormRepo *GormUsageMeterRepository
}

// NewUsageMeterRepositoryAdapter cria um novo adapter
func NewUsageMeterRepositoryAdapter(db *gorm.DB) billing.UsageMeterRepository {
	return &UsageMeterRepositoryAdapter{
		gormRepo: NewGormUsageMeterRepository(db),
	}
}

// Create cria um novo usage meter
func (a *UsageMeterRepositoryAdapter) Create(ctx context.Context, meter *billing.UsageMeter) error {
	entity, err := a.domainToEntity(meter)
	if err != nil {
		return err
	}
	return a.gormRepo.Create(ctx, entity)
}

// FindByID busca um usage meter por ID
func (a *UsageMeterRepositoryAdapter) FindByID(ctx context.Context, id uuid.UUID) (*billing.UsageMeter, error) {
	entity, err := a.gormRepo.FindByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, billing.ErrNotFound
		}
		return nil, err
	}
	return a.entityToDomain(entity)
}

// FindByBillingAccount busca todos os usage meters de uma billing account
func (a *UsageMeterRepositoryAdapter) FindByBillingAccount(ctx context.Context, billingAccountID uuid.UUID) ([]*billing.UsageMeter, error) {
	entities, err := a.gormRepo.FindByBillingAccountID(ctx, billingAccountID)
	if err != nil {
		return nil, err
	}

	meters := make([]*billing.UsageMeter, len(entities))
	for i, entity := range entities {
		meter, err := a.entityToDomain(entity)
		if err != nil {
			return nil, err
		}
		meters[i] = meter
	}
	return meters, nil
}

// FindByMetricName busca usage meter por billing account e nome da métrica
func (a *UsageMeterRepositoryAdapter) FindByMetricName(ctx context.Context, billingAccountID uuid.UUID, metricName string) (*billing.UsageMeter, error) {
	entity, err := a.gormRepo.FindByBillingAccountAndMetric(ctx, billingAccountID, metricName)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, billing.ErrNotFound
		}
		return nil, err
	}
	return a.entityToDomain(entity)
}

// FindByStripeMeterID busca usage meter pelo Stripe Meter ID
func (a *UsageMeterRepositoryAdapter) FindByStripeMeterID(ctx context.Context, stripeMeterID string) (*billing.UsageMeter, error) {
	var entity entities.UsageMeterEntity
	err := a.gormRepo.db.WithContext(ctx).
		Where("stripe_meter_id = ?", stripeMeterID).
		First(&entity).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, billing.ErrNotFound
		}
		return nil, err
	}
	return a.entityToDomain(&entity)
}

// FindInCurrentPeriod busca meters que estão no período atual
func (a *UsageMeterRepositoryAdapter) FindInCurrentPeriod(ctx context.Context) ([]*billing.UsageMeter, error) {
	now := time.Now()
	entities, err := a.gormRepo.FindByPeriod(ctx, now.Add(-30*24*time.Hour), now)
	if err != nil {
		return nil, err
	}

	meters := make([]*billing.UsageMeter, len(entities))
	for i, entity := range entities {
		meter, err := a.entityToDomain(entity)
		if err != nil {
			return nil, err
		}
		meters[i] = meter
	}
	return meters, nil
}

// FindPendingReport busca meters que precisam reportar uso ao Stripe
func (a *UsageMeterRepositoryAdapter) FindPendingReport(ctx context.Context, since time.Duration) ([]*billing.UsageMeter, error) {
	entities, err := a.gormRepo.FindPendingReports(ctx, 100)
	if err != nil {
		return nil, err
	}

	meters := make([]*billing.UsageMeter, len(entities))
	for i, entity := range entities {
		meter, err := a.entityToDomain(entity)
		if err != nil {
			return nil, err
		}
		meters[i] = meter
	}
	return meters, nil
}

// FindByPeriod busca meters de um período específico
func (a *UsageMeterRepositoryAdapter) FindByPeriod(ctx context.Context, start, end time.Time) ([]*billing.UsageMeter, error) {
	entities, err := a.gormRepo.FindByPeriod(ctx, start, end)
	if err != nil {
		return nil, err
	}

	meters := make([]*billing.UsageMeter, len(entities))
	for i, entity := range entities {
		meter, err := a.entityToDomain(entity)
		if err != nil {
			return nil, err
		}
		meters[i] = meter
	}
	return meters, nil
}

// Update atualiza um usage meter existente
func (a *UsageMeterRepositoryAdapter) Update(ctx context.Context, meter *billing.UsageMeter) error {
	entity, err := a.domainToEntity(meter)
	if err != nil {
		return err
	}
	return a.gormRepo.Update(ctx, entity)
}

// Delete remove um usage meter
func (a *UsageMeterRepositoryAdapter) Delete(ctx context.Context, id uuid.UUID) error {
	return a.gormRepo.Delete(ctx, id)
}

// domainToEntity converte billing.UsageMeter para entities.UsageMeterEntity
func (a *UsageMeterRepositoryAdapter) domainToEntity(meter *billing.UsageMeter) (*entities.UsageMeterEntity, error) {
	// Converter metadata para JSON bytes
	var metadataBytes []byte
	if len(meter.Metadata()) > 0 {
		var err error
		metadataBytes, err = json.Marshal(meter.Metadata())
		if err != nil {
			return nil, err
		}
	}

	return &entities.UsageMeterEntity{
		ID:               meter.ID(),
		Version:          meter.Version(),
		BillingAccountID: meter.BillingAccountID(),
		StripeCustomerID: meter.StripeCustomerID(),
		StripeMeterID:    meter.StripeMeterID(),
		MetricName:       meter.MetricName(),
		EventName:        meter.EventName(),
		Quantity:         meter.Quantity(),
		PeriodStart:      meter.PeriodStart(),
		PeriodEnd:        meter.PeriodEnd(),
		LastReportedAt:   meter.LastReportedAt(),
		Metadata:         metadataBytes,
		CreatedAt:        meter.CreatedAt(),
		UpdatedAt:        meter.UpdatedAt(),
	}, nil
}

// entityToDomain converte entities.UsageMeterEntity para billing.UsageMeter
func (a *UsageMeterRepositoryAdapter) entityToDomain(entity *entities.UsageMeterEntity) (*billing.UsageMeter, error) {
	// Converter JSON bytes para metadata
	var metadata map[string]string
	if len(entity.Metadata) > 0 {
		err := json.Unmarshal(entity.Metadata, &metadata)
		if err != nil {
			return nil, err
		}
	}

	return billing.ReconstructUsageMeter(
		entity.ID,
		entity.Version,
		entity.BillingAccountID,
		entity.StripeCustomerID,
		entity.StripeMeterID,
		entity.MetricName,
		entity.EventName,
		entity.Quantity,
		entity.PeriodStart,
		entity.PeriodEnd,
		entity.LastReportedAt,
		metadata,
		entity.CreatedAt,
		entity.UpdatedAt,
	), nil
}
