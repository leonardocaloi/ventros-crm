package contact_event

import "errors"

var (
	ErrContactEventNotFound = errors.New("contact event not found")

	ErrInvalidContactEvent = errors.New("invalid contact event")
)
