-- Migration: Rename automation_rules to automations (better naming for generic automations)

-- Rename the table
ALTER TABLE automation_rules RENAME TO automations;

-- Rename all indexes
ALTER INDEX idx_automation_pipeline RENAME TO idx_automations_pipeline;
ALTER INDEX idx_automation_tenant RENAME TO idx_automations_tenant;
ALTER INDEX idx_automation_trigger RENAME TO idx_automations_trigger;
ALTER INDEX idx_automation_priority RENAME TO idx_automations_priority;
ALTER INDEX idx_automation_enabled RENAME TO idx_automations_enabled;
ALTER INDEX idx_automation_last_executed RENAME TO idx_automations_last_executed;
ALTER INDEX idx_automation_next_execution RENAME TO idx_automations_next_execution;
ALTER INDEX idx_automation_pipeline_trigger RENAME TO idx_automations_pipeline_trigger;
ALTER INDEX idx_automation_tenant_enabled RENAME TO idx_automations_tenant_enabled;
ALTER INDEX idx_automation_active_rules RENAME TO idx_automations_active;
ALTER INDEX idx_automation_scheduled_ready RENAME TO idx_automations_scheduled_ready;
ALTER INDEX idx_automation_type RENAME TO idx_automations_type;
ALTER INDEX idx_automation_type_trigger RENAME TO idx_automations_type_trigger;
ALTER INDEX idx_automation_tenant_type RENAME TO idx_automations_tenant_type;

-- Rename sequence
ALTER SEQUENCE IF EXISTS automation_rules_id_seq RENAME TO automations_id_seq;

-- Rename foreign key constraint
ALTER TABLE automations
DROP CONSTRAINT IF EXISTS automation_rules_pipeline_id_fkey;

ALTER TABLE automations
ADD CONSTRAINT automations_pipeline_id_fkey
FOREIGN KEY (pipeline_id)
REFERENCES pipelines(id)
ON DELETE SET NULL;

-- Update trigger function name
DROP TRIGGER IF EXISTS trigger_update_automation_rules_updated_at ON automations;

CREATE TRIGGER trigger_update_automations_updated_at
    BEFORE UPDATE ON automations
    FOR EACH ROW
    EXECUTE FUNCTION update_automation_rules_updated_at();

-- Update comments
COMMENT ON TABLE automations IS 'Generic automation rules (pipeline, scheduled reports, notifications, webhooks, etc)';
