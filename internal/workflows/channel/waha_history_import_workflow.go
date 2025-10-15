package channel

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// WAHAHistoryImportWorkflowInput representa a entrada do workflow
type WAHAHistoryImportWorkflowInput struct {
	ChannelID             string `json:"channel_id"`
	SessionID             string `json:"session_id"`
	Strategy              string `json:"strategy"`                // "all", "recent", "custom"
	Limit                 int    `json:"limit"`                   // Mensagens por chat (0 = todas)
	TimeRangeDays         int    `json:"time_range_days"`         // Dias para filtrar mensagens (0 = sem filtro)
	SessionTimeoutMinutes int    `json:"session_timeout_minutes"` // Timeout de inatividade para agrupar sessões (default: 30)
	ProjectID             string `json:"project_id"`
	TenantID              string `json:"tenant_id"`
	UserID                string `json:"user_id"`
}

// WAHAHistoryImportWorkflowResult representa o resultado do workflow
type WAHAHistoryImportWorkflowResult struct {
	ChannelID        string    `json:"channel_id"`
	ChatsProcessed   int       `json:"chats_processed"`
	MessagesImported int       `json:"messages_imported"`
	SessionsCreated  int       `json:"sessions_created"`
	ContactsCreated  int       `json:"contacts_created"`
	Errors           []string  `json:"errors"`
	StartedAt        time.Time `json:"started_at"`
	CompletedAt      time.Time `json:"completed_at"`
	Status           string    `json:"status"` // "completed", "failed", "partial"
}

// WAHAHistoryImportWorkflow é o workflow principal para importação de histórico
// Usa Temporal para garantir durabilidade, retry e observabilidade
func WAHAHistoryImportWorkflow(ctx workflow.Context, input WAHAHistoryImportWorkflowInput) (*WAHAHistoryImportWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)

	// Default session timeout se não especificado
	if input.SessionTimeoutMinutes == 0 {
		input.SessionTimeoutMinutes = 240 // Default: 4 horas (para máxima consolidação)
	}

	logger.Info("========================================")
	logger.Info("🚀 Starting WAHA History Import Workflow")
	logger.Info("========================================")
	logger.Info("Configuration:",
		"channel_id", input.ChannelID,
		"session_id", input.SessionID,
		"strategy", input.Strategy,
		"limit", input.Limit,
		"time_range_days", input.TimeRangeDays,
		"session_timeout_minutes", input.SessionTimeoutMinutes,
		"project_id", input.ProjectID,
		"tenant_id", input.TenantID)
	logger.Info("========================================")

	result := &WAHAHistoryImportWorkflowResult{
		ChannelID: input.ChannelID,
		StartedAt: workflow.Now(ctx),
		Errors:    []string{},
		Status:    "processing",
	}

	// Setup query handler para consultar status durante execução
	err := workflow.SetQueryHandler(ctx, "import-status", func() (*WAHAHistoryImportWorkflowResult, error) {
		return result, nil
	})
	if err != nil {
		return nil, err
	}

	// Configurar retry policy para activities
	retryPolicy := &temporal.RetryPolicy{
		InitialInterval:    time.Second * 5,
		BackoffCoefficient: 2.0,
		MaximumInterval:    time.Minute * 5,
		MaximumAttempts:    3,
	}

	activityOptions := workflow.ActivityOptions{
		StartToCloseTimeout: time.Minute * 10, // Tempo máximo por activity
		RetryPolicy:         retryPolicy,
	}
	ctx = workflow.WithActivityOptions(ctx, activityOptions)

	// STEP 0: Determinar melhor range de tempo baseado no chat mais antigo
	logger.Info("Step 0: Determining optimal import time range")
	var timeRangeResult DetermineTimeRangeActivityResult
	timeRangeInput := DetermineTimeRangeActivityInput{
		ChannelID:     input.ChannelID,
		SessionID:     input.SessionID,
		TimeRangeDays: input.TimeRangeDays,
	}

	err = workflow.ExecuteActivity(ctx, "DetermineImportTimeRangeActivity", timeRangeInput).Get(ctx, &timeRangeResult)
	if err != nil {
		logger.Error("Failed to determine time range, using requested range", "error", err.Error())
		// Fallback: continua com o range solicitado
	} else if timeRangeResult.OptimizedStartDate != nil {
		// Usar data otimizada
		actualDays := int(workflow.Now(ctx).Sub(*timeRangeResult.OptimizedStartDate).Hours() / 24)
		logger.Info("✅ Using optimized time range",
			"requested_days", input.TimeRangeDays,
			"actual_days", actualDays,
			"optimized_start", timeRangeResult.OptimizedStartDate.Format(time.RFC3339))

		// Atualizar TimeRangeDays com o valor otimizado
		input.TimeRangeDays = actualDays
	}

	// STEP 1: Buscar lista de chats do WAHA
	logger.Info("📋 Step 1: Fetching chats from WAHA")
	var chats []ChatInfo
	fetchChatsInput := FetchChatsActivityInput{
		ChannelID: input.ChannelID,
		SessionID: input.SessionID,
		Strategy:  input.Strategy,
	}

	logger.Info("🔄 Executing FetchWAHAChatsActivity...",
		"channel_id", input.ChannelID,
		"session_id", input.SessionID)

	err = workflow.ExecuteActivity(ctx, "FetchWAHAChatsActivity", fetchChatsInput).Get(ctx, &chats)
	if err != nil {
		logger.Error("❌ Failed to fetch chats", "error", err.Error())
		result.Status = "failed"
		result.Errors = append(result.Errors, "Failed to fetch chats: "+err.Error())
		result.CompletedAt = workflow.Now(ctx)
		return result, err
	}

	logger.Info("✅ Chats fetched successfully", "count", len(chats))

	// STEP 2: Processar cada chat em paralelo (com limite de concorrência)
	logger.Info("Step 2: Processing chats")

	// Limitar paralelismo para não sobrecarregar WAHA
	maxConcurrentChats := 5
	chatBatches := batchChats(chats, maxConcurrentChats)

	for batchIndex, batch := range chatBatches {
		logger.Info("Processing chat batch", "batch", batchIndex+1, "total_batches", len(chatBatches), "chats_in_batch", len(batch))

		// Processar batch em paralelo
		futures := []workflow.Future{}
		for _, chat := range batch {
			importInput := ImportChatHistoryActivityInput{
				ChannelID:             input.ChannelID,
				SessionID:             input.SessionID,
				ChatID:                chat.ID,
				ChatName:              chat.Name,
				Limit:                 input.Limit,
				TimeRangeDays:         input.TimeRangeDays,
				SessionTimeoutMinutes: input.SessionTimeoutMinutes,
				ProjectID:             input.ProjectID,
				TenantID:              input.TenantID,
			}

			future := workflow.ExecuteActivity(ctx, "ImportChatHistoryActivity", importInput)
			futures = append(futures, future)
		}

		// Aguardar todas as activities do batch
		for i, future := range futures {
			var chatResult ImportChatHistoryActivityResult
			err := future.Get(ctx, &chatResult)
			if err != nil {
				errMsg := "Failed to import chat " + batch[i].ID + ": " + err.Error()
				logger.Warn(errMsg)
				result.Errors = append(result.Errors, errMsg)
				continue
			}

			// Acumular resultados
			result.ChatsProcessed++
			result.MessagesImported += chatResult.MessagesImported
			result.SessionsCreated += chatResult.SessionsCreated
			result.ContactsCreated += chatResult.ContactsCreated
		}
	}

	// STEP 2.5: Consolidar sessions criadas durante import (pós-processamento determinístico)
	logger.Info("Step 2.5: Consolidating history sessions")
	consolidateInput := ConsolidateHistorySessionsActivityInput{
		ChannelID:             input.ChannelID,
		SessionTimeoutMinutes: input.SessionTimeoutMinutes,
	}

	var consolidateResult ConsolidateHistorySessionsActivityResult
	err = workflow.ExecuteActivity(ctx, "ConsolidateHistorySessionsActivity", consolidateInput).Get(ctx, &consolidateResult)
	if err != nil {
		logger.Warn("Failed to consolidate sessions", "error", err.Error())
		result.Errors = append(result.Errors, "Failed to consolidate sessions: "+err.Error())
	} else {
		// Atualizar contagem de sessions (após consolidação)
		result.SessionsCreated = consolidateResult.SessionsAfter
		logger.Info("Sessions consolidated",
			"sessions_before", consolidateResult.SessionsBefore,
			"sessions_after", consolidateResult.SessionsAfter,
			"sessions_deleted", consolidateResult.SessionsDeleted,
			"messages_updated", consolidateResult.MessagesUpdated)
	}

	// STEP 3: Processar webhooks buffered (SAGA Pattern compensation)
	logger.Info("Step 3: Processing buffered webhooks")
	importDuration := workflow.Now(ctx).Sub(result.StartedAt)
	processBufferedInput := ProcessBufferedWebhooksActivityInput{
		ChannelID:             input.ChannelID,
		TotalMessagesImported: result.MessagesImported,
		ImportDurationSeconds: int64(importDuration.Seconds()),
	}

	var processBufferedResult ProcessBufferedWebhooksActivityResult
	err = workflow.ExecuteActivity(ctx, "ProcessBufferedWebhooksActivity", processBufferedInput).Get(ctx, &processBufferedResult)
	if err != nil {
		logger.Warn("Failed to process buffered webhooks", "error", err.Error())
		result.Errors = append(result.Errors, "Failed to process buffered webhooks: "+err.Error())
	} else {
		logger.Info("Buffered webhooks processed",
			"webhooks_count", processBufferedResult.WebhooksProcessed,
			"errors_count", len(processBufferedResult.Errors))
	}

	// Finalizar
	result.CompletedAt = workflow.Now(ctx)
	duration := result.CompletedAt.Sub(result.StartedAt)

	if len(result.Errors) == 0 {
		result.Status = "completed"
	} else if result.ChatsProcessed > 0 {
		result.Status = "partial" // Alguns chats importados, mas houve erros
	} else {
		result.Status = "failed"
	}

	// Calcular estatísticas
	avgSessionsPerChat := float64(0)
	if result.ChatsProcessed > 0 {
		avgSessionsPerChat = float64(result.SessionsCreated) / float64(result.ChatsProcessed)
	}

	avgMessagesPerSession := float64(0)
	if result.SessionsCreated > 0 {
		avgMessagesPerSession = float64(result.MessagesImported) / float64(result.SessionsCreated)
	}

	// Log do relatório detalhado final
	logger.Info("========================================")
	logger.Info("✅ WAHA History Import COMPLETED")
	logger.Info("========================================")
	logger.Info("Status:", "value", result.Status)
	logger.Info("Duration:", "value", duration.String())
	logger.Info("----------------------------------------")
	logger.Info("📊 Statistics:")
	logger.Info("  • Chats Processed:", "count", result.ChatsProcessed)
	logger.Info("  • Messages Imported:", "count", result.MessagesImported)
	logger.Info("  • Sessions Created:", "count", result.SessionsCreated)
	logger.Info("  • Contacts Created:", "count", result.ContactsCreated)
	logger.Info("----------------------------------------")
	logger.Info("📈 Averages:")
	logger.Info("  • Sessions per Chat:", "avg", avgSessionsPerChat)
	logger.Info("  • Messages per Session:", "avg", avgMessagesPerSession)
	logger.Info("----------------------------------------")
	logger.Info("⚙️  Configuration Used:")
	logger.Info("  • Session Timeout:", "minutes", input.SessionTimeoutMinutes)
	logger.Info("  • Limit per Chat:", "value", input.Limit)
	logger.Info("  • Time Range Days:", "value", input.TimeRangeDays)

	if len(result.Errors) > 0 {
		logger.Info("----------------------------------------")
		logger.Warn("⚠️  Errors Encountered:", "count", len(result.Errors))
		for i, err := range result.Errors {
			logger.Warn("  Error", "index", i+1, "message", err)
		}
	}

	logger.Info("========================================")

	return result, nil
}

// batchChats divide chats em lotes para processamento paralelo controlado
func batchChats(chats []ChatInfo, batchSize int) [][]ChatInfo {
	batches := [][]ChatInfo{}
	for i := 0; i < len(chats); i += batchSize {
		end := i + batchSize
		if end > len(chats) {
			end = len(chats)
		}
		batches = append(batches, chats[i:end])
	}
	return batches
}

// ChatInfo representa informações básicas de um chat
type ChatInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// DetermineTimeRangeActivityInput representa entrada para determinar range de tempo
type DetermineTimeRangeActivityInput struct {
	ChannelID     string `json:"channel_id"` // ID do canal (para buscar config WAHA)
	SessionID     string `json:"session_id"`
	TimeRangeDays int    `json:"time_range_days"` // Dias solicitados pelo usuário
}

// DetermineTimeRangeActivityResult representa resultado da otimização de range
type DetermineTimeRangeActivityResult struct {
	TimeRangeDays      int        `json:"time_range_days"`       // Dias originalmente solicitados
	OptimizedStartDate *time.Time `json:"optimized_start_date"`  // Data otimizada (chat mais antigo - 1 dia), ou nil se usar default
	ActualDays         int        `json:"actual_days,omitempty"` // Dias reais que serão importados
}

// FetchChatsActivityInput representa a entrada para buscar chats
type FetchChatsActivityInput struct {
	ChannelID string `json:"channel_id"` // ID do canal (para buscar config WAHA)
	SessionID string `json:"session_id"`
	Strategy  string `json:"strategy"`
}

// ImportChatHistoryActivityInput representa a entrada para importar chat
type ImportChatHistoryActivityInput struct {
	ChannelID             string `json:"channel_id"`
	SessionID             string `json:"session_id"`
	ChatID                string `json:"chat_id"`
	ChatName              string `json:"chat_name"`
	Limit                 int    `json:"limit"`
	TimeRangeDays         int    `json:"time_range_days"`         // Dias para filtrar mensagens (0 = sem filtro)
	SessionTimeoutMinutes int    `json:"session_timeout_minutes"` // Timeout de inatividade para agrupar sessões
	ProjectID             string `json:"project_id"`
	TenantID              string `json:"tenant_id"`
}

// ImportChatHistoryActivityResult representa o resultado da importação de um chat
type ImportChatHistoryActivityResult struct {
	ChatID           string `json:"chat_id"`
	MessagesImported int    `json:"messages_imported"`
	SessionsCreated  int    `json:"sessions_created"`
	ContactsCreated  int    `json:"contacts_created"`
}

// MarkImportCompletedActivityInput representa a entrada para marcar importação completa
type MarkImportCompletedActivityInput struct {
	ChannelID string `json:"channel_id"`
}

// ProcessBufferedWebhooksActivityInput representa a entrada para processar webhooks buffered
type ProcessBufferedWebhooksActivityInput struct {
	ChannelID             string `json:"channel_id"`
	TotalMessagesImported int    `json:"total_messages_imported"`
	ImportDurationSeconds int64  `json:"import_duration_seconds"`
}

// ProcessBufferedWebhooksActivityResult representa o resultado do processamento de webhooks
type ProcessBufferedWebhooksActivityResult struct {
	ChannelID         string   `json:"channel_id"`
	WebhooksProcessed int      `json:"webhooks_processed"`
	Errors            []string `json:"errors"`
}

// ConsolidateHistorySessionsActivityInput representa entrada para consolidação de sessions
type ConsolidateHistorySessionsActivityInput struct {
	ChannelID             string `json:"channel_id"`
	SessionTimeoutMinutes int    `json:"session_timeout_minutes"`
}

// ConsolidateHistorySessionsActivityResult representa resultado da consolidação
type ConsolidateHistorySessionsActivityResult struct {
	ChannelID       string `json:"channel_id"`
	SessionsBefore  int    `json:"sessions_before"`
	SessionsAfter   int    `json:"sessions_after"`
	SessionsDeleted int64  `json:"sessions_deleted"`
	MessagesUpdated int64  `json:"messages_updated"`
}
