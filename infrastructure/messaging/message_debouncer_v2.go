package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// MessageDebouncerV2 implementa debouncing igual ao n8n
// Usa Redis LIST (RPUSH/LRANGE) + timestamp check
type MessageDebouncerV2 struct {
	redis         *redis.Client
	waitDuration  time.Duration // tempo para aguardar (15s default)
	keyPrefix     string
	ttl           time.Duration
	processorFunc ProcessorFunc // callback para processar batch
	maxRetries    int           // max tentativas de pull antes de forçar
}

// ProcessorFunc processa um batch de mensagens
// Retorna se foi processado com sucesso
type ProcessorFunc func(ctx context.Context, sessionKey string, messages []BufferedMessage) error

// BufferedMessage representa uma mensagem no buffer Redis
type BufferedMessage struct {
	MessageID   string                 `json:"message_id"`
	Text        string                 `json:"text,omitempty"`
	Type        string                 `json:"type"`
	Timestamp   int64                  `json:"timestamp"`
	FromContact bool                   `json:"from_contact"`
	ContactID   string                 `json:"contact_id"`
	SessionID   string                 `json:"session_id,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// NewMessageDebouncerV2 cria um novo debouncer estilo n8n
func NewMessageDebouncerV2(
	redisClient *redis.Client,
	waitDuration time.Duration,
	processor ProcessorFunc,
) *MessageDebouncerV2 {
	if waitDuration == 0 {
		waitDuration = 15 * time.Second // padrão n8n
	}

	return &MessageDebouncerV2{
		redis:         redisClient,
		waitDuration:  waitDuration,
		keyPrefix:     "msg:buffer:",
		ttl:           5 * time.Minute,
		processorFunc: processor,
		maxRetries:    10, // máximo 10 retries (10 * 15s = 2.5min)
	}
}

// Push adiciona mensagem ao buffer (RPUSH - tail)
func (d *MessageDebouncerV2) Push(ctx context.Context, sessionKey string, msg BufferedMessage) error {
	key := d.getBufferKey(sessionKey)

	// Serializa mensagem
	msgJSON, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// RPUSH (adiciona no final da lista)
	if err := d.redis.RPush(ctx, key, msgJSON).Err(); err != nil {
		return fmt.Errorf("failed to push message: %w", err)
	}

	// Atualiza TTL
	d.redis.Expire(ctx, key, d.ttl)

	fmt.Printf("📥 [Debouncer] Pushed message: session=%s, type=%s\n",
		sessionKey, msg.Type)

	return nil
}

// PushAndCheck adiciona mensagem e inicia loop de verificação
// Este é o método principal que replica o fluxo n8n
func (d *MessageDebouncerV2) PushAndCheck(
	ctx context.Context,
	sessionKey string,
	msg BufferedMessage,
) error {
	// 1. Push mensagem
	if err := d.Push(ctx, sessionKey, msg); err != nil {
		return err
	}

	// 2. Pull lista completa
	messages, err := d.Pull(ctx, sessionKey)
	if err != nil {
		return err
	}

	// 3. Check se deve processar (Switch logic do n8n)
	retries := 0
	for {
		decision := d.checkProcessingDecision(msg.MessageID, messages)

		switch decision {
		case DecisionNothing:
			// Mensagem duplicada ou já está no buffer
			fmt.Printf("⏭️  [Debouncer] Duplicate message, skipping: session=%s\n", sessionKey)
			return nil

		case DecisionProceed:
			// Timeout atingido, processar agora
			fmt.Printf("✅ [Debouncer] Timeout reached, processing: session=%s, count=%d\n",
				sessionKey, len(messages))
			return d.processAndDelete(ctx, sessionKey, messages)

		case DecisionWait:
			// Aguardar mais tempo
			retries++
			if retries > d.maxRetries {
				// Forçar processamento após max retries
				fmt.Printf("⚠️  [Debouncer] Max retries reached, forcing process: session=%s\n", sessionKey)
				return d.processAndDelete(ctx, sessionKey, messages)
			}

			fmt.Printf("⏳ [Debouncer] Waiting %v: session=%s, retry=%d/%d\n",
				d.waitDuration, sessionKey, retries, d.maxRetries)

			// Wait (como no n8n)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(d.waitDuration):
				// Continua loop
			}

			// Pull novamente após wait
			messages, err = d.Pull(ctx, sessionKey)
			if err != nil {
				return err
			}
		}
	}
}

// ProcessingDecision representa a decisão do switch
type ProcessingDecision int

const (
	DecisionNothing ProcessingDecision = iota // Ignora (duplicada)
	DecisionProceed                           // Processa agora
	DecisionWait                              // Aguarda mais
)

// checkProcessingDecision implementa a lógica do Switch do n8n
func (d *MessageDebouncerV2) checkProcessingDecision(
	currentMessageID string,
	messages []BufferedMessage,
) ProcessingDecision {
	if len(messages) == 0 {
		return DecisionNothing
	}

	// Condição 1: Primeira mensagem da lista ≠ mensagem atual?
	// Se sim, mensagem já está no buffer (duplicada)
	firstMsg := messages[0]
	if firstMsg.MessageID != currentMessageID {
		return DecisionNothing
	}

	// Condição 2: Última mensagem tem > waitDuration de idade?
	lastMsg := messages[len(messages)-1]
	lastMsgTime := time.UnixMilli(lastMsg.Timestamp)
	age := time.Since(lastMsgTime)

	if age >= d.waitDuration {
		return DecisionProceed
	}

	// Senão: aguardar mais
	return DecisionWait
}

// Pull busca todas as mensagens do buffer (LRANGE 0 -1)
func (d *MessageDebouncerV2) Pull(ctx context.Context, sessionKey string) ([]BufferedMessage, error) {
	key := d.getBufferKey(sessionKey)

	// LRANGE 0 -1 (todas as mensagens)
	results, err := d.redis.LRange(ctx, key, 0, -1).Result()
	if err == redis.Nil {
		return []BufferedMessage{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to pull messages: %w", err)
	}

	// Deserializa mensagens
	messages := make([]BufferedMessage, 0, len(results))
	for _, result := range results {
		var msg BufferedMessage
		if err := json.Unmarshal([]byte(result), &msg); err != nil {
			fmt.Printf("⚠️  [Debouncer] Failed to unmarshal message: %v\n", err)
			continue
		}
		messages = append(messages, msg)
	}

	return messages, nil
}

// processAndDelete processa o batch e deleta do Redis
func (d *MessageDebouncerV2) processAndDelete(
	ctx context.Context,
	sessionKey string,
	messages []BufferedMessage,
) error {
	key := d.getBufferKey(sessionKey)

	// 1. Chama processor (se configurado)
	if d.processorFunc != nil {
		if err := d.processorFunc(ctx, sessionKey, messages); err != nil {
			return fmt.Errorf("processor failed: %w", err)
		}
	}

	// 2. Delete buffer (como no n8n)
	if err := d.redis.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to delete buffer: %w", err)
	}

	fmt.Printf("🗑️  [Debouncer] Deleted buffer: session=%s, processed=%d messages\n",
		sessionKey, len(messages))

	return nil
}

// ForceProcess força o processamento imediato (útil para testes)
func (d *MessageDebouncerV2) ForceProcess(ctx context.Context, sessionKey string) error {
	messages, err := d.Pull(ctx, sessionKey)
	if err != nil {
		return err
	}

	if len(messages) == 0 {
		return nil
	}

	return d.processAndDelete(ctx, sessionKey, messages)
}

// GetBufferSize retorna quantas mensagens estão no buffer
func (d *MessageDebouncerV2) GetBufferSize(ctx context.Context, sessionKey string) (int, error) {
	key := d.getBufferKey(sessionKey)
	count, err := d.redis.LLen(ctx, key).Result()
	if err == redis.Nil {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return int(count), nil
}

// ClearBuffer limpa o buffer sem processar
func (d *MessageDebouncerV2) ClearBuffer(ctx context.Context, sessionKey string) error {
	key := d.getBufferKey(sessionKey)
	return d.redis.Del(ctx, key).Err()
}

// Helper methods

func (d *MessageDebouncerV2) getBufferKey(sessionKey string) string {
	return fmt.Sprintf("%s%s", d.keyPrefix, sessionKey)
}

// BuildSessionKey constrói chave única para sessão
// Formato: {contact_id}:{inbox_type}:{inbox_id}
func BuildSessionKey(contactID, inboxType, inboxID string) string {
	return fmt.Sprintf("%s:%s:%s", contactID, inboxType, inboxID)
}

// ExtractSessionKeyFromUUIDs versão usando UUIDs
func ExtractSessionKeyFromUUIDs(contactID, channelID uuid.UUID, channelType string) string {
	return BuildSessionKey(contactID.String(), channelType, channelID.String())
}
