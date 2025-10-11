-- Rollback migration: Move custom fields data back to original locations

-- Restore channel labels to config if they exist
UPDATE channels
SET config = jsonb_set(
    COALESCE(config, '{}'::jsonb),
    '{labels}',
    custom_field_definitions->'labels'->'definitions'
)
WHERE custom_field_definitions ? 'labels'
AND custom_field_definitions->'labels' ? 'definitions';

-- Restore chat label_ids to metadata if they exist
UPDATE chats
SET metadata = jsonb_set(
    COALESCE(metadata, '{}'::jsonb),
    '{label_ids}',
    custom_fields->'labels'->'value'
)
WHERE custom_fields ? 'labels'
AND custom_fields->'labels' ? 'value';

-- Drop indexes
DROP INDEX IF EXISTS idx_chats_custom_fields;
DROP INDEX IF EXISTS idx_channels_custom_field_definitions;

-- Drop columns
ALTER TABLE chats DROP COLUMN IF EXISTS custom_fields;
ALTER TABLE channels DROP COLUMN IF EXISTS custom_field_definitions;
