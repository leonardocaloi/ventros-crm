package message

import (
	"context"
	"fmt"
	"time"

	"github.com/caloi/ventros-crm/internal/domain/crm/message_group"
	"go.uber.org/zap"
)

// MessageGroupWorker processa grupos de mensagens expirados periodicamente
type MessageGroupWorker struct {
	logger            *zap.Logger
	debouncerService  *MessageDebouncerService
	enrichmentService *MessageEnrichmentService
	aiAgentService    *AIAgentService
	messageGroupRepo  message_group.Repository
	tickerInterval    time.Duration
	batchSize         int
	stopChan          chan struct{}
}

// NewMessageGroupWorker cria um novo worker de processamento de grupos
func NewMessageGroupWorker(
	logger *zap.Logger,
	debouncerService *MessageDebouncerService,
	enrichmentService *MessageEnrichmentService,
	aiAgentService *AIAgentService,
	messageGroupRepo message_group.Repository,
) *MessageGroupWorker {
	return &MessageGroupWorker{
		logger:            logger,
		debouncerService:  debouncerService,
		enrichmentService: enrichmentService,
		aiAgentService:    aiAgentService,
		messageGroupRepo:  messageGroupRepo,
		tickerInterval:    5 * time.Second, // Processa a cada 5 segundos
		batchSize:         100,             // Processa até 100 grupos por vez
		stopChan:          make(chan struct{}),
	}
}

// Start inicia o worker
func (w *MessageGroupWorker) Start(ctx context.Context) error {
	w.logger.Info("Starting message group worker",
		zap.Duration("interval", w.tickerInterval),
		zap.Int("batch_size", w.batchSize))

	ticker := time.NewTicker(w.tickerInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("Message group worker stopped (context done)")
			return ctx.Err()

		case <-w.stopChan:
			w.logger.Info("Message group worker stopped (stop signal)")
			return nil

		case <-ticker.C:
			if err := w.processExpiredGroups(ctx); err != nil {
				w.logger.Error("Failed to process expired groups",
					zap.Error(err))
				// Continua processando mesmo se houver erro
			}
		}
	}
}

// Stop para o worker
func (w *MessageGroupWorker) Stop() {
	close(w.stopChan)
}

// processExpiredGroups processa grupos que expiraram
func (w *MessageGroupWorker) processExpiredGroups(ctx context.Context) error {
	// 1. Buscar grupos expirados
	expiredGroups, err := w.messageGroupRepo.FindExpired(ctx, w.batchSize)
	if err != nil {
		return fmt.Errorf("failed to find expired groups: %w", err)
	}

	if len(expiredGroups) == 0 {
		return nil // Nenhum grupo expirado
	}

	w.logger.Info("Processing expired message groups",
		zap.Int("count", len(expiredGroups)))

	// 2. Processar cada grupo
	successCount := 0
	errorCount := 0

	for _, group := range expiredGroups {
		if err := w.processGroup(ctx, group); err != nil {
			w.logger.Error("Failed to process group",
				zap.Error(err),
				zap.String("group_id", group.ID().String()))
			errorCount++
			continue
		}
		successCount++
	}

	w.logger.Info("Expired groups processed",
		zap.Int("success", successCount),
		zap.Int("errors", errorCount),
		zap.Int("total", len(expiredGroups)))

	return nil
}

// processGroup processa um grupo individual seguindo o fluxo completo
func (w *MessageGroupWorker) processGroup(ctx context.Context, group *message_group.MessageGroup) error {
	w.logger.Info("Processing message group",
		zap.String("group_id", group.ID().String()),
		zap.Int("message_count", group.MessageCount()),
		zap.String("status", string(group.Status())))

	// 1. Marcar como processando
	if err := group.MarkAsProcessing(); err != nil {
		return fmt.Errorf("failed to mark as processing: %w", err)
	}

	if err := w.messageGroupRepo.Save(ctx, group); err != nil {
		return fmt.Errorf("failed to save group status: %w", err)
	}

	// 2. Processar enriquecimentos (transcrição, OCR, etc)
	if err := w.enrichmentService.ProcessGroupEnrichments(ctx, group); err != nil {
		w.logger.Error("Failed to process enrichments",
			zap.Error(err),
			zap.String("group_id", group.ID().String()))
		// Continua mesmo se enrichment falhar - algumas mensagens podem não ter mídia
	}

	// 3. Aguardar enriquecimentos completarem (polling simples)
	// TODO: Substituir por event-driven quando workers assíncronos estiverem prontos
	if err := w.waitForEnrichments(ctx, group); err != nil {
		w.logger.Warn("Failed to wait for enrichments, continuing anyway",
			zap.Error(err),
			zap.String("group_id", group.ID().String()))
	}

	// 4. Enviar para AI Agent (concatena tudo)
	if err := w.aiAgentService.ProcessCompletedGroup(ctx, group); err != nil {
		return fmt.Errorf("failed to send to AI agent: %w", err)
	}

	w.logger.Info("Message group processed successfully",
		zap.String("group_id", group.ID().String()),
		zap.Int("message_count", group.MessageCount()))

	return nil
}

// waitForEnrichments aguarda enriquecimentos completarem (polling simples)
// TODO: Substituir por event-driven quando workers assíncronos estiverem prontos
func (w *MessageGroupWorker) waitForEnrichments(ctx context.Context, group *message_group.MessageGroup) error {
	maxWaitTime := 30 * time.Second
	pollInterval := 500 * time.Millisecond
	deadline := time.Now().Add(maxWaitTime)

	for time.Now().Before(deadline) {
		// Buscar enriquecimentos
		enrichments, err := w.enrichmentService.GetGroupEnrichments(ctx, group.ID())
		if err != nil {
			return fmt.Errorf("failed to get enrichments: %w", err)
		}

		// Verificar se todos estão completos ou falharam
		allDone := true
		for _, e := range enrichments {
			if !e.IsFinal() {
				allDone = false
				break
			}
		}

		if allDone {
			w.logger.Debug("All enrichments completed",
				zap.String("group_id", group.ID().String()),
				zap.Int("enrichment_count", len(enrichments)))
			return nil
		}

		// Aguardar antes de próxima verificação
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(pollInterval):
			// Continua polling
		}
	}

	// Timeout - continua mesmo sem todos enrichments
	w.logger.Warn("Timeout waiting for enrichments",
		zap.String("group_id", group.ID().String()),
		zap.Duration("waited", maxWaitTime))

	return nil
}

// SetTickerInterval configura o intervalo do ticker (útil para testes)
func (w *MessageGroupWorker) SetTickerInterval(interval time.Duration) {
	w.tickerInterval = interval
}

// SetBatchSize configura o tamanho do batch (útil para testes)
func (w *MessageGroupWorker) SetBatchSize(size int) {
	w.batchSize = size
}
