-- Reverte otimização dos índices

DROP INDEX IF EXISTS idx_messages_channel_msg_status;
DROP INDEX IF EXISTS idx_messages_channel_message_id_lookup;

-- Recria índice simples do GORM
CREATE INDEX idx_messages_channel_message_id ON messages(channel_message_id);
