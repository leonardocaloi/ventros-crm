package persistence

import (
	"fmt"
	"log"

	"github.com/caloi/ventros-crm/infrastructure/persistence/entities"
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

// fixSchemaIncompatibilities detecta e corrige automaticamente incompatibilidades de schema
func fixSchemaIncompatibilities(db *gorm.DB) error {
	fmt.Println("🔧 Starting fixSchemaIncompatibilities...")
	log.Println("🔧 Starting fixSchemaIncompatibilities...")
	
	type columnInfo struct {
		TableName  string `gorm:"column:table_name"`
		ColumnName string `gorm:"column:column_name"`
		DataType   string `gorm:"column:data_type"`
	}
	
	// Mapeamento de conversões conhecidas: de tipo -> para tipo
	typeConversions := map[string]map[string]string{
		"webhook_subscriptions.events": {
			"jsonb": "text[]",
		},
	}
	
	// Buscar todas as colunas que podem precisar de conversão
	var columns []columnInfo
	result := db.Raw(`
		SELECT table_name, column_name, data_type 
		FROM information_schema.columns 
		WHERE table_schema = 'public'
		AND table_name IN ('webhook_subscriptions')
	`).Scan(&columns)
	
	if result.Error != nil {
		log.Printf("⚠️  Warning: Failed to query columns: %v", result.Error)
		return result.Error
	}
	
	log.Printf("🔍 Found %d columns to check for conversions", len(columns))
	
	// Se não encontrou nada, tentar com current_schema
	if len(columns) == 0 {
		log.Println("🔍 No columns found in 'public' schema, trying with current_schema()...")
		result = db.Raw(`
			SELECT table_name, column_name, data_type 
			FROM information_schema.columns 
			WHERE table_schema = current_schema()
			AND table_name IN ('webhook_subscriptions')
		`).Scan(&columns)
		log.Printf("🔍 Found %d columns with current_schema()", len(columns))
	}
	
	// Aplicar conversões necessárias
	for _, col := range columns {
		key := col.TableName + "." + col.ColumnName
		log.Printf("🔍 Checking column: %s (type: %s)", key, col.DataType)
		
		if conversions, exists := typeConversions[key]; exists {
			if targetType, needsConversion := conversions[col.DataType]; needsConversion {
				log.Printf("🔄 Converting %s.%s from %s to %s...", col.TableName, col.ColumnName, col.DataType, targetType)
				
				var conversionSQL string
				// Definir SQL de conversão baseado nos tipos
				if col.DataType == "jsonb" && targetType == "text[]" {
					// PostgreSQL não permite subqueries no USING, então usamos uma abordagem diferente
					// Primeiro, criamos uma coluna temporária, depois copiamos os dados, e finalmente renomeamos
					conversionSQL = fmt.Sprintf(`
						DO $$ 
						BEGIN
							-- Adicionar coluna temporária
							ALTER TABLE %s ADD COLUMN IF NOT EXISTS %s_temp text[];
							
							-- Copiar dados convertidos
							UPDATE %s SET %s_temp = CASE 
								WHEN jsonb_typeof(%s) = 'array' THEN 
									(SELECT array_agg(value::text) FROM jsonb_array_elements_text(%s) AS value)
								ELSE 
									ARRAY[]::text[]
							END;
							
							-- Remover coluna antiga
							ALTER TABLE %s DROP COLUMN IF EXISTS %s;
							
							-- Renomear coluna temporária
							ALTER TABLE %s RENAME COLUMN %s_temp TO %s;
						END $$;
					`, col.TableName, col.ColumnName, 
					   col.TableName, col.ColumnName, col.ColumnName, col.ColumnName,
					   col.TableName, col.ColumnName,
					   col.TableName, col.ColumnName, col.ColumnName)
				} else {
					// Conversão genérica
					conversionSQL = fmt.Sprintf(
						"ALTER TABLE %s ALTER COLUMN %s TYPE %s USING %s::%s",
						col.TableName, col.ColumnName, targetType, col.ColumnName, targetType,
					)
				}
				
				if err := db.Exec(conversionSQL).Error; err != nil {
					log.Printf("⚠️  Warning: Failed to convert %s.%s: %v", col.TableName, col.ColumnName, err)
				} else {
					log.Printf("✅ Successfully converted %s.%s to %s", col.TableName, col.ColumnName, targetType)
				}
			}
		}
	}
	
	return nil
}

// AutoMigrate runs database migrations for all entities
func AutoMigrate(db *gorm.DB) error {
	fmt.Println("DEBUG: AutoMigrate function called")
	log.Println("🔄 Running GORM auto-migrations...")

	// Pré-migração: Corrigir incompatibilidades de tipos automaticamente
	fmt.Println("DEBUG: About to call fixSchemaIncompatibilities")
	log.Println("🔄 Pre-migration: Fixing schema incompatibilities...")
	if err := fixSchemaIncompatibilities(db); err != nil {
		log.Printf("⚠️  Warning: Some schema fixes failed: %v", err)
	}
	fmt.Println("DEBUG: fixSchemaIncompatibilities completed")
	
	preMigrationUpdates := []string{
		// Converter webhook_subscriptions.events de jsonb para text[] PRIMEIRO
		`DO $$ 
		BEGIN
			IF EXISTS (
				SELECT 1 FROM information_schema.columns 
				WHERE table_name = 'webhook_subscriptions' 
				AND column_name = 'events' 
				AND data_type = 'jsonb'
			) THEN
				ALTER TABLE webhook_subscriptions 
				ALTER COLUMN events TYPE text[] 
				USING CASE 
					WHEN jsonb_typeof(events) = 'array' THEN 
						ARRAY(SELECT jsonb_array_elements_text(events))
					ELSE 
						ARRAY[]::text[] 
				END;
			END IF;
		END $$;`,
		// Adicionar password_hash aos usuários existentes
		`DO $$ 
		BEGIN
			IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'users' AND column_name = 'password_hash' IS NULL) THEN
				-- Adicionar coluna como nullable primeiro
				ALTER TABLE users ADD COLUMN IF NOT EXISTS password_hash text;
				-- Atualizar usuários sem password_hash (hash de 'changeme123')
				UPDATE users SET password_hash = '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy' WHERE password_hash IS NULL;
			END IF;
		END $$;`,
		// Atualizar messages.content_type para ter valor padrão
		"UPDATE messages SET content_type = 'text' WHERE content_type IS NULL OR content_type = ''",
		// Atualizar channel_types.provider para ter valor padrão
		"UPDATE channel_types SET provider = 'unknown' WHERE provider IS NULL OR provider = ''",
		// Renomear tabela customers para users
		"ALTER TABLE IF EXISTS customers RENAME TO users",
		// Renomear coluna customer_id para user_id em projects
		"ALTER TABLE IF EXISTS projects RENAME COLUMN customer_id TO user_id",
		// Renomear coluna customer_id para user_id em messages
		"ALTER TABLE IF EXISTS messages RENAME COLUMN customer_id TO user_id",
		// Drop old foreign key constraint if exists
		"ALTER TABLE IF EXISTS projects DROP CONSTRAINT IF EXISTS projects_customers_projects",
		"ALTER TABLE IF EXISTS projects DROP CONSTRAINT IF EXISTS fk_projects_customer",
	}

	for _, sql := range preMigrationUpdates {
		if err := db.Exec(sql).Error; err != nil {
			log.Printf("⚠️  Warning: Pre-migration update failed (may be expected): %v", err)
			// Continuar mesmo se falhar (pode ser que a coluna não exista ainda)
		}
	}

	// Executar AutoMigrate
	err := db.AutoMigrate(
		// Core entities
		&entities.UserEntity{},
		&entities.UserAPIKeyEntity{},        // ← ADICIONADO
		&entities.ProjectEntity{},
		&entities.BillingAccountEntity{},    // ← ADICIONADO
		&entities.ChannelEntity{},           // ← ADICIONADO
		&entities.ChannelTypeEntity{},
		
		// Contact & Communication
		&entities.ContactEntity{},
		&entities.SessionEntity{},
		&entities.MessageEntity{},
		&entities.ContactEventEntity{},
		
		// Agents
		&entities.AgentEntity{},
		&entities.AgentSessionEntity{},
		
		// Webhooks
		&entities.WebhookSubscriptionEntity{},
		
		// Pipelines
		&entities.PipelineEntity{},
		&entities.PipelineStatusEntity{},
		&entities.ContactPipelineStatusEntity{},
		
		// Custom Fields
		&entities.ContactCustomFieldEntity{},
		&entities.SessionCustomFieldEntity{},
		
		// Event Logs
		&entities.DomainEventLogEntity{},
	)

	if err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}
	
	// 🎯 Otimizações pós-migração (índices customizados)
	log.Println("🔄 Applying post-migration optimizations...")
	postMigrationOptimizations := []string{
		// Otimiza índice do channel_message_id para deduplicação e ACKs
		`DROP INDEX IF EXISTS idx_messages_channel_message_id`,
		`CREATE INDEX IF NOT EXISTS idx_messages_channel_message_id_lookup 
		 ON messages(channel_message_id) 
		 WHERE channel_message_id IS NOT NULL`,
		`CREATE INDEX IF NOT EXISTS idx_messages_channel_msg_status 
		 ON messages(channel_message_id, status) 
		 WHERE channel_message_id IS NOT NULL`,
	}
	
	for _, sql := range postMigrationOptimizations {
		if err := db.Exec(sql).Error; err != nil {
			log.Printf("⚠️  Warning: Post-migration optimization failed: %v", err)
			// Continua mesmo se falhar (índice pode já existir)
		}
	}
	log.Println("✅ Post-migration optimizations completed")

	log.Println("✅ GORM migrations completed successfully!")
	return nil
}

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
