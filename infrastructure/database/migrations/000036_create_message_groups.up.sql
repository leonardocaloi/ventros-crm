-- Tabela de grupos de mensagens para debouncer
CREATE TABLE message_groups (
    id UUID PRIMARY KEY,
    contact_id UUID NOT NULL,
    channel_id UUID NOT NULL,
    session_id UUID NOT NULL,
    tenant_id VARCHAR(255) NOT NULL,
    message_ids TEXT[] NOT NULL DEFAULT '{}',
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    started_at TIMESTAMP NOT NULL,
    completed_at TIMESTAMP,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Índices para performance
CREATE INDEX idx_message_groups_contact_channel ON message_groups(contact_id, channel_id);
CREATE INDEX idx_message_groups_session ON message_groups(session_id);
CREATE INDEX idx_message_groups_tenant ON message_groups(tenant_id);
CREATE INDEX idx_message_groups_status ON message_groups(status);
CREATE INDEX idx_message_groups_expires_at ON message_groups(expires_at) WHERE status = 'pending';

-- Tabela de mensagens enriquecidas (partes individuais processadas pela IA)
CREATE TABLE message_enrichments (
    id UUID PRIMARY KEY,
    message_id UUID NOT NULL REFERENCES messages(id) ON DELETE CASCADE,
    group_id UUID REFERENCES message_groups(id) ON DELETE SET NULL,
    tenant_id VARCHAR(255) NOT NULL,
    content_type VARCHAR(50) NOT NULL, -- text, audio, image, video, document, voice
    original_content TEXT, -- Conteúdo original (texto ou URL de mídia)
    enriched_content TEXT, -- Conteúdo enriquecido pela IA (transcrição, OCR, etc)
    provider VARCHAR(50), -- openai, anthropic, google, deepgram, llamaparse
    model VARCHAR(100),
    processing_status VARCHAR(50) NOT NULL DEFAULT 'pending', -- pending, processing, completed, failed
    processing_started_at TIMESTAMP,
    processing_completed_at TIMESTAMP,
    error_message TEXT,
    metadata JSONB, -- Metadados extras (confidence, language, etc)
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Índices para message_enrichments
CREATE INDEX idx_message_enrichments_message_id ON message_enrichments(message_id);
CREATE INDEX idx_message_enrichments_group_id ON message_enrichments(group_id);
CREATE INDEX idx_message_enrichments_tenant ON message_enrichments(tenant_id);
CREATE INDEX idx_message_enrichments_status ON message_enrichments(processing_status);

-- Tabela de histórico de envio para AI Agent (resultado concatenado após debouncer)
CREATE TABLE ai_agent_history (
    id UUID PRIMARY KEY,
    group_id UUID NOT NULL REFERENCES message_groups(id) ON DELETE CASCADE,
    session_id UUID NOT NULL,
    contact_id UUID NOT NULL,
    channel_id UUID NOT NULL,
    tenant_id VARCHAR(255) NOT NULL,
    concatenated_content TEXT NOT NULL, -- Texto concatenado: msgs originais + enriquecimentos
    message_count INT NOT NULL, -- Quantas mensagens foram agrupadas
    enrichment_count INT NOT NULL, -- Quantos enriquecimentos foram incluídos
    sent_to_ai BOOLEAN NOT NULL DEFAULT false,
    ai_response TEXT, -- Resposta do AI Agent
    ai_provider VARCHAR(50),
    ai_model VARCHAR(100),
    processing_time_ms INT, -- Tempo de processamento em ms
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    sent_at TIMESTAMP,
    response_received_at TIMESTAMP
);

-- Índices para ai_agent_history
CREATE INDEX idx_ai_agent_history_group_id ON ai_agent_history(group_id);
CREATE INDEX idx_ai_agent_history_session_id ON ai_agent_history(session_id);
CREATE INDEX idx_ai_agent_history_contact ON ai_agent_history(contact_id);
CREATE INDEX idx_ai_agent_history_tenant ON ai_agent_history(tenant_id);
CREATE INDEX idx_ai_agent_history_sent_to_ai ON ai_agent_history(sent_to_ai);
