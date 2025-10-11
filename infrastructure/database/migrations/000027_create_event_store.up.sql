-- Event Store for Event Sourcing Pattern
-- Stores all domain events as append-only immutable log

CREATE TABLE IF NOT EXISTS contact_event_store (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Aggregate identification
    aggregate_id UUID NOT NULL,
    aggregate_type VARCHAR(50) NOT NULL DEFAULT 'contact',

    -- Event metadata
    event_type VARCHAR(100) NOT NULL,
    event_version VARCHAR(10) NOT NULL DEFAULT 'v1',
    sequence_number BIGINT NOT NULL,

    -- Event data (JSONB for flexibility and indexing)
    event_data JSONB NOT NULL,
    metadata JSONB,

    -- Temporal tracking
    occurred_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),

    -- Multi-tenancy and isolation
    tenant_id VARCHAR(255) NOT NULL,
    project_id UUID,

    -- Causation and correlation tracking for distributed systems
    causation_id UUID,       -- Which command/event caused this event
    correlation_id UUID,     -- Workflow/saga tracking across aggregates

    -- Ensure events are sequential within an aggregate
    CONSTRAINT unique_aggregate_sequence UNIQUE(aggregate_id, sequence_number)
);

-- Comments for documentation
COMMENT ON TABLE contact_event_store IS 'Append-only event store for Contact aggregate using Event Sourcing pattern';
COMMENT ON COLUMN contact_event_store.aggregate_id IS 'Contact UUID - the aggregate root ID';
COMMENT ON COLUMN contact_event_store.event_type IS 'Event type in format: contact.created, contact.email_changed';
COMMENT ON COLUMN contact_event_store.event_version IS 'Event schema version for evolution: v1, v2, etc';
COMMENT ON COLUMN contact_event_store.sequence_number IS 'Sequential number within aggregate for ordering';
COMMENT ON COLUMN contact_event_store.event_data IS 'Full event payload as JSON';
COMMENT ON COLUMN contact_event_store.metadata IS 'Additional context: user_id, ip_address, user_agent, etc';
COMMENT ON COLUMN contact_event_store.causation_id IS 'ID of the command or event that caused this event';
COMMENT ON COLUMN contact_event_store.correlation_id IS 'Correlation ID for tracing across services/aggregates';

-- Indexes for performance optimization
CREATE INDEX idx_contact_events_aggregate
    ON contact_event_store(aggregate_id, sequence_number);

CREATE INDEX idx_contact_events_type
    ON contact_event_store(event_type, occurred_at DESC);

CREATE INDEX idx_contact_events_tenant
    ON contact_event_store(tenant_id, occurred_at DESC);

CREATE INDEX idx_contact_events_correlation
    ON contact_event_store(correlation_id)
    WHERE correlation_id IS NOT NULL;

CREATE INDEX idx_contact_events_occurred
    ON contact_event_store(occurred_at DESC);

-- GIN index for JSONB queries on event_data
CREATE INDEX idx_contact_events_data
    ON contact_event_store USING gin(event_data);

-- Snapshots table for performance optimization
-- Stores periodic snapshots of aggregate state to avoid replaying all events
CREATE TABLE IF NOT EXISTS contact_snapshots (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Aggregate identification
    aggregate_id UUID NOT NULL,

    -- Snapshot data
    snapshot_data JSONB NOT NULL,
    last_sequence_number BIGINT NOT NULL,

    -- Temporal tracking
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),

    -- Multi-tenancy
    tenant_id VARCHAR(255) NOT NULL,

    -- Ensure only one snapshot per aggregate at each sequence number
    CONSTRAINT unique_aggregate_snapshot UNIQUE(aggregate_id, last_sequence_number)
);

COMMENT ON TABLE contact_snapshots IS 'Snapshots of Contact aggregate state for performance optimization';
COMMENT ON COLUMN contact_snapshots.snapshot_data IS 'Serialized Contact state at this point in time';
COMMENT ON COLUMN contact_snapshots.last_sequence_number IS 'Last event sequence number included in this snapshot';

-- Index to quickly find latest snapshot for an aggregate
CREATE INDEX idx_contact_snapshots_aggregate
    ON contact_snapshots(aggregate_id, last_sequence_number DESC);

CREATE INDEX idx_contact_snapshots_tenant
    ON contact_snapshots(tenant_id, created_at DESC);
