-- Rollback: Torna channel_id opcional novamente

-- 1. Remove FK RESTRICT
ALTER TABLE messages DROP CONSTRAINT IF EXISTS fk_messages_channel;

-- 2. Torna coluna nullable
ALTER TABLE messages ALTER COLUMN channel_id DROP NOT NULL;

-- 3. Recria FK com SET NULL
ALTER TABLE messages
ADD CONSTRAINT fk_messages_channel 
FOREIGN KEY (channel_id) 
REFERENCES channels(id) 
ON DELETE SET NULL;

-- 4. Atualiza comentário
COMMENT ON COLUMN messages.channel_id IS 'ID do canal por onde a mensagem foi recebida/enviada. Pode ser NULL para mensagens antigas.';
