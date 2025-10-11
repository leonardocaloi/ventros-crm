package customer

import (
	"testing"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCustomer(t *testing.T) {
	t.Run("valid customer", func(t *testing.T) {
		customer, err := NewCustomer("Acme Corp", "customer@example.com")
		require.NoError(t, err)
		assert.NotNil(t, customer)
		assert.NotEqual(t, uuid.Nil, customer.ID())
		assert.Equal(t, "Acme Corp", customer.Name())
		assert.Equal(t, "customer@example.com", customer.Email())
		assert.True(t, customer.IsActive())
	})

	t.Run("validation - empty name", func(t *testing.T) {
		_, err := NewCustomer("", "email@test.com")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "name cannot be empty")
	})

	t.Run("validation - empty email", func(t *testing.T) {
		_, err := NewCustomer("Test Company", "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "email cannot be empty")
	})
}
