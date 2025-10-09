-- Rollback: Remove automation_type field and restore pipeline_id NOT NULL constraint

-- Drop new indexes
DROP INDEX IF EXISTS idx_automation_type;
DROP INDEX IF EXISTS idx_automation_type_trigger;
DROP INDEX IF EXISTS idx_automation_tenant_type;

-- Restore NOT NULL constraint on pipeline_id
-- WARNING: This will fail if there are any rows with NULL pipeline_id
-- You may need to manually handle those rows before running this rollback
ALTER TABLE automation_rules
ALTER COLUMN pipeline_id SET NOT NULL;

-- Restore CASCADE delete behavior
ALTER TABLE automation_rules
DROP CONSTRAINT IF EXISTS automation_rules_pipeline_id_fkey;

ALTER TABLE automation_rules
ADD CONSTRAINT automation_rules_pipeline_id_fkey
FOREIGN KEY (pipeline_id)
REFERENCES pipelines(id)
ON DELETE CASCADE;

-- Remove automation_type column
ALTER TABLE automation_rules
DROP COLUMN IF EXISTS automation_type;

-- Restore original comments
COMMENT ON COLUMN automation_rules.pipeline_id IS 'Pipeline that owns this rule';
