-- Rollback: Remove connection_mode column from channels table

-- 1. Drop the index
DROP INDEX IF EXISTS idx_channels_connection_mode;

-- 2. Drop the column
ALTER TABLE channels
DROP COLUMN IF EXISTS connection_mode;
