package handlers

import (
	"net/http"
	"strconv"
	"strings"

	apierrors "github.com/caloi/ventros-crm/infrastructure/http/errors"
	"github.com/caloi/ventros-crm/infrastructure/http/middleware"
	contactapp "github.com/caloi/ventros-crm/internal/application/contact"
	"github.com/caloi/ventros-crm/internal/application/queries"
	"github.com/caloi/ventros-crm/internal/domain/core/shared"
	"github.com/caloi/ventros-crm/internal/domain/crm/contact"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type ContactHandler struct {
	logger                      *zap.Logger
	contactRepo                 contact.Repository
	changePipelineStatusUseCase *contactapp.ChangePipelineStatusUseCase
	// Query handlers
	listContactsQueryHandler   *queries.ListContactsQueryHandler
	searchContactsQueryHandler *queries.SearchContactsQueryHandler
	// TODO: Add use cases when needed
	// createContactUseCase *contactapp.CreateContactUseCase
}

func NewContactHandler(logger *zap.Logger, contactRepo contact.Repository, changePipelineStatusUseCase *contactapp.ChangePipelineStatusUseCase) *ContactHandler {
	return &ContactHandler{
		logger:                      logger,
		contactRepo:                 contactRepo,
		changePipelineStatusUseCase: changePipelineStatusUseCase,
		listContactsQueryHandler:    queries.NewListContactsQueryHandler(contactRepo, logger),
		searchContactsQueryHandler:  queries.NewSearchContactsQueryHandler(contactRepo, logger),
	}
}

// CreateContactRequest represents the request to create a contact.
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

// UpdateContactRequest represents the request to update a contact.
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
//	@Description	List all contacts with optional filters (authenticated user only)
//	@Tags			CRM - Contacts
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
	_, exists := middleware.GetAuthContext(c)
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
		apierrors.InternalError(c, "Failed to count contacts", err)
		return
	}

	// Get contacts
	contacts, err := h.contactRepo.FindByProject(c.Request.Context(), projectID, pageSize, offset)
	if err != nil {
		h.logger.Error("Failed to list contacts", zap.Error(err))
		apierrors.InternalError(c, "Failed to retrieve contacts", err)
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
//	@Tags			CRM - Contacts
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
		apierrors.ValidationError(c, "project_id", "project_id query parameter is required")
		return
	}

	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		apierrors.ValidationError(c, "project_id", "Invalid project_id format (must be UUID)")
		return
	}

	var req CreateContactRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to parse contact request", zap.Error(err))
		apierrors.BadRequest(c, "Invalid request body: "+err.Error())
		return
	}

	// TODO: Get tenant_id from context/auth
	tenantID := "default" // Placeholder

	// Create domain contact
	domainContact, err := contact.NewContact(projectID, tenantID, req.Name)
	if err != nil {
		h.logger.Error("Failed to create domain contact", zap.Error(err))
		apierrors.RespondWithError(c, err)
		return
	}

	// Set optional fields
	if req.Email != "" {
		if err := domainContact.SetEmail(req.Email); err != nil {
			apierrors.ValidationError(c, "email", "Invalid email format")
			return
		}
	}

	if req.Phone != "" {
		if err := domainContact.SetPhone(req.Phone); err != nil {
			apierrors.ValidationError(c, "phone", "Invalid phone format")
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
		apierrors.InternalError(c, "Failed to save contact", err)
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
//	@Tags			CRM - Contacts
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
		apierrors.ValidationError(c, "id", "Invalid contact ID format (must be UUID)")
		return
	}

	domainContact, err := h.contactRepo.FindByID(c.Request.Context(), contactID)
	if err != nil {
		// Error is already a DomainError from repository
		apierrors.RespondWithError(c, err)
		return
	}

	if domainContact == nil {
		apierrors.NotFound(c, "contact", contactID.String())
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
//	@Tags			CRM - Contacts
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
		apierrors.ValidationError(c, "id", "Invalid contact ID format (must be UUID)")
		return
	}

	var req UpdateContactRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to parse update request", zap.Error(err))
		apierrors.BadRequest(c, "Invalid request body: "+err.Error())
		return
	}

	// Find existing contact
	domainContact, err := h.contactRepo.FindByID(c.Request.Context(), contactID)
	if err != nil {
		apierrors.RespondWithError(c, err)
		return
	}

	if domainContact == nil {
		apierrors.NotFound(c, "contact", contactID.String())
		return
	}

	// Update fields
	if req.Name != nil {
		domainContact.UpdateName(*req.Name)
	}

	if req.Email != nil {
		if err := domainContact.SetEmail(*req.Email); err != nil {
			apierrors.ValidationError(c, "email", "Invalid email format")
			return
		}
	}

	if req.Phone != nil {
		if err := domainContact.SetPhone(*req.Phone); err != nil {
			apierrors.ValidationError(c, "phone", "Invalid phone format")
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
		apierrors.InternalError(c, "Failed to update contact", err)
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
//	@Tags			CRM - Contacts
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
		apierrors.ValidationError(c, "id", "Invalid contact ID format (must be UUID)")
		return
	}

	// Find existing contact
	domainContact, err := h.contactRepo.FindByID(c.Request.Context(), contactID)
	if err != nil {
		apierrors.RespondWithError(c, err)
		return
	}

	if domainContact == nil {
		apierrors.NotFound(c, "contact", contactID.String())
		return
	}

	// Soft delete
	domainContact.Delete()

	// Save updated contact
	if err := h.contactRepo.Save(c.Request.Context(), domainContact); err != nil {
		h.logger.Error("Failed to delete contact", zap.Error(err))
		apierrors.InternalError(c, "Failed to delete contact", err)
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

// ChangePipelineStatusRequest represents the request to change pipeline status.
type ChangePipelineStatusRequest struct {
	StatusID uuid.UUID `json:"status_id" binding:"required"`
	Reason   string    `json:"reason,omitempty"`
}

// ChangePipelineStatus changes a contact's status in a pipeline
//
//	@Summary		Change contact pipeline status
//	@Description	Change contact status in a specific pipeline
//	@Tags			CRM - Contacts
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
		apierrors.Unauthorized(c, "Authentication required")
		return
	}

	contactID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		apierrors.ValidationError(c, "id", "Invalid contact ID format (must be UUID)")
		return
	}

	pipelineID, err := uuid.Parse(c.Param("pipeline_id"))
	if err != nil {
		apierrors.ValidationError(c, "pipeline_id", "Invalid pipeline ID format (must be UUID)")
		return
	}

	var req ChangePipelineStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierrors.BadRequest(c, "Invalid request body: "+err.Error())
		return
	}

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
		apierrors.RespondWithError(c, err)
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

// ListContactsAdvanced lists contacts with advanced filters, pagination, and sorting
//
//	@Summary		List contacts with advanced filters
//	@Description	List contacts with advanced filters (name, phone, email, tags, dates), pagination, and sorting
//	@Tags			CRM - Contacts
//	@Accept			json
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			name			query		string					false	"Filter by name (partial match)"
//	@Param			phone			query		string					false	"Filter by phone (partial match)"
//	@Param			email			query		string					false	"Filter by email (partial match)"
//	@Param			tags			query		[]string				false	"Filter by tags (comma-separated)"
//	@Param			created_after	query		string					false	"Filter by created_after (YYYY-MM-DD)"
//	@Param			created_before	query		string					false	"Filter by created_before (YYYY-MM-DD)"
//	@Param			page			query		int						false	"Page number"		default(1)
//	@Param			limit			query		int						false	"Page size"			default(20)
//	@Param			sort_by			query		string					false	"Sort by field"		default(created_at)
//	@Param			sort_dir		query		string					false	"Sort direction"	default(desc)	Enums(asc, desc)
//	@Success		200				{object}	map[string]interface{}	"List of contacts with pagination"
//	@Failure		400				{object}	map[string]interface{}	"Invalid parameters"
//	@Failure		401				{object}	map[string]interface{}	"Not authenticated"
//	@Failure		500				{object}	map[string]interface{}	"Internal server error"
//	@Router			/api/v1/contacts/advanced [get]
func (h *ContactHandler) ListContactsAdvanced(c *gin.Context) {
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
	sortBy := c.DefaultQuery("sort_by", "created_at")
	sortDir := c.DefaultQuery("sort_dir", "desc")

	// Parse filters
	filters := queries.ContactFilters{
		Name:          c.Query("name"),
		Phone:         c.Query("phone"),
		Email:         c.Query("email"),
		CreatedAfter:  c.Query("created_after"),
		CreatedBefore: c.Query("created_before"),
	}

	// Parse tags (comma-separated)
	if tagsStr := c.Query("tags"); tagsStr != "" {
		filters.Tags = strings.Split(tagsStr, ",")
	}

	// Create tenant ID
	tenantID, err := shared.NewTenantID(authCtx.TenantID)
	if err != nil {
		apierrors.ValidationError(c, "tenant_id", "Invalid tenant ID")
		return
	}

	// Execute query
	query := queries.ListContactsQuery{
		TenantID: tenantID,
		Filters:  filters,
		Page:     page,
		Limit:    limit,
		SortBy:   sortBy,
		SortDir:  sortDir,
	}

	response, err := h.listContactsQueryHandler.Handle(c.Request.Context(), query)
	if err != nil {
		h.logger.Error("Failed to list contacts", zap.Error(err))
		apierrors.InternalError(c, "Failed to retrieve contacts", err)
		return
	}

	c.JSON(http.StatusOK, response)
}

// SearchContacts performs full-text search on contacts
//
//	@Summary		Search contacts
//	@Description	Full-text search on contact name, phone, and email with relevance scoring
//	@Tags			CRM - Contacts
//	@Accept			json
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			q		query		string					true	"Search query"
//	@Param			limit	query		int						false	"Result limit"	default(20)
//	@Success		200		{object}	map[string]interface{}	"Search results with match scores"
//	@Failure		400		{object}	map[string]interface{}	"Invalid parameters"
//	@Failure		401		{object}	map[string]interface{}	"Not authenticated"
//	@Failure		500		{object}	map[string]interface{}	"Internal server error"
//	@Router			/api/v1/contacts/search [get]
func (h *ContactHandler) SearchContacts(c *gin.Context) {
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
	query := queries.SearchContactsQuery{
		TenantID:   tenantID,
		SearchText: searchText,
		Limit:      limit,
	}

	response, err := h.searchContactsQueryHandler.Handle(c.Request.Context(), query)
	if err != nil {
		h.logger.Error("Failed to search contacts", zap.Error(err))
		apierrors.InternalError(c, "Failed to search contacts", err)
		return
	}

	c.JSON(http.StatusOK, response)
}
