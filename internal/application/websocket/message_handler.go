package websocket

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	ws "github.com/ventros/crm/infrastructure/websocket"
	"github.com/ventros/crm/internal/domain/crm/message"
	"go.uber.org/zap"
)

// WebSocketMessageHandler implementa ws.MessageHandler
// Integra WebSocket com domain Message
type WebSocketMessageHandler struct {
	messageRepo message.Repository
	logger      *zap.Logger
}

// NewWebSocketMessageHandler cria novo handler
func NewWebSocketMessageHandler(messageRepo message.Repository, logger *zap.Logger) *WebSocketMessageHandler {
	return &WebSocketMessageHandler{
		messageRepo: messageRepo,
		logger:      logger,
	}
}

// SendMessage implementa ws.MessageHandler.SendMessage
func (h *WebSocketMessageHandler) SendMessage(ctx context.Context, userID uuid.UUID, payload ws.SendMessagePayload) (*ws.NewMessagePayload, error) {
	h.logger.Debug("Sending message via WebSocket",
		zap.String("user_id", userID.String()),
		zap.String("session_id", payload.SessionID.String()),
		zap.String("contact_id", payload.ContactID.String()))

	// TODO: Buscar customerID e projectID do contexto de autenticação
	// Por ora, usando valores do contexto (deveria vir do auth middleware)
	customerID := uuid.New() // FIXME: Pegar do contexto de auth
	projectID := uuid.New()  // FIXME: Pegar do contexto de auth

	// Converter ContentType string para domain type
	contentTypeStr := payload.ContentType
	if contentTypeStr == "" {
		contentTypeStr = "text" // Default
	}
	contentType, err := message.ParseContentType(contentTypeStr)
	if err != nil {
		return nil, fmt.Errorf("invalid content type: %w", err)
	}

	// Criar mensagem no domain
	msg, err := message.NewMessage(
		payload.ContactID,
		projectID,
		customerID,
		contentType,
		true, // fromMe = true (mensagem enviada pelo agente)
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	// Configurar mensagem
	if err := msg.SetText(payload.Text); err != nil {
		return nil, fmt.Errorf("failed to set text: %w", err)
	}

	msg.AssignToSession(payload.SessionID)

	// TODO: AssignToChannel - buscar channelID da sessão
	// TODO: Implementar método AssignToAgent no domain Message se necessário

	// Persistir mensagem
	if err := h.messageRepo.Save(ctx, msg); err != nil {
		return nil, fmt.Errorf("failed to save message: %w", err)
	}

	h.logger.Info("Message sent successfully",
		zap.String("message_id", msg.ID().String()),
		zap.String("session_id", payload.SessionID.String()),
		zap.String("contact_id", payload.ContactID.String()))

	// Converter para payload de resposta
	response := &ws.NewMessagePayload{
		MessageID:   msg.ID(),
		SessionID:   payload.SessionID,
		ContactID:   payload.ContactID,
		Text:        payload.Text,
		ContentType: payload.ContentType,
		FromMe:      true,
		Timestamp:   msg.Timestamp(),
		AgentID:     &userID,
		Status:      msg.Status().String(),
	}

	return response, nil
}
