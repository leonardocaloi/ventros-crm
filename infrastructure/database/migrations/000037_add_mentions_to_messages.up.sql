-- Adicionar campo de menções às mensagens
ALTER TABLE messages ADD COLUMN mentions TEXT[] DEFAULT '{}';

-- Índice GIN para buscas eficientes em menções (opcional, mas útil)
CREATE INDEX idx_messages_mentions ON messages USING GIN(mentions);
