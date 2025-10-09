-- Add optional timeout override to channels (can override project default)
ALTER TABLE channels
ADD COLUMN session_timeout_minutes INT NULL; -- NULL = use project default

-- Add optional timeout override to pipelines (can override channel OR project default)
ALTER TABLE pipelines
ADD COLUMN session_timeout_minutes INT NULL; -- NULL = use channel or project default

-- Add comments explaining the hierarchy
COMMENT ON COLUMN projects.session_timeout_minutes IS 'Base session timeout (default: 30 min). Inherited by all channels and pipelines unless overridden.';
COMMENT ON COLUMN channels.session_timeout_minutes IS 'Optional session timeout override. NULL = inherit from project. Overrides project default.';
COMMENT ON COLUMN pipelines.session_timeout_minutes IS 'Optional session timeout override. NULL = inherit from channel or project. Final override.';

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_channels_session_timeout ON channels(session_timeout_minutes) WHERE session_timeout_minutes IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_pipelines_session_timeout ON pipelines(session_timeout_minutes) WHERE session_timeout_minutes IS NOT NULL;
