-- Rollback product schemas
-- Warning: This will fail if there are tables in these schemas

DROP SCHEMA IF EXISTS ai CASCADE;
DROP SCHEMA IF EXISTS bi CASCADE;
DROP SCHEMA IF EXISTS workflows CASCADE;
DROP SCHEMA IF NOT EXISTS crm CASCADE;
DROP SCHEMA IF EXISTS shared CASCADE;

-- Reset search_path
ALTER DATABASE ventros_crm SET search_path TO public;
