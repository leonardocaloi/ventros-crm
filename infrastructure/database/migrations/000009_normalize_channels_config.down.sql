-- Rollback: Restaura colunas específicas do WAHA

-- 1. Adiciona colunas WAHA de volta
ALTER TABLE channels ADD COLUMN IF NOT EXISTS waha_base_url TEXT;
ALTER TABLE channels ADD COLUMN IF NOT EXISTS waha_token TEXT;
ALTER TABLE channels ADD COLUMN IF NOT EXISTS waha_session_id TEXT;
ALTER TABLE channels ADD COLUMN IF NOT EXISTS waha_webhook_url TEXT;

-- 2. Restaura dados do Config JSONB para colunas
UPDATE channels
SET 
    waha_base_url = config->>'base_url',
    waha_token = config->>'token',
    waha_session_id = config->>'session_id',
    waha_webhook_url = config->>'webhook_url'
WHERE type = 'waha' AND config IS NOT NULL;

-- 3. Remove índices criados
DROP INDEX IF EXISTS idx_channels_external_id;
DROP INDEX IF EXISTS idx_channels_config_gin;

-- 4. Remove comentários
COMMENT ON COLUMN channels.external_id IS NULL;
COMMENT ON COLUMN channels.config IS NULL;
