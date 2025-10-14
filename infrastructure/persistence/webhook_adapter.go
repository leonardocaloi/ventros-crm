package persistence

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/ventros/crm/infrastructure/persistence/entities"
	"github.com/ventros/crm/internal/domain/crm/webhook"
	"gorm.io/gorm"
)

// WebhookRepositoryAdapter adapta o GormWebhookRepository para implementar webhook.Repository
type WebhookRepositoryAdapter struct {
	gormRepo *GormWebhookRepository
}

// NewWebhookRepositoryAdapter cria um novo adapter
func NewWebhookRepositoryAdapter(db *gorm.DB) webhook.Repository {
	return &WebhookRepositoryAdapter{
		gormRepo: NewGormWebhookRepository(db),
	}
}

// Create cria uma nova inscrição de webhook
func (a *WebhookRepositoryAdapter) Create(ctx context.Context, w *webhook.WebhookSubscription) error {
	entity, err := a.domainToEntity(w)
	if err != nil {
		return err
	}
	return a.gormRepo.Create(ctx, entity)
}

// FindByID busca um webhook por ID
func (a *WebhookRepositoryAdapter) FindByID(ctx context.Context, id uuid.UUID) (*webhook.WebhookSubscription, error) {
	entity, err := a.gormRepo.FindByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, webhook.ErrNotFound
		}
		return nil, err
	}
	return a.entityToDomain(entity)
}

// FindAll busca todos os webhooks
func (a *WebhookRepositoryAdapter) FindAll(ctx context.Context) ([]*webhook.WebhookSubscription, error) {
	entities, err := a.gormRepo.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	webhooks := make([]*webhook.WebhookSubscription, len(entities))
	for i, entity := range entities {
		w, err := a.entityToDomain(entity)
		if err != nil {
			return nil, err
		}
		webhooks[i] = w
	}
	return webhooks, nil
}

// FindActiveByEvent busca webhooks ativos inscritos em um evento
func (a *WebhookRepositoryAdapter) FindActiveByEvent(ctx context.Context, eventType string) ([]*webhook.WebhookSubscription, error) {
	// Para simplificar, vamos buscar todos os ativos e filtrar em memória
	// Em produção, seria melhor fazer isso no banco com uma query SQL customizada
	activeWebhooks, err := a.gormRepo.FindActive(ctx)
	if err != nil {
		return nil, err
	}

	var result []*webhook.WebhookSubscription
	for _, entity := range activeWebhooks {
		w, err := a.entityToDomain(entity)
		if err != nil {
			return nil, err
		}
		if w.IsSubscribedTo(eventType) {
			result = append(result, w)
		}
	}
	return result, nil
}

// FindByActive busca webhooks por status ativo
func (a *WebhookRepositoryAdapter) FindByActive(ctx context.Context, active bool) ([]*webhook.WebhookSubscription, error) {
	var webhookEntities []*entities.WebhookSubscriptionEntity
	var err error

	if active {
		webhookEntities, err = a.gormRepo.FindActive(ctx)
	} else {
		// Implementar FindInactive se necessário
		webhookEntities, err = a.gormRepo.FindAll(ctx)
		if err != nil {
			return nil, err
		}
		// Filtrar apenas os inativos
		var inactiveEntities []*entities.WebhookSubscriptionEntity
		for _, entity := range webhookEntities {
			if !entity.Active {
				inactiveEntities = append(inactiveEntities, entity)
			}
		}
		webhookEntities = inactiveEntities
	}

	if err != nil {
		return nil, err
	}

	webhooks := make([]*webhook.WebhookSubscription, len(webhookEntities))
	for i, entity := range webhookEntities {
		w, err := a.entityToDomain(entity)
		if err != nil {
			return nil, err
		}
		webhooks[i] = w
	}
	return webhooks, nil
}

// Update atualiza um webhook existente
func (a *WebhookRepositoryAdapter) Update(ctx context.Context, w *webhook.WebhookSubscription) error {
	entity, err := a.domainToEntity(w)
	if err != nil {
		return err
	}
	return a.gormRepo.Update(ctx, entity)
}

// Delete remove um webhook
func (a *WebhookRepositoryAdapter) Delete(ctx context.Context, id uuid.UUID) error {
	return a.gormRepo.Delete(ctx, id)
}

// RecordTrigger atualiza estatísticas de disparo do webhook
func (a *WebhookRepositoryAdapter) RecordTrigger(ctx context.Context, id uuid.UUID, success bool) error {
	w, err := a.FindByID(ctx, id)
	if err != nil {
		return err
	}

	w.RecordTrigger(success)
	return a.Update(ctx, w)
}

// domainToEntity converte webhook.WebhookSubscription para entities.WebhookSubscriptionEntity
func (a *WebhookRepositoryAdapter) domainToEntity(w *webhook.WebhookSubscription) (*entities.WebhookSubscriptionEntity, error) {
	// Converter headers map[string]string para JSON bytes
	var headersBytes []byte
	if len(w.Headers) > 0 {
		var err error
		headersBytes, err = json.Marshal(w.Headers)
		if err != nil {
			return nil, err
		}
	}

	// Converter []string para pq.StringArray
	events := make(pq.StringArray, len(w.Events))
	copy(events, w.Events)

	return &entities.WebhookSubscriptionEntity{
		ID:              w.ID,
		UserID:          w.UserID,
		ProjectID:       w.ProjectID,
		TenantID:        w.TenantID,
		Name:            w.Name,
		URL:             w.URL,
		Events:          events,
		Active:          w.Active,
		Secret:          w.Secret,
		Headers:         headersBytes,
		RetryCount:      w.RetryCount,
		TimeoutSeconds:  w.TimeoutSeconds,
		LastTriggeredAt: w.LastTriggeredAt,
		LastSuccessAt:   w.LastSuccessAt,
		LastFailureAt:   w.LastFailureAt,
		SuccessCount:    w.SuccessCount,
		FailureCount:    w.FailureCount,
		CreatedAt:       w.CreatedAt,
		UpdatedAt:       w.UpdatedAt,
	}, nil
}

// entityToDomain converte entities.WebhookSubscriptionEntity para webhook.WebhookSubscription
func (a *WebhookRepositoryAdapter) entityToDomain(e *entities.WebhookSubscriptionEntity) (*webhook.WebhookSubscription, error) {
	// Converter JSON bytes para map[string]string
	headers := make(map[string]string)
	if len(e.Headers) > 0 {
		err := json.Unmarshal(e.Headers, &headers)
		if err != nil {
			return nil, err
		}
	}

	// Converter pq.StringArray para []string
	events := make([]string, len(e.Events))
	copy(events, e.Events)

	return &webhook.WebhookSubscription{
		ID:              e.ID,
		UserID:          e.UserID,
		ProjectID:       e.ProjectID,
		TenantID:        e.TenantID,
		Name:            e.Name,
		URL:             e.URL,
		Events:          events,
		Active:          e.Active,
		Secret:          e.Secret,
		Headers:         headers,
		RetryCount:      e.RetryCount,
		TimeoutSeconds:  e.TimeoutSeconds,
		LastTriggeredAt: e.LastTriggeredAt,
		LastSuccessAt:   e.LastSuccessAt,
		LastFailureAt:   e.LastFailureAt,
		SuccessCount:    e.SuccessCount,
		FailureCount:    e.FailureCount,
		CreatedAt:       e.CreatedAt,
		UpdatedAt:       e.UpdatedAt,
	}, nil
}
