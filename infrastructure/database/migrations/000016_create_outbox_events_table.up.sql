-- Outbox Pattern: Tabela para armazenar eventos de domínio antes de publicá-los
-- Garante consistência transacional entre mudanças de estado e publicação de eventos
CREATE TABLE IF NOT EXISTS outbox_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id UUID NOT NULL UNIQUE, -- ID único do evento de domínio (para deduplicação)
    aggregate_id UUID NOT NULL, -- ID do agregado que gerou o evento
    aggregate_type VARCHAR(100) NOT NULL, -- Tipo do agregado (Contact, Session, Message, etc)
    event_type VARCHAR(100) NOT NULL, -- Tipo do evento (contact.created, session.started, etc)
    event_version VARCHAR(20) NOT NULL DEFAULT 'v1', -- Versão do schema do evento
    event_data JSONB NOT NULL, -- Payload completo do evento
    tenant_id VARCHAR(100), -- Tenant ID para multi-tenancy
    project_id UUID, -- Project ID para filtros
    created_at TIMESTAMP NOT NULL DEFAULT NOW(), -- Quando o evento foi gerado
    processed_at TIMESTAMP, -- Quando foi publicado no RabbitMQ
    status VARCHAR(20) NOT NULL DEFAULT 'pending', -- pending | processing | processed | failed
    retry_count INT NOT NULL DEFAULT 0, -- Número de tentativas de processamento
    last_error TEXT, -- Última mensagem de erro (se houver)
    last_retry_at TIMESTAMP, -- Quando foi a última tentativa

    -- Constraints
    CONSTRAINT chk_outbox_status CHECK (status IN ('pending', 'processing', 'processed', 'failed'))
);

-- Índices para performance
CREATE INDEX idx_outbox_status_created ON outbox_events(status, created_at) WHERE status IN ('pending', 'processing');
CREATE INDEX idx_outbox_aggregate ON outbox_events(aggregate_type, aggregate_id);
CREATE INDEX idx_outbox_tenant ON outbox_events(tenant_id) WHERE tenant_id IS NOT NULL;
CREATE INDEX idx_outbox_event_type ON outbox_events(event_type);

-- Índice para retry exponencial
CREATE INDEX idx_outbox_retry ON outbox_events(status, retry_count, last_retry_at)
    WHERE status = 'failed' AND retry_count < 5;

-- Comentários para documentação
COMMENT ON TABLE outbox_events IS 'Transactional Outbox Pattern: eventos de domínio aguardando publicação no RabbitMQ';
COMMENT ON COLUMN outbox_events.event_id IS 'ID único do evento para deduplicação e rastreabilidade';
COMMENT ON COLUMN outbox_events.event_version IS 'Versão do schema para schema evolution';
COMMENT ON COLUMN outbox_events.status IS 'pending: aguardando processamento | processing: sendo processado | processed: publicado com sucesso | failed: erro após retries';
