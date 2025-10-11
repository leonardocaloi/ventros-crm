package pipeline

import (
	"context"
	"fmt"
	"time"

	"github.com/caloi/ventros-crm/internal/domain/crm/pipeline"
	"github.com/caloi/ventros-crm/internal/domain/crm/session"
	"github.com/google/uuid"
)

// AutomationIntegration integra Follow-up Rules com eventos do sistema
type AutomationIntegration struct {
	engine       *AutomationEngine
	sessionRepo  session.Repository
	pipelineRepo pipeline.Repository
	logger       Logger
}

// NewAutomationIntegration cria nova instância da integração
func NewAutomationIntegration(
	engine *AutomationEngine,
	sessionRepo session.Repository,
	pipelineRepo pipeline.Repository,
	logger Logger,
) *AutomationIntegration {
	return &AutomationIntegration{
		engine:       engine,
		sessionRepo:  sessionRepo,
		pipelineRepo: pipelineRepo,
		logger:       logger,
	}
}

// OnSessionEnded processa evento de sessão encerrada
func (i *AutomationIntegration) OnSessionEnded(ctx context.Context, sessionID uuid.UUID) error {
	// Busca sessão
	sess, err := i.sessionRepo.FindByID(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("failed to get session: %w", err)
	}

	// Verifica se sessão tem pipeline associado
	if sess.PipelineID() == nil {
		i.logger.Debug("session has no pipeline, skipping follow-up rules", "sessionID", sessionID)
		return nil
	}

	pipelineID := *sess.PipelineID()

	// Prepara metadata para avaliação
	metadata := map[string]interface{}{
		"session_duration_minutes": time.Since(sess.StartedAt()).Minutes(),
		"message_count":            sess.MessageCount(),
		"resolved":                 sess.IsResolved(),
	}

	// Adiciona agent_id se houver agentes
	agentIDs := sess.AgentIDs()
	if len(agentIDs) > 0 {
		metadata["agent_id"] = agentIDs[0].String()
	}

	// Dispara engine
	return i.engine.ProcessSessionEvent(
		ctx,
		pipelineID,
		pipeline.TriggerSessionEnded,
		sessionID,
		sess.ContactID(),
		uuid.Nil, // channelID não disponível no domain
		sess.TenantID(),
		metadata,
	)
}

// OnSessionTimeout processa evento de timeout de sessão
func (i *AutomationIntegration) OnSessionTimeout(ctx context.Context, sessionID uuid.UUID) error {
	sess, err := i.sessionRepo.FindByID(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("failed to get session: %w", err)
	}

	if sess.PipelineID() == nil {
		return nil
	}

	pipelineID := *sess.PipelineID()

	// LastActivityAt é o equivalente de LastMessageAt
	metadata := map[string]interface{}{
		"session_duration_minutes": time.Since(sess.StartedAt()).Minutes(),
		"last_message_at":          sess.LastActivityAt(),
		"hours_since_last_message": time.Since(sess.LastActivityAt()).Hours(),
		"message_count":            sess.MessageCount(),
	}

	return i.engine.ProcessSessionEvent(
		ctx,
		pipelineID,
		pipeline.TriggerSessionTimeout,
		sessionID,
		sess.ContactID(),
		uuid.Nil, // channelID não disponível no domain
		sess.TenantID(),
		metadata,
	)
}

// OnSessionResolved processa evento de sessão resolvida
func (i *AutomationIntegration) OnSessionResolved(ctx context.Context, sessionID uuid.UUID) error {
	sess, err := i.sessionRepo.FindByID(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("failed to get session: %w", err)
	}

	if sess.PipelineID() == nil {
		return nil
	}

	pipelineID := *sess.PipelineID()

	metadata := map[string]interface{}{
		"session_duration_minutes": time.Since(sess.StartedAt()).Minutes(),
		"message_count":            sess.MessageCount(),
	}

	// Adiciona agent_id se houver agentes
	agentIDs := sess.AgentIDs()
	if len(agentIDs) > 0 {
		metadata["agent_id"] = agentIDs[0].String()
	}

	return i.engine.ProcessSessionEvent(
		ctx,
		pipelineID,
		pipeline.TriggerSessionResolved,
		sessionID,
		sess.ContactID(),
		uuid.Nil, // channelID não disponível no domain
		sess.TenantID(),
		metadata,
	)
}

// OnNoResponse processa trigger de ausência de resposta
func (i *AutomationIntegration) OnNoResponse(
	ctx context.Context,
	sessionID uuid.UUID,
	hoursSinceLastMessage float64,
) error {
	sess, err := i.sessionRepo.FindByID(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("failed to get session: %w", err)
	}

	if sess.PipelineID() == nil {
		return nil
	}

	pipelineID := *sess.PipelineID()

	metadata := map[string]interface{}{
		"hours_since_last_message": hoursSinceLastMessage,
		"last_message_at":          sess.LastActivityAt(),
		"message_count":            sess.MessageCount(),
	}

	return i.engine.ProcessSessionEvent(
		ctx,
		pipelineID,
		pipeline.TriggerNoResponse,
		sessionID,
		sess.ContactID(),
		uuid.Nil, // channelID não disponível no domain
		sess.TenantID(),
		metadata,
	)
}

// OnMessageReceived processa trigger de mensagem recebida
func (i *AutomationIntegration) OnMessageReceived(
	ctx context.Context,
	sessionID uuid.UUID,
	messageID uuid.UUID,
) error {
	sess, err := i.sessionRepo.FindByID(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("failed to get session: %w", err)
	}

	if sess.PipelineID() == nil {
		return nil
	}

	pipelineID := *sess.PipelineID()

	metadata := map[string]interface{}{
		"message_id":    messageID.String(),
		"message_count": sess.MessageCount(),
	}

	return i.engine.ProcessSessionEvent(
		ctx,
		pipelineID,
		pipeline.TriggerMessageReceived,
		sessionID,
		sess.ContactID(),
		uuid.Nil, // channelID não disponível no domain
		sess.TenantID(),
		metadata,
	)
}

// OnStatusChanged processa trigger de mudança de status
func (i *AutomationIntegration) OnStatusChanged(
	ctx context.Context,
	contactID uuid.UUID,
	pipelineID uuid.UUID,
	oldStatusID *uuid.UUID,
	newStatusID uuid.UUID,
	tenantID string,
) error {
	metadata := map[string]interface{}{
		"old_status_id": oldStatusID,
		"new_status_id": newStatusID.String(),
	}

	return i.engine.ProcessContactEvent(
		ctx,
		pipelineID,
		pipeline.TriggerStatusChanged,
		contactID,
		tenantID,
		metadata,
	)
}

// ScheduleDelayedCheck agenda checagem de follow-up após delay
// Útil para triggers como "after.delay" ou "scheduled"
func (i *AutomationIntegration) ScheduleDelayedCheck(
	ctx context.Context,
	pipelineID uuid.UUID,
	contactID uuid.UUID,
	delayMinutes int,
	metadata map[string]interface{},
) error {
	// TODO: Integrar com Temporal para agendamento real
	i.logger.Info("scheduling delayed follow-up check",
		"pipelineID", pipelineID,
		"contactID", contactID,
		"delayMinutes", delayMinutes,
	)

	// Placeholder para integração futura
	return nil
}

// ProcessScheduledRules processa regras agendadas
// Chamado por worker/cron job
func (i *AutomationIntegration) ProcessScheduledRules(ctx context.Context) error {
	// TODO: Implementar quando houver sistema de agendamento
	i.logger.Info("processing scheduled follow-up rules")
	return nil
}
