-- Remove foreign key constraint if it exists
ALTER TABLE messages DROP CONSTRAINT IF EXISTS fk_messages_chat;

-- Drop index
DROP INDEX IF EXISTS idx_messages_chat_id;

-- Remove chat_id column
ALTER TABLE messages DROP COLUMN IF EXISTS chat_id;
