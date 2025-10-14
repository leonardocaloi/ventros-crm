package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ventros/crm/internal/application/user"
	"go.uber.org/zap"
)

// AuthContext representa o contexto de autenticação
type AuthContext struct {
	UserID    uuid.UUID `json:"user_id"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	TenantID  string    `json:"tenant_id"`
	ProjectID uuid.UUID `json:"project_id"`
}

// AuthMiddleware é um middleware flexível para desenvolvimento
type AuthMiddleware struct {
	logger      *zap.Logger
	devMode     bool
	userService *user.UserService
}

func NewAuthMiddleware(logger *zap.Logger, devMode bool, userService *user.UserService) *AuthMiddleware {
	return &AuthMiddleware{
		logger:      logger,
		devMode:     devMode,
		userService: userService,
	}
}

// Authenticate middleware flexível para desenvolvimento
func (a *AuthMiddleware) Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Em modo dev, permite bypass com headers especiais
		if a.devMode {
			if authCtx := a.handleDevAuth(c); authCtx != nil {
				c.Set("auth", authCtx)
				c.Next()
				return
			}
		}

		// Auth normal (API Key ou JWT)
		if authCtx := a.handleAPIKeyAuth(c); authCtx != nil {
			c.Set("auth", authCtx)
			c.Next()
			return
		}

		// Se chegou aqui, não autenticado
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Authentication required",
			"hint":  "Use X-Dev-User-ID header in dev mode or Authorization: Bearer <api_key>",
		})
		c.Abort()
	}
}

// handleDevAuth permite bypass em desenvolvimento usando headers
func (a *AuthMiddleware) handleDevAuth(c *gin.Context) *AuthContext {
	userID := c.GetHeader("X-Dev-User-ID")
	if userID == "" {
		return nil
	}

	parsedUserID, err := uuid.Parse(userID)
	if err != nil {
		a.logger.Warn("Invalid dev user ID", zap.String("user_id", userID))
		return nil
	}

	// Headers opcionais para desenvolvimento
	email := c.GetHeader("X-Dev-Email")
	if email == "" {
		email = "dev@example.com"
	}

	role := c.GetHeader("X-Dev-Role")
	if role == "" {
		role = "admin"
	}

	tenantID := c.GetHeader("X-Dev-Tenant-ID")
	if tenantID == "" {
		tenantID = "dev-tenant"
	}

	// Pegar project_id do header
	projectIDStr := c.GetHeader("X-Dev-Project-ID")
	var projectID uuid.UUID
	if projectIDStr != "" {
		parsedProjectID, err := uuid.Parse(projectIDStr)
		if err == nil {
			projectID = parsedProjectID
		}
	}

	a.logger.Info("Dev auth bypass",
		zap.String("user_id", userID),
		zap.String("email", email),
		zap.String("role", role),
	)

	return &AuthContext{
		UserID:    parsedUserID,
		Email:     email,
		Role:      role,
		TenantID:  tenantID,
		ProjectID: projectID,
	}
}

// handleAPIKeyAuth autentica usando API Key
func (a *AuthMiddleware) handleAPIKeyAuth(c *gin.Context) *AuthContext {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return nil
	}

	// Suporte para Bearer token
	if strings.HasPrefix(authHeader, "Bearer ") {
		apiKey := strings.TrimPrefix(authHeader, "Bearer ")
		return a.validateAPIKey(apiKey)
	}

	// Suporte para API Key direto
	return a.validateAPIKey(authHeader)
}

// validateAPIKey valida a API key usando o UserService
func (a *AuthMiddleware) validateAPIKey(apiKey string) *AuthContext {
	if len(apiKey) < 10 {
		return nil
	}

	// Em modo dev, mantém keys de desenvolvimento para facilitar testes
	if a.devMode {
		switch apiKey {
		case "dev-admin-key":
			return &AuthContext{
				UserID:    uuid.MustParse("00000000-0000-0000-0000-000000000001"),
				Email:     "admin@dev.com",
				Role:      "admin",
				TenantID:  "dev-tenant",
				ProjectID: uuid.MustParse("11111111-1111-1111-1111-111111111111"),
			}
		case "dev-user-key":
			return &AuthContext{
				UserID:    uuid.MustParse("00000000-0000-0000-0000-000000000002"),
				Email:     "user@dev.com",
				Role:      "user",
				TenantID:  "dev-tenant",
				ProjectID: uuid.MustParse("11111111-1111-1111-1111-111111111111"),
			}
		}
	}

	// Validação real usando UserService
	if a.userService != nil {
		userEntity, projectEntity, err := a.userService.ValidateAPIKey(apiKey)
		if err != nil {
			a.logger.Debug("API key validation failed", zap.Error(err))
			return nil
		}

		return &AuthContext{
			UserID:    userEntity.ID,
			Email:     userEntity.Email,
			Role:      userEntity.Role,
			TenantID:  projectEntity.TenantID,
			ProjectID: projectEntity.ID,
		}
	}

	return nil
}

// GetAuthContext extrai o contexto de auth da request
func GetAuthContext(c *gin.Context) (*AuthContext, bool) {
	auth, exists := c.Get("auth")
	if !exists {
		return nil, false
	}

	authCtx, ok := auth.(*AuthContext)
	return authCtx, ok
}

// RequireRole middleware que exige uma role específica
func RequireRole(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authCtx, exists := GetAuthContext(c)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			c.Abort()
			return
		}

		if authCtx.Role != role && authCtx.Role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{
				"error":         "Insufficient permissions",
				"required_role": role,
				"your_role":     authCtx.Role,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
