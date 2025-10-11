package persistence

import (
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// NewDatabase creates a new GORM database connection
func NewDatabase(config DatabaseConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s",
		config.Host, config.User, config.Password, config.DBName, config.Port, config.SSLMode)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	return db, nil
}

// ⚠️ DEPRECATED: AutoMigrate function REMOVED
//
// This project uses 100% SQL migrations (NO GORM AutoMigrate).
// All schema changes must be in SQL migration files located at:
// infrastructure/database/migrations/*.sql
//
// Migration files follow the pattern:
// - 000001_initial_schema.up.sql    (create tables/indexes)
// - 000001_initial_schema.down.sql  (drop tables/indexes)
//
// To apply migrations, use:
// - Atlas CLI: atlas migrate apply --url "postgres://..."
// - OR any other migration tool that supports SQL files
//
// The AutoMigrate function and related schema fixes have been removed because:
// 1. GORM AutoMigrate is not reliable for production (can't handle complex migrations)
// 2. SQL migrations provide full control over schema changes
// 3. Migration rollback (.down.sql files) is critical for production safety
// 4. SQL migrations are version-controlled and reviewable
//
// All previous schema fixes, pre-migration updates, and post-migration optimizations
// have been moved to proper SQL migration files.
//
// For local development, ensure migrations are applied before starting the API:
//   atlas migrate apply --url "postgres://localhost:5432/ventros_crm?sslmode=disable"

// CreateIndexes creates additional database indexes for performance
func CreateIndexes(db *gorm.DB) error {
	log.Println("🔄 Creating additional database indexes...")

	// Indexes for better query performance
	indexes := []string{
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_contacts_project_external ON contacts(project_id, external_id) WHERE deleted_at IS NULL",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_sessions_contact_status ON sessions(contact_id, status) WHERE deleted_at IS NULL",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_messages_session_timestamp ON messages(session_id, timestamp) WHERE deleted_at IS NULL",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_messages_contact_timestamp ON messages(contact_id, timestamp) WHERE deleted_at IS NULL",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_custom_fields_contact_key ON contact_custom_fields(contact_id, field_key) WHERE deleted_at IS NULL",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_custom_fields_session_key ON session_custom_fields(session_id, field_key) WHERE deleted_at IS NULL",
	}

	for _, indexSQL := range indexes {
		if err := db.Exec(indexSQL).Error; err != nil {
			log.Printf("⚠️  Warning: Failed to create index: %v", err)
			// Continue with other indexes even if one fails
		}
	}

	log.Println("✅ Database indexes created successfully!")
	return nil
}

// SetupRLS configures Row Level Security (RLS) policies for multi-tenancy
func SetupRLS(db *gorm.DB) error {
	log.Println("🔒 Setting up Row Level Security (RLS)...")

	// 1. Criar role app_user se não existir
	if err := db.Exec(`
		DO $$
		BEGIN
			IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'app_user') THEN
				CREATE ROLE app_user;
			END IF;
		END
		$$;
	`).Error; err != nil {
		log.Printf("⚠️  Warning: Failed to create app_user role: %v", err)
	}

	// 2. Garantir permissões para o role
	if err := db.Exec(`GRANT app_user TO CURRENT_USER`).Error; err != nil {
		log.Printf("⚠️  Warning: Failed to grant app_user role: %v", err)
	}

	// 3. Criar funções helper para RLS
	rlsFunctions := []string{
		// Função para definir o usuário atual na sessão
		`CREATE OR REPLACE FUNCTION set_current_user_id(user_uuid uuid)
		RETURNS void AS $$
		BEGIN
			PERFORM set_config('app.current_user_id', user_uuid::text, false);
		END;
		$$ LANGUAGE plpgsql SECURITY DEFINER;`,

		// Função para obter o usuário atual
		`CREATE OR REPLACE FUNCTION get_current_user_id()
		RETURNS uuid AS $$
		BEGIN
			RETURN current_setting('app.current_user_id', true)::uuid;
		EXCEPTION
			WHEN OTHERS THEN
				RETURN NULL;
		END;
		$$ LANGUAGE plpgsql SECURITY DEFINER;`,
	}

	for _, fn := range rlsFunctions {
		if err := db.Exec(fn).Error; err != nil {
			log.Printf("⚠️  Warning: Failed to create RLS function: %v", err)
		}
	}

	// 4. Habilitar RLS nas tabelas principais
	tables := []string{
		"projects",
		"pipelines",
		"pipeline_statuses",
		"contacts",
		"contact_pipeline_statuses",
		"messages",
		"sessions",
		"webhook_subscriptions",
		"channels",
	}

	for _, table := range tables {
		if err := db.Exec(fmt.Sprintf("ALTER TABLE %s ENABLE ROW LEVEL SECURITY", table)).Error; err != nil {
			log.Printf("⚠️  Warning: Failed to enable RLS on %s: %v", table, err)
		}
	}

	// 5. Criar políticas RLS
	policies := []string{
		// Projects: usuário só vê seus próprios projetos
		`DROP POLICY IF EXISTS user_projects_policy ON projects;
		CREATE POLICY user_projects_policy ON projects
			FOR ALL TO app_user
			USING (user_id = current_setting('app.current_user_id', true)::uuid);`,

		// Pipelines: usuário só vê pipelines de seus projetos
		`DROP POLICY IF EXISTS user_pipelines_policy ON pipelines;
		CREATE POLICY user_pipelines_policy ON pipelines
			FOR ALL TO app_user
			USING (
				project_id IN (
					SELECT id FROM projects 
					WHERE user_id = current_setting('app.current_user_id', true)::uuid
				)
			);`,

		// Pipeline Statuses: usuário só vê status de pipelines de seus projetos
		`DROP POLICY IF EXISTS user_pipeline_statuses_policy ON pipeline_statuses;
		CREATE POLICY user_pipeline_statuses_policy ON pipeline_statuses
			FOR ALL TO app_user
			USING (
				pipeline_id IN (
					SELECT p.id FROM pipelines p
					JOIN projects pr ON p.project_id = pr.id
					WHERE pr.user_id = current_setting('app.current_user_id', true)::uuid
				)
			);`,

		// Contacts: usuário só vê contatos de seus projetos
		`DROP POLICY IF EXISTS user_contacts_policy ON contacts;
		CREATE POLICY user_contacts_policy ON contacts
			FOR ALL TO app_user
			USING (
				project_id IN (
					SELECT id FROM projects 
					WHERE user_id = current_setting('app.current_user_id', true)::uuid
				)
			);`,

		// Contact Pipeline Statuses: usuário só vê status de seus contatos
		`DROP POLICY IF EXISTS user_contact_pipeline_statuses_policy ON contact_pipeline_statuses;
		CREATE POLICY user_contact_pipeline_statuses_policy ON contact_pipeline_statuses
			FOR ALL TO app_user
			USING (
				contact_id IN (
					SELECT c.id FROM contacts c
					JOIN projects p ON c.project_id = p.id
					WHERE p.user_id = current_setting('app.current_user_id', true)::uuid
				)
			);`,

		// Messages: usuário só vê mensagens de seus contatos
		`DROP POLICY IF EXISTS user_messages_policy ON messages;
		CREATE POLICY user_messages_policy ON messages
			FOR ALL TO app_user
			USING (
				contact_id IN (
					SELECT c.id FROM contacts c
					JOIN projects p ON c.project_id = p.id
					WHERE p.user_id = current_setting('app.current_user_id', true)::uuid
				)
			);`,

		// Sessions: usuário só vê sessões de seus contatos
		`DROP POLICY IF EXISTS user_sessions_policy ON sessions;
		CREATE POLICY user_sessions_policy ON sessions
			FOR ALL TO app_user
			USING (
				contact_id IN (
					SELECT c.id FROM contacts c
					JOIN projects p ON c.project_id = p.id
					WHERE p.user_id = current_setting('app.current_user_id', true)::uuid
				)
			);`,

		// Webhook Subscriptions: usuário só vê seus próprios webhooks
		`DROP POLICY IF EXISTS user_webhook_subscriptions_policy ON webhook_subscriptions;
		CREATE POLICY user_webhook_subscriptions_policy ON webhook_subscriptions
			FOR ALL TO app_user
			USING (user_id = current_setting('app.current_user_id', true)::uuid);`,

		// Channels: usuário só vê seus próprios canais
		`DROP POLICY IF EXISTS user_channels_policy ON channels;
		CREATE POLICY user_channels_policy ON channels
			FOR ALL TO app_user
			USING (user_id = current_setting('app.current_user_id', true)::uuid);`,
	}

	for _, policy := range policies {
		if err := db.Exec(policy).Error; err != nil {
			log.Printf("⚠️  Warning: Failed to create RLS policy: %v", err)
		}
	}

	// 6. Conceder permissões para app_user
	permissions := []string{
		"GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO app_user",
		"GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO app_user",
	}

	for _, perm := range permissions {
		if err := db.Exec(perm).Error; err != nil {
			log.Printf("⚠️  Warning: Failed to grant permissions: %v", err)
		}
	}

	log.Println("✅ Row Level Security (RLS) configured successfully!")
	log.Println("   📋 Tables with RLS enabled:")
	log.Println("      - projects, pipelines, pipeline_statuses")
	log.Println("      - contacts, contact_pipeline_statuses")
	log.Println("      - messages, sessions, webhook_subscriptions, channels")
	log.Println("   🔒 Policies created for user isolation")
	log.Println("   🔄 RLS será aplicado via GORM callbacks usando SET LOCAL")

	return nil
}
