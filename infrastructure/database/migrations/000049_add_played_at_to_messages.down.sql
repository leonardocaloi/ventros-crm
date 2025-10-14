-- Remove played_at column from messages table
DROP INDEX IF EXISTS idx_messages_played_at;
ALTER TABLE messages DROP COLUMN IF EXISTS played_at;
