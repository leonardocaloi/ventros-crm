package contact

import (
	"errors"
	"regexp"
	"strings"
)

// Email é um Value Object representando email válido.
type Email struct {
	value string
}

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

// NewEmail cria um novo Email com validação.
func NewEmail(value string) (Email, error) {
	value = strings.TrimSpace(strings.ToLower(value))

	if value == "" {
		return Email{}, errors.New("email cannot be empty")
	}

	if !emailRegex.MatchString(value) {
		return Email{}, errors.New("invalid email format")
	}

	return Email{value: value}, nil
}

// String retorna o valor do email.
func (e Email) String() string {
	return e.value
}

// Equals compara dois emails.
func (e Email) Equals(other Email) bool {
	return e.value == other.value
}

// Phone é um Value Object representando telefone.
type Phone struct {
	value string
}

// NewPhone cria um novo Phone com validação básica.
func NewPhone(value string) (Phone, error) {
	value = strings.TrimSpace(value)

	if value == "" {
		return Phone{}, errors.New("phone cannot be empty")
	}

	// Remove caracteres não numéricos
	cleaned := regexp.MustCompile(`[^0-9+]`).ReplaceAllString(value, "")

	if len(cleaned) < 8 {
		return Phone{}, errors.New("phone too short")
	}

	return Phone{value: cleaned}, nil
}

// String retorna o valor do telefone.
func (p Phone) String() string {
	return p.value
}

// Equals compara dois telefones.
func (p Phone) Equals(other Phone) bool {
	return p.value == other.value
}
