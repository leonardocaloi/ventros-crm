package contact_event

import "errors"

var (
	// ErrContactEventNotFound é retornado quando um evento de contato não é encontrado.
	ErrContactEventNotFound = errors.New("contact event not found")
	
	// ErrInvalidContactEvent é retornado quando um evento de contato é inválido.
	ErrInvalidContactEvent = errors.New("invalid contact event")
)
