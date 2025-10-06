package main

import (
	"log"
	"os"

	"github.com/caloi/ventros-crm/infrastructure/persistence"
)

func main() {
	log.Println("🔄 Starting GORM database migration...")

	// Database configuration from environment
	config := persistence.DatabaseConfig{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     5432,
		User:     getEnv("DB_USER", "ventros"),
		Password: getEnv("DB_PASSWORD", "ventros123"),
		DBName:   getEnv("DB_NAME", "ventros_crm"),
		SSLMode:  getEnv("DB_SSLMODE", "disable"),
	}

	// Connect to database
	db, err := persistence.NewDatabase(config)
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}

	// Get underlying SQL DB for connection management
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("❌ Failed to get SQL DB: %v", err)
	}
	defer sqlDB.Close()

	log.Println("✅ Connected to database successfully!")

	// Run auto-migrations
	if err := persistence.AutoMigrate(db); err != nil {
		log.Fatalf("❌ Migration failed: %v", err)
	}

	// Create additional indexes
	if err := persistence.CreateIndexes(db); err != nil {
		log.Printf("⚠️  Warning: Failed to create some indexes: %v", err)
	}

	log.Println("🎉 Database migration completed successfully!")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
