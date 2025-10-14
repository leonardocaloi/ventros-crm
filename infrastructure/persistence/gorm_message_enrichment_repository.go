package persistence

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/ventros/crm/infrastructure/persistence/entities"
	"github.com/ventros/crm/internal/domain/crm/message_enrichment"
)

// GormMessageEnrichmentRepository implementa message_enrichment.Repository
type GormMessageEnrichmentRepository struct {
	db *gorm.DB
}

// NewGormMessageEnrichmentRepository cria uma nova instância do repositório
func NewGormMessageEnrichmentRepository(db *gorm.DB) *GormMessageEnrichmentRepository {
	return &GormMessageEnrichmentRepository{db: db}
}

// Save persiste um enrichment (create ou update)
func (r *GormMessageEnrichmentRepository) Save(ctx context.Context, enrichment *message_enrichment.MessageEnrichment) error {
	entity := r.toEntity(enrichment)

	// Check if exists
	var exists bool
	err := r.db.WithContext(ctx).
		Model(&entities.MessageEnrichmentEntity{}).
		Select("count(*) > 0").
		Where("id = ?", entity.ID).
		Find(&exists).Error

	if err != nil {
		return fmt.Errorf("failed to check enrichment existence: %w", err)
	}

	if exists {
		// Update
		if err := r.db.WithContext(ctx).Save(entity).Error; err != nil {
			return fmt.Errorf("failed to update enrichment: %w", err)
		}
	} else {
		// Create
		if err := r.db.WithContext(ctx).Create(entity).Error; err != nil {
			return fmt.Errorf("failed to create enrichment: %w", err)
		}
	}

	return nil
}

// FindByID busca um enrichment por ID
func (r *GormMessageEnrichmentRepository) FindByID(ctx context.Context, id uuid.UUID) (*message_enrichment.MessageEnrichment, error) {
	var entity entities.MessageEnrichmentEntity

	err := r.db.WithContext(ctx).
		Where("id = ?", id).
		First(&entity).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, message_enrichment.ErrEnrichmentNotFound
		}
		return nil, fmt.Errorf("failed to find enrichment: %w", err)
	}

	return r.toDomain(&entity)
}

// FindByMessageID busca todos os enrichments de uma mensagem
func (r *GormMessageEnrichmentRepository) FindByMessageID(ctx context.Context, messageID uuid.UUID) ([]*message_enrichment.MessageEnrichment, error) {
	var entities []entities.MessageEnrichmentEntity

	err := r.db.WithContext(ctx).
		Where("message_id = ?", messageID).
		Order("created_at ASC").
		Find(&entities).Error

	if err != nil {
		return nil, fmt.Errorf("failed to find enrichments by message ID: %w", err)
	}

	return r.toDomainSlice(entities)
}

// FindByMessageGroupID busca todos os enrichments de um grupo de mensagens
func (r *GormMessageEnrichmentRepository) FindByMessageGroupID(ctx context.Context, messageGroupID uuid.UUID) ([]*message_enrichment.MessageEnrichment, error) {
	var entities []entities.MessageEnrichmentEntity

	err := r.db.WithContext(ctx).
		Where("message_group_id = ?", messageGroupID).
		Order("created_at ASC").
		Find(&entities).Error

	if err != nil {
		return nil, fmt.Errorf("failed to find enrichments by message group ID: %w", err)
	}

	return r.toDomainSlice(entities)
}

// FindPending busca enrichments pendentes (para processamento)
// Retorna até 'limit' enrichments ordenados por prioridade (voice primeiro)
func (r *GormMessageEnrichmentRepository) FindPending(ctx context.Context, limit int) ([]*message_enrichment.MessageEnrichment, error) {
	var entities []entities.MessageEnrichmentEntity

	// Order by priority: voice (10) > audio (8) > image (7) > document (6) > video (3)
	// Using CASE statement to convert content_type to priority
	err := r.db.WithContext(ctx).
		Where("status = ?", "pending").
		Order(`
			CASE content_type
				WHEN 'voice' THEN 10
				WHEN 'audio' THEN 8
				WHEN 'image' THEN 7
				WHEN 'document' THEN 6
				WHEN 'video' THEN 3
				ELSE 5
			END DESC, created_at ASC
		`).
		Limit(limit).
		Find(&entities).Error

	if err != nil {
		return nil, fmt.Errorf("failed to find pending enrichments: %w", err)
	}

	return r.toDomainSlice(entities)
}

// FindProcessing busca enrichments em processamento
// Útil para detectar jobs travados (olderThan em minutos)
func (r *GormMessageEnrichmentRepository) FindProcessing(ctx context.Context, olderThanMinutes int) ([]*message_enrichment.MessageEnrichment, error) {
	var entities []entities.MessageEnrichmentEntity

	threshold := time.Now().Add(-time.Duration(olderThanMinutes) * time.Minute)

	err := r.db.WithContext(ctx).
		Where("status = ? AND created_at < ?", "processing", threshold).
		Order("created_at ASC").
		Find(&entities).Error

	if err != nil {
		return nil, fmt.Errorf("failed to find stuck processing enrichments: %w", err)
	}

	return r.toDomainSlice(entities)
}

// CountByStatus conta enrichments por status
func (r *GormMessageEnrichmentRepository) CountByStatus(ctx context.Context, status message_enrichment.EnrichmentStatus) (int, error) {
	var count int64

	err := r.db.WithContext(ctx).
		Model(&entities.MessageEnrichmentEntity{}).
		Where("status = ?", string(status)).
		Count(&count).Error

	if err != nil {
		return 0, fmt.Errorf("failed to count enrichments by status: %w", err)
	}

	return int(count), nil
}

// Delete remove um enrichment (raramente usado)
func (r *GormMessageEnrichmentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).
		Where("id = ?", id).
		Delete(&entities.MessageEnrichmentEntity{})

	if result.Error != nil {
		return fmt.Errorf("failed to delete enrichment: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return message_enrichment.ErrEnrichmentNotFound
	}

	return nil
}

// ==================== Conversion Methods ====================

// toDomain converte entity para domain
func (r *GormMessageEnrichmentRepository) toDomain(entity *entities.MessageEnrichmentEntity) (*message_enrichment.MessageEnrichment, error) {
	// Convert metadata from JSONB to map[string]interface{}
	metadata := make(map[string]interface{})
	if len(entity.Metadata) > 0 {
		// datatypes.JSON is already a []byte, just unmarshal it
		if err := json.Unmarshal(entity.Metadata, &metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
	}

	// Convert processing time from *int (milliseconds) to *time.Duration
	var processingTime *time.Duration
	if entity.ProcessingTime != nil {
		duration := time.Duration(*entity.ProcessingTime) * time.Millisecond
		processingTime = &duration
	}

	// Use Reconstitute to restore aggregate from database
	enrichment, err := message_enrichment.Reconstitute(
		entity.ID,
		entity.MessageID,
		entity.MessageGroupID,
		message_enrichment.EnrichmentContentType(entity.ContentType),
		message_enrichment.EnrichmentProvider(entity.Provider),
		entity.MediaURL,
		message_enrichment.EnrichmentStatus(entity.Status),
		entity.ExtractedText,
		metadata,
		processingTime,
		entity.Error,
		entity.Context,
		entity.CreatedAt,
		entity.ProcessedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to reconstitute enrichment: %w", err)
	}

	return enrichment, nil
}

// toEntity converte domain para entity
func (r *GormMessageEnrichmentRepository) toEntity(enrichment *message_enrichment.MessageEnrichment) *entities.MessageEnrichmentEntity {
	entity := &entities.MessageEnrichmentEntity{
		ID:             enrichment.ID(),
		MessageID:      enrichment.MessageID(),
		MessageGroupID: enrichment.MessageGroupID(),
		ContentType:    string(enrichment.ContentType()),
		Provider:       string(enrichment.Provider()),
		MediaURL:       enrichment.MediaURL(),
		Status:         string(enrichment.Status()),
		ExtractedText:  enrichment.ExtractedText(),
		Error:          enrichment.Error(),
		Context:        enrichment.Context(),
		CreatedAt:      enrichment.CreatedAt(),
		ProcessedAt:    enrichment.ProcessedAt(),
	}

	// Convert metadata map to JSONB
	if metadata := enrichment.Metadata(); len(metadata) > 0 {
		// Marshal to JSON bytes
		if jsonBytes, err := json.Marshal(metadata); err == nil {
			entity.Metadata = jsonBytes
		}
	}

	// Convert processing time from *time.Duration to *int (milliseconds)
	if pt := enrichment.ProcessingTime(); pt != nil {
		ms := int(pt.Milliseconds())
		entity.ProcessingTime = &ms
	}

	return entity
}

// toDomainSlice converte slice de entities para slice de domains
func (r *GormMessageEnrichmentRepository) toDomainSlice(entities []entities.MessageEnrichmentEntity) ([]*message_enrichment.MessageEnrichment, error) {
	result := make([]*message_enrichment.MessageEnrichment, 0, len(entities))

	for i := range entities {
		domain, err := r.toDomain(&entities[i])
		if err != nil {
			return nil, err
		}
		result = append(result, domain)
	}

	return result, nil
}
