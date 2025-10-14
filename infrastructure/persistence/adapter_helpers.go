package persistence

import (
	"time"

	"github.com/google/uuid"
	"github.com/ventros/crm/internal/domain/core/shared"
)

// BaseEntityFields contém campos comuns a todas as entities GORM
type BaseEntityFields struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key"`
	Version   int       `gorm:"default:1;not null"` // Optimistic locking
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

// CopyBaseToEntity copia campos base do domain aggregate para entity
// Garante que version sempre será copiado (compile-time safety)
func CopyBaseToEntity(aggregate shared.AggregateRoot, entity *BaseEntityFields) {
	entity.ID = aggregate.ID()
	entity.Version = aggregate.Version()
	// CreatedAt e UpdatedAt são gerenciados pelo GORM
}

// BaseFromEntity retorna campos base da entity para reconstrução do domain
// Usado no toDomain() para criar aggregates a partir do banco
func BaseFromEntity(entity BaseEntityFields) (id uuid.UUID, version int, createdAt, updatedAt time.Time) {
	return entity.ID, entity.Version, entity.CreatedAt, entity.UpdatedAt
}

// ValidateAggregateRoot verifica se um aggregate implementa AggregateRoot corretamente
// Útil para testes e validação em tempo de compilação
func ValidateAggregateRoot(aggregate interface{}) bool {
	_, ok := aggregate.(shared.AggregateRoot)
	return ok
}

// NewBaseEntityFields cria campos base para nova entity
func NewBaseEntityFields(id uuid.UUID, version int) BaseEntityFields {
	now := time.Now()
	if version == 0 {
		version = 1 // Default para backwards compatibility
	}
	return BaseEntityFields{
		ID:        id,
		Version:   version,
		CreatedAt: now,
		UpdatedAt: now,
	}
}
