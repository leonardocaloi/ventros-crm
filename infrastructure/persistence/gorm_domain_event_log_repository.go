package persistence

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/ventros/crm/infrastructure/persistence/entities"
	"github.com/ventros/crm/internal/domain/core/shared"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// DomainEventLogRepository é o repositório para logs de eventos de domínio
type DomainEventLogRepository struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewDomainEventLogRepository cria um novo repositório
func NewDomainEventLogRepository(db *gorm.DB, logger *zap.Logger) *DomainEventLogRepository {
	return &DomainEventLogRepository{
		db:     db,
		logger: logger,
	}
}

// LogEvent salva um evento de domínio no histórico
func (r *DomainEventLogRepository) LogEvent(ctx context.Context, event shared.DomainEvent, tenantID string, projectID, userID *uuid.UUID) error {
	// Serializa o evento completo como JSON
	payloadBytes, err := json.Marshal(event)
	if err != nil {
		r.logger.Error("Failed to marshal event payload", zap.Error(err))
		return err
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		r.logger.Error("Failed to unmarshal event payload", zap.Error(err))
		return err
	}

	// Extrai aggregate ID e type do evento (se disponível)
	aggregateID, aggregateType := r.extractAggregateInfo(event)

	entity := entities.DomainEventLogEntity{
		EventType:     event.EventName(),
		AggregateID:   aggregateID,
		AggregateType: aggregateType,
		TenantID:      tenantID,
		ProjectID:     projectID,
		UserID:        userID,
		Payload:       payload,
		OccurredAt:    event.OccurredAt(),
		PublishedAt:   time.Now(),
	}

	if err := r.db.WithContext(ctx).Create(&entity).Error; err != nil {
		r.logger.Error("Failed to log domain event",
			zap.String("event_type", event.EventName()),
			zap.Error(err))
		return err
	}

	r.logger.Debug("Domain event logged",
		zap.String("event_type", event.EventName()),
		zap.String("aggregate_id", aggregateID.String()),
		zap.String("aggregate_type", aggregateType))

	return nil
}

// extractAggregateInfo extrai ID e tipo da agregação do evento
func (r *DomainEventLogRepository) extractAggregateInfo(event shared.DomainEvent) (uuid.UUID, string) {
	// Usa type assertion para extrair informações específicas de cada tipo de evento
	// Isso pode ser melhorado com uma interface comum

	switch e := event.(type) {
	// Contact events
	case interface{ ContactID() uuid.UUID }:
		return e.ContactID(), "contact"

	// Session events
	case interface{ SessionID() uuid.UUID }:
		return e.SessionID(), "session"

	// Message events
	case interface{ MessageID() uuid.UUID }:
		return e.MessageID(), "message"

	default:
		// Se não conseguir extrair, retorna UUID zero
		return uuid.Nil, "unknown"
	}
}

// FindByAggregateID busca eventos por ID da agregação
func (r *DomainEventLogRepository) FindByAggregateID(ctx context.Context, aggregateID uuid.UUID) ([]entities.DomainEventLogEntity, error) {
	var logs []entities.DomainEventLogEntity

	err := r.db.WithContext(ctx).
		Where("aggregate_id = ?", aggregateID).
		Order("occurred_at DESC").
		Find(&logs).Error

	if err != nil {
		return nil, err
	}

	return logs, nil
}

// FindByEventType busca eventos por tipo
func (r *DomainEventLogRepository) FindByEventType(ctx context.Context, eventType string, limit int) ([]entities.DomainEventLogEntity, error) {
	var logs []entities.DomainEventLogEntity

	query := r.db.WithContext(ctx).
		Where("event_type = ?", eventType).
		Order("occurred_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Find(&logs).Error
	if err != nil {
		return nil, err
	}

	return logs, nil
}

// FindByProjectID busca eventos por projeto
func (r *DomainEventLogRepository) FindByProjectID(ctx context.Context, projectID uuid.UUID, limit int) ([]entities.DomainEventLogEntity, error) {
	var logs []entities.DomainEventLogEntity

	query := r.db.WithContext(ctx).
		Where("project_id = ?", projectID).
		Order("occurred_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Find(&logs).Error
	if err != nil {
		return nil, err
	}

	return logs, nil
}
