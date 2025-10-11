-- Remover campo de menções
DROP INDEX IF EXISTS idx_messages_mentions;
ALTER TABLE messages DROP COLUMN IF EXISTS mentions;
