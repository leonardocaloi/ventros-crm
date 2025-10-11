package websocket

import (
	"html"
	"regexp"
	"strings"
)

// SanitizeText remove caracteres perigosos para prevenir XSS
// SECURITY: Chamado em ambas direções (client → server e server → client)
func SanitizeText(text string) string {
	// Remove control characters e null bytes
	sanitized := removeControlChars(text)

	// Escape HTML entities
	sanitized = html.EscapeString(sanitized)

	// Remove scripts e tags perigosas (defense in depth)
	sanitized = removeScriptTags(sanitized)

	// Limita tamanho máximo
	const maxLength = 10000
	if len(sanitized) > maxLength {
		sanitized = sanitized[:maxLength]
	}

	return sanitized
}

// removeControlChars remove caracteres de controle perigosos
func removeControlChars(s string) string {
	var result strings.Builder
	for _, r := range s {
		// Permitir: printable chars, newlines, tabs
		if r >= 32 || r == '\n' || r == '\t' || r == '\r' {
			result.WriteRune(r)
		}
		// Remove: null bytes, control chars
	}
	return result.String()
}

// removeScriptTags remove tags <script> e similares (defense in depth)
func removeScriptTags(s string) string {
	// Remove <script>...</script>
	scriptRegex := regexp.MustCompile(`(?i)<script[^>]*>.*?</script>`)
	s = scriptRegex.ReplaceAllString(s, "")

	// Remove event handlers (onclick, onerror, etc)
	eventRegex := regexp.MustCompile(`(?i)\son\w+\s*=`)
	s = eventRegex.ReplaceAllString(s, "")

	// Remove javascript: protocol
	jsProtocolRegex := regexp.MustCompile(`(?i)javascript:`)
	s = jsProtocolRegex.ReplaceAllString(s, "")

	return s
}

// ValidateOrigin valida Origin header para prevenir CSWSH
// SECURITY: CRITICAL - previne Cross-Site WebSocket Hijacking
func ValidateOrigin(origin string, allowedOrigins []string) bool {
	if origin == "" {
		// Conexões sem Origin são suspeitas (possível ataque)
		return false
	}

	for _, allowed := range allowedOrigins {
		if origin == allowed {
			return true
		}
	}

	return false
}

// GetAllowedOrigins retorna lista de origens permitidas
func GetAllowedOrigins(production bool) []string {
	if production {
		return []string{
			"https://app.ventros.io",
			"https://ventros.io",
			"https://www.ventros.io",
		}
	}

	// Development origins
	return []string{
		"http://localhost:3000",
		"http://localhost:3001",
		"http://localhost:5173", // Vite
		"http://localhost:8080",
		"http://127.0.0.1:3000",
		"http://127.0.0.1:5173",
		"https://app.ventros.io",
		"https://ventros.io",
	}
}
