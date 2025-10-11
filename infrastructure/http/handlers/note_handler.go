package handlers

import (
	"net/http"
	"strconv"

	apierrors "github.com/caloi/ventros-crm/infrastructure/http/errors"
	"github.com/caloi/ventros-crm/infrastructure/http/middleware"
	"github.com/caloi/ventros-crm/internal/application/queries"
	"github.com/caloi/ventros-crm/internal/domain/core/shared"
	"github.com/caloi/ventros-crm/internal/domain/crm/note"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type NoteHandler struct {
	logger                  *zap.Logger
	noteRepo                note.Repository
	listNotesQueryHandler   *queries.ListNotesQueryHandler
	searchNotesQueryHandler *queries.SearchNotesQueryHandler
}

func NewNoteHandler(logger *zap.Logger, noteRepo note.Repository) *NoteHandler {
	return &NoteHandler{
		logger:                  logger,
		noteRepo:                noteRepo,
		listNotesQueryHandler:   queries.NewListNotesQueryHandler(noteRepo, logger),
		searchNotesQueryHandler: queries.NewSearchNotesQueryHandler(noteRepo, logger),
	}
}

// ListNotesAdvanced lists notes with advanced filters, pagination, and sorting
//
//	@Summary		List notes with advanced filters and pagination
//	@Description	Retrieve all notes with comprehensive filtering capabilities. Notes are annotations, comments, and action items added by agents, automation, or the system during customer interactions. Essential for maintaining conversation context, tracking follow-ups, and audit trails.
//	@Description
//	@Description	**Filtering Capabilities:**
//	@Description	- Filter by contact_id to view all notes for a specific customer
//	@Description	- Filter by session_id to see notes from a particular conversation
//	@Description	- Filter by author_id to track notes from specific agents or automation rules
//	@Description	- Filter by author_type (agent, system, automation) to distinguish note sources
//	@Description	- Filter by note_type (comment, action, follow-up, escalation, resolution) to organize by purpose
//	@Description	- Filter by priority (low, medium, high, urgent) for task management
//	@Description	- Filter by visible_to_client flag to separate internal vs customer-facing notes
//	@Description	- Filter by pinned flag to identify important or starred notes
//	@Description
//	@Description	**Common Use Cases:**
//	@Description	- Build contact history timelines with all interactions and annotations
//	@Description	- Generate session summaries with agent notes and context
//	@Description	- Track agent activity and note-taking patterns
//	@Description	- Manage follow-up tasks and action items
//	@Description	- Identify escalated issues requiring attention
//	@Description	- Create customer-facing summaries (visible_to_client=true)
//	@Description	- Audit trail for compliance and quality assurance
//	@Description	- Filter high-priority notes for urgent follow-ups
//	@Description
//	@Description	**Sorting Options:**
//	@Description	- Sort by created_at (chronological order)
//	@Description	- Sort by priority (task prioritization)
//	@Description	- Ascending or descending order
//	@Description
//	@Description	**Performance:**
//	@Description	- Optimized GORM indexes on tenant+contact for fast contact note queries
//	@Description	- Composite indexes on tenant+session for session note retrieval
//	@Description	- Indexes on tenant+author for agent activity tracking
//	@Description	- Efficiently handles large note volumes per contact
//	@Tags			CRM - Notes
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			contact_id			query		string						false	"Filter by contact UUID - Example: 550e8400-e29b-41d4-a716-446655440000"
//	@Param			session_id			query		string						false	"Filter by session UUID - Example: 660e8400-e29b-41d4-a716-446655440001"
//	@Param			author_id			query		string						false	"Filter by author (agent/automation) UUID"
//	@Param			author_type			query		string						false	"Filter by author type"												Enums(agent, system, automation)							example(agent)
//	@Param			note_type			query		string						false	"Filter by note purpose"											Enums(comment, action, follow-up, escalation, resolution)	example(follow-up)
//	@Param			priority			query		string						false	"Filter by priority level"											Enums(low, medium, high, urgent)							example(high)
//	@Param			visible_to_client	query		bool						false	"Filter by client visibility - true: customer-facing notes only"	example(false)
//	@Param			pinned				query		bool						false	"Filter by pinned status - true: important/starred notes only"		example(true)
//	@Param			page				query		int							false	"Page number for pagination (starts at 1)"							default(1)					minimum(1)			example(1)
//	@Param			limit				query		int							false	"Results per page (max 100)"										default(20)					minimum(1)			maximum(100)	example(20)
//	@Param			sort_by				query		string						false	"Field to sort by"													Enums(created_at, priority)	default(created_at)	example(created_at)
//	@Param			sort_dir			query		string						false	"Sort direction"													Enums(asc, desc)			default(desc)		example(desc)
//	@Success		200					{object}	queries.ListNotesResponse	"Successfully retrieved notes with full details"
//	@Failure		400					{object}	map[string]interface{}		"Bad Request - Invalid UUID or parameter format"
//	@Failure		401					{object}	map[string]interface{}		"Unauthorized - Authentication required"
//	@Failure		403					{object}	map[string]interface{}		"Forbidden - No access to this tenant's notes"
//	@Failure		500					{object}	map[string]interface{}		"Internal Server Error"
//	@Router			/api/v1/crm/notes/advanced [get]
func (h *NoteHandler) ListNotesAdvanced(c *gin.Context) {
	authCtx, exists := middleware.GetAuthContext(c)
	if !exists {
		apierrors.Unauthorized(c, "Authentication required")
		return
	}

	tenantID, err := shared.NewTenantID(authCtx.TenantID)
	if err != nil {
		h.logger.Error("Invalid tenant ID", zap.Error(err))
		apierrors.InternalError(c, "Invalid tenant configuration", err)
		return
	}

	// Parse pagination
	page := 1
	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	limit := 20
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	// Parse sorting
	sortBy := c.DefaultQuery("sort_by", "created_at")
	sortDir := c.DefaultQuery("sort_dir", "desc")

	// Parse contact_id filter
	var contactID *uuid.UUID
	if contactIDStr := c.Query("contact_id"); contactIDStr != "" {
		if cid, err := uuid.Parse(contactIDStr); err == nil {
			contactID = &cid
		} else {
			apierrors.ValidationError(c, "contact_id", "Invalid UUID format")
			return
		}
	}

	// Parse session_id filter
	var sessionID *uuid.UUID
	if sessionIDStr := c.Query("session_id"); sessionIDStr != "" {
		if sid, err := uuid.Parse(sessionIDStr); err == nil {
			sessionID = &sid
		} else {
			apierrors.ValidationError(c, "session_id", "Invalid UUID format")
			return
		}
	}

	// Parse author_id filter
	var authorID *uuid.UUID
	if authorIDStr := c.Query("author_id"); authorIDStr != "" {
		if aid, err := uuid.Parse(authorIDStr); err == nil {
			authorID = &aid
		} else {
			apierrors.ValidationError(c, "author_id", "Invalid UUID format")
			return
		}
	}

	// Parse author_type filter
	var authorType *string
	if authorTypeStr := c.Query("author_type"); authorTypeStr != "" {
		authorType = &authorTypeStr
	}

	// Parse note_type filter
	var noteType *string
	if noteTypeStr := c.Query("note_type"); noteTypeStr != "" {
		noteType = &noteTypeStr
	}

	// Parse priority filter
	var priority *string
	if priorityStr := c.Query("priority"); priorityStr != "" {
		priority = &priorityStr
	}

	// Parse visible_to_client filter
	var visibleToClient *bool
	if visibleStr := c.Query("visible_to_client"); visibleStr != "" {
		if v, err := strconv.ParseBool(visibleStr); err == nil {
			visibleToClient = &v
		}
	}

	// Parse pinned filter
	var pinned *bool
	if pinnedStr := c.Query("pinned"); pinnedStr != "" {
		if p, err := strconv.ParseBool(pinnedStr); err == nil {
			pinned = &p
		}
	}

	query := queries.ListNotesQuery{
		TenantID:        tenantID,
		ContactID:       contactID,
		SessionID:       sessionID,
		AuthorID:        authorID,
		AuthorType:      authorType,
		NoteType:        noteType,
		Priority:        priority,
		VisibleToClient: visibleToClient,
		Pinned:          pinned,
		Page:            page,
		Limit:           limit,
		SortBy:          sortBy,
		SortDir:         sortDir,
	}

	response, err := h.listNotesQueryHandler.Handle(c.Request.Context(), query)
	if err != nil {
		h.logger.Error("Failed to list notes", zap.Error(err))
		apierrors.InternalError(c, "Failed to retrieve notes", err)
		return
	}

	c.JSON(http.StatusOK, response)
}

// SearchNotes performs full-text search on notes
//
//	@Summary		Search notes by content and author
//	@Description	Full-text search across note content and author names. Perfect for finding specific annotations, comments, or action items across all customer interactions and conversations.
//	@Description
//	@Description	**Search Capabilities:**
//	@Description	- Searches note content/body (primary field)
//	@Description	- Searches author names (secondary field)
//	@Description	- Case-insensitive ILIKE matching
//	@Description
//	@Description	**Match Scoring:**
//	@Description	- Content matches: 1.5 score (higher priority)
//	@Description	- Author name matches: 1.2 score (lower priority)
//	@Description
//	@Description	**Search Examples:**
//	@Description	- "follow-up required" - Find notes about pending follow-ups
//	@Description	- "escalated to manager" - Locate escalation notes
//	@Description	- "pricing question" - Find pricing-related annotations
//	@Description	- "technical issue" - Search for technical problem notes
//	@Description	- "João" - Find notes written by agent João
//	@Description	- "urgent" - Locate urgent action items
//	@Description
//	@Description	**Performance:**
//	@Description	- Optimized GORM indexes on tenant_id for fast tenant isolation
//	@Description	- ILIKE operator uses PostgreSQL's text search capabilities
//	@Description	- Maximum 100 results to ensure sub-second response times
//	@Tags			CRM - Notes
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			q		query		string						true	"Search query - content or author name"	minlength(1)	example(follow-up required)
//	@Param			limit	query		int							false	"Maximum results (max 100)"				default(20)		minimum(1)	maximum(100)	example(10)
//	@Success		200		{object}	queries.SearchNotesResponse	"Found notes with match scores"
//	@Failure		400		{object}	map[string]interface{}		"Bad Request - Empty search query"
//	@Failure		401		{object}	map[string]interface{}		"Unauthorized"
//	@Failure		403		{object}	map[string]interface{}		"Forbidden"
//	@Failure		500		{object}	map[string]interface{}		"Internal Server Error"
//	@Router			/api/v1/crm/notes/search [get]
func (h *NoteHandler) SearchNotes(c *gin.Context) {
	authCtx, exists := middleware.GetAuthContext(c)
	if !exists {
		apierrors.Unauthorized(c, "Authentication required")
		return
	}

	tenantID, err := shared.NewTenantID(authCtx.TenantID)
	if err != nil {
		h.logger.Error("Invalid tenant ID", zap.Error(err))
		apierrors.InternalError(c, "Invalid tenant configuration", err)
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

	query := queries.SearchNotesQuery{
		TenantID:   tenantID,
		SearchText: searchText,
		Limit:      limit,
	}

	response, err := h.searchNotesQueryHandler.Handle(c.Request.Context(), query)
	if err != nil {
		h.logger.Error("Failed to search notes", zap.Error(err))
		apierrors.InternalError(c, "Failed to search notes", err)
		return
	}

	c.JSON(http.StatusOK, response)
}
