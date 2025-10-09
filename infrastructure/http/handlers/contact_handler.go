package handlers

import (
	"net/http"
	"strconv"

	"github.com/caloi/ventros-crm/infrastructure/http/middleware"
	contactapp "github.com/caloi/ventros-crm/internal/application/contact"
	"github.com/caloi/ventros-crm/internal/domain/contact"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type ContactHandler struct {
	logger                      *zap.Logger
	contactRepo                 contact.Repository
	changePipelineStatusUseCase *contactapp.ChangePipelineStatusUseCase
	// TODO: Add use cases when needed
	// createContactUseCase *contactapp.CreateContactUseCase
}

func NewContactHandler(logger *zap.Logger, contactRepo contact.Repository, changePipelineStatusUseCase *contactapp.ChangePipelineStatusUseCase) *ContactHandler {
	return &ContactHandler{
		logger:                      logger,
		contactRepo:                 contactRepo,
		changePipelineStatusUseCase: changePipelineStatusUseCase,
	}
}

// CreateContactRequest representa o payload para criar um contato
type CreateContactRequest struct {
	Name          string            `json:"name" binding:"required" example:"João Silva"`
	Email         string            `json:"email" example:"joao@example.com"`
	Phone         string            `json:"phone" example:"+5511999999999"`
	ExternalID    string            `json:"external_id" example:"ext_123"`
	SourceChannel string            `json:"source_channel" example:"whatsapp"`
	Language      string            `json:"language" example:"pt-BR"`
	Timezone      string            `json:"timezone" example:"America/Sao_Paulo"`
	Tags          []string          `json:"tags" example:"lead,whatsapp"`
	CustomFields  map[string]string `json:"custom_fields" example:"company:Empresa XYZ"`
}

// UpdateContactRequest representa o payload para atualizar um contato
type UpdateContactRequest struct {
	Name          *string           `json:"name,omitempty"`
	Email         *string           `json:"email,omitempty"`
	Phone         *string           `json:"phone,omitempty"`
	ExternalID    *string           `json:"external_id,omitempty"`
	SourceChannel *string           `json:"source_channel,omitempty"`
	Language      *string           `json:"language,omitempty"`
	Timezone      *string           `json:"timezone,omitempty"`
	Tags          []string          `json:"tags,omitempty"`
	CustomFields  map[string]string `json:"custom_fields,omitempty"`
}

// ListContacts lists all contacts with optional filters
//
//	@Summary		List contacts
//	@Description	Lista todos os contatos com filtros opcionais (apenas do usuário autenticado)
//	@Tags			contacts
//	@Accept			json
//	@Produce		json
//	@Param			project_id	query		string					true	"Project ID"
//	@Param			page		query		int						false	"Page number"	default(1)
//	@Param			page_size	query		int						false	"Page size"		default(20)
//	@Success		200			{object}	map[string]interface{}	"List of contacts"
//	@Failure		400			{object}	map[string]interface{}	"Invalid parameters"
//	@Failure		401			{object}	map[string]interface{}	"Not authenticated"
//	@Failure		403			{object}	map[string]interface{}	"Project not owned by user"
//	@Failure		500			{object}	map[string]interface{}	"Internal server error"
//	@Router			/api/v1/contacts [get]
func (h *ContactHandler) ListContacts(c *gin.Context) {
	// Verificar autenticação
	_, exists := middleware.GetAuthContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	projectIDStr := c.Query("project_id")
	if projectIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "project_id is required"})
		return
	}

	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project_id format"})
		return
	}

	// Parse pagination parameters
	page := 1
	pageSize := 20

	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if pageSizeStr := c.Query("page_size"); pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 && ps <= 100 {
			pageSize = ps
		}
	}

	// Calculate offset
	offset := (page - 1) * pageSize

	// Get total count
	total, err := h.contactRepo.CountByProject(c.Request.Context(), projectID)
	if err != nil {
		h.logger.Error("Failed to count contacts", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve contacts"})
		return
	}

	// Get contacts
	contacts, err := h.contactRepo.FindByProject(c.Request.Context(), projectID, pageSize, offset)
	if err != nil {
		h.logger.Error("Failed to list contacts", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve contacts"})
		return
	}

	// Convert to response
	contactResponses := make([]map[string]interface{}, len(contacts))
	for i, contact := range contacts {
		contactResponses[i] = h.contactToResponse(contact)
	}

	c.JSON(http.StatusOK, gin.H{
		"contacts":  contactResponses,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// CreateContact creates a new contact
//
//	@Summary		Create a new contact
//	@Description	Create a new contact in the system
//	@Tags			contacts
//	@Accept			json
//	@Produce		json
//	@Param			project_id	query		string					true	"Project ID"
//	@Param			contact		body		CreateContactRequest	true	"Contact data"
//	@Success		201			{object}	map[string]interface{}	"Contact created successfully"
//	@Failure		400			{object}	map[string]interface{}	"Invalid request"
//	@Failure		500			{object}	map[string]interface{}	"Internal server error"
//	@Router			/api/v1/contacts [post]
func (h *ContactHandler) CreateContact(c *gin.Context) {
	projectIDStr := c.Query("project_id")
	if projectIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "project_id is required"})
		return
	}

	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project_id format"})
		return
	}

	var req CreateContactRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to parse contact request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// TODO: Get tenant_id from context/auth
	tenantID := "default" // Placeholder

	// Create domain contact
	domainContact, err := contact.NewContact(projectID, tenantID, req.Name)
	if err != nil {
		h.logger.Error("Failed to create domain contact", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set optional fields
	if req.Email != "" {
		if err := domainContact.SetEmail(req.Email); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email: " + err.Error()})
			return
		}
	}

	if req.Phone != "" {
		if err := domainContact.SetPhone(req.Phone); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid phone: " + err.Error()})
			return
		}
	}

	if req.ExternalID != "" {
		domainContact.SetExternalID(req.ExternalID)
	}

	if req.SourceChannel != "" {
		domainContact.SetSourceChannel(req.SourceChannel)
	}

	if req.Language != "" {
		domainContact.SetLanguage(req.Language)
	}

	if req.Timezone != "" {
		domainContact.SetTimezone(req.Timezone)
	}

	// Add tags
	for _, tag := range req.Tags {
		domainContact.AddTag(tag)
	}

	// Save contact
	if err := h.contactRepo.Save(c.Request.Context(), domainContact); err != nil {
		h.logger.Error("Failed to save contact", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create contact"})
		return
	}

	// TODO: Publish domain events (should be done by use case)
	// For now, we'll skip event publishing until use case is properly integrated
	// This means webhooks won't be triggered from direct API calls
	h.logger.Info("Contact created successfully",
		zap.String("contact_id", domainContact.ID().String()),
		zap.String("name", domainContact.Name()),
		zap.Int("domain_events", len(domainContact.DomainEvents())),
	)

	// Convert to response
	response := h.contactToResponse(domainContact)
	c.JSON(http.StatusCreated, response)
}

// GetContact gets a contact by ID
//
//	@Summary		Get contact by ID
//	@Description	Get detailed information about a specific contact
//	@Tags			contacts
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string					true	"Contact ID"
//	@Success		200	{object}	map[string]interface{}	"Contact details"
//	@Failure		400	{object}	map[string]interface{}	"Invalid contact ID"
//	@Failure		404	{object}	map[string]interface{}	"Contact not found"
//	@Failure		500	{object}	map[string]interface{}	"Internal server error"
//	@Router			/api/v1/contacts/{id} [get]
func (h *ContactHandler) GetContact(c *gin.Context) {
	idStr := c.Param("id")
	contactID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid contact ID format"})
		return
	}

	domainContact, err := h.contactRepo.FindByID(c.Request.Context(), contactID)
	if err != nil {
		h.logger.Error("Failed to find contact", zap.String("contact_id", contactID.String()), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve contact"})
		return
	}

	if domainContact == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Contact not found"})
		return
	}

	// Convert domain contact to API response
	response := h.contactToResponse(domainContact)
	c.JSON(http.StatusOK, response)
}

// UpdateContact updates a contact
//
//	@Summary		Update contact
//	@Description	Atualiza um contato existente
//	@Tags			contacts
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string					true	"Contact ID (UUID)"
//	@Param			contact	body		UpdateContactRequest	true	"Contact update data"
//	@Success		200		{object}	map[string]interface{}	"Contact updated successfully"
//	@Failure		400		{object}	map[string]interface{}	"Invalid request"
//	@Failure		404		{object}	map[string]interface{}	"Contact not found"
//	@Failure		500		{object}	map[string]interface{}	"Internal server error"
//	@Router			/api/v1/contacts/{id} [put]
func (h *ContactHandler) UpdateContact(c *gin.Context) {
	idStr := c.Param("id")
	contactID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid contact ID format"})
		return
	}

	var req UpdateContactRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to parse update request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// Find existing contact
	domainContact, err := h.contactRepo.FindByID(c.Request.Context(), contactID)
	if err != nil {
		h.logger.Error("Failed to find contact", zap.String("contact_id", contactID.String()), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve contact"})
		return
	}

	if domainContact == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Contact not found"})
		return
	}

	// Update fields
	if req.Name != nil {
		domainContact.UpdateName(*req.Name)
	}

	if req.Email != nil {
		if err := domainContact.SetEmail(*req.Email); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email: " + err.Error()})
			return
		}
	}

	if req.Phone != nil {
		if err := domainContact.SetPhone(*req.Phone); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid phone: " + err.Error()})
			return
		}
	}

	if req.ExternalID != nil {
		domainContact.SetExternalID(*req.ExternalID)
	}

	if req.SourceChannel != nil {
		domainContact.SetSourceChannel(*req.SourceChannel)
	}

	if req.Language != nil {
		domainContact.SetLanguage(*req.Language)
	}

	if req.Timezone != nil {
		domainContact.SetTimezone(*req.Timezone)
	}

	// Update tags if provided
	if req.Tags != nil {
		// Clear existing tags and add new ones
		domainContact.ClearTags()
		for _, tag := range req.Tags {
			domainContact.AddTag(tag)
		}
	}

	// Save updated contact
	if err := h.contactRepo.Save(c.Request.Context(), domainContact); err != nil {
		h.logger.Error("Failed to save contact", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update contact"})
		return
	}

	// Convert to response
	response := h.contactToResponse(domainContact)
	c.JSON(http.StatusOK, response)
}

// DeleteContact deletes a contact
//
//	@Summary		Delete contact
//	@Description	Remove um contato (soft delete)
//	@Tags			contacts
//	@Produce		json
//	@Param			id	path	string	true	"Contact ID (UUID)"
//	@Success		204	"Contact deleted successfully"
//	@Failure		400	{object}	map[string]interface{}	"Invalid contact ID"
//	@Failure		404	{object}	map[string]interface{}	"Contact not found"
//	@Failure		500	{object}	map[string]interface{}	"Internal server error"
//	@Router			/api/v1/contacts/{id} [delete]
func (h *ContactHandler) DeleteContact(c *gin.Context) {
	idStr := c.Param("id")
	contactID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid contact ID format"})
		return
	}

	// Find existing contact
	domainContact, err := h.contactRepo.FindByID(c.Request.Context(), contactID)
	if err != nil {
		h.logger.Error("Failed to find contact", zap.String("contact_id", contactID.String()), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve contact"})
		return
	}

	if domainContact == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Contact not found"})
		return
	}

	// Soft delete
	domainContact.Delete()

	// Save updated contact
	if err := h.contactRepo.Save(c.Request.Context(), domainContact); err != nil {
		h.logger.Error("Failed to delete contact", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete contact"})
		return
	}

	c.Status(http.StatusNoContent)
}

// contactToResponse converts domain contact to API response
func (h *ContactHandler) contactToResponse(c *contact.Contact) map[string]interface{} {
	response := map[string]interface{}{
		"id":                   c.ID(),
		"project_id":           c.ProjectID(),
		"tenant_id":            c.TenantID(),
		"name":                 c.Name(),
		"email":                c.Email(),
		"phone":                c.Phone(),
		"external_id":          c.ExternalID(),
		"source_channel":       c.SourceChannel(),
		"language":             c.Language(),
		"timezone":             c.Timezone(),
		"tags":                 c.Tags(),
		"first_interaction_at": c.FirstInteractionAt(),
		"last_interaction_at":  c.LastInteractionAt(),
		"created_at":           c.CreatedAt(),
		"updated_at":           c.UpdatedAt(),
		"deleted_at":           c.DeletedAt(),
		"is_deleted":           c.IsDeleted(),
	}

	return response
}

// ChangePipelineStatusRequest representa o request para mudar status no pipeline
type ChangePipelineStatusRequest struct {
	StatusID uuid.UUID `json:"status_id" binding:"required"`
	Reason   string    `json:"reason,omitempty"`
}

// ChangePipelineStatus muda o status de um contato em um pipeline
//
//	@Summary		Change contact pipeline status
//	@Description	Altera o status de um contato em um pipeline específico
//	@Tags			contacts
//	@Accept			json
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			id			path		string						true	"Contact ID (UUID)"
//	@Param			pipeline_id	path		string						true	"Pipeline ID (UUID)"
//	@Param			request		body		ChangePipelineStatusRequest	true	"Status change request"
//	@Success		200			{object}	map[string]interface{}		"Status changed successfully"
//	@Failure		400			{object}	map[string]interface{}		"Invalid request"
//	@Failure		401			{object}	map[string]interface{}		"Authentication required"
//	@Failure		404			{object}	map[string]interface{}		"Contact or pipeline not found"
//	@Failure		500			{object}	map[string]interface{}		"Internal server error"
//	@Router			/api/v1/contacts/{id}/pipelines/{pipeline_id}/status [put]
func (h *ContactHandler) ChangePipelineStatus(c *gin.Context) {
	authCtx, exists := middleware.GetAuthContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	contactID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid contact ID"})
		return
	}

	pipelineID, err := uuid.Parse(c.Param("pipeline_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid pipeline ID"})
		return
	}

	var req ChangePipelineStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// Execute use case
	input := contactapp.ChangePipelineStatusInput{
		ContactID:  contactID,
		PipelineID: pipelineID,
		StatusID:   req.StatusID,
		ChangedBy:  &authCtx.UserID,
		Reason:     req.Reason,
		TenantID:   authCtx.TenantID,
		ProjectID:  authCtx.ProjectID,
	}

	output, err := h.changePipelineStatusUseCase.Execute(c.Request.Context(), input)
	if err != nil {
		h.logger.Error("Failed to change pipeline status", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":              "Pipeline status changed successfully",
		"contact_id":           output.ContactID,
		"pipeline_id":          output.PipelineID,
		"previous_status_id":   output.PreviousStatusID,
		"previous_status_name": output.PreviousStatusName,
		"new_status_id":        output.NewStatusID,
		"new_status_name":      output.NewStatusName,
		"changed_at":           output.ChangedAt,
	})
}
