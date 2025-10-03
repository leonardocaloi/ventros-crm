package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/caloi/ventros-crm/ent"
	"github.com/caloi/ventros-crm/ent/migrate"
	_ "github.com/lib/pq"
)

func main() {
	// Get DB connection from env or use defaults
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "ventros")
	dbPass := getEnv("DB_PASSWORD", "ventros123")
	dbName := getEnv("DB_NAME", "ventros_crm")
	sslMode := getEnv("DB_SSL_MODE", "disable")

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		dbHost, dbPort, dbUser, dbPass, dbName, sslMode,
	)

	log.Println("🔄 Connecting to database...")
	client, err := ent.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	// Step 1: Detect and drop extra tables not in schema
	log.Println("🔍 Detecting tables not declared in Ent schema...")
	if err := dropExtraTables(ctx, client); err != nil {
		log.Printf("⚠️  Warning: Could not drop extra tables: %v", err)
	}

	// Step 2: Run Ent migration
	log.Println("🔄 Running Ent schema migration...")
	log.Println("   ✅ Creating missing tables")
	log.Println("   ✅ Dropping unused columns")
	log.Println("   ✅ Dropping unused indexes")
	log.Println("   ✅ Syncing foreign keys")
	log.Println("")

	// Run migration with available options
	err = client.Schema.Create(
		ctx,
		migrate.WithDropIndex(true),   // Drop unused indexes
		migrate.WithDropColumn(true),  // Drop unused columns
		migrate.WithForeignKeys(true), // Create foreign keys
	)

	if err != nil {
		log.Fatalf("❌ Migration failed: %v", err)
	}

	log.Println("✅ Database migration completed successfully!")
	log.Println("✅ Database is now 100% synchronized with domain (DDD/Ent)")
}

func dropExtraTables(ctx context.Context, client *ent.Client) error {
	// Tables declared in Ent schema
	declaredTables := map[string]bool{
		"agents":                    true,
		"agent_sessions":            true,
		"channel_types":             true,
		"contacts":                  true,
		"contact_custom_fields":     true,
		"contact_events":            true,
		"contact_pipeline_statuses": true,
		"contact_status_histories":  true,
		"customers":                 true,
		"events":                    true,
		"messages":                  true,
		"pipelines":                 true,
		"projects":                  true,
		"sessions":                  true,
		"session_custom_fields":     true,
		"statuses":                  true,
		"webhook_subscriptions":     true,
	}

	// Open direct SQL connection
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "ventros")
	dbPass := getEnv("DB_PASSWORD", "ventros123")
	dbName := getEnv("DB_NAME", "ventros_crm")
	sslMode := getEnv("DB_SSL_MODE", "disable")

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		dbHost, dbPort, dbUser, dbPass, dbName, sslMode,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	// Query existing tables
	rows, err := db.QueryContext(ctx, `
		SELECT tablename 
		FROM pg_tables 
		WHERE schemaname = 'public'
	`)
	if err != nil {
		return fmt.Errorf("failed to query tables: %w", err)
	}
	defer rows.Close()

	var extraTables []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			continue
		}
		
		// Skip if table is declared in schema
		if declaredTables[tableName] {
			continue
		}

		extraTables = append(extraTables, tableName)
	}

	// Drop extra tables
	if len(extraTables) > 0 {
		log.Printf("⚠️  Found %d tables not declared in schema:", len(extraTables))
		for _, table := range extraTables {
			log.Printf("   - Dropping: %s", table)
			_, err := db.ExecContext(ctx, fmt.Sprintf("DROP TABLE IF EXISTS %s CASCADE", table))
			if err != nil {
				log.Printf("   ❌ Failed to drop %s: %v", table, err)
			} else {
				log.Printf("   ✅ Dropped: %s", table)
			}
		}
	} else {
		log.Println("✅ No extra tables found")
	}

	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
