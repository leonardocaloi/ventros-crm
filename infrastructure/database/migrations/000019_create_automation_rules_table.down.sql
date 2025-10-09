-- Drop trigger
DROP TRIGGER IF EXISTS trigger_update_automation_rules_updated_at ON automation_rules;
DROP FUNCTION IF EXISTS update_automation_rules_updated_at();

-- Drop indexes
DROP INDEX IF EXISTS idx_automation_scheduled_ready;
DROP INDEX IF EXISTS idx_automation_active_rules;
DROP INDEX IF EXISTS idx_automation_tenant_enabled;
DROP INDEX IF EXISTS idx_automation_pipeline_trigger;
DROP INDEX IF EXISTS idx_automation_next_execution;
DROP INDEX IF EXISTS idx_automation_last_executed;
DROP INDEX IF EXISTS idx_automation_enabled;
DROP INDEX IF EXISTS idx_automation_priority;
DROP INDEX IF EXISTS idx_automation_trigger;
DROP INDEX IF EXISTS idx_automation_tenant;
DROP INDEX IF EXISTS idx_automation_pipeline;

-- Drop table
DROP TABLE IF EXISTS automation_rules;
