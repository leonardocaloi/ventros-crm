-- Remove debounce_timeout_ms column from channels table

DROP INDEX IF EXISTS idx_channels_debounce_timeout;

ALTER TABLE channels
DROP COLUMN IF EXISTS debounce_timeout_ms;
