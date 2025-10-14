package session

import "errors"

var (
	// Command validation errors
	ErrSessionIDRequired = errors.New("session_id is required")
	ErrReasonRequired    = errors.New("reason is required")
	ErrInvalidReason     = errors.New("invalid reason")

	// Business logic errors
	ErrSessionNotFound      = errors.New("session not found")
	ErrSessionAlreadyEnded  = errors.New("session is already ended")
	ErrSessionCloseFailed   = errors.New("failed to close session")
	ErrRepositorySaveFailed = errors.New("failed to save session")
)
