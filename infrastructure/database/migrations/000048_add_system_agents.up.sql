-- ========================================
-- Migration 000048: Add System Agents
-- ========================================
-- This migration adds:
-- 1. source column to messages table
-- 2. Creates system agents with reserved UUIDs
-- 3. Adds validation constraints

-- ========================================
-- 1. Add source column to messages table
-- ========================================
ALTER TABLE messages
ADD COLUMN IF NOT EXISTS source VARCHAR(20) DEFAULT 'manual' NOT NULL;

-- Add check constraint for valid source values
ALTER TABLE messages
ADD CONSTRAINT chk_messages_source CHECK (
    source IN ('manual', 'broadcast', 'sequence', 'trigger', 'bot', 'system', 'webhook', 'scheduled', 'test')
);

-- Add index for source filtering
CREATE INDEX IF NOT EXISTS idx_messages_source ON messages(source);
CREATE INDEX IF NOT EXISTS idx_messages_agent_source ON messages(agent_id, source);

-- Add comment
COMMENT ON COLUMN messages.source IS 'Origin of the message: manual (human agent), broadcast (campaign), sequence (automation), trigger (pipeline rule), bot (AI), system (internal), webhook (webhook response), scheduled (scheduled message), test (E2E test)';

-- ========================================
-- 2. Create System Agents
-- ========================================
-- System agents use reserved UUID range: 00000000-0000-0000-0000-0000000000XX (00-99)
-- These agents are immutable and cannot be created/modified via API

-- First, we need to ensure we have a default project and tenant
-- Get the first project for system agents OR create a system project
DO $$
DECLARE
    default_project_id UUID;
    default_tenant_id TEXT;
    system_user_id UUID;
    system_billing_account_id UUID;
BEGIN
    -- Get first active project
    SELECT id, tenant_id INTO default_project_id, default_tenant_id
    FROM projects
    WHERE active = true AND deleted_at IS NULL
    ORDER BY created_at ASC
    LIMIT 1;

    -- If no project exists, create a system project for system agents
    IF default_project_id IS NULL THEN
        RAISE NOTICE 'No projects found. Creating system project for system agents.';

        -- Create system user if doesn't exist
        INSERT INTO users (
            id, name, email, password_hash, status, role, created_at, updated_at
        ) VALUES (
            '00000000-0000-0000-0000-000000000001'::UUID,
            'System User',
            'system@ventros.crm',
            '', -- No password (cannot login)
            'active',
            'admin',
            NOW(),
            NOW()
        ) ON CONFLICT (id) DO NOTHING
        RETURNING id INTO system_user_id;

        -- If INSERT didn't return (conflict), get the existing ID
        IF system_user_id IS NULL THEN
            system_user_id := '00000000-0000-0000-0000-000000000001'::UUID;
        END IF;

        -- Create system billing account if doesn't exist
        INSERT INTO billing_accounts (
            id, user_id, name, billing_email, payment_status, created_at, updated_at
        ) VALUES (
            '00000000-0000-0000-0000-000000000001'::UUID,
            system_user_id,
            'System Billing Account',
            'billing@ventros.crm',
            'active',
            NOW(),
            NOW()
        ) ON CONFLICT (id) DO NOTHING
        RETURNING id INTO system_billing_account_id;

        -- If INSERT didn't return (conflict), get the existing ID
        IF system_billing_account_id IS NULL THEN
            system_billing_account_id := '00000000-0000-0000-0000-000000000001'::UUID;
        END IF;

        -- Create system project
        INSERT INTO projects (
            id, user_id, billing_account_id, tenant_id, name, description, active, created_at, updated_at
        ) VALUES (
            '00000000-0000-0000-0000-000000000001'::UUID,
            system_user_id,
            system_billing_account_id,
            'system',
            'System Project',
            'Internal project for system agents and automation',
            true,
            NOW(),
            NOW()
        ) ON CONFLICT (id) DO NOTHING;

        default_project_id := '00000000-0000-0000-0000-000000000001'::UUID;
        default_tenant_id := 'system';
    END IF;

    -- Insert System Agent: Broadcast
    INSERT INTO agents (
        id, project_id, user_id, tenant_id, name, email, type, status, active,
        config, sessions_handled, average_response_ms, created_at, updated_at
    ) VALUES (
        '00000000-0000-0000-0000-000000000001'::UUID,
        default_project_id,
        NULL,
        default_tenant_id,
        'System - Broadcast',
        '',
        'system',
        'offline',
        true,
        '{"description": "Handles broadcast campaign messages", "immutable": true}'::JSONB,
        0,
        0,
        NOW(),
        NOW()
    ) ON CONFLICT (id) DO NOTHING;

    -- Insert System Agent: Sequence
    INSERT INTO agents (
        id, project_id, user_id, tenant_id, name, email, type, status, active,
        config, sessions_handled, average_response_ms, created_at, updated_at
    ) VALUES (
        '00000000-0000-0000-0000-000000000002'::UUID,
        default_project_id,
        NULL,
        default_tenant_id,
        'System - Sequence',
        '',
        'system',
        'offline',
        true,
        '{"description": "Handles automation sequence messages", "immutable": true}'::JSONB,
        0,
        0,
        NOW(),
        NOW()
    ) ON CONFLICT (id) DO NOTHING;

    -- Insert System Agent: Trigger
    INSERT INTO agents (
        id, project_id, user_id, tenant_id, name, email, type, status, active,
        config, sessions_handled, average_response_ms, created_at, updated_at
    ) VALUES (
        '00000000-0000-0000-0000-000000000003'::UUID,
        default_project_id,
        NULL,
        default_tenant_id,
        'System - Trigger',
        '',
        'system',
        'offline',
        true,
        '{"description": "Handles pipeline trigger/rule messages", "immutable": true}'::JSONB,
        0,
        0,
        NOW(),
        NOW()
    ) ON CONFLICT (id) DO NOTHING;

    -- Insert System Agent: Webhook
    INSERT INTO agents (
        id, project_id, user_id, tenant_id, name, email, type, status, active,
        config, sessions_handled, average_response_ms, created_at, updated_at
    ) VALUES (
        '00000000-0000-0000-0000-000000000004'::UUID,
        default_project_id,
        NULL,
        default_tenant_id,
        'System - Webhook',
        '',
        'system',
        'offline',
        true,
        '{"description": "Handles webhook automated responses", "immutable": true}'::JSONB,
        0,
        0,
        NOW(),
        NOW()
    ) ON CONFLICT (id) DO NOTHING;

    -- Insert System Agent: Scheduled
    INSERT INTO agents (
        id, project_id, user_id, tenant_id, name, email, type, status, active,
        config, sessions_handled, average_response_ms, created_at, updated_at
    ) VALUES (
        '00000000-0000-0000-0000-000000000005'::UUID,
        default_project_id,
        NULL,
        default_tenant_id,
        'System - Scheduled',
        '',
        'system',
        'offline',
        true,
        '{"description": "Handles scheduled messages", "immutable": true}'::JSONB,
        0,
        0,
        NOW(),
        NOW()
    ) ON CONFLICT (id) DO NOTHING;

    -- Insert System Agent: Test
    INSERT INTO agents (
        id, project_id, user_id, tenant_id, name, email, type, status, active,
        config, sessions_handled, average_response_ms, created_at, updated_at
    ) VALUES (
        '00000000-0000-0000-0000-000000000010'::UUID,
        default_project_id,
        NULL,
        default_tenant_id,
        'System - Test',
        '',
        'system',
        'offline',
        true,
        '{"description": "Handles E2E test messages and development testing", "immutable": true}'::JSONB,
        0,
        0,
        NOW(),
        NOW()
    ) ON CONFLICT (id) DO NOTHING;

    -- Insert System Agent: Default
    INSERT INTO agents (
        id, project_id, user_id, tenant_id, name, email, type, status, active,
        config, sessions_handled, average_response_ms, created_at, updated_at
    ) VALUES (
        '00000000-0000-0000-0000-000000000099'::UUID,
        default_project_id,
        NULL,
        default_tenant_id,
        'System - Default',
        '',
        'system',
        'offline',
        true,
        '{"description": "Fallback system agent for generic automation", "immutable": true}'::JSONB,
        0,
        0,
        NOW(),
        NOW()
    ) ON CONFLICT (id) DO NOTHING;

END $$;

-- ========================================
-- 3. Add comment for system agents
-- ========================================
-- System agents are protected via application layer (domain + repository)
-- See: internal/domain/crm/agent/agent.go
COMMENT ON COLUMN agents.type IS 'Agent type: human (real person), ai (AI assistant), bot (rule-based), channel (communication channel), virtual (historical), system (protected - cannot be modified)';
