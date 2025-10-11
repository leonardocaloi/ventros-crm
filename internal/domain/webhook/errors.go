package webhook

import "errors"

var (
	ErrInvalidName = errors.New("webhook name cannot be empty")

	ErrInvalidURL = errors.New("webhook URL cannot be empty")

	ErrNoEvents = errors.New("at least one event must be specified")

	ErrNotFound = errors.New("webhook subscription not found")

	ErrAlreadyExists = errors.New("webhook subscription already exists")
)
