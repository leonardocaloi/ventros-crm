-- ========================================
-- Migration 000052 Rollback: Remove UNIQUE Index
-- ========================================
-- Removes the deduplication index added for import/webhook concurrency

-- Drop index (no constraint to drop since we use index directly)
-- Note: Not using CONCURRENTLY as it cannot run inside transaction block
DROP INDEX IF EXISTS idx_messages_unique_channel_msg_id;
