-- Create message_enrichments table
CREATE TABLE IF NOT EXISTS message_enrichments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    message_id UUID NOT NULL,
    message_group_id UUID NOT NULL,
    content_type VARCHAR(50) NOT NULL,
    provider VARCHAR(50) NOT NULL,
    media_url TEXT NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    extracted_text TEXT,
    metadata JSONB,
    processing_time_ms INTEGER,
    error TEXT,
    context VARCHAR(50),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    processed_at TIMESTAMP,

    -- Foreign keys
    CONSTRAINT fk_enrichments_message FOREIGN KEY (message_id) REFERENCES messages(id) ON DELETE CASCADE,
    CONSTRAINT fk_enrichments_message_group FOREIGN KEY (message_group_id) REFERENCES message_groups(id) ON DELETE CASCADE,

    -- Check constraints
    CONSTRAINT chk_enrichments_content_type CHECK (content_type IN ('audio', 'voice', 'image', 'video', 'document')),
    CONSTRAINT chk_enrichments_provider CHECK (provider IN ('whisper', 'deepgram', 'vision', 'llamaparse', 'ffmpeg', 'tesseract')),
    CONSTRAINT chk_enrichments_status CHECK (status IN ('pending', 'processing', 'completed', 'failed'))
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_enrichments_message
    ON message_enrichments(message_id);

CREATE INDEX IF NOT EXISTS idx_enrichments_group
    ON message_enrichments(message_group_id);

CREATE INDEX IF NOT EXISTS idx_enrichments_status
    ON message_enrichments(status)
    WHERE status IN ('pending', 'processing');

CREATE INDEX IF NOT EXISTS idx_enrichments_content_type
    ON message_enrichments(content_type);

CREATE INDEX IF NOT EXISTS idx_enrichments_created
    ON message_enrichments(created_at DESC);

-- Composite index for priority queue (used by FindPending)
CREATE INDEX IF NOT EXISTS idx_enrichments_pending_priority
    ON message_enrichments(
        CASE content_type
            WHEN 'voice' THEN 10
            WHEN 'audio' THEN 8
            WHEN 'image' THEN 7
            WHEN 'document' THEN 6
            WHEN 'video' THEN 3
            ELSE 5
        END DESC,
        created_at ASC
    )
    WHERE status = 'pending';

-- Index for stuck jobs detection
CREATE INDEX IF NOT EXISTS idx_enrichments_processing_stuck
    ON message_enrichments(created_at ASC)
    WHERE status = 'processing';

-- Comments for documentation
COMMENT ON TABLE message_enrichments IS 'Armazena enriquecimentos de mensagens com mídia (transcrição, OCR, parsing)';
COMMENT ON COLUMN message_enrichments.content_type IS 'Tipo de conteúdo: audio, voice (PTT), image, video, document';
COMMENT ON COLUMN message_enrichments.provider IS 'Provider de IA: whisper, deepgram, vision, llamaparse, ffmpeg, tesseract';
COMMENT ON COLUMN message_enrichments.status IS 'Status: pending, processing, completed, failed';
COMMENT ON COLUMN message_enrichments.extracted_text IS 'Texto extraído (transcrição, OCR, parsing)';
COMMENT ON COLUMN message_enrichments.metadata IS 'Metadados do provider (segments, objects, etc)';
COMMENT ON COLUMN message_enrichments.processing_time_ms IS 'Tempo de processamento em milissegundos';
COMMENT ON COLUMN message_enrichments.error IS 'Mensagem de erro (se falhou)';
COMMENT ON COLUMN message_enrichments.context IS 'Contexto de processamento (chat_message, profile_picture, pipeline_product, etc)';
