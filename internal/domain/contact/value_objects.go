package contact

import (
	"errors"
	"regexp"
	"strings"
)

type Email struct {
	value string
}

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

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

func (e Email) String() string {
	return e.value
}

func (e Email) Equals(other Email) bool {
	return e.value == other.value
}

type Phone struct {
	value string
}

func NewPhone(value string) (Phone, error) {
	value = strings.TrimSpace(value)

	if value == "" {
		return Phone{}, errors.New("phone cannot be empty")
	}

	cleaned := regexp.MustCompile(`[^0-9+]`).ReplaceAllString(value, "")

	if len(cleaned) < 8 {
		return Phone{}, errors.New("phone too short")
	}

	return Phone{value: cleaned}, nil
}

func (p Phone) String() string {
	return p.value
}

func (p Phone) Equals(other Phone) bool {
	return p.value == other.value
}
