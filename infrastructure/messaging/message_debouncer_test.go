package messaging

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestRedis(t *testing.T) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   15, // usa DB separado para testes
	})

	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		t.Skipf("Redis not available: %v", err)
	}

	// Limpa DB de teste
	client.FlushDB(ctx)

	return client
}

func TestMessageDebouncer_AddMessage(t *testing.T) {
	redisClient := setupTestRedis(t)
	defer redisClient.Close()

	ctx := context.Background()
	sessionID := uuid.New()

	var flushedMessages []DebouncedMessage
	callback := func(sid uuid.UUID, msgs []DebouncedMessage) error {
		flushedMessages = msgs
		return nil
	}

	debouncer := NewMessageDebouncer(redisClient, 2*time.Second, 5, callback)

	// Adiciona mensagem
	msg := DebouncedMessage{
		MessageID:   uuid.New(),
		Text:        "Hello",
		Type:        "text",
		Timestamp:   time.Now().UnixMilli(),
		FromContact: true,
		Metadata:    map[string]interface{}{},
	}

	err := debouncer.AddMessage(ctx, sessionID, msg)
	require.NoError(t, err)

	// Verifica que está no buffer
	buffered, err := debouncer.GetBufferedMessages(ctx, sessionID)
	require.NoError(t, err)
	assert.Len(t, buffered, 1)
	assert.Equal(t, "Hello", buffered[0].Text)
}

func TestMessageDebouncer_FlushOnMaxBatchSize(t *testing.T) {
	redisClient := setupTestRedis(t)
	defer redisClient.Close()

	ctx := context.Background()
	sessionID := uuid.New()

	var flushedMessages []DebouncedMessage
	callback := func(sid uuid.UUID, msgs []DebouncedMessage) error {
		flushedMessages = msgs
		return nil
	}

	// Max batch size = 3
	debouncer := NewMessageDebouncer(redisClient, 10*time.Second, 3, callback)

	// Adiciona 3 mensagens
	for i := 0; i < 3; i++ {
		msg := DebouncedMessage{
			MessageID:   uuid.New(),
			Text:        "Message " + string(rune('1'+i)),
			Type:        "text",
			Timestamp:   time.Now().UnixMilli() + int64(i),
			FromContact: true,
			Metadata:    map[string]interface{}{},
		}
		err := debouncer.AddMessage(ctx, sessionID, msg)
		require.NoError(t, err)
	}

	// Deve ter feito flush automaticamente
	assert.Len(t, flushedMessages, 3)

	// Buffer deve estar vazio
	buffered, err := debouncer.GetBufferedMessages(ctx, sessionID)
	require.NoError(t, err)
	assert.Len(t, buffered, 0)
}

func TestMessageDebouncer_FlushOnTimeout(t *testing.T) {
	redisClient := setupTestRedis(t)
	defer redisClient.Close()

	ctx := context.Background()
	sessionID := uuid.New()

	var flushedMessages []DebouncedMessage
	callback := func(sid uuid.UUID, msgs []DebouncedMessage) error {
		flushedMessages = msgs
		return nil
	}

	// Window de 100ms
	debouncer := NewMessageDebouncer(redisClient, 100*time.Millisecond, 10, callback)

	// Adiciona mensagem
	msg := DebouncedMessage{
		MessageID:   uuid.New(),
		Text:        "Test",
		Type:        "text",
		Timestamp:   time.Now().UnixMilli(),
		FromContact: true,
		Metadata:    map[string]interface{}{},
	}
	err := debouncer.AddMessage(ctx, sessionID, msg)
	require.NoError(t, err)

	// Aguarda tempo da janela
	time.Sleep(150 * time.Millisecond)

	// Verifica e faz flush
	err = debouncer.CheckAndFlush(ctx, sessionID)
	require.NoError(t, err)

	// Deve ter feito flush
	assert.Len(t, flushedMessages, 1)
}

func TestMessageDebouncer_ForceFlush(t *testing.T) {
	redisClient := setupTestRedis(t)
	defer redisClient.Close()

	ctx := context.Background()
	sessionID := uuid.New()

	var flushedMessages []DebouncedMessage
	callback := func(sid uuid.UUID, msgs []DebouncedMessage) error {
		flushedMessages = msgs
		return nil
	}

	debouncer := NewMessageDebouncer(redisClient, 10*time.Second, 10, callback)

	// Adiciona mensagens
	for i := 0; i < 2; i++ {
		msg := DebouncedMessage{
			MessageID:   uuid.New(),
			Text:        "Message " + string(rune('1'+i)),
			Type:        "text",
			Timestamp:   time.Now().UnixMilli() + int64(i),
			FromContact: true,
			Metadata:    map[string]interface{}{},
		}
		err := debouncer.AddMessage(ctx, sessionID, msg)
		require.NoError(t, err)
	}

	// Force flush imediato
	err := debouncer.ForceFlush(ctx, sessionID)
	require.NoError(t, err)

	// Deve ter feito flush
	assert.Len(t, flushedMessages, 2)
}

func TestMessageDebouncer_ConcatenateMessages(t *testing.T) {
	messages := []DebouncedMessage{
		{
			MessageID: uuid.New(),
			Text:      "Hello",
			Timestamp: 1000,
		},
		{
			MessageID: uuid.New(),
			Text:      "World",
			Timestamp: 2000,
		},
		{
			MessageID: uuid.New(),
			Text:      "!",
			Timestamp: 3000,
		},
	}

	result := ConcatenateMessages(messages)
	assert.Equal(t, "Hello\nWorld\n!", result)
}

func TestMessageDebouncer_ConcatenateMessagesJSON(t *testing.T) {
	messages := []DebouncedMessage{
		{
			MessageID: uuid.New(),
			Text:      "Hello",
			Type:      "text",
			Timestamp: 1000,
		},
	}

	result, err := ConcatenateMessagesJSON(messages)
	require.NoError(t, err)
	assert.Contains(t, result, "Hello")
	assert.Contains(t, result, "count")
}

func TestMessageDebouncer_MessageOrdering(t *testing.T) {
	redisClient := setupTestRedis(t)
	defer redisClient.Close()

	ctx := context.Background()
	sessionID := uuid.New()

	var flushedMessages []DebouncedMessage
	callback := func(sid uuid.UUID, msgs []DebouncedMessage) error {
		flushedMessages = msgs
		return nil
	}

	debouncer := NewMessageDebouncer(redisClient, 2*time.Second, 5, callback)

	// Adiciona mensagens fora de ordem (pode acontecer em sistemas distribuídos)
	messages := []DebouncedMessage{
		{MessageID: uuid.New(), Text: "Third", Timestamp: 3000},
		{MessageID: uuid.New(), Text: "First", Timestamp: 1000},
		{MessageID: uuid.New(), Text: "Second", Timestamp: 2000},
	}

	for _, msg := range messages {
		err := debouncer.AddMessage(ctx, sessionID, msg)
		require.NoError(t, err)
	}

	// Force flush
	err := debouncer.ForceFlush(ctx, sessionID)
	require.NoError(t, err)

	// Verifica ordenação por timestamp
	require.Len(t, flushedMessages, 3)
	assert.Equal(t, "First", flushedMessages[0].Text)
	assert.Equal(t, "Second", flushedMessages[1].Text)
	assert.Equal(t, "Third", flushedMessages[2].Text)
}
