-- Drop indexes
DROP INDEX IF EXISTS idx_channels_history_import;
DROP INDEX IF EXISTS idx_channels_default_agent;

-- Remove history import fields from channels table
ALTER TABLE channels
DROP COLUMN IF EXISTS history_import_enabled,
DROP COLUMN IF EXISTS last_import_date,
DROP COLUMN IF EXISTS history_import_status,
DROP COLUMN IF EXISTS history_import_stats,
DROP COLUMN IF EXISTS default_agent_id,
DROP COLUMN IF EXISTS history_import_max_days,
DROP COLUMN IF EXISTS history_import_max_messages_per_chat;
