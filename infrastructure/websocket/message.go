package websocket

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// MessageType define o tipo de mensagem WebSocket
type MessageType string

const (
	// Client → Server
	MessageTypeSendMessage  MessageType = "send_message"  // Enviar nova mensagem
	MessageTypeTyping       MessageType = "typing"        // Usuário está digitando
	MessageTypeJoinSession  MessageType = "join_session"  // Entrar em uma sessão
	MessageTypeLeaveSession MessageType = "leave_session" // Sair de uma sessão
	MessageTypePing         MessageType = "ping"          // Heartbeat

	// Server → Client
	MessageTypeNewMessage  MessageType = "new_message"  // Nova mensagem recebida
	MessageTypeMessageSent MessageType = "message_sent" // Confirmação de envio
	MessageTypeMessageRead MessageType = "message_read" // Mensagem lida
	MessageTypeUserTyping  MessageType = "user_typing"  // Outro usuário digitando
	MessageTypeError       MessageType = "error"        // Erro
	MessageTypePong        MessageType = "pong"         // Resposta ao ping
	MessageTypeConnected   MessageType = "connected"    // Conexão estabelecida
)

// WSMessage representa uma mensagem WebSocket
type WSMessage struct {
	Type      MessageType `json:"type"`
	Payload   interface{} `json:"payload,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
	MessageID string      `json:"message_id,omitempty"`
	Error     string      `json:"error,omitempty"`
}

// SendMessagePayload para enviar mensagens
type SendMessagePayload struct {
	SessionID   uuid.UUID  `json:"session_id"`
	ContactID   uuid.UUID  `json:"contact_id"`
	Text        string     `json:"text"`
	ContentType string     `json:"content_type"`
	ReplyToID   *uuid.UUID `json:"reply_to_id,omitempty"`
}

// NewMessagePayload para notificar novas mensagens
type NewMessagePayload struct {
	MessageID   uuid.UUID  `json:"message_id"`
	SessionID   uuid.UUID  `json:"session_id"`
	ContactID   uuid.UUID  `json:"contact_id"`
	Text        string     `json:"text"`
	ContentType string     `json:"content_type"`
	FromMe      bool       `json:"from_me"`
	Timestamp   time.Time  `json:"timestamp"`
	AgentID     *uuid.UUID `json:"agent_id,omitempty"`
	Status      string     `json:"status"`
}

// TypingPayload para notificações de digitação
type TypingPayload struct {
	SessionID uuid.UUID `json:"session_id"`
	ContactID uuid.UUID `json:"contact_id"`
	IsTyping  bool      `json:"is_typing"`
	UserID    uuid.UUID `json:"user_id,omitempty"`
	UserName  string    `json:"user_name,omitempty"`
}

// JoinSessionPayload para entrar em uma sessão
type JoinSessionPayload struct {
	SessionID uuid.UUID `json:"session_id"`
}

// ErrorPayload para erros
type ErrorPayload struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// ConnectedPayload para confirmação de conexão
type ConnectedPayload struct {
	ClientID  string    `json:"client_id"`
	UserID    uuid.UUID `json:"user_id"`
	Timestamp time.Time `json:"timestamp"`
}

// NewWSMessage cria nova mensagem WebSocket
func NewWSMessage(msgType MessageType, payload interface{}) *WSMessage {
	return &WSMessage{
		Type:      msgType,
		Payload:   payload,
		Timestamp: time.Now(),
		MessageID: uuid.New().String(),
	}
}

// NewErrorMessage cria mensagem de erro
func NewErrorMessage(code, message string) *WSMessage {
	return &WSMessage{
		Type:      MessageTypeError,
		Timestamp: time.Now(),
		MessageID: uuid.New().String(),
		Payload: ErrorPayload{
			Code:    code,
			Message: message,
		},
	}
}

// ToJSON converte mensagem para JSON
func (m *WSMessage) ToJSON() ([]byte, error) {
	return json.Marshal(m)
}

// FromJSON parse mensagem de JSON
func FromJSON(data []byte) (*WSMessage, error) {
	var msg WSMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		return nil, err
	}
	return &msg, nil
}

// ParsePayload extrai payload tipado
func (m *WSMessage) ParsePayload(target interface{}) error {
	payloadJSON, err := json.Marshal(m.Payload)
	if err != nil {
		return err
	}
	return json.Unmarshal(payloadJSON, target)
}
