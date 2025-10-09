-- Tabela de credenciais criptografadas para integrações externas
CREATE TABLE IF NOT EXISTS credentials (
    id UUID PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL,
    project_id UUID,

    -- Tipo e identificação
    credential_type VARCHAR(50) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,

    -- Valor criptografado (AES-256-GCM)
    encrypted_value_ciphertext TEXT NOT NULL,
    encrypted_value_nonce TEXT NOT NULL,

    -- OAuth tokens (quando aplicável)
    oauth_access_token_ciphertext TEXT,
    oauth_access_token_nonce TEXT,
    oauth_refresh_token_ciphertext TEXT,
    oauth_refresh_token_nonce TEXT,
    oauth_token_type VARCHAR(20) DEFAULT 'Bearer',
    oauth_expires_at TIMESTAMP,

    -- Metadata adicional (JSONB para flexibilidade)
    metadata JSONB DEFAULT '{}'::jsonb,

    -- Status e lifecycle
    is_active BOOLEAN NOT NULL DEFAULT true,
    expires_at TIMESTAMP,
    last_used_at TIMESTAMP,

    -- Auditoria
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),

    -- Foreign Keys
    CONSTRAINT fk_credentials_projects FOREIGN KEY (project_id)
        REFERENCES projects(id) ON DELETE SET NULL,

    -- Constraints
    CONSTRAINT chk_credential_type CHECK (
        credential_type IN (
            'meta_whatsapp_cloud',
            'meta_ads',
            'meta_conversions_api',
            'google_ads',
            'google_analytics',
            'webhook_auth',
            'api_key',
            'basic_auth',
            'waha_instance'
        )
    )
);

-- Índices para performance
CREATE INDEX idx_credentials_tenant ON credentials(tenant_id);
CREATE INDEX idx_credentials_project ON credentials(project_id);
CREATE INDEX idx_credentials_type ON credentials(credential_type);
CREATE INDEX idx_credentials_tenant_type ON credentials(tenant_id, credential_type);
CREATE INDEX idx_credentials_active ON credentials(is_active) WHERE is_active = true;
CREATE INDEX idx_credentials_expires_at ON credentials(expires_at) WHERE expires_at IS NOT NULL;

-- Índice único para evitar duplicatas de nome por tenant/projeto
CREATE UNIQUE INDEX idx_credentials_unique_name
ON credentials(tenant_id, COALESCE(project_id::text, 'global'), name);

-- Índice GIN para busca em metadata
CREATE INDEX idx_credentials_metadata ON credentials USING GIN (metadata);

-- Comentários
COMMENT ON TABLE credentials IS 'Armazena credenciais criptografadas para integrações externas (Meta, Google, etc.)';
COMMENT ON COLUMN credentials.encrypted_value_ciphertext IS 'Texto cifrado (base64) do valor principal da credencial';
COMMENT ON COLUMN credentials.encrypted_value_nonce IS 'Nonce (base64) usado na criptografia AES-GCM';
COMMENT ON COLUMN credentials.oauth_access_token_ciphertext IS 'Access token OAuth criptografado (quando aplicável)';
COMMENT ON COLUMN credentials.oauth_refresh_token_ciphertext IS 'Refresh token OAuth criptografado (quando aplicável)';
COMMENT ON COLUMN credentials.metadata IS 'Dados adicionais em formato JSON (scopes, URLs, etc.)';
COMMENT ON COLUMN credentials.is_active IS 'Indica se a credencial está ativa e pode ser usada';
COMMENT ON COLUMN credentials.expires_at IS 'Data de expiração da credencial (se aplicável)';
COMMENT ON COLUMN credentials.last_used_at IS 'Última vez que a credencial foi utilizada';

-- Trigger para atualizar updated_at automaticamente
CREATE OR REPLACE FUNCTION update_credentials_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_credentials_updated_at
    BEFORE UPDATE ON credentials
    FOR EACH ROW
    EXECUTE FUNCTION update_credentials_updated_at();
