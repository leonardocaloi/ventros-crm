-- Create sequences table
CREATE TABLE IF NOT EXISTS sequences (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    status VARCHAR(50) NOT NULL DEFAULT 'draft',
    trigger_type VARCHAR(50) NOT NULL,
    trigger_data JSONB,
    exit_on_reply BOOLEAN NOT NULL DEFAULT true,
    total_enrolled INTEGER NOT NULL DEFAULT 0,
    active_count INTEGER NOT NULL DEFAULT 0,
    completed_count INTEGER NOT NULL DEFAULT 0,
    exited_count INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for sequences
CREATE INDEX IF NOT EXISTS idx_sequences_tenant ON sequences(tenant_id);
CREATE INDEX IF NOT EXISTS idx_sequences_status ON sequences(status);
CREATE INDEX IF NOT EXISTS idx_sequences_trigger_type ON sequences(trigger_type);

-- Create sequence_steps table
CREATE TABLE IF NOT EXISTS sequence_steps (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sequence_id UUID NOT NULL REFERENCES sequences(id) ON DELETE CASCADE,
    "order" INTEGER NOT NULL,
    name VARCHAR(255) NOT NULL,
    delay_amount INTEGER NOT NULL,
    delay_unit VARCHAR(20) NOT NULL,
    message_template JSONB NOT NULL,
    conditions JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for sequence_steps
CREATE INDEX IF NOT EXISTS idx_sequence_steps_sequence_id ON sequence_steps(sequence_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_sequence_steps_sequence_order ON sequence_steps(sequence_id, "order");

-- Create sequence_enrollments table
CREATE TABLE IF NOT EXISTS sequence_enrollments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sequence_id UUID NOT NULL REFERENCES sequences(id) ON DELETE CASCADE,
    contact_id UUID NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    current_step_order INTEGER NOT NULL DEFAULT 0,
    next_scheduled_at TIMESTAMP,
    exited_at TIMESTAMP,
    exit_reason TEXT,
    completed_at TIMESTAMP,
    enrolled_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for sequence_enrollments
CREATE INDEX IF NOT EXISTS idx_enrollments_sequence_id ON sequence_enrollments(sequence_id);
CREATE INDEX IF NOT EXISTS idx_enrollments_contact_id ON sequence_enrollments(contact_id);
CREATE INDEX IF NOT EXISTS idx_enrollments_status ON sequence_enrollments(status);
CREATE INDEX IF NOT EXISTS idx_enrollments_next_scheduled ON sequence_enrollments(next_scheduled_at) WHERE next_scheduled_at IS NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_enrollments_sequence_contact_unique
    ON sequence_enrollments(sequence_id, contact_id)
    WHERE status = 'active';

-- Add comments
COMMENT ON TABLE sequences IS 'Automated message sequences (drip campaigns)';
COMMENT ON TABLE sequence_steps IS 'Individual steps/messages in a sequence';
COMMENT ON TABLE sequence_enrollments IS 'Contact enrollments in sequences';
COMMENT ON COLUMN sequences.trigger_type IS 'How contacts enter this sequence (manual, tag_added, list_joined, etc)';
COMMENT ON COLUMN sequences.exit_on_reply IS 'Whether to exit the sequence when contact replies';
COMMENT ON COLUMN sequence_steps.delay_amount IS 'Amount of time to wait before sending this step';
COMMENT ON COLUMN sequence_steps.delay_unit IS 'Unit of delay: minutes, hours, days';
COMMENT ON COLUMN sequence_enrollments.current_step_order IS 'Current step order the contact is on';
COMMENT ON COLUMN sequence_enrollments.next_scheduled_at IS 'When to send the next message';
