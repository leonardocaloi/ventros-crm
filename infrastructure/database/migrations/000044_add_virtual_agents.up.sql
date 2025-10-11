-- Add virtual_metadata JSONB column for virtual agents
ALTER TABLE agents ADD COLUMN IF NOT EXISTS virtual_metadata JSONB;

-- Add index for virtual_metadata GIN operations
CREATE INDEX IF NOT EXISTS idx_agents_virtual_metadata ON agents USING gin (virtual_metadata);

-- Add comment to explain virtual agents
COMMENT ON COLUMN agents.virtual_metadata IS 'Metadata for virtual agents representing historical users. Contains: represents_person_name, period_start, period_end, reason, source_device, notes';
