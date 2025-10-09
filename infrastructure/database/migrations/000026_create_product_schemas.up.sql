-- Create product schemas for multi-product architecture
-- Ventros platform has multiple products: CRM, Workflows, BI, AI
-- Each product gets its own schema, with shared schema for common resources

-- 1. Create shared schema for auth, users, billing
CREATE SCHEMA IF NOT EXISTS shared;

-- 2. Create CRM schema (current tables will be moved here)
CREATE SCHEMA IF NOT EXISTS crm;

-- 3. Create future product schemas
CREATE SCHEMA IF NOT EXISTS workflows;
CREATE SCHEMA IF NOT EXISTS bi;
CREATE SCHEMA IF NOT EXISTS ai;

-- 4. Grant permissions (adjust as needed for your user)
GRANT ALL ON SCHEMA shared TO ventros;
GRANT ALL ON SCHEMA crm TO ventros;
GRANT ALL ON SCHEMA workflows TO ventros;
GRANT ALL ON SCHEMA bi TO ventros;
GRANT ALL ON SCHEMA ai TO ventros;

-- 5. Set search_path to include all schemas (for convenience)
-- This should also be set in application config
ALTER DATABASE ventros_crm SET search_path TO crm, shared, public;

-- Comments explaining the architecture
COMMENT ON SCHEMA shared IS 'Shared resources across all Ventros products (auth, users, billing, projects)';
COMMENT ON SCHEMA crm IS 'Ventros CRM product tables (contacts, sessions, messages, channels, pipelines)';
COMMENT ON SCHEMA workflows IS 'Ventros Workflows product tables (future)';
COMMENT ON SCHEMA bi IS 'Ventros BI product tables (future)';
COMMENT ON SCHEMA ai IS 'Ventros AI product tables (future)';

-- Note: Tables are currently in public schema
-- Future migrations can move specific tables to appropriate schemas:
-- - shared: users, projects, billing_accounts, customers
-- - crm: contacts, channels, sessions, messages, pipelines, etc
