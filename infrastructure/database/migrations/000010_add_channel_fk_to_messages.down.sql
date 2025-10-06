-- Rollback: Remove Foreign Key de messages → channels

-- 1. Remove Foreign Key
ALTER TABLE messages DROP CONSTRAINT IF EXISTS fk_messages_channel;

-- 2. Remove comentário
COMMENT ON COLUMN messages.channel_id IS NULL;

-- 3. Mantém índice por performance (não remove)
-- DROP INDEX IF EXISTS idx_messages_channel_id;
