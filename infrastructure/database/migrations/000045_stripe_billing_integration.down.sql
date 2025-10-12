-- Migration 000045: Stripe Billing Integration - Rollback

-- Drop usage_meters table
DROP INDEX IF EXISTS idx_usage_meters_last_reported;
DROP INDEX IF EXISTS idx_usage_meters_period;
DROP INDEX IF EXISTS idx_usage_meters_metric_name;
DROP INDEX IF EXISTS idx_usage_meters_stripe_meter_id;
DROP INDEX IF EXISTS idx_usage_meters_stripe_customer_id;
DROP INDEX IF EXISTS idx_usage_meters_billing_account_id;
DROP TABLE IF EXISTS usage_meters;

-- Drop invoices table
DROP INDEX IF EXISTS idx_invoices_paid_at;
DROP INDEX IF EXISTS idx_invoices_due_date;
DROP INDEX IF EXISTS idx_invoices_status;
DROP INDEX IF EXISTS idx_invoices_stripe_invoice_id;
DROP INDEX IF EXISTS idx_invoices_subscription_id;
DROP INDEX IF EXISTS idx_invoices_billing_account_id;
DROP TABLE IF EXISTS invoices;

-- Drop subscriptions table
DROP INDEX IF EXISTS idx_subscriptions_current_period_end;
DROP INDEX IF EXISTS idx_subscriptions_status;
DROP INDEX IF EXISTS idx_subscriptions_stripe_subscription_id;
DROP INDEX IF EXISTS idx_subscriptions_billing_account_id;
DROP TABLE IF EXISTS subscriptions;

-- Remove stripe_customer_id from billing_accounts
DROP INDEX IF EXISTS idx_billing_accounts_stripe_customer_id;
ALTER TABLE billing_accounts DROP COLUMN IF EXISTS stripe_customer_id;
