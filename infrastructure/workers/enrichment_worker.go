package workers

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/caloi/ventros-crm/infrastructure/ai"
	"github.com/caloi/ventros-crm/internal/domain/message_enrichment"
)

// EnrichmentWorker processa enrichments pendentes periodicamente usando polling
// Arquitetura simples sem RabbitMQ/Temporal - apenas busca enrichments pendentes e processa
type EnrichmentWorker struct {
	logger            *zap.Logger
	enrichmentRepo    message_enrichment.Repository
	providerFactory   *ai.ProviderFactory
	tickerInterval    time.Duration // Default: 5s
	batchSize         int           // Default: 10
	stopChan          chan struct{}
	maxRetries        int           // Máximo de retries antes de marcar como failed
	retryDelay        time.Duration // Delay entre retries
}

// NewEnrichmentWorker cria um novo worker de processamento de enrichments
func NewEnrichmentWorker(
	logger *zap.Logger,
	enrichmentRepo message_enrichment.Repository,
	providerFactory *ai.ProviderFactory,
) *EnrichmentWorker {
	return &EnrichmentWorker{
		logger:          logger,
		enrichmentRepo:  enrichmentRepo,
		providerFactory: providerFactory,
		tickerInterval:  5 * time.Second,  // Processa a cada 5 segundos
		batchSize:       10,                // Processa até 10 enrichments por vez
		maxRetries:      3,                 // Até 3 tentativas
		retryDelay:      30 * time.Second,  // 30s entre retries
		stopChan:        make(chan struct{}),
	}
}

// Start inicia o worker
func (w *EnrichmentWorker) Start(ctx context.Context) error {
	w.logger.Info("Starting enrichment worker",
		zap.Duration("interval", w.tickerInterval),
		zap.Int("batch_size", w.batchSize))

	ticker := time.NewTicker(w.tickerInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("Enrichment worker stopped (context done)")
			return ctx.Err()

		case <-w.stopChan:
			w.logger.Info("Enrichment worker stopped (stop signal)")
			return nil

		case <-ticker.C:
			if err := w.processPendingEnrichments(ctx); err != nil {
				w.logger.Error("Failed to process pending enrichments",
					zap.Error(err))
				// Continua processando mesmo se houver erro
			}
		}
	}
}

// Stop para o worker
func (w *EnrichmentWorker) Stop() {
	close(w.stopChan)
}

// processPendingEnrichments processa enrichments pendentes
func (w *EnrichmentWorker) processPendingEnrichments(ctx context.Context) error {
	// 1. Buscar enrichments pendentes (ordenados por prioridade)
	pendingEnrichments, err := w.enrichmentRepo.FindPending(ctx, w.batchSize)
	if err != nil {
		return fmt.Errorf("failed to find pending enrichments: %w", err)
	}

	if len(pendingEnrichments) == 0 {
		return nil // Nenhum enrichment pendente
	}

	w.logger.Info("Processing pending enrichments",
		zap.Int("count", len(pendingEnrichments)))

	// 2. Processar cada enrichment
	successCount := 0
	errorCount := 0

	for _, enrichment := range pendingEnrichments {
		if err := w.processEnrichment(ctx, enrichment); err != nil {
			w.logger.Error("Failed to process enrichment",
				zap.Error(err),
				zap.String("enrichment_id", enrichment.ID().String()),
				zap.String("content_type", string(enrichment.ContentType())))
			errorCount++
			continue
		}
		successCount++
	}

	w.logger.Info("Pending enrichments processed",
		zap.Int("success", successCount),
		zap.Int("errors", errorCount),
		zap.Int("total", len(pendingEnrichments)))

	return nil
}

// processEnrichment processa um enrichment individual
func (w *EnrichmentWorker) processEnrichment(
	ctx context.Context,
	enrichment *message_enrichment.MessageEnrichment,
) error {
	w.logger.Info("Processing enrichment",
		zap.String("enrichment_id", enrichment.ID().String()),
		zap.String("message_id", enrichment.MessageID().String()),
		zap.String("content_type", string(enrichment.ContentType())),
		zap.String("provider", string(enrichment.Provider())),
		zap.Uint8("priority", enrichment.Priority()))

	// 1. Marcar como processando
	if err := enrichment.MarkAsProcessing(); err != nil {
		return fmt.Errorf("failed to mark as processing: %w", err)
	}

	if err := w.enrichmentRepo.Save(ctx, enrichment); err != nil {
		return fmt.Errorf("failed to save processing status: %w", err)
	}

	// 2. Obter provider apropriado
	provider, err := w.providerFactory.GetProvider(enrichment.Provider())
	if err != nil {
		// Provider não disponível - marcar como failed
		enrichment.MarkAsFailed(fmt.Sprintf("Provider not available: %v", err))
		w.enrichmentRepo.Save(ctx, enrichment)
		return fmt.Errorf("provider not available: %w", err)
	}

	// 3. Processar com provider (passar context para providers que suportam, como Vision)
	result, err := provider.Process(ctx, enrichment.MediaURL(), enrichment.ContentType(), enrichment.Context())
	if err != nil {
		// Processar falhou - marcar como failed
		errorMsg := fmt.Sprintf("Processing failed: %v", err)
		enrichment.MarkAsFailed(errorMsg)
		w.enrichmentRepo.Save(ctx, enrichment)
		return fmt.Errorf("processing failed: %w", err)
	}

	// 4. Marcar como concluído
	if err := enrichment.MarkAsCompleted(
		result.ExtractedText,
		result.Metadata,
		result.ProcessingTime,
	); err != nil {
		return fmt.Errorf("failed to mark as completed: %w", err)
	}

	if err := w.enrichmentRepo.Save(ctx, enrichment); err != nil {
		return fmt.Errorf("failed to save completed enrichment: %w", err)
	}

	w.logger.Info("Enrichment processed successfully",
		zap.String("enrichment_id", enrichment.ID().String()),
		zap.String("message_id", enrichment.MessageID().String()),
		zap.Int("text_length", len(result.ExtractedText)),
		zap.Duration("processing_time", result.ProcessingTime))

	return nil
}

// RecoverStuckEnrichments recupera enrichments travados em "processing"
// Deve ser chamado periodicamente (ex: a cada 5 minutos)
func (w *EnrichmentWorker) RecoverStuckEnrichments(ctx context.Context) error {
	// Buscar enrichments em processing há mais de 10 minutos
	stuckEnrichments, err := w.enrichmentRepo.FindProcessing(ctx, 10)
	if err != nil {
		return fmt.Errorf("failed to find stuck enrichments: %w", err)
	}

	if len(stuckEnrichments) == 0 {
		return nil
	}

	w.logger.Warn("Found stuck enrichments, marking as failed",
		zap.Int("count", len(stuckEnrichments)))

	for _, enrichment := range stuckEnrichments {
		// Marcar como failed
		errorMsg := "Processing timeout - enrichment was stuck for more than 10 minutes"
		if err := enrichment.MarkAsFailed(errorMsg); err != nil {
			w.logger.Error("Failed to mark stuck enrichment as failed",
				zap.Error(err),
				zap.String("enrichment_id", enrichment.ID().String()))
			continue
		}

		if err := w.enrichmentRepo.Save(ctx, enrichment); err != nil {
			w.logger.Error("Failed to save failed enrichment",
				zap.Error(err),
				zap.String("enrichment_id", enrichment.ID().String()))
		}
	}

	return nil
}

// GetStats retorna estatísticas do worker
func (w *EnrichmentWorker) GetStats(ctx context.Context) (map[string]int, error) {
	stats := make(map[string]int)

	pending, err := w.enrichmentRepo.CountByStatus(ctx, message_enrichment.StatusPending)
	if err != nil {
		return nil, err
	}
	stats["pending"] = pending

	processing, err := w.enrichmentRepo.CountByStatus(ctx, message_enrichment.StatusProcessing)
	if err != nil {
		return nil, err
	}
	stats["processing"] = processing

	completed, err := w.enrichmentRepo.CountByStatus(ctx, message_enrichment.StatusCompleted)
	if err != nil {
		return nil, err
	}
	stats["completed"] = completed

	failed, err := w.enrichmentRepo.CountByStatus(ctx, message_enrichment.StatusFailed)
	if err != nil {
		return nil, err
	}
	stats["failed"] = failed

	return stats, nil
}

// SetTickerInterval configura o intervalo do ticker (útil para testes)
func (w *EnrichmentWorker) SetTickerInterval(interval time.Duration) {
	w.tickerInterval = interval
}

// SetBatchSize configura o tamanho do batch (útil para testes)
func (w *EnrichmentWorker) SetBatchSize(size int) {
	w.batchSize = size
}

// SetMaxRetries configura o número máximo de retries
func (w *EnrichmentWorker) SetMaxRetries(retries int) {
	w.maxRetries = retries
}
