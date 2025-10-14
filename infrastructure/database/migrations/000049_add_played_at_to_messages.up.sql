-- Add played_at column to messages table for tracking when media was played/viewed
ALTER TABLE messages ADD COLUMN IF NOT EXISTS played_at TIMESTAMP WITH TIME ZONE;

-- Add index for played_at queries
CREATE INDEX IF NOT EXISTS idx_messages_played_at ON messages(played_at);

-- Add comment for documentation
COMMENT ON COLUMN messages.played_at IS 'Timestamp when media message was played/viewed (ACK 4 from WhatsApp)';
