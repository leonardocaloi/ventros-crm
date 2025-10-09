-- Remove timeout hierarchy
DROP INDEX IF EXISTS idx_pipelines_session_timeout;
DROP INDEX IF EXISTS idx_channels_session_timeout;

ALTER TABLE pipelines DROP COLUMN IF EXISTS session_timeout_minutes;
ALTER TABLE channels DROP COLUMN IF EXISTS session_timeout_minutes;
