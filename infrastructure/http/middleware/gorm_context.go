package middleware

import (
	"context"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// GORMContextMiddleware injeta o contexto Gin no GORM para uso nos callbacks
func GORMContextMiddleware(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Criar um novo contexto com o Gin context embutido
		ctx := context.WithValue(c.Request.Context(), "gin_context", c)

		// Atualizar o request context
		c.Request = c.Request.WithContext(ctx)

		// Criar uma nova sessão GORM com o contexto atualizado
		c.Set("db", db.WithContext(ctx))

		c.Next()
	}
}
