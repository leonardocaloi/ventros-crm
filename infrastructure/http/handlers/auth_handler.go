package handlers

import (
	"net/http"

	"github.com/caloi/ventros-crm/infrastructure/http/middleware"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type AuthHandler struct {
	logger *zap.Logger
}

func NewAuthHandler(logger *zap.Logger) *AuthHandler {
	return &AuthHandler{
		logger: logger,
	}
}

// CreateUserRequest representa o payload para criar um usuário
type CreateUserRequest struct {
	Name     string `json:"name" binding:"required" example:"João Silva"`
	Email    string `json:"email" binding:"required" example:"joao@empresa.com"`
	Password string `json:"password" binding:"required" example:"senha123"`
	Role     string `json:"role" example:"user"`
}

// LoginRequest representa o payload de login
type LoginRequest struct {
	Email    string `json:"email" binding:"required" example:"joao@empresa.com"`
	Password string `json:"password" binding:"required" example:"senha123"`
}

// GenerateAPIKeyRequest representa o payload para gerar API key
type GenerateAPIKeyRequest struct {
	Name string `json:"name" example:"Minha API Key"`
}

// CreateUser creates a new user
// @Summary Create user
// @Description Cria um novo usuário no sistema
// @Tags auth
// @Accept json
// @Produce json
// @Param user body CreateUserRequest true "User data"
// @Success 201 {object} map[string]interface{} "User created successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/auth/register [post]
func (h *AuthHandler) CreateUser(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to parse user request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// TODO: Implementar criação real de usuário
	userID := uuid.New()
	apiKey := uuid.New().String()

	c.JSON(http.StatusCreated, gin.H{
		"message":  "User created successfully (mock)",
		"user_id":  userID,
		"name":     req.Name,
		"email":    req.Email,
		"role":     req.Role,
		"api_key":  apiKey,
		"note":     "Save this API key - it won't be shown again",
	})
}

// Login authenticates a user
// @Summary User login
// @Description Autentica um usuário e retorna API key
// @Tags auth
// @Accept json
// @Produce json
// @Param credentials body LoginRequest true "Login credentials"
// @Success 200 {object} map[string]interface{} "Login successful"
// @Failure 400 {object} map[string]interface{} "Invalid credentials"
// @Failure 401 {object} map[string]interface{} "Authentication failed"
// @Router /api/v1/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// TODO: Implementar validação real de credenciais
	// Mock para desenvolvimento
	if req.Email == "admin@dev.com" && req.Password == "admin123" {
		c.JSON(http.StatusOK, gin.H{
			"message":  "Login successful",
			"api_key":  "dev-admin-key",
			"user_id":  "00000000-0000-0000-0000-000000000001",
			"role":     "admin",
			"email":    req.Email,
		})
		return
	}

	if req.Email == "user@dev.com" && req.Password == "user123" {
		c.JSON(http.StatusOK, gin.H{
			"message":  "Login successful",
			"api_key":  "dev-user-key",
			"user_id":  "00000000-0000-0000-0000-000000000002",
			"role":     "user",
			"email":    req.Email,
		})
		return
	}

	c.JSON(http.StatusUnauthorized, gin.H{
		"error": "Invalid credentials",
		"hint":  "Try admin@dev.com/admin123 or user@dev.com/user123",
	})
}

// GetProfile gets current user profile
// @Summary Get user profile
// @Description Obtém o perfil do usuário autenticado
// @Tags auth
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} map[string]interface{} "User profile"
// @Failure 401 {object} map[string]interface{} "Not authenticated"
// @Router /api/v1/auth/profile [get]
func (h *AuthHandler) GetProfile(c *gin.Context) {
	authCtx, exists := middleware.GetAuthContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id":   authCtx.UserID,
		"email":     authCtx.Email,
		"role":      authCtx.Role,
		"tenant_id": authCtx.TenantID,
	})
}

// GenerateAPIKey generates a new API key for the user
// @Summary Generate API key
// @Description Gera uma nova API key para o usuário
// @Tags auth
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body GenerateAPIKeyRequest true "API key request"
// @Success 200 {object} map[string]interface{} "API key generated"
// @Failure 401 {object} map[string]interface{} "Not authenticated"
// @Router /api/v1/auth/api-key [post]
func (h *AuthHandler) GenerateAPIKey(c *gin.Context) {
	authCtx, exists := middleware.GetAuthContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	var req GenerateAPIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		req.Name = "Default API Key"
	}

	// TODO: Implementar geração real de API key
	apiKey := uuid.New().String()

	c.JSON(http.StatusOK, gin.H{
		"message":  "API key generated successfully",
		"api_key":  apiKey,
		"name":     req.Name,
		"user_id":  authCtx.UserID,
		"note":     "Save this API key - it won't be shown again",
	})
}

// GetAuthInfo provides authentication information for development
// @Summary Get auth info
// @Description Informações sobre autenticação para desenvolvimento
// @Tags auth
// @Produce json
// @Success 200 {object} map[string]interface{} "Auth information"
// @Router /api/v1/auth/info [get]
func (h *AuthHandler) GetAuthInfo(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Authentication methods available",
		"methods": map[string]interface{}{
			"dev_headers": map[string]string{
				"X-Dev-User-ID":   "UUID of the user (bypasses auth in dev mode)",
				"X-Dev-Email":     "Email (optional, defaults to dev@example.com)",
				"X-Dev-Role":      "Role (optional, defaults to admin)",
				"X-Dev-Tenant-ID": "Tenant ID (optional, defaults to dev-tenant)",
			},
			"api_key": map[string]string{
				"header":      "Authorization: Bearer <api_key>",
				"dev_keys":    "dev-admin-key, dev-user-key",
				"custom_keys": "Any UUID can be used as API key in dev mode",
			},
			"predefined_users": []map[string]string{
				{
					"email":    "admin@dev.com",
					"password": "admin123",
					"api_key":  "dev-admin-key",
					"role":     "admin",
				},
				{
					"email":    "user@dev.com",
					"password": "user123",
					"api_key":  "dev-user-key",
					"role":     "user",
				},
			},
		},
		"examples": map[string]interface{}{
			"curl_with_dev_header": "curl -H 'X-Dev-User-ID: 123e4567-e89b-12d3-a456-426614174000' http://localhost:8080/api/v1/auth/profile",
			"curl_with_api_key":    "curl -H 'Authorization: Bearer dev-admin-key' http://localhost:8080/api/v1/auth/profile",
		},
	})
}
