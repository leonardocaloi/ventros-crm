-- Add debounce_timeout_ms column to channels table
-- Used for message grouping/debouncing configuration

ALTER TABLE channels
ADD COLUMN debounce_timeout_ms INTEGER NOT NULL DEFAULT 15000;

-- Add comment for documentation
COMMENT ON COLUMN channels.debounce_timeout_ms IS 'Timeout do debouncer em milissegundos (default: 15000ms = 15s). Usado para agrupar mensagens sequenciais com mídia.';

-- Create index for performance (optional, but useful for queries)
CREATE INDEX IF NOT EXISTS idx_channels_debounce_timeout
ON channels(debounce_timeout_ms)
WHERE debounce_timeout_ms != 15000; -- Only index non-default values
