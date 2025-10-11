package message

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/caloi/ventros-crm/infrastructure/ai"
	"github.com/caloi/ventros-crm/internal/domain/channel"
	"github.com/caloi/ventros-crm/internal/domain/message"
	"github.com/caloi/ventros-crm/internal/domain/message_enrichment"
	"github.com/caloi/ventros-crm/internal/domain/message_group"
)

// MessageEnrichmentService processa enriquecimentos de mensagens usando DDD
type MessageEnrichmentService struct {
	logger         *zap.Logger
	enrichmentRepo message_enrichment.Repository
	messageRepo    message.Repository
	channelRepo    channel.Repository
	mimetypeRouter *ai.MimetypeRouter
	audioSplitter  *ai.AudioSplitter
}

// NewMessageEnrichmentService cria um novo serviço de enriquecimento
func NewMessageEnrichmentService(
	logger *zap.Logger,
	enrichmentRepo message_enrichment.Repository,
	messageRepo message.Repository,
	channelRepo channel.Repository,
) *MessageEnrichmentService {
	return &MessageEnrichmentService{
		logger:         logger,
		enrichmentRepo: enrichmentRepo,
		messageRepo:    messageRepo,
		channelRepo:    channelRepo,
		mimetypeRouter: ai.NewMimetypeRouter(),
		audioSplitter:  ai.NewAudioSplitter(logger),
	}
}

// ProcessGroupEnrichments processa enriquecimentos para um grupo de mensagens
func (s *MessageEnrichmentService) ProcessGroupEnrichments(
	ctx context.Context,
	group *message_group.MessageGroup,
) error {
	s.logger.Info("Processing group enrichments",
		zap.String("group_id", group.ID().String()),
		zap.Int("message_count", group.MessageCount()))

	// Buscar canal para configurações de IA
	ch, err := s.channelRepo.GetByID(group.ChannelID())
	if err != nil {
		return fmt.Errorf("failed to get channel: %w", err)
	}

	// Processar cada mensagem do grupo
	for _, messageID := range group.MessageIDs() {
		if err := s.processMessageEnrichment(ctx, messageID, group.ID(), ch); err != nil {
			s.logger.Error("Failed to process message enrichment",
				zap.Error(err),
				zap.String("message_id", messageID.String()))
			// Continuar processando outras mensagens mesmo se uma falhar
			continue
		}
	}

	return nil
}

// processMessageEnrichment processa enriquecimento de uma mensagem individual
func (s *MessageEnrichmentService) processMessageEnrichment(
	ctx context.Context,
	messageID uuid.UUID,
	groupID uuid.UUID,
	ch *channel.Channel,
) error {
	// 1. Buscar mensagem
	msg, err := s.messageRepo.FindByID(ctx, messageID)
	if err != nil {
		return fmt.Errorf("failed to get message: %w", err)
	}

	// 2. Se é texto puro, não precisa enriquecer
	if msg.MediaURL() == nil && msg.ContentType() == message.ContentTypeText {
		s.logger.Debug("Skipping text-only message",
			zap.String("message_id", messageID.String()))
		return nil
	}

	// 3. Verificar se tem media URL
	if msg.MediaURL() == nil {
		s.logger.Debug("Message has no media URL",
			zap.String("message_id", messageID.String()))
		return nil
	}

	// 4. Determinar tipo de conteúdo e provider
	enrichmentContentType, provider := s.determineEnrichmentType(msg)
	if enrichmentContentType == "" {
		s.logger.Debug("No enrichment needed for content type",
			zap.String("message_id", messageID.String()),
			zap.String("content_type", string(msg.ContentType())))
		return nil
	}

	// 5. Determinar contexto de processamento (para providers como Vision)
	// Default: chat_message para imagens em conversa
	var context *string
	if enrichmentContentType == message_enrichment.EnrichmentTypeImage {
		ctx := "chat_message" // Default context para imagens
		context = &ctx
	}

	// 6. Criar registro de enriquecimento usando domain aggregate
	enrichment, err := message_enrichment.NewMessageEnrichment(
		messageID,
		groupID,
		enrichmentContentType,
		provider,
		*msg.MediaURL(),
		context,
	)
	if err != nil {
		return fmt.Errorf("failed to create enrichment aggregate: %w", err)
	}

	// 7. Persistir
	if err := s.enrichmentRepo.Save(ctx, enrichment); err != nil {
		return fmt.Errorf("failed to save enrichment: %w", err)
	}

	s.logger.Info("Created enrichment record",
		zap.String("enrichment_id", enrichment.ID().String()),
		zap.String("message_id", messageID.String()),
		zap.String("content_type", string(enrichmentContentType)),
		zap.String("provider", string(provider)))

	// Enrichment criado com sucesso - worker irá processar assincronamente

	return nil
}

// determineEnrichmentType determina o tipo de enriquecimento e provider baseado na mensagem
func (s *MessageEnrichmentService) determineEnrichmentType(msg *message.Message) (message_enrichment.EnrichmentContentType, message_enrichment.EnrichmentProvider) {
	// Map message content type to enrichment content type
	switch msg.ContentType() {
	case message.ContentTypeVoice:
		// PTT - Prioridade máxima
		return message_enrichment.EnrichmentTypeVoice, message_enrichment.ProviderWhisper

	case message.ContentTypeAudio:
		// Áudio geral
		return message_enrichment.EnrichmentTypeAudio, message_enrichment.ProviderWhisper

	case message.ContentTypeImage:
		// Imagem - OCR + descrição
		return message_enrichment.EnrichmentTypeImage, message_enrichment.ProviderVision

	case message.ContentTypeVideo:
		// Vídeo - precisa FFmpeg primeiro para extrair áudio
		return message_enrichment.EnrichmentTypeVideo, message_enrichment.ProviderFFmpeg

	case message.ContentTypeDocument:
		// Documento - parsing com LlamaParse
		return message_enrichment.EnrichmentTypeDocument, message_enrichment.ProviderLlamaParse

	default:
		// Texto ou outros tipos - sem enriquecimento
		return "", ""
	}
}

// GetGroupEnrichments retorna todos os enriquecimentos de um grupo
func (s *MessageEnrichmentService) GetGroupEnrichments(
	ctx context.Context,
	groupID uuid.UUID,
) ([]*message_enrichment.MessageEnrichment, error) {
	return s.enrichmentRepo.FindByMessageGroupID(ctx, groupID)
}

// WaitForEnrichmentsCompletion aguarda enriquecimentos completarem (polling)
// Retorna quando todos estiverem completed/failed ou timeout expirar
func (s *MessageEnrichmentService) WaitForEnrichmentsCompletion(
	ctx context.Context,
	groupID uuid.UUID,
	timeout time.Duration,
) error {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	timeoutChan := time.After(timeout)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case <-timeoutChan:
			return fmt.Errorf("timeout waiting for enrichments completion")

		case <-ticker.C:
			// Buscar enriquecimentos do grupo
			enrichments, err := s.enrichmentRepo.FindByMessageGroupID(ctx, groupID)
			if err != nil {
				return fmt.Errorf("failed to get enrichments: %w", err)
			}

			// Verificar se todos estão finalizados
			allCompleted := true
			for _, e := range enrichments {
				if !e.IsFinal() {
					allCompleted = false
					break
				}
			}

			if allCompleted {
				s.logger.Info("All enrichments completed",
					zap.String("group_id", groupID.String()),
					zap.Int("total", len(enrichments)))
				return nil
			}
		}
	}
}

// GetEnrichmentStats retorna estatísticas de enriquecimento
func (s *MessageEnrichmentService) GetEnrichmentStats(ctx context.Context) (map[string]int, error) {
	stats := make(map[string]int)

	pending, err := s.enrichmentRepo.CountByStatus(ctx, message_enrichment.StatusPending)
	if err != nil {
		return nil, err
	}
	stats["pending"] = pending

	processing, err := s.enrichmentRepo.CountByStatus(ctx, message_enrichment.StatusProcessing)
	if err != nil {
		return nil, err
	}
	stats["processing"] = processing

	completed, err := s.enrichmentRepo.CountByStatus(ctx, message_enrichment.StatusCompleted)
	if err != nil {
		return nil, err
	}
	stats["completed"] = completed

	failed, err := s.enrichmentRepo.CountByStatus(ctx, message_enrichment.StatusFailed)
	if err != nil {
		return nil, err
	}
	stats["failed"] = failed

	return stats, nil
}
