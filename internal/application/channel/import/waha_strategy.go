package importpkg

import (
	"context"
	"fmt"
	"time"

	"github.com/ventros/crm/infrastructure/channels/waha"
	"github.com/ventros/crm/internal/domain/crm/channel"
	"go.temporal.io/sdk/client"
	"go.uber.org/zap"
)

// WAHAImportStrategy implementa importação de histórico para canais WAHA
// Valida que a sessão WAHA está rodando e inicia workflow Temporal
type WAHAImportStrategy struct {
	temporalClient client.Client
	logger         *zap.Logger
}

// NewWAHAImportStrategy cria uma nova instância da strategy WAHA
func NewWAHAImportStrategy(temporalClient client.Client, logger *zap.Logger) *WAHAImportStrategy {
	return &WAHAImportStrategy{
		temporalClient: temporalClient,
		logger:         logger,
	}
}

// CanImport verifica se o canal WAHA pode realizar importação de histórico
// Valida que:
// 1. Canal é do tipo WAHA
// 2. Canal está ativo
// 3. Configuração WAHA está completa
// 4. Sessão WAHA está WORKING
// 5. Não há importação em andamento
func (s *WAHAImportStrategy) CanImport(ctx context.Context, ch *channel.Channel, strategy string) error {
	if ch.Type != channel.TypeWAHA {
		return fmt.Errorf("channel type is not WAHA: %s", ch.Type)
	}

	// Verificar se canal está ativo
	if ch.Status != channel.StatusActive {
		return fmt.Errorf("channel is not active: status=%s", ch.Status)
	}

	// Verificar configuração WAHA
	config, err := ch.GetWAHAConfig()
	if err != nil {
		return fmt.Errorf("failed to get WAHA config: %w", err)
	}

	if config.BaseURL == "" {
		return fmt.Errorf("WAHA base_url is required")
	}

	if config.Auth.APIKey == "" && config.Auth.Token == "" {
		return fmt.Errorf("WAHA authentication is required")
	}

	if config.SessionID == "" {
		return fmt.Errorf("WAHA session_id is required")
	}

	// Verificar status da sessão WAHA
	authToken := config.Auth.APIKey
	if authToken == "" {
		authToken = config.Auth.Token
	}

	wahaClient := waha.NewWAHAClient(config.BaseURL, authToken, s.logger)
	isHealthy, status, err := wahaClient.HealthCheck(ctx, config.SessionID)
	if err != nil {
		return fmt.Errorf("WAHA health check failed: %w", err)
	}

	if !isHealthy || status != "WORKING" {
		return fmt.Errorf("WAHA session is not working: status=%s", status)
	}

	// Verificar se já há importação em andamento
	if ch.HistoryImportStatus == channel.HistoryImportInProgress {
		return fmt.Errorf("import already in progress")
	}

	// Validar strategy
	validStrategies := map[string]bool{
		"time_range": true,
		"full":       true,
		"recent":     true,
	}
	if !validStrategies[strategy] {
		return fmt.Errorf("invalid import strategy: %s", strategy)
	}

	s.logger.Info("WAHA channel pre-import checks passed",
		zap.String("channel_id", ch.ID.String()),
		zap.String("session_id", config.SessionID),
		zap.String("strategy", strategy))

	return nil
}

// Import inicia o workflow Temporal de importação de histórico WAHA
// Retorna o workflowID para tracking
func (s *WAHAImportStrategy) Import(ctx context.Context, ch *channel.Channel, params ImportParams) (string, error) {
	config, err := ch.GetWAHAConfig()
	if err != nil {
		return "", fmt.Errorf("failed to get WAHA config: %w", err)
	}

	// Determinar session timeout (default 30 minutos se não configurado no canal)
	sessionTimeoutMinutes := 30 // Default
	if ch.DefaultSessionTimeoutMinutes > 0 {
		sessionTimeoutMinutes = ch.DefaultSessionTimeoutMinutes
	}

	// Montar input do workflow
	workflowInput := map[string]interface{}{
		"channel_id":              ch.ID.String(),
		"session_id":              config.SessionID,
		"strategy":                params.Strategy,
		"limit":                   params.Limit,
		"time_range_days":         params.TimeRangeDays,
		"session_timeout_minutes": sessionTimeoutMinutes,
		"project_id":              ch.ProjectID.String(),
		"tenant_id":               ch.TenantID,
		"user_id":                 params.UserID,
	}

	// Gerar workflowID único
	workflowID := fmt.Sprintf("waha-import-%s", ch.ID.String())

	// Iniciar workflow Temporal
	workflowOptions := client.StartWorkflowOptions{
		ID:        workflowID,
		TaskQueue: "waha-imports", // Must match worker task queue in waha_import_worker.go:46
	}

	_, err = s.temporalClient.ExecuteWorkflow(ctx, workflowOptions, "WAHAHistoryImportWorkflow", workflowInput)
	if err != nil {
		s.logger.Error("Failed to start WAHA import workflow",
			zap.String("channel_id", ch.ID.String()),
			zap.String("workflow_id", workflowID),
			zap.Error(err))
		return "", fmt.Errorf("failed to start workflow: %w", err)
	}

	s.logger.Info("WAHA import workflow started successfully",
		zap.String("channel_id", ch.ID.String()),
		zap.String("workflow_id", workflowID),
		zap.String("correlation_id", params.CorrelationID),
		zap.String("strategy", params.Strategy),
		zap.Int("time_range_days", params.TimeRangeDays))

	return workflowID, nil
}

// GetImportStatus consulta o status atual da importação via Temporal
func (s *WAHAImportStrategy) GetImportStatus(ctx context.Context, ch *channel.Channel) (*ImportStatusInfo, error) {
	workflowID := fmt.Sprintf("waha-import-%s", ch.ID.String())

	// Consultar workflow no Temporal
	run := s.temporalClient.GetWorkflow(ctx, workflowID, "")

	var result map[string]interface{}
	err := run.Get(ctx, &result)

	// Se workflow ainda está rodando, err será do tipo WorkflowExecutionError
	if err != nil {
		// Workflow ainda em execução ou erro
		s.logger.Debug("Workflow status check",
			zap.String("workflow_id", workflowID),
			zap.Error(err))
	}

	// Montar response com dados do channel
	status := &ImportStatusInfo{
		Status:           string(ch.HistoryImportStatus),
		WorkflowID:       workflowID,
		CorrelationID:    ch.HistoryImportCorrelationID,
		MessagesImported: ch.HistoryImportMessagesCount,
	}

	// Converter Stats para map se existir
	if ch.HistoryImportStats != nil {
		status.Stats = map[string]interface{}{
			"total":      ch.HistoryImportStats.Total,
			"processed":  ch.HistoryImportStats.Processed,
			"failed":     ch.HistoryImportStats.Failed,
			"started_at": ch.HistoryImportStats.StartedAt,
		}
		if ch.HistoryImportStats.EndedAt != nil {
			status.Stats["ended_at"] = ch.HistoryImportStats.EndedAt
		}
	}

	// Converter LastImportDate para string RFC3339 se existir
	if ch.LastImportDate != nil {
		formatted := ch.LastImportDate.Format(time.RFC3339)
		status.StartedAt = &formatted
	}

	return status, nil
}

// CancelImport cancela uma importação em andamento
func (s *WAHAImportStrategy) CancelImport(ctx context.Context, ch *channel.Channel, reason string) error {
	workflowID := fmt.Sprintf("waha-import-%s", ch.ID.String())

	err := s.temporalClient.CancelWorkflow(ctx, workflowID, "")
	if err != nil {
		s.logger.Error("Failed to cancel WAHA import workflow",
			zap.String("channel_id", ch.ID.String()),
			zap.String("workflow_id", workflowID),
			zap.String("reason", reason),
			zap.Error(err))
		return fmt.Errorf("failed to cancel workflow: %w", err)
	}

	s.logger.Info("WAHA import workflow cancelled",
		zap.String("channel_id", ch.ID.String()),
		zap.String("workflow_id", workflowID),
		zap.String("reason", reason))

	return nil
}
