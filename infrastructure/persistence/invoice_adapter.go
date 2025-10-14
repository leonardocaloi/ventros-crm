package persistence

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/ventros/crm/infrastructure/persistence/entities"
	"github.com/ventros/crm/internal/domain/core/billing"
	"gorm.io/gorm"
)

// InvoiceRepositoryAdapter adapta o GormInvoiceRepository para implementar billing.InvoiceRepository
type InvoiceRepositoryAdapter struct {
	gormRepo *GormInvoiceRepository
}

// NewInvoiceRepositoryAdapter cria um novo adapter
func NewInvoiceRepositoryAdapter(db *gorm.DB) billing.InvoiceRepository {
	return &InvoiceRepositoryAdapter{
		gormRepo: NewGormInvoiceRepository(db),
	}
}

// Create cria uma nova invoice
func (a *InvoiceRepositoryAdapter) Create(ctx context.Context, invoice *billing.Invoice) error {
	entity, err := a.domainToEntity(invoice)
	if err != nil {
		return err
	}
	return a.gormRepo.Create(ctx, entity)
}

// FindByID busca uma invoice por ID
func (a *InvoiceRepositoryAdapter) FindByID(ctx context.Context, id uuid.UUID) (*billing.Invoice, error) {
	entity, err := a.gormRepo.FindByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, billing.ErrNotFound
		}
		return nil, err
	}
	return a.entityToDomain(entity)
}

// FindByStripeInvoiceID busca uma invoice pelo Stripe Invoice ID
func (a *InvoiceRepositoryAdapter) FindByStripeInvoiceID(ctx context.Context, stripeInvoiceID string) (*billing.Invoice, error) {
	entity, err := a.gormRepo.FindByStripeInvoiceID(ctx, stripeInvoiceID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, billing.ErrNotFound
		}
		return nil, err
	}
	return a.entityToDomain(entity)
}

// FindByBillingAccount busca todas as invoices de uma billing account
func (a *InvoiceRepositoryAdapter) FindByBillingAccount(ctx context.Context, billingAccountID uuid.UUID) ([]*billing.Invoice, error) {
	entities, err := a.gormRepo.FindByBillingAccountID(ctx, billingAccountID)
	if err != nil {
		return nil, err
	}

	invoices := make([]*billing.Invoice, len(entities))
	for i, entity := range entities {
		invoice, err := a.entityToDomain(entity)
		if err != nil {
			return nil, err
		}
		invoices[i] = invoice
	}
	return invoices, nil
}

// FindBySubscription busca todas as invoices de uma subscription
func (a *InvoiceRepositoryAdapter) FindBySubscription(ctx context.Context, subscriptionID uuid.UUID) ([]*billing.Invoice, error) {
	entities, err := a.gormRepo.FindBySubscriptionID(ctx, subscriptionID)
	if err != nil {
		return nil, err
	}

	invoices := make([]*billing.Invoice, len(entities))
	for i, entity := range entities {
		invoice, err := a.entityToDomain(entity)
		if err != nil {
			return nil, err
		}
		invoices[i] = invoice
	}
	return invoices, nil
}

// FindByStatus busca invoices por status
func (a *InvoiceRepositoryAdapter) FindByStatus(ctx context.Context, status billing.InvoiceStatus) ([]*billing.Invoice, error) {
	entities, err := a.gormRepo.FindByStatus(ctx, string(status), 0)
	if err != nil {
		return nil, err
	}

	invoices := make([]*billing.Invoice, len(entities))
	for i, entity := range entities {
		invoice, err := a.entityToDomain(entity)
		if err != nil {
			return nil, err
		}
		invoices[i] = invoice
	}
	return invoices, nil
}

// FindOverdue busca invoices vencidas
func (a *InvoiceRepositoryAdapter) FindOverdue(ctx context.Context) ([]*billing.Invoice, error) {
	entities, err := a.gormRepo.FindOverdueInvoices(ctx)
	if err != nil {
		return nil, err
	}

	invoices := make([]*billing.Invoice, len(entities))
	for i, entity := range entities {
		invoice, err := a.entityToDomain(entity)
		if err != nil {
			return nil, err
		}
		invoices[i] = invoice
	}
	return invoices, nil
}

// FindUnpaidByBillingAccount busca invoices não pagas de uma billing account
func (a *InvoiceRepositoryAdapter) FindUnpaidByBillingAccount(ctx context.Context, billingAccountID uuid.UUID) ([]*billing.Invoice, error) {
	var entities []*entities.InvoiceEntity
	err := a.gormRepo.db.WithContext(ctx).
		Where("billing_account_id = ? AND status IN (?)", billingAccountID, []string{"open", "past_due"}).
		Order("created_at DESC").
		Find(&entities).Error
	if err != nil {
		return nil, err
	}

	invoices := make([]*billing.Invoice, len(entities))
	for i, entity := range entities {
		invoice, err := a.entityToDomain(entity)
		if err != nil {
			return nil, err
		}
		invoices[i] = invoice
	}
	return invoices, nil
}

// FindDueInPeriod busca invoices com vencimento em um período
func (a *InvoiceRepositoryAdapter) FindDueInPeriod(ctx context.Context, start, end time.Time) ([]*billing.Invoice, error) {
	var entities []*entities.InvoiceEntity
	err := a.gormRepo.db.WithContext(ctx).
		Where("due_date >= ? AND due_date <= ?", start, end).
		Order("due_date ASC").
		Find(&entities).Error
	if err != nil {
		return nil, err
	}

	invoices := make([]*billing.Invoice, len(entities))
	for i, entity := range entities {
		invoice, err := a.entityToDomain(entity)
		if err != nil {
			return nil, err
		}
		invoices[i] = invoice
	}
	return invoices, nil
}

// Update atualiza uma invoice existente
func (a *InvoiceRepositoryAdapter) Update(ctx context.Context, invoice *billing.Invoice) error {
	entity, err := a.domainToEntity(invoice)
	if err != nil {
		return err
	}
	return a.gormRepo.Update(ctx, entity)
}

// Delete remove uma invoice
func (a *InvoiceRepositoryAdapter) Delete(ctx context.Context, id uuid.UUID) error {
	return a.gormRepo.Delete(ctx, id)
}

// domainToEntity converte billing.Invoice para entities.InvoiceEntity
func (a *InvoiceRepositoryAdapter) domainToEntity(invoice *billing.Invoice) (*entities.InvoiceEntity, error) {
	// Converter metadata para JSON bytes
	var metadataBytes []byte
	if len(invoice.Metadata()) > 0 {
		var err error
		metadataBytes, err = json.Marshal(invoice.Metadata())
		if err != nil {
			return nil, err
		}
	}

	return &entities.InvoiceEntity{
		ID:                   invoice.ID(),
		Version:              invoice.Version(),
		BillingAccountID:     invoice.BillingAccountID(),
		StripeInvoiceID:      invoice.StripeInvoiceID(),
		SubscriptionID:       invoice.SubscriptionID(),
		StripeSubscriptionID: invoice.StripeSubscriptionID(),
		AmountDue:            invoice.AmountDue(),
		AmountPaid:           invoice.AmountPaid(),
		AmountRemaining:      invoice.AmountRemaining(),
		Currency:             invoice.Currency(),
		Status:               string(invoice.Status()),
		HostedInvoiceURL:     invoice.HostedInvoiceURL(),
		InvoicePDF:           invoice.InvoicePDF(),
		DueDate:              invoice.DueDate(),
		PaidAt:               invoice.PaidAt(),
		Metadata:             metadataBytes,
		CreatedAt:            invoice.CreatedAt(),
		UpdatedAt:            invoice.UpdatedAt(),
	}, nil
}

// entityToDomain converte entities.InvoiceEntity para billing.Invoice
func (a *InvoiceRepositoryAdapter) entityToDomain(entity *entities.InvoiceEntity) (*billing.Invoice, error) {
	// Converter JSON bytes para metadata
	var metadata map[string]string
	if len(entity.Metadata) > 0 {
		err := json.Unmarshal(entity.Metadata, &metadata)
		if err != nil {
			return nil, err
		}
	}

	return billing.ReconstructInvoice(
		entity.ID,
		entity.Version,
		entity.BillingAccountID,
		entity.StripeInvoiceID,
		entity.SubscriptionID,
		entity.StripeSubscriptionID,
		entity.AmountDue,
		entity.AmountPaid,
		entity.AmountRemaining,
		entity.Currency,
		billing.InvoiceStatus(entity.Status),
		entity.HostedInvoiceURL,
		entity.InvoicePDF,
		entity.DueDate,
		entity.PaidAt,
		metadata,
		entity.CreatedAt,
		entity.UpdatedAt,
	), nil
}
