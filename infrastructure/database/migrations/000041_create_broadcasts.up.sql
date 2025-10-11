-- Create broadcasts table
CREATE TABLE IF NOT EXISTS broadcasts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    list_id UUID NOT NULL,
    message_template JSONB NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'draft',
    scheduled_for TIMESTAMP,
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    total_contacts INTEGER NOT NULL DEFAULT 0,
    sent_count INTEGER NOT NULL DEFAULT 0,
    failed_count INTEGER NOT NULL DEFAULT 0,
    pending_count INTEGER NOT NULL DEFAULT 0,
    rate_limit INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for broadcasts
CREATE INDEX idx_broadcasts_tenant ON broadcasts(tenant_id);
CREATE INDEX idx_broadcasts_list ON broadcasts(list_id);
CREATE INDEX idx_broadcasts_status ON broadcasts(status);
CREATE INDEX idx_broadcasts_scheduled ON broadcasts(scheduled_for) WHERE status = 'scheduled';

-- Create broadcast_executions table
CREATE TABLE IF NOT EXISTS broadcast_executions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    broadcast_id UUID NOT NULL REFERENCES broadcasts(id) ON DELETE CASCADE,
    contact_id UUID NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    message_id UUID,
    error TEXT,
    sent_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for broadcast_executions
CREATE INDEX idx_executions_broadcast ON broadcast_executions(broadcast_id);
CREATE INDEX idx_executions_contact ON broadcast_executions(contact_id);
CREATE INDEX idx_executions_status ON broadcast_executions(status);
CREATE INDEX idx_executions_broadcast_status ON broadcast_executions(broadcast_id, status);

-- Add comments
COMMENT ON TABLE broadcasts IS 'Mass broadcast messages to contact lists';
COMMENT ON TABLE broadcast_executions IS 'Individual execution tracking per contact';
COMMENT ON COLUMN broadcasts.message_template IS 'JSON template with type, content, variables, etc';
COMMENT ON COLUMN broadcasts.rate_limit IS 'Messages per minute (0 = unlimited)';
COMMENT ON COLUMN broadcasts.status IS 'draft, scheduled, running, completed, failed, cancelled';
COMMENT ON COLUMN broadcast_executions.status IS 'pending, sending, sent, failed, skipped';
