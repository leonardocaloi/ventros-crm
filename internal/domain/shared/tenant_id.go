package shared

import "errors"

// TenantID é um Value Object representando o ID do tenant (multi-tenancy).
type TenantID struct {
	value string
}

// NewTenantID cria um novo TenantID com validação.
func NewTenantID(value string) (TenantID, error) {
	if value == "" {
		return TenantID{}, errors.New("tenantID cannot be empty")
	}
	
	// Pode adicionar mais validações (formato, comprimento, etc.)
	if len(value) < 3 {
		return TenantID{}, errors.New("tenantID too short")
	}
	
	return TenantID{value: value}, nil
}

// String retorna o valor do tenant ID.
func (t TenantID) String() string {
	return t.value
}

// Equals compara dois tenant IDs.
func (t TenantID) Equals(other TenantID) bool {
	return t.value == other.value
}
