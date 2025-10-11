-- Drop message_enrichments table and all related objects
DROP INDEX IF EXISTS idx_enrichments_processing_stuck;
DROP INDEX IF EXISTS idx_enrichments_pending_priority;
DROP INDEX IF EXISTS idx_enrichments_created;
DROP INDEX IF EXISTS idx_enrichments_content_type;
DROP INDEX IF EXISTS idx_enrichments_status;
DROP INDEX IF EXISTS idx_enrichments_group;
DROP INDEX IF EXISTS idx_enrichments_message;

DROP TABLE IF EXISTS message_enrichments;
