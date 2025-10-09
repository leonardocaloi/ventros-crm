package contact

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ===========================
// 1.5.1 - Testes de Email Value Object
// ===========================

func TestNewEmail_ValidEmail(t *testing.T) {
	validEmails := []string{
		"test@example.com",
		"user.name@example.com",
		"user+tag@example.co.uk",
		"test123@test-domain.com",
		"UPPERCASE@EXAMPLE.COM", // Should be normalized to lowercase
	}

	for _, emailStr := range validEmails {
		t.Run(emailStr, func(t *testing.T) {
			// Act
			email, err := NewEmail(emailStr)

			// Assert
			require.NoError(t, err, "Email %s should be valid", emailStr)
			assert.NotEmpty(t, email.String())
		})
	}
}

func TestNewEmail_InvalidFormat(t *testing.T) {
	invalidEmails := []string{
		"not-an-email",
		"@example.com",
		"user@",
		"user @example.com", // space
		"user@.com",
		"user@domain",     // missing TLD
		"",                // empty
		"   ",             // spaces only
		"user@@domain.com", // double @
	}

	for _, emailStr := range invalidEmails {
		t.Run(emailStr, func(t *testing.T) {
			// Act
			email, err := NewEmail(emailStr)

			// Assert
			require.Error(t, err, "Email %s should be invalid", emailStr)
			assert.Empty(t, email.String())
		})
	}
}

func TestNewEmail_EmptyString(t *testing.T) {
	// Act
	email, err := NewEmail("")

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "email cannot be empty")
	assert.Empty(t, email.String())
}

func TestEmail_String(t *testing.T) {
	// Arrange
	emailStr := "test@example.com"
	email, _ := NewEmail(emailStr)

	// Act
	result := email.String()

	// Assert
	assert.Equal(t, emailStr, result)
}

func TestEmail_Equals(t *testing.T) {
	// Arrange
	email1, _ := NewEmail("test@example.com")
	email2, _ := NewEmail("test@example.com")
	email3, _ := NewEmail("other@example.com")

	// Assert
	assert.True(t, email1.Equals(email2), "Same emails should be equal")
	assert.False(t, email1.Equals(email3), "Different emails should not be equal")
}

func TestEmail_Normalization(t *testing.T) {
	// Arrange - same email with different casing
	email1, _ := NewEmail("Test@Example.COM")
	email2, _ := NewEmail("test@example.com")

	// Assert - should be normalized to lowercase
	assert.Equal(t, "test@example.com", email1.String())
	assert.True(t, email1.Equals(email2), "Emails should be case-insensitive")
}

func TestEmail_TrimsSpaces(t *testing.T) {
	// Arrange
	email, _ := NewEmail("  test@example.com  ")

	// Assert
	assert.Equal(t, "test@example.com", email.String())
}
