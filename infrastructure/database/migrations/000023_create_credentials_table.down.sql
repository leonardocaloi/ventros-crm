-- Remove trigger
DROP TRIGGER IF EXISTS trigger_credentials_updated_at ON credentials;
DROP FUNCTION IF EXISTS update_credentials_updated_at();

-- Remove índices
DROP INDEX IF EXISTS idx_credentials_metadata;
DROP INDEX IF EXISTS idx_credentials_unique_name;
DROP INDEX IF EXISTS idx_credentials_expires_at;
DROP INDEX IF EXISTS idx_credentials_active;
DROP INDEX IF EXISTS idx_credentials_tenant_type;
DROP INDEX IF EXISTS idx_credentials_type;
DROP INDEX IF EXISTS idx_credentials_project;
DROP INDEX IF EXISTS idx_credentials_tenant;

-- Remove tabela
DROP TABLE IF EXISTS credentials;
