-- Remove pipeline association from channels
DROP INDEX IF EXISTS idx_channels_pipeline_id;

ALTER TABLE channels
DROP COLUMN IF EXISTS pipeline_id,
DROP COLUMN IF EXISTS default_session_timeout_minutes;
