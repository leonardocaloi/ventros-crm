package middleware

import (
	"context"
	"database/sql"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// RLSMiddleware configura o user_id na sessão PostgreSQL para RLS
type RLSMiddleware struct {
	logger *zap.Logger
	db     *sql.DB
}

func NewRLSMiddleware(logger *zap.Logger, db *sql.DB) *RLSMiddleware {
	return &RLSMiddleware{
		logger: logger,
		db:     db,
	}
}

// SetUserContext define o user_id na sessão PostgreSQL para RLS
func (r *RLSMiddleware) SetUserContext() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Obter contexto de auth
		authCtx, exists := GetAuthContext(c)
		if !exists {
			// Se não autenticado, continua sem definir contexto RLS
			c.Next()
			return
		}

		// Definir user_id na sessão PostgreSQL
		if err := r.setPostgreSQLUserContext(c.Request.Context(), authCtx.UserID.String()); err != nil {
			r.logger.Error("Failed to set PostgreSQL user context for RLS", 
				zap.String("user_id", authCtx.UserID.String()),
				zap.Error(err))
			// Não falha a request, apenas loga o erro
		} else {
			r.logger.Debug("PostgreSQL user context set for RLS", 
				zap.String("user_id", authCtx.UserID.String()))
		}

		c.Next()
	}
}

// setPostgreSQLUserContext executa a função para definir o user_id na sessão
func (r *RLSMiddleware) setPostgreSQLUserContext(ctx context.Context, userID string) error {
	query := "SELECT set_current_user_id($1::uuid)"
	_, err := r.db.ExecContext(ctx, query, userID)
	return err
}

// ClearUserContext limpa o contexto do usuário (opcional, para cleanup)
func (r *RLSMiddleware) ClearUserContext() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			// Limpar contexto após a request
			if err := r.clearPostgreSQLUserContext(c.Request.Context()); err != nil {
				r.logger.Debug("Failed to clear PostgreSQL user context", zap.Error(err))
			}
		}()
		c.Next()
	}
}

// clearPostgreSQLUserContext limpa o user_id da sessão
func (r *RLSMiddleware) clearPostgreSQLUserContext(ctx context.Context) error {
	query := "SELECT set_config('app.current_user_id', '', false)"
	_, err := r.db.ExecContext(ctx, query)
	return err
}
