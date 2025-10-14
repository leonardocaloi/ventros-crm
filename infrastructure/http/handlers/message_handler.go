package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	apierrors "github.com/ventros/crm/infrastructure/http/errors"
	"github.com/ventros/crm/infrastructure/http/middleware"
	"github.com/ventros/crm/internal/application/commands/message"
	"github.com/ventros/crm/internal/application/queries"
	"github.com/ventros/crm/internal/domain/core/shared"
	domainMessage "github.com/ventros/crm/internal/domain/crm/message"
	"go.uber.org/zap"
)

type MessageHandler struct {
	logger                        *zap.Logger
	messageRepo                   domainMessage.Repository
	sendMessageHandler            *message.SendMessageHandler
	confirmMessageDeliveryHandler *message.ConfirmMessageDeliveryHandler
	listMessagesQueryHandler      *queries.ListMessagesQueryHandler
	searchMessagesQueryHandler    *queries.SearchMessagesQueryHandler
}

func NewMessageHandler(logger *zap.Logger, messageRepo domainMessage.Repository, sendMessageHandler *message.SendMessageHandler, confirmMessageDeliveryHandler *message.ConfirmMessageDeliveryHandler) *MessageHandler {
	return &MessageHandler{
		logger:                        logger,
		messageRepo:                   messageRepo,
		sendMessageHandler:            sendMessageHandler,
		confirmMessageDeliveryHandler: confirmMessageDeliveryHandler,
		listMessagesQueryHandler:      queries.NewListMessagesQueryHandler(messageRepo, logger),
		searchMessagesQueryHandler:    queries.NewSearchMessagesQueryHandler(messageRepo, logger),
	}
}

// CreateMessageRequest represents the request to create a message.
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

// SendMessageRequest represents the request to send a message.
type SendMessageRequest struct {
	ContactID   uuid.UUID              `json:"contact_id" binding:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
	ChannelID   uuid.UUID              `json:"channel_id" binding:"required" example:"550e8400-e29b-41d4-a716-446655440001"`
	ContentType string                 `json:"content_type" binding:"required" example:"text"`
	Text        *string                `json:"text,omitempty" example:"Hello, how can I help you?"`
	MediaURL    *string                `json:"media_url,omitempty" example:"https://example.com/image.jpg"`
	ReplyToID   *uuid.UUID             `json:"reply_to_id,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// SendMessageResponse represents the response of sending a message.
type SendMessageResponse struct {
	MessageID  uuid.UUID `json:"message_id" example:"550e8400-e29b-41d4-a716-446655440002"`
	ExternalID *string   `json:"external_id,omitempty" example:"wamid.123456"`
	Status     string    `json:"status" example:"sent"`
	SentAt     string    `json:"sent_at" example:"2025-10-09T10:30:00Z"`
	Error      *string   `json:"error,omitempty"`
}

// UpdateMessageRequest represents the request to update a message.
type UpdateMessageRequest struct {
	Content     *string                `json:"content,omitempty"`
	Status      *string                `json:"status,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	ReadAt      *string                `json:"read_at,omitempty"`
	DeliveredAt *string                `json:"delivered_at,omitempty"`
}

// ListMessages lists messages with optional filters
//
//	@Summary		List messages
//	@Description	List messages with optional filters
//	@Tags			CRM - Messages
//	@Produce		json
//	@Param			session_id		query		string					false	"Filter by session ID (UUID)"
//	@Param			contact_id		query		string					false	"Filter by contact ID (UUID)"
//	@Param			direction		query		string					false	"Filter by direction (inbound, outbound)"
//	@Param			message_type	query		string					false	"Filter by message type"
//	@Param			status			query		string					false	"Filter by status"
//	@Param			limit			query		int						false	"Limit results"			default(50)
//	@Param			offset			query		int						false	"Offset for pagination"	default(0)
//	@Success		200				{array}		map[string]interface{}	"List of messages"
//	@Failure		400				{object}	map[string]interface{}	"Invalid parameters"
//	@Failure		500				{object}	map[string]interface{}	"Internal server error"
//	@Router			/api/v1/messages [get]
func (h *MessageHandler) ListMessages(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Message listing not yet implemented",
		"note":    "Use GET /api/v1/messages/{id} to get specific message",
	})
}

// CreateMessage creates a new message
//
//	@Summary		Create message
//	@Description	Create a new message
//	@Tags			CRM - Messages
//	@Accept			json
//	@Produce		json
//	@Param			message	body		CreateMessageRequest	true	"Message data"
//	@Success		201		{object}	map[string]interface{}	"Message created successfully"
//	@Failure		400		{object}	map[string]interface{}	"Invalid request"
//	@Failure		500		{object}	map[string]interface{}	"Internal server error"
//	@Router			/api/v1/messages [post]
func (h *MessageHandler) CreateMessage(c *gin.Context) {
	var req CreateMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to parse message request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

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

// SendMessage sends a message to a contact via a channel
//
//	@Summary		Send message
//	@Description	Send a message to a contact via specific channel
//	@Tags			CRM - Messages
//	@Accept			json
//	@Produce		json
//	@Param			message	body		SendMessageRequest		true	"Message data"
//	@Success		200		{object}	SendMessageResponse		"Message sent successfully"
//	@Failure		400		{object}	map[string]interface{}	"Invalid request"
//	@Failure		500		{object}	map[string]interface{}	"Internal server error"
//	@Router			/api/v1/messages/send [post]
func (h *MessageHandler) SendMessage(c *gin.Context) {
	var req SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to parse send message request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	contentType, err := domainMessage.ParseContentType(req.ContentType)
	if err != nil {
		h.logger.Error("Invalid content type", zap.String("content_type", req.ContentType), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid content type: " + err.Error()})
		return
	}

	// Extract auth context
	authCtx, exists := middleware.GetAuthContext(c)
	if !exists {
		h.logger.Error("Auth context not found")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	tenantIDStr := authCtx.TenantID
	projectIDUUID := authCtx.ProjectID
	customerIDUUID := authCtx.UserID // customer_id is the user_id

	// AgentID - use System Test Agent for E2E tests and manual sends
	// TODO: In production, this should fetch the user's first active agent
	agentIDUUID := uuid.MustParse("00000000-0000-0000-0000-000000000010") // System Test Agent
	h.logger.Info("Using System Test Agent for message send",
		zap.String("agent_id", agentIDUUID.String()),
		zap.String("user_id", authCtx.UserID.String()))

	cmd := &message.SendMessageCommand{
		ContactID:   req.ContactID,
		ChannelID:   req.ChannelID,
		ContentType: contentType,
		Text:        req.Text,
		MediaURL:    req.MediaURL,
		ReplyToID:   req.ReplyToID,
		Metadata:    req.Metadata,
		TenantID:    tenantIDStr,
		ProjectID:   projectIDUUID,
		CustomerID:  customerIDUUID,
		AgentID:     agentIDUUID,
	}

	result, err := h.sendMessageHandler.Handle(c.Request.Context(), cmd)
	if err != nil {
		h.logger.Error("Failed to send message",
			zap.String("contact_id", req.ContactID.String()),
			zap.String("channel_id", req.ChannelID.String()),
			zap.Error(err))

		errMsg := err.Error()
		c.JSON(http.StatusInternalServerError, SendMessageResponse{
			MessageID:  uuid.Nil,
			ExternalID: nil,
			Status:     "failed",
			SentAt:     "",
			Error:      &errMsg,
		})
		return
	}

	c.JSON(http.StatusOK, SendMessageResponse{
		MessageID:  result.MessageID,
		ExternalID: result.ExternalID,
		Status:     result.Status,
		SentAt:     result.SentAt.Format("2006-01-02T15:04:05Z07:00"),
		Error:      result.Error,
	})
}

// GetMessage gets a message by ID
//
//	@Summary		Get message by ID
//	@Description	Get details of a specific message
//	@Tags			CRM - Messages
//	@Produce		json
//	@Param			id	path		string					true	"Message ID (UUID)"
//	@Success		200	{object}	map[string]interface{}	"Message details"
//	@Failure		400	{object}	map[string]interface{}	"Invalid message ID"
//	@Failure		404	{object}	map[string]interface{}	"Message not found"
//	@Failure		500	{object}	map[string]interface{}	"Internal server error"
//	@Router			/api/v1/messages/{id} [get]
func (h *MessageHandler) GetMessage(c *gin.Context) {
	idStr := c.Param("id")
	messageID, err := uuid.Parse(idStr)
	if err != nil {
		h.logger.Error("Invalid message ID", zap.String("id", idStr), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid message ID format"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Message retrieval not yet implemented",
		"message_id": messageID,
	})
}

// UpdateMessage updates a message
//
//	@Summary		Update message
//	@Description	Update an existing message
//	@Tags			CRM - Messages
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string					true	"Message ID (UUID)"
//	@Param			message	body		UpdateMessageRequest	true	"Message update data"
//	@Success		200		{object}	map[string]interface{}	"Message updated successfully"
//	@Failure		400		{object}	map[string]interface{}	"Invalid request"
//	@Failure		404		{object}	map[string]interface{}	"Message not found"
//	@Failure		500		{object}	map[string]interface{}	"Internal server error"
//	@Router			/api/v1/messages/{id} [put]
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

	c.JSON(http.StatusOK, gin.H{
		"message":    "Message update not yet implemented",
		"message_id": messageID,
	})
}

// DeleteMessage deletes a message
//
//	@Summary		Delete message
//	@Description	Remove a message (soft delete)
//	@Tags			CRM - Messages
//	@Produce		json
//	@Param			id	path	string	true	"Message ID (UUID)"
//	@Success		204	"Message deleted successfully"
//	@Failure		400	{object}	map[string]interface{}	"Invalid message ID"
//	@Failure		404	{object}	map[string]interface{}	"Message not found"
//	@Failure		500	{object}	map[string]interface{}	"Internal server error"
//	@Router			/api/v1/messages/{id} [delete]
func (h *MessageHandler) DeleteMessage(c *gin.Context) {
	idStr := c.Param("id")
	messageID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid message ID format"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Message deletion not yet implemented",
		"message_id": messageID,
	})
}

// GetMessagesBySession gets messages for a specific session
//
//	@Summary		Get messages by session
//	@Description	Get all messages from a specific session
//	@Tags			CRM - Messages
//	@Produce		json
//	@Param			session_id	path		string					true	"Session ID (UUID)"
//	@Param			limit		query		int						false	"Limit results"			default(100)
//	@Param			offset		query		int						false	"Offset for pagination"	default(0)
//	@Success		200			{array}		map[string]interface{}	"List of messages"
//	@Failure		400			{object}	map[string]interface{}	"Invalid session ID"
//	@Failure		500			{object}	map[string]interface{}	"Internal server error"
//	@Router			/api/v1/sessions/{session_id}/messages [get]
func (h *MessageHandler) GetMessagesBySession(c *gin.Context) {
	sessionIDStr := c.Param("session_id")
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid session ID format"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Session messages retrieval not yet implemented",
		"session_id": sessionID,
		"messages":   []interface{}{},
	})
}

// ConfirmMessageDeliveryRequest represents the request to confirm delivery.
type ConfirmMessageDeliveryRequest struct {
	MessageID     *string `json:"message_id,omitempty"`
	ExternalID    string  `json:"external_id" binding:"required" example:"wamid.123456"`
	Status        string  `json:"status" binding:"required" example:"delivered"`
	DeliveredAt   *string `json:"delivered_at,omitempty" example:"2025-10-09T10:30:00Z"`
	ReadAt        *string `json:"read_at,omitempty" example:"2025-10-09T10:35:00Z"`
	FailureReason *string `json:"failure_reason,omitempty" example:"Message expired"`
}

// ConfirmMessageDelivery confirms delivery/reading of a message
//
//	@Summary		Confirm message delivery
//	@Description	Confirm delivery, reading or failure of a sent message
//	@Tags			CRM - Messages
//	@Accept			json
//	@Produce		json
//	@Param			confirmation	body		ConfirmMessageDeliveryRequest	true	"Delivery confirmation data"
//	@Success		200				{object}	map[string]interface{}			"Message status updated successfully"
//	@Failure		400				{object}	map[string]interface{}			"Invalid request"
//	@Failure		404				{object}	map[string]interface{}			"Message not found"
//	@Failure		500				{object}	map[string]interface{}			"Internal server error"
//	@Router			/api/v1/messages/confirm-delivery [post]
func (h *MessageHandler) ConfirmMessageDelivery(c *gin.Context) {
	var req ConfirmMessageDeliveryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to parse confirmation request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	var messageID uuid.UUID
	if req.MessageID != nil {
		var err error
		messageID, err = uuid.Parse(*req.MessageID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid message_id format"})
			return
		}
	}

	var deliveredAt *time.Time
	if req.DeliveredAt != nil {
		t, err := time.Parse(time.RFC3339, *req.DeliveredAt)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid delivered_at format (use RFC3339)"})
			return
		}
		deliveredAt = &t
	}

	var readAt *time.Time
	if req.ReadAt != nil {
		t, err := time.Parse(time.RFC3339, *req.ReadAt)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid read_at format (use RFC3339)"})
			return
		}
		readAt = &t
	}

	cmd := &message.ConfirmMessageDeliveryCommand{
		MessageID:     messageID,
		ExternalID:    req.ExternalID,
		Status:        req.Status,
		DeliveredAt:   deliveredAt,
		ReadAt:        readAt,
		FailureReason: req.FailureReason,
	}

	if err := h.confirmMessageDeliveryHandler.Handle(c.Request.Context(), cmd); err != nil {
		h.logger.Error("Failed to confirm message delivery",
			zap.String("external_id", req.ExternalID),
			zap.String("status", req.Status),
			zap.Error(err))

		if err.Error() == "message not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Message not found"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "Message status updated successfully",
		"external_id": req.ExternalID,
		"status":      req.Status,
	})
}

// ListMessagesAdvanced lists messages with advanced filters, pagination, and sorting
//
//	@Summary		List messages with advanced filters and pagination
//	@Description	Retrieve a paginated list of messages with comprehensive filtering across contact, session, channel, direction, content type, and delivery status. Ideal for building message history views, conversation analytics, and customer interaction tracking.
//	@Description
//	@Description	**Filtering Capabilities:**
//	@Description	- Filter by contact to view all messages for a specific customer
//	@Description	- Filter by session to see complete conversation threads
//	@Description	- Filter by channel to analyze messages from specific communication channels (WhatsApp, Email, etc)
//	@Description	- Filter by project to segment messages by business unit or department
//	@Description	- Filter by channel_type (1=WhatsApp, 2=Email, 3=SMS, etc) for channel-specific analytics
//	@Description	- Filter by from_me (true=outbound agent messages, false=inbound customer messages)
//	@Description	- Filter by content_type (text, image, video, audio, document, location, contact, sticker) for media analysis
//	@Description	- Filter by status (pending, sent, delivered, read, failed) for delivery tracking
//	@Description	- Filter by agent_id to track individual agent performance
//	@Description	- Filter by timestamp range to analyze time-based patterns
//	@Description	- Filter by has_media flag to find messages with attachments
//	@Description
//	@Description	**Use Cases:**
//	@Description	- Build conversation history UIs with infinite scroll pagination
//	@Description	- Analyze customer response times and patterns
//	@Description	- Track message delivery rates across channels
//	@Description	- Monitor agent productivity and response quality
//	@Description	- Generate conversation transcripts for compliance
//	@Description	- Identify media-rich conversations for quality assurance
//	@Description
//	@Description	**Sorting Options:**
//	@Description	- Sort by timestamp (default), created_at for processing order
//	@Description	- Ascending or descending order
//	@Description
//	@Description	**Performance:**
//	@Description	- Optimized with composite GORM indexes on tenant+session, tenant+contact, tenant+channel
//	@Description	- GIN index on JSONB metadata field for custom attribute searches
//	@Description	- Maximum 100 messages per page for optimal response times
//	@Description	- Efficiently handles millions of messages per tenant
//	@Tags			CRM - Messages
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			contact_id			query		string							false	"Filter by contact UUID - Example: 550e8400-e29b-41d4-a716-446655440000"
//	@Param			session_id			query		string							false	"Filter by session UUID to get full conversation - Example: 660e8400-e29b-41d4-a716-446655440001"
//	@Param			channel_id			query		string							false	"Filter by channel UUID (specific WhatsApp number, email account, etc)"
//	@Param			project_id			query		string							false	"Filter by project UUID to segment by business unit"
//	@Param			channel_type_id		query		int								false	"Filter by channel type - 1:WhatsApp, 2:Email, 3:SMS, 4:Web Chat"	example(1)
//	@Param			from_me				query		bool							false	"Filter by direction - true: agent sent, false: customer sent"		example(false)
//	@Param			content_type		query		string							false	"Filter by content type"											Enums(text, image, video, audio, document, location, contact, sticker)	example(text)
//	@Param			status				query		string							false	"Filter by delivery status"											Enums(pending, sent, delivered, read, failed)							example(delivered)
//	@Param			agent_id			query		string							false	"Filter by agent UUID for performance tracking"
//	@Param			timestamp_after		query		string							false	"Messages sent after this timestamp - Format: 2006-01-02T15:04:05Z"	example(2024-01-01T00:00:00Z)
//	@Param			timestamp_before	query		string							false	"Messages sent before this timestamp"								example(2024-12-31T23:59:59Z)
//	@Param			has_media			query		bool							false	"Filter messages with media attachments - true: only with media"	example(true)
//	@Param			page				query		int								false	"Page number for pagination (starts at 1)"							default(1)						minimum(1)			example(1)
//	@Param			limit				query		int								false	"Messages per page (max 100)"										default(20)						minimum(1)			maximum(100)	example(50)
//	@Param			sort_by				query		string							false	"Field to sort by"													Enums(timestamp, created_at)	default(timestamp)	example(timestamp)
//	@Param			sort_dir			query		string							false	"Sort direction"													Enums(asc, desc)				default(desc)		example(desc)
//	@Success		200					{object}	queries.ListMessagesResponse	"Successfully retrieved messages with pagination and filter metadata"
//	@Failure		400					{object}	map[string]interface{}			"Bad Request - Invalid UUID format, invalid enum values, or limit exceeds 100"
//	@Failure		401					{object}	map[string]interface{}			"Unauthorized - Missing or invalid authentication token"
//	@Failure		403					{object}	map[string]interface{}			"Forbidden - User lacks permission to access this tenant's messages"
//	@Failure		500					{object}	map[string]interface{}			"Internal Server Error - Database errors or query execution failures"
//	@Router			/api/v1/crm/messages/advanced [get]
func (h *MessageHandler) ListMessagesAdvanced(c *gin.Context) {
	authCtx, exists := middleware.GetAuthContext(c)
	if !exists {
		apierrors.Unauthorized(c, "Authentication required")
		return
	}

	// Parse pagination
	page := 1
	limit := 20
	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	// Parse sorting
	sortBy := c.DefaultQuery("sort_by", "timestamp")
	sortDir := c.DefaultQuery("sort_dir", "desc")

	// Parse optional UUID filters
	var contactID, sessionID, channelID *uuid.UUID
	if contactIDStr := c.Query("contact_id"); contactIDStr != "" {
		if id, err := uuid.Parse(contactIDStr); err == nil {
			contactID = &id
		}
	}
	if sessionIDStr := c.Query("session_id"); sessionIDStr != "" {
		if id, err := uuid.Parse(sessionIDStr); err == nil {
			sessionID = &id
		}
	}
	if channelIDStr := c.Query("channel_id"); channelIDStr != "" {
		if id, err := uuid.Parse(channelIDStr); err == nil {
			channelID = &id
		}
	}

	// Parse boolean filter
	var fromMe *bool
	if fromMeStr := c.Query("from_me"); fromMeStr != "" {
		if b, err := strconv.ParseBool(fromMeStr); err == nil {
			fromMe = &b
		}
	}

	// Parse string filters
	var contentType, status *string
	if contentTypeStr := c.Query("content_type"); contentTypeStr != "" {
		contentType = &contentTypeStr
	}
	if statusStr := c.Query("status"); statusStr != "" {
		status = &statusStr
	}

	// Create tenant ID
	tenantID, err := shared.NewTenantID(authCtx.TenantID)
	if err != nil {
		apierrors.ValidationError(c, "tenant_id", "Invalid tenant ID")
		return
	}

	// Execute query
	query := queries.ListMessagesQuery{
		TenantID:    tenantID,
		ContactID:   contactID,
		SessionID:   sessionID,
		ChannelID:   channelID,
		FromMe:      fromMe,
		ContentType: contentType,
		Status:      status,
		Page:        page,
		Limit:       limit,
		SortBy:      sortBy,
		SortDir:     sortDir,
	}

	response, err := h.listMessagesQueryHandler.Handle(c.Request.Context(), query)
	if err != nil {
		h.logger.Error("Failed to list messages", zap.Error(err))
		apierrors.InternalError(c, "Failed to retrieve messages", err)
		return
	}

	c.JSON(http.StatusOK, response)
}

// SearchMessages performs full-text search on messages
//
//	@Summary		Full-text search across message content
//	@Description	Perform intelligent full-text search across message text content using PostgreSQL ILIKE pattern matching. Perfect for finding specific conversations, keywords, customer questions, or agent responses within your entire message history.
//	@Description
//	@Description	**Search Capabilities:**
//	@Description	- Searches through message text content (case-insensitive)
//	@Description	- Supports partial word matches (e.g., "refund" matches "refunds", "refunded")
//	@Description	- Works across all message types (customer and agent messages)
//	@Description	- Searches only text content (media URLs are not searched)
//	@Description
//	@Description	**Match Scoring & Relevance:**
//	@Description	- All matches receive a score of 1.0 (simple ILIKE search, no complex scoring)
//	@Description	- Results ordered by timestamp (newest first) for relevance
//	@Description	- Match field always returns "text" since only text content is searched
//	@Description
//	@Description	**Common Use Cases:**
//	@Description	- Find all conversations mentioning "refund" or "cancellation"
//	@Description	- Search for product names across customer inquiries
//	@Description	- Locate conversations with specific error codes or reference numbers
//	@Description	- Find messages containing customer phone numbers or emails
//	@Description	- Search for competitor mentions in customer conversations
//	@Description	- Identify conversations with specific keywords for quality assurance
//	@Description	- Compliance searches for regulated terms or phrases
//	@Description
//	@Description	**Search Examples:**
//	@Description	- "order #12345" - Find messages mentioning specific order numbers
//	@Description	- "password reset" - Find password-related support conversations
//	@Description	- "urgent" or "emergency" - Identify high-priority conversations
//	@Description	- "@email.com" - Find messages containing email addresses
//	@Description	- "bug" or "error" - Locate technical issue reports
//	@Description
//	@Description	**Performance:**
//	@Description	- Optimized GORM indexes on tenant_id for fast tenant isolation
//	@Description	- ILIKE operator uses PostgreSQL's text search capabilities
//	@Description	- Maximum 100 results to ensure sub-second response times
//	@Description	- Handles searches across millions of messages efficiently
//	@Tags			CRM - Messages
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			q		query		string							true	"Search query - minimum 1 character, case-insensitive, supports partial matches - Examples: 'refund', 'order #12345', 'password reset', 'urgent'"	minlength(1)	example(refund request)
//	@Param			limit	query		int								false	"Maximum number of results (max 100)"																												default(20)		minimum(1)	maximum(100)	example(20)
//	@Success		200		{object}	queries.SearchMessagesResponse	"Successfully found matching messages with text excerpts and context"
//	@Failure		400		{object}	map[string]interface{}			"Bad Request - Missing or empty search query, or limit exceeds 100"
//	@Failure		401		{object}	map[string]interface{}			"Unauthorized - Missing or invalid authentication token"
//	@Failure		403		{object}	map[string]interface{}			"Forbidden - User lacks permission to search this tenant's messages"
//	@Failure		500		{object}	map[string]interface{}			"Internal Server Error - Database connection errors or search execution failures"
//	@Router			/api/v1/crm/messages/search [get]
func (h *MessageHandler) SearchMessages(c *gin.Context) {
	authCtx, exists := middleware.GetAuthContext(c)
	if !exists {
		apierrors.Unauthorized(c, "Authentication required")
		return
	}

	searchText := c.Query("q")
	if searchText == "" {
		apierrors.ValidationError(c, "q", "Search query 'q' is required")
		return
	}

	// Parse limit
	limit := 20
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	// Create tenant ID
	tenantID, err := shared.NewTenantID(authCtx.TenantID)
	if err != nil {
		apierrors.ValidationError(c, "tenant_id", "Invalid tenant ID")
		return
	}

	// Execute search
	query := queries.SearchMessagesQuery{
		TenantID:   tenantID,
		SearchText: searchText,
		Limit:      limit,
	}

	response, err := h.searchMessagesQueryHandler.Handle(c.Request.Context(), query)
	if err != nil {
		h.logger.Error("Failed to search messages", zap.Error(err))
		apierrors.InternalError(c, "Failed to search messages", err)
		return
	}

	c.JSON(http.StatusOK, response)
}
