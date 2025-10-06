package customer

import "errors"

// Status representa o status de um cliente.
type Status string

const (
	StatusActive    Status = "active"
	StatusInactive  Status = "inactive"
	StatusSuspended Status = "suspended"
)

func (s Status) String() string {
	return string(s)
}

// IsValid verifica se o status é válido.
func (s Status) IsValid() bool {
	switch s {
	case StatusActive, StatusInactive, StatusSuspended:
		return true
	default:
		return false
	}
}

// ParseStatus converte string para Status.
func ParseStatus(s string) (Status, error) {
	status := Status(s)
	if !status.IsValid() {
		return "", errors.New("invalid status")
	}
	return status, nil
}
