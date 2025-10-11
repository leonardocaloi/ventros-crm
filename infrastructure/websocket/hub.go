package websocket

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// Hub mantém conjunto de clientes ativos e broadcast de mensagens via Redis Pub/Sub
type Hub struct {
	// Clientes registrados
	clients map[*Client]bool
	mu      sync.RWMutex

	// Índice: sessionID → [clients]
	sessionClients map[uuid.UUID]map[*Client]bool
	sessionMu      sync.RWMutex

	// Canais de comunicação
	Register   chan *Client
	Unregister chan *Client
	Broadcast  chan *BroadcastMessage

	// Redis Pub/Sub para multi-server
	redis  *redis.Client
	pubsub *redis.PubSub

	// Message handler (integração com domain)
	messageHandler MessageHandler

	logger *zap.Logger
	ctx    context.Context
	cancel context.CancelFunc
}

// BroadcastMessage mensagem para broadcast
type BroadcastMessage struct {
	SessionID       uuid.UUID
	Message         *WSMessage
	ExcludeClientID string // Não enviar para este cliente
}

// MessageHandler processa mensagens de negócio
type MessageHandler interface {
	// SendMessage persiste e envia mensagem
	SendMessage(ctx context.Context, userID uuid.UUID, payload SendMessagePayload) (*NewMessagePayload, error)
}

// NewHub cria novo Hub WebSocket
func NewHub(redis *redis.Client, messageHandler MessageHandler, logger *zap.Logger) *Hub {
	ctx, cancel := context.WithCancel(context.Background())

	return &Hub{
		clients:        make(map[*Client]bool),
		sessionClients: make(map[uuid.UUID]map[*Client]bool),
		Register:       make(chan *Client),
		Unregister:     make(chan *Client),
		Broadcast:      make(chan *BroadcastMessage, 256),
		redis:          redis,
		messageHandler: messageHandler,
		logger:         logger,
		ctx:            ctx,
		cancel:         cancel,
	}
}

// Run inicia o hub
func (h *Hub) Run() {
	// Iniciar Redis Pub/Sub
	if h.redis != nil {
		h.startRedisPubSub()
	}

	for {
		select {
		case client := <-h.Register:
			h.registerClient(client)

		case client := <-h.Unregister:
			h.unregisterClient(client)

		case broadcast := <-h.Broadcast:
			h.broadcastMessage(broadcast)

		case <-h.ctx.Done():
			h.logger.Info("WebSocket hub stopping")
			return
		}
	}
}

// registerClient registra novo cliente
func (h *Hub) registerClient(client *Client) {
	h.mu.Lock()
	h.clients[client] = true
	h.mu.Unlock()

	h.logger.Info("Client registered",
		zap.String("client_id", client.id),
		zap.String("user_id", client.userID.String()),
		zap.Int("total_clients", len(h.clients)))

	// Enviar mensagem de boas-vindas
	welcomeMsg := NewWSMessage(MessageTypeConnected, ConnectedPayload{
		ClientID:  client.id,
		UserID:    client.userID,
		Timestamp: time.Now(),
	})
	client.SendMessage(welcomeMsg)
}

// unregisterClient remove cliente
func (h *Hub) unregisterClient(client *Client) {
	h.mu.Lock()
	if _, ok := h.clients[client]; ok {
		delete(h.clients, client)
		client.Close()
	}
	h.mu.Unlock()

	// Remover de todas as sessões
	h.sessionMu.Lock()
	for sessionID, clients := range h.sessionClients {
		delete(clients, client)
		if len(clients) == 0 {
			delete(h.sessionClients, sessionID)
		}
	}
	h.sessionMu.Unlock()

	h.logger.Info("Client unregistered",
		zap.String("client_id", client.id),
		zap.Int("total_clients", len(h.clients)))
}

// broadcastMessage envia mensagem para clientes
func (h *Hub) broadcastMessage(broadcast *BroadcastMessage) {
	h.sessionMu.RLock()
	clients := h.sessionClients[broadcast.SessionID]
	h.sessionMu.RUnlock()

	for client := range clients {
		// Não enviar para cliente que originou mensagem
		if client.id == broadcast.ExcludeClientID {
			continue
		}

		// Verificar se cliente está observando essa sessão
		if client.IsWatchingSession(broadcast.SessionID) {
			client.SendMessage(broadcast.Message)
		}
	}
}

// BroadcastToSession envia mensagem para todos clientes de uma sessão
func (h *Hub) BroadcastToSession(sessionID uuid.UUID, msg *WSMessage, excludeClientID string) {
	h.Broadcast <- &BroadcastMessage{
		SessionID:       sessionID,
		Message:         msg,
		ExcludeClientID: excludeClientID,
	}
}

// HandleSendMessage processa envio de mensagem
func (h *Hub) HandleSendMessage(client *Client, payload SendMessagePayload) {
	// Delegar para message handler (domain logic)
	newMsg, err := h.messageHandler.SendMessage(h.ctx, client.userID, payload)
	if err != nil {
		h.logger.Error("Failed to send message",
			zap.Error(err),
			zap.String("client_id", client.id))

		// Enviar erro para cliente
		client.SendMessage(NewErrorMessage("send_failed", err.Error()))
		return
	}

	// Confirmar envio para o cliente que enviou
	client.SendMessage(NewWSMessage(MessageTypeMessageSent, newMsg))

	// Broadcast para outros clientes na sessão
	broadcastMsg := NewWSMessage(MessageTypeNewMessage, newMsg)
	h.BroadcastToSession(payload.SessionID, broadcastMsg, client.id)

	// Publicar no Redis para outros servidores
	if h.redis != nil {
		h.publishToRedis(payload.SessionID, broadcastMsg)
	}
}

// startRedisPubSub inicia Redis Pub/Sub para multi-server
func (h *Hub) startRedisPubSub() {
	channel := "websocket:messages"
	h.pubsub = h.redis.Subscribe(h.ctx, channel)

	h.logger.Info("Redis Pub/Sub started",
		zap.String("channel", channel))

	// Goroutine para receber mensagens do Redis
	go func() {
		for {
			select {
			case <-h.ctx.Done():
				h.pubsub.Close()
				return

			case msg := <-h.pubsub.Channel():
				h.handleRedisMessage(msg.Payload)
			}
		}
	}()
}

// handleRedisMessage processa mensagem recebida do Redis
func (h *Hub) handleRedisMessage(payload string) {
	var redisMsg RedisMessage
	if err := json.Unmarshal([]byte(payload), &redisMsg); err != nil {
		h.logger.Error("Failed to parse Redis message", zap.Error(err))
		return
	}

	// Broadcast para clientes locais
	h.broadcastMessage(&BroadcastMessage{
		SessionID: redisMsg.SessionID,
		Message:   redisMsg.WSMessage,
	})
}

// publishToRedis publica mensagem no Redis para outros servidores
func (h *Hub) publishToRedis(sessionID uuid.UUID, msg *WSMessage) {
	redisMsg := RedisMessage{
		SessionID: sessionID,
		WSMessage: msg,
		Timestamp: time.Now(),
	}

	data, err := json.Marshal(redisMsg)
	if err != nil {
		h.logger.Error("Failed to marshal Redis message", zap.Error(err))
		return
	}

	channel := "websocket:messages"
	if err := h.redis.Publish(h.ctx, channel, data).Err(); err != nil {
		h.logger.Error("Failed to publish to Redis",
			zap.Error(err),
			zap.String("channel", channel))
	}
}

// RedisMessage mensagem enviada via Redis Pub/Sub
type RedisMessage struct {
	SessionID uuid.UUID  `json:"session_id"`
	WSMessage *WSMessage `json:"ws_message"`
	Timestamp time.Time  `json:"timestamp"`
}

// GetStats retorna estatísticas do hub
func (h *Hub) GetStats() map[string]interface{} {
	h.mu.RLock()
	totalClients := len(h.clients)
	h.mu.RUnlock()

	h.sessionMu.RLock()
	totalSessions := len(h.sessionClients)
	h.sessionMu.RUnlock()

	return map[string]interface{}{
		"total_clients":  totalClients,
		"total_sessions": totalSessions,
		"redis_enabled":  h.redis != nil,
	}
}

// Shutdown para o hub gracefully
func (h *Hub) Shutdown() error {
	h.cancel()

	// Fechar todos os clientes
	h.mu.Lock()
	for client := range h.clients {
		client.Close()
	}
	h.mu.Unlock()

	// Fechar Redis Pub/Sub
	if h.pubsub != nil {
		return h.pubsub.Close()
	}

	return nil
}

// NotifyNewMessage notifica nova mensagem recebida externamente (webhook, etc)
// Usado quando mensagem chega via WAHA ou outro canal
func (h *Hub) NotifyNewMessage(sessionID uuid.UUID, msg *NewMessagePayload) {
	wsMsg := NewWSMessage(MessageTypeNewMessage, msg)

	// Broadcast local
	h.BroadcastToSession(sessionID, wsMsg, "")

	// Broadcast via Redis
	if h.redis != nil {
		h.publishToRedis(sessionID, wsMsg)
	}
}

// NotifyMessageRead notifica que mensagem foi lida
func (h *Hub) NotifyMessageRead(sessionID, messageID uuid.UUID) {
	payload := map[string]interface{}{
		"message_id": messageID.String(),
		"read_at":    time.Now(),
	}

	wsMsg := NewWSMessage(MessageTypeMessageRead, payload)

	// Broadcast local
	h.BroadcastToSession(sessionID, wsMsg, "")

	// Broadcast via Redis
	if h.redis != nil {
		h.publishToRedis(sessionID, wsMsg)
	}
}
