package ai

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/ventros/crm/internal/domain/crm/channel"
	"github.com/ventros/crm/internal/domain/crm/message"
	"go.uber.org/zap"
)

// MessageEnrichmentRequest representa uma requisição de enriquecimento de mensagem
type MessageEnrichmentRequest struct {
	MessageID   uuid.UUID
	ChannelID   uuid.UUID
	ProjectID   uuid.UUID
	TenantID    string
	ContentType string
	Mimetype    string
	MediaURL    string
	Text        string
	IsPTT       bool // Push-to-Talk (áudio de voz do WhatsApp)
	SizeBytes   int64
}

// MessageEnrichmentProcessor processa mensagens para enriquecimento com IA
type MessageEnrichmentProcessor struct {
	logger         *zap.Logger
	channelRepo    channel.Repository
	messageRepo    message.Repository
	mimetypeRouter *MimetypeRouter
	debouncer      *AIDebouncer
}

// NewMessageEnrichmentProcessor cria um novo processador de enriquecimento
func NewMessageEnrichmentProcessor(
	logger *zap.Logger,
	channelRepo channel.Repository,
	messageRepo message.Repository,
) *MessageEnrichmentProcessor {
	return &MessageEnrichmentProcessor{
		logger:         logger,
		channelRepo:    channelRepo,
		messageRepo:    messageRepo,
		mimetypeRouter: NewMimetypeRouter(),
		debouncer:      NewAIDebouncer(logger),
	}
}

// ProcessMessage processa uma mensagem para enriquecimento com IA
func (p *MessageEnrichmentProcessor) ProcessMessage(ctx context.Context, req MessageEnrichmentRequest) error {
	// 1. Buscar canal para verificar configurações de IA
	ch, err := p.channelRepo.GetByID(req.ChannelID)
	if err != nil {
		return fmt.Errorf("failed to get channel: %w", err)
	}

	// 2. Verificar se IA está habilitada no canal
	if !ch.AIEnabled {
		p.logger.Debug("AI not enabled for channel, skipping enrichment",
			zap.String("channel_id", req.ChannelID.String()),
			zap.String("message_id", req.MessageID.String()))
		return nil
	}

	// 3. Determinar tipo de conteúdo baseado no mimetype
	aiContentType := p.mimetypeRouter.RouteToContentType(req.Mimetype, req.IsPTT)

	p.logger.Info("Routing message for AI processing",
		zap.String("message_id", req.MessageID.String()),
		zap.String("mimetype", req.Mimetype),
		zap.String("ai_content_type", string(aiContentType)),
		zap.Bool("is_ptt", req.IsPTT))

	// 4. Verificar se deve processar este tipo de conteúdo
	if !ch.ShouldProcessAIContent(aiContentType) {
		p.logger.Debug("AI processing not enabled for content type",
			zap.String("channel_id", req.ChannelID.String()),
			zap.String("content_type", string(aiContentType)),
			zap.String("message_id", req.MessageID.String()))
		return nil
	}

	// 5. Obter configuração de processamento para este tipo
	config := ch.GetAIProcessingConfig(aiContentType)
	if config == nil {
		// Usar configuração padrão
		defaultConfig := channel.GetDefaultAIConfig(aiContentType)
		config = &defaultConfig
	}

	// 6. Verificar tamanho do arquivo
	if req.SizeBytes > config.MaxSizeBytes {
		p.logger.Warn("File size exceeds maximum for AI processing",
			zap.String("message_id", req.MessageID.String()),
			zap.Int64("size_bytes", req.SizeBytes),
			zap.Int64("max_size_bytes", config.MaxSizeBytes))
		return fmt.Errorf("file size %d bytes exceeds maximum %d bytes", req.SizeBytes, config.MaxSizeBytes)
	}

	// 7. Aplicar debouncer
	if config.DebounceMs > 0 {
		shouldProcess := p.debouncer.ShouldProcess(
			req.MessageID.String(),
			time.Duration(config.DebounceMs)*time.Millisecond,
		)
		if !shouldProcess {
			p.logger.Debug("Message debounced, skipping AI processing",
				zap.String("message_id", req.MessageID.String()),
				zap.Int("debounce_ms", config.DebounceMs))
			return nil
		}
	}

	// 8. Criar job de processamento assíncrono
	enrichmentJob := EnrichmentJob{
		MessageID:   req.MessageID,
		ChannelID:   req.ChannelID,
		ProjectID:   req.ProjectID,
		TenantID:    req.TenantID,
		ContentType: aiContentType,
		Provider:    config.Provider,
		Model:       config.Model,
		Priority:    config.Priority,
		MediaURL:    req.MediaURL,
		Text:        req.Text,
		IsPTT:       req.IsPTT,
		SizeBytes:   req.SizeBytes,
		Config:      *config,
		CreatedAt:   time.Now(),
	}

	// 9. Publicar job para fila de processamento
	if err := p.publishEnrichmentJob(ctx, enrichmentJob); err != nil {
		p.logger.Error("Failed to publish enrichment job",
			zap.Error(err),
			zap.String("message_id", req.MessageID.String()))
		return fmt.Errorf("failed to publish enrichment job: %w", err)
	}

	p.logger.Info("Message enrichment job created",
		zap.String("message_id", req.MessageID.String()),
		zap.String("content_type", string(aiContentType)),
		zap.String("provider", config.Provider),
		zap.String("model", config.Model),
		zap.Int("priority", config.Priority))

	return nil
}

// publishEnrichmentJob publica job para fila de processamento
func (p *MessageEnrichmentProcessor) publishEnrichmentJob(ctx context.Context, job EnrichmentJob) error {
	// TODO: Implementar publicação para fila RabbitMQ ou usar Temporal workflow
	// Por enquanto, apenas log
	p.logger.Info("Publishing enrichment job to queue",
		zap.String("message_id", job.MessageID.String()),
		zap.String("provider", job.Provider),
		zap.Int("priority", job.Priority))

	// Placeholder para implementação futura
	return nil
}

// EnrichmentJob representa um job de enriquecimento de mensagem
type EnrichmentJob struct {
	MessageID   uuid.UUID
	ChannelID   uuid.UUID
	ProjectID   uuid.UUID
	TenantID    string
	ContentType channel.AIContentType
	Provider    string
	Model       string
	Priority    int
	MediaURL    string
	Text        string
	IsPTT       bool
	SizeBytes   int64
	Config      channel.AIProcessingConfig
	CreatedAt   time.Time
}
