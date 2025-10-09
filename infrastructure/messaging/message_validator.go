package messaging

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"

	messageports "github.com/caloi/ventros-crm/internal/application/message"
)

// MessageValidator implementa validação de mensagens
// Seguindo Single Responsibility Principle (SRP)
type DefaultMessageValidator struct {
	maxContentLength int
	maxMediaSize     int64
	allowedDomains   []string
	blockedWords     []string
}

// NewMessageValidator cria uma nova instância do validador
func NewMessageValidator() messageports.MessageValidator {
	return &DefaultMessageValidator{
		maxContentLength: 4096,
		maxMediaSize:     64 * 1024 * 1024, // 64MB
		allowedDomains:   []string{},       // Vazio = todos permitidos
		blockedWords:     []string{},       // Lista de palavras bloqueadas
	}
}

// ValidateContent valida o conteúdo da mensagem
func (v *DefaultMessageValidator) ValidateContent(messageType messageports.MessageType, content string) error {
	// Validação básica de tamanho
	if len(content) > v.maxContentLength {
		return fmt.Errorf("content exceeds maximum length of %d characters", v.maxContentLength)
	}

	// Validações específicas por tipo
	switch messageType {
	case messageports.MessageTypeText:
		return v.validateTextContent(content)
	case messageports.MessageTypeTemplate:
		return v.validateTemplateContent(content)
	default:
		// Para outros tipos, validação básica
		if strings.TrimSpace(content) == "" {
			return fmt.Errorf("content cannot be empty for message type %s", messageType)
		}
	}

	return nil
}

// ValidateMedia valida URLs de mídia
func (v *DefaultMessageValidator) ValidateMedia(mediaURL string, mediaType string) error {
	// Validar URL
	parsedURL, err := url.Parse(mediaURL)
	if err != nil {
		return fmt.Errorf("invalid media URL: %w", err)
	}

	// Verificar esquema
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return fmt.Errorf("media URL must use http or https scheme")
	}

	// Verificar domínios permitidos (se configurado)
	if len(v.allowedDomains) > 0 {
		allowed := false
		for _, domain := range v.allowedDomains {
			if strings.Contains(parsedURL.Host, domain) {
				allowed = true
				break
			}
		}
		if !allowed {
			return fmt.Errorf("media domain %s is not allowed", parsedURL.Host)
		}
	}

	// Validar tipo de mídia
	return v.validateMediaType(mediaType)
}

// ValidateTemplate valida dados de template
func (v *DefaultMessageValidator) ValidateTemplate(templateData map[string]interface{}) error {
	if templateData == nil {
		return fmt.Errorf("template data cannot be nil")
	}

	// Verificar se tem pelo menos um campo
	if len(templateData) == 0 {
		return fmt.Errorf("template data cannot be empty")
	}

	// Validar tipos de dados suportados
	for key, value := range templateData {
		if err := v.validateTemplateValue(key, value); err != nil {
			return fmt.Errorf("invalid template value for key %s: %w", key, err)
		}
	}

	return nil
}

// validateTextContent valida conteúdo de texto
func (v *DefaultMessageValidator) validateTextContent(content string) error {
	// Verificar se não está vazio
	if strings.TrimSpace(content) == "" {
		return fmt.Errorf("text content cannot be empty")
	}

	// Verificar palavras bloqueadas
	for _, blockedWord := range v.blockedWords {
		if strings.Contains(strings.ToLower(content), strings.ToLower(blockedWord)) {
			return fmt.Errorf("content contains blocked word: %s", blockedWord)
		}
	}

	// Verificar caracteres especiais excessivos
	if v.hasExcessiveSpecialChars(content) {
		return fmt.Errorf("content contains excessive special characters")
	}

	return nil
}

// validateTemplateContent valida conteúdo de template
func (v *DefaultMessageValidator) validateTemplateContent(content string) error {
	// Verificar se tem placeholders válidos
	placeholderRegex := regexp.MustCompile(`\{\{[a-zA-Z_][a-zA-Z0-9_]*\}\}`)
	placeholders := placeholderRegex.FindAllString(content, -1)

	if len(placeholders) == 0 {
		return fmt.Errorf("template content must contain at least one placeholder")
	}

	// Verificar sintaxe dos placeholders
	for _, placeholder := range placeholders {
		if !v.isValidPlaceholder(placeholder) {
			return fmt.Errorf("invalid placeholder syntax: %s", placeholder)
		}
	}

	return nil
}

// validateMediaType valida tipos de mídia
func (v *DefaultMessageValidator) validateMediaType(mediaType string) error {
	allowedTypes := map[string][]string{
		"image": {"image/jpeg", "image/png", "image/gif", "image/webp"},
		"audio": {"audio/mpeg", "audio/ogg", "audio/wav", "audio/mp4", "audio/aac"},
		"video": {"video/mp4", "video/webm", "video/quicktime", "video/avi"},
		"document": {
			"application/pdf",
			"application/msword",
			"application/vnd.openxmlformats-officedocument.wordprocessingml.document",
			"application/vnd.ms-excel",
			"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
			"text/plain",
		},
	}

	// Verificar se o tipo é suportado
	for _, types := range allowedTypes {
		for _, allowedType := range types {
			if mediaType == allowedType {
				return nil
			}
		}
	}

	return fmt.Errorf("unsupported media type: %s", mediaType)
}

// validateTemplateValue valida valores de template
func (v *DefaultMessageValidator) validateTemplateValue(key string, value interface{}) error {
	// Verificar nome da chave
	if !v.isValidTemplateKey(key) {
		return fmt.Errorf("invalid template key format")
	}

	// Verificar tipos suportados
	switch v := value.(type) {
	case string:
		if len(v) > 1000 {
			return fmt.Errorf("string value too long")
		}
	case int, int32, int64, float32, float64:
		// Números são válidos
	case bool:
		// Booleanos são válidos
	case nil:
		return fmt.Errorf("template value cannot be nil")
	default:
		return fmt.Errorf("unsupported template value type: %T", value)
	}

	return nil
}

// hasExcessiveSpecialChars verifica caracteres especiais excessivos
func (v *DefaultMessageValidator) hasExcessiveSpecialChars(content string) bool {
	specialChars := regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?~` + "`" + `]`)
	matches := specialChars.FindAllString(content, -1)

	// Se mais de 30% do conteúdo são caracteres especiais
	return float64(len(matches))/float64(len(content)) > 0.3
}

// isValidPlaceholder verifica se um placeholder é válido
func (v *DefaultMessageValidator) isValidPlaceholder(placeholder string) bool {
	// Remover {{ e }}
	inner := strings.TrimPrefix(strings.TrimSuffix(placeholder, "}}"), "{{")

	// Verificar formato: deve começar com letra ou _, seguido de letras, números ou _
	validFormat := regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)
	return validFormat.MatchString(inner)
}

// isValidTemplateKey verifica se uma chave de template é válida
func (v *DefaultMessageValidator) isValidTemplateKey(key string) bool {
	// Mesmo formato que placeholders
	validFormat := regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)
	return validFormat.MatchString(key) && len(key) <= 50
}

// SetMaxContentLength configura o tamanho máximo do conteúdo
func (v *DefaultMessageValidator) SetMaxContentLength(length int) {
	v.maxContentLength = length
}

// SetMaxMediaSize configura o tamanho máximo de mídia
func (v *DefaultMessageValidator) SetMaxMediaSize(size int64) {
	v.maxMediaSize = size
}

// SetAllowedDomains configura domínios permitidos para mídia
func (v *DefaultMessageValidator) SetAllowedDomains(domains []string) {
	v.allowedDomains = domains
}

// SetBlockedWords configura palavras bloqueadas
func (v *DefaultMessageValidator) SetBlockedWords(words []string) {
	v.blockedWords = words
}

// ValidatePhoneNumber valida números de telefone (para contatos)
func (v *DefaultMessageValidator) ValidatePhoneNumber(phone string) error {
	// Remover caracteres não numéricos exceto +
	cleaned := regexp.MustCompile(`[^\d+]`).ReplaceAllString(phone, "")

	// Verificar formato básico
	phoneRegex := regexp.MustCompile(`^\+?[1-9]\d{1,14}$`)
	if !phoneRegex.MatchString(cleaned) {
		return fmt.Errorf("invalid phone number format")
	}

	return nil
}

// ValidateEmail valida endereços de email
func (v *DefaultMessageValidator) ValidateEmail(email string) error {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return fmt.Errorf("invalid email format")
	}

	return nil
}

// ValidateLocation valida dados de localização
func (v *DefaultMessageValidator) ValidateLocation(latitude, longitude float64) error {
	if latitude < -90 || latitude > 90 {
		return fmt.Errorf("latitude must be between -90 and 90")
	}

	if longitude < -180 || longitude > 180 {
		return fmt.Errorf("longitude must be between -180 and 180")
	}

	return nil
}

// ValidateVCard valida dados de vCard
func (v *DefaultMessageValidator) ValidateVCard(vcard string) error {
	if !strings.HasPrefix(vcard, "BEGIN:VCARD") {
		return fmt.Errorf("vCard must start with BEGIN:VCARD")
	}

	if !strings.HasSuffix(vcard, "END:VCARD") {
		return fmt.Errorf("vCard must end with END:VCARD")
	}

	// Verificar se tem pelo menos um campo obrigatório
	requiredFields := []string{"FN:", "N:"}
	hasRequired := false
	for _, field := range requiredFields {
		if strings.Contains(vcard, field) {
			hasRequired = true
			break
		}
	}

	if !hasRequired {
		return fmt.Errorf("vCard must contain at least one name field (FN or N)")
	}

	return nil
}
