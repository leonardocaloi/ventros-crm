package testing

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/testsuite"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/caloi/ventros-crm/infrastructure/persistence/entities"
)

// IntegrationTestSuite holds all external dependencies for integration tests
type IntegrationTestSuite struct {
	Ctx               context.Context
	DB                *gorm.DB
	RedisClient       *redis.Client
	RabbitMQConn      *amqp.Connection
	RabbitMQChannel   *amqp.Channel
	TemporalClient    client.Client
	PostgresContainer *postgres.PostgresContainer

	// Test data
	TenantID  string
	ProjectID uuid.UUID
	UserID    uuid.UUID
}

// SetupIntegrationTest initializes all services for integration testing
func SetupIntegrationTest(t *testing.T) *IntegrationTestSuite {
	t.Helper()

	ctx := context.Background()
	suite := &IntegrationTestSuite{
		Ctx:      ctx,
		TenantID: "test-tenant-" + uuid.New().String()[:8],
	}

	// Setup PostgreSQL
	suite.setupPostgres(t)

	// Setup Redis
	suite.setupRedis(t)

	// Setup RabbitMQ
	suite.setupRabbitMQ(t)

	// Setup Temporal (using test suite for unit-style tests)
	suite.setupTemporal(t)

	// Seed initial data
	suite.seedTestData(t)

	return suite
}

// setupPostgres starts a PostgreSQL container
func (s *IntegrationTestSuite) setupPostgres(t *testing.T) {
	t.Helper()

	postgresContainer, err := postgres.Run(s.Ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("ventros_test"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second)),
	)
	require.NoError(t, err)
	s.PostgresContainer = postgresContainer

	dsn, err := postgresContainer.ConnectionString(s.Ctx, "sslmode=disable")
	require.NoError(t, err)

	db, err := gorm.Open(gormpostgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	s.DB = db

	// Run migrations
	err = s.DB.AutoMigrate(
		&entities.UserEntity{},
		&entities.BillingAccountEntity{},
		&entities.ProjectEntity{},
		&entities.PipelineEntity{},
		&entities.PipelineStatusEntity{},
		&entities.ContactEntity{},
		&entities.SessionEntity{},
		&entities.MessageEntity{},
		&entities.ChannelEntity{},
		&entities.ChannelTypeEntity{},
		&entities.AgentEntity{},
		&entities.WebhookSubscriptionEntity{},
		&entities.DomainEventLogEntity{},
		&entities.ContactEventEntity{},
		&entities.TrackingEntity{},
		&entities.TrackingEnrichmentEntity{},
		&entities.OutboxEventEntity{},
		&entities.ProcessedEventEntity{},
		&entities.AutomationEntity{},
		&entities.NoteEntity{},
		&entities.ContactListEntity{},
	)
	require.NoError(t, err)
}

// setupRedis connects to Redis (assumes running on localhost:6379 or via docker-compose.test.yml)
func (s *IntegrationTestSuite) setupRedis(t *testing.T) {
	t.Helper()

	s.RedisClient = redis.NewClient(&redis.Options{
		Addr:     "localhost:6380",
		Password: "",
		DB:       0,
	})

	// Ping to ensure connection
	_, err := s.RedisClient.Ping(s.Ctx).Result()
	if err != nil {
		t.Logf("Redis not available on localhost:6380, skipping Redis tests. Error: %v", err)
		s.RedisClient = nil
	}
}

// setupRabbitMQ connects to RabbitMQ
func (s *IntegrationTestSuite) setupRabbitMQ(t *testing.T) {
	t.Helper()

	conn, err := amqp.Dial("amqp://test:test@localhost:5673/")
	if err != nil {
		t.Logf("RabbitMQ not available on localhost:5673, skipping RabbitMQ tests. Error: %v", err)
		return
	}
	s.RabbitMQConn = conn

	ch, err := conn.Channel()
	require.NoError(t, err)
	s.RabbitMQChannel = ch
}

// setupTemporal initializes Temporal test suite
func (s *IntegrationTestSuite) setupTemporal(t *testing.T) {
	t.Helper()

	// Try to connect to real Temporal service for integration tests
	c, err := client.Dial(client.Options{
		HostPort:  "localhost:7234",
		Namespace: "default",
	})
	if err != nil {
		t.Logf("Temporal not available on localhost:7234, will use test suite. Error: %v", err)
		return
	}
	s.TemporalClient = c
}

// seedTestData creates basic test data
func (s *IntegrationTestSuite) seedTestData(t *testing.T) {
	t.Helper()

	// Create user
	user := &entities.UserEntity{
		ID:           uuid.New(),
		Email:        "test@example.com",
		Name:         "Test User",
		PasswordHash: "hashed_password",
		Status:       "active",
		Role:         "admin",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	require.NoError(t, s.DB.Create(user).Error)
	s.UserID = user.ID

	// Create billing account
	billingAccount := &entities.BillingAccountEntity{
		ID:            uuid.New(),
		UserID:        user.ID,
		Name:          "Test Billing Account",
		PaymentStatus: "active",
		BillingEmail:  "billing@example.com",
		Suspended:     false,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	require.NoError(t, s.DB.Create(billingAccount).Error)

	// Create project
	project := &entities.ProjectEntity{
		ID:               uuid.New(),
		UserID:           user.ID,
		BillingAccountID: billingAccount.ID,
		Name:             "Test Project",
		Description:      "Test project for integration tests",
		TenantID:         s.TenantID,
		Active:           true,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}
	require.NoError(t, s.DB.Create(project).Error)
	s.ProjectID = project.ID
}

// TeardownIntegrationTest cleans up all resources
func (s *IntegrationTestSuite) TeardownIntegrationTest(t *testing.T) {
	t.Helper()

	// Close connections
	if s.RabbitMQChannel != nil {
		_ = s.RabbitMQChannel.Close()
	}
	if s.RabbitMQConn != nil {
		_ = s.RabbitMQConn.Close()
	}
	if s.RedisClient != nil {
		_ = s.RedisClient.Close()
	}
	if s.TemporalClient != nil {
		s.TemporalClient.Close()
	}

	// Terminate containers
	if s.PostgresContainer != nil {
		_ = s.PostgresContainer.Terminate(s.Ctx)
	}
}

// TemporalTestSuite wraps Temporal test suite for workflow testing
type TemporalTestSuite struct {
	*testsuite.WorkflowTestSuite
	env *testsuite.TestWorkflowEnvironment
}

// SetupTemporalTest creates a new Temporal test environment
func SetupTemporalTest() *TemporalTestSuite {
	ts := &TemporalTestSuite{
		WorkflowTestSuite: &testsuite.WorkflowTestSuite{},
	}
	ts.env = ts.NewTestWorkflowEnvironment()
	return ts
}

// GetEnv returns the test workflow environment
func (ts *TemporalTestSuite) GetEnv() *testsuite.TestWorkflowEnvironment {
	return ts.env
}

// Teardown cleans up the test environment
func (ts *TemporalTestSuite) Teardown() {
	if ts.env != nil {
		ts.env.AssertExpectations(ts.WorkflowTestSuite.T())
	}
}

// MockHelpers provides common mock data generators
type MockHelpers struct{}

// NewMockContact creates a mock contact entity
func (m *MockHelpers) NewMockContact(tenantID string, projectID uuid.UUID) *entities.ContactEntity {
	return &entities.ContactEntity{
		ID:        uuid.New(),
		TenantID:  tenantID,
		ProjectID: projectID,
		Name:      "Test Contact",
		Phone:     "+5511999999999",
		Email:     stringPtr("test@example.com"),
		Language:  "pt",
		Tags:      []string{"test"},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// NewMockSession creates a mock session entity
func (m *MockHelpers) NewMockSession(tenantID string, contactID uuid.UUID) *entities.SessionEntity {
	channelTypeID := 1
	return &entities.SessionEntity{
		ID:                  uuid.New(),
		TenantID:            tenantID,
		ContactID:           contactID,
		ChannelTypeID:       &channelTypeID,
		Status:              "active",
		TimeoutDuration:     1800000000000,
		StartedAt:           time.Now(),
		LastActivityAt:      time.Now(),
		MessageCount:        0,
		MessagesFromContact: 0,
		MessagesFromAgent:   0,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}
}

// NewMockMessage creates a mock message entity
func (m *MockHelpers) NewMockMessage(tenantID string, sessionID, contactID uuid.UUID) *entities.MessageEntity {
	return &entities.MessageEntity{
		ID:          uuid.New(),
		TenantID:    tenantID,
		SessionID:   &sessionID,
		ContactID:   contactID,
		ExternalID:  stringPtr("ext-" + uuid.New().String()),
		Direction:   "inbound",
		MessageType: "text",
		Status:      "received",
		Body:        stringPtr("Test message"),
		ReceivedAt:  timePtr(time.Now()),
		ProcessedAt: timePtr(time.Now()),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// Helper functions
func stringPtr(s string) *string {
	return &s
}

func timePtr(t time.Time) *time.Time {
	return &t
}

func intPtr(i int) *int {
	return &i
}

// AssertNoRabbitMQMessages ensures queue is empty
func AssertNoRabbitMQMessages(t *testing.T, ch *amqp.Channel, queueName string) {
	t.Helper()

	queue, err := ch.QueueInspect(queueName)
	if err != nil {
		// Queue might not exist, which is also acceptable
		return
	}

	require.Equal(t, 0, queue.Messages, "Expected no messages in queue %s", queueName)
}

// ConsumeRabbitMQMessage consumes one message from queue with timeout
func ConsumeRabbitMQMessage(t *testing.T, ch *amqp.Channel, queueName string, timeout time.Duration) *amqp.Delivery {
	t.Helper()

	msgs, err := ch.Consume(
		queueName,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	require.NoError(t, err)

	select {
	case msg := <-msgs:
		return &msg
	case <-time.After(timeout):
		t.Fatalf("Timeout waiting for message on queue %s", queueName)
		return nil
	}
}
