-- Adiciona Foreign Key de messages.channel_id → channels.id
-- E garante integridade referencial

-- 1. Adiciona índice no channel_id se não existir
CREATE INDEX IF NOT EXISTS idx_messages_channel_id ON messages(channel_id) WHERE channel_id IS NOT NULL;

-- 2. Adiciona Foreign Key com ON DELETE SET NULL
-- (Se canal for deletado, mensagens ficam mas channel_id vira NULL)
ALTER TABLE messages
ADD CONSTRAINT fk_messages_channel 
FOREIGN KEY (channel_id) 
REFERENCES channels(id) 
ON DELETE SET NULL;

-- 3. Adiciona comentário para documentação
COMMENT ON COLUMN messages.channel_id IS 'ID do canal por onde a mensagem foi recebida/enviada. Pode ser NULL para mensagens antigas ou se canal for deletado.';
