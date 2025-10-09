-- Migration: Add automation_type field and make pipeline_id nullable
-- This enables generic automations beyond pipeline-specific ones

-- Add automation_type column with default value for existing rows
ALTER TABLE automation_rules
ADD COLUMN IF NOT EXISTS automation_type VARCHAR(100) NOT NULL DEFAULT 'pipeline_automation';

-- Create index for automation_type
CREATE INDEX IF NOT EXISTS idx_automation_type ON automation_rules(automation_type);

-- Make pipeline_id nullable (drop NOT NULL constraint)
-- First, check if any rules don't have a pipeline_id and set a default if needed
-- This ensures data integrity during migration
ALTER TABLE automation_rules
ALTER COLUMN pipeline_id DROP NOT NULL;

-- Update the foreign key constraint to SET NULL on delete instead of CASCADE
-- First drop the existing constraint
ALTER TABLE automation_rules
DROP CONSTRAINT IF EXISTS automation_rules_pipeline_id_fkey;

-- Recreate with SET NULL
ALTER TABLE automation_rules
ADD CONSTRAINT automation_rules_pipeline_id_fkey
FOREIGN KEY (pipeline_id)
REFERENCES pipelines(id)
ON DELETE SET NULL;

-- Add composite index for generic automations (type + trigger + enabled)
CREATE INDEX IF NOT EXISTS idx_automation_type_trigger
ON automation_rules(automation_type, trigger)
WHERE enabled = true;

-- Add composite index for tenant-scoped generic automations
CREATE INDEX IF NOT EXISTS idx_automation_tenant_type
ON automation_rules(tenant_id, automation_type, enabled);

-- Update comments
COMMENT ON COLUMN automation_rules.automation_type IS 'Type of automation: pipeline_automation, scheduled_report, time_based_notification, webhook_automation, custom';
COMMENT ON COLUMN automation_rules.pipeline_id IS 'Pipeline ID (nullable - only required for pipeline automations)';
