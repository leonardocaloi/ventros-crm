-- Remove allow_groups and tracking_enabled columns from channels table

DROP INDEX IF EXISTS idx_channels_tracking_enabled;
DROP INDEX IF EXISTS idx_channels_allow_groups;

ALTER TABLE channels
DROP COLUMN IF EXISTS tracking_enabled,
DROP COLUMN IF EXISTS allow_groups;
