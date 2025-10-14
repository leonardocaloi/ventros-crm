-- ========================================
-- Migration 000051 Rollback: Remove history_import Source
-- ========================================
-- This migration removes 'history_import' from valid message source values

-- Drop the constraint
ALTER TABLE messages DROP CONSTRAINT IF EXISTS chk_messages_source;

-- Recreate constraint without history_import
ALTER TABLE messages
ADD CONSTRAINT chk_messages_source CHECK (
    source IN ('manual', 'broadcast', 'sequence', 'trigger', 'bot', 'system', 'webhook', 'scheduled', 'test')
);

-- Restore original comment
COMMENT ON COLUMN messages.source IS 'Origin of the message: manual (human agent), broadcast (campaign), sequence (automation), trigger (pipeline rule), bot (AI), system (internal), webhook (webhook response), scheduled (scheduled message), test (E2E test)';
