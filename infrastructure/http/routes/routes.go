package routes

import (
	"github.com/caloi/ventros-crm/infrastructure/health"
	"github.com/caloi/ventros-crm/infrastructure/http/handlers"
	"github.com/caloi/ventros-crm/infrastructure/http/middleware"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
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

	// API v1 routes (minimal)
	v1 := router.Group("/api/v1")
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

func SetupRoutes(router *gin.Engine, logger *zap.Logger, healthChecker *health.HealthChecker, wahaHandler *handlers.WAHAWebhookHandler, webhookHandler *handlers.WebhookSubscriptionHandler, queueHandler *handlers.QueueHandler, sessionHandler *handlers.SessionHandler, contactHandler *handlers.ContactHandler, pipelineHandler *handlers.PipelineHandler, projectHandler *handlers.ProjectHandler, agentHandler *handlers.AgentHandler, messageHandler *handlers.MessageHandler, customerHandler *handlers.CustomerHandler, channelHandler *handlers.ChannelHandler) {
	// Middlewares
	router.Use(gin.Recovery())
	router.Use(LoggerMiddleware(logger))
	router.Use(CORSMiddleware())

	// Health check handlers
	healthHandler := handlers.NewHealthHandler(logger, healthChecker)
	
	// Webhook routes (sem auth para receber da WAHA)
	webhooks := router.Group("/api/v1/webhooks")
	{
		webhooks.POST("/waha/:session", wahaHandler.ReceiveWebhook)
		webhooks.GET("/waha", wahaHandler.GetWebhookInfo)
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

		// Customer routes
		// customers := v1.Group("/customers")
		// {
		// }

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
			contacts.GET("", contactHandler.ListContacts)
			contacts.POST("", contactHandler.CreateContact)
			contacts.GET("/:id", contactHandler.GetContact)
			contacts.PUT("/:id", contactHandler.UpdateContact)
			contacts.DELETE("/:id", contactHandler.DeleteContact)
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
	
	// Webhook routes (sem auth para receber da WAHA)
	webhooks := router.Group("/api/v1/webhooks")
	{
		webhooks.POST("/waha/:session", wahaHandler.ReceiveWebhook)
		webhooks.GET("/waha", wahaHandler.GetWebhookInfo)
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

		// Customer routes
		// customers := v1.Group("/customers")
		// {
		// }

		// Project routes
		// projects := v1.Group("/projects")
		// {
		// }

		// Contact routes
		contacts := v1.Group("/contacts")
		contacts.Use(authMiddleware.Authenticate())
		contacts.Use(rlsMiddleware.SetUserContext())
		{
			contacts.GET("", contactHandler.ListContacts)
			contacts.POST("", contactHandler.CreateContact)
			contacts.GET("/:id", contactHandler.GetContact)
			contacts.PUT("/:id", contactHandler.UpdateContact)
			contacts.DELETE("/:id", contactHandler.DeleteContact)
			
			// Nested session routes under contact (using :id for contact)
			contacts.GET("/:id/sessions", sessionHandler.ListSessions)
			contacts.GET("/:id/sessions/:session_id", sessionHandler.GetSession)
		}

		// Session routes (protected) - global with required filters
		sessions := v1.Group("/sessions")
		sessions.Use(authMiddleware.Authenticate())
		sessions.Use(rlsMiddleware.SetUserContext())
		{
			sessions.GET("", sessionHandler.ListSessions)        // Requires ?contact_id or ?channel_id
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

// SetupRoutesBasicWithTest configura as rotas básicas com endpoints de teste, auth, channels e projects
func SetupRoutesBasicWithTest(router *gin.Engine, logger *zap.Logger, healthChecker *health.HealthChecker, authHandler *handlers.AuthHandler, channelHandler *handlers.ChannelHandler, projectHandler *handlers.ProjectHandler, wahaHandler *handlers.WAHAWebhookHandler, webhookHandler *handlers.WebhookSubscriptionHandler, queueHandler *handlers.QueueHandler, sessionHandler *handlers.SessionHandler, contactHandler *handlers.ContactHandler, gormDB *gorm.DB, authMiddleware *middleware.AuthMiddleware, rlsMiddleware *middleware.RLSMiddleware) {
	// Add GORM context middleware FIRST (before any other middleware)
	router.Use(middleware.GORMContextMiddleware(gormDB))
	
	// Use the basic setup first
	SetupRoutesBasic(router, logger, healthChecker, wahaHandler, webhookHandler, queueHandler, sessionHandler, contactHandler, authMiddleware, rlsMiddleware)
	
	// Add auth routes (public routes)
	authRoutes := router.Group("/api/v1/auth")
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
	
	// Add channel routes (all protected)
	channels := router.Group("/api/v1/channels")
	channels.Use(authMiddleware.Authenticate())
	channels.Use(rlsMiddleware.SetUserContext())
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
		
		// Nested session routes under channel (using :id for channel)
		channels.GET("/:id/sessions", sessionHandler.ListSessions)
		channels.GET("/:id/sessions/:session_id", sessionHandler.GetSession)
	}
	
	// Add project routes (all protected)
	projects := router.Group("/api/v1/projects")
	projects.Use(authMiddleware.Authenticate())
	projects.Use(rlsMiddleware.SetUserContext())
	{
		projects.GET("", projectHandler.ListProjects)
		projects.POST("", projectHandler.CreateProject)
		projects.GET("/:id", projectHandler.GetProject)
		projects.PUT("/:id", projectHandler.UpdateProject)
		projects.DELETE("/:id", projectHandler.DeleteProject)
	}
	
	// Add test routes
	v1 := router.Group("/api/v1")
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
