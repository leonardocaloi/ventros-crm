package message

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/ventros/crm/internal/domain/crm/channel"
	"github.com/ventros/crm/internal/domain/crm/message"
	"github.com/ventros/crm/internal/domain/crm/message_group"
	"go.uber.org/zap"
)

// MessageDebouncerService gerencia o agrupamento e debouncing de mensagens
type MessageDebouncerService struct {
	logger           *zap.Logger
	messageGroupRepo message_group.Repository
	messageRepo      message.Repository
	channelRepo      channel.Repository
	redisClient      *redis.Client
}

// NewMessageDebouncerService cria um novo serviço de debouncer
func NewMessageDebouncerService(
	logger *zap.Logger,
	messageGroupRepo message_group.Repository,
	messageRepo message.Repository,
	channelRepo channel.Repository,
	redisClient *redis.Client,
) *MessageDebouncerService {
	return &MessageDebouncerService{
		logger:           logger,
		messageGroupRepo: messageGroupRepo,
		messageRepo:      messageRepo,
		channelRepo:      channelRepo,
		redisClient:      redisClient,
	}
}

// ProcessInboundMessage processa mensagem inbound e sempre agrupa
// TODAS as mensagens passam pelo debouncer (incluindo texto puro)
// para permitir concatenação e envio para AI Agent
func (s *MessageDebouncerService) ProcessInboundMessage(
	ctx context.Context,
	msg *message.Message,
	channelID uuid.UUID,
	sessionID uuid.UUID,
) error {
	// 1. Buscar configuração do canal
	ch, err := s.channelRepo.GetByID(channelID)
	if err != nil {
		return fmt.Errorf("failed to get channel: %w", err)
	}

	// 2. Obter timeout do debouncer do canal (default 15s se não configurado)
	debounceTimeout := ch.GetDebounceDuration()

	// 3. Verificar se existe grupo ativo para este contato/canal
	activeGroup, err := s.messageGroupRepo.FindActiveByContact(ctx, msg.ContactID(), channelID)
	if err != nil {
		return fmt.Errorf("failed to find active group: %w", err)
	}

	// 4. Se não existe grupo ativo, criar novo
	if activeGroup == nil {
		group, err := message_group.NewMessageGroup(
			msg.ContactID(),
			channelID,
			sessionID,
			msg.CustomerID().String(), // CustomerID como TenantID
			msg.ID(),
			debounceTimeout,
		)
		if err != nil {
			return fmt.Errorf("failed to create message group: %w", err)
		}

		if err := s.messageGroupRepo.Save(ctx, group); err != nil {
			return fmt.Errorf("failed to save message group: %w", err)
		}

		s.logger.Info("Created new message group",
			zap.String("group_id", group.ID().String()),
			zap.String("message_id", msg.ID().String()),
			zap.Duration("timeout", debounceTimeout))

		// Agendar processamento do grupo no Redis
		if err := s.scheduleGroupProcessing(ctx, group.ID(), debounceTimeout); err != nil {
			s.logger.Error("Failed to schedule group processing", zap.Error(err))
		}

		return nil
	}

	// 5. Adicionar mensagem ao grupo existente (reinicia o timer)
	if err := activeGroup.AddMessage(msg.ID(), debounceTimeout); err != nil {
		return fmt.Errorf("failed to add message to group: %w", err)
	}

	if err := s.messageGroupRepo.Save(ctx, activeGroup); err != nil {
		return fmt.Errorf("failed to update message group: %w", err)
	}

	s.logger.Info("Added message to existing group",
		zap.String("group_id", activeGroup.ID().String()),
		zap.String("message_id", msg.ID().String()),
		zap.Int("total_messages", activeGroup.MessageCount()))

	// Re-agendar processamento (reset do timer)
	if err := s.scheduleGroupProcessing(ctx, activeGroup.ID(), debounceTimeout); err != nil {
		s.logger.Error("Failed to reschedule group processing", zap.Error(err))
	}

	return nil
}

// ProcessExpiredGroups processa grupos que expiraram
func (s *MessageDebouncerService) ProcessExpiredGroups(ctx context.Context) error {
	// Buscar grupos expirados
	expiredGroups, err := s.messageGroupRepo.FindExpired(ctx, 100)
	if err != nil {
		return fmt.Errorf("failed to find expired groups: %w", err)
	}

	s.logger.Info("Processing expired groups", zap.Int("count", len(expiredGroups)))

	for _, group := range expiredGroups {
		if err := s.processGroup(ctx, group); err != nil {
			s.logger.Error("Failed to process expired group",
				zap.Error(err),
				zap.String("group_id", group.ID().String()))
			continue
		}
	}

	return nil
}

// processGroup processa um grupo de mensagens
func (s *MessageDebouncerService) processGroup(ctx context.Context, group *message_group.MessageGroup) error {
	// 1. Marcar como processando
	if err := group.MarkAsProcessing(); err != nil {
		return fmt.Errorf("failed to mark as processing: %w", err)
	}

	if err := s.messageGroupRepo.Save(ctx, group); err != nil {
		return fmt.Errorf("failed to save group: %w", err)
	}

	s.logger.Info("Processing message group",
		zap.String("group_id", group.ID().String()),
		zap.Int("message_count", group.MessageCount()))

	// 2. Publicar evento para processamento assíncrono
	// O worker pegará este evento e processará os enriquecimentos
	if err := s.publishGroupForEnrichment(ctx, group); err != nil {
		return fmt.Errorf("failed to publish group for enrichment: %w", err)
	}

	return nil
}

// scheduleGroupProcessing agenda processamento do grupo no Redis
func (s *MessageDebouncerService) scheduleGroupProcessing(ctx context.Context, groupID uuid.UUID, delay time.Duration) error {
	// Usar Redis sorted set com score = timestamp de expiração
	score := float64(time.Now().Add(delay).Unix())
	key := "message_groups:scheduled"

	return s.redisClient.ZAdd(ctx, key, redis.Z{
		Score:  score,
		Member: groupID.String(),
	}).Err()
}

// publishGroupForEnrichment publica grupo para processamento de enriquecimento
func (s *MessageDebouncerService) publishGroupForEnrichment(ctx context.Context, group *message_group.MessageGroup) error {
	// TODO: Publicar para fila RabbitMQ ou Temporal workflow
	// Por enquanto, apenas log
	payload := map[string]interface{}{
		"group_id":      group.ID().String(),
		"contact_id":    group.ContactID().String(),
		"channel_id":    group.ChannelID().String(),
		"session_id":    group.SessionID().String(),
		"message_count": group.MessageCount(),
		"message_ids":   group.MessageIDs(),
	}

	jsonPayload, _ := json.Marshal(payload)
	s.logger.Info("Publishing group for enrichment",
		zap.String("payload", string(jsonPayload)))

	// Placeholder: Publicar no Redis pub/sub por enquanto
	return s.redisClient.Publish(ctx, "message_enrichment:jobs", jsonPayload).Err()
}
