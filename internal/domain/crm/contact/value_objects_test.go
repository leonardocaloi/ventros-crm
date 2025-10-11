package contact

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Email Value Object Tests

func TestNewEmail_ValidEmail(t *testing.T) {
	validEmails := []string{
		"test@example.com",
		"user.name@example.com",
		"user+tag@example.co.uk",
		"test123@test-domain.com",
		"UPPERCASE@EXAMPLE.COM",
	}

	for _, emailStr := range validEmails {
		t.Run(emailStr, func(t *testing.T) {
			email, err := NewEmail(emailStr)
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
		"user @example.com",
		"user@.com",
		"user@domain",
		"",
		"   ",
		"user@@domain.com",
	}

	for _, emailStr := range invalidEmails {
		t.Run(emailStr, func(t *testing.T) {
			email, err := NewEmail(emailStr)
			require.Error(t, err, "Email %s should be invalid", emailStr)
			assert.Empty(t, email.String())
		})
	}
}

func TestNewEmail_EmptyString(t *testing.T) {
	email, err := NewEmail("")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "email cannot be empty")
	assert.Empty(t, email.String())
}

func TestEmail_String(t *testing.T) {
	emailStr := "test@example.com"
	email, _ := NewEmail(emailStr)
	result := email.String()
	assert.Equal(t, emailStr, result)
}

func TestEmail_Equals(t *testing.T) {
	email1, _ := NewEmail("test@example.com")
	email2, _ := NewEmail("test@example.com")
	email3, _ := NewEmail("other@example.com")

	assert.True(t, email1.Equals(email2), "Same emails should be equal")
	assert.False(t, email1.Equals(email3), "Different emails should not be equal")
}

func TestEmail_Normalization(t *testing.T) {
	email1, _ := NewEmail("Test@Example.COM")
	email2, _ := NewEmail("test@example.com")

	assert.Equal(t, "test@example.com", email1.String())
	assert.True(t, email1.Equals(email2), "Emails should be case-insensitive")
}

func TestEmail_TrimsSpaces(t *testing.T) {
	email, _ := NewEmail("  test@example.com  ")
	assert.Equal(t, "test@example.com", email.String())
}

// Phone Value Object Tests

func TestNewPhone_ValidPhone(t *testing.T) {
	validPhones := []string{
		"5511999999999",
		"+5511999999999",
		"(11) 99999-9999",
		"11 9 9999-9999",
		"+1 (555) 123-4567",
		"44 20 1234 5678",
	}

	for _, phoneStr := range validPhones {
		t.Run(phoneStr, func(t *testing.T) {
			phone, err := NewPhone(phoneStr)
			require.NoError(t, err, "Phone %s should be valid", phoneStr)
			assert.NotEmpty(t, phone.String())
			assert.Regexp(t, `^[0-9+]+$`, phone.String())
		})
	}
}

func TestNewPhone_InvalidFormat(t *testing.T) {
	invalidPhones := []string{
		"123",
		"12345",
		"",
		"   ",
		"abc",
		"!!!",
	}

	for _, phoneStr := range invalidPhones {
		t.Run(phoneStr, func(t *testing.T) {
			phone, err := NewPhone(phoneStr)
			require.Error(t, err, "Phone %s should be invalid", phoneStr)
			assert.Empty(t, phone.String())
		})
	}
}

func TestNewPhone_EmptyString(t *testing.T) {
	phone, err := NewPhone("")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "phone cannot be empty")
	assert.Empty(t, phone.String())
}

func TestNewPhone_TooShort(t *testing.T) {
	phone, err := NewPhone("123")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "phone too short")
	assert.Empty(t, phone.String())
}

func TestPhone_String(t *testing.T) {
	phoneStr := "5511999999999"
	phone, _ := NewPhone(phoneStr)
	result := phone.String()
	assert.Equal(t, phoneStr, result)
}

func TestPhone_Equals(t *testing.T) {
	phone1, _ := NewPhone("5511999999999")
	phone2, _ := NewPhone("5511999999999")
	phone3, _ := NewPhone("5511888888888")

	assert.True(t, phone1.Equals(phone2), "Same phones should be equal")
	assert.False(t, phone1.Equals(phone3), "Different phones should not be equal")
}

func TestPhone_RemovesFormatting(t *testing.T) {
	phone, _ := NewPhone("(11) 99999-9999")
	assert.Equal(t, "11999999999", phone.String())
	assert.NotContains(t, phone.String(), "(")
	assert.NotContains(t, phone.String(), ")")
	assert.NotContains(t, phone.String(), "-")
	assert.NotContains(t, phone.String(), " ")
}

func TestPhone_PreservesCountryCode(t *testing.T) {
	phone, _ := NewPhone("+5511999999999")
	assert.Equal(t, "+5511999999999", phone.String())
	assert.Contains(t, phone.String(), "+")
}

func TestPhone_TrimsSpaces(t *testing.T) {
	phone, _ := NewPhone("  5511999999999  ")
	assert.Equal(t, "5511999999999", phone.String())
}

func TestPhone_MixedFormatting(t *testing.T) {
	phones := []string{
		"+55 11 99999-9999",
		"+55 (11) 99999-9999",
		"+5511999999999",
		"+55 11999999999",
	}

	expected := "+5511999999999"

	for _, phoneStr := range phones {
		t.Run(phoneStr, func(t *testing.T) {
			phone, err := NewPhone(phoneStr)
			require.NoError(t, err)
			assert.Equal(t, expected, phone.String(), "All formats should normalize to %s", expected)
		})
	}
}
