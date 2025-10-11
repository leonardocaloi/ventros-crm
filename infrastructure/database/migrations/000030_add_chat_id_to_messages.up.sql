-- Add chat_id to messages table to link messages to chats
ALTER TABLE messages ADD COLUMN IF NOT EXISTS chat_id UUID;

-- Create index for chat_id
CREATE INDEX IF NOT EXISTS idx_messages_chat_id ON messages(chat_id) WHERE deleted_at IS NULL;

-- Add foreign key constraint (but don't enforce it yet for existing data)
-- ALTER TABLE messages ADD CONSTRAINT fk_messages_chat FOREIGN KEY (chat_id) REFERENCES chats(id) ON DELETE CASCADE;

-- Note: After running this migration, you should:
-- 1. Create individual chats for all existing messages
-- 2. Update messages.chat_id to reference the appropriate chats
-- 3. Uncomment and run the foreign key constraint above
-- 4. Make chat_id NOT NULL after data migration is complete
