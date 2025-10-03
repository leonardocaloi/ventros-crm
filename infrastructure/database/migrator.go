package database

import (
	"context"

	"github.com/caloi/ventros-crm/ent"
	"github.com/caloi/ventros-crm/ent/migrate"
	"go.uber.org/zap"
)

// AutoMigrator handles automatic database migrations using Ent
type AutoMigrator struct {
	client *ent.Client
	logger *zap.Logger
}

// NewAutoMigrator creates a new auto migrator
func NewAutoMigrator(client *ent.Client, logger *zap.Logger) *AutoMigrator {
	return &AutoMigrator{
		client: client,
		logger: logger,
	}
}

// Run executes automatic migrations
// This is idempotent and safe to run multiple times
// Note: Use 'make db-sync' for full DB sync (including dropping extra tables)
func (m *AutoMigrator) Run(ctx context.Context) error {
	m.logger.Info("Starting automatic database migration...")

	// Run auto-migration (creates tables, drops unused columns/indexes)
	err := m.client.Schema.Create(
		ctx,
		migrate.WithDropIndex(true),   // Drop unused indexes
		migrate.WithDropColumn(true),  // Drop unused columns
		migrate.WithForeignKeys(true), // Create foreign keys
	)

	if err != nil {
		m.logger.Error("Migration failed", zap.Error(err))
		return err
	}

	m.logger.Info("✅ Database migration completed successfully")
	return nil
}

// CheckStatus verifies if migrations are applied by counting sessions
func (m *AutoMigrator) CheckStatus(ctx context.Context) (bool, error) {
	// Try to query the sessions table - if it doesn't exist, query will fail
	count, err := m.client.Session.Query().Count(ctx)
	if err != nil {
		m.logger.Warn("Database schema not initialized", zap.Error(err))
		return false, nil
	}

	m.logger.Info("✅ Database schema exists", zap.Int("sessions_count", count))
	return true, nil
}
