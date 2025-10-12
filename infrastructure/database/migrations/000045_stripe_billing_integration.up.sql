-- Migration 000045: Stripe Billing Integration
-- Add Stripe Customer ID to billing_accounts and create subscription/invoice/usage_meter tables

-- 1. Add stripe_customer_id to billing_accounts
ALTER TABLE billing_accounts
ADD COLUMN IF NOT EXISTS stripe_customer_id VARCHAR(255);

CREATE INDEX IF NOT EXISTS idx_billing_accounts_stripe_customer_id
ON billing_accounts(stripe_customer_id);

-- 2. Create subscriptions table
CREATE TABLE IF NOT EXISTS subscriptions (
    id UUID PRIMARY KEY,
    billing_account_id UUID NOT NULL REFERENCES billing_accounts(id) ON DELETE CASCADE,
    stripe_subscription_id VARCHAR(255) NOT NULL UNIQUE,
    stripe_price_id VARCHAR(255) NOT NULL,
    status VARCHAR(50) NOT NULL,
    current_period_start TIMESTAMP WITH TIME ZONE NOT NULL,
    current_period_end TIMESTAMP WITH TIME ZONE NOT NULL,
    trial_start TIMESTAMP WITH TIME ZONE,
    trial_end TIMESTAMP WITH TIME ZONE,
    cancel_at TIMESTAMP WITH TIME ZONE,
    canceled_at TIMESTAMP WITH TIME ZONE,
    cancel_at_period_end BOOLEAN NOT NULL DEFAULT FALSE,
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_subscriptions_billing_account_id
ON subscriptions(billing_account_id);

CREATE INDEX IF NOT EXISTS idx_subscriptions_stripe_subscription_id
ON subscriptions(stripe_subscription_id);

CREATE INDEX IF NOT EXISTS idx_subscriptions_status
ON subscriptions(status);

CREATE INDEX IF NOT EXISTS idx_subscriptions_current_period_end
ON subscriptions(current_period_end)
WHERE status IN ('active', 'trialing');

-- 3. Create invoices table
CREATE TABLE IF NOT EXISTS invoices (
    id UUID PRIMARY KEY,
    billing_account_id UUID NOT NULL REFERENCES billing_accounts(id) ON DELETE CASCADE,
    subscription_id UUID REFERENCES subscriptions(id) ON DELETE SET NULL,
    stripe_invoice_id VARCHAR(255) NOT NULL UNIQUE,
    stripe_subscription_id VARCHAR(255),
    amount_due BIGINT NOT NULL,
    amount_paid BIGINT NOT NULL DEFAULT 0,
    amount_remaining BIGINT NOT NULL,
    currency VARCHAR(3) NOT NULL,
    status VARCHAR(50) NOT NULL,
    hosted_invoice_url TEXT,
    invoice_pdf TEXT,
    due_date TIMESTAMP WITH TIME ZONE,
    paid_at TIMESTAMP WITH TIME ZONE,
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_invoices_billing_account_id
ON invoices(billing_account_id);

CREATE INDEX IF NOT EXISTS idx_invoices_subscription_id
ON invoices(subscription_id);

CREATE INDEX IF NOT EXISTS idx_invoices_stripe_invoice_id
ON invoices(stripe_invoice_id);

CREATE INDEX IF NOT EXISTS idx_invoices_status
ON invoices(status);

CREATE INDEX IF NOT EXISTS idx_invoices_due_date
ON invoices(due_date)
WHERE status = 'open';

CREATE INDEX IF NOT EXISTS idx_invoices_paid_at
ON invoices(paid_at)
WHERE status = 'paid';

-- 4. Create usage_meters table
CREATE TABLE IF NOT EXISTS usage_meters (
    id UUID PRIMARY KEY,
    billing_account_id UUID NOT NULL REFERENCES billing_accounts(id) ON DELETE CASCADE,
    stripe_customer_id VARCHAR(255) NOT NULL,
    stripe_meter_id VARCHAR(255) NOT NULL,
    metric_name VARCHAR(100) NOT NULL,
    event_name VARCHAR(100) NOT NULL,
    quantity BIGINT NOT NULL DEFAULT 0,
    period_start TIMESTAMP WITH TIME ZONE NOT NULL,
    period_end TIMESTAMP WITH TIME ZONE NOT NULL,
    last_reported_at TIMESTAMP WITH TIME ZONE,
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_usage_meters_billing_account_id
ON usage_meters(billing_account_id);

CREATE INDEX IF NOT EXISTS idx_usage_meters_stripe_customer_id
ON usage_meters(stripe_customer_id);

CREATE INDEX IF NOT EXISTS idx_usage_meters_stripe_meter_id
ON usage_meters(stripe_meter_id);

CREATE INDEX IF NOT EXISTS idx_usage_meters_metric_name
ON usage_meters(billing_account_id, metric_name);

CREATE INDEX IF NOT EXISTS idx_usage_meters_period
ON usage_meters(period_start, period_end);

CREATE INDEX IF NOT EXISTS idx_usage_meters_last_reported
ON usage_meters(last_reported_at);

-- 5. Add comments for documentation
COMMENT ON TABLE subscriptions IS 'Stripe subscriptions for recurring billing';
COMMENT ON TABLE invoices IS 'Stripe invoices for billing charges';
COMMENT ON TABLE usage_meters IS 'Usage meters for usage-based billing (Stripe Billing Meters V2)';

COMMENT ON COLUMN billing_accounts.stripe_customer_id IS 'Stripe Customer ID (cus_xxx)';
COMMENT ON COLUMN subscriptions.stripe_subscription_id IS 'Stripe Subscription ID (sub_xxx)';
COMMENT ON COLUMN subscriptions.stripe_price_id IS 'Stripe Price ID (price_xxx)';
COMMENT ON COLUMN invoices.stripe_invoice_id IS 'Stripe Invoice ID (in_xxx)';
COMMENT ON COLUMN usage_meters.stripe_meter_id IS 'Stripe Billing Meter ID (mtr_xxx)';
COMMENT ON COLUMN usage_meters.metric_name IS 'Metric name (ex: messages_sent, ai_tokens, contacts)';
COMMENT ON COLUMN usage_meters.event_name IS 'Stripe event name (ex: message.sent, ai.token.used)';
COMMENT ON COLUMN usage_meters.quantity IS 'Running total of usage for the current period';
