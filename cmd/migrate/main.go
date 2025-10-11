package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/caloi/ventros-crm/infrastructure/config"
	"github.com/caloi/ventros-crm/infrastructure/database"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

// migrate is a CLI tool for managing database migrations
//
// Usage:
//
//	go run cmd/migrate/main.go up              - Apply all pending migrations
//	go run cmd/migrate/main.go down            - Rollback last migration
//	go run cmd/migrate/main.go status          - Show migration status
//	go run cmd/migrate/main.go force <version> - Force version (recovery only)
//	go run cmd/migrate/main.go steps <n>       - Apply/rollback N migrations
//
// Examples:
//
//	go run cmd/migrate/main.go up
//	go run cmd/migrate/main.go down
//	go run cmd/migrate/main.go steps 2
//	go run cmd/migrate/main.go steps -1
//	go run cmd/migrate/main.go force 28
func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	// Initialize logger
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	// Load config
	cfg := config.Load()

	// Connect to database
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Name,
		cfg.Database.SSLMode,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	// Create migration runner
	runner, err := database.NewMigrationRunner(db, logger)
	if err != nil {
		logger.Fatal("Failed to create migration runner", zap.Error(err))
	}
	defer runner.Close()

	// Execute command
	switch command {
	case "up":
		handleUp(runner, logger)
	case "down":
		handleDown(runner, logger)
	case "status":
		handleStatus(runner, logger)
	case "force":
		handleForce(runner, logger, os.Args)
	case "steps":
		handleSteps(runner, logger, os.Args)
	default:
		fmt.Printf("Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

func handleUp(runner *database.MigrationRunner, logger *zap.Logger) {
	fmt.Println("📦 Applying all pending migrations...")

	if err := runner.Up(); err != nil {
		logger.Fatal("Failed to apply migrations", zap.Error(err))
	}

	status, err := runner.Status()
	if err != nil {
		logger.Fatal("Failed to get status", zap.Error(err))
	}

	fmt.Printf("\n%s\n", status.Message)
	fmt.Printf("Current version: %d\n", status.Version)
}

func handleDown(runner *database.MigrationRunner, logger *zap.Logger) {
	fmt.Println("⚠️  Rolling back last migration...")
	fmt.Println("Are you sure? (type 'yes' to confirm)")

	var confirm string
	fmt.Scanln(&confirm)

	if confirm != "yes" {
		fmt.Println("Rollback cancelled")
		return
	}

	if err := runner.Down(); err != nil {
		logger.Fatal("Failed to rollback migration", zap.Error(err))
	}

	status, err := runner.Status()
	if err != nil {
		logger.Fatal("Failed to get status", zap.Error(err))
	}

	fmt.Printf("\n%s\n", status.Message)
	fmt.Printf("Current version: %d\n", status.Version)
}

func handleStatus(runner *database.MigrationRunner, logger *zap.Logger) {
	status, err := runner.Status()
	if err != nil {
		logger.Fatal("Failed to get migration status", zap.Error(err))
	}

	fmt.Println("📊 Migration Status")
	fmt.Println("==================")
	fmt.Printf("Version: %d\n", status.Version)
	fmt.Printf("Dirty: %v\n", status.Dirty)
	fmt.Printf("Status: %s\n", status.Message)

	if status.Dirty {
		fmt.Println("\n⚠️  WARNING: Database is in DIRTY state!")
		fmt.Println("This means a migration failed mid-way.")
		fmt.Println("Steps to recover:")
		fmt.Println("  1. Manually inspect and fix database schema")
		fmt.Println("  2. Run: go run cmd/migrate/main.go force <version>")
		fmt.Println("  3. Continue with normal migrations")
	}
}

func handleForce(runner *database.MigrationRunner, logger *zap.Logger, args []string) {
	if len(args) < 3 {
		fmt.Println("Error: version argument required")
		fmt.Println("Usage: go run cmd/migrate/main.go force <version>")
		os.Exit(1)
	}

	version, err := strconv.Atoi(args[2])
	if err != nil {
		logger.Fatal("Invalid version number", zap.Error(err))
	}

	fmt.Printf("⚠️  FORCING migration version to %d\n", version)
	fmt.Println("This is DANGEROUS and should ONLY be used to recover from dirty state!")
	fmt.Println("Have you manually fixed the database schema? (type 'yes' to confirm)")

	var confirm string
	fmt.Scanln(&confirm)

	if confirm != "yes" {
		fmt.Println("Force cancelled")
		return
	}

	if err := runner.Force(version); err != nil {
		logger.Fatal("Failed to force version", zap.Error(err))
	}

	fmt.Printf("\n✅ Version forced to %d\n", version)
	fmt.Println("You can now continue with normal migrations")
}

func handleSteps(runner *database.MigrationRunner, logger *zap.Logger, args []string) {
	if len(args) < 3 {
		fmt.Println("Error: steps argument required")
		fmt.Println("Usage: go run cmd/migrate/main.go steps <n>")
		fmt.Println("  Positive N: Apply next N migrations")
		fmt.Println("  Negative N: Rollback N migrations")
		os.Exit(1)
	}

	steps, err := strconv.Atoi(args[2])
	if err != nil {
		logger.Fatal("Invalid steps number", zap.Error(err))
	}

	if steps < 0 {
		fmt.Printf("⚠️  Rolling back %d migration(s)...\n", -steps)
		fmt.Println("Are you sure? (type 'yes' to confirm)")

		var confirm string
		fmt.Scanln(&confirm)

		if confirm != "yes" {
			fmt.Println("Rollback cancelled")
			return
		}
	} else {
		fmt.Printf("📦 Applying %d migration(s)...\n", steps)
	}

	if err := runner.Steps(steps); err != nil {
		logger.Fatal("Failed to apply steps", zap.Error(err))
	}

	status, err := runner.Status()
	if err != nil {
		logger.Fatal("Failed to get status", zap.Error(err))
	}

	fmt.Printf("\n%s\n", status.Message)
	fmt.Printf("Current version: %d\n", status.Version)
}

func printUsage() {
	fmt.Println("Ventros CRM Migration Tool")
	fmt.Println("===========================")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  go run cmd/migrate/main.go <command> [arguments]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  up              Apply all pending migrations")
	fmt.Println("  down            Rollback last migration (with confirmation)")
	fmt.Println("  status          Show current migration status")
	fmt.Println("  force <version> Force version (recovery only - DANGEROUS)")
	fmt.Println("  steps <n>       Apply/rollback N migrations")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  go run cmd/migrate/main.go up")
	fmt.Println("  go run cmd/migrate/main.go down")
	fmt.Println("  go run cmd/migrate/main.go status")
	fmt.Println("  go run cmd/migrate/main.go steps 2    # Apply next 2 migrations")
	fmt.Println("  go run cmd/migrate/main.go steps -1   # Rollback 1 migration")
	fmt.Println("  go run cmd/migrate/main.go force 28   # Force version to 28")
	fmt.Println()
	fmt.Println("Note:")
	fmt.Println("  - Migrations are located in infrastructure/database/migrations/")
	fmt.Println("  - Each migration has .up.sql (apply) and .down.sql (rollback)")
	fmt.Println("  - The API applies migrations automatically on startup")
	fmt.Println("  - Use this CLI for manual migration management")
}
