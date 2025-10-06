-- Add webhook fields to channels table
ALTER TABLE channels
ADD COLUMN webhook_url TEXT,
ADD COLUMN webhook_configured_at TIMESTAMP WITH TIME ZONE,
ADD COLUMN webhook_active BOOLEAN DEFAULT FALSE;

-- Create index on webhook_url for faster queries
CREATE INDEX idx_channels_webhook_url ON channels(webhook_url) WHERE webhook_url IS NOT NULL;

-- Add comment explaining the fields
COMMENT ON COLUMN channels.webhook_url IS 'URL do webhook configurada para receber eventos do canal';
COMMENT ON COLUMN channels.webhook_configured_at IS 'Data/hora em que o webhook foi configurado';
COMMENT ON COLUMN channels.webhook_active IS 'Indica se o webhook está ativo e recebendo eventos';
