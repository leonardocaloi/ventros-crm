-- Add history import control fields to channels table
ALTER TABLE channels
ADD COLUMN history_import_enabled BOOLEAN DEFAULT false NOT NULL,
ADD COLUMN last_import_date TIMESTAMP,
ADD COLUMN history_import_status VARCHAR(50) DEFAULT 'idle',
ADD COLUMN history_import_stats JSONB,
ADD COLUMN default_agent_id UUID REFERENCES agents(id) ON DELETE SET NULL,
ADD COLUMN history_import_max_days INTEGER DEFAULT NULL,
ADD COLUMN history_import_max_messages_per_chat INTEGER DEFAULT NULL;

-- Add index for efficient querying of channels with history import enabled
CREATE INDEX idx_channels_history_import ON channels(history_import_enabled, history_import_status)
WHERE history_import_enabled = true;

-- Add index for default agent lookup
CREATE INDEX idx_channels_default_agent ON channels(default_agent_id)
WHERE default_agent_id IS NOT NULL;

-- Add comments for documentation
COMMENT ON COLUMN channels.history_import_enabled IS 'Flag to enable/disable history import for this channel';
COMMENT ON COLUMN channels.last_import_date IS 'Timestamp of the last successful import (for incremental sync)';
COMMENT ON COLUMN channels.history_import_status IS 'Current import status: idle, importing, completed, failed';
COMMENT ON COLUMN channels.history_import_stats IS 'JSON object with import statistics: {total, processed, failed, started_at, ended_at}';
COMMENT ON COLUMN channels.default_agent_id IS 'Default agent to assign imported messages to';
COMMENT ON COLUMN channels.history_import_max_days IS 'Maximum number of days to import history (NULL = unlimited, 7 = last week, 30 = last month, 90 = last 3 months)';
COMMENT ON COLUMN channels.history_import_max_messages_per_chat IS 'Maximum messages to import per chat (NULL = unlimited, prevents memory issues)';
