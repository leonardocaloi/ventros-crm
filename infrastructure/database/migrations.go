package database

// ⚠️ DEPRECATED FILE
//
// This file previously contained a manual migration manager.
// It has been replaced by migration_runner.go which uses golang-migrate.
//
// Migration functionality now provided by:
//   - migration_runner.go: Production-ready migration runner using golang-migrate
//   - cmd/migrate/main.go: CLI tool for manual migration management
//
// For migration usage, see:
//   - MIGRATIONS.md: Complete migration guide
//   - migration_runner.go: MigrationRunner implementation
//
// All migrations are in infrastructure/database/migrations/*.sql
// and are embedded in the binary using go:embed directive.
