-- Migração para normalizar configurações dos canais
-- Remove colunas específicas do WAHA e move para Config JSONB

-- 1. Adiciona coluna external_id se não existir
ALTER TABLE channels ADD COLUMN IF NOT EXISTS external_id VARCHAR(255);

-- 2. Migra dados existentes de colunas WAHA para Config JSONB
UPDATE channels
SET config = jsonb_build_object(
    'base_url', COALESCE(waha_base_url, ''),
    'token', COALESCE(waha_token, ''),
    'session_id', COALESCE(waha_session_id, ''),
    'webhook_url', COALESCE(waha_webhook_url, '')
)
WHERE type = 'waha' AND (
    waha_base_url IS NOT NULL OR 
    waha_token IS NOT NULL OR 
    waha_session_id IS NOT NULL OR 
    waha_webhook_url IS NOT NULL
);

-- 3. Copia waha_session_id para external_id (se ainda não estiver preenchido)
UPDATE channels
SET external_id = waha_session_id
WHERE type = 'waha' 
  AND waha_session_id IS NOT NULL 
  AND (external_id IS NULL OR external_id = '');

-- 3. Remove colunas específicas do WAHA (agora no Config JSONB)
ALTER TABLE channels DROP COLUMN IF EXISTS waha_base_url;
ALTER TABLE channels DROP COLUMN IF EXISTS waha_token;
ALTER TABLE channels DROP COLUMN IF EXISTS waha_session_id;
ALTER TABLE channels DROP COLUMN IF EXISTS waha_webhook_url;

-- 4. Adiciona índice no external_id para buscas rápidas
CREATE INDEX IF NOT EXISTS idx_channels_external_id ON channels(external_id) WHERE external_id IS NOT NULL AND external_id != '';

-- 5. Adiciona índice GIN no Config JSONB para queries eficientes
CREATE INDEX IF NOT EXISTS idx_channels_config_gin ON channels USING GIN (config);

-- 6. Adiciona comentários para documentação
COMMENT ON COLUMN channels.external_id IS 'ID externo do canal na plataforma (session_id para WAHA, bot_id para Telegram, page_id para Messenger, etc)';
COMMENT ON COLUMN channels.config IS 'Configurações específicas de cada tipo de canal em formato JSONB';
