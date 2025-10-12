package persistence

import (
	"context"
	"errors"
	"time"

	"github.com/caloi/ventros-crm/infrastructure/persistence/entities"
	"github.com/caloi/ventros-crm/internal/domain/core/shared"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// GormInvoiceRepository implementa o repositório de invoices usando GORM
type GormInvoiceRepository struct {
	db *gorm.DB
}

// NewGormInvoiceRepository cria uma nova instância do repositório
func NewGormInvoiceRepository(db *gorm.DB) *GormInvoiceRepository {
	return &GormInvoiceRepository{db: db}
}

// Create cria uma nova invoice
func (r *GormInvoiceRepository) Create(ctx context.Context, invoice *entities.InvoiceEntity) error {
	return r.db.WithContext(ctx).Create(invoice).Error
}

// FindByID busca uma invoice por ID
func (r *GormInvoiceRepository) FindByID(ctx context.Context, id uuid.UUID) (*entities.InvoiceEntity, error) {
	var invoice entities.InvoiceEntity
	err := r.db.WithContext(ctx).First(&invoice, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &invoice, nil
}

// FindByStripeInvoiceID busca uma invoice pelo Stripe Invoice ID
func (r *GormInvoiceRepository) FindByStripeInvoiceID(ctx context.Context, stripeInvoiceID string) (*entities.InvoiceEntity, error) {
	var invoice entities.InvoiceEntity
	err := r.db.WithContext(ctx).Where("stripe_invoice_id = ?", stripeInvoiceID).First(&invoice).Error
	if err != nil {
		return nil, err
	}
	return &invoice, nil
}

// FindByBillingAccountID busca todas as invoices de uma billing account
func (r *GormInvoiceRepository) FindByBillingAccountID(ctx context.Context, billingAccountID uuid.UUID) ([]*entities.InvoiceEntity, error) {
	var invoices []*entities.InvoiceEntity
	err := r.db.WithContext(ctx).
		Where("billing_account_id = ?", billingAccountID).
		Order("created_at DESC").
		Find(&invoices).Error
	return invoices, err
}

// FindBySubscriptionID busca todas as invoices de uma subscription
func (r *GormInvoiceRepository) FindBySubscriptionID(ctx context.Context, subscriptionID uuid.UUID) ([]*entities.InvoiceEntity, error) {
	var invoices []*entities.InvoiceEntity
	err := r.db.WithContext(ctx).
		Where("subscription_id = ?", subscriptionID).
		Order("created_at DESC").
		Find(&invoices).Error
	return invoices, err
}

// FindOverdueInvoices busca invoices vencidas (open e past due)
func (r *GormInvoiceRepository) FindOverdueInvoices(ctx context.Context) ([]*entities.InvoiceEntity, error) {
	var invoices []*entities.InvoiceEntity
	now := time.Now()
	err := r.db.WithContext(ctx).
		Where("status IN (?) AND due_date < ?", []string{"open", "past_due"}, now).
		Order("due_date ASC").
		Find(&invoices).Error
	return invoices, err
}

// FindByStatus busca invoices por status
func (r *GormInvoiceRepository) FindByStatus(ctx context.Context, status string, limit int) ([]*entities.InvoiceEntity, error) {
	var invoices []*entities.InvoiceEntity
	query := r.db.WithContext(ctx).
		Where("status = ?", status).
		Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Find(&invoices).Error
	return invoices, err
}

// Update atualiza uma invoice
func (r *GormInvoiceRepository) Update(ctx context.Context, invoice *entities.InvoiceEntity) error {
	// Check if exists
	var existing entities.InvoiceEntity
	err := r.db.WithContext(ctx).Where("id = ?", invoice.ID).First(&existing).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Insert if not found
			return r.db.WithContext(ctx).Create(invoice).Error
		}
		return err
	}

	// Update with optimistic locking
	result := r.db.WithContext(ctx).Model(&entities.InvoiceEntity{}).
		Where("id = ? AND version = ?", invoice.ID, existing.Version).
		Updates(map[string]interface{}{
			"version":                 existing.Version + 1, // Increment version
			"billing_account_id":      invoice.BillingAccountID,
			"stripe_invoice_id":       invoice.StripeInvoiceID,
			"subscription_id":         invoice.SubscriptionID,
			"stripe_subscription_id":  invoice.StripeSubscriptionID,
			"amount_due":              invoice.AmountDue,
			"amount_paid":             invoice.AmountPaid,
			"amount_remaining":        invoice.AmountRemaining,
			"currency":                invoice.Currency,
			"status":                  invoice.Status,
			"hosted_invoice_url":      invoice.HostedInvoiceURL,
			"invoice_pdf":             invoice.InvoicePDF,
			"due_date":                invoice.DueDate,
			"paid_at":                 invoice.PaidAt,
			"metadata":                invoice.Metadata,
			"updated_at":              invoice.UpdatedAt,
		})

	if result.Error != nil {
		return result.Error
	}

	// Check optimistic locking - if 0 rows affected, version mismatch (concurrent update)
	if result.RowsAffected == 0 {
		return shared.NewOptimisticLockError(
			"Invoice",
			invoice.ID.String(),
			existing.Version,
			invoice.Version,
		)
	}

	return nil
}

// Delete deleta uma invoice (soft delete)
func (r *GormInvoiceRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entities.InvoiceEntity{}, "id = ?", id).Error
}
