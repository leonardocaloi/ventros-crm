-- Torna channel_id obrigatório em messages
-- Uma mensagem SEMPRE vem de um canal

-- 1. Deleta mensagens órfãs (se existirem)
DELETE FROM messages WHERE channel_id IS NULL;

-- 2. Torna coluna NOT NULL
ALTER TABLE messages ALTER COLUMN channel_id SET NOT NULL;

-- 3. Remove FK antiga (SET NULL) e cria nova (RESTRICT)
ALTER TABLE messages DROP CONSTRAINT IF EXISTS fk_messages_channel;

ALTER TABLE messages
ADD CONSTRAINT fk_messages_channel 
FOREIGN KEY (channel_id) 
REFERENCES channels(id) 
ON DELETE RESTRICT;  -- NÃO permite deletar canal com mensagens!

-- 4. Atualiza comentário
COMMENT ON COLUMN messages.channel_id IS 'ID do canal por onde a mensagem foi recebida/enviada. OBRIGATÓRIO - toda mensagem vem de um canal. FK com ON DELETE RESTRICT impede exclusão de canais com mensagens.';
