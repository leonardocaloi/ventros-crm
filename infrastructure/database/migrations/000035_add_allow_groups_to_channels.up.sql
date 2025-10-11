-- Add allow_groups and tracking_enabled columns to channels table
-- This allows channels to configure group processing and message tracking

ALTER TABLE channels
ADD COLUMN allow_groups BOOLEAN NOT NULL DEFAULT false,
ADD COLUMN tracking_enabled BOOLEAN NOT NULL DEFAULT false;

COMMENT ON COLUMN channels.allow_groups IS 'Whether the channel processes WhatsApp group messages';
COMMENT ON COLUMN channels.tracking_enabled IS 'Whether the channel tracks message origins (basic tracking)';

-- Create indexes for filtering
CREATE INDEX idx_channels_allow_groups ON channels(allow_groups) WHERE allow_groups = true;
CREATE INDEX idx_channels_tracking_enabled ON channels(tracking_enabled) WHERE tracking_enabled = true;
