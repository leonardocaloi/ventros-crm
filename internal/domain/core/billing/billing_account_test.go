package billing

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBillingAccount(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name         string
		userID       uuid.UUID
		accountName  string
		billingEmail string
		wantErr      error
	}{
		{
			name:         "valid billing account",
			userID:       userID,
			accountName:  "Test Account",
			billingEmail: "billing@example.com",
			wantErr:      nil,
		},
		{
			name:         "nil userID",
			userID:       uuid.Nil,
			accountName:  "Test Account",
			billingEmail: "billing@example.com",
			wantErr:      ErrInvalidUserID,
		},
		{
			name:         "empty name",
			userID:       userID,
			accountName:  "",
			billingEmail: "billing@example.com",
			wantErr:      ErrInvalidName,
		},
		{
			name:         "empty billing email",
			userID:       userID,
			accountName:  "Test Account",
			billingEmail: "",
			wantErr:      ErrInvalidEmail,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			account, err := NewBillingAccount(tt.userID, tt.accountName, tt.billingEmail)

			if tt.wantErr != nil {
				require.Error(t, err)
				assert.Equal(t, tt.wantErr, err)
				assert.Nil(t, account)
			} else {
				require.NoError(t, err)
				require.NotNil(t, account)

				assert.NotEqual(t, uuid.Nil, account.ID())
				assert.Equal(t, tt.userID, account.UserID())
				assert.Equal(t, tt.accountName, account.Name())
				assert.Equal(t, tt.billingEmail, account.BillingEmail())
				assert.Equal(t, PaymentStatusPending, account.PaymentStatus())
				assert.False(t, account.IsSuspended())
				assert.NotZero(t, account.CreatedAt())
				assert.NotZero(t, account.UpdatedAt())

				// Check domain event
				events := account.DomainEvents()
				require.Len(t, events, 1)
				event, ok := events[0].(BillingAccountCreatedEvent)
				require.True(t, ok)
				assert.Equal(t, account.ID(), event.AccountID)
			}
		})
	}
}

func TestBillingAccount_ActivatePayment(t *testing.T) {
	userID := uuid.New()
	account, err := NewBillingAccount(userID, "Test Account", "billing@example.com")
	require.NoError(t, err)
	account.ClearEvents()

	expiresAt := time.Now().AddDate(1, 0, 0)
	method := PaymentMethod{
		Type:       "credit_card",
		LastDigits: "1234",
		ExpiresAt:  &expiresAt,
		IsDefault:  true,
	}

	t.Run("activate payment method", func(t *testing.T) {
		err := account.ActivatePayment(method)
		require.NoError(t, err)

		assert.Equal(t, PaymentStatusActive, account.PaymentStatus())
		assert.Len(t, account.PaymentMethods(), 1)
		assert.True(t, account.PaymentMethods()[0].IsDefault)

		events := account.DomainEvents()
		require.Len(t, events, 1)
		_, ok := events[0].(PaymentMethodActivatedEvent)
		require.True(t, ok)
	})

	t.Run("add second payment method", func(t *testing.T) {
		account.ClearEvents()
		secondMethod := PaymentMethod{
			Type:       "pix",
			LastDigits: "",
			IsDefault:  true,
		}

		err := account.ActivatePayment(secondMethod)
		require.NoError(t, err)

		methods := account.PaymentMethods()
		assert.Len(t, methods, 2)

		// First method should no longer be default
		assert.False(t, methods[0].IsDefault)
		// Second method should be default
		assert.True(t, methods[1].IsDefault)
	})
}

func TestBillingAccount_Suspend(t *testing.T) {
	userID := uuid.New()
	account, err := NewBillingAccount(userID, "Test Account", "billing@example.com")
	require.NoError(t, err)

	// Activate payment first
	method := PaymentMethod{Type: "credit_card", LastDigits: "1234"}
	_ = account.ActivatePayment(method)
	account.ClearEvents()

	t.Run("suspend active account", func(t *testing.T) {
		assert.False(t, account.IsSuspended())

		account.Suspend("Payment failed")

		assert.True(t, account.IsSuspended())
		assert.NotNil(t, account.SuspendedAt())
		assert.Equal(t, "Payment failed", account.SuspensionReason())
		assert.Equal(t, PaymentStatusSuspended, account.PaymentStatus())

		events := account.DomainEvents()
		require.Len(t, events, 1)
		event, ok := events[0].(BillingAccountSuspendedEvent)
		require.True(t, ok)
		assert.Equal(t, "Payment failed", event.Reason)
	})

	t.Run("suspend already suspended account", func(t *testing.T) {
		account.ClearEvents()
		account.Suspend("Another reason")

		// Should not generate duplicate event
		events := account.DomainEvents()
		assert.Len(t, events, 0)
	})
}

func TestBillingAccount_Reactivate(t *testing.T) {
	userID := uuid.New()
	account, err := NewBillingAccount(userID, "Test Account", "billing@example.com")
	require.NoError(t, err)

	// Setup: activate payment and then suspend
	method := PaymentMethod{Type: "credit_card", LastDigits: "1234"}
	_ = account.ActivatePayment(method)
	account.Suspend("Payment failed")
	account.ClearEvents()

	t.Run("reactivate suspended account", func(t *testing.T) {
		err := account.Reactivate()
		require.NoError(t, err)

		assert.False(t, account.IsSuspended())
		assert.Nil(t, account.SuspendedAt())
		assert.Empty(t, account.SuspensionReason())
		assert.Equal(t, PaymentStatusActive, account.PaymentStatus())

		events := account.DomainEvents()
		require.Len(t, events, 1)
		_, ok := events[0].(BillingAccountReactivatedEvent)
		require.True(t, ok)
	})

	t.Run("reactivate already active account", func(t *testing.T) {
		account.ClearEvents()
		err := account.Reactivate()
		require.NoError(t, err)

		// Should not generate event
		events := account.DomainEvents()
		assert.Len(t, events, 0)
	})

	t.Run("cannot reactivate without payment method", func(t *testing.T) {
		// Create new account without payment method
		newAccount, err := NewBillingAccount(userID, "Test", "test@example.com")
		require.NoError(t, err)
		newAccount.Suspend("Test")

		err = newAccount.Reactivate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot reactivate without payment method")
	})
}

func TestBillingAccount_Cancel(t *testing.T) {
	userID := uuid.New()
	account, err := NewBillingAccount(userID, "Test Account", "billing@example.com")
	require.NoError(t, err)

	method := PaymentMethod{Type: "credit_card", LastDigits: "1234"}
	_ = account.ActivatePayment(method)
	account.ClearEvents()

	t.Run("cancel account", func(t *testing.T) {
		account.Cancel()

		assert.Equal(t, PaymentStatusCanceled, account.PaymentStatus())
		assert.True(t, account.IsSuspended())
		assert.NotNil(t, account.SuspendedAt())
		assert.Equal(t, "Canceled by user", account.SuspensionReason())

		events := account.DomainEvents()
		require.Len(t, events, 1)
		_, ok := events[0].(BillingAccountCanceledEvent)
		require.True(t, ok)
	})

	t.Run("cannot activate payment on canceled account", func(t *testing.T) {
		newMethod := PaymentMethod{Type: "pix"}
		err := account.ActivatePayment(newMethod)

		require.Error(t, err)
		assert.Equal(t, ErrAccountCanceled, err)
	})
}

func TestBillingAccount_CanCreateProject(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name          string
		setupFn       func(*BillingAccount)
		canCreate     bool
	}{
		{
			name: "pending account cannot create project",
			setupFn: func(a *BillingAccount) {
				// Do nothing, account is pending by default
			},
			canCreate: false,
		},
		{
			name: "active account can create project",
			setupFn: func(a *BillingAccount) {
				method := PaymentMethod{Type: "credit_card"}
				_ = a.ActivatePayment(method)
			},
			canCreate: true,
		},
		{
			name: "suspended account cannot create project",
			setupFn: func(a *BillingAccount) {
				method := PaymentMethod{Type: "credit_card"}
				_ = a.ActivatePayment(method)
				a.Suspend("Payment failed")
			},
			canCreate: false,
		},
		{
			name: "canceled account cannot create project",
			setupFn: func(a *BillingAccount) {
				method := PaymentMethod{Type: "credit_card"}
				_ = a.ActivatePayment(method)
				a.Cancel()
			},
			canCreate: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			account, err := NewBillingAccount(userID, "Test Account", "billing@example.com")
			require.NoError(t, err)

			tt.setupFn(account)

			assert.Equal(t, tt.canCreate, account.CanCreateProject())
			assert.Equal(t, tt.canCreate, account.IsActive())
		})
	}
}

func TestBillingAccount_UpdateBillingEmail(t *testing.T) {
	userID := uuid.New()
	account, err := NewBillingAccount(userID, "Test Account", "billing@example.com")
	require.NoError(t, err)

	t.Run("update billing email", func(t *testing.T) {
		err := account.UpdateBillingEmail("new-billing@example.com")
		require.NoError(t, err)
		assert.Equal(t, "new-billing@example.com", account.BillingEmail())
	})

	t.Run("empty billing email", func(t *testing.T) {
		err := account.UpdateBillingEmail("")
		require.Error(t, err)
		assert.Equal(t, ErrInvalidEmail, err)
	})
}

func TestReconstructBillingAccount(t *testing.T) {
	id := uuid.New()
	userID := uuid.New()
	createdAt := time.Now().Add(-24 * time.Hour)
	updatedAt := time.Now()
	suspendedAt := time.Now().Add(-1 * time.Hour)

	expiresAt := time.Now().AddDate(1, 0, 0)
	methods := []PaymentMethod{
		{
			Type:       "credit_card",
			LastDigits: "1234",
			ExpiresAt:  &expiresAt,
			IsDefault:  true,
		},
	}

	t.Run("reconstruct active account", func(t *testing.T) {
		account := ReconstructBillingAccount(
			id, userID,
			"Reconstructed Account",
			PaymentStatusActive,
			methods,
			"billing@example.com",
			false,
			nil,
			"",
			createdAt,
			updatedAt,
		)

		assert.Equal(t, id, account.ID())
		assert.Equal(t, userID, account.UserID())
		assert.Equal(t, "Reconstructed Account", account.Name())
		assert.Equal(t, PaymentStatusActive, account.PaymentStatus())
		assert.Len(t, account.PaymentMethods(), 1)
		assert.False(t, account.IsSuspended())
		assert.Len(t, account.DomainEvents(), 0) // No events on reconstruction
	})

	t.Run("reconstruct suspended account", func(t *testing.T) {
		account := ReconstructBillingAccount(
			id, userID,
			"Suspended Account",
			PaymentStatusSuspended,
			methods,
			"billing@example.com",
			true,
			&suspendedAt,
			"Payment failed",
			createdAt,
			updatedAt,
		)

		assert.True(t, account.IsSuspended())
		assert.Equal(t, &suspendedAt, account.SuspendedAt())
		assert.Equal(t, "Payment failed", account.SuspensionReason())
	})

	t.Run("reconstruct with nil payment methods", func(t *testing.T) {
		account := ReconstructBillingAccount(
			id, userID,
			"Account",
			PaymentStatusPending,
			nil, // nil methods
			"billing@example.com",
			false,
			nil,
			"",
			createdAt,
			updatedAt,
		)

		// Should initialize empty slice
		assert.NotNil(t, account.PaymentMethods())
		assert.Len(t, account.PaymentMethods(), 0)
	})
}

func TestBillingAccount_ActivatePaymentOnSuspendedAccount(t *testing.T) {
	userID := uuid.New()
	account, err := NewBillingAccount(userID, "Test Account", "billing@example.com")
	require.NoError(t, err)

	// Suspend account
	account.Suspend("Test suspension")

	t.Run("cannot activate payment on suspended account", func(t *testing.T) {
		method := PaymentMethod{Type: "credit_card"}
		err := account.ActivatePayment(method)

		require.Error(t, err)
		assert.Equal(t, ErrAccountSuspended, err)
	})
}

func TestBillingAccount_EventManagement(t *testing.T) {
	userID := uuid.New()
	account, err := NewBillingAccount(userID, "Test Account", "billing@example.com")
	require.NoError(t, err)

	t.Run("clear events", func(t *testing.T) {
		assert.Len(t, account.DomainEvents(), 1) // Creation event

		account.ClearEvents()
		assert.Len(t, account.DomainEvents(), 0)
	})

	t.Run("multiple operations generate events", func(t *testing.T) {
		account.ClearEvents()

		method := PaymentMethod{Type: "credit_card"}
		_ = account.ActivatePayment(method)
		account.Suspend("Test")
		_ = account.Reactivate()

		events := account.DomainEvents()
		assert.Len(t, events, 3)
	})

	t.Run("events are immutable copies", func(t *testing.T) {
		account.ClearEvents()
		account.Cancel()

		events1 := account.DomainEvents()
		events2 := account.DomainEvents()

		// Should be different slices (copies)
		assert.NotSame(t, &events1, &events2)
		assert.Equal(t, len(events1), len(events2))
	})
}
