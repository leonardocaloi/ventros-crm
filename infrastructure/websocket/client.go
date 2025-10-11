package websocket

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

const (
	// Time allowed to write a message to the peer
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer (1MB)
	maxMessageSize = 1024 * 1024
)

// Client representa um cliente WebSocket conectado
type Client struct {
	hub  *Hub
	conn *websocket.Conn
	send chan []byte

	// Identificação
	id        string
	userID    uuid.UUID
	tenantID  string
	projectID uuid.UUID

	// Sessões que o cliente está observando
	sessions map[uuid.UUID]bool
	mu       sync.RWMutex

	logger *zap.Logger
	ctx    context.Context
	cancel context.CancelFunc
}

// NewClient cria novo cliente WebSocket
func NewClient(hub *Hub, conn *websocket.Conn, userID uuid.UUID, tenantID string, projectID uuid.UUID, logger *zap.Logger) *Client {
	ctx, cancel := context.WithCancel(context.Background())

	return &Client{
		hub:       hub,
		conn:      conn,
		send:      make(chan []byte, 256),
		id:        uuid.New().String(),
		userID:    userID,
		tenantID:  tenantID,
		projectID: projectID,
		sessions:  make(map[uuid.UUID]bool),
		logger:    logger,
		ctx:       ctx,
		cancel:    cancel,
	}
}

// ReadPump bombeia mensagens do cliente WebSocket para o hub
func (c *Client) ReadPump() {
	defer func() {
		c.hub.Unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.logger.Error("WebSocket read error",
					zap.String("client_id", c.id),
					zap.Error(err))
			}
			break
		}

		// Processar mensagem
		if err := c.handleMessage(message); err != nil {
			c.logger.Error("Failed to handle message",
				zap.String("client_id", c.id),
				zap.Error(err))

			// Enviar erro para o cliente
			errorMsg := NewErrorMessage("message_error", err.Error())
			c.SendMessage(errorMsg)
		}
	}
}

// WritePump bombeia mensagens do hub para o cliente WebSocket
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// Hub fechou o canal
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Adicionar mensagens enfileiradas ao frame atual
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}

		case <-c.ctx.Done():
			return
		}
	}
}

// handleMessage processa mensagem recebida do cliente
func (c *Client) handleMessage(data []byte) error {
	// Parse mensagem
	msg, err := FromJSON(data)
	if err != nil {
		return err
	}

	c.logger.Debug("Message received from client",
		zap.String("client_id", c.id),
		zap.String("message_type", string(msg.Type)))

	// Processar por tipo
	switch msg.Type {
	case MessageTypeJoinSession:
		return c.handleJoinSession(msg)
	case MessageTypeLeaveSession:
		return c.handleLeaveSession(msg)
	case MessageTypeSendMessage:
		return c.handleSendMessage(msg)
	case MessageTypeTyping:
		return c.handleTyping(msg)
	case MessageTypePing:
		return c.handlePing()
	default:
		c.logger.Warn("Unknown message type",
			zap.String("type", string(msg.Type)))
	}

	return nil
}

// handleJoinSession adiciona cliente a uma sessão
func (c *Client) handleJoinSession(msg *WSMessage) error {
	var payload JoinSessionPayload
	if err := msg.ParsePayload(&payload); err != nil {
		return err
	}

	c.mu.Lock()
	c.sessions[payload.SessionID] = true
	c.mu.Unlock()

	c.logger.Info("Client joined session",
		zap.String("client_id", c.id),
		zap.String("session_id", payload.SessionID.String()))

	return nil
}

// handleLeaveSession remove cliente de uma sessão
func (c *Client) handleLeaveSession(msg *WSMessage) error {
	var payload JoinSessionPayload
	if err := msg.ParsePayload(&payload); err != nil {
		return err
	}

	c.mu.Lock()
	delete(c.sessions, payload.SessionID)
	c.mu.Unlock()

	c.logger.Info("Client left session",
		zap.String("client_id", c.id),
		zap.String("session_id", payload.SessionID.String()))

	return nil
}

// handleSendMessage processa envio de mensagem
func (c *Client) handleSendMessage(msg *WSMessage) error {
	var payload SendMessagePayload
	if err := msg.ParsePayload(&payload); err != nil {
		return err
	}

	// Sanitizar texto (anti-XSS)
	payload.Text = SanitizeText(payload.Text)

	// SECURITY: Validar que usuário tem acesso a essa sessão
	// TODO: Implementar validação de permissão

	// Publicar mensagem via hub
	c.hub.HandleSendMessage(c, payload)

	return nil
}

// handleTyping processa notificação de digitação
func (c *Client) handleTyping(msg *WSMessage) error {
	var payload TypingPayload
	if err := msg.ParsePayload(&payload); err != nil {
		return err
	}

	// Broadcast para outros clientes na mesma sessão
	c.hub.BroadcastToSession(payload.SessionID, NewWSMessage(MessageTypeUserTyping, payload), c.id)

	return nil
}

// handlePing responde ao ping
func (c *Client) handlePing() error {
	c.SendMessage(NewWSMessage(MessageTypePong, nil))
	return nil
}

// SendMessage envia mensagem para o cliente
func (c *Client) SendMessage(msg *WSMessage) {
	data, err := msg.ToJSON()
	if err != nil {
		c.logger.Error("Failed to serialize message",
			zap.Error(err))
		return
	}

	select {
	case c.send <- data:
	case <-c.ctx.Done():
		return
	default:
		// Canal cheio, remover cliente
		c.hub.Unregister <- c
	}
}

// IsWatchingSession verifica se cliente está observando uma sessão
func (c *Client) IsWatchingSession(sessionID uuid.UUID) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.sessions[sessionID]
}

// Close fecha o cliente
func (c *Client) Close() {
	c.cancel()
	close(c.send)
}

// ID retorna o ID do cliente
func (c *Client) ID() string {
	return c.id
}

// UserID retorna o ID do usuário
func (c *Client) UserID() uuid.UUID {
	return c.userID
}

// TenantID retorna o tenant ID
func (c *Client) TenantID() string {
	return c.tenantID
}

// ProjectID retorna o project ID
func (c *Client) ProjectID() uuid.UUID {
	return c.projectID
}
