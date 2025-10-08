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
func (uc *ManageSubscriptionUseCase) GetAvailableEvents() map[string][]string {
	return map[string][]string{
		// ===== EVENTOS DE DOMÍNIO/APLICAÇÃO (Internos) =====
		"domain_contacts": {
			"contact.created",     // Novo contato criado
			"contact.updated",     // Contato atualizado
			"contact.deleted",     // Contato deletado
			"contact.merged",      // Contatos duplicados merged
			"contact.enriched",    // Dados externos adicionados
		},
		"domain_sessions": {
			"session.started",           // Nova sessão iniciada
			"session.ended",             // Sessão encerrada
			"session.message_recorded",  // Mensagem registrada na sessão
			"session.agent_assigned",    // Agente atribuído
			"session.resolved",          // Sessão resolvida
			"session.escalated",         // Sessão escalada
			"session.summarized",        // Resumo gerado por IA
			"session.abandoned",         // Sessão abandonada
		},
		"domain_messages": {
			"message.created",    // Mensagem criada no sistema
			"message.delivered",  // Mensagem entregue
			"message.read",       // Mensagem lida
			"message.failed",     // Mensagem falhou
		},
		"domain_tracking": {
			"tracking.message.meta_ads", // Conversão de anúncio rastreada (Meta Ads: FB/Instagram)
		},
		"domain_pipelines": {
			"pipeline.created",          // Pipeline criado
			"pipeline.updated",          // Pipeline atualizado
			"pipeline.activated",        // Pipeline ativado
			"pipeline.deactivated",      // Pipeline desativado
			"status.created",            // Status criado
			"status.updated",            // Status atualizado
			"contact.status_changed",    // Status do contato alterado
			"contact.entered_pipeline",  // Contato entrou no pipeline
			"contact.exited_pipeline",   // Contato saiu do pipeline
		},
		
		// ===== EVENTOS WAHA (Externos - Canal WhatsApp) =====
		"waha_messages": {
			"message",           // Mensagens incoming (fromMe: false)
			"message.any",       // Todas as mensagens (fromMe: true/false)
			"message.ack",       // Confirmações de leitura
			"message.reaction",  // Reações
			"message.edited",    // Mensagens editadas
		},
		"waha_calls": {
			"call.received",
			"call.accepted", 
			"call.rejected",
		},
		"waha_labels": {
			"label.upsert",
			"label.deleted",
			"label.chat.added",
			"label.chat.deleted",
		},
		"waha_groups": {
			"group.v2.join",
			"group.v2.leave",
			"group.v2.update",
			"group.v2.participants",
		},
		"waha_other": {
			"event.response.failed",
		},
	}
}
