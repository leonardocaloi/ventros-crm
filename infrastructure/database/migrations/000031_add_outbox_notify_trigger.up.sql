-- Create function that sends NOTIFY when a new outbox event is inserted
-- This enables push-based event processing (no polling!)
CREATE OR REPLACE FUNCTION notify_outbox_event()
RETURNS TRIGGER AS $$
BEGIN
    -- Send notification with event_id as payload
    -- PostgresNotifyOutboxProcessor will receive this < 100ms
    PERFORM pg_notify('outbox_events', NEW.id::text);
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger that fires AFTER INSERT on outbox_events
-- Only fires after COMMIT (transaction safety!)
CREATE TRIGGER trigger_notify_outbox_event
    AFTER INSERT ON outbox_events
    FOR EACH ROW
    WHEN (NEW.status = 'pending')
    EXECUTE FUNCTION notify_outbox_event();

COMMENT ON FUNCTION notify_outbox_event IS 'Sends PostgreSQL NOTIFY when new outbox event is created (push-based, no polling!)';
COMMENT ON TRIGGER trigger_notify_outbox_event ON outbox_events IS 'Triggers NOTIFY after INSERT (only for pending events)';
