-- Add custom_field_definitions JSONB column to channels table
-- This stores the catalog/schema of available custom fields for each channel
ALTER TABLE channels
ADD COLUMN IF NOT EXISTS custom_field_definitions JSONB DEFAULT '{}';

-- Add custom_fields JSONB column to chats table
-- This stores the actual values of custom fields for each chat
ALTER TABLE chats
ADD COLUMN IF NOT EXISTS custom_fields JSONB DEFAULT '{}';

-- Create GIN indexes for efficient JSONB queries
CREATE INDEX IF NOT EXISTS idx_channels_custom_field_definitions
ON channels USING gin(custom_field_definitions);

CREATE INDEX IF NOT EXISTS idx_chats_custom_fields
ON chats USING gin(custom_fields);

-- Migrate existing labels data from channels
-- Move Config["labels"] to custom_field_definitions["labels"]
-- Only for WAHA-based channels (waha, whatsapp_business)
UPDATE channels
SET custom_field_definitions = jsonb_build_object(
    'labels',
    jsonb_build_object(
        'type', 'label',
        'required', true,
        'fixed', true,
        'system', true,
        'description', 'WhatsApp labels synchronized from WAHA',
        'definitions', COALESCE(config->'labels', '[]'::jsonb)
    )
)
WHERE type IN ('waha', 'whatsapp_business')
AND config ? 'labels';

-- Migrate existing label IDs from chats
-- Move Metadata["label_ids"] to custom_fields["labels"]
UPDATE chats
SET custom_fields = jsonb_build_object(
    'labels',
    jsonb_build_object(
        'type', 'label',
        'value', COALESCE(metadata->'label_ids', '[]'::jsonb)
    )
)
WHERE metadata ? 'label_ids';

-- Add comment to explain the columns
COMMENT ON COLUMN channels.custom_field_definitions IS 'JSONB storage for custom field definitions/schema specific to each channel type';
COMMENT ON COLUMN chats.custom_fields IS 'JSONB storage for custom field values for each chat';
