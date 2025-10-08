-- Otimiza índice do channel_message_id para queries de deduplicação
-- Remove índice antigo se existir e cria um novo otimizado

-- Remove índice antigo (GORM cria com nome genérico)
DROP INDEX IF EXISTS idx_messages_channel_message_id;

-- Cria índice otimizado (não-único porque pode ser NULL)
-- Usa BTREE para busca rápida e WHERE para indexar apenas não-nulos
CREATE INDEX idx_messages_channel_message_id_lookup 
ON messages(channel_message_id) 
WHERE channel_message_id IS NOT NULL;

-- Adiciona índice composto para queries de ACK (channel_message_id + status)
CREATE INDEX idx_messages_channel_msg_status 
ON messages(channel_message_id, status) 
WHERE channel_message_id IS NOT NULL;
