package contact_list

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ventros/crm/internal/domain/core/shared"
)

func TestParseFieldType_ValidTypes(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    shared.FieldType
		description string
	}{
		{
			name:        "text field type",
			input:       "text",
			expected:    shared.FieldTypeText,
			description: "Should parse text field type correctly",
		},
		{
			name:        "number field type",
			input:       "number",
			expected:    shared.FieldTypeNumber,
			description: "Should parse number field type correctly",
		},
		{
			name:        "boolean field type",
			input:       "boolean",
			expected:    shared.FieldTypeBoolean,
			description: "Should parse boolean field type correctly",
		},
		{
			name:        "date field type",
			input:       "date",
			expected:    shared.FieldTypeDate,
			description: "Should parse date field type correctly",
		},
		{
			name:        "json field type",
			input:       "json",
			expected:    shared.FieldTypeJSON,
			description: "Should parse json field type correctly",
		},
		{
			name:        "url field type",
			input:       "url",
			expected:    shared.FieldTypeURL,
			description: "Should parse url field type correctly",
		},
		{
			name:        "email field type",
			input:       "email",
			expected:    shared.FieldTypeEmail,
			description: "Should parse email field type correctly",
		},
		{
			name:        "phone field type",
			input:       "phone",
			expected:    shared.FieldTypePhone,
			description: "Should parse phone field type correctly",
		},
		{
			name:        "select field type",
			input:       "select",
			expected:    shared.FieldTypeSelect,
			description: "Should parse select field type correctly",
		},
		{
			name:        "multi_select field type",
			input:       "multi_select",
			expected:    shared.FieldTypeMultiSelect,
			description: "Should parse multi_select field type correctly",
		},
		{
			name:        "label field type",
			input:       "label",
			expected:    shared.FieldTypeLabel,
			description: "Should parse label field type correctly",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			result := parseFieldType(tt.input)

			// Assert
			assert.Equal(t, tt.expected, result, tt.description)
			assert.True(t, result.IsValid(), "Result should be a valid field type")
		})
	}
}

func TestParseFieldType_InvalidType(t *testing.T) {
	// Arrange
	invalidType := "invalid_type"

	// Act
	result := parseFieldType(invalidType)

	// Assert
	assert.Equal(t, shared.FieldType("invalid_type"), result)
	assert.False(t, result.IsValid(), "Invalid type should not be valid")
}

func TestParseFieldType_EmptyString(t *testing.T) {
	// Arrange
	emptyType := ""

	// Act
	result := parseFieldType(emptyType)

	// Assert
	assert.Equal(t, shared.FieldType(""), result)
	assert.False(t, result.IsValid(), "Empty type should not be valid")
}

func TestParseFieldType_CaseSensitivity(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected shared.FieldType
	}{
		{
			name:     "uppercase TEXT",
			input:    "TEXT",
			expected: shared.FieldType("TEXT"),
		},
		{
			name:     "mixed case Text",
			input:    "Text",
			expected: shared.FieldType("Text"),
		},
		{
			name:     "mixed case NuMbEr",
			input:    "NuMbEr",
			expected: shared.FieldType("NuMbEr"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			result := parseFieldType(tt.input)

			// Assert
			assert.Equal(t, tt.expected, result)
			assert.False(t, result.IsValid(), "Case-sensitive strings should not be valid")
		})
	}
}

func TestParseFieldType_WithWhitespace(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "leading whitespace",
			input: " text",
		},
		{
			name:  "trailing whitespace",
			input: "text ",
		},
		{
			name:  "both whitespace",
			input: " text ",
		},
		{
			name:  "whitespace in middle",
			input: "te xt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			result := parseFieldType(tt.input)

			// Assert
			assert.Equal(t, shared.FieldType(tt.input), result)
			assert.False(t, result.IsValid(), "Strings with whitespace should not be valid")
		})
	}
}

func TestParseFieldType_SpecialCharacters(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "with underscore prefix",
			input: "_text",
		},
		{
			name:  "with hyphen",
			input: "text-field",
		},
		{
			name:  "with dot",
			input: "text.field",
		},
		{
			name:  "with special chars",
			input: "text@field",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			result := parseFieldType(tt.input)

			// Assert
			assert.Equal(t, shared.FieldType(tt.input), result)
			assert.False(t, result.IsValid(), "Strings with special characters should not be valid")
		})
	}
}

func TestParseFieldType_UnicodeCharacters(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "unicode emoji",
			input: "text😀",
		},
		{
			name:  "unicode chinese",
			input: "文本",
		},
		{
			name:  "unicode arabic",
			input: "نص",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			result := parseFieldType(tt.input)

			// Assert
			assert.Equal(t, shared.FieldType(tt.input), result)
			assert.False(t, result.IsValid(), "Unicode strings should not be valid")
		})
	}
}

func TestParseFieldType_NumericStrings(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "numeric string",
			input: "123",
		},
		{
			name:  "numeric with text",
			input: "text123",
		},
		{
			name:  "text with numeric",
			input: "123text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			result := parseFieldType(tt.input)

			// Assert
			assert.Equal(t, shared.FieldType(tt.input), result)
			assert.False(t, result.IsValid(), "Numeric strings should not be valid")
		})
	}
}

func TestParseFieldType_ReturnType(t *testing.T) {
	// Arrange
	input := "text"

	// Act
	result := parseFieldType(input)

	// Assert
	assert.IsType(t, shared.FieldType(""), result, "Should return shared.FieldType type")
}

func TestParseFieldType_Consistency(t *testing.T) {
	// Arrange
	input := "text"

	// Act
	result1 := parseFieldType(input)
	result2 := parseFieldType(input)

	// Assert
	assert.Equal(t, result1, result2, "Multiple calls with same input should return same result")
}

func TestParseFieldType_AllValidTypesAreValid(t *testing.T) {
	// Test that all parsed valid types are actually valid according to IsValid()
	validTypes := []string{
		"text", "number", "boolean", "date", "json",
		"url", "email", "phone", "select", "multi_select", "label",
	}

	for _, typeStr := range validTypes {
		t.Run("validate_"+typeStr, func(t *testing.T) {
			// Act
			result := parseFieldType(typeStr)

			// Assert
			assert.True(t, result.IsValid(), "Parsed type %s should be valid", typeStr)
		})
	}
}

func TestParseFieldType_ConversionToString(t *testing.T) {
	// Arrange
	input := "text"

	// Act
	result := parseFieldType(input)

	// Assert
	assert.Equal(t, input, string(result), "Should be convertible back to original string")
	assert.Equal(t, input, result.String(), "String() method should return original string")
}
