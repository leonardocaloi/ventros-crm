package webhook

import "errors"

var (
	// ErrInvalidName indica que o nome do webhook é inválido
	ErrInvalidName = errors.New("webhook name cannot be empty")

	// ErrInvalidURL indica que a URL do webhook é inválida
	ErrInvalidURL = errors.New("webhook URL cannot be empty")

	// ErrNoEvents indica que nenhum evento foi especificado
	ErrNoEvents = errors.New("at least one event must be specified")

	// ErrNotFound indica que o webhook não foi encontrado
	ErrNotFound = errors.New("webhook subscription not found")

	// ErrAlreadyExists indica que já existe um webhook com o mesmo nome
	ErrAlreadyExists = errors.New("webhook subscription already exists")
)
