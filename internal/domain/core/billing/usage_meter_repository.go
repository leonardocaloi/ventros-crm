package billing

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// UsageMeterRepository define operações de persistência para UsageMeters
type UsageMeterRepository interface {
	// Create cria um novo usage meter
	Create(ctx context.Context, meter *UsageMeter) error

	// FindByID busca um meter por ID
	FindByID(ctx context.Context, id uuid.UUID) (*UsageMeter, error)

	// FindByBillingAccount busca todos os meters de uma billing account
	FindByBillingAccount(ctx context.Context, billingAccountID uuid.UUID) ([]*UsageMeter, error)

	// FindByMetricName busca meters por nome de métrica
	FindByMetricName(ctx context.Context, billingAccountID uuid.UUID, metricName string) (*UsageMeter, error)

	// FindByStripeMeterID busca meter pelo ID do Stripe
	FindByStripeMeterID(ctx context.Context, stripeMeterID string) (*UsageMeter, error)

	// Update atualiza um meter existente
	Update(ctx context.Context, meter *UsageMeter) error

	// Delete remove um meter
	Delete(ctx context.Context, id uuid.UUID) error

	// FindInCurrentPeriod busca meters que estão no período atual
	FindInCurrentPeriod(ctx context.Context) ([]*UsageMeter, error)

	// FindPendingReport busca meters que precisam reportar uso ao Stripe
	// (não reportados ou reportados há mais de X tempo)
	FindPendingReport(ctx context.Context, since time.Duration) ([]*UsageMeter, error)

	// FindByPeriod busca meters de um período específico
	FindByPeriod(ctx context.Context, start, end time.Time) ([]*UsageMeter, error)
}
