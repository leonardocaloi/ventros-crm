package customer

import (
	"testing"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCustomer(t *testing.T) {
	contactID := uuid.New()

	t.Run("valid customer", func(t *testing.T) {
		customer, err := NewCustomer(contactID, "tenant-123", "customer@example.com")
		require.NoError(t, err)
		assert.NotNil(t, customer)
		assert.Equal(t, contactID, customer.ContactID())
	})

	t.Run("validation", func(t *testing.T) {
		_, err := NewCustomer(uuid.Nil, "tenant", "email@test.com")
		assert.Error(t, err)

		_, err = NewCustomer(contactID, "", "email@test.com")
		assert.Error(t, err)
	})
}
