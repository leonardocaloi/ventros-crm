-- Rollback: Rename automations back to automation_rules

-- Rename the table back
ALTER TABLE automations RENAME TO automation_rules;

-- Rename all indexes back
ALTER INDEX idx_automations_pipeline RENAME TO idx_automation_pipeline;
ALTER INDEX idx_automations_tenant RENAME TO idx_automation_tenant;
ALTER INDEX idx_automations_trigger RENAME TO idx_automation_trigger;
ALTER INDEX idx_automations_priority RENAME TO idx_automation_priority;
ALTER INDEX idx_automations_enabled RENAME TO idx_automation_enabled;
ALTER INDEX idx_automations_last_executed RENAME TO idx_automation_last_executed;
ALTER INDEX idx_automations_next_execution RENAME TO idx_automation_next_execution;
ALTER INDEX idx_automations_pipeline_trigger RENAME TO idx_automation_pipeline_trigger;
ALTER INDEX idx_automations_tenant_enabled RENAME TO idx_automation_tenant_enabled;
ALTER INDEX idx_automations_active RENAME TO idx_automation_active_rules;
ALTER INDEX idx_automations_scheduled_ready RENAME TO idx_automation_scheduled_ready;
ALTER INDEX idx_automations_type RENAME TO idx_automation_type;
ALTER INDEX idx_automations_type_trigger RENAME TO idx_automation_type_trigger;
ALTER INDEX idx_automations_tenant_type RENAME TO idx_automation_tenant_type;

-- Rename sequence back
ALTER SEQUENCE IF EXISTS automations_id_seq RENAME TO automation_rules_id_seq;

-- Rename foreign key constraint back
ALTER TABLE automation_rules
DROP CONSTRAINT IF EXISTS automations_pipeline_id_fkey;

ALTER TABLE automation_rules
ADD CONSTRAINT automation_rules_pipeline_id_fkey
FOREIGN KEY (pipeline_id)
REFERENCES pipelines(id)
ON DELETE SET NULL;

-- Rename trigger back
DROP TRIGGER IF EXISTS trigger_update_automations_updated_at ON automation_rules;

CREATE TRIGGER trigger_update_automation_rules_updated_at
    BEFORE UPDATE ON automation_rules
    FOR EACH ROW
    EXECUTE FUNCTION update_automation_rules_updated_at();

-- Restore original comment
COMMENT ON TABLE automation_rules IS 'Automatic follow-up rules associated with pipelines';
