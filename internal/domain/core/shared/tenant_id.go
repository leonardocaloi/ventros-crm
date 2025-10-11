package shared

import "errors"

type TenantID struct {
	value string
}

func NewTenantID(value string) (TenantID, error) {
	if value == "" {
		return TenantID{}, errors.New("tenantID cannot be empty")
	}

	if len(value) < 3 {
		return TenantID{}, errors.New("tenantID too short")
	}

	return TenantID{value: value}, nil
}

func (t TenantID) String() string {
	return t.value
}

func (t TenantID) Equals(other TenantID) bool {
	return t.value == other.value
}
