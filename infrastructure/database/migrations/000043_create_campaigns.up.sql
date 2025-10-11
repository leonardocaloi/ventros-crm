-- Create campaigns table
CREATE TABLE IF NOT EXISTS campaigns (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    status VARCHAR(50) NOT NULL DEFAULT 'draft',
    goal_type VARCHAR(50) NOT NULL,
    goal_value INTEGER NOT NULL DEFAULT 0,
    contacts_reached INTEGER NOT NULL DEFAULT 0,
    conversions_count INTEGER NOT NULL DEFAULT 0,
    start_date TIMESTAMP,
    end_date TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for campaigns
CREATE INDEX IF NOT EXISTS idx_campaigns_tenant ON campaigns(tenant_id);
CREATE INDEX IF NOT EXISTS idx_campaigns_status ON campaigns(status);

-- Create campaign_steps table
CREATE TABLE IF NOT EXISTS campaign_steps (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    campaign_id UUID NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
    "order" INTEGER NOT NULL,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL,
    config JSONB NOT NULL,
    conditions JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for campaign_steps
CREATE INDEX IF NOT EXISTS idx_campaign_steps_campaign_id ON campaign_steps(campaign_id);
CREATE INDEX IF NOT EXISTS idx_campaign_steps_campaign_order ON campaign_steps(campaign_id, "order");

-- Create campaign_enrollments table
CREATE TABLE IF NOT EXISTS campaign_enrollments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    campaign_id UUID NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
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

-- Create indexes for campaign_enrollments
CREATE INDEX IF NOT EXISTS idx_campaign_enrollments_campaign_id ON campaign_enrollments(campaign_id);
CREATE INDEX IF NOT EXISTS idx_campaign_enrollments_contact_id ON campaign_enrollments(contact_id);
CREATE INDEX IF NOT EXISTS idx_campaign_enrollments_status ON campaign_enrollments(status);
CREATE INDEX IF NOT EXISTS idx_campaign_enrollments_next_scheduled ON campaign_enrollments(next_scheduled_at);

-- Create unique index to prevent duplicate active enrollments
CREATE UNIQUE INDEX IF NOT EXISTS idx_campaign_enrollments_campaign_contact_unique
    ON campaign_enrollments(campaign_id, contact_id)
    WHERE status = 'active';
