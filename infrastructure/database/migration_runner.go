package database

import (
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"io/fs"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"go.uber.org/zap"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// MigrationRunner manages database migrations using golang-migrate
//
// Features:
// - Embedded SQL files (no external dependencies)
// - Automatic version tracking
// - Rollback support (.down.sql files)
// - Thread-safe
// - Production-ready
type MigrationRunner struct {
	db     *sql.DB
	logger *zap.Logger
	m      *migrate.Migrate
}

// NewMigrationRunner creates a new migration runner
//
// Example:
//   runner, err := NewMigrationRunner(sqlDB, logger)
//   if err != nil {
//       log.Fatal(err)
//   }
//   defer runner.Close()
//
//   // Apply all pending migrations
//   if err := runner.Up(); err != nil {
//       log.Fatal(err)
//   }
func NewMigrationRunner(db *sql.DB, logger *zap.Logger) (*MigrationRunner, error) {
	// Create postgres driver instance
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to create postgres driver: %w", err)
	}

	// Create migration source from embedded FS
	// migrations/ subdirectory contains *.sql files
	migrationsSubFS, err := fs.Sub(migrationsFS, "migrations")
	if err != nil {
		return nil, fmt.Errorf("failed to access migrations directory: %w", err)
	}

	source, err := iofs.New(migrationsSubFS, ".")
	if err != nil {
		return nil, fmt.Errorf("failed to create migration source: %w", err)
	}

	// Create migrate instance
	m, err := migrate.NewWithInstance("iofs", source, "postgres", driver)
	if err != nil {
		return nil, fmt.Errorf("failed to create migrate instance: %w", err)
	}

	return &MigrationRunner{
		db:     db,
		logger: logger,
		m:      m,
	}, nil
}

// Up applies all pending migrations
//
// This is idempotent - safe to call multiple times.
// If all migrations are already applied, this is a no-op.
//
// Returns:
//   - nil if migrations were applied successfully or already up to date
//   - error if migrations failed
func (r *MigrationRunner) Up() error {
	r.logger.Info("🔄 Applying database migrations...")

	err := r.m.Up()
	if err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			r.logger.Info("✅ Database is already up to date (no pending migrations)")
			return nil
		}
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	version, dirty, _ := r.m.Version()
	r.logger.Info("✅ Migrations applied successfully",
		zap.Uint("version", version),
		zap.Bool("dirty", dirty))

	return nil
}

// Down rolls back the last migration
//
// ⚠️ USE WITH CAUTION IN PRODUCTION!
//
// This executes the .down.sql file for the most recent migration.
// Always test rollbacks in staging first!
//
// Returns:
//   - nil if rollback was successful
//   - error if rollback failed
func (r *MigrationRunner) Down() error {
	r.logger.Warn("⚠️  Rolling back last migration...")

	err := r.m.Down()
	if err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			r.logger.Info("ℹ️  No migrations to roll back")
			return nil
		}
		return fmt.Errorf("failed to rollback migration: %w", err)
	}

	version, dirty, _ := r.m.Version()
	r.logger.Info("✅ Migration rolled back successfully",
		zap.Uint("version", version),
		zap.Bool("dirty", dirty))

	return nil
}

// Steps applies N migrations (positive) or rolls back N migrations (negative)
//
// Examples:
//   Steps(2)  - Apply next 2 migrations
//   Steps(-1) - Rollback 1 migration
//   Steps(-3) - Rollback 3 migrations
//
// Returns:
//   - nil if successful
//   - error if failed
func (r *MigrationRunner) Steps(n int) error {
	if n == 0 {
		return nil
	}

	action := "Applying"
	if n < 0 {
		action = "Rolling back"
	}

	r.logger.Info(fmt.Sprintf("🔄 %s %d migration(s)...", action, abs(n)),
		zap.Int("steps", n))

	err := r.m.Steps(n)
	if err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			r.logger.Info("ℹ️  No migrations to apply/rollback")
			return nil
		}
		return fmt.Errorf("failed to apply %d steps: %w", n, err)
	}

	version, dirty, _ := r.m.Version()
	r.logger.Info("✅ Migration steps completed",
		zap.Uint("version", version),
		zap.Bool("dirty", dirty))

	return nil
}

// Version returns the current migration version
//
// Returns:
//   - version: Current migration version number
//   - dirty: true if database is in a dirty state (migration failed mid-way)
//   - error: if unable to determine version
func (r *MigrationRunner) Version() (version uint, dirty bool, err error) {
	version, dirty, err = r.m.Version()
	if err != nil {
		if errors.Is(err, migrate.ErrNilVersion) {
			// No migrations applied yet
			return 0, false, nil
		}
		return 0, false, fmt.Errorf("failed to get migration version: %w", err)
	}
	return version, dirty, nil
}

// Status returns detailed migration status
//
// Returns:
//   - MigrationStatus with version, dirty flag, and message
//   - error if unable to determine status
func (r *MigrationRunner) Status() (*MigrationStatus, error) {
	version, dirty, err := r.Version()
	if err != nil {
		return nil, err
	}

	status := &MigrationStatus{
		Version: version,
		Dirty:   dirty,
	}

	if version == 0 {
		status.Message = "No migrations applied yet"
	} else if dirty {
		status.Message = fmt.Sprintf("⚠️  Database is DIRTY at version %d - migration failed mid-way!", version)
	} else {
		status.Message = fmt.Sprintf("✅ Database is up to date at version %d", version)
	}

	return status, nil
}

// Force sets the migration version without running migrations
//
// ⚠️ DANGEROUS - ONLY USE TO RECOVER FROM DIRTY STATE!
//
// This is used when a migration failed and left the database in a dirty state.
// You must manually fix the database schema first, then force the version.
//
// Example recovery from dirty state:
//   1. Manually inspect database and fix issues
//   2. Force version: runner.Force(currentVersion)
//   3. Continue with normal migrations
//
// Parameters:
//   - version: The version to force the database to
//
// Returns:
//   - error if forcing version failed
func (r *MigrationRunner) Force(version int) error {
	r.logger.Warn("⚠️  FORCING migration version (DANGEROUS!)",
		zap.Int("version", version))

	if err := r.m.Force(version); err != nil {
		return fmt.Errorf("failed to force version %d: %w", version, err)
	}

	r.logger.Info("✅ Migration version forced successfully",
		zap.Int("version", version))

	return nil
}

// Close closes the migration runner
//
// This should be called when done with migrations to clean up resources.
// Always defer this after creating the runner:
//
//   runner, err := NewMigrationRunner(db, logger)
//   if err != nil {
//       return err
//   }
//   defer runner.Close()
func (r *MigrationRunner) Close() error {
	sourceErr, dbErr := r.m.Close()
	if sourceErr != nil {
		return fmt.Errorf("failed to close migration source: %w", sourceErr)
	}
	if dbErr != nil {
		return fmt.Errorf("failed to close migration database: %w", dbErr)
	}
	return nil
}

// MigrationStatus represents the current state of database migrations
type MigrationStatus struct {
	Version uint   // Current migration version
	Dirty   bool   // True if migration failed mid-way
	Message string // Human-readable status message
}

// Helper function to get absolute value
func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}
