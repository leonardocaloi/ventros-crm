-- Add pipeline association to channels
ALTER TABLE channels
ADD COLUMN pipeline_id UUID REFERENCES pipelines(id) ON DELETE SET NULL,
ADD COLUMN default_session_timeout_minutes INT NOT NULL DEFAULT 30;

-- Create index for faster queries
CREATE INDEX idx_channels_pipeline_id ON channels(pipeline_id);

-- Add comment
COMMENT ON COLUMN channels.pipeline_id IS 'Optional pipeline associated with this channel';
COMMENT ON COLUMN channels.default_session_timeout_minutes IS 'Session timeout in minutes. PRIORITY: Channel timeout > Pipeline timeout > Default (30 min)';
