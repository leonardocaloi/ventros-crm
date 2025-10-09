-- Idempotency Pattern: Tabela para rastrear eventos já processados por cada consumer
-- Previne processamento duplicado de eventos no sistema event-driven
CREATE TABLE IF NOT EXISTS processed_events (
    id BIGSERIAL PRIMARY KEY,
    event_id UUID NOT NULL, -- ID do evento de domínio
    consumer_name VARCHAR(100) NOT NULL, -- Nome do consumer que processou (ContactEventConsumer, SessionWorker, etc)
    processed_at TIMESTAMP NOT NULL DEFAULT NOW(), -- Quando foi processado
    processing_duration_ms INT, -- Duração do processamento em ms (para métricas)

    -- Garante que cada consumer processa cada evento apenas uma vez
    CONSTRAINT uq_processed_event_consumer UNIQUE(event_id, consumer_name)
);

-- Índice para lookup rápido de idempotência
CREATE INDEX idx_processed_events_lookup ON processed_events(event_id, consumer_name);

-- Índice para cleanup de eventos antigos (opcional, para manutenção)
CREATE INDEX idx_processed_events_cleanup ON processed_events(processed_at);

-- Comentários para documentação
COMMENT ON TABLE processed_events IS 'Idempotency tracking: garante que cada evento é processado apenas uma vez por consumer';
COMMENT ON COLUMN processed_events.event_id IS 'ID único do evento de domínio (vem do outbox_events.event_id)';
COMMENT ON COLUMN processed_events.consumer_name IS 'Identificador do consumer (ex: ContactEventConsumer, WAHAMessageConsumer)';
COMMENT ON COLUMN processed_events.processing_duration_ms IS 'Duração do processamento para métricas de performance';
