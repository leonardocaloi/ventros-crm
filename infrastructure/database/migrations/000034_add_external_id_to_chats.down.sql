-- Remove external_id from chats table

ALTER TABLE chats
DROP CONSTRAINT IF EXISTS uq_chats_external_id;

DROP INDEX IF EXISTS idx_chats_external_id;

ALTER TABLE chats
DROP COLUMN IF EXISTS external_id;
