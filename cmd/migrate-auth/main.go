package main

import (
	"fmt"
	"log"
	"os"

	"github.com/caloi/ventros-crm/infrastructure/persistence/entities"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	// Database connection from environment variables
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "ventros")
	dbPassword := getEnv("DB_PASSWORD", "ventros123")
	dbName := getEnv("DB_NAME", "ventros_crm")
	dbSSLMode := getEnv("DB_SSLMODE", "disable")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=UTC",
		dbHost, dbPort, dbUser, dbPassword, dbName, dbSSLMode)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	fmt.Println("🔄 Running authentication system migrations...")

	// Enable UUID extension (replaces init.sql)
	if err := db.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"").Error; err != nil {
		log.Printf("⚠️  Failed to create uuid-ossp extension (may already exist): %v", err)
	}

	// Auto migrate all entities
	err = db.AutoMigrate(
		&entities.UserEntity{},
		&entities.UserAPIKeyEntity{},
		&entities.ProjectEntity{},
		&entities.PipelineEntity{},
		&entities.ContactEntity{},
		&entities.MessageEntity{},
		&entities.WebhookSubscriptionEntity{},
		&entities.ChannelEntity{},
	)
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	fmt.Println("✅ Authentication migrations completed successfully!")

	// Create default admin user
	createDefaultAdmin(db)

	fmt.Println("🎯 Next steps:")
	fmt.Println("  1. Start the API: make full-up")
	fmt.Println("  2. Login: POST /api/v1/auth/login")
	fmt.Println("  3. Use the returned API key in Authorization header")
}

func createDefaultAdmin(db *gorm.DB) {
	adminEmail := os.Getenv("ADMIN_EMAIL")
	adminPassword := os.Getenv("ADMIN_PASSWORD")
	adminName := os.Getenv("ADMIN_NAME")

	// Use defaults if not set
	if adminEmail == "" {
		adminEmail = "admin@ventros.com"
	}
	if adminPassword == "" {
		adminPassword = "admin123"
	}
	if adminName == "" {
		adminName = "Administrator"
	}

	fmt.Println("👤 Checking for admin user...")

	// Check if admin user already exists
	var existingUser entities.UserEntity
	result := db.Where("email = ?", adminEmail).First(&existingUser)

	if result.Error == gorm.ErrRecordNotFound {
		// Hash password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(adminPassword), bcrypt.DefaultCost)
		if err != nil {
			log.Printf("⚠️  Failed to hash admin password: %v", err)
			return
		}

		// Create admin user
		adminUser := entities.UserEntity{
			ID:           uuid.New(),
			Name:         adminName,
			Email:        adminEmail,
			PasswordHash: string(hashedPassword),
			Role:         "admin",
			Status:       "active",
		}

		if err := db.Create(&adminUser).Error; err != nil {
			log.Printf("⚠️  Failed to create admin user: %v", err)
			return
		}

		// Create billing account
		billingAccount := entities.BillingAccountEntity{
			ID:           uuid.New(),
			UserID:       adminUser.ID,
			Name:         fmt.Sprintf("Conta de %s", adminName),
			BillingEmail: adminEmail,
		}

		if err := db.Create(&billingAccount).Error; err != nil {
			log.Printf("⚠️  Failed to create billing account: %v", err)
			return
		}

		// Create default project
		defaultProject := entities.ProjectEntity{
			ID:               uuid.New(),
			UserID:           adminUser.ID,
			BillingAccountID: billingAccount.ID,
			TenantID:         adminUser.ID.String(),
			Name:             "Admin Default Project",
			Active:           true,
		}

		if err := db.Create(&defaultProject).Error; err != nil {
			log.Printf("⚠️  Failed to create default project: %v", err)
			return
		}

		// Create default pipeline
		defaultPipeline := entities.PipelineEntity{
			ID:        uuid.New(),
			ProjectID: defaultProject.ID,
			TenantID:  adminUser.ID.String(),
			Name:      "Admin Default Pipeline",
			Active:    true,
		}

		if err := db.Create(&defaultPipeline).Error; err != nil {
			log.Printf("⚠️  Failed to create default pipeline: %v", err)
			return
		}

		fmt.Printf("✅ Admin user created!\n")
		fmt.Printf("   Email: %s\n", adminEmail)
		fmt.Printf("   Password: %s\n", adminPassword)
		fmt.Printf("   Project: %s\n", defaultProject.ID)
		fmt.Printf("   Pipeline: %s\n", defaultPipeline.ID)
		fmt.Println("   ⚠️  CHANGE THIS PASSWORD IN PRODUCTION!")
	} else if result.Error != nil {
		log.Printf("⚠️  Error checking for admin user: %v", result.Error)
	} else {
		fmt.Printf("✓ Admin user already exists: %s\n", adminEmail)
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
