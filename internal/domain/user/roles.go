package user

import "fmt"

type Role string

const (
	RoleAdmin Role = "admin"

	RoleUser Role = "user"

	RoleManager Role = "manager"

	RoleReadOnly Role = "readonly"
)

func AllRoles() []Role {
	return []Role{RoleAdmin, RoleUser, RoleManager, RoleReadOnly}
}

func (r Role) IsValid() bool {
	for _, validRole := range AllRoles() {
		if r == validRole {
			return true
		}
	}
	return false
}

func (r Role) String() string {
	return string(r)
}

func (r Role) HasPermission(permission Permission) bool {
	switch r {
	case RoleAdmin:
		return true
	case RoleManager:
		return permission.IsManagerAllowed()
	case RoleUser:
		return permission.IsUserAllowed()
	case RoleReadOnly:
		return permission.IsReadOnlyAllowed()
	default:
		return false
	}
}

func (r Role) CanAccessResource(resourceType ResourceType, operation Operation) bool {
	permission := NewPermission(resourceType, operation)
	return r.HasPermission(permission)
}

func ParseRole(s string) (Role, error) {
	role := Role(s)
	if !role.IsValid() {
		return "", fmt.Errorf("invalid role: %s", s)
	}
	return role, nil
}
