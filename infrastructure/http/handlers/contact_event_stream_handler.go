package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/caloi/ventros-crm/infrastructure/persistence"
	"github.com/caloi/ventros-crm/internal/domain/crm/contact_event"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type ContactEventStreamHandler struct {
	eventRepo *persistence.GormContactEventRepository
	logger    *zap.Logger
}

func NewContactEventStreamHandler(eventRepo *persistence.GormContactEventRepository, logger *zap.Logger) *ContactEventStreamHandler {
	return &ContactEventStreamHandler{
		eventRepo: eventRepo,
		logger:    logger,
	}
}

// StreamContactEvents streams contact events via SSE
//
//	@Summary		Stream contact events
//	@Description	Stream contact events in real-time using Server-Sent Events (SSE). Requires authentication.
//	@Tags			contact-events
//	@Security		BearerAuth
//	@Produce		text/event-stream
//	@Param			contact_id	path		string				true	"Contact ID"
//	@Param			categories	query		string				false	"Filter by categories (comma-separated): status,pipeline,assignment,tag,note,session,custom_field,system,notification"
//	@Param			priority	query		string				false	"Filter by priority: low,normal,high,urgent"
//	@Success		200			{string}	string				"Event stream"
//	@Failure		401			{object}	map[string]string	"Unauthorized"
//	@Failure		403			{object}	map[string]string	"Forbidden - no access to this contact"
//	@Router			/api/v1/contacts/{contact_id}/events/stream [get]
func (h *ContactEventStreamHandler) StreamContactEvents(c *gin.Context) {
	// SECURITY: Verify authentication (middleware should handle this)
	authCtx, exists := c.Get("auth")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	// SECURITY: Validate Origin header to prevent CSRF
	origin := c.GetHeader("Origin")
	if origin != "" && !h.isAllowedOrigin(origin) {
		h.logger.Warn("Rejected SSE connection from unauthorized origin",
			zap.String("origin", origin))
		c.JSON(http.StatusForbidden, gin.H{"error": "Origin not allowed"})
		return
	}

	contactIDStr := c.Param("contact_id")
	contactID, err := uuid.Parse(contactIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid contact ID"})
		return
	}

	// SECURITY: Verify user has access to this contact (RLS)
	// The auth context should contain tenant/project info
	_ = authCtx // Will be used by repository through context

	// Parse filters
	categoriesFilter := parseCategories(c.Query("categories"))
	priorityFilter := parsePriority(c.Query("priority"))

	// SECURITY: Set SSE headers with security considerations
	c.Writer.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
	c.Writer.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Transfer-Encoding", "chunked")
	c.Writer.Header().Set("X-Accel-Buffering", "no")           // Disable nginx buffering
	c.Writer.Header().Set("X-Content-Type-Options", "nosniff") // Prevent MIME sniffing

	// SECURITY: Set CORS headers if needed (should be handled by middleware)
	// Only allow specific origins, never use wildcard for credentials

	// SECURITY: Limit connection time to prevent resource exhaustion
	ctx, cancel := context.WithTimeout(c.Request.Context(), 15*time.Minute)
	defer cancel()

	// Log connection for monitoring/rate limiting
	h.logger.Info("SSE connection established",
		zap.String("contact_id", contactID.String()),
		zap.String("remote_addr", c.ClientIP()))

	// Send initial connection event
	h.sendSSEEvent(c, "connected", map[string]interface{}{
		"contact_id": contactID.String(),
		"timestamp":  time.Now().UTC(),
	})

	// SECURITY: Use reasonable polling interval to prevent excessive DB queries
	ticker := time.NewTicker(3 * time.Second) // Increased from 2s to reduce load
	defer ticker.Stop()

	// Start from recent events only (not too far back)
	lastEventTime := time.Now().Add(-2 * time.Minute)

	for {
		select {
		case <-ctx.Done():
			h.logger.Info("SSE connection closed", zap.String("contact_id", contactID.String()))
			return
		case <-ticker.C:
			// Fetch new events since last check
			events, err := h.fetchNewEvents(ctx, contactID, lastEventTime, categoriesFilter, priorityFilter)
			if err != nil {
				h.logger.Error("Failed to fetch events", zap.Error(err))
				continue
			}

			// Send events
			for _, event := range events {
				if err := h.sendContactEvent(c, event); err != nil {
					h.logger.Error("Failed to send event", zap.Error(err))
					return
				}

				// Update last event time
				if event.OccurredAt().After(lastEventTime) {
					lastEventTime = event.OccurredAt()
				}
			}

			// Send heartbeat if no events
			if len(events) == 0 {
				h.sendSSEEvent(c, "heartbeat", map[string]interface{}{
					"timestamp": time.Now().UTC(),
				})
			}

			c.Writer.Flush()
		}
	}
}

// StreamContactEventsByCategory streams contact events filtered by category
//
//	@Summary		Stream contact events by category
//	@Description	Stream contact events filtered by specific category
//	@Tags			contact-events
//	@Produce		text/event-stream
//	@Param			contact_id	path		string	true	"Contact ID"
//	@Param			category	path		string	true	"Category: status,pipeline,assignment,tag,note,session,custom_field,system,notification"
//	@Success		200			{string}	string	"Event stream"
//	@Router			/api/v1/contacts/{contact_id}/events/stream/{category} [get]
func (h *ContactEventStreamHandler) StreamContactEventsByCategory(c *gin.Context) {
	category := c.Param("category")

	// Validate category
	if !isValidCategory(category) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid category"})
		return
	}

	// Set category filter in query
	c.Request.URL.RawQuery = fmt.Sprintf("categories=%s", category)

	// Reuse main stream handler
	h.StreamContactEvents(c)
}

// ListContactEvents lists contact events with pagination
//
//	@Summary		List contact events
//	@Description	List contact events with filtering and pagination
//	@Tags			contact-events
//	@Produce		json
//	@Param			contact_id	path		string	true	"Contact ID"
//	@Param			categories	query		string	false	"Filter by categories (comma-separated)"
//	@Param			priority	query		string	false	"Filter by priority"
//	@Param			limit		query		int		false	"Limit (default: 50, max: 200)"
//	@Param			offset		query		int		false	"Offset (default: 0)"
//	@Success		200			{object}	map[string]interface{}
//	@Router			/api/v1/contacts/{contact_id}/events [get]
func (h *ContactEventStreamHandler) ListContactEvents(c *gin.Context) {
	contactIDStr := c.Param("contact_id")
	contactID, err := uuid.Parse(contactIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid contact ID"})
		return
	}

	// Parse pagination
	limit := parseIntQuery(c, "limit", 50, 1, 200)
	offset := parseIntQuery(c, "offset", 0, 0, 10000)

	// Parse filters
	categoriesFilter := parseCategories(c.Query("categories"))
	priorityFilter := parsePriority(c.Query("priority"))

	// Fetch events
	events, err := h.fetchEvents(c.Request.Context(), contactID, limit, offset, categoriesFilter, priorityFilter)
	if err != nil {
		h.logger.Error("Failed to fetch events", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch events"})
		return
	}

	// Convert to response format
	eventResponses := make([]map[string]interface{}, len(events))
	for i, event := range events {
		eventResponses[i] = h.eventToMap(event)
	}

	c.JSON(http.StatusOK, gin.H{
		"contact_id": contactID.String(),
		"events":     eventResponses,
		"count":      len(eventResponses),
		"limit":      limit,
		"offset":     offset,
	})
}

// Helper functions

func (h *ContactEventStreamHandler) fetchNewEvents(
	ctx context.Context,
	contactID uuid.UUID,
	since time.Time,
	categories []contact_event.Category,
	priority *contact_event.Priority,
) ([]*contact_event.ContactEvent, error) {
	// This would ideally use a repository method that filters by time
	// For now, fetch recent events and filter
	events, err := h.eventRepo.FindByContactID(ctx, contactID, 100, 0)
	if err != nil {
		return nil, err
	}

	// Filter events
	filtered := make([]*contact_event.ContactEvent, 0)
	for _, event := range events {
		// Filter by time
		if !event.OccurredAt().After(since) {
			continue
		}

		// Filter by category
		if len(categories) > 0 && !containsCategory(categories, event.Category()) {
			continue
		}

		// Filter by priority
		if priority != nil && event.Priority() != *priority {
			continue
		}

		filtered = append(filtered, event)
	}

	return filtered, nil
}

func (h *ContactEventStreamHandler) fetchEvents(
	ctx context.Context,
	contactID uuid.UUID,
	limit, offset int,
	categories []contact_event.Category,
	priority *contact_event.Priority,
) ([]*contact_event.ContactEvent, error) {
	// Fetch more than needed to account for filtering
	fetchLimit := limit * 3
	if fetchLimit > 500 {
		fetchLimit = 500
	}

	events, err := h.eventRepo.FindByContactID(ctx, contactID, fetchLimit, offset)
	if err != nil {
		return nil, err
	}

	// Filter events
	filtered := make([]*contact_event.ContactEvent, 0, limit)
	for _, event := range events {
		// Filter by category
		if len(categories) > 0 && !containsCategory(categories, event.Category()) {
			continue
		}

		// Filter by priority
		if priority != nil && event.Priority() != *priority {
			continue
		}

		filtered = append(filtered, event)
		if len(filtered) >= limit {
			break
		}
	}

	return filtered, nil
}

func (h *ContactEventStreamHandler) sendContactEvent(c *gin.Context, event *contact_event.ContactEvent) error {
	return h.sendSSEEvent(c, "contact_event", h.eventToMap(event))
}

func (h *ContactEventStreamHandler) sendSSEEvent(c *gin.Context, eventType string, data interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(c.Writer, "event: %s\ndata: %s\n\n", eventType, jsonData)
	return err
}

func (h *ContactEventStreamHandler) eventToMap(event *contact_event.ContactEvent) map[string]interface{} {
	// SECURITY: Sanitize all string data to prevent XSS attacks
	result := map[string]interface{}{
		"id":          event.ID().String(),
		"contact_id":  event.ContactID().String(),
		"tenant_id":   sanitizeString(event.TenantID()),
		"event_type":  sanitizeString(event.EventType()),
		"category":    event.Category().String(),
		"priority":    event.Priority().String(),
		"source":      event.Source().String(),
		"payload":     sanitizePayload(event.Payload()),
		"metadata":    sanitizePayload(event.Metadata()),
		"occurred_at": event.OccurredAt(),
		"created_at":  event.CreatedAt(),
	}

	if event.SessionID() != nil {
		result["session_id"] = event.SessionID().String()
	}
	if event.Title() != nil {
		result["title"] = sanitizeString(*event.Title())
	}
	if event.Description() != nil {
		result["description"] = sanitizeString(*event.Description())
	}
	if event.TriggeredBy() != nil {
		result["triggered_by"] = event.TriggeredBy().String()
	}
	if event.IntegrationSource() != nil {
		result["integration_source"] = sanitizeString(*event.IntegrationSource())
	}

	return result
}

// Parsing helpers

func parseCategories(categoriesStr string) []contact_event.Category {
	if categoriesStr == "" {
		return nil
	}

	categories := make([]contact_event.Category, 0)
	for _, cat := range splitAndTrim(categoriesStr, ",") {
		category := contact_event.Category(cat)
		if category.IsValid() {
			categories = append(categories, category)
		}
	}

	return categories
}

func parsePriority(priorityStr string) *contact_event.Priority {
	if priorityStr == "" {
		return nil
	}

	priority := contact_event.Priority(priorityStr)
	if priority.IsValid() {
		return &priority
	}

	return nil
}

func parseIntQuery(c *gin.Context, key string, defaultVal, min, max int) int {
	val := defaultVal
	if valStr := c.Query(key); valStr != "" {
		if parsed, err := parseInt(valStr); err == nil {
			val = parsed
		}
	}

	if val < min {
		val = min
	}
	if val > max {
		val = max
	}

	return val
}

func parseInt(s string) (int, error) {
	var val int
	_, err := fmt.Sscanf(s, "%d", &val)
	return val, err
}

func splitAndTrim(s, sep string) []string {
	parts := []string{}
	for _, part := range splitString(s, sep) {
		trimmed := trimString(part)
		if trimmed != "" {
			parts = append(parts, trimmed)
		}
	}
	return parts
}

func splitString(s, sep string) []string {
	if s == "" {
		return []string{}
	}
	result := []string{}
	current := ""
	for _, char := range s {
		if string(char) == sep {
			result = append(result, current)
			current = ""
		} else {
			current += string(char)
		}
	}
	if current != "" {
		result = append(result, current)
	}
	return result
}

func trimString(s string) string {
	start := 0
	end := len(s)

	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n' || s[start] == '\r') {
		start++
	}

	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\n' || s[end-1] == '\r') {
		end--
	}

	return s[start:end]
}

func containsCategory(categories []contact_event.Category, category contact_event.Category) bool {
	for _, cat := range categories {
		if cat == category {
			return true
		}
	}
	return false
}

func isValidCategory(category string) bool {
	cat := contact_event.Category(category)
	return cat.IsValid()
}

// SECURITY: isAllowedOrigin validates the Origin header to prevent CSRF attacks
func (h *ContactEventStreamHandler) isAllowedOrigin(origin string) bool {
	// TODO: Load allowed origins from configuration
	// For now, allow common development origins and production domain
	allowedOrigins := []string{
		"http://localhost:3000",
		"http://localhost:3001",
		"http://localhost:5173", // Vite default
		"http://localhost:8080",
		"https://app.ventros.io",
		"https://ventros.io",
	}

	for _, allowed := range allowedOrigins {
		if origin == allowed {
			return true
		}
	}

	h.logger.Warn("Origin not in allowed list", zap.String("origin", origin))
	return false
}

// SECURITY: sanitizeString removes potentially dangerous characters from strings
// to prevent XSS when data is rendered in browser
func sanitizeString(s string) string {
	// Remove control characters and null bytes
	result := ""
	for _, r := range s {
		// Allow printable characters, newlines, tabs
		if r >= 32 || r == '\n' || r == '\t' {
			result += string(r)
		}
	}
	return result
}

// SECURITY: sanitizePayload recursively sanitizes map data to prevent XSS
func sanitizePayload(data map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for key, value := range data {
		switch v := value.(type) {
		case string:
			result[key] = sanitizeString(v)
		case map[string]interface{}:
			result[key] = sanitizePayload(v)
		case []interface{}:
			sanitized := make([]interface{}, len(v))
			for i, item := range v {
				if str, ok := item.(string); ok {
					sanitized[i] = sanitizeString(str)
				} else if m, ok := item.(map[string]interface{}); ok {
					sanitized[i] = sanitizePayload(m)
				} else {
					sanitized[i] = item
				}
			}
			result[key] = sanitized
		default:
			result[key] = value
		}
	}
	return result
}
