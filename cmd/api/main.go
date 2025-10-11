package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	_ "github.com/caloi/ventros-crm/docs" // Import swagger docs
	"github.com/caloi/ventros-crm/infrastructure/cache"
	"github.com/caloi/ventros-crm/infrastructure/channels/waha"
	"github.com/caloi/ventros-crm/infrastructure/config"
	"github.com/caloi/ventros-crm/infrastructure/database"
	"github.com/caloi/ventros-crm/infrastructure/health"
	"github.com/caloi/ventros-crm/infrastructure/http/handlers"
	"github.com/caloi/ventros-crm/infrastructure/http/middleware"
	"github.com/caloi/ventros-crm/infrastructure/http/routes"
	"github.com/caloi/ventros-crm/infrastructure/messaging"
	"github.com/caloi/ventros-crm/infrastructure/persistence"
	"github.com/caloi/ventros-crm/infrastructure/webhooks"
	ws "github.com/caloi/ventros-crm/infrastructure/websocket"
	"github.com/caloi/ventros-crm/infrastructure/workflow"
	channelapp "github.com/caloi/ventros-crm/internal/application/channel"
	chatapp "github.com/caloi/ventros-crm/internal/application/chat"
	messagecommand "github.com/caloi/ventros-crm/internal/application/commands/message"
	appconfig "github.com/caloi/ventros-crm/internal/application/config"
	contactapp "github.com/caloi/ventros-crm/internal/application/contact"
	contacteventapp "github.com/caloi/ventros-crm/internal/application/contact_event"
	messageapp "github.com/caloi/ventros-crm/internal/application/message"
	pipelineapp "github.com/caloi/ventros-crm/internal/application/pipeline"
	sessionapp "github.com/caloi/ventros-crm/internal/application/session"
	"github.com/caloi/ventros-crm/internal/application/shared"
	trackingapp "github.com/caloi/ventros-crm/internal/application/tracking"
	"github.com/caloi/ventros-crm/internal/application/user"
	webhookapp "github.com/caloi/ventros-crm/internal/application/webhook"
	wsapp "github.com/caloi/ventros-crm/internal/application/websocket"

	// contact_event "github.com/caloi/ventros-crm/internal/domain/contact/events" // Temporariamente comentado
	domainPipeline "github.com/caloi/ventros-crm/internal/domain/crm/pipeline"
	channelworkflow "github.com/caloi/ventros-crm/internal/workflows/channel"
	sagaworkflow "github.com/caloi/ventros-crm/internal/workflows/saga"
	sessionworkflow "github.com/caloi/ventros-crm/internal/workflows/session"
	"github.com/gin-gonic/gin"
	"go.temporal.io/sdk/worker"
	"go.uber.org/zap"
)

//	@title			Ventros CRM API
//	@version		1.0
//	@description	API para gerenciamento de CRM com eventos e workflows
//	@termsOfService	http://swagger.io/terms/

//	@contact.name	API Support
//	@contact.email	support@ventros.com

//	@license.name	MIT
//	@license.url	https://opensource.org/licenses/MIT

//	@BasePath	/
//	@schemes	http https

//	@securityDefinitions.apikey	BearerAuth
//	@in							header
//	@name						Authorization
//	@description				Bearer token authentication. Format: "Bearer {token}"

//	@securityDefinitions.apikey	ApiKeyAuth
//	@in							header
//	@name						X-API-Key
//	@description				API Key authentication for service-to-service communication

// loggerAdapter adapts zap.Logger to pipeline.Logger interface
type loggerAdapter struct {
	logger *zap.Logger
}

func (l *loggerAdapter) Info(msg string, args ...interface{}) {
	l.logger.Sugar().Infof(msg, args...)
}

func (l *loggerAdapter) Error(msg string, args ...interface{}) {
	l.logger.Sugar().Errorf(msg, args...)
}

func (l *loggerAdapter) Debug(msg string, args ...interface{}) {
	l.logger.Sugar().Debugf(msg, args...)
}

// loggingActionExecutor is a minimal MVP executor that only logs actions
type loggingActionExecutor struct {
	logger *zap.Logger
}

func (e *loggingActionExecutor) Execute(ctx context.Context, action domainPipeline.RuleAction, actionCtx pipelineapp.ActionContext) error {
	e.logger.Info("📋 Scheduled automation action (MVP - logging only)",
		zap.String("action_type", string(action.Type)),
		zap.String("rule_id", actionCtx.RuleID.String()),
		zap.String("tenant_id", actionCtx.TenantID),
		zap.Any("params", action.Params),
	)
	return nil
}

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

	// 🗄️ DATABASE MIGRATIONS (golang-migrate with embedded SQL files)
	// Migrations are located in: infrastructure/database/migrations/*.sql
	// Uses golang-migrate library for versioned migrations
	ctx := context.Background()

	// Create migration runner
	migrationRunner, err := database.NewMigrationRunner(sqlDB, logger)
	if err != nil {
		logger.Fatal("Failed to create migration runner", zap.Error(err))
	}
	defer migrationRunner.Close()

	// Apply all pending migrations automatically
	// This is safe to run on every startup (idempotent)
	if err := migrationRunner.Up(); err != nil {
		logger.Fatal("Failed to apply database migrations", zap.Error(err))
	}

	// Log migration status
	status, err := migrationRunner.Status()
	if err != nil {
		logger.Fatal("Failed to get migration status", zap.Error(err))
	}
	logger.Info(status.Message,
		zap.Uint("version", status.Version),
		zap.Bool("dirty", status.Dirty))

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
			zap.String("hint", "Run: make db-seed or check database configuration"))
	}
	logger.Info("✅ App config loaded",
		zap.Int("channel_types", len(appCfg.ChannelTypes)))

	// Initialize repositories (infrastructure layer)
	contactRepo := persistence.NewGormContactRepository(gormDB)
	sessionRepo := persistence.NewGormSessionRepository(gormDB)
	contactEventRepo := persistence.NewGormContactEventRepository(gormDB)
	channelRepo := persistence.NewGormChannelRepository(gormDB)
	chatRepo := persistence.NewGormChatRepository(gormDB)
	pipelineRepo := persistence.NewGormPipelineRepository(gormDB)
	trackingRepo := persistence.NewGormTrackingRepository(gormDB)
	eventLogRepo := persistence.NewDomainEventLogRepository(gormDB, logger)
	outboxRepo := persistence.NewGormOutboxRepository(gormDB)
	agentRepo := persistence.NewGormAgentRepository(gormDB)
	noteRepo := persistence.NewGormNoteRepository(gormDB)
	messageGroupRepo := persistence.NewGormMessageGroupRepository(gormDB)
	automationRepo := persistence.NewGormAutomationRuleRepository(gormDB)
	logger.Info("Repositories initialized")

	// Initialize webhook repository and use case
	webhookRepo := persistence.NewWebhookRepositoryAdapter(gormDB)
	webhookUseCase := webhookapp.NewManageSubscriptionUseCase(webhookRepo, logger)

	// Initialize webhook notifier
	webhookNotifier := webhooks.NewWebhookNotifier(logger, webhookRepo)

	// Initialize event bus with Outbox Pattern (SEM POLLING!)
	eventBus := messaging.NewDomainEventBus(gormDB, outboxRepo, webhookNotifier, eventLogRepo, rabbitConn)
	logger.Info("Event bus initialized with Transactional Outbox Pattern (push-based, no polling!)")

	// Initialize and start PostgreSQL LISTEN/NOTIFY Outbox Processor (push-based, < 100ms latency!)
	// Processa eventos imediatamente após commit via database trigger NOTIFY
	dbConnStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Name,
		cfg.Database.SSLMode,
	)
	postgresNotifyProcessor := messaging.NewPostgresNotifyOutboxProcessor(gormDB, outboxRepo, rabbitConn, logger, dbConnStr)
	if err := postgresNotifyProcessor.Start(ctx); err != nil {
		logger.Fatal("Failed to start PostgreSQL NOTIFY processor (required for push-based event processing)", zap.Error(err))
	}
	logger.Info("✅ PostgreSQL LISTEN/NOTIFY Outbox Processor started (push-based, < 100ms latency, NO POLLING!)")
	// Cleanup on shutdown
	defer postgresNotifyProcessor.Stop()

	// Initialize session manager (Temporal workflows) - will be used by processMessageUseCase

	// Initialize message repository (needed for session enrichment)
	// Keep concrete type for adapter, interface for use cases
	gormMessageRepoImpl := &persistence.GormMessageRepository{}
	gormMessageRepoImpl = persistence.NewGormMessageRepository(gormDB).(*persistence.GormMessageRepository)
	messageRepoAdapter := persistence.NewMessageRepositoryAdapter(gormMessageRepoImpl)

	// Initialize and start session worker
	sessionWorker := workflow.NewSessionWorker(temporalClient, sessionRepo, messageRepoAdapter, eventBus, logger)
	if err := sessionWorker.Start(ctx); err != nil {
		logger.Fatal("Failed to start session worker", zap.Error(err))
	}

	// Schedule session cleanup
	if err := sessionWorker.ScheduleCleanup(ctx); err != nil {
		logger.Warn("Failed to schedule session cleanup", zap.Error(err))
	}

	// Initialize repositories and services needed for saga (BEFORE worker registration)
	messageRepo := persistence.NewGormMessageRepository(gormDB)
	sagaEventBus := messaging.NewSagaEventBusAdapter(eventBus)
	messageEventBus := messaging.NewMessageEventBusAdapter(eventBus)
	timeoutResolver := sessionapp.NewSessionTimeoutResolver(channelRepo, pipelineRepo)
	txManagerShared := shared.NewGormTransactionManager(gormDB)

	// Initialize and start saga worker (Temporal)
	if cfg.UseSagaOrchestration && temporalClient != nil {
		sagaWorker := worker.New(temporalClient, "message-processing", worker.Options{
			MaxConcurrentActivityExecutionSize:    10,
			MaxConcurrentWorkflowTaskExecutionSize: 10,
		})

		// Register workflow
		sagaWorker.RegisterWorkflow(sagaworkflow.ProcessInboundMessageSaga)

		// Create activities instance
		sagaActivities := sagaworkflow.NewActivities(
			contactRepo,
			sessionRepo,
			messageRepo,
			txManagerShared,
			sagaEventBus,
			timeoutResolver,
		)

		// Register forward activities
		sagaWorker.RegisterActivity(sagaActivities.FindOrCreateContactActivity)
		sagaWorker.RegisterActivity(sagaActivities.FindOrCreateSessionActivity)
		sagaWorker.RegisterActivity(sagaActivities.CreateMessageActivity)
		sagaWorker.RegisterActivity(sagaActivities.PublishDomainEventsActivity)
		sagaWorker.RegisterActivity(sagaActivities.ProcessMessageDebouncerActivity)
		sagaWorker.RegisterActivity(sagaActivities.TrackAdConversionActivity)

		// Register compensation activities
		sagaWorker.RegisterActivity(sagaActivities.DeleteContactActivity)
		sagaWorker.RegisterActivity(sagaActivities.CloseSessionActivity)
		sagaWorker.RegisterActivity(sagaActivities.DeleteMessageActivity)

		// Start worker in background
		go func() {
			err := sagaWorker.Run(worker.InterruptCh())
			if err != nil {
				logger.Error("Saga worker stopped", zap.Error(err))
			}
		}()
		logger.Info("✅ Saga Orchestration worker started (Temporal)", zap.String("task_queue", "message-processing"))
	} else {
		logger.Info("Saga Orchestration disabled (using transaction-based processing)")
	}

	// ❌ REMOVIDO: Temporal Outbox Worker (fazia polling a cada 30 segundos)
	// PostgreSQL LISTEN/NOTIFY é suficiente (push-based, <100ms latency, SEM POLLING!)
	// Se precisar de fallback no futuro, considerar aumentar PollInterval para 5-10 minutos
	logger.Info("Outbox processing: Using PostgreSQL LISTEN/NOTIFY only (NO POLLING!)")

	// 🤖 SCHEDULED AUTOMATION WORKER (MVP - Logging Only)
	// Phase 1: Minimal integration with logging-only action executor
	// Future: Wire full action executors (MessageSender, PipelineStatusChanger, etc.)

	// Create adapter instances
	logAdapter := &loggerAdapter{logger: logger}
	loggingExecutor := &loggingActionExecutor{logger: logger}

	// Initialize automation engine with logging executor
	automationEngine := pipelineapp.NewAutomationEngine(automationRepo, loggingExecutor, logAdapter)

	// Initialize and start scheduled rules worker
	scheduledWorker := workflow.NewScheduledRulesWorker(
		gormDB,
		automationEngine,
		1*time.Minute, // Poll every 1 minute
		logAdapter,
	)

	// Start worker in background
	go func() {
		scheduledWorker.Start(ctx)
	}()
	defer scheduledWorker.Stop()

	logger.Info("✅ Scheduled Automation Worker started (polling every 1 minute, MVP mode)")

	// Initialize use cases with event bus adapters (DDD: Application Layer)
	contactEventBus := messaging.NewContactEventBusAdapter(eventBus)
	sessionEventBus := messaging.NewSessionEventBusAdapter(eventBus)
	chatEventBus := messaging.NewChatEventBusAdapter(eventBus)

	createContactUseCase := contactapp.NewCreateContactUseCase(contactRepo, contactEventBus, txManagerShared)
	changePipelineStatusUseCase := contactapp.NewChangePipelineStatusUseCase(contactRepo, pipelineRepo, contactEventBus, txManagerShared)
	createSessionUseCase := sessionapp.NewCreateSessionUseCase(sessionRepo, sessionEventBus, txManagerShared)
	closeSessionUseCase := sessionapp.NewCloseSessionUseCase(sessionRepo, sessionEventBus, txManagerShared)

	// Initialize Chat use cases (DDD: Application Service)
	createChatUseCase := chatapp.NewCreateChatUseCase(chatRepo, chatEventBus)
	findChatUseCase := chatapp.NewFindChatUseCase(chatRepo)
	manageParticipantsUseCase := chatapp.NewManageParticipantsUseCase(chatRepo, chatEventBus)
	archiveChatUseCase := chatapp.NewArchiveChatUseCase(chatRepo, chatEventBus)
	updateChatUseCase := chatapp.NewUpdateChatUseCase(chatRepo, chatEventBus)

	// Initialize Tracking use cases (DDD: Application Service)
	createTrackingUseCase := trackingapp.NewCreateTrackingUseCase(trackingRepo, eventBus, logger, txManagerShared)
	getTrackingUseCase := trackingapp.NewGetTrackingUseCase(trackingRepo, logger)
	getContactTrackingsUseCase := trackingapp.NewGetContactTrackingsUseCase(trackingRepo, logger)

	// Initialize Contact Event use case (DDD: Application Service)
	createContactEventUseCase := contacteventapp.NewCreateContactEventUseCase(contactEventRepo)

	// Initialize idempotency checker for consumers
	idempotencyChecker := persistence.NewIdempotencyChecker(gormDB)

	// Initialize Contact Event Consumer (DDD: Infrastructure -> Application)
	// Consome Domain Events e cria Contact Events para timeline/SSE
	contactEventConsumer := messaging.NewContactEventConsumer(rabbitConn, createContactEventUseCase, idempotencyChecker, logger)

	// Start consuming domain events in background
	go func() {
		if err := contactEventConsumer.Start(ctx); err != nil {
			logger.Error("Failed to start contact event consumer", zap.Error(err))
		}
	}()
	logger.Info("Contact Event Consumer started")

	// Initialize processMessageUseCase
	// (messageRepo, messageEventBus, timeoutResolver, txManagerShared already created above for Saga)

	// Initialize session manager
	sessionManager := sessionworkflow.NewSessionManager(temporalClient)

	// Note: Project and customer IDs are now obtained from authenticated user context
	// via the auth middleware and RLS system

	// Initialize MessageDebouncerService (message grouping with Redis)
	var messageDebouncerSvc *messageapp.MessageDebouncerService
	if redisClient != nil {
		messageDebouncerSvc = messageapp.NewMessageDebouncerService(
			logger,
			messageGroupRepo,
			messageRepo,
			channelRepo,
			redisClient,
		)
		logger.Info("Message debouncer service initialized")
	} else {
		logger.Warn("Redis not available, message debouncer disabled")
	}

	// Initialize ProcessInboundMessageUseCase with Saga support
	processMessageUseCase := messageapp.NewProcessInboundMessageUseCase(
		contactRepo,
		sessionRepo,
		messageRepo,
		contactEventRepo,
		messageEventBus,
		sessionManager,
		timeoutResolver,
		gormDB,              // Usado apenas para invisible tracking detection
		messageDebouncerSvc, // Opcional - passa nil se Redis não disponível
		txManagerShared,
		temporalClient,
		cfg.UseSagaOrchestration,
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

	// Setup nova arquitetura WAHA (eventos raw)
	wahaIntegration := messaging.NewWAHAIntegration(
		rabbitConn,
		wahaMessageService,
		messageRepo,
		channelRepo,
		contactRepo,
		chatRepo,
		logger,
	)

	// Configura as filas da nova arquitetura
	if err := wahaIntegration.SetupQueues(); err != nil {
		logger.Fatal("Failed to setup WAHA integration queues", zap.Error(err))
	}

	// Inicia os processors da nova arquitetura
	if err := wahaIntegration.StartProcessors(ctx, rabbitConn); err != nil {
		logger.Fatal("Failed to start WAHA integration processors", zap.Error(err))
	}

	// Initialize WAHA consumer (legado - manter para compatibilidade)
	wahaConsumer := messaging.NewWAHAMessageConsumer(wahaMessageService, idempotencyChecker)
	if err := wahaConsumer.Start(ctx, rabbitConn); err != nil {
		logger.Fatal("Failed to start WAHA consumer", zap.Error(err))
	}
	logger.Info("Use cases and WAHA service initialized successfully")

	// TODO: Initialize MessageGroupWorker when enrichment services are ready
	//
	// O MessageGroupWorker processa grupos de mensagens expirados:
	// 1. Busca grupos expirados no banco (via MessageGroupRepo)
	// 2. Processa enriquecimentos (transcrição, OCR, etc) via MessageEnrichmentService
	// 3. Aguarda enriquecimentos completarem
	// 4. Concatena todas as mensagens do grupo e envia para AI Agent
	//
	// Dependências necessárias:
	// - MessageEnrichmentService (transcrição de áudio, OCR de imagem, etc)
	// - AIAgentService (envia mensagens concatenadas para AI)
	//
	// Exemplo de inicialização (quando serviços estiverem prontos):
	/*
		enrichmentService := messageapp.NewMessageEnrichmentService(...)
		aiAgentService := messageapp.NewAIAgentService(...)

		messageGroupWorker := messageapp.NewMessageGroupWorker(
			logger,
			messageDebouncerSvc,
			enrichmentService,
			aiAgentService,
			messageGroupRepo,
		)

		go func() {
			if err := messageGroupWorker.Start(ctx); err != nil {
				logger.Error("Message group worker stopped", zap.Error(err))
			}
		}()
		logger.Info("✅ Message group worker started")
	*/

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

	// Initialize WAHA client for channel service
	wahaClient := waha.NewWAHAClientFromEnv(logger)

	// Initialize WAHA history importer
	historyImporter := channelapp.NewWAHAHistoryImporter(
		logger,
		wahaClient,
		channelRepo,
		contactRepo,
		sessionRepo,
		messageRepo,
	)

	// Initialize channel service with history importer
	channelService := channelapp.NewChannelService(channelRepo, logger, wahaClient, historyImporter)

	// Initialize and start WAHA import worker (Temporal)
	wahaImportWorker := channelworkflow.NewWAHAImportWorker(
		temporalClient,
		wahaClient,
		channelRepo,
		contactRepo,
		sessionRepo,
		messageRepo,
		logger,
	)
	if err := wahaImportWorker.Start(ctx); err != nil {
		logger.Fatal("Failed to start WAHA import worker", zap.Error(err))
	}
	logger.Info("WAHA import worker started successfully")

	// Create adapter for WAHA message sender
	wahaMessageSender := persistence.NewWAHAMessageSenderAdapter(wahaClient, logger)

	// Create adapter for session repository (adds GetActiveSessionByContact)
	sessionRepoAdapter := persistence.NewSessionRepositoryAdapter(sessionRepo)

	// Initialize message sending (CQRS Command)
	sendMessageHandler := messagecommand.NewSendMessageHandler(
		contactRepo,
		sessionRepoAdapter,
		messageRepo,
		wahaMessageSender,
		txManagerShared,
	)

	// Initialize message delivery confirmation (CQRS Command)
	confirmMessageDeliveryHandler := messagecommand.NewConfirmMessageDeliveryHandler(messageRepo)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(logger, userService)
	channelHandler := handlers.NewChannelHandler(logger, channelService, temporalClient)
	wahaHandler := handlers.NewWAHAWebhookHandler(logger, wahaIntegration.RawEventBus, channelRepo)
	webhookHandler := handlers.NewWebhookSubscriptionHandler(logger, webhookUseCase)
	queueHandler := handlers.NewQueueHandler(logger, rabbitConn)
	sessionHandler := handlers.NewSessionHandler(logger, sessionRepo)
	contactHandler := handlers.NewContactHandler(logger, contactRepo, changePipelineStatusUseCase)
	chatHandler := handlers.NewChatHandler(logger, createChatUseCase, findChatUseCase, manageParticipantsUseCase, archiveChatUseCase, updateChatUseCase)
	messageHandler := handlers.NewMessageHandler(logger, messageRepo, sendMessageHandler, confirmMessageDeliveryHandler)
	trackingHandler := handlers.NewTrackingHandler(createTrackingUseCase, getTrackingUseCase, getContactTrackingsUseCase, logger)
	agentHandler := handlers.NewAgentHandler(logger, agentRepo)
	noteHandler := handlers.NewNoteHandler(logger, noteRepo)
	domainEventHandler := handlers.NewDomainEventHandler(eventLogRepo, logger)

	// Create auth middleware
	authMiddleware := middleware.NewAuthMiddleware(logger, cfg.Server.Env != "production", userService)

	// Create RLS middleware (agora só precisa do logger)
	rlsMiddleware := middleware.NewRLSMiddleware(logger)

	// TODO: Update handlers to use use cases instead of repositories directly
	_ = createContactUseCase
	_ = createSessionUseCase
	_ = closeSessionUseCase
	_ = domainEventHandler // TODO: Add domain event routes

	// Initialize pipeline handler
	pipelineHandler := handlers.NewPipelineHandler(logger, pipelineRepo)

	// Initialize project handler (placeholder - using mock for now)
	projectRepo := persistence.NewMockProjectRepository()
	projectHandler := handlers.NewProjectHandler(logger, projectRepo)

	// Initialize TriggerRegistry for automation discovery
	triggerRegistry := domainPipeline.NewTriggerRegistry()

	// Initialize automation discovery handler
	automationDiscoveryHandler := handlers.NewAutomationDiscoveryHandler(triggerRegistry)

	// Initialize automation handler (cross-product AUTOMATION product)
	automationHandler := handlers.NewAutomationHandler(logger, gormDB)

	// Initialize broadcast handler (AUTOMATION product)
	broadcastHandler := handlers.NewBroadcastHandler(logger, gormDB)

	// Initialize sequence handler (AUTOMATION product)
	sequenceHandler := handlers.NewSequenceHandler(logger, gormDB)

	// Initialize campaign handler (AUTOMATION product)
	campaignHandler := handlers.NewCampaignHandler(logger, gormDB)

	// Initialize WebSocket infrastructure
	// WebSocket message handler (integra com domain Message)
	wsMessageHandler := wsapp.NewWebSocketMessageHandler(messageRepo, logger)

	// WebSocket Hub (Redis Pub/Sub para multi-server)
	wsHub := ws.NewHub(redisClient, wsMessageHandler, logger)

	// Start Hub em goroutine (event loop)
	go wsHub.Run()
	logger.Info("✅ WebSocket Hub started (Redis Pub/Sub enabled)")

	// Cleanup on shutdown
	defer func() {
		if err := wsHub.Shutdown(); err != nil {
			logger.Error("Failed to shutdown WebSocket hub", zap.Error(err))
		}
	}()

	// WebSocket HTTP handler
	isProduction := cfg.Server.Env == "production"
	websocketHandler := handlers.NewWebSocketMessageHandler(wsHub, isProduction, logger)

	// WebSocket auth middleware
	wsAuthMiddleware := middleware.NewWebSocketAuthMiddleware(authMiddleware, logger)

	// WebSocket rate limiter (max 5 connections per minute per IP)
	wsRateLimiter := middleware.NewWebSocketRateLimiter(redisClient, logger)

	// Initialize HTTP Rate Limiter (global rate limiting for API endpoints)
	rateLimiter := middleware.NewRateLimiter(redisClient, logger)

	// Set Gin mode
	if cfg.Server.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create Gin router
	router := gin.Default()

	// Setup basic routes (health, queue, session, contact, webhooks, auth, automation, broadcasts, sequences, campaigns, channels, projects, pipelines, trackings, automation discovery, messages, chats, agents, notes, WebSocket)
	routes.SetupRoutesBasicWithTest(router, logger, healthChecker, authHandler, automationHandler, broadcastHandler, sequenceHandler, campaignHandler, channelHandler, projectHandler, pipelineHandler, wahaHandler, webhookHandler, queueHandler, sessionHandler, contactHandler, trackingHandler, messageHandler, chatHandler, agentHandler, noteHandler, automationDiscoveryHandler, websocketHandler, wsRateLimiter, gormDB, authMiddleware, wsAuthMiddleware, rlsMiddleware, rateLimiter)

	// Start server with graceful shutdown
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.Server.Port),
		Handler: router,
	}

	// Start server in goroutine
	go func() {
		logger.Info("🚀 Server ready to accept connections", zap.String("port", cfg.Server.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with a timeout
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("🛑 Shutting down server...")

	// Create shutdown context with 5s timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("✅ Server exited gracefully")
}

func initLogger(level, env string) (*zap.Logger, error) {
	if env == "production" {
		return zap.NewProduction()
	}
	return zap.NewDevelopment()
}
