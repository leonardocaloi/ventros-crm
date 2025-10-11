package message_enrichment

import (
	"context"

	"github.com/google/uuid"
)

// Repository define a interface para persistência de MessageEnrichment
type Repository interface {
	// Save persiste um enrichment (create ou update)
	Save(ctx context.Context, enrichment *MessageEnrichment) error

	// FindByID busca um enrichment por ID
	FindByID(ctx context.Context, id uuid.UUID) (*MessageEnrichment, error)

	// FindByMessageID busca todos os enrichments de uma mensagem
	FindByMessageID(ctx context.Context, messageID uuid.UUID) ([]*MessageEnrichment, error)

	// FindByMessageGroupID busca todos os enrichments de um grupo de mensagens
	FindByMessageGroupID(ctx context.Context, messageGroupID uuid.UUID) ([]*MessageEnrichment, error)

	// FindPending busca enrichments pendentes (para processamento)
	// Retorna até 'limit' enrichments ordenados por prioridade (voice primeiro)
	FindPending(ctx context.Context, limit int) ([]*MessageEnrichment, error)

	// FindProcessing busca enrichments em processamento
	// Útil para detectar jobs travados
	FindProcessing(ctx context.Context, olderThan int) ([]*MessageEnrichment, error)

	// CountByStatus conta enrichments por status
	CountByStatus(ctx context.Context, status EnrichmentStatus) (int, error)

	// Delete remove um enrichment (raramente usado)
	Delete(ctx context.Context, id uuid.UUID) error
}
