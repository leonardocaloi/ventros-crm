package routes

import (
	"github.com/caloi/ventros-crm/infrastructure/health"
	"github.com/caloi/ventros-crm/infrastructure/http/handlers"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// SetupRoutesMinimal configura apenas as rotas básicas (health, queue, session, contact)
func SetupRoutesMinimal(router *gin.Engine, logger *zap.Logger, healthChecker *health.HealthChecker, queueHandler *handlers.QueueHandler, sessionHandler *handlers.SessionHandler, contactHandler *handlers.ContactHandler) {
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
		
		// Contact management
		contacts := v1.Group("/contacts")
		{
			contacts.GET("/:id", contactHandler.GetContact)
			contacts.GET("/", contactHandler.ListContacts)
		}
	}
}

func SetupRoutes(router *gin.Engine, logger *zap.Logger, healthChecker *health.HealthChecker, wahaHandler *handlers.WAHAWebhookHandler, webhookHandler *handlers.WebhookSubscriptionHandler, queueHandler *handlers.QueueHandler, sessionHandler *handlers.SessionHandler, contactHandler *handlers.ContactHandler, pipelineHandler *handlers.PipelineHandler, projectHandler *handlers.ProjectHandler, agentHandler *handlers.AgentHandler, messageHandler *handlers.MessageHandler, customerHandler *handlers.CustomerHandler) {
	// Middlewares
	router.Use(gin.Recovery())
	router.Use(LoggerMiddleware(logger))
	router.Use(CORSMiddleware())

	// Health check handlers
	healthHandler := handlers.NewHealthHandler(logger, healthChecker)
	
	// Webhook routes (sem auth para receber da WAHA)
	webhooks := router.Group("/webhooks")
	{
		waha := webhooks.Group("/waha")
		{
			waha.POST("/message", wahaHandler.HandleMessage)
			waha.POST("/status", wahaHandler.HandleStatus)
			waha.GET("/health", wahaHandler.Health)
		}
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
		// projects := v1.Group("/projects")
		// {
		// }

		// Contact routes
		contacts := v1.Group("/contacts")
		{
			contacts.GET("", contactHandler.ListContacts)
			contacts.POST("", contactHandler.CreateContact)
			contacts.GET("/:id", contactHandler.GetContact)
			contacts.PUT("/:id", contactHandler.UpdateContact)
			contacts.DELETE("/:id", contactHandler.DeleteContact)
		}

		// Session routes
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
func SetupRoutesBasic(router *gin.Engine, logger *zap.Logger, healthChecker *health.HealthChecker, wahaHandler *handlers.WAHAWebhookHandler, webhookHandler *handlers.WebhookSubscriptionHandler, queueHandler *handlers.QueueHandler, sessionHandler *handlers.SessionHandler, contactHandler *handlers.ContactHandler) {
	// Middlewares
	router.Use(gin.Recovery())
	router.Use(LoggerMiddleware(logger))
	router.Use(CORSMiddleware())

	// Health check handlers
	healthHandler := handlers.NewHealthHandler(logger, healthChecker)
	
	// Webhook routes (sem auth para receber da WAHA)
	webhooks := router.Group("/webhooks")
	{
		waha := webhooks.Group("/waha")
		{
			waha.POST("/message", wahaHandler.HandleMessage)
			waha.POST("/status", wahaHandler.HandleStatus)
			waha.GET("/health", wahaHandler.Health)
		}
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
		// projects := v1.Group("/projects")
		// {
		// }

		// Contact routes
		contacts := v1.Group("/contacts")
		{
			contacts.GET("", contactHandler.ListContacts)
			contacts.POST("", contactHandler.CreateContact)
			contacts.GET("/:id", contactHandler.GetContact)
			contacts.PUT("/:id", contactHandler.UpdateContact)
			contacts.DELETE("/:id", contactHandler.DeleteContact)
		}

		// Session routes
		sessions := v1.Group("/sessions")
		{
			sessions.GET("", sessionHandler.ListSessions)
			sessions.GET("/:id", sessionHandler.GetSession)
			sessions.GET("/stats", sessionHandler.GetSessionStats)
		}

		// Message routes
		// messages := v1.Group("/messages")
		// {
		// }
	}
}

// SetupRoutesBasicWithTest configura as rotas básicas com endpoints de teste
func SetupRoutesBasicWithTest(router *gin.Engine, logger *zap.Logger, healthChecker *health.HealthChecker, wahaHandler *handlers.WAHAWebhookHandler, webhookHandler *handlers.WebhookSubscriptionHandler, queueHandler *handlers.QueueHandler, sessionHandler *handlers.SessionHandler, contactHandler *handlers.ContactHandler, gormDB *gorm.DB) {
	// Use the basic setup first
	SetupRoutesBasic(router, logger, healthChecker, wahaHandler, webhookHandler, queueHandler, sessionHandler, contactHandler)
	
	// Add test routes
	v1 := router.Group("/api/v1")
	{
		// Test routes
		testHandler := handlers.NewTestHandler(gormDB, logger)
		test := v1.Group("/test")
		{
			test.POST("/setup", testHandler.SetupTestEnvironment)
			test.POST("/cleanup", testHandler.CleanupTestEnvironment)
			test.POST("/waha-message", testHandler.TestWAHAMessage)
			test.POST("/send-waha-message", testHandler.SendWAHAMessage)
		}
	}
}
