package customer

import "errors"

type Status string

const (
	StatusActive    Status = "active"
	StatusInactive  Status = "inactive"
	StatusSuspended Status = "suspended"
)

func (s Status) String() string {
	return string(s)
}

func (s Status) IsValid() bool {
	switch s {
	case StatusActive, StatusInactive, StatusSuspended:
		return true
	default:
		return false
	}
}

func ParseStatus(s string) (Status, error) {
	status := Status(s)
	if !status.IsValid() {
		return "", errors.New("invalid status")
	}
	return status, nil
}
