package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// DebouncedMessage representa uma mensagem no buffer
type DebouncedMessage struct {
	MessageID   uuid.UUID              `json:"message_id"`
	Text        string                 `json:"text"`
	Type        string                 `json:"type"` // text, audio, image, document, etc
	Timestamp   int64                  `json:"timestamp"`
	FromContact bool                   `json:"from_contact"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// FlushCallback é chamado quando um batch de mensagens é processado
type FlushCallback func(sessionID uuid.UUID, messages []DebouncedMessage) error

// MessageDebouncer agrupa mensagens rápidas e fragmentadas usando Redis
type MessageDebouncer struct {
	redis          *redis.Client
	windowDuration time.Duration // tempo de espera para agrupar mensagens
	maxBatchSize   int           // número máximo de mensagens antes de forçar flush
	flushCallback  FlushCallback
	keyPrefix      string
	ttl            time.Duration // TTL das chaves Redis
}

// NewMessageDebouncer cria um novo debouncer
func NewMessageDebouncer(
	redisClient *redis.Client,
	windowDuration time.Duration,
	maxBatchSize int,
	callback FlushCallback,
) *MessageDebouncer {
	if windowDuration == 0 {
		windowDuration = 2 * time.Second // padrão: 2s
	}
	if maxBatchSize == 0 {
		maxBatchSize = 10 // padrão: 10 mensagens
	}

	return &MessageDebouncer{
		redis:          redisClient,
		windowDuration: windowDuration,
		maxBatchSize:   maxBatchSize,
		flushCallback:  callback,
		keyPrefix:      "debouncer:session:",
		ttl:            5 * time.Minute,
	}
}

// AddMessage adiciona uma mensagem ao buffer da sessão
func (d *MessageDebouncer) AddMessage(
	ctx context.Context,
	sessionID uuid.UUID,
	msg DebouncedMessage,
) error {
	// Serializa mensagem
	msgJSON, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// Chave do sorted set: debouncer:session:{uuid}:messages
	key := d.getMessagesKey(sessionID)

	// Adiciona ao sorted set com timestamp como score
	if err := d.redis.ZAdd(ctx, key, redis.Z{
		Score:  float64(msg.Timestamp),
		Member: msgJSON,
	}).Err(); err != nil {
		return fmt.Errorf("failed to add message to buffer: %w", err)
	}

	// Atualiza TTL
	d.redis.Expire(ctx, key, d.ttl)

	// Atualiza timestamp da última mensagem
	lastMsgKey := d.getLastMessageTimestampKey(sessionID)
	d.redis.Set(ctx, lastMsgKey, time.Now().Unix(), d.ttl)

	fmt.Printf("📥 [Debouncer] Buffered message: session=%s, type=%s, count=%d\n",
		sessionID.String()[:8], msg.Type, d.getBufferSize(ctx, sessionID))

	// Verifica se deve fazer flush por tamanho
	if d.getBufferSize(ctx, sessionID) >= d.maxBatchSize {
		fmt.Printf("🚀 [Debouncer] Max batch size reached, flushing: session=%s\n", sessionID.String()[:8])
		return d.flush(ctx, sessionID)
	}

	return nil
}

// CheckAndFlush verifica se deve fazer flush baseado no tempo
func (d *MessageDebouncer) CheckAndFlush(ctx context.Context, sessionID uuid.UUID) error {
	lastMsgKey := d.getLastMessageTimestampKey(sessionID)
	lastMsgTimestamp, err := d.redis.Get(ctx, lastMsgKey).Int64()
	if err == redis.Nil {
		// Sem mensagens no buffer
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to get last message timestamp: %w", err)
	}

	lastMsgTime := time.Unix(lastMsgTimestamp, 0)
	age := time.Since(lastMsgTime)

	// Se passou o tempo da janela, faz flush
	if age >= d.windowDuration {
		fmt.Printf("⏰ [Debouncer] Window elapsed, flushing: session=%s, age=%.1fs\n",
			sessionID.String()[:8], age.Seconds())
		return d.flush(ctx, sessionID)
	}

	return nil
}

// flush processa todas as mensagens do buffer
func (d *MessageDebouncer) flush(ctx context.Context, sessionID uuid.UUID) error {
	key := d.getMessagesKey(sessionID)

	// Busca todas as mensagens do sorted set (ordenadas por timestamp)
	results, err := d.redis.ZRangeWithScores(ctx, key, 0, -1).Result()
	if err != nil {
		return fmt.Errorf("failed to get messages from buffer: %w", err)
	}

	if len(results) == 0 {
		return nil // buffer vazio
	}

	// Deserializa mensagens
	messages := make([]DebouncedMessage, 0, len(results))
	for _, result := range results {
		var msg DebouncedMessage
		if err := json.Unmarshal([]byte(result.Member.(string)), &msg); err != nil {
			fmt.Printf("⚠️  [Debouncer] Failed to unmarshal message: %v\n", err)
			continue
		}
		messages = append(messages, msg)
	}

	if len(messages) == 0 {
		return nil
	}

	// Chama callback com as mensagens
	if d.flushCallback != nil {
		if err := d.flushCallback(sessionID, messages); err != nil {
			return fmt.Errorf("flush callback failed: %w", err)
		}
	}

	// Limpa buffer após flush bem-sucedido
	d.redis.Del(ctx, key)
	d.redis.Del(ctx, d.getLastMessageTimestampKey(sessionID))

	fmt.Printf("✅ [Debouncer] Flushed session=%s, messages=%d\n",
		sessionID.String()[:8], len(messages))

	return nil
}

// StartFlushWorker inicia um worker que monitora sessões e faz flush periódico
func (d *MessageDebouncer) StartFlushWorker(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Second) // verifica a cada 1s
	defer ticker.Stop()

	fmt.Println("🔄 [Debouncer] Flush worker started")

	for {
		select {
		case <-ctx.Done():
			fmt.Println("🛑 [Debouncer] Flush worker stopped")
			return
		case <-ticker.C:
			// Busca todas as chaves de sessão
			pattern := d.keyPrefix + "*:messages"
			keys, err := d.redis.Keys(ctx, pattern).Result()
			if err != nil {
				fmt.Printf("⚠️  [Debouncer] Failed to list keys: %v\n", err)
				continue
			}

			// Verifica cada sessão
			for _, key := range keys {
				sessionID, err := d.extractSessionID(key)
				if err != nil {
					continue
				}

				if err := d.CheckAndFlush(ctx, sessionID); err != nil {
					fmt.Printf("⚠️  [Debouncer] Failed to flush session %s: %v\n",
						sessionID.String()[:8], err)
				}
			}
		}
	}
}

// ForceFlush força o flush de uma sessão específica
func (d *MessageDebouncer) ForceFlush(ctx context.Context, sessionID uuid.UUID) error {
	fmt.Printf("🔥 [Debouncer] Force flushing session=%s\n", sessionID.String()[:8])
	return d.flush(ctx, sessionID)
}

// GetBufferedMessages retorna as mensagens no buffer sem removê-las
func (d *MessageDebouncer) GetBufferedMessages(
	ctx context.Context,
	sessionID uuid.UUID,
) ([]DebouncedMessage, error) {
	key := d.getMessagesKey(sessionID)

	results, err := d.redis.ZRangeWithScores(ctx, key, 0, -1).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get messages: %w", err)
	}

	messages := make([]DebouncedMessage, 0, len(results))
	for _, result := range results {
		var msg DebouncedMessage
		if err := json.Unmarshal([]byte(result.Member.(string)), &msg); err != nil {
			continue
		}
		messages = append(messages, msg)
	}

	return messages, nil
}

// ConcatenateMessages concatena as mensagens em texto simples
func ConcatenateMessages(messages []DebouncedMessage) string {
	if len(messages) == 0 {
		return ""
	}

	// Ordena por timestamp (garantia extra)
	sort.Slice(messages, func(i, j int) bool {
		return messages[i].Timestamp < messages[j].Timestamp
	})

	var result string
	for i, msg := range messages {
		if i > 0 {
			result += "\n"
		}
		result += msg.Text
	}

	return result
}

// ConcatenateMessagesJSON retorna as mensagens em formato JSON estruturado
func ConcatenateMessagesJSON(messages []DebouncedMessage) (string, error) {
	if len(messages) == 0 {
		return "{}", nil
	}

	// Ordena por timestamp
	sort.Slice(messages, func(i, j int) bool {
		return messages[i].Timestamp < messages[j].Timestamp
	})

	data, err := json.Marshal(map[string]interface{}{
		"messages": messages,
		"count":    len(messages),
	})
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// Helper methods

func (d *MessageDebouncer) getMessagesKey(sessionID uuid.UUID) string {
	return fmt.Sprintf("%s%s:messages", d.keyPrefix, sessionID.String())
}

func (d *MessageDebouncer) getLastMessageTimestampKey(sessionID uuid.UUID) string {
	return fmt.Sprintf("%s%s:last_timestamp", d.keyPrefix, sessionID.String())
}

func (d *MessageDebouncer) getBufferSize(ctx context.Context, sessionID uuid.UUID) int {
	key := d.getMessagesKey(sessionID)
	count, err := d.redis.ZCard(ctx, key).Result()
	if err != nil {
		return 0
	}
	return int(count)
}

func (d *MessageDebouncer) extractSessionID(key string) (uuid.UUID, error) {
	// Extrai UUID da chave: "debouncer:session:{uuid}:messages"
	start := len(d.keyPrefix)
	end := len(key) - len(":messages")
	if end <= start {
		return uuid.Nil, fmt.Errorf("invalid key format")
	}

	uuidStr := key[start:end]
	return uuid.Parse(uuidStr)
}
