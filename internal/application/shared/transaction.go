package shared

import (
	"context"

	"gorm.io/gorm"
)

// TransactionManager gerencia transações de banco de dados.
// Permite executar múltiplas operações (Save + Publish) de forma atômica.
type TransactionManager interface {
	// ExecuteInTransaction executa uma função dentro de uma transação.
	// Se a função retornar erro, a transação é revertida (rollback).
	// Se a função retornar nil, a transação é confirmada (commit).
	ExecuteInTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}

// GormTransactionManager implementa TransactionManager usando GORM.
type GormTransactionManager struct {
	db *gorm.DB
}

// NewGormTransactionManager cria um novo transaction manager.
func NewGormTransactionManager(db *gorm.DB) TransactionManager {
	return &GormTransactionManager{db: db}
}

// ExecuteInTransaction executa uma função dentro de uma transação GORM.
//
// Exemplo de uso:
//
//	txManager.ExecuteInTransaction(ctx, func(ctx context.Context) error {
//	    // Todas as operações aqui usarão a mesma transação
//	    if err := contactRepo.SaveInTransaction(ctx, contact); err != nil {
//	        return err // Rollback automático
//	    }
//
//	    if err := eventBus.PublishInTransaction(ctx, contact.DomainEvents()...); err != nil {
//	        return err // Rollback automático
//	    }
//
//	    return nil // Commit automático
//	})
func (tm *GormTransactionManager) ExecuteInTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	// Inicia transação
	tx := tm.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return tx.Error
	}

	// Adiciona a transação ao contexto para que repositories e event bus possam usá-la
	ctx = ContextWithTransaction(ctx, tx)

	// Defer para garantir rollback em caso de panic
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r) // Re-panic após rollback
		}
	}()

	// Executa a função
	err := fn(ctx)
	if err != nil {
		// Rollback em caso de erro
		tx.Rollback()
		return err
	}

	// Commit se tudo ocorreu bem
	return tx.Commit().Error
}

// Chave para armazenar a transação no contexto
type transactionKey struct{}

// ContextWithTransaction adiciona a transação ao contexto.
func ContextWithTransaction(ctx context.Context, tx *gorm.DB) context.Context {
	return context.WithValue(ctx, transactionKey{}, tx)
}

// TransactionFromContext extrai a transação do contexto.
// Retorna nil se não houver transação no contexto.
func TransactionFromContext(ctx context.Context) *gorm.DB {
	if tx, ok := ctx.Value(transactionKey{}).(*gorm.DB); ok {
		return tx
	}
	return nil
}
