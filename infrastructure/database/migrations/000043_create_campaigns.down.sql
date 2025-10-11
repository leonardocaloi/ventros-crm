-- Drop indexes for campaign_enrollments
DROP INDEX IF EXISTS idx_campaign_enrollments_campaign_contact_unique;
DROP INDEX IF EXISTS idx_campaign_enrollments_next_scheduled;
DROP INDEX IF EXISTS idx_campaign_enrollments_status;
DROP INDEX IF EXISTS idx_campaign_enrollments_contact_id;
DROP INDEX IF EXISTS idx_campaign_enrollments_campaign_id;

-- Drop campaign_enrollments table
DROP TABLE IF EXISTS campaign_enrollments;

-- Drop indexes for campaign_steps
DROP INDEX IF EXISTS idx_campaign_steps_campaign_order;
DROP INDEX IF EXISTS idx_campaign_steps_campaign_id;

-- Drop campaign_steps table
DROP TABLE IF EXISTS campaign_steps;

-- Drop indexes for campaigns
DROP INDEX IF EXISTS idx_campaigns_status;
DROP INDEX IF EXISTS idx_campaigns_tenant;

-- Drop campaigns table
DROP TABLE IF EXISTS campaigns;
