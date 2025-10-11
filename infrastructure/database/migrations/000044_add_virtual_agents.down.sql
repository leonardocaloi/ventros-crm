-- Remove virtual_metadata column
DROP INDEX IF EXISTS idx_agents_virtual_metadata;
ALTER TABLE agents DROP COLUMN IF EXISTS virtual_metadata;
