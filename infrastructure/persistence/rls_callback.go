package persistence

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RegisterRLSCallbacks registra callbacks GORM para aplicar RLS em cada query
func RegisterRLSCallbacks(db *gorm.DB) error {
	// Callback ANTES de cada query (SELECT)
	// Registrar com Replace para garantir que seja executado
	err := db.Callback().Query().Before("gorm:query").Register("rls:set_user_query", func(db *gorm.DB) {
		setRLSContext(db)
	})
	if err != nil {
		return fmt.Errorf("failed to register RLS query callback: %w", err)
	}

	// Callback ANTES de cada CREATE
	err = db.Callback().Create().Before("gorm:create").Register("rls:set_user_create", func(db *gorm.DB) {
		setRLSContext(db)
	})
	if err != nil {
		return fmt.Errorf("failed to register RLS create callback: %w", err)
	}

	// Callback ANTES de cada UPDATE
	err = db.Callback().Update().Before("gorm:update").Register("rls:set_user_update", func(db *gorm.DB) {
		setRLSContext(db)
	})
	if err != nil {
		return fmt.Errorf("failed to register RLS update callback: %w", err)
	}

	// Callback ANTES de cada DELETE
	err = db.Callback().Delete().Before("gorm:delete").Register("rls:set_user_delete", func(db *gorm.DB) {
		setRLSContext(db)
	})
	if err != nil {
		return fmt.Errorf("failed to register RLS delete callback: %w", err)
	}

	// Callback ANTES de cada RAW query
	err = db.Callback().Raw().Before("gorm:raw").Register("rls:set_user_raw", func(db *gorm.DB) {
		setRLSContext(db)
	})
	if err != nil {
		return fmt.Errorf("failed to register RLS raw callback: %w", err)
	}

	// Callback ANTES de cada ROW query
	err = db.Callback().Row().Before("gorm:row").Register("rls:set_user_row", func(db *gorm.DB) {
		setRLSContext(db)
	})
	if err != nil {
		return fmt.Errorf("failed to register RLS row callback: %w", err)
	}

	return nil
}

// setRLSContext define o contexto RLS na sessão PostgreSQL usando SET LOCAL
func setRLSContext(db *gorm.DB) {
	// Verificar se já executamos SET LOCAL nesta statement para evitar recursão
	if _, ok := db.Statement.Settings.Load("rls_set"); ok {
		return
	}

	// Marcar que já executamos para esta statement
	db.Statement.Settings.Store("rls_set", true)

	// Debug: verificar se o contexto existe
	if db.Statement.Context == nil {
		return
	}

	// Tentar obter o user_id do contexto Gin
	ginCtxValue := db.Statement.Context.Value("gin_context")
	if ginCtxValue == nil {
		return
	}

	ginCtx, ok := ginCtxValue.(*gin.Context)
	if !ok {
		return
	}

	userID, exists := ginCtx.Get("rls_user_id")
	if !exists {
		return
	}

	userIDStr, ok := userID.(string)
	if !ok {
		return
	}

	// Executar SET LOCAL diretamente na conexão SQL para evitar recursão
	sql := fmt.Sprintf("SET LOCAL app.current_user_id = '%s'", userIDStr)

	// Usar a conexão SQL direta do statement
	if db.Statement.ConnPool != nil {
		_, err := db.Statement.ConnPool.ExecContext(db.Statement.Context, sql)
		if err != nil {
			log.Printf("⚠️  RLS: Failed to execute SET LOCAL: %v", err)
		} else {
			log.Printf("✅ RLS: SET LOCAL app.current_user_id = '%s'", userIDStr)
		}
	}
}
