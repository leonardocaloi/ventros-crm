package webhook

import (
	"context"
	"fmt"

	"github.com/caloi/ventros-crm/internal/domain/webhook"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// ManageSubscriptionUseCase gerencia inscrições de webhook
type ManageSubscriptionUseCase struct {
	repo   webhook.Repository
	logger *zap.Logger
}

// NewManageSubscriptionUseCase cria um novo use case
func NewManageSubscriptionUseCase(repo webhook.Repository, logger *zap.Logger) *ManageSubscriptionUseCase {
	return &ManageSubscriptionUseCase{
		repo:   repo,
		logger: logger,
	}
}

// CreateWebhook cria uma nova inscrição de webhook
func (uc *ManageSubscriptionUseCase) CreateWebhook(ctx context.Context, dto CreateWebhookDTO) (*WebhookDTO, error) {
	// Cria entidade de domínio
	sub, err := webhook.NewWebhookSubscription(dto.UserID, dto.ProjectID, dto.TenantID, dto.Name, dto.URL, dto.Events)
	if err != nil {
		return nil, fmt.Errorf("failed to create webhook: %w", err)
	}

	// Aplica configurações opcionais
	if dto.Secret != "" {
		sub.SetSecret(dto.Secret)
	}
	if dto.Headers != nil {
		sub.SetHeaders(dto.Headers)
	}
	if dto.RetryCount > 0 && dto.TimeoutSeconds > 0 {
		sub.SetRetryPolicy(dto.RetryCount, dto.TimeoutSeconds)
	}

	// Persiste
	if err := uc.repo.Create(ctx, sub); err != nil {
		uc.logger.Error("Failed to create webhook subscription",
			zap.Error(err),
			zap.String("name", dto.Name),
		)
		return nil, fmt.Errorf("failed to save webhook: %w", err)
	}

	uc.logger.Info("Webhook subscription created",
		zap.String("id", sub.ID.String()),
		zap.String("name", sub.Name),
		zap.Strings("events", sub.Events),
	)

	result := ToDTO(sub)
	return &result, nil
}

// GetWebhook busca um webhook por ID
func (uc *ManageSubscriptionUseCase) GetWebhook(ctx context.Context, id uuid.UUID) (*WebhookDTO, error) {
	sub, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	result := ToDTO(sub)
	return &result, nil
}

// ListWebhooks lista todos os webhooks
func (uc *ManageSubscriptionUseCase) ListWebhooks(ctx context.Context, activeOnly *bool) ([]WebhookDTO, error) {
	var webhooks []*webhook.WebhookSubscription
	var err error

	if activeOnly != nil {
		webhooks, err = uc.repo.FindByActive(ctx, *activeOnly)
	} else {
		webhooks, err = uc.repo.FindAll(ctx)
	}

	if err != nil {
		uc.logger.Error("Failed to list webhooks", zap.Error(err))
		return nil, fmt.Errorf("failed to list webhooks: %w", err)
	}

	return ToDTOList(webhooks), nil
}

// UpdateWebhook atualiza um webhook
func (uc *ManageSubscriptionUseCase) UpdateWebhook(ctx context.Context, id uuid.UUID, dto UpdateWebhookDTO) (*WebhookDTO, error) {
	// Busca webhook existente
	sub, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Aplica updates
	if dto.Name != nil {
		if err := sub.UpdateName(*dto.Name); err != nil {
			return nil, err
		}
	}
	if dto.URL != nil {
		if err := sub.UpdateURL(*dto.URL); err != nil {
			return nil, err
		}
	}
	if dto.Events != nil {
		if err := sub.UpdateEvents(dto.Events); err != nil {
			return nil, err
		}
	}
	if dto.Active != nil {
		if *dto.Active {
			sub.SetActive()
		} else {
			sub.SetInactive()
		}
	}
	if dto.Secret != nil {
		sub.SetSecret(*dto.Secret)
	}
	if dto.Headers != nil {
		sub.SetHeaders(dto.Headers)
	}
	if dto.RetryCount != nil && dto.TimeoutSeconds != nil {
		sub.SetRetryPolicy(*dto.RetryCount, *dto.TimeoutSeconds)
	}

	// Persiste
	if err := uc.repo.Update(ctx, sub); err != nil {
		uc.logger.Error("Failed to update webhook subscription",
			zap.Error(err),
			zap.String("id", id.String()),
		)
		return nil, fmt.Errorf("failed to update webhook: %w", err)
	}

	uc.logger.Info("Webhook subscription updated",
		zap.String("id", sub.ID.String()),
		zap.String("name", sub.Name),
	)

	result := ToDTO(sub)
	return &result, nil
}

// DeleteWebhook remove um webhook
func (uc *ManageSubscriptionUseCase) DeleteWebhook(ctx context.Context, id uuid.UUID) error {
	// Verifica se existe
	if _, err := uc.repo.FindByID(ctx, id); err != nil {
		return err
	}

	// Remove
	if err := uc.repo.Delete(ctx, id); err != nil {
		uc.logger.Error("Failed to delete webhook subscription",
			zap.Error(err),
			zap.String("id", id.String()),
		)
		return fmt.Errorf("failed to delete webhook: %w", err)
	}

	uc.logger.Info("Webhook subscription deleted", zap.String("id", id.String()))
	return nil
}

// GetAvailableEvents retorna eventos disponíveis (WAHA + Domínio/Aplicação)
func (uc *ManageSubscriptionUseCase) GetAvailableEvents() map[string]interface{} {
	return map[string]interface{}{
		// ===== EVENTOS DE DOMÍNIO/APLICAÇÃO (Internos) =====
		"domain_contacts": map[string]interface{}{
			"wildcard": "contact.*", // Subscreve todos os eventos de contato
			"events": []string{
				"contact.created",                 // Novo contato criado
				"contact.updated",                 // Contato atualizado
				"contact.deleted",                 // Contato deletado
				"contact.merged",                  // Contatos duplicados merged
				"contact.enriched",                // Dados externos adicionados
				"contact.profile_picture_updated", // Foto de perfil atualizada
			},
		},
		"domain_sessions": map[string]interface{}{
			"wildcard": "session.*", // Subscreve todos os eventos de sessão
			"events": []string{
				"session.created",        // Nova sessão criada (alias: session.started)
				"session.closed",         // Sessão encerrada (alias: session.ended)
				"session.agent_assigned", // Agente atribuído
				"session.resolved",       // Sessão resolvida
				"session.escalated",      // Sessão escalada
				"session.summarized",     // Resumo gerado por IA
				"session.abandoned",      // Sessão abandonada
			},
		},
		"domain_notes": map[string]interface{}{
			"wildcard": "note.*", // Subscreve todos os eventos de nota
			"events": []string{
				"note.added",   // Nota adicionada ao contato
				"note.updated", // Nota atualizada
				"note.deleted", // Nota deletada
				"note.pinned",  // Nota fixada
			},
		},
		"domain_tracking": map[string]interface{}{
			"wildcard": "tracking.*", // Subscreve todos os eventos de tracking
			"events": []string{
				"tracking.message.meta_ads", // Conversão de anúncio rastreada (Meta Ads: FB/Instagram)
				"tracking.created",          // Tracking criado
				"tracking.enriched",         // Tracking enriquecido com dados adicionais
			},
		},
		"domain_pipelines": map[string]interface{}{
			"wildcard": "pipeline.*", // Subscreve todos os eventos de pipeline
			"events": []string{
				"pipeline.created",         // Pipeline criado
				"pipeline.updated",         // Pipeline atualizado
				"pipeline.activated",       // Pipeline ativado
				"pipeline.deactivated",     // Pipeline desativado
				"pipeline.status.created",  // Status criado
				"pipeline.status.updated",  // Status atualizado
				"pipeline.status.changed",  // Status do contato alterado
				"contact.entered_pipeline", // Contato entrou no pipeline
				"contact.exited_pipeline",  // Contato saiu do pipeline
			},
		},
		"domain_agents": map[string]interface{}{
			"wildcard": "agent.*", // Subscreve todos os eventos de agente
			"events": []string{
				"agent.created",     // Agente criado
				"agent.updated",     // Agente atualizado
				"agent.activated",   // Agente ativado
				"agent.deactivated", // Agente desativado
			},
		},
		"domain_channels": map[string]interface{}{
			"wildcard": "channel.*", // Subscreve todos os eventos de canal
			"events": []string{
				"channel.created",     // Canal criado
				"channel.activated",   // Canal ativado
				"channel.deactivated", // Canal desativado
				"channel.deleted",     // Canal deletado
			},
		},

		// ===== EVENTOS WAHA (Externos - Canal WhatsApp) =====
		// Nota: Eventos de mensagens foram removidos conforme solicitado
		"waha_calls": map[string]interface{}{
			"wildcard": "call.*", // Subscreve todos os eventos de call
			"events": []string{
				"call.received",
				"call.accepted",
				"call.rejected",
			},
		},
		"waha_labels": map[string]interface{}{
			"wildcard": "label.*", // Subscreve todos os eventos de label
			"events": []string{
				"label.upsert",
				"label.deleted",
				"label.chat.added",
				"label.chat.deleted",
			},
		},
		"waha_groups": map[string]interface{}{
			"wildcard": "group.*", // Subscreve todos os eventos de group
			"events": []string{
				"group.v2.join",
				"group.v2.leave",
				"group.v2.update",
				"group.v2.participants",
			},
		},
		"waha_presence": map[string]interface{}{
			"wildcard": "presence.*", // Subscreve todos os eventos de presence
			"events": []string{
				"presence.update", // Status de presença (online/offline/typing)
			},
		},
		"waha_other": map[string]interface{}{
			"wildcard": "event.*", // Subscreve todos os eventos genéricos
			"events": []string{
				"event.response.failed",
			},
		},
	}
}
