-- ========================================
-- Migration 000048 Rollback: Remove System Agents
-- ========================================

-- 1. Remove system agents (order doesn't matter for deletes)
DELETE FROM agents WHERE id = '00000000-0000-0000-0000-000000000001'::UUID; -- Broadcast
DELETE FROM agents WHERE id = '00000000-0000-0000-0000-000000000002'::UUID; -- Sequence
DELETE FROM agents WHERE id = '00000000-0000-0000-0000-000000000003'::UUID; -- Trigger
DELETE FROM agents WHERE id = '00000000-0000-0000-0000-000000000004'::UUID; -- Webhook
DELETE FROM agents WHERE id = '00000000-0000-0000-0000-000000000005'::UUID; -- Scheduled
DELETE FROM agents WHERE id = '00000000-0000-0000-0000-000000000010'::UUID; -- Test
DELETE FROM agents WHERE id = '00000000-0000-0000-0000-000000000099'::UUID; -- Default

-- 2. Remove system project if it exists (will cascade to billing_account and user via FK)
-- Only delete if it has the exact system values
DELETE FROM projects
WHERE id = '00000000-0000-0000-0000-000000000001'::UUID
  AND tenant_id = 'system'
  AND name = 'System Project';

-- Note: billing_accounts and users have ON DELETE CASCADE, so they'll be removed automatically
-- But we'll manually delete them for safety since we created them with specific IDs
DELETE FROM billing_accounts WHERE id = '00000000-0000-0000-0000-000000000001'::UUID;
DELETE FROM users WHERE id = '00000000-0000-0000-0000-000000000001'::UUID AND email = 'system@ventros.crm';

-- 3. Remove indexes for source column
DROP INDEX IF EXISTS idx_messages_agent_source;
DROP INDEX IF EXISTS idx_messages_source;

-- 4. Remove check constraint for source
ALTER TABLE messages DROP CONSTRAINT IF EXISTS chk_messages_source;

-- 5. Remove source column from messages table
ALTER TABLE messages DROP COLUMN IF EXISTS source;
