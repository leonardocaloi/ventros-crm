package contact

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ===========================
// 1.5.2 - Testes de Phone Value Object
// ===========================

func TestNewPhone_ValidPhone(t *testing.T) {
	validPhones := []string{
		"5511999999999",        // Brazilian mobile
		"+5511999999999",       // With country code
		"(11) 99999-9999",      // Formatted
		"11 9 9999-9999",       // With spaces
		"+1 (555) 123-4567",    // US format
		"44 20 1234 5678",      // UK format
	}

	for _, phoneStr := range validPhones {
		t.Run(phoneStr, func(t *testing.T) {
			// Act
			phone, err := NewPhone(phoneStr)

			// Assert
			require.NoError(t, err, "Phone %s should be valid", phoneStr)
			assert.NotEmpty(t, phone.String())
			// Should contain only digits and +
			assert.Regexp(t, `^[0-9+]+$`, phone.String())
		})
	}
}

func TestNewPhone_InvalidFormat(t *testing.T) {
	invalidPhones := []string{
		"123",      // Too short
		"12345",    // Too short
		"",         // Empty
		"   ",      // Spaces only
		"abc",      // Letters
		"!!!",      // Special chars only
	}

	for _, phoneStr := range invalidPhones {
		t.Run(phoneStr, func(t *testing.T) {
			// Act
			phone, err := NewPhone(phoneStr)

			// Assert
			require.Error(t, err, "Phone %s should be invalid", phoneStr)
			assert.Empty(t, phone.String())
		})
	}
}

func TestNewPhone_EmptyString(t *testing.T) {
	// Act
	phone, err := NewPhone("")

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "phone cannot be empty")
	assert.Empty(t, phone.String())
}

func TestNewPhone_TooShort(t *testing.T) {
	// Act
	phone, err := NewPhone("123")

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "phone too short")
	assert.Empty(t, phone.String())
}

func TestPhone_String(t *testing.T) {
	// Arrange
	phoneStr := "5511999999999"
	phone, _ := NewPhone(phoneStr)

	// Act
	result := phone.String()

	// Assert
	assert.Equal(t, phoneStr, result)
}

func TestPhone_Equals(t *testing.T) {
	// Arrange
	phone1, _ := NewPhone("5511999999999")
	phone2, _ := NewPhone("5511999999999")
	phone3, _ := NewPhone("5511888888888")

	// Assert
	assert.True(t, phone1.Equals(phone2), "Same phones should be equal")
	assert.False(t, phone1.Equals(phone3), "Different phones should not be equal")
}

func TestPhone_RemovesFormatting(t *testing.T) {
	// Arrange - formatted phone
	phone, _ := NewPhone("(11) 99999-9999")

	// Assert - should remove all non-numeric chars (except +)
	assert.Equal(t, "11999999999", phone.String())
	assert.NotContains(t, phone.String(), "(")
	assert.NotContains(t, phone.String(), ")")
	assert.NotContains(t, phone.String(), "-")
	assert.NotContains(t, phone.String(), " ")
}

func TestPhone_PreservesCountryCode(t *testing.T) {
	// Arrange - phone with + prefix
	phone, _ := NewPhone("+5511999999999")

	// Assert - should preserve the +
	assert.Equal(t, "+5511999999999", phone.String())
	assert.Contains(t, phone.String(), "+")
}

func TestPhone_TrimsSpaces(t *testing.T) {
	// Arrange
	phone, _ := NewPhone("  5511999999999  ")

	// Assert
	assert.Equal(t, "5511999999999", phone.String())
}

func TestPhone_MixedFormatting(t *testing.T) {
	// Arrange - various formatting styles should normalize to the same value
	phones := []string{
		"+55 11 99999-9999",
		"+55 (11) 99999-9999",
		"+5511999999999",
		"+55 11999999999",
	}

	expected := "+5511999999999"

	for _, phoneStr := range phones {
		t.Run(phoneStr, func(t *testing.T) {
			// Act
			phone, err := NewPhone(phoneStr)

			// Assert
			require.NoError(t, err)
			assert.Equal(t, expected, phone.String(), "All formats should normalize to %s", expected)
		})
	}
}
