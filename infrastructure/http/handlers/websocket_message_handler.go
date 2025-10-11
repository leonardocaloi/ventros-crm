package handlers

import (
	"net/http"

	"github.com/caloi/ventros-crm/infrastructure/http/middleware"
	ws "github.com/caloi/ventros-crm/infrastructure/websocket"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

// WebSocketMessageHandler gerencia conexões WebSocket para mensagens
type WebSocketMessageHandler struct {
	hub      *ws.Hub
	upgrader websocket.Upgrader
	logger   *zap.Logger
}

// NewWebSocketMessageHandler cria novo handler
func NewWebSocketMessageHandler(hub *ws.Hub, production bool, logger *zap.Logger) *WebSocketMessageHandler {
	allowedOrigins := ws.GetAllowedOrigins(production)

	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		// SECURITY: Origin validation para prevenir CSWSH
		CheckOrigin: func(r *http.Request) bool {
			origin := r.Header.Get("Origin")
			isAllowed := ws.ValidateOrigin(origin, allowedOrigins)

			if !isAllowed {
				logger.Warn("WebSocket connection rejected - invalid origin",
					zap.String("origin", origin),
					zap.String("remote_addr", r.RemoteAddr))
			}

			return isAllowed
		},
	}

	return &WebSocketMessageHandler{
		hub:      hub,
		upgrader: upgrader,
		logger:   logger,
	}
}

// HandleWebSocket faz upgrade HTTP → WebSocket
//
//	@Summary		WebSocket connection for real-time messages
//	@Description	Establishes WebSocket connection for bi-directional real-time messaging
//	@Tags			websocket
//	@Security		BearerAuth
//	@Param			token	query	string	false	"Authentication token (alternative to Bearer header)"
//	@Success		101		"Switching Protocols - WebSocket connection established"
//	@Failure		401		{object}	map[string]string	"Unauthorized - invalid or missing token"
//	@Failure		403		{object}	map[string]string	"Forbidden - origin not allowed"
//	@Failure		500		{object}	map[string]string	"Internal server error"
//	@Router			/api/v1/ws/messages [get]
func (h *WebSocketMessageHandler) HandleWebSocket(c *gin.Context) {
	// SECURITY: Verificar autenticação
	authCtx, exists := middleware.GetAuthContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Authentication required",
			"hint":  "Use Authorization: Bearer <token> header or ?token=<token> query param",
		})
		return
	}

	// SECURITY: Log de conexão para auditoria
	h.logger.Info("WebSocket connection attempt",
		zap.String("user_id", authCtx.UserID.String()),
		zap.String("tenant_id", authCtx.TenantID),
		zap.String("remote_addr", c.ClientIP()),
		zap.String("origin", c.GetHeader("Origin")),
		zap.String("user_agent", c.GetHeader("User-Agent")))

	// Upgrade HTTP → WebSocket
	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		h.logger.Error("Failed to upgrade to WebSocket",
			zap.Error(err),
			zap.String("user_id", authCtx.UserID.String()))
		return
	}

	// Criar cliente WebSocket
	client := ws.NewClient(h.hub, conn, authCtx.UserID, authCtx.TenantID, authCtx.ProjectID, h.logger)

	// Registrar no hub
	h.hub.Register <- client

	// Iniciar goroutines de leitura/escrita
	go client.WritePump()
	go client.ReadPump()

	h.logger.Info("WebSocket connection established",
		zap.String("client_id", client.ID()),
		zap.String("user_id", authCtx.UserID.String()))
}

// GetStats retorna estatísticas do WebSocket
//
//	@Summary		WebSocket statistics
//	@Description	Returns current WebSocket connection statistics
//	@Tags			websocket
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	map[string]interface{}
//	@Router			/api/v1/ws/stats [get]
func (h *WebSocketMessageHandler) GetStats(c *gin.Context) {
	stats := h.hub.GetStats()
	c.JSON(http.StatusOK, stats)
}
