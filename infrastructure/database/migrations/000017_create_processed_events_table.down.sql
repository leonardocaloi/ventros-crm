-- Rollback: Remove tabela de eventos processados
DROP INDEX IF EXISTS idx_processed_events_cleanup;
DROP INDEX IF EXISTS idx_processed_events_lookup;
DROP TABLE IF EXISTS processed_events;
