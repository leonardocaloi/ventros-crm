package user

import "fmt"

// Role representa uma role de usuário no sistema
type Role string

const (
	// RoleAdmin - Administrador do sistema (acesso total)
	RoleAdmin Role = "admin"

	// RoleUser - Usuário padrão (acesso aos próprios recursos)
	RoleUser Role = "user"

	// RoleManager - Gerente (pode ver recursos de sua equipe)
	RoleManager Role = "manager"

	// RoleReadOnly - Apenas leitura
	RoleReadOnly Role = "readonly"
)

// AllRoles retorna todas as roles válidas
func AllRoles() []Role {
	return []Role{RoleAdmin, RoleUser, RoleManager, RoleReadOnly}
}

// IsValid verifica se a role é válida
func (r Role) IsValid() bool {
	for _, validRole := range AllRoles() {
		if r == validRole {
			return true
		}
	}
	return false
}

// String implementa fmt.Stringer
func (r Role) String() string {
	return string(r)
}

// HasPermission verifica se a role tem uma permissão específica
func (r Role) HasPermission(permission Permission) bool {
	switch r {
	case RoleAdmin:
		return true // Admin tem todas as permissões
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

// CanAccessResource verifica se a role pode acessar um recurso específico
func (r Role) CanAccessResource(resourceType ResourceType, operation Operation) bool {
	permission := NewPermission(resourceType, operation)
	return r.HasPermission(permission)
}

// ParseRole converte string para Role
func ParseRole(s string) (Role, error) {
	role := Role(s)
	if !role.IsValid() {
		return "", fmt.Errorf("invalid role: %s", s)
	}
	return role, nil
}
