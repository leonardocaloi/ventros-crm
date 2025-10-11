package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// WebSocketAuthMiddleware autentica conexões WebSocket
// Suporta autenticação via:
// 1. Authorization header (Bearer token)
// 2. Query parameter ?token=<token> (para compatibilidade com clientes que não suportam headers em WebSocket)
type WebSocketAuthMiddleware struct {
	authMiddleware *AuthMiddleware
	logger         *zap.Logger
}

// NewWebSocketAuthMiddleware cria novo middleware
func NewWebSocketAuthMiddleware(authMiddleware *AuthMiddleware, logger *zap.Logger) *WebSocketAuthMiddleware {
	return &WebSocketAuthMiddleware{
		authMiddleware: authMiddleware,
		logger:         logger,
	}
}

// Authenticate middleware para WebSocket
func (m *WebSocketAuthMiddleware) Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Tentar autenticação via header primeiro
		if authCtx := m.tryHeaderAuth(c); authCtx != nil {
			c.Set("auth", authCtx)
			c.Next()
			return
		}

		// Tentar autenticação via query parameter (fallback para WebSocket)
		if authCtx := m.tryQueryAuth(c); authCtx != nil {
			c.Set("auth", authCtx)
			c.Next()
			return
		}

		// Não autenticado
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Authentication required",
			"hint":  "Use Authorization: Bearer <token> header or ?token=<token> query param",
		})
		c.Abort()
	}
}

// tryHeaderAuth tenta autenticação via Authorization header
func (m *WebSocketAuthMiddleware) tryHeaderAuth(c *gin.Context) *AuthContext {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return nil
	}

	// Usar middleware de auth existente
	return m.authMiddleware.handleAPIKeyAuth(c)
}

// tryQueryAuth tenta autenticação via query parameter
func (m *WebSocketAuthMiddleware) tryQueryAuth(c *gin.Context) *AuthContext {
	token := c.Query("token")
	if token == "" {
		return nil
	}

	m.logger.Debug("Authenticating WebSocket via query parameter",
		zap.String("remote_addr", c.ClientIP()))

	// SECURITY: Validar token
	// Temporariamente seta no header para reusar lógica existente
	c.Request.Header.Set("Authorization", "Bearer "+token)
	authCtx := m.authMiddleware.handleAPIKeyAuth(c)

	// Limpar header (não poluir)
	c.Request.Header.Del("Authorization")

	return authCtx
}

// ValidateWebSocketSession valida se usuário tem acesso a uma sessão
// SECURITY: CRITICAL - previne acesso não autorizado a sessões
func ValidateWebSocketSession(c *gin.Context, sessionID uuid.UUID) (*AuthContext, error) {
	authCtx, exists := GetAuthContext(c)
	if !exists {
		return nil, ErrUnauthorized
	}

	// TODO: Implementar validação de permissão via repository
	// Verificar se usuário tem acesso a essa sessão/projeto
	// Por ora, assume que auth middleware já validou tenant/project

	return authCtx, nil
}

var (
	ErrUnauthorized = &AuthError{Code: "unauthorized", Message: "Authentication required"}
	ErrForbidden    = &AuthError{Code: "forbidden", Message: "Access denied"}
)

type AuthError struct {
	Code    string
	Message string
}

func (e *AuthError) Error() string {
	return e.Message
}
