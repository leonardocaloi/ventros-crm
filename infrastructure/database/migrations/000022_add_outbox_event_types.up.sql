-- Adiciona categoria de evento à tabela outbox_events
-- Permite classificar eventos em: domain_event, message_delivery, webhook, meta_conversion, google_ads

ALTER TABLE outbox_events
ADD COLUMN IF NOT EXISTS event_category VARCHAR(50) NOT NULL DEFAULT 'domain_event';

-- Cria índice para melhorar performance em queries por categoria
CREATE INDEX IF NOT EXISTS idx_outbox_events_category ON outbox_events(event_category);

-- Cria índice composto para queries comuns (categoria + status + tenant)
CREATE INDEX IF NOT EXISTS idx_outbox_events_category_status_tenant
ON outbox_events(event_category, status, tenant_id);

-- Comentários sobre as categorias
COMMENT ON COLUMN outbox_events.event_category IS 'Categoria do evento: domain_event, message_delivery, webhook, meta_conversion, google_ads';
