-- ========================================
-- Migration 000051: Add history_import Source
-- ========================================
-- This migration adds 'history_import' as a valid message source value
-- for the WAHA history import feature.
--
-- Context: When importing historical messages from WAHA, we need to mark
-- them with source='history_import' to distinguish from real-time messages.

-- Drop the existing constraint
ALTER TABLE messages DROP CONSTRAINT IF EXISTS chk_messages_source;

-- Recreate constraint with the new value
ALTER TABLE messages
ADD CONSTRAINT chk_messages_source CHECK (
    source IN ('manual', 'broadcast', 'sequence', 'trigger', 'bot', 'system', 'webhook', 'scheduled', 'test', 'history_import')
);

-- Update comment to include new source
COMMENT ON COLUMN messages.source IS 'Origin of the message: manual (human agent), broadcast (campaign), sequence (automation), trigger (pipeline rule), bot (AI), system (internal), webhook (webhook response), scheduled (scheduled message), test (E2E test), history_import (imported from WAHA history)';
