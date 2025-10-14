package routes

import (
	"time"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/ventros/crm/infrastructure/health"
	"github.com/ventros/crm/infrastructure/http/handlers"
	"github.com/ventros/crm/infrastructure/http/middleware"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// SetupRoutesMinimal configura apenas as rotas básicas (health, queue, session, contact)
func SetupRoutesMinimal(router *gin.Engine, logger *zap.Logger, healthChecker *health.HealthChecker, queueHandler *handlers.QueueHandler, sessionHandler *handlers.SessionHandler, contactHandler *handlers.ContactHandler, authMiddleware *middleware.AuthMiddleware, rlsMiddleware *middleware.RLSMiddleware) {
	// Middlewares
	router.Use(gin.Recovery())
	router.Use(LoggerMiddleware(logger))
	router.Use(CORSMiddleware())

	// Health check handlers
	healthHandler := handlers.NewHealthHandler(logger, healthChecker)

	// Health endpoints - General
	router.GET("/health", healthHandler.Health)
	router.GET("/ready", healthHandler.Ready)
	router.GET("/live", healthHandler.Live)

	// Health endpoints - Specific components
	healthGroup := router.Group("/health")
	{
		healthGroup.GET("/database", healthHandler.CheckDatabase)
		healthGroup.GET("/migrations", healthHandler.CheckMigrations)
		healthGroup.GET("/redis", healthHandler.CheckRedis)
		healthGroup.GET("/rabbitmq", healthHandler.CheckRabbitMQ)
		healthGroup.GET("/temporal", healthHandler.CheckTemporal)
	}

	// Swagger documentation
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API v1 CRM routes (minimal)
	v1 := router.Group("/api/v1/crm")
	{
		// Queue management
		v1.GET("/queues", queueHandler.ListQueues)

		// Session management
		sessions := v1.Group("/sessions")
		{
			sessions.GET("/:id", sessionHandler.GetSession)
			sessions.GET("/", sessionHandler.ListSessions)
		}

		// Contact management (protected routes)
		contacts := v1.Group("/contacts")
		contacts.Use(authMiddleware.Authenticate())
		contacts.Use(rlsMiddleware.SetUserContext())
		{
			contacts.GET("/:id", contactHandler.GetContact)
			contacts.GET("/", contactHandler.ListContacts)
		}
	}
}

func SetupRoutes(router *gin.Engine, logger *zap.Logger, healthChecker *health.HealthChecker, wahaHandler *handlers.WAHAWebhookHandler, webhookHandler *handlers.WebhookSubscriptionHandler, queueHandler *handlers.QueueHandler, sessionHandler *handlers.SessionHandler, contactHandler *handlers.ContactHandler, pipelineHandler *handlers.PipelineHandler, projectHandler *handlers.ProjectHandler, agentHandler *handlers.AgentHandler, messageHandler *handlers.MessageHandler, channelHandler *handlers.ChannelHandler, trackingHandler *handlers.TrackingHandler) {
	// Middlewares
	router.Use(gin.Recovery())
	router.Use(LoggerMiddleware(logger))
	router.Use(CORSMiddleware())

	// Health check handlers
	healthHandler := handlers.NewHealthHandler(logger, healthChecker)

	// Webhook inbound routes (sem auth para receber de serviços externos)
	// Padrão indústria: /api/v1/webhooks/:webhook_id
	webhooks := router.Group("/api/v1/webhooks")
	{
		webhooks.POST("/:webhook_id", wahaHandler.ReceiveWebhook)
		webhooks.GET("/info", wahaHandler.GetWebhookInfo)
	}

	// Health endpoints - General
	router.GET("/health", healthHandler.Health)
	router.GET("/ready", healthHandler.Ready)
	router.GET("/live", healthHandler.Live)

	// Health endpoints - Specific components
	healthGroup := router.Group("/health")
	{
		healthGroup.GET("/database", healthHandler.CheckDatabase)
		healthGroup.GET("/migrations", healthHandler.CheckMigrations)
		healthGroup.GET("/redis", healthHandler.CheckRedis)
		healthGroup.GET("/rabbitmq", healthHandler.CheckRabbitMQ)
		healthGroup.GET("/temporal", healthHandler.CheckTemporal)
	}

	// Swagger documentation
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Queue management routes
		v1.GET("/queues", queueHandler.ListQueues)

		// Webhook subscription routes
		webhookSubs := v1.Group("/webhook-subscriptions")
		{
			webhookSubs.GET("/available-events", webhookHandler.GetAvailableEvents)
			webhookSubs.POST("", webhookHandler.CreateWebhook)
			webhookSubs.GET("", webhookHandler.ListWebhooks)
			webhookSubs.GET("/:id", webhookHandler.GetWebhook)
			webhookSubs.PUT("/:id", webhookHandler.UpdateWebhook)
			webhookSubs.DELETE("/:id", webhookHandler.DeleteWebhook)
		}

		// Project routes
		projects := v1.Group("/projects")
		{
			projects.GET("", projectHandler.ListProjects)
			projects.POST("", projectHandler.CreateProject)
			projects.GET("/:id", projectHandler.GetProject)
			projects.PUT("/:id", projectHandler.UpdateProject)
			projects.DELETE("/:id", projectHandler.DeleteProject)
		}

		// Contact routes
		contacts := v1.Group("/contacts")
		{
			contacts.GET("/search", contactHandler.SearchContacts)         // Must be before /:id
			contacts.GET("/advanced", contactHandler.ListContactsAdvanced) // Must be before /:id
			contacts.GET("", contactHandler.ListContacts)
			contacts.POST("", contactHandler.CreateContact)
			contacts.GET("/:id", contactHandler.GetContact)
			contacts.PUT("/:id", contactHandler.UpdateContact)
			contacts.DELETE("/:id", contactHandler.DeleteContact)

			// Tracking routes dentro de contacts
			contacts.GET("/:contact_id/trackings", trackingHandler.GetContactTrackings)
		}

		// Channel routes
		channels := v1.Group("/channels")
		{
			channels.GET("", channelHandler.ListChannels)
			channels.POST("", channelHandler.CreateChannel)
			channels.GET("/:id", channelHandler.GetChannel)
			channels.POST("/:id/activate", channelHandler.ActivateChannel)
			channels.POST("/:id/deactivate", channelHandler.DeactivateChannel)
			channels.DELETE("/:id", channelHandler.DeleteChannel)

			// Webhook endpoints for channels
			channels.GET("/:id/webhook-url", channelHandler.GetChannelWebhookURL)
			channels.POST("/:id/configure-webhook", channelHandler.ConfigureChannelWebhook)
			channels.GET("/:id/webhook-info", channelHandler.GetChannelWebhookInfo)
		}

		// Session routes (TODO: add auth when SetupRoutes is used)
		sessions := v1.Group("/sessions")
		{
			sessions.GET("", sessionHandler.ListSessions)
			sessions.GET("/:id", sessionHandler.GetSession)
			sessions.GET("/stats", sessionHandler.GetSessionStats)
		}

		// Pipeline routes
		pipelines := v1.Group("/pipelines")
		{
			pipelines.GET("", pipelineHandler.ListPipelines)
			pipelines.POST("", pipelineHandler.CreatePipeline)
			pipelines.GET("/:id", pipelineHandler.GetPipeline)

			// Status routes within pipelines
			pipelines.POST("/:id/statuses", pipelineHandler.CreateStatus)

			// Contact status routes
			pipelines.PUT("/:pipeline_id/contacts/:contact_id/status", pipelineHandler.ChangeContactStatus)
			pipelines.GET("/:pipeline_id/contacts/:contact_id/status", pipelineHandler.GetContactStatus)
		}

		// Tracking routes
		trackings := v1.Group("/trackings")
		{
			trackings.GET("/enums", trackingHandler.GetTrackingEnums) // Must be before /:id
			trackings.POST("/encode", trackingHandler.EncodeTracking)
			trackings.POST("/decode", trackingHandler.DecodeTracking)
			trackings.POST("", trackingHandler.CreateTracking)
			trackings.GET("/:id", trackingHandler.GetTracking)
		}

		// Message routes
		// messages := v1.Group("/messages")
		// {
		// }
	}
}

// SetupRoutesBasic configura as rotas básicas sem pipeline handler (temporário)
func SetupRoutesBasic(router *gin.Engine, logger *zap.Logger, healthChecker *health.HealthChecker, wahaHandler *handlers.WAHAWebhookHandler, webhookHandler *handlers.WebhookSubscriptionHandler, queueHandler *handlers.QueueHandler, sessionHandler *handlers.SessionHandler, contactHandler *handlers.ContactHandler, authMiddleware *middleware.AuthMiddleware, rlsMiddleware *middleware.RLSMiddleware) {
	// Middlewares
	router.Use(gin.Recovery())
	router.Use(LoggerMiddleware(logger))
	router.Use(CORSMiddleware())

	// Health check handlers
	healthHandler := handlers.NewHealthHandler(logger, healthChecker)

	// Webhook inbound routes (sem auth para receber de serviços externos)
	// Padrão indústria: /api/v1/webhooks/:webhook_id
	webhooks := router.Group("/api/v1/webhooks")
	{
		webhooks.POST("/:webhook_id", wahaHandler.ReceiveWebhook)
		webhooks.GET("/info", wahaHandler.GetWebhookInfo)
	}

	// Health endpoints - General
	router.GET("/health", healthHandler.Health)
	router.GET("/ready", healthHandler.Ready)
	router.GET("/live", healthHandler.Live)

	// Health endpoints - Specific components
	healthGroup := router.Group("/health")
	{
		healthGroup.GET("/database", healthHandler.CheckDatabase)
		healthGroup.GET("/migrations", healthHandler.CheckMigrations)
		healthGroup.GET("/redis", healthHandler.CheckRedis)
		healthGroup.GET("/rabbitmq", healthHandler.CheckRabbitMQ)
		healthGroup.GET("/temporal", healthHandler.CheckTemporal)
	}

	// Swagger documentation
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Queue management routes
		v1.GET("/queues", queueHandler.ListQueues)

		// Webhook subscription routes
		webhookSubs := v1.Group("/webhook-subscriptions")
		webhookSubs.Use(authMiddleware.Authenticate())
		webhookSubs.Use(rlsMiddleware.SetUserContext())
		{
			webhookSubs.GET("/available-events", webhookHandler.GetAvailableEvents)
			webhookSubs.POST("", webhookHandler.CreateWebhook)
			webhookSubs.GET("", webhookHandler.ListWebhooks)
			webhookSubs.GET("/:id", webhookHandler.GetWebhook)
			webhookSubs.PUT("/:id", webhookHandler.UpdateWebhook)
			webhookSubs.DELETE("/:id", webhookHandler.DeleteWebhook)
		}

		// Contact routes
		contacts := v1.Group("/contacts")
		contacts.Use(authMiddleware.Authenticate())
		contacts.Use(rlsMiddleware.SetUserContext())
		{
			contacts.GET("/search", contactHandler.SearchContacts)         // Must be before /:id
			contacts.GET("/advanced", contactHandler.ListContactsAdvanced) // Must be before /:id
			contacts.GET("", contactHandler.ListContacts)
			contacts.POST("", contactHandler.CreateContact)
			contacts.GET("/:id", contactHandler.GetContact)
			contacts.PUT("/:id", contactHandler.UpdateContact)
			contacts.DELETE("/:id", contactHandler.DeleteContact)

			// Nested session routes under contact (using :id for contact)
			contacts.GET("/:id/sessions", sessionHandler.ListSessions)
			contacts.GET("/:id/sessions/:session_id", sessionHandler.GetSession)

			// Pipeline status routes under contact
			// PUT /api/v1/contacts/:id/pipelines/:pipeline_id/status
			contacts.PUT("/:id/pipelines/:pipeline_id/status", contactHandler.ChangePipelineStatus)
		}

		// Session routes (protected) - global with required filters
		sessions := v1.Group("/sessions")
		sessions.Use(authMiddleware.Authenticate())
		sessions.Use(rlsMiddleware.SetUserContext())
		{
			sessions.GET("", sessionHandler.ListSessions) // Requires ?contact_id or ?channel_id
			sessions.GET("/:id", sessionHandler.GetSession)
			sessions.POST("/:id/close", sessionHandler.CloseSession) // Agente encerra sessão manualmente
			sessions.GET("/stats", sessionHandler.GetSessionStats)
		}

		// Message routes
		// messages := v1.Group("/messages")
		// {
		// }
	}
}

// SetupRoutesBasicWithTest configura as rotas básicas com endpoints de teste, auth, channels, projects, pipelines, messages, chats e WebSocket
// LEGACY: Mantido para compatibilidade
func SetupRoutesBasicWithTest(router *gin.Engine, logger *zap.Logger, healthChecker *health.HealthChecker, authHandler *handlers.AuthHandler, automationHandler *handlers.AutomationHandler, broadcastHandler *handlers.BroadcastHandler, sequenceHandler *handlers.SequenceHandler, campaignHandler *handlers.CampaignHandler, channelHandler *handlers.ChannelHandler, projectHandler *handlers.ProjectHandler, pipelineHandler *handlers.PipelineHandler, wahaHandler *handlers.WAHAWebhookHandler, webhookHandler *handlers.WebhookSubscriptionHandler, queueHandler *handlers.QueueHandler, sessionHandler *handlers.SessionHandler, contactHandler *handlers.ContactHandler, trackingHandler *handlers.TrackingHandler, messageHandler *handlers.MessageHandler, chatHandler *handlers.ChatHandler, agentHandler *handlers.AgentHandler, noteHandler *handlers.NoteHandler, automationDiscoveryHandler *handlers.AutomationDiscoveryHandler, websocketHandler *handlers.WebSocketMessageHandler, wsRateLimiter *middleware.WebSocketRateLimiter, gormDB *gorm.DB, authMiddleware *middleware.AuthMiddleware, wsAuthMiddleware *middleware.WebSocketAuthMiddleware, rlsMiddleware *middleware.RLSMiddleware) {
	// Add GORM context middleware FIRST (before any other middleware)
	router.Use(middleware.GORMContextMiddleware(gormDB))

	// Add Correlation ID middleware for distributed tracing
	router.Use(middleware.CorrelationIDMiddleware())

	// Use the basic setup first
	SetupRoutesBasic(router, logger, healthChecker, wahaHandler, webhookHandler, queueHandler, sessionHandler, contactHandler, authMiddleware, rlsMiddleware)

	// Auth routes (INDEPENDENT - não faz parte do CRM)
	// Moved from /api/v1/crm/auth to /api/v1/auth
	// Rate limit: 10 requests per minute for auth endpoints (prevent brute force)
	authRoutes := router.Group("/api/v1/auth")
	authRoutes.Use(middleware.AuthRateLimitMiddleware()) // 10 req/min
	{
		authRoutes.POST("/register", authHandler.CreateUser)
		authRoutes.POST("/login", authHandler.Login)
		authRoutes.GET("/info", authHandler.GetAuthInfo)

		// Protected auth routes
		authProtected := authRoutes.Group("")
		authProtected.Use(authMiddleware.Authenticate())
		{
			authProtected.GET("/profile", authHandler.GetProfile)
			authProtected.POST("/api-key", authHandler.GenerateAPIKey)
		}
	}

	// Add automation routes (cross-product AUTOMATION product - NOT CRM)
	// Rate limit: 1000 requests per minute for authenticated API endpoints
	automation := router.Group("/api/v1/automation")
	automation.Use(authMiddleware.Authenticate())
	automation.Use(rlsMiddleware.SetUserContext())
	automation.Use(middleware.UserBasedRateLimitMiddleware("1000-M")) // 1000 req/min per user
	{
		// Discovery endpoints (metadata)
		automation.GET("/types", automationHandler.GetAutomationTypes)
		automation.GET("/actions", automationHandler.GetAvailableActions)
		automation.GET("/operators", automationHandler.GetAvailableOperators)

		// CRUD endpoints
		automation.GET("", automationHandler.ListAutomations)
		automation.POST("", automationHandler.CreateAutomation)
		automation.GET("/:id", automationHandler.GetAutomation)
		automation.PUT("/:id", automationHandler.UpdateAutomation)
		automation.DELETE("/:id", automationHandler.DeleteAutomation)

		// Broadcast endpoints
		broadcasts := automation.Group("/broadcasts")
		{
			broadcasts.GET("", broadcastHandler.ListBroadcasts)
			broadcasts.POST("", broadcastHandler.CreateBroadcast)
			broadcasts.GET("/:id", broadcastHandler.GetBroadcast)
			broadcasts.PUT("/:id", broadcastHandler.UpdateBroadcast)
			broadcasts.DELETE("/:id", broadcastHandler.DeleteBroadcast)
			broadcasts.POST("/:id/schedule", broadcastHandler.ScheduleBroadcast)
			broadcasts.POST("/:id/execute", broadcastHandler.ExecuteBroadcast)
			broadcasts.POST("/:id/cancel", broadcastHandler.CancelBroadcast)
			broadcasts.GET("/:id/stats", broadcastHandler.GetBroadcastStats)
		}

		// Sequence endpoints
		sequences := automation.Group("/sequences")
		{
			sequences.GET("", sequenceHandler.ListSequences)
			sequences.POST("", sequenceHandler.CreateSequence)
			sequences.GET("/:id", sequenceHandler.GetSequence)
			sequences.PUT("/:id", sequenceHandler.UpdateSequence)
			sequences.DELETE("/:id", sequenceHandler.DeleteSequence)
			sequences.POST("/:id/activate", sequenceHandler.ActivateSequence)
			sequences.POST("/:id/pause", sequenceHandler.PauseSequence)
			sequences.POST("/:id/resume", sequenceHandler.ResumeSequence)
			sequences.POST("/:id/archive", sequenceHandler.ArchiveSequence)
			sequences.GET("/:id/stats", sequenceHandler.GetSequenceStats)
			sequences.POST("/:id/enroll", sequenceHandler.EnrollContact)
			sequences.GET("/:id/enrollments", sequenceHandler.ListEnrollments)
		}

		// Campaign endpoints
		campaigns := automation.Group("/campaigns")
		{
			campaigns.GET("", campaignHandler.ListCampaigns)
			campaigns.POST("", campaignHandler.CreateCampaign)
			campaigns.GET("/:id", campaignHandler.GetCampaign)
			campaigns.PUT("/:id", campaignHandler.UpdateCampaign)
			campaigns.DELETE("/:id", campaignHandler.DeleteCampaign)
			campaigns.POST("/:id/activate", campaignHandler.ActivateCampaign)
			campaigns.POST("/:id/schedule", campaignHandler.ScheduleCampaign)
			campaigns.POST("/:id/pause", campaignHandler.PauseCampaign)
			campaigns.POST("/:id/resume", campaignHandler.ResumeCampaign)
			campaigns.POST("/:id/complete", campaignHandler.CompleteCampaign)
			campaigns.POST("/:id/archive", campaignHandler.ArchiveCampaign)
			campaigns.GET("/:id/stats", campaignHandler.GetCampaignStats)
			campaigns.POST("/:id/enroll", campaignHandler.EnrollContact)
			campaigns.GET("/:id/enrollments", campaignHandler.ListEnrollments)
		}
	}

	// Add channel routes (all protected)
	// Rate limit: 1000 requests per minute for authenticated API endpoints
	channels := router.Group("/api/v1/crm/channels")
	channels.Use(authMiddleware.Authenticate())
	channels.Use(rlsMiddleware.SetUserContext())
	channels.Use(middleware.UserBasedRateLimitMiddleware("1000-M")) // 1000 req/min per user
	{
		channels.GET("", channelHandler.ListChannels)
		channels.POST("", channelHandler.CreateChannel)
		channels.GET("/:id", channelHandler.GetChannel)
		channels.POST("/:id/activate", channelHandler.ActivateChannel)
		channels.POST("/:id/deactivate", channelHandler.DeactivateChannel)
		channels.DELETE("/:id", channelHandler.DeleteChannel)

		// Webhook endpoints for channels
		channels.GET("/:id/webhook-url", channelHandler.GetChannelWebhookURL)
		channels.POST("/:id/configure-webhook", channelHandler.ConfigureChannelWebhook)
		channels.GET("/:id/webhook-info", channelHandler.GetChannelWebhookInfo)

		// WAHA-specific endpoints
		channels.POST("/:id/activate-waha", channelHandler.ActivateWAHAChannel)
		channels.POST("/:id/import-history", channelHandler.ImportWAHAHistory)
		channels.GET("/:id/import-status", channelHandler.GetWAHAImportStatus)

		// Nested session routes under channel (using :id for channel)
		channels.GET("/:id/sessions", sessionHandler.ListSessions)
		channels.GET("/:id/sessions/:session_id", sessionHandler.GetSession)
	}

	// Add project routes (all protected)
	projects := router.Group("/api/v1/crm/projects")
	projects.Use(authMiddleware.Authenticate())
	projects.Use(rlsMiddleware.SetUserContext())
	{
		projects.GET("/search", projectHandler.SearchProjects)         // Must be before /:id
		projects.GET("/advanced", projectHandler.ListProjectsAdvanced) // Must be before /:id
		projects.GET("", projectHandler.ListProjects)
		projects.POST("", projectHandler.CreateProject)
		projects.GET("/:id", projectHandler.GetProject)
		projects.PUT("/:id", projectHandler.UpdateProject)
		projects.DELETE("/:id", projectHandler.DeleteProject)
	}

	// Add pipeline routes (all protected)
	pipelines := router.Group("/api/v1/crm/pipelines")
	pipelines.Use(authMiddleware.Authenticate())
	pipelines.Use(rlsMiddleware.SetUserContext())
	{
		pipelines.GET("/search", pipelineHandler.SearchPipelines)         // Must be before /:id
		pipelines.GET("/advanced", pipelineHandler.ListPipelinesAdvanced) // Must be before /:id
		pipelines.GET("", pipelineHandler.ListPipelines)
		pipelines.POST("", pipelineHandler.CreatePipeline)
		pipelines.GET("/:id", pipelineHandler.GetPipeline)

		// Status routes within pipelines
		pipelines.POST("/:id/statuses", pipelineHandler.CreateStatus)

		// Contact status routes (usando :id para pipeline)
		pipelines.PUT("/:id/contacts/:contact_id/status", pipelineHandler.ChangeContactStatus)
		pipelines.GET("/:id/contacts/:contact_id/status", pipelineHandler.GetContactStatus)
	}

	// Add session routes (all protected) - advanced query endpoints
	sessions := router.Group("/api/v1/crm/sessions")
	sessions.Use(authMiddleware.Authenticate())
	sessions.Use(rlsMiddleware.SetUserContext())
	{
		sessions.GET("/search", sessionHandler.SearchSessions)         // Must be before /:id
		sessions.GET("/advanced", sessionHandler.ListSessionsAdvanced) // Must be before /:id
	}

	// Add tracking routes (all protected)
	trackings := router.Group("/api/v1/crm/trackings")
	trackings.Use(authMiddleware.Authenticate())
	trackings.Use(rlsMiddleware.SetUserContext())
	{
		trackings.GET("/enums", trackingHandler.GetTrackingEnums) // Must be before /:id
		trackings.POST("/encode", trackingHandler.EncodeTracking)
		trackings.POST("/decode", trackingHandler.DecodeTracking)
		trackings.POST("", trackingHandler.CreateTracking)
		trackings.GET("/:id", trackingHandler.GetTracking)
	}

	// Add message routes (all protected)
	messages := router.Group("/api/v1/crm/messages")
	messages.Use(authMiddleware.Authenticate())
	messages.Use(rlsMiddleware.SetUserContext())
	{
		messages.GET("/search", messageHandler.SearchMessages)         // Must be before /:id
		messages.GET("/advanced", messageHandler.ListMessagesAdvanced) // Must be before /:id
		messages.GET("", messageHandler.ListMessages)
		messages.POST("", messageHandler.CreateMessage)
		messages.POST("/send", messageHandler.SendMessage)
		messages.POST("/confirm-delivery", messageHandler.ConfirmMessageDelivery)
		messages.GET("/:id", messageHandler.GetMessage)
		messages.PUT("/:id", messageHandler.UpdateMessage)
		messages.DELETE("/:id", messageHandler.DeleteMessage)
	}

	// Add chat routes (all protected) - NOTE: chatHandler will be nil initially, add when ready
	if chatHandler != nil {
		chats := router.Group("/api/v1/crm/chats")
		chats.Use(authMiddleware.Authenticate())
		chats.Use(rlsMiddleware.SetUserContext())
		{
			chats.POST("", chatHandler.CreateChat)
			chats.GET("", chatHandler.ListChats)
			chats.GET("/:id", chatHandler.GetChat)
			chats.POST("/:id/participants", chatHandler.AddParticipant)
			chats.DELETE("/:id/participants/:participant_id", chatHandler.RemoveParticipant)
			chats.POST("/:id/archive", chatHandler.ArchiveChat)
			chats.POST("/:id/unarchive", chatHandler.UnarchiveChat)
			chats.POST("/:id/close", chatHandler.CloseChat)
			chats.PATCH("/:id/subject", chatHandler.UpdateChatSubject)
		}
	}

	// Add agent routes (all protected)
	if agentHandler != nil {
		agents := router.Group("/api/v1/crm/agents")
		agents.Use(authMiddleware.Authenticate())
		agents.Use(rlsMiddleware.SetUserContext())
		{
			agents.GET("/search", agentHandler.SearchAgents)         // Must be before /:id
			agents.GET("/advanced", agentHandler.ListAgentsAdvanced) // Must be before /:id
			agents.POST("/virtual", agentHandler.CreateVirtualAgent) // Must be before /:id
			agents.GET("", agentHandler.ListAgents)
			agents.POST("", agentHandler.CreateAgent)
			agents.GET("/:id", agentHandler.GetAgent)
			agents.PUT("/:id", agentHandler.UpdateAgent)
			agents.DELETE("/:id", agentHandler.DeleteAgent)
			agents.GET("/:id/stats", agentHandler.GetAgentStats)
			agents.PUT("/:id/virtual/end-period", agentHandler.EndVirtualAgentPeriod)
		}
	}

	// Add note routes (all protected)
	if noteHandler != nil {
		notes := router.Group("/api/v1/crm/notes")
		notes.Use(authMiddleware.Authenticate())
		notes.Use(rlsMiddleware.SetUserContext())
		{
			notes.GET("/search", noteHandler.SearchNotes)         // Must be before /:id
			notes.GET("/advanced", noteHandler.ListNotesAdvanced) // Must be before /:id
		}
	}

	// WebSocket routes (real-time messaging)
	// SECURITY: Autenticação obrigatória via token + rate limiting
	if websocketHandler != nil && wsRateLimiter != nil {
		ws := router.Group("/api/v1/ws")
		ws.Use(wsRateLimiter.RateLimit(5, 1*time.Minute)) // Max 5 conexões por minuto
		ws.Use(wsAuthMiddleware.Authenticate())
		{
			ws.GET("/messages", websocketHandler.HandleWebSocket)
			ws.GET("/stats", websocketHandler.GetStats) // Stats protegidas
		}
	}

	// Add automation discovery routes (all protected)
	if automationDiscoveryHandler != nil {
		automation := router.Group("/api/v1/crm/automation")
		automation.Use(authMiddleware.Authenticate())
		{
			// Discovery endpoints - read-only (no RLS needed)
			automation.GET("/types", automationDiscoveryHandler.GetAutomationTypes)
			automation.GET("/triggers", automationDiscoveryHandler.GetTriggers)
			automation.GET("/triggers/:code", automationDiscoveryHandler.GetTriggerDetails)
			automation.GET("/actions", automationDiscoveryHandler.GetActions)
			automation.GET("/conditions/operators", automationDiscoveryHandler.GetConditionOperators)
			automation.GET("/logic-operators", automationDiscoveryHandler.GetLogicOperators)
			automation.GET("/discovery", automationDiscoveryHandler.GetFullDiscovery)

			// Custom trigger management (admin only - add RBAC middleware if needed)
			automation.POST("/triggers/custom", automationDiscoveryHandler.RegisterCustomTrigger)
			automation.DELETE("/triggers/custom/:code", automationDiscoveryHandler.UnregisterCustomTrigger)
		}
	}

	// Add test routes
	v1 := router.Group("/api/v1/crm")
	{
		// Test routes
		testHandler := handlers.NewTestHandler(gormDB, logger)
		test := v1.Group("/test")
		{
			test.POST("/setup", testHandler.SetupTestEnvironment)
			test.POST("/cleanup", testHandler.CleanupTestEnvironment)
			test.PUT("/pipeline/:id/timeout", testHandler.UpdatePipelineTimeout)
			test.POST("/waha-message", testHandler.TestWAHAMessage)
			test.POST("/send-waha-message", testHandler.SendWAHAMessage)
			test.POST("/waha-connection", testHandler.TestWAHAConnection)
			test.POST("/waha-qr", testHandler.TestWAHAQRCode)
		}
	}
}
