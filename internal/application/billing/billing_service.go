package billing

import (
	"context"
	"fmt"

	"github.com/caloi/ventros-crm/internal/domain/core/billing"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// BillingService gerencia contas de faturamento
type BillingService struct {
	repo   billing.Repository
	logger *zap.Logger
}

// NewBillingService cria uma nova instância do serviço
func NewBillingService(repo billing.Repository, logger *zap.Logger) *BillingService {
	return &BillingService{
		repo:   repo,
		logger: logger,
	}
}

// CreateBillingAccount cria uma nova conta de faturamento
func (s *BillingService) CreateBillingAccount(ctx context.Context, userID uuid.UUID, name, billingEmail string) (*billing.BillingAccount, error) {
	// Criar conta de faturamento
	account, err := billing.NewBillingAccount(userID, name, billingEmail)
	if err != nil {
		return nil, fmt.Errorf("failed to create billing account: %w", err)
	}

	// Persistir
	if err := s.repo.Create(ctx, account); err != nil {
		s.logger.Error("Failed to save billing account",
			zap.Error(err),
			zap.String("user_id", userID.String()),
		)
		return nil, fmt.Errorf("failed to save billing account: %w", err)
	}

	s.logger.Info("Billing account created",
		zap.String("account_id", account.ID().String()),
		zap.String("user_id", userID.String()),
		zap.String("name", name),
	)

	return account, nil
}

// GetBillingAccount busca uma conta por ID
func (s *BillingService) GetBillingAccount(ctx context.Context, id uuid.UUID) (*billing.BillingAccount, error) {
	return s.repo.FindByID(ctx, id)
}

// GetUserBillingAccounts busca todas as contas de um usuário
func (s *BillingService) GetUserBillingAccounts(ctx context.Context, userID uuid.UUID) ([]*billing.BillingAccount, error) {
	return s.repo.FindByUserID(ctx, userID)
}

// GetActiveBillingAccount busca a primeira conta ativa de um usuário
func (s *BillingService) GetActiveBillingAccount(ctx context.Context, userID uuid.UUID) (*billing.BillingAccount, error) {
	return s.repo.FindActiveByUserID(ctx, userID)
}

// ActivatePaymentFake ativa o pagamento de forma fake (para desenvolvimento)
func (s *BillingService) ActivatePaymentFake(ctx context.Context, accountID uuid.UUID) error {
	account, err := s.repo.FindByID(ctx, accountID)
	if err != nil {
		return err
	}

	// Criar método de pagamento fake
	fakePaymentMethod := billing.PaymentMethod{
		Type:       "fake_card",
		LastDigits: "1234",
		IsDefault:  true,
	}

	if err := account.ActivatePayment(fakePaymentMethod); err != nil {
		return fmt.Errorf("failed to activate payment: %w", err)
	}

	if err := s.repo.Update(ctx, account); err != nil {
		s.logger.Error("Failed to update billing account",
			zap.Error(err),
			zap.String("account_id", accountID.String()),
		)
		return fmt.Errorf("failed to update billing account: %w", err)
	}

	s.logger.Info("Payment activated (fake)",
		zap.String("account_id", accountID.String()),
	)

	return nil
}

// SuspendAccount suspende uma conta
func (s *BillingService) SuspendAccount(ctx context.Context, accountID uuid.UUID, reason string) error {
	account, err := s.repo.FindByID(ctx, accountID)
	if err != nil {
		return err
	}

	account.Suspend(reason)

	if err := s.repo.Update(ctx, account); err != nil {
		s.logger.Error("Failed to suspend billing account",
			zap.Error(err),
			zap.String("account_id", accountID.String()),
		)
		return fmt.Errorf("failed to suspend billing account: %w", err)
	}

	s.logger.Info("Billing account suspended",
		zap.String("account_id", accountID.String()),
		zap.String("reason", reason),
	)

	return nil
}

// ReactivateAccount reativa uma conta suspensa
func (s *BillingService) ReactivateAccount(ctx context.Context, accountID uuid.UUID) error {
	account, err := s.repo.FindByID(ctx, accountID)
	if err != nil {
		return err
	}

	if err := account.Reactivate(); err != nil {
		return fmt.Errorf("failed to reactivate account: %w", err)
	}

	if err := s.repo.Update(ctx, account); err != nil {
		s.logger.Error("Failed to reactivate billing account",
			zap.Error(err),
			zap.String("account_id", accountID.String()),
		)
		return fmt.Errorf("failed to reactivate billing account: %w", err)
	}

	s.logger.Info("Billing account reactivated",
		zap.String("account_id", accountID.String()),
	)

	return nil
}
