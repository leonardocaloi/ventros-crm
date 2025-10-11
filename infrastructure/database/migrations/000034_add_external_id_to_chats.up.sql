-- Add external_id to chats table for storing external identifiers
-- (WhatsApp group ID @g.us, Telegram group ID, etc)

ALTER TABLE chats
ADD COLUMN external_id TEXT;

-- Create index for fast lookups by external_id
CREATE INDEX idx_chats_external_id ON chats(external_id) WHERE external_id IS NOT NULL;

-- Add unique constraint to prevent duplicate external_ids
ALTER TABLE chats
ADD CONSTRAINT uq_chats_external_id UNIQUE (external_id);

COMMENT ON COLUMN chats.external_id IS 'External identifier from the channel (WhatsApp group ID @g.us, Telegram group ID, etc)';
