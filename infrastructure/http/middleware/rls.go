package middleware

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// RLSMiddleware configura o user_id no contexto Gin para ser usado pelo GORM callback
type RLSMiddleware struct {
	logger *zap.Logger
}

func NewRLSMiddleware(logger *zap.Logger) *RLSMiddleware {
	return &RLSMiddleware{
		logger: logger,
	}
}

// SetUserContext armazena o user_id no contexto Gin para uso posterior
func (r *RLSMiddleware) SetUserContext() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Obter contexto de auth
		authCtx, exists := GetAuthContext(c)
		if !exists {
			// Se não autenticado, continua sem definir contexto RLS
			c.Next()
			return
		}

		// Armazenar user_id no contexto Gin para o GORM callback usar
		c.Set("rls_user_id", authCtx.UserID.String())
		
		r.logger.Debug("RLS user context stored in Gin context", 
			zap.String("user_id", authCtx.UserID.String()),
			zap.String("project_id", authCtx.ProjectID.String()))

		c.Next()
	}
}
