-- ========================================
-- Migration 000052: Add UNIQUE Constraint for Message Deduplication
-- ========================================
-- Context: Prevents duplicate messages during history import vs webhook concurrency
--
-- SAGA Pattern Compensation:
-- - History import sets channel.history_import_status = 'importing'
-- - Webhooks arriving during import are buffered in RabbitMQ
-- - After import, buffered webhooks are processed
-- - This UNIQUE constraint ensures no duplicates if race condition occurs
--
-- Idempotency Key: (channel_id, channel_message_id)
-- ========================================

-- Remove NULL values first (messages without channel_message_id)
UPDATE messages
SET channel_message_id = id::text
WHERE channel_message_id IS NULL;

-- Add UNIQUE INDEX for deduplication (partial index)
-- PostgreSQL partial indexes cannot be used with UNIQUE constraints,
-- so we use a unique index directly (functionally equivalent)
-- Note: Not using CONCURRENTLY as it cannot run inside transaction block
CREATE UNIQUE INDEX IF NOT EXISTS idx_messages_unique_channel_msg_id
ON messages (channel_id, channel_message_id)
WHERE channel_message_id IS NOT NULL AND deleted_at IS NULL;

-- Add comment explaining idempotency strategy
COMMENT ON INDEX idx_messages_unique_channel_msg_id IS
'Ensures message idempotency: prevents duplicates from concurrent import + webhooks. Part of SAGA compensation pattern for history import.';
