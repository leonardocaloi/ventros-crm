package message_enrichment

import "errors"

var (
	// ErrEnrichmentNotFound é retornado quando um enrichment não é encontrado
	ErrEnrichmentNotFound = errors.New("enrichment not found")

	// ErrInvalidContentType é retornado quando o tipo de conteúdo é inválido
	ErrInvalidContentType = errors.New("invalid content type")

	// ErrInvalidProvider é retornado quando o provider é inválido
	ErrInvalidProvider = errors.New("invalid provider")

	// ErrInvalidStatus é retornado quando o status é inválido
	ErrInvalidStatus = errors.New("invalid status")

	// ErrProcessingNotAllowed é retornado quando a transição de status não é permitida
	ErrProcessingNotAllowed = errors.New("processing not allowed in current status")

	// ErrMediaURLEmpty é retornado quando a URL da mídia está vazia
	ErrMediaURLEmpty = errors.New("media URL cannot be empty")

	// ErrProviderUnavailable é retornado quando o provider não está disponível
	ErrProviderUnavailable = errors.New("provider unavailable")

	// ErrProcessingTimeout é retornado quando o processamento excede o timeout
	ErrProcessingTimeout = errors.New("processing timeout exceeded")
)
