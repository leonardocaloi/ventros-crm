-- Create automation_rules table
CREATE TABLE IF NOT EXISTS automation_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    pipeline_id UUID NOT NULL REFERENCES pipelines(id) ON DELETE CASCADE,
    tenant_id VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    trigger VARCHAR(100) NOT NULL,
    conditions JSONB DEFAULT '[]'::jsonb,
    actions JSONB DEFAULT '[]'::jsonb,
    priority INT NOT NULL DEFAULT 0,
    enabled BOOLEAN NOT NULL DEFAULT true,
    schedule JSONB,
    last_executed TIMESTAMP WITH TIME ZONE,
    next_execution TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Indexes for efficient querying
CREATE INDEX idx_automation_pipeline ON automation_rules(pipeline_id);
CREATE INDEX idx_automation_tenant ON automation_rules(tenant_id);
CREATE INDEX idx_automation_trigger ON automation_rules(trigger);
CREATE INDEX idx_automation_priority ON automation_rules(priority);
CREATE INDEX idx_automation_enabled ON automation_rules(enabled);
CREATE INDEX idx_automation_last_executed ON automation_rules(last_executed);
CREATE INDEX idx_automation_next_execution ON automation_rules(next_execution);
CREATE INDEX idx_automation_pipeline_trigger ON automation_rules(pipeline_id, trigger) WHERE enabled = true;
CREATE INDEX idx_automation_tenant_enabled ON automation_rules(tenant_id, enabled);

-- Composite index for most common query pattern
CREATE INDEX idx_automation_active_rules ON automation_rules(pipeline_id, trigger, priority) WHERE enabled = true;

-- Index for scheduled rules worker (busca regras prontas para executar)
CREATE INDEX idx_automation_scheduled_ready ON automation_rules(next_execution, enabled)
WHERE trigger = 'scheduled' AND enabled = true AND next_execution IS NOT NULL;

-- Comments
COMMENT ON TABLE automation_rules IS 'Automatic follow-up rules associated with pipelines';
COMMENT ON COLUMN automation_rules.pipeline_id IS 'Pipeline that owns this rule';
COMMENT ON COLUMN automation_rules.tenant_id IS 'Tenant isolation';
COMMENT ON COLUMN automation_rules.trigger IS 'Event that triggers rule evaluation (session.ended, no_response.timeout, scheduled, etc)';
COMMENT ON COLUMN automation_rules.conditions IS 'JSON array of RuleCondition objects (field, operator, value)';
COMMENT ON COLUMN automation_rules.actions IS 'JSON array of RuleAction objects (type, params, delay_minutes)';
COMMENT ON COLUMN automation_rules.priority IS 'Execution priority (lower value = higher priority)';
COMMENT ON COLUMN automation_rules.enabled IS 'Whether rule is active';
COMMENT ON COLUMN automation_rules.schedule IS 'ScheduledRuleConfig for scheduled/cron rules (type, cron_expr, day_of_week, etc)';
COMMENT ON COLUMN automation_rules.last_executed IS 'Timestamp of last execution (for scheduled rules)';
COMMENT ON COLUMN automation_rules.next_execution IS 'Timestamp of next scheduled execution';

-- Trigger para atualizar updated_at
CREATE OR REPLACE FUNCTION update_automation_rules_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_automation_rules_updated_at
    BEFORE UPDATE ON automation_rules
    FOR EACH ROW
    EXECUTE FUNCTION update_automation_rules_updated_at();
