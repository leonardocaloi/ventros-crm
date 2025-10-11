package middleware

import (
	"net/http"

	"github.com/caloi/ventros-crm/internal/domain/core/user"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// RBACMiddleware implementa Role-Based Access Control
type RBACMiddleware struct {
	logger *zap.Logger
}

// NewRBACMiddleware cria um novo middleware RBAC
func NewRBACMiddleware(logger *zap.Logger) *RBACMiddleware {
	return &RBACMiddleware{
		logger: logger,
	}
}

// RequirePermission verifica se o usuário tem permissão para acessar um recurso
func (m *RBACMiddleware) RequirePermission(resource user.ResourceType, operation user.Operation) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Obter contexto de autenticação
		authCtx, exists := GetAuthContext(c)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			c.Abort()
			return
		}

		// Parse da role do usuário
		userRole, err := user.ParseRole(authCtx.Role)
		if err != nil {
			m.logger.Error("Invalid user role",
				zap.String("role", authCtx.Role),
				zap.String("user_id", authCtx.UserID.String()),
				zap.Error(err))
			c.JSON(http.StatusForbidden, gin.H{"error": "Invalid user role"})
			c.Abort()
			return
		}

		// Verificar permissão
		if !userRole.CanAccessResource(resource, operation) {
			m.logger.Warn("Access denied",
				zap.String("user_id", authCtx.UserID.String()),
				zap.String("role", authCtx.Role),
				zap.String("resource", string(resource)),
				zap.String("operation", string(operation)))

			c.JSON(http.StatusForbidden, gin.H{
				"error": "Access denied",
				"details": map[string]string{
					"resource":  string(resource),
					"operation": string(operation),
					"required":  "Insufficient permissions",
				},
			})
			c.Abort()
			return
		}

		// Log acesso autorizado
		m.logger.Debug("Access granted",
			zap.String("user_id", authCtx.UserID.String()),
			zap.String("role", authCtx.Role),
			zap.String("resource", string(resource)),
			zap.String("operation", string(operation)))

		c.Next()
	}
}

// RequireRole verifica se o usuário tem uma role específica
func (m *RBACMiddleware) RequireRole(requiredRole user.Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		authCtx, exists := GetAuthContext(c)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			c.Abort()
			return
		}

		userRole, err := user.ParseRole(authCtx.Role)
		if err != nil || userRole != requiredRole {
			c.JSON(http.StatusForbidden, gin.H{
				"error":    "Insufficient role",
				"required": string(requiredRole),
				"current":  authCtx.Role,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAnyRole verifica se o usuário tem pelo menos uma das roles especificadas
func (m *RBACMiddleware) RequireAnyRole(roles ...user.Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		authCtx, exists := GetAuthContext(c)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			c.Abort()
			return
		}

		userRole, err := user.ParseRole(authCtx.Role)
		if err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "Invalid user role"})
			c.Abort()
			return
		}

		// Verificar se tem pelo menos uma das roles
		for _, role := range roles {
			if userRole == role {
				c.Next()
				return
			}
		}

		c.JSON(http.StatusForbidden, gin.H{
			"error":        "Insufficient role",
			"required_any": roles,
			"current":      authCtx.Role,
		})
		c.Abort()
	}
}

// IsAdmin verifica se o usuário é admin
func (m *RBACMiddleware) IsAdmin() gin.HandlerFunc {
	return m.RequireRole(user.RoleAdmin)
}

// CanManage verifica se o usuário pode gerenciar recursos (admin ou manager)
func (m *RBACMiddleware) CanManage() gin.HandlerFunc {
	return m.RequireAnyRole(user.RoleAdmin, user.RoleManager)
}
