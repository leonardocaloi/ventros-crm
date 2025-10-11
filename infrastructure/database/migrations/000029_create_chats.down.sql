-- Drop indexes first
DROP INDEX IF EXISTS idx_chats_subject;
DROP INDEX IF EXISTS idx_chats_metadata;
DROP INDEX IF EXISTS idx_chats_participants;
DROP INDEX IF EXISTS idx_chats_deleted;
DROP INDEX IF EXISTS idx_chats_updated;
DROP INDEX IF EXISTS idx_chats_created;
DROP INDEX IF EXISTS idx_chats_last_message;
DROP INDEX IF EXISTS idx_chats_status;
DROP INDEX IF EXISTS idx_chats_type;
DROP INDEX IF EXISTS idx_chats_tenant_type;
DROP INDEX IF EXISTS idx_chats_tenant_status;
DROP INDEX IF EXISTS idx_chats_tenant;
DROP INDEX IF EXISTS idx_chats_project;

-- Drop table
DROP TABLE IF EXISTS chats;
