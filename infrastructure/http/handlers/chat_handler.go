package handlers

import (
	"net/http"

	apierrors "github.com/caloi/ventros-crm/infrastructure/http/errors"
	"github.com/caloi/ventros-crm/infrastructure/http/middleware"
	chatapp "github.com/caloi/ventros-crm/internal/application/chat"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type ChatHandler struct {
	logger                    *zap.Logger
	createChatUseCase         *chatapp.CreateChatUseCase
	findChatUseCase           *chatapp.FindChatUseCase
	manageParticipantsUseCase *chatapp.ManageParticipantsUseCase
	archiveChatUseCase        *chatapp.ArchiveChatUseCase
	updateChatUseCase         *chatapp.UpdateChatUseCase
}

func NewChatHandler(
	logger *zap.Logger,
	createChatUseCase *chatapp.CreateChatUseCase,
	findChatUseCase *chatapp.FindChatUseCase,
	manageParticipantsUseCase *chatapp.ManageParticipantsUseCase,
	archiveChatUseCase *chatapp.ArchiveChatUseCase,
	updateChatUseCase *chatapp.UpdateChatUseCase,
) *ChatHandler {
	return &ChatHandler{
		logger:                    logger,
		createChatUseCase:         createChatUseCase,
		findChatUseCase:           findChatUseCase,
		manageParticipantsUseCase: manageParticipantsUseCase,
		archiveChatUseCase:        archiveChatUseCase,
		updateChatUseCase:         updateChatUseCase,
	}
}

// CreateChatRequest represents the request to create a chat
type CreateChatRequest struct {
	ChatType  string     `json:"chat_type" binding:"required" example:"individual"`
	ContactID *uuid.UUID `json:"contact_id,omitempty"`
	CreatorID *uuid.UUID `json:"creator_id,omitempty"`
	Subject   *string    `json:"subject,omitempty" example:"Team Discussion"`
}

// AddParticipantRequest represents the request to add a participant
type AddParticipantRequest struct {
	ParticipantID   uuid.UUID `json:"participant_id" binding:"required"`
	ParticipantType string    `json:"participant_type" binding:"required" example:"agent"`
}

// UpdateSubjectRequest represents the request to update chat subject
type UpdateSubjectRequest struct {
	Subject string `json:"subject" binding:"required" example:"New Team Discussion"`
}

// CreateChat creates a new chat
//
//	@Summary		Create a new chat
//	@Description	Create a new chat (individual, group, or channel)
//	@Tags			CRM - Chats
//	@Accept			json
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			project_id	query		string					true	"Project ID"
//	@Param			chat		body		CreateChatRequest		true	"Chat data"
//	@Success		201			{object}	map[string]interface{}	"Chat created successfully"
//	@Failure		400			{object}	map[string]interface{}	"Invalid request"
//	@Failure		401			{object}	map[string]interface{}	"Authentication required"
//	@Failure		500			{object}	map[string]interface{}	"Internal server error"
//	@Router			/api/v1/chats [post]
func (h *ChatHandler) CreateChat(c *gin.Context) {
	authCtx, exists := middleware.GetAuthContext(c)
	if !exists {
		apierrors.Unauthorized(c, "Authentication required")
		return
	}

	projectIDStr := c.Query("project_id")
	if projectIDStr == "" {
		apierrors.ValidationError(c, "project_id", "project_id query parameter is required")
		return
	}

	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		apierrors.ValidationError(c, "project_id", "Invalid project_id format (must be UUID)")
		return
	}

	var req CreateChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to parse chat request", zap.Error(err))
		apierrors.BadRequest(c, "Invalid request body: "+err.Error())
		return
	}

	// Build input
	input := chatapp.CreateChatInput{
		ProjectID: projectID,
		TenantID:  authCtx.TenantID,
		ChatType:  req.ChatType,
		ContactID: req.ContactID,
		CreatorID: req.CreatorID,
		Subject:   req.Subject,
	}

	// Execute use case
	output, err := h.createChatUseCase.Execute(c.Request.Context(), input)
	if err != nil {
		h.logger.Error("Failed to create chat", zap.Error(err))
		apierrors.RespondWithError(c, err)
		return
	}

	c.JSON(http.StatusCreated, output.Chat)
}

// GetChat gets a chat by ID
//
//	@Summary		Get chat by ID
//	@Description	Get detailed information about a specific chat
//	@Tags			CRM - Chats
//	@Accept			json
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			id	path		string					true	"Chat ID"
//	@Success		200	{object}	map[string]interface{}	"Chat details"
//	@Failure		400	{object}	map[string]interface{}	"Invalid chat ID"
//	@Failure		401	{object}	map[string]interface{}	"Authentication required"
//	@Failure		404	{object}	map[string]interface{}	"Chat not found"
//	@Failure		500	{object}	map[string]interface{}	"Internal server error"
//	@Router			/api/v1/chats/{id} [get]
func (h *ChatHandler) GetChat(c *gin.Context) {
	_, exists := middleware.GetAuthContext(c)
	if !exists {
		apierrors.Unauthorized(c, "Authentication required")
		return
	}

	idStr := c.Param("id")
	chatID, err := uuid.Parse(idStr)
	if err != nil {
		apierrors.ValidationError(c, "id", "Invalid chat ID format (must be UUID)")
		return
	}

	input := chatapp.FindChatInput{
		ChatID: chatID,
	}

	output, err := h.findChatUseCase.FindByID(c.Request.Context(), input)
	if err != nil {
		h.logger.Error("Failed to find chat", zap.Error(err))
		apierrors.RespondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, output.Chat)
}

// ListChats lists chats with filters
//
//	@Summary		List chats
//	@Description	List chats with optional filters (project_id, contact_id, status, chat_type)
//	@Tags			CRM - Chats
//	@Accept			json
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			project_id	query		string					false	"Filter by project ID"
//	@Param			contact_id	query		string					false	"Filter by contact ID (participant)"
//	@Param			status		query		string					false	"Filter by status (active, archived, closed)"
//	@Param			chat_type	query		string					false	"Filter by chat type (individual, group, channel)"
//	@Success		200			{object}	map[string]interface{}	"List of chats"
//	@Failure		400			{object}	map[string]interface{}	"Invalid parameters"
//	@Failure		401			{object}	map[string]interface{}	"Authentication required"
//	@Failure		500			{object}	map[string]interface{}	"Internal server error"
//	@Router			/api/v1/chats [get]
func (h *ChatHandler) ListChats(c *gin.Context) {
	authCtx, exists := middleware.GetAuthContext(c)
	if !exists {
		apierrors.Unauthorized(c, "Authentication required")
		return
	}

	input := chatapp.ListChatsInput{
		TenantID: &authCtx.TenantID,
	}

	// Parse filters
	if projectIDStr := c.Query("project_id"); projectIDStr != "" {
		projectID, err := uuid.Parse(projectIDStr)
		if err != nil {
			apierrors.ValidationError(c, "project_id", "Invalid project_id format (must be UUID)")
			return
		}
		input.ProjectID = &projectID
	}

	if contactIDStr := c.Query("contact_id"); contactIDStr != "" {
		contactID, err := uuid.Parse(contactIDStr)
		if err != nil {
			apierrors.ValidationError(c, "contact_id", "Invalid contact_id format (must be UUID)")
			return
		}
		input.ContactID = &contactID
	}

	if status := c.Query("status"); status != "" {
		input.Status = &status
	}

	if chatType := c.Query("chat_type"); chatType != "" {
		input.ChatType = &chatType
	}

	output, err := h.findChatUseCase.ListChats(c.Request.Context(), input)
	if err != nil {
		h.logger.Error("Failed to list chats", zap.Error(err))
		apierrors.RespondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"chats": output.Chats,
		"total": output.Total,
	})
}

// AddParticipant adds a participant to a chat
//
//	@Summary		Add participant to chat
//	@Description	Add a contact or agent as a participant to a chat
//	@Tags			CRM - Chats
//	@Accept			json
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			id			path		string					true	"Chat ID"
//	@Param			participant	body		AddParticipantRequest	true	"Participant data"
//	@Success		200			{object}	map[string]interface{}	"Participant added successfully"
//	@Failure		400			{object}	map[string]interface{}	"Invalid request"
//	@Failure		401			{object}	map[string]interface{}	"Authentication required"
//	@Failure		404			{object}	map[string]interface{}	"Chat not found"
//	@Failure		500			{object}	map[string]interface{}	"Internal server error"
//	@Router			/api/v1/chats/{id}/participants [post]
func (h *ChatHandler) AddParticipant(c *gin.Context) {
	_, exists := middleware.GetAuthContext(c)
	if !exists {
		apierrors.Unauthorized(c, "Authentication required")
		return
	}

	idStr := c.Param("id")
	chatID, err := uuid.Parse(idStr)
	if err != nil {
		apierrors.ValidationError(c, "id", "Invalid chat ID format (must be UUID)")
		return
	}

	var req AddParticipantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to parse add participant request", zap.Error(err))
		apierrors.BadRequest(c, "Invalid request body: "+err.Error())
		return
	}

	input := chatapp.AddParticipantInput{
		ChatID:          chatID,
		ParticipantID:   req.ParticipantID,
		ParticipantType: req.ParticipantType,
	}

	output, err := h.manageParticipantsUseCase.AddParticipant(c.Request.Context(), input)
	if err != nil {
		h.logger.Error("Failed to add participant", zap.Error(err))
		apierrors.RespondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, output.Chat)
}

// RemoveParticipant removes a participant from a chat
//
//	@Summary		Remove participant from chat
//	@Description	Remove a participant from a chat
//	@Tags			CRM - Chats
//	@Accept			json
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			id				path		string					true	"Chat ID"
//	@Param			participant_id	path		string					true	"Participant ID"
//	@Success		200				{object}	map[string]interface{}	"Participant removed successfully"
//	@Failure		400				{object}	map[string]interface{}	"Invalid request"
//	@Failure		401				{object}	map[string]interface{}	"Authentication required"
//	@Failure		404				{object}	map[string]interface{}	"Chat or participant not found"
//	@Failure		500				{object}	map[string]interface{}	"Internal server error"
//	@Router			/api/v1/chats/{id}/participants/{participant_id} [delete]
func (h *ChatHandler) RemoveParticipant(c *gin.Context) {
	_, exists := middleware.GetAuthContext(c)
	if !exists {
		apierrors.Unauthorized(c, "Authentication required")
		return
	}

	idStr := c.Param("id")
	chatID, err := uuid.Parse(idStr)
	if err != nil {
		apierrors.ValidationError(c, "id", "Invalid chat ID format (must be UUID)")
		return
	}

	participantIDStr := c.Param("participant_id")
	participantID, err := uuid.Parse(participantIDStr)
	if err != nil {
		apierrors.ValidationError(c, "participant_id", "Invalid participant_id format (must be UUID)")
		return
	}

	input := chatapp.RemoveParticipantInput{
		ChatID:        chatID,
		ParticipantID: participantID,
	}

	output, err := h.manageParticipantsUseCase.RemoveParticipant(c.Request.Context(), input)
	if err != nil {
		h.logger.Error("Failed to remove participant", zap.Error(err))
		apierrors.RespondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, output.Chat)
}

// ArchiveChat archives a chat
//
//	@Summary		Archive chat
//	@Description	Archive a chat (can be unarchived later)
//	@Tags			CRM - Chats
//	@Accept			json
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			id	path		string					true	"Chat ID"
//	@Success		200	{object}	map[string]interface{}	"Chat archived successfully"
//	@Failure		400	{object}	map[string]interface{}	"Invalid chat ID"
//	@Failure		401	{object}	map[string]interface{}	"Authentication required"
//	@Failure		404	{object}	map[string]interface{}	"Chat not found"
//	@Failure		500	{object}	map[string]interface{}	"Internal server error"
//	@Router			/api/v1/chats/{id}/archive [post]
func (h *ChatHandler) ArchiveChat(c *gin.Context) {
	_, exists := middleware.GetAuthContext(c)
	if !exists {
		apierrors.Unauthorized(c, "Authentication required")
		return
	}

	idStr := c.Param("id")
	chatID, err := uuid.Parse(idStr)
	if err != nil {
		apierrors.ValidationError(c, "id", "Invalid chat ID format (must be UUID)")
		return
	}

	input := chatapp.ArchiveChatInput{
		ChatID: chatID,
	}

	output, err := h.archiveChatUseCase.Archive(c.Request.Context(), input)
	if err != nil {
		h.logger.Error("Failed to archive chat", zap.Error(err))
		apierrors.RespondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, output.Chat)
}

// UnarchiveChat unarchives a chat
//
//	@Summary		Unarchive chat
//	@Description	Unarchive a chat (reactivate it)
//	@Tags			CRM - Chats
//	@Accept			json
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			id	path		string					true	"Chat ID"
//	@Success		200	{object}	map[string]interface{}	"Chat unarchived successfully"
//	@Failure		400	{object}	map[string]interface{}	"Invalid chat ID"
//	@Failure		401	{object}	map[string]interface{}	"Authentication required"
//	@Failure		404	{object}	map[string]interface{}	"Chat not found"
//	@Failure		500	{object}	map[string]interface{}	"Internal server error"
//	@Router			/api/v1/chats/{id}/unarchive [post]
func (h *ChatHandler) UnarchiveChat(c *gin.Context) {
	_, exists := middleware.GetAuthContext(c)
	if !exists {
		apierrors.Unauthorized(c, "Authentication required")
		return
	}

	idStr := c.Param("id")
	chatID, err := uuid.Parse(idStr)
	if err != nil {
		apierrors.ValidationError(c, "id", "Invalid chat ID format (must be UUID)")
		return
	}

	input := chatapp.UnarchiveChatInput{
		ChatID: chatID,
	}

	output, err := h.archiveChatUseCase.Unarchive(c.Request.Context(), input)
	if err != nil {
		h.logger.Error("Failed to unarchive chat", zap.Error(err))
		apierrors.RespondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, output.Chat)
}

// CloseChat permanently closes a chat
//
//	@Summary		Close chat
//	@Description	Permanently close a chat (cannot be reopened)
//	@Tags			CRM - Chats
//	@Accept			json
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			id	path		string					true	"Chat ID"
//	@Success		200	{object}	map[string]interface{}	"Chat closed successfully"
//	@Failure		400	{object}	map[string]interface{}	"Invalid chat ID"
//	@Failure		401	{object}	map[string]interface{}	"Authentication required"
//	@Failure		404	{object}	map[string]interface{}	"Chat not found"
//	@Failure		500	{object}	map[string]interface{}	"Internal server error"
//	@Router			/api/v1/chats/{id}/close [post]
func (h *ChatHandler) CloseChat(c *gin.Context) {
	_, exists := middleware.GetAuthContext(c)
	if !exists {
		apierrors.Unauthorized(c, "Authentication required")
		return
	}

	idStr := c.Param("id")
	chatID, err := uuid.Parse(idStr)
	if err != nil {
		apierrors.ValidationError(c, "id", "Invalid chat ID format (must be UUID)")
		return
	}

	input := chatapp.CloseChatInput{
		ChatID: chatID,
	}

	output, err := h.archiveChatUseCase.Close(c.Request.Context(), input)
	if err != nil {
		h.logger.Error("Failed to close chat", zap.Error(err))
		apierrors.RespondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, output.Chat)
}

// UpdateChatSubject updates the subject of a group or channel chat
//
//	@Summary		Update chat subject
//	@Description	Update the subject/name of a group or channel chat
//	@Tags			CRM - Chats
//	@Accept			json
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			id		path		string					true	"Chat ID"
//	@Param			subject	body		UpdateSubjectRequest	true	"New subject"
//	@Success		200		{object}	map[string]interface{}	"Subject updated successfully"
//	@Failure		400		{object}	map[string]interface{}	"Invalid request"
//	@Failure		401		{object}	map[string]interface{}	"Authentication required"
//	@Failure		404		{object}	map[string]interface{}	"Chat not found"
//	@Failure		500		{object}	map[string]interface{}	"Internal server error"
//	@Router			/api/v1/chats/{id}/subject [patch]
func (h *ChatHandler) UpdateChatSubject(c *gin.Context) {
	_, exists := middleware.GetAuthContext(c)
	if !exists {
		apierrors.Unauthorized(c, "Authentication required")
		return
	}

	idStr := c.Param("id")
	chatID, err := uuid.Parse(idStr)
	if err != nil {
		apierrors.ValidationError(c, "id", "Invalid chat ID format (must be UUID)")
		return
	}

	var req UpdateSubjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to parse update subject request", zap.Error(err))
		apierrors.BadRequest(c, "Invalid request body: "+err.Error())
		return
	}

	input := chatapp.UpdateChatSubjectInput{
		ChatID:  chatID,
		Subject: req.Subject,
	}

	output, err := h.updateChatUseCase.UpdateSubject(c.Request.Context(), input)
	if err != nil {
		h.logger.Error("Failed to update chat subject", zap.Error(err))
		apierrors.RespondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, output.Chat)
}
