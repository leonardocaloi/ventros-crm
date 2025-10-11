package persistence

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/caloi/ventros-crm/infrastructure/persistence/entities"
	gormpostgres "gorm.io/driver/postgres"
)

// TestDatabase holds the test database container and connection
type TestDatabase struct {
	Container *postgres.PostgresContainer
	DB        *gorm.DB
	DSN       string
	ctx       context.Context
}

// SetupTestDatabase starts a PostgreSQL test container and returns a GORM connection
func SetupTestDatabase(t *testing.T) *TestDatabase {
	t.Helper()

	ctx := context.Background()

	// Create PostgreSQL container
	postgresContainer, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("ventros_test"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second)),
	)
	if err != nil {
		t.Fatalf("Failed to start postgres container: %v", err)
	}

	// Get connection string
	dsn, err := postgresContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("Failed to get connection string: %v", err)
	}

	// Connect to database with GORM
	db, err := gorm.Open(gormpostgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent), // Silencia logs durante testes
	})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Run migrations
	err = AutoMigrateTest(db)
	if err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	return &TestDatabase{
		Container: postgresContainer,
		DB:        db,
		DSN:       dsn,
		ctx:       ctx,
	}
}

// TeardownTestDatabase stops the test container and cleans up
func (td *TestDatabase) TeardownTestDatabase(t *testing.T) {
	t.Helper()

	if td.Container != nil {
		if err := td.Container.Terminate(td.ctx); err != nil {
			t.Logf("Failed to terminate container: %v", err)
		}
	}
}

// AutoMigrateTest runs all migrations for test database
//
// ⚠️ WARNING: This function uses GORM AutoMigrate and is ONLY for tests!
// ⚠️ NEVER use AutoMigrate in production code - use SQL migrations instead.
//
// This is acceptable for tests because:
// 1. Tests need quick schema setup for isolated test databases
// 2. Tests run in containers (testcontainers) that are destroyed after each test
// 3. Test schema doesn't need rollback capabilities
//
// For production, use SQL migrations located in infrastructure/database/migrations/*.sql
func AutoMigrateTest(db *gorm.DB) error {
	return db.AutoMigrate(
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
	)
}

// SeedTestData populates the database with test data
func (td *TestDatabase) SeedTestData(t *testing.T) *TestData {
	t.Helper()

	testData := &TestData{
		TenantID: "test-tenant-" + uuid.New().String()[:8],
	}

	// Create user first (required by billing account and project)
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
	if err := td.DB.Create(user).Error; err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Create billing account (required by project)
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
	if err := td.DB.Create(billingAccount).Error; err != nil {
		t.Fatalf("Failed to create test billing account: %v", err)
	}

	// Create test project
	project := &entities.ProjectEntity{
		ID:               uuid.New(),
		UserID:           user.ID,
		BillingAccountID: billingAccount.ID,
		Name:             "Test Project",
		Description:      "Test project for integration tests",
		TenantID:         testData.TenantID,
		Active:           true,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}
	if err := td.DB.Create(project).Error; err != nil {
		t.Fatalf("Failed to create test project: %v", err)
	}
	testData.ProjectID = project.ID

	// Create test pipeline
	timeoutMinutes := 30
	pipeline := &entities.PipelineEntity{
		ID:                    uuid.New(),
		TenantID:              testData.TenantID,
		ProjectID:             project.ID,
		Name:                  "Test Pipeline",
		Description:           "Test pipeline",
		Active:                true,
		SessionTimeoutMinutes: &timeoutMinutes,
		Position:              1,
		CreatedAt:             time.Now(),
		UpdatedAt:             time.Now(),
	}
	if err := td.DB.Create(pipeline).Error; err != nil {
		t.Fatalf("Failed to create test pipeline: %v", err)
	}
	testData.PipelineID = pipeline.ID

	// Create default pipeline statuses
	statuses := []entities.PipelineStatusEntity{
		{
			ID:         uuid.New(),
			PipelineID: pipeline.ID,
			Name:       "New",
			StatusType: "open",
			Position:   1,
			Active:     true,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		},
		{
			ID:         uuid.New(),
			PipelineID: pipeline.ID,
			Name:       "Contacted",
			StatusType: "active",
			Position:   2,
			Active:     true,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		},
		{
			ID:         uuid.New(),
			PipelineID: pipeline.ID,
			Name:       "Qualified",
			StatusType: "active",
			Position:   3,
			Active:     true,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		},
	}
	for _, status := range statuses {
		if err := td.DB.Create(&status).Error; err != nil {
			t.Fatalf("Failed to create pipeline status: %v", err)
		}
	}
	testData.InitialStatusID = statuses[0].ID

	// Create test channel type
	channelType := &entities.ChannelTypeEntity{
		ID:          1,
		Name:        "WhatsApp",
		Description: "WhatsApp channel",
		Provider:    "waha",
		Active:      true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	if err := td.DB.Create(channelType).Error; err != nil {
		// Ignore error if channel type already exists
		if err := td.DB.First(channelType, 1).Error; err != nil {
			t.Fatalf("Failed to get channel type: %v", err)
		}
	}
	testData.ChannelTypeID = channelType.ID

	return testData
}

// TestData holds IDs of seeded test data
type TestData struct {
	TenantID        string
	ProjectID       uuid.UUID
	PipelineID      uuid.UUID
	InitialStatusID uuid.UUID
	ChannelTypeID   int
}

// CreateTestContact creates a test contact
func (td *TestDatabase) CreateTestContact(t *testing.T, testData *TestData, name string) *entities.ContactEntity {
	t.Helper()

	contact := &entities.ContactEntity{
		ID:        uuid.New(),
		TenantID:  testData.TenantID,
		ProjectID: testData.ProjectID,
		Name:      name,
		Language:  "en",
		Tags:      []string{},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := td.DB.Create(contact).Error; err != nil {
		t.Fatalf("Failed to create test contact: %v", err)
	}

	return contact
}

// CreateTestSession creates a test session
func (td *TestDatabase) CreateTestSession(t *testing.T, testData *TestData, contactID uuid.UUID) *entities.SessionEntity {
	t.Helper()

	session := &entities.SessionEntity{
		ID:                  uuid.New(),
		TenantID:            testData.TenantID,
		ContactID:           contactID,
		ChannelTypeID:       &testData.ChannelTypeID,
		PipelineID:          &testData.PipelineID,
		Status:              "active",
		TimeoutDuration:     1800000000000, // 30 minutes in nanoseconds
		StartedAt:           time.Now(),
		LastActivityAt:      time.Now(),
		MessageCount:        0,
		MessagesFromContact: 0,
		MessagesFromAgent:   0,
		AgentTransfers:      0,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}

	if err := td.DB.Create(session).Error; err != nil {
		t.Fatalf("Failed to create test session: %v", err)
	}

	return session
}

// CleanupTestData removes all test data for a tenant
func (td *TestDatabase) CleanupTestData(t *testing.T, tenantID string) {
	t.Helper()

	// Delete in reverse order of dependencies
	td.DB.Where("tenant_id = ?", tenantID).Delete(&entities.MessageEntity{})
	td.DB.Where("tenant_id = ?", tenantID).Delete(&entities.SessionEntity{})
	td.DB.Where("tenant_id = ?", tenantID).Delete(&entities.ContactEntity{})
	td.DB.Where("tenant_id = ?", tenantID).Delete(&entities.PipelineStatusEntity{})
	td.DB.Where("tenant_id = ?", tenantID).Delete(&entities.PipelineEntity{})
	td.DB.Where("tenant_id = ?", tenantID).Delete(&entities.ProjectEntity{})
}
