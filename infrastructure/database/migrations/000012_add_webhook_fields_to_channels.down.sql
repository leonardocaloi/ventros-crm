-- Remove webhook fields from channels table
DROP INDEX IF EXISTS idx_channels_webhook_url;

ALTER TABLE channels
DROP COLUMN IF EXISTS webhook_url,
DROP COLUMN IF EXISTS webhook_configured_at,
DROP COLUMN IF EXISTS webhook_active;
