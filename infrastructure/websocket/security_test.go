package websocket

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSanitizeText(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "normal text",
			input:    "Hello World",
			expected: "Hello World",
		},
		{
			name:     "text with HTML",
			input:    "<b>Bold</b> text",
			expected: "&lt;b&gt;Bold&lt;/b&gt; text",
		},
		{
			name:     "text with script tag",
			input:    "Hello <script>alert('xss')</script>",
			expected: "Hello &lt;script&gt;alert(&#39;xss&#39;)&lt;/script&gt;", // HTML escaped = safe
		},
		{
			name:     "text with null bytes",
			input:    "Hello\x00World",
			expected: "HelloWorld",
		},
		{
			name:     "text with control chars",
			input:    "Hello\x01\x02World",
			expected: "HelloWorld",
		},
		{
			name:     "text with event handlers",
			input:    "<div onclick='alert()'>Click</div>",
			expected: "&lt;div&#39;alert()&#39;&gt;Click&lt;/div&gt;", // HTML escaped + onclick removed
		},
		{
			name:     "text with javascript protocol",
			input:    "<a href='javascript:alert()'>Link</a>",
			expected: "&lt;a href=&#39;alert()&#39;&gt;Link&lt;/a&gt;",
		},
		{
			name:     "very long text",
			input:    strings.Repeat("A", 11000),
			expected: strings.Repeat("A", 10000),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeText(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestValidateOrigin(t *testing.T) {
	allowedOrigins := []string{
		"http://localhost:3000",
		"https://app.ventros.io",
	}

	tests := []struct {
		name     string
		origin   string
		expected bool
	}{
		{
			name:     "allowed localhost",
			origin:   "http://localhost:3000",
			expected: true,
		},
		{
			name:     "allowed production",
			origin:   "https://app.ventros.io",
			expected: true,
		},
		{
			name:     "disallowed origin",
			origin:   "http://evil.com",
			expected: false,
		},
		{
			name:     "empty origin",
			origin:   "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateOrigin(tt.origin, allowedOrigins)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetAllowedOrigins(t *testing.T) {
	t.Run("production", func(t *testing.T) {
		origins := GetAllowedOrigins(true)
		assert.Contains(t, origins, "https://app.ventros.io")
		assert.NotContains(t, origins, "http://localhost:3000")
	})

	t.Run("development", func(t *testing.T) {
		origins := GetAllowedOrigins(false)
		assert.Contains(t, origins, "http://localhost:3000")
		assert.Contains(t, origins, "https://app.ventros.io")
	})
}
