package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"

	_ "github.com/lib/pq"

	"github.com/caloi/ventros-crm/infrastructure/persistence"
	"github.com/caloi/ventros-crm/infrastructure/persistence/entities"
)

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func main() {
	log.Println("🔄 Running GORM AutoMigrate...")

	// Load DB config from env
	port, _ := strconv.Atoi(getEnv("DB_PORT", "5432"))
	dbConfig := persistence.DatabaseConfig{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     port,
		User:     getEnv("DB_USER", "ventros"),
		Password: getEnv("DB_PASSWORD", "ventros123"),
		DBName:   getEnv("DB_NAME", "ventros_crm"),
		SSLMode:  "disable",
	}

	// Connect to database
	db, err := persistence.NewDatabase(dbConfig)
	if err != nil {
		log.Fatal("❌ Failed to connect to database:", err)
	}

	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	// AutoMigrate all entities
	log.Println("📋 Syncing schema with entities...")

	// Core tables (no dependencies)
	if err := db.AutoMigrate(
		&entities.UserEntity{},
		&entities.BillingAccountEntity{},
		&entities.ProjectEntity{},
		&entities.ChannelTypeEntity{},
	); err != nil {
		log.Fatal("❌ Failed to migrate core tables:", err)
	}

	// Dependent tables
	if err := db.AutoMigrate(
		&entities.ChannelEntity{},
		&entities.PipelineEntity{},
		&entities.PipelineStatusEntity{},
		&entities.ContactEntity{},
		&entities.ContactPipelineStatusEntity{},
		&entities.SessionEntity{},
		&entities.MessageEntity{},
		&entities.NoteEntity{},
		&entities.AgentEntity{},
		&entities.AgentSessionEntity{},
		&entities.AutomationEntity{},
		&entities.WebhookSubscriptionEntity{},
		&entities.UserAPIKeyEntity{},
		&entities.CredentialEntity{},
		&entities.ContactEventEntity{},
		&entities.ContactListEntity{},
		&entities.TrackingEntity{},
		&entities.TrackingEnrichmentEntity{},
		&entities.OutboxEventEntity{},
		&entities.ProcessedEventEntity{},
		&entities.DomainEventLogEntity{},
		&entities.ChatEntity{},
		&entities.MessageGroupEntity{},
		&entities.MessageEnrichmentEntity{},
		&entities.AIAgentHistoryEntity{},
		&entities.ContactEventStoreEntity{},
		&entities.ContactSnapshotEntity{},
	); err != nil {
		log.Fatal("❌ Failed to migrate dependent tables:", err)
	}

	log.Println("✅ Schema synced successfully!")

	// Mark all migrations as applied
	log.Println("📝 Marking all migrations as applied...")

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		dbConfig.User, dbConfig.Password, dbConfig.Host, dbConfig.Port, dbConfig.DBName)

	conn, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal("❌ Failed to connect:", err)
	}
	defer conn.Close()

	// Get latest migration version (from files)
	latestVersion := 41 // Update this when adding new migrations

	// Update schema_migrations table
	_, err = conn.Exec(`
		INSERT INTO schema_migrations (version, dirty)
		VALUES ($1, false)
		ON CONFLICT (version) DO UPDATE SET dirty = false
	`, latestVersion)
	if err != nil {
		log.Fatal("❌ Failed to update schema_migrations:", err)
	}

	log.Printf("✅ Marked database as version %d (all migrations applied)\n", latestVersion)
	log.Println("")
	log.Println("🎉 AutoMigrate completed successfully!")
	log.Println("   You can now run: make api")
}
