package main

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	_ "github.com/caloi/ventros-crm/docs" // Import swagger docs
	"github.com/caloi/ventros-crm/infrastructure/cache"
	"github.com/caloi/ventros-crm/infrastructure/channels/waha"
	"github.com/caloi/ventros-crm/infrastructure/config"
	"github.com/caloi/ventros-crm/infrastructure/health"
	"github.com/caloi/ventros-crm/infrastructure/http/handlers"
	"github.com/caloi/ventros-crm/infrastructure/http/middleware"
	"github.com/caloi/ventros-crm/infrastructure/http/routes"
	"github.com/caloi/ventros-crm/infrastructure/messaging"
	"github.com/caloi/ventros-crm/infrastructure/persistence"
	"github.com/caloi/ventros-crm/infrastructure/webhooks"
	"github.com/caloi/ventros-crm/infrastructure/workflow"
	appconfig "github.com/caloi/ventros-crm/internal/application/config"
	contactapp "github.com/caloi/ventros-crm/internal/application/contact"
	messageapp "github.com/caloi/ventros-crm/internal/application/message"
	sessionapp "github.com/caloi/ventros-crm/internal/application/session"
	"github.com/caloi/ventros-crm/internal/application/user"
	webhookapp "github.com/caloi/ventros-crm/internal/application/webhook"
	channelapp "github.com/caloi/ventros-crm/internal/application/channel"
	// contact_event "github.com/caloi/ventros-crm/internal/domain/contact/events" // Temporariamente comentado
	sessionworkflow "github.com/caloi/ventros-crm/internal/workflows/session"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// @title Ventros CRM API
// @version 1.0
// @description API para gerenciamento de CRM com eventos e workflows
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.email support@ventros.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host
// @BasePath /
// @schemes http https

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize logger
	logger, err := initLogger(cfg.Log.Level, cfg.Server.Env)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	logger.Info("Starting Ventros CRM API",
		zap.String("port", cfg.Server.Port),
		zap.String("env", cfg.Server.Env),
	)

	// Initialize GORM database
	dbPort, err := strconv.Atoi(cfg.Database.Port)
	if err != nil {
		logger.Fatal("Invalid database port", zap.Error(err))
	}
	gormDB, err := persistence.NewDatabase(persistence.DatabaseConfig{
		Host:     cfg.Database.Host,
		Port:     dbPort,
		User:     cfg.Database.User,
		Password: cfg.Database.Password,
		DBName:   cfg.Database.Name,
		SSLMode:  cfg.Database.SSLMode,
	})
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	
	// Get underlying SQL DB for connection management
	sqlDB, err := gormDB.DB()
	if err != nil {
		logger.Fatal("Failed to get SQL DB", zap.Error(err))
	}
	defer sqlDB.Close()
	logger.Info("Database connected successfully")

	// Run GORM migrations
	ctx := context.Background()
	if err := persistence.AutoMigrate(gormDB); err != nil {
		logger.Fatal("Failed to run database migrations", zap.Error(err))
	}
	logger.Info("✅ Database migrations completed")

	// Setup Row Level Security (RLS)
	if err := persistence.SetupRLS(gormDB); err != nil {
		logger.Warn("Failed to setup RLS, continuing without it", zap.Error(err))
	}

	// Register RLS callbacks for GORM
	if err := persistence.RegisterRLSCallbacks(gormDB); err != nil {
		logger.Fatal("Failed to register RLS callbacks", zap.Error(err))
	}
	logger.Info("✅ RLS callbacks registered successfully")

	// Initialize Redis
	redisClient, err := cache.NewRedisClient(cache.Config{
		Host:     cfg.Redis.Host,
		Port:     cfg.Redis.Port,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	if err != nil {
		logger.Warn("Failed to connect to Redis, continuing without cache", zap.Error(err))
		redisClient = nil
	} else {
		defer redisClient.Close()
		logger.Info("Redis connected successfully")
	}

	// Initialize Temporal client
	temporalClient, err := workflow.NewTemporalClient(workflow.Config{
		Host:      cfg.Temporal.Host,
		Namespace: cfg.Temporal.Namespace,
	})
	if err != nil {
		logger.Warn("Failed to connect to Temporal, continuing without workflow engine", zap.Error(err))
		temporalClient = nil
	} else {
		defer temporalClient.Close()
		logger.Info("Temporal connected successfully")
	}

	// Initialize RabbitMQ
	rabbitConn, err := messaging.NewRabbitMQConnection(messaging.RabbitMQConfig{
		URL:            cfg.RabbitMQ.URL,
		ReconnectDelay: 5 * time.Second,
		MaxReconnects:  10,
	})
	if err != nil {
		logger.Fatal("Failed to connect to RabbitMQ", zap.Error(err))
	}
	defer rabbitConn.Close()
	logger.Info("RabbitMQ connected successfully")

	// Setup RabbitMQ queues
	logger.Info("Setting up RabbitMQ queues...")
	if err := rabbitConn.SetupAllQueues(); err != nil {
		logger.Fatal("Failed to setup RabbitMQ queues", zap.Error(err))
	}
	logger.Info("✅ All RabbitMQ queues declared successfully")

	// Load application config from database
	configService := appconfig.NewAppConfigService(gormDB)
	appCfg, err := configService.LoadConfig(ctx)
	if err != nil {
		logger.Fatal("Failed to load app config from database", 
			zap.Error(err),
			zap.String("hint", "Run: make seed or execute SQL seeds in deployments/docker/seeds/"))
	}
	logger.Info("✅ App config loaded",
		zap.Int("channel_types", len(appCfg.ChannelTypes)))
	
	// Initialize repositories (infrastructure layer)
	contactRepo := persistence.NewGormContactRepository(gormDB)
	sessionRepo := persistence.NewGormSessionRepository(gormDB)
	contactEventRepo := persistence.NewGormContactEventRepository(gormDB)
	channelRepo := persistence.NewGormChannelRepository(gormDB)
	logger.Info("Repositories initialized")

	// Initialize webhook repository and use case
	webhookRepo := persistence.NewWebhookRepositoryAdapter(gormDB)
	webhookUseCase := webhookapp.NewManageSubscriptionUseCase(webhookRepo, logger)
	
	// Initialize webhook notifier
	webhookNotifier := webhooks.NewWebhookNotifier(logger, webhookRepo)

	// Initialize event bus with webhook notifier
	eventBus := messaging.NewDomainEventBus(rabbitConn, webhookNotifier)
	logger.Info("Event bus initialized")

	// Initialize session manager (Temporal workflows) - will be used by processMessageUseCase
	
	// Initialize and start session worker
	sessionWorker := workflow.NewSessionWorker(temporalClient, sessionRepo, eventBus, logger)
	if err := sessionWorker.Start(ctx); err != nil {
		logger.Fatal("Failed to start session worker", zap.Error(err))
	}
	
	// Schedule session cleanup
	if err := sessionWorker.ScheduleCleanup(ctx); err != nil {
		logger.Warn("Failed to schedule session cleanup", zap.Error(err))
	}
	
	// Initialize use cases with event bus adapters
	contactEventBus := messaging.NewContactEventBusAdapter(eventBus)
	sessionEventBus := messaging.NewSessionEventBusAdapter(eventBus)
	
	createContactUseCase := contactapp.NewCreateContactUseCase(contactRepo, contactEventBus)
	createSessionUseCase := sessionapp.NewCreateSessionUseCase(sessionRepo, sessionEventBus)
	closeSessionUseCase := sessionapp.NewCloseSessionUseCase(sessionRepo, sessionEventBus)
	
	// Initialize processMessageUseCase
	messageRepo := persistence.NewGormMessageRepository(gormDB)
	
	// Create message event bus adapter
	messageEventBus := messaging.NewMessageEventBusAdapter(eventBus)
	
	// Initialize session manager
	sessionManager := sessionworkflow.NewSessionManager(temporalClient)
	
	// Note: Project and customer IDs are now obtained from authenticated user context
	// via the auth middleware and RLS system
	
	// Initialize ProcessInboundMessageUseCase
	processMessageUseCase := messageapp.NewProcessInboundMessageUseCase(
		contactRepo,
		sessionRepo, 
		messageRepo,
		contactEventRepo,
		messageEventBus,
		sessionManager,
	)
	
	// Load AppConfig (channel types, etc)
	appConfigService := appconfig.NewAppConfigService(gormDB)
	appCfg, err3 := appConfigService.LoadConfig(ctx)
	if err3 != nil {
		logger.Fatal("Failed to load app config", zap.Error(err3))
	}
	
	// Initialize message adapter and WAHA service
	messageAdapter := waha.NewMessageAdapter()
	wahaMessageService := messageapp.NewWAHAMessageService(
		logger,
		channelRepo,
		processMessageUseCase,
		appCfg,
		messageAdapter,
	)
	
	// Initialize WAHA consumer
	wahaConsumer := messaging.NewWAHAMessageConsumer(wahaMessageService)
	if err := wahaConsumer.Start(ctx, rabbitConn); err != nil {
		logger.Fatal("Failed to start WAHA consumer", zap.Error(err))
	}
	logger.Info("Use cases and WAHA service initialized successfully")

	// Initialize health checker
	healthChecker := health.NewHealthChecker(
		sqlDB,
		redisClient,
		cfg.RabbitMQ.URL,
		temporalClient,
		gormDB, // Pass GORM DB for migration checks
	)
	
	// Initialize user service for auth
	userService := user.NewUserService(gormDB)
	
	// Initialize channel service
	channelService := channelapp.NewChannelService(channelRepo, logger)
	
	// Initialize handlers
	authHandler := handlers.NewAuthHandler(logger, userService)
	channelHandler := handlers.NewChannelHandler(logger, channelService)
	wahaHandler := handlers.NewWAHAWebhookHandler(logger, wahaMessageService)
	webhookHandler := handlers.NewWebhookSubscriptionHandler(logger, webhookUseCase)
	queueHandler := handlers.NewQueueHandler(logger, rabbitConn)
	sessionHandler := handlers.NewSessionHandler(logger, sessionRepo)
	contactHandler := handlers.NewContactHandler(logger, contactRepo)
	
	// Create auth middleware
	authMiddleware := middleware.NewAuthMiddleware(logger, cfg.Server.Env != "production", userService)
	
	// Create RLS middleware (agora só precisa do logger)
	rlsMiddleware := middleware.NewRLSMiddleware(logger)
	
	// TODO: Update handlers to use use cases instead of repositories directly
	_ = createContactUseCase
	_ = createSessionUseCase
	_ = closeSessionUseCase

	// Initialize pipeline handler (placeholder - will need pipeline repo)
	// pipelineHandler := handlers.NewPipelineHandler(logger, pipelineRepo)

	// Initialize project handler (placeholder - using mock for now)
	projectRepo := persistence.NewMockProjectRepository()
	projectHandler := handlers.NewProjectHandler(logger, projectRepo)

	// Set Gin mode
	if cfg.Server.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create Gin router
	router := gin.Default()

	// Setup basic routes (health, queue, session, contact, webhooks, auth, channels, projects)
	routes.SetupRoutesBasicWithTest(router, logger, healthChecker, authHandler, channelHandler, projectHandler, wahaHandler, webhookHandler, queueHandler, sessionHandler, contactHandler, gormDB, authMiddleware, rlsMiddleware)

	// Start server
	logger.Info(" Server ready to accept connections", zap.String("port", cfg.Server.Port))
	if err := router.Run(fmt.Sprintf(":%s", cfg.Server.Port)); err != nil {
		logger.Fatal("Failed to start server", zap.Error(err))
	}
}

func initLogger(level, env string) (*zap.Logger, error) {
	if env == "production" {
		return zap.NewProduction()
	}
	return zap.NewDevelopment()
}

