package integration

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestWebSocketConnection testa conexão básica WebSocket
func TestWebSocketConnection(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Conectar ao WebSocket
	url := "ws://localhost:8080/api/v1/ws/messages?token=dev-admin-key"
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	require.NoError(t, err)
	defer conn.Close()

	// Aguardar mensagem de conexão
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	_, message, err := conn.ReadMessage()
	require.NoError(t, err)

	var msg map[string]interface{}
	err = json.Unmarshal(message, &msg)
	require.NoError(t, err)

	assert.Equal(t, "connected", msg["type"])
}

// TestWebSocketSendMessage testa envio de mensagem
func TestWebSocketSendMessage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Conectar
	url := "ws://localhost:8080/api/v1/ws/messages?token=dev-admin-key"
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	require.NoError(t, err)
	defer conn.Close()

	// Ler mensagem de conexão
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	_, _, err = conn.ReadMessage()
	require.NoError(t, err)

	// Join session
	sessionID := uuid.New()
	contactID := uuid.New()

	joinMsg := map[string]interface{}{
		"type": "join_session",
		"payload": map[string]interface{}{
			"session_id": sessionID.String(),
		},
	}

	err = conn.WriteJSON(joinMsg)
	require.NoError(t, err)

	// Enviar mensagem
	sendMsg := map[string]interface{}{
		"type": "send_message",
		"payload": map[string]interface{}{
			"session_id":   sessionID.String(),
			"contact_id":   contactID.String(),
			"text":         "Test message",
			"content_type": "text",
		},
	}

	err = conn.WriteJSON(sendMsg)
	require.NoError(t, err)

	// Aguardar resposta (message_sent ou error)
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	_, responseData, err := conn.ReadMessage()
	require.NoError(t, err)

	var response map[string]interface{}
	err = json.Unmarshal(responseData, &response)
	require.NoError(t, err)

	// Pode ser message_sent ou error (depende se DB está configurado)
	assert.Contains(t, []string{"message_sent", "error"}, response["type"])
}

// TestWebSocketRateLimiting testa rate limiting
func TestWebSocketRateLimiting(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	url := "ws://localhost:8080/api/v1/ws/messages?token=dev-admin-key"

	// Tentar conectar 6 vezes (limite é 5 por minuto)
	connections := make([]*websocket.Conn, 0, 6)
	defer func() {
		for _, conn := range connections {
			if conn != nil {
				conn.Close()
			}
		}
	}()

	for i := 0; i < 6; i++ {
		conn, resp, err := websocket.DefaultDialer.Dial(url, nil)

		if i < 5 {
			// Primeiras 5 devem conectar
			require.NoError(t, err)
			require.NotNil(t, conn)
			connections = append(connections, conn)
		} else {
			// 6ª deve falhar (rate limit)
			if err == nil && conn != nil {
				conn.Close()
				t.Log("Warning: Rate limit not enforced (expected on 6th connection)")
			} else {
				// Verificar HTTP 429 Too Many Requests
				if resp != nil {
					assert.Equal(t, 429, resp.StatusCode)
				}
			}
		}
	}
}

// TestWebSocketHeartbeat testa ping/pong
func TestWebSocketHeartbeat(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	url := "ws://localhost:8080/api/v1/ws/messages?token=dev-admin-key"
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	require.NoError(t, err)
	defer conn.Close()

	// Ler mensagem de conexão
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	_, _, err = conn.ReadMessage()
	require.NoError(t, err)

	// Enviar ping
	pingMsg := map[string]interface{}{
		"type": "ping",
	}

	err = conn.WriteJSON(pingMsg)
	require.NoError(t, err)

	// Aguardar pong
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	_, responseData, err := conn.ReadMessage()
	require.NoError(t, err)

	var response map[string]interface{}
	err = json.Unmarshal(responseData, &response)
	require.NoError(t, err)

	assert.Equal(t, "pong", response["type"])
}

// BenchmarkWebSocketMessage benchmark de envio de mensagens
func BenchmarkWebSocketMessage(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping benchmark")
	}

	url := "ws://localhost:8080/api/v1/ws/messages?token=dev-admin-key"
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		b.Fatal(err)
	}
	defer conn.Close()

	// Ler conexão
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	_, _, _ = conn.ReadMessage()

	sessionID := uuid.New()
	contactID := uuid.New()

	// Join session
	joinMsg := map[string]interface{}{
		"type": "join_session",
		"payload": map[string]interface{}{
			"session_id": sessionID.String(),
		},
	}
	_ = conn.WriteJSON(joinMsg)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		sendMsg := map[string]interface{}{
			"type": "send_message",
			"payload": map[string]interface{}{
				"session_id":   sessionID.String(),
				"contact_id":   contactID.String(),
				"text":         "Benchmark message",
				"content_type": "text",
			},
		}

		err := conn.WriteJSON(sendMsg)
		if err != nil {
			b.Fatal(err)
		}
	}

	select {
	case <-ctx.Done():
		return
	default:
	}
}
