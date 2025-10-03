package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type MessageHandler struct {
	logger *zap.Logger
}

func NewMessageHandler(logger *zap.Logger) *MessageHandler {
	return &MessageHandler{
		logger: logger,
	}
}

// CreateMessageRequest representa o payload para criar uma mensagem
type CreateMessageRequest struct {
	SessionID     uuid.UUID              `json:"session_id" binding:"required"`
	ContactID     uuid.UUID              `json:"contact_id" binding:"required"`
	Content       string                 `json:"content" binding:"required" example:"Olá, como posso ajudar?"`
	MessageType   string                 `json:"message_type" example:"text"`
	Direction     string                 `json:"direction" example:"inbound"`
	ExternalID    string                 `json:"external_id" example:"msg_123"`
	ReplyToID     *uuid.UUID             `json:"reply_to_id,omitempty"`
	MediaURL      string                 `json:"media_url,omitempty"`
	MediaType     string                 `json:"media_type,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
	ChannelTypeID uuid.UUID              `json:"channel_type_id" binding:"required"`
}

// UpdateMessageRequest representa o payload para atualizar uma mensagem
type UpdateMessageRequest struct {
	Content   *string                `json:"content,omitempty"`
	Status    *string                `json:"status,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	ReadAt    *string                `json:"read_at,omitempty"`
	DeliveredAt *string              `json:"delivered_at,omitempty"`
}

// ListMessages lists messages with optional filters
// @Summary List messages
// @Description Lista mensagens com filtros opcionais
// @Tags messages
// @Produce json
// @Param session_id query string false "Filter by session ID (UUID)"
// @Param contact_id query string false "Filter by contact ID (UUID)"
// @Param direction query string false "Filter by direction (inbound, outbound)"
// @Param message_type query string false "Filter by message type"
// @Param status query string false "Filter by status"
// @Param limit query int false "Limit results" default(50)
// @Param offset query int false "Offset for pagination" default(0)
// @Success 200 {array} map[string]interface{} "List of messages"
// @Failure 400 {object} map[string]interface{} "Invalid parameters"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/messages [get]
func (h *MessageHandler) ListMessages(c *gin.Context) {
	// TODO: Implement proper message listing with filters
	c.JSON(http.StatusOK, gin.H{
		"message": "Message listing not yet implemented",
		"note":    "Use GET /api/v1/messages/{id} to get specific message",
	})
}

// CreateMessage creates a new message
// @Summary Create message
// @Description Cria uma nova mensagem
// @Tags messages
// @Accept json
// @Produce json
// @Param message body CreateMessageRequest true "Message data"
// @Success 201 {object} map[string]interface{} "Message created successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/messages [post]
func (h *MessageHandler) CreateMessage(c *gin.Context) {
	var req CreateMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to parse message request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// TODO: Implement message creation
	c.JSON(http.StatusCreated, gin.H{
		"message":         "Message creation not yet implemented",
		"session_id":      req.SessionID,
		"contact_id":      req.ContactID,
		"content":         req.Content,
		"message_type":    req.MessageType,
		"direction":       req.Direction,
		"channel_type_id": req.ChannelTypeID,
	})
}

// GetMessage gets a message by ID
// @Summary Get message by ID
// @Description Obtém detalhes de uma mensagem específica
// @Tags messages
// @Produce json
// @Param id path string true "Message ID (UUID)"
// @Success 200 {object} map[string]interface{} "Message details"
// @Failure 400 {object} map[string]interface{} "Invalid message ID"
// @Failure 404 {object} map[string]interface{} "Message not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/messages/{id} [get]
func (h *MessageHandler) GetMessage(c *gin.Context) {
	idStr := c.Param("id")
	messageID, err := uuid.Parse(idStr)
	if err != nil {
		h.logger.Error("Invalid message ID", zap.String("id", idStr), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid message ID format"})
		return
	}

	// TODO: Implement message retrieval
	c.JSON(http.StatusOK, gin.H{
		"message":    "Message retrieval not yet implemented",
		"message_id": messageID,
	})
}

// UpdateMessage updates a message
// @Summary Update message
// @Description Atualiza uma mensagem existente
// @Tags messages
// @Accept json
// @Produce json
// @Param id path string true "Message ID (UUID)"
// @Param message body UpdateMessageRequest true "Message update data"
// @Success 200 {object} map[string]interface{} "Message updated successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Failure 404 {object} map[string]interface{} "Message not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/messages/{id} [put]
func (h *MessageHandler) UpdateMessage(c *gin.Context) {
	idStr := c.Param("id")
	messageID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid message ID format"})
		return
	}

	var req UpdateMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to parse update request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// TODO: Implement message update
	c.JSON(http.StatusOK, gin.H{
		"message":    "Message update not yet implemented",
		"message_id": messageID,
	})
}

// DeleteMessage deletes a message
// @Summary Delete message
// @Description Remove uma mensagem (soft delete)
// @Tags messages
// @Produce json
// @Param id path string true "Message ID (UUID)"
// @Success 204 "Message deleted successfully"
// @Failure 400 {object} map[string]interface{} "Invalid message ID"
// @Failure 404 {object} map[string]interface{} "Message not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/messages/{id} [delete]
func (h *MessageHandler) DeleteMessage(c *gin.Context) {
	idStr := c.Param("id")
	messageID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid message ID format"})
		return
	}

	// TODO: Implement message deletion
	c.JSON(http.StatusOK, gin.H{
		"message":    "Message deletion not yet implemented",
		"message_id": messageID,
	})
}

// GetMessagesBySession gets messages for a specific session
// @Summary Get messages by session
// @Description Obtém todas as mensagens de uma sessão específica
// @Tags messages
// @Produce json
// @Param session_id path string true "Session ID (UUID)"
// @Param limit query int false "Limit results" default(100)
// @Param offset query int false "Offset for pagination" default(0)
// @Success 200 {array} map[string]interface{} "List of messages"
// @Failure 400 {object} map[string]interface{} "Invalid session ID"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/sessions/{session_id}/messages [get]
func (h *MessageHandler) GetMessagesBySession(c *gin.Context) {
	sessionIDStr := c.Param("session_id")
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid session ID format"})
		return
	}

	// TODO: Implement session messages retrieval
	c.JSON(http.StatusOK, gin.H{
		"message":    "Session messages retrieval not yet implemented",
		"session_id": sessionID,
		"messages":   []interface{}{},
	})
}
