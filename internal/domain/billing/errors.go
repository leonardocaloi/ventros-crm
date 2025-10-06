package billing

import "errors"

var (
	// ErrNotFound é retornado quando a conta de faturamento não é encontrada
	ErrNotFound = errors.New("billing account not found")
)
