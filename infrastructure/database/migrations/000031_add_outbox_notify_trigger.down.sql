-- Remove trigger
DROP TRIGGER IF EXISTS trigger_notify_outbox_event ON outbox_events;

-- Remove function
DROP FUNCTION IF EXISTS notify_outbox_event();
