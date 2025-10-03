package database

import (
	"context"
	"database/sql"
	"fmt"

	"go.uber.org/zap"
)

// MigrationManager handles database migrations
// NOTE: In production, migrations should be handled by Atlas CLI
// This is only for checking migration status
type MigrationManager struct {
	db     *sql.DB
	logger *zap.Logger
}

// NewMigrationManager creates a new migration manager
func NewMigrationManager(db *sql.DB, logger *zap.Logger) *MigrationManager {
	return &MigrationManager{
		db:     db,
		logger: logger,
	}
}

// ⚠️ NEVER USE IN PRODUCTION!
// This is only for local development auto-migration
// In production, use Atlas CLI in docker-entrypoint.sh
func (m *MigrationManager) RunAutoMigration(ctx context.Context, client interface{}) error {
	m.logger.Warn("⚠️  Running AUTO-MIGRATION - Only use in development!")
	// This would call: client.Schema.Create(ctx)
	// But we're not implementing it here to prevent misuse
	return fmt.Errorf("auto-migration not implemented - use Atlas migrations")
}

// createMigrationsTable creates the migrations tracking table
func (m *MigrationManager) createMigrationsTable(ctx context.Context) error {
	query := `
		CREATE TABLE IF NOT EXISTS atlas_schema_revisions (
			version VARCHAR(255) PRIMARY KEY,
			description TEXT,
			executed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`
	_, err := m.db.ExecContext(ctx, query)
	return err
}

// getCurrentVersion returns the current migration version
func (m *MigrationManager) getCurrentVersion(ctx context.Context) (string, error) {
	var version string
	query := `
		SELECT version 
		FROM atlas_schema_revisions 
		ORDER BY executed_at DESC 
		LIMIT 1
	`
	err := m.db.QueryRowContext(ctx, query).Scan(&version)
	if err == sql.ErrNoRows {
		return "none", nil
	}
	if err != nil {
		return "", err
	}
	return version, nil
}

// CheckMigrationsStatus verifies if migrations are up to date
func (m *MigrationManager) CheckMigrationsStatus(ctx context.Context) (bool, error) {
	// Check if migrations table exists
	var exists bool
	query := `
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_schema = 'public' 
			AND table_name = 'atlas_schema_revisions'
		)
	`
	if err := m.db.QueryRowContext(ctx, query).Scan(&exists); err != nil {
		return false, err
	}

	if !exists {
		m.logger.Warn("Migrations table does not exist - database not initialized")
		return false, nil
	}

	// Get current version
	version, err := m.getCurrentVersion(ctx)
	if err != nil {
		return false, err
	}

	if version == "none" {
		m.logger.Warn("No migrations applied yet")
		return false, nil
	}

	m.logger.Info("Migrations are present", zap.String("version", version))
	return true, nil
}
