package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/caloi/ventros-crm/internal/application/pipeline"
	domainPipeline "github.com/caloi/ventros-crm/internal/domain/pipeline"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ScheduledRulesWorker executa regras agendadas (cron, recorrentes)
type ScheduledRulesWorker struct {
	db             *gorm.DB
	followUpEngine *pipeline.AutomationEngine
	pollInterval   time.Duration
	logger         pipeline.Logger
	stopChan       chan struct{}
}

// NewScheduledRulesWorker cria novo worker
func NewScheduledRulesWorker(
	db *gorm.DB,
	followUpEngine *pipeline.AutomationEngine,
	pollInterval time.Duration,
	logger pipeline.Logger,
) *ScheduledRulesWorker {
	if pollInterval == 0 {
		pollInterval = 1 * time.Minute // default: 1 minuto
	}

	if logger == nil {
		logger = &defaultLogger{}
	}

	return &ScheduledRulesWorker{
		db:             db,
		followUpEngine: followUpEngine,
		pollInterval:   pollInterval,
		logger:         logger,
		stopChan:       make(chan struct{}),
	}
}

// Start inicia o worker
func (w *ScheduledRulesWorker) Start(ctx context.Context) {
	w.logger.Info("starting scheduled rules worker", "pollInterval", w.pollInterval)

	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()

	// Primeira execução imediata
	w.processScheduledRules(ctx)

	for {
		select {
		case <-ticker.C:
			w.processScheduledRules(ctx)

		case <-w.stopChan:
			w.logger.Info("scheduled rules worker stopped")
			return

		case <-ctx.Done():
			w.logger.Info("scheduled rules worker context cancelled")
			return
		}
	}
}

// Stop para o worker
func (w *ScheduledRulesWorker) Stop() {
	close(w.stopChan)
}

// processScheduledRules busca e executa regras agendadas prontas
func (w *ScheduledRulesWorker) processScheduledRules(ctx context.Context) {
	w.logger.Debug("processing scheduled follow-up rules")

	now := time.Now()

	// Busca regras agendadas prontas para executar
	// next_execution <= now AND enabled = true AND trigger = 'scheduled'
	var rules []scheduledRuleRow
	err := w.db.WithContext(ctx).
		Table("automation_rules").
		Where("trigger = ?", "scheduled").
		Where("enabled = ?", true).
		Where("next_execution IS NOT NULL").
		Where("next_execution <= ?", now).
		Order("next_execution ASC, priority ASC").
		Find(&rules).Error

	if err != nil {
		w.logger.Error("failed to fetch scheduled rules", "error", err)
		return
	}

	if len(rules) == 0 {
		w.logger.Debug("no scheduled rules ready to execute")
		return
	}

	w.logger.Info("found scheduled rules ready to execute", "count", len(rules))

	// Executa cada regra
	for _, ruleRow := range rules {
		if err := w.executeScheduledRule(ctx, &ruleRow, now); err != nil {
			w.logger.Error("failed to execute scheduled rule",
				"ruleID", ruleRow.ID,
				"name", ruleRow.Name,
				"error", err,
			)
			// Continua para próximas regras
		}
	}
}

// executeScheduledRule executa uma regra agendada individual
func (w *ScheduledRulesWorker) executeScheduledRule(
	ctx context.Context,
	ruleRow *scheduledRuleRow,
	executedAt time.Time,
) error {
	w.logger.Info("executing scheduled rule",
		"ruleID", ruleRow.ID,
		"name", ruleRow.Name,
		"pipelineID", ruleRow.PipelineID,
	)

	// Reconstrói o ScheduledRuleConfig do JSON
	scheduledRule, err := w.reconstructScheduledRule(ruleRow)
	if err != nil {
		return fmt.Errorf("failed to reconstruct scheduled rule: %w", err)
	}

	// Verifica se realmente está pronto (double-check)
	if !scheduledRule.IsReadyToExecute(executedAt) {
		w.logger.Debug("rule not ready to execute (double-check failed)",
			"ruleID", ruleRow.ID,
		)
		return nil
	}

	// Prepara contexto de avaliação
	evalContext := map[string]interface{}{
		"executed_at": executedAt,
		"tenant_id":   ruleRow.TenantID,
		"pipeline_id": ruleRow.PipelineID,
	}

	// Prepara contexto de ação
	pipelineID := ruleRow.GetPipelineID()
	actionCtx := pipeline.ActionContext{
		PipelineID: &pipelineID,
		TenantID:   ruleRow.TenantID,
		Trigger:    domainPipeline.TriggerScheduled,
		Metadata:   evalContext,
	}

	// Executa via engine
	err = w.followUpEngine.EvaluateAndExecute(
		ctx,
		ruleRow.GetPipelineID(),
		domainPipeline.TriggerScheduled,
		evalContext,
		actionCtx,
	)

	if err != nil {
		w.logger.Error("scheduled rule execution failed",
			"ruleID", ruleRow.ID,
			"error", err,
		)
		// Não atualiza next_execution em caso de erro
		// Vai tentar novamente no próximo ciclo
		return err
	}

	// Marca como executada e calcula próxima execução
	scheduledRule.MarkExecuted(executedAt)

	// Atualiza no banco
	updates := map[string]interface{}{
		"last_executed": executedAt,
	}

	if scheduledRule.NextExecution != nil {
		updates["next_execution"] = *scheduledRule.NextExecution
		w.logger.Info("scheduled rule executed, next execution at",
			"ruleID", ruleRow.ID,
			"nextExecution", *scheduledRule.NextExecution,
		)
	} else {
		// Não há próxima execução (ex: schedule type = once)
		updates["next_execution"] = nil
		updates["enabled"] = false // desativa regra
		w.logger.Info("scheduled rule executed, no more executions (disabled)",
			"ruleID", ruleRow.ID,
		)
	}

	if err := w.db.WithContext(ctx).Table("automation_rules").
		Where("id = ?", ruleRow.ID).
		Updates(updates).Error; err != nil {
		w.logger.Error("failed to update scheduled rule execution status",
			"ruleID", ruleRow.ID,
			"error", err,
		)
		return err
	}

	return nil
}

// reconstructScheduledRule reconstrói ScheduledAutomationRule do banco
func (w *ScheduledRulesWorker) reconstructScheduledRule(row *scheduledRuleRow) (*domainPipeline.ScheduledAutomationRule, error) {
	// Deserializa schedule config
	var scheduleConfig domainPipeline.ScheduledRuleConfig
	if err := json.Unmarshal(row.Schedule, &scheduleConfig); err != nil {
		return nil, fmt.Errorf("failed to unmarshal schedule config: %w", err)
	}

	// Reconstrói regra base
	pipelineID := row.GetPipelineID()
	rule := domainPipeline.ReconstructAutomation(
		row.GetID(),
		domainPipeline.AutomationTypeScheduled, // Tipo da automação
		&pipelineID,
		row.TenantID,
		row.Name,
		row.Description,
		domainPipeline.AutomationTrigger(row.Trigger),
		nil, // conditions não necessárias para execução
		nil, // actions não necessárias aqui (engine processa)
		row.Priority,
		row.Enabled,
		row.CreatedAt,
		row.UpdatedAt,
	)

	// Cria ScheduledAutomationRule
	return domainPipeline.ReconstructScheduledAutomationRule(
		rule,
		scheduleConfig,
		row.LastExecuted,
		row.NextExecution,
	), nil
}

// scheduledRuleRow representa uma linha do banco para regras agendadas
type scheduledRuleRow struct {
	ID            string     `gorm:"column:id"`
	PipelineID    string     `gorm:"column:pipeline_id"`
	TenantID      string     `gorm:"column:tenant_id"`
	Name          string     `gorm:"column:name"`
	Description   string     `gorm:"column:description"`
	Trigger       string     `gorm:"column:trigger"`
	Priority      int        `gorm:"column:priority"`
	Enabled       bool       `gorm:"column:enabled"`
	Schedule      []byte     `gorm:"column:schedule"`
	LastExecuted  *time.Time `gorm:"column:last_executed"`
	NextExecution *time.Time `gorm:"column:next_execution"`
	CreatedAt     time.Time  `gorm:"column:created_at"`
	UpdatedAt     time.Time  `gorm:"column:updated_at"`
}

func (r *scheduledRuleRow) GetID() uuid.UUID {
	id, _ := uuid.Parse(r.ID)
	return id
}

func (r *scheduledRuleRow) GetPipelineID() uuid.UUID {
	id, _ := uuid.Parse(r.PipelineID)
	return id
}

// defaultLogger implementação simples
type defaultLogger struct{}

func (l *defaultLogger) Info(msg string, args ...interface{}) {
	log.Printf("[INFO] "+msg, args...)
}

func (l *defaultLogger) Error(msg string, args ...interface{}) {
	log.Printf("[ERROR] "+msg, args...)
}

func (l *defaultLogger) Debug(msg string, args ...interface{}) {
	log.Printf("[DEBUG] "+msg, args...)
}
