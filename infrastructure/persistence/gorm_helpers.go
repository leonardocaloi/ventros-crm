package persistence

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// OnConflictDoNothing retorna uma cláusula ON CONFLICT DO NOTHING para PostgreSQL.
// Útil para operações idempotentes.
func OnConflictDoNothing(constraintName string) clause.OnConflict {
	return clause.OnConflict{
		Columns:   []clause.Column{}, // Vazio usa a constraint definida
		DoNothing: true,
	}
}

// WithTransaction executa uma função dentro de uma transação.
// Se a função retornar erro, faz rollback. Caso contrário, commit.
func WithTransaction(db *gorm.DB, fn func(tx *gorm.DB) error) error {
	tx := db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	if err := fn(tx); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}
