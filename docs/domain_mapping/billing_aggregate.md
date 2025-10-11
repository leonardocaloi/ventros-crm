# Billing Aggregate

**Last Updated**: 2025-10-10
**Status**: ⚠️ Basic Implementation - Needs Stripe Integration
**Lines of Code**: ~250
**Test Coverage**: ~60%

---

## Overview

- **Purpose**: Manage billing accounts, payment methods, and subscriptions
- **Location**: `internal/domain/billing/`
- **Entity**: `infrastructure/persistence/entities/billing_account.go` (not implemented yet)
- **Repository**: Not implemented yet
- **Aggregate Root**: `BillingAccount`

**Business Problem**:
The Billing aggregate manages **payment processing and subscription lifecycle** for the SaaS platform. Critical for:
- **Revenue generation** - Collect payments from customers
- **Subscription management** - Handle monthly/annual billing cycles
- **Usage-based billing** - Charge for messages, AI processing, contacts
- **Payment compliance** - PCI-DSS, SCA (Strong Customer Authentication), 3D Secure
- **Dunning management** - Handle failed payments and retries
- **Multi-currency** - Support 135+ currencies via Stripe

---

## Domain Model

### Aggregate Root: BillingAccount

```go
type BillingAccount struct {
    id               uuid.UUID
    userID           uuid.UUID       // Owner of the billing account
    name             string          // Display name
    paymentStatus    PaymentStatus   // pending, active, suspended, canceled
    paymentMethods   []PaymentMethod // Credit cards, ACH, etc.
    billingEmail     string          // Email for invoices
    suspended        bool            // Is account suspended?
    suspendedAt      *time.Time
    suspensionReason string
    createdAt        time.Time
    updatedAt        time.Time
}
```

### Value Objects

#### PaymentStatus

```go
type PaymentStatus string
const (
    PaymentStatusPending   PaymentStatus = "pending"   // No payment method yet
    PaymentStatusActive    PaymentStatus = "active"    // Payment method active
    PaymentStatusSuspended PaymentStatus = "suspended" // Payment failed
    PaymentStatusCanceled  PaymentStatus = "canceled"  // Account canceled
)
```

#### PaymentMethod

```go
type PaymentMethod struct {
    Type       string      // card, ach_debit, sepa_debit, boleto
    LastDigits string      // Last 4 digits of card/account
    ExpiresAt  *time.Time  // Expiration date (for cards)
    IsDefault  bool        // Is this the default payment method?
}
```

### Business Invariants

1. **Billing account must have user and email**
   - `userID` required
   - `billingEmail` required for invoices

2. **Payment status lifecycle**
   - Created as `pending` (no payment method)
   - Becomes `active` when payment method added
   - Becomes `suspended` when payment fails
   - Becomes `canceled` when user cancels

3. **Suspended accounts**
   - Cannot create new projects
   - Existing projects become read-only
   - Data retained for 30 days

4. **Reactivation**
   - Requires active payment method
   - Cannot reactivate canceled accounts

---

## Events Emitted

| Event | When | Purpose |
|-------|------|---------|
| `billing.account.created` | New billing account created | Initialize billing |
| `billing.payment.activated` | Payment method activated | Enable billing |
| `billing.account.suspended` | Account suspended (payment failed) | Stop services |
| `billing.account.reactivated` | Account reactivated | Resume services |
| `billing.account.canceled` | Account canceled by user | Archive data |

**⚠️ Missing Events** (needed for Stripe integration):
- `billing.subscription.created`
- `billing.subscription.updated`
- `billing.subscription.canceled`
- `billing.invoice.created`
- `billing.invoice.paid`
- `billing.invoice.payment_failed`
- `billing.usage.recorded`
- `billing.payment_intent.succeeded`
- `billing.payment_intent.failed`

---

## Repository Interface

```go
type Repository interface {
    Create(ctx context.Context, account *BillingAccount) error
    FindByID(ctx context.Context, id uuid.UUID) (*BillingAccount, error)
    FindByUserID(ctx context.Context, userID uuid.UUID) ([]*BillingAccount, error)
    FindActiveByUserID(ctx context.Context, userID uuid.UUID) (*BillingAccount, error)
    Update(ctx context.Context, account *BillingAccount) error
    Delete(ctx context.Context, id uuid.UUID) error
}
```

**Note**: Repository interface defined but **GORM implementation not created yet**.

---

## Current Implementation Status

### ✅ What's Implemented

1. **BillingAccount aggregate root**
   - Basic lifecycle: create, activate, suspend, reactivate, cancel
   - Payment method management (add, set default)
   - Status validation

2. **Domain events** (5 events)
   - Account created, activated, suspended, reactivated, canceled

3. **Business rules**
   - Cannot reactivate without payment method
   - Suspended accounts cannot create projects
   - Payment status state machine

4. **Unit tests** (~60% coverage)

### ❌ What's Missing (Critical for Production)

1. **Stripe Integration** - No Stripe API integration
2. **Subscription Management** - No subscription entity
3. **Invoice Management** - No invoice entity
4. **Payment Intents** - No payment processing
5. **Usage Metering** - No usage tracking
6. **Webhook Handling** - No Stripe webhook receiver
7. **3D Secure / SCA** - No strong authentication
8. **Dunning Management** - No failed payment retries
9. **Multi-currency** - No currency support
10. **Proration** - No mid-cycle plan changes

---

## Stripe Integration Architecture

### Core Entities Needed

```go
// 1. Subscription (NEW)
type Subscription struct {
    id                uuid.UUID
    billingAccountID  uuid.UUID
    stripeSubscriptionID string      // Stripe subscription ID
    stripePriceID     string         // Stripe price ID
    status            SubscriptionStatus
    currentPeriodStart time.Time
    currentPeriodEnd   time.Time
    cancelAt          *time.Time
    metadata          map[string]interface{}
}

type SubscriptionStatus string
const (
    SubscriptionStatusTrialing       SubscriptionStatus = "trialing"
    SubscriptionStatusActive         SubscriptionStatus = "active"
    SubscriptionStatusIncomplete     SubscriptionStatus = "incomplete"
    SubscriptionStatusPastDue        SubscriptionStatus = "past_due"
    SubscriptionStatusCanceled       SubscriptionStatus = "canceled"
    SubscriptionStatusUnpaid         SubscriptionStatus = "unpaid"
)

// 2. Invoice (NEW)
type Invoice struct {
    id              uuid.UUID
    billingAccountID uuid.UUID
    stripeInvoiceID string
    subscriptionID  *uuid.UUID
    amountDue       int64      // Amount in cents
    amountPaid      int64      // Amount paid in cents
    currency        string     // USD, BRL, EUR
    status          InvoiceStatus
    hostedInvoiceURL string    // Stripe hosted invoice
    invoicePDF      string     // PDF URL
    dueDate         *time.Time
}

type InvoiceStatus string
const (
    InvoiceStatusDraft         InvoiceStatus = "draft"
    InvoiceStatusOpen          InvoiceStatus = "open"
    InvoiceStatusPaid          InvoiceStatus = "paid"
    InvoiceStatusUncollectible InvoiceStatus = "uncollectible"
    InvoiceStatusVoid          InvoiceStatus = "void"
)

// 3. UsageMeter (NEW)
type UsageMeter struct {
    id                uuid.UUID
    billingAccountID  uuid.UUID
    stripeMeterId     string
    metricName        string   // messages_sent, ai_tokens, contacts
    quantity          int64    // Current usage
    periodStart       time.Time
    periodEnd         time.Time
}

// 4. PaymentIntent (NEW)
type PaymentIntent struct {
    id                   uuid.UUID
    billingAccountID     uuid.UUID
    stripePaymentIntentID string
    amount               int64
    currency             string
    status               PaymentIntentStatus
    clientSecret         string  // For frontend
    requiresAction       bool    // 3D Secure needed
}

type PaymentIntentStatus string
const (
    PaymentIntentStatusRequiresPaymentMethod PaymentIntentStatus = "requires_payment_method"
    PaymentIntentStatusRequiresConfirmation  PaymentIntentStatus = "requires_confirmation"
    PaymentIntentStatusRequiresAction        PaymentIntentStatus = "requires_action"
    PaymentIntentStatusProcessing            PaymentIntentStatus = "processing"
    PaymentIntentStatusSucceeded             PaymentIntentStatus = "succeeded"
    PaymentIntentStatusCanceled              PaymentIntentStatus = "canceled"
)
```

---

## Real-World Usage

### Scenario 1: Create Billing Account with Stripe Customer

```go
// Step 1: Create billing account in domain
billingAccount, _ := billing.NewBillingAccount(
    userID,
    "Acme Corporation",
    "billing@acme.com",
)

// Step 2: Create Stripe customer
stripeCustomer, _ := stripeClient.Customers.New(&stripe.CustomerParams{
    Email: stripe.String("billing@acme.com"),
    Name:  stripe.String("Acme Corporation"),
    Metadata: map[string]string{
        "ventros_billing_account_id": billingAccount.ID().String(),
    },
})

// Step 3: Store Stripe customer ID in billing account
billingAccount.SetStripeCustomerID(stripeCustomer.ID)

billingRepo.Create(ctx, billingAccount)

// Event emitted: billing.account.created
```

### Scenario 2: Add Payment Method with 3D Secure (SCA Compliant)

```go
// Step 1: Create SetupIntent in Stripe (for future payments)
setupIntent, _ := stripeClient.SetupIntents.New(&stripe.SetupIntentParams{
    Customer: stripe.String(billingAccount.StripeCustomerID()),
    PaymentMethodTypes: []*string{
        stripe.String("card"),
    },
    Usage: stripe.String("off_session"), // For recurring payments
})

// Step 2: Return client_secret to frontend
// Frontend uses Stripe.js to collect card and confirm SetupIntent
// Stripe handles 3D Secure authentication automatically

// Step 3: Handle webhook: setup_intent.succeeded
func HandleSetupIntentSucceeded(event stripe.Event) {
    var setupIntent stripe.SetupIntent
    json.Unmarshal(event.Data.Raw, &setupIntent)

    // Retrieve payment method
    paymentMethod, _ := stripeClient.PaymentMethods.Get(
        setupIntent.PaymentMethod.ID,
        nil,
    )

    // Update domain
    billingAccount, _ := billingRepo.FindByStripeCustomerID(ctx, setupIntent.Customer.ID)

    billingAccount.ActivatePayment(billing.PaymentMethod{
        Type:       "card",
        LastDigits: paymentMethod.Card.Last4,
        ExpiresAt:  time.Date(
            int(paymentMethod.Card.ExpYear),
            time.Month(paymentMethod.Card.ExpMonth),
            1, 0, 0, 0, 0, time.UTC,
        ),
        IsDefault:  true,
    })

    billingRepo.Update(ctx, billingAccount)

    // Event emitted: billing.payment.activated
}
```

### Scenario 3: Create Subscription (Monthly Plan)

```go
// Step 1: Create subscription in Stripe
subscription, _ := stripeClient.Subscriptions.New(&stripe.SubscriptionParams{
    Customer: stripe.String(billingAccount.StripeCustomerID()),
    Items: []*stripe.SubscriptionItemsParams{
        {
            Price: stripe.String("price_monthly_basic"), // Stripe Price ID
        },
    },
    BillingMode: stripe.SubscriptionBillingModeFlexible, // New 2025 API
    PaymentBehavior: stripe.String("default_incomplete"),
    PaymentSettings: &stripe.SubscriptionPaymentSettingsParams{
        SaveDefaultPaymentMethod: stripe.String("on_subscription"),
    },
    Metadata: map[string]string{
        "ventros_billing_account_id": billingAccount.ID().String(),
    },
})

// Step 2: Create Subscription in domain
domainSubscription := billing.NewSubscription(
    billingAccount.ID(),
    subscription.ID,
    subscription.Items.Data[0].Price.ID,
    billing.SubscriptionStatusActive,
    time.Unix(subscription.CurrentPeriodStart, 0),
    time.Unix(subscription.CurrentPeriodEnd, 0),
)

subscriptionRepo.Create(ctx, domainSubscription)

// Event emitted: billing.subscription.created
```

### Scenario 4: Usage-Based Billing (Messages Sent)

```go
// Step 1: Define usage meter in Stripe (one-time setup)
meter, _ := stripeClient.Billing.Meters.New(&stripe.BillingMeterParams{
    DisplayName:  stripe.String("Messages Sent"),
    EventName:    stripe.String("message_sent"),
    ValueSettings: &stripe.BillingMeterValueSettingsParams{
        EventPayloadKey: stripe.String("message_count"),
    },
})

// Step 2: Record usage in real-time
func RecordMessageSent(billingAccountID uuid.UUID, messageCount int) {
    // Report to Stripe
    stripeClient.Billing.MeterEvents.Create(&stripe.BillingMeterEventParams{
        EventName: stripe.String("message_sent"),
        Payload: map[string]interface{}{
            "message_count": messageCount,
            "stripe_customer_id": billingAccount.StripeCustomerID(),
        },
    })

    // Update domain usage meter
    usageMeter.IncrementUsage(messageCount)
    usageMeterRepo.Update(ctx, usageMeter)

    // Event emitted: billing.usage.recorded
}

// Step 3: Stripe automatically creates invoice at end of billing period
// with metered charges
```

### Scenario 5: Handle Failed Payment (Dunning)

```go
// Webhook: invoice.payment_failed
func HandleInvoicePaymentFailed(event stripe.Event) {
    var invoice stripe.Invoice
    json.Unmarshal(event.Data.Raw, &invoice)

    billingAccount, _ := billingRepo.FindByStripeCustomerID(ctx, invoice.Customer.ID)

    // Attempt 1: Suspend account temporarily
    billingAccount.Suspend("Payment failed - retry in 3 days")
    billingRepo.Update(ctx, billingAccount)

    // Event emitted: billing.account.suspended

    // Schedule retry via Temporal workflow
    temporalClient.ExecuteWorkflow(ctx, "DunningWorkflow", DunningWorkflowInput{
        BillingAccountID: billingAccount.ID(),
        InvoiceID:        invoice.ID,
        AttemptNumber:    1,
        RetryAt:          time.Now().Add(3 * 24 * time.Hour),
    })
}

// Dunning workflow (Temporal)
func DunningWorkflow(ctx workflow.Context, input DunningWorkflowInput) error {
    // Retry 1: After 3 days
    workflow.Sleep(ctx, 3*24*time.Hour)

    success := workflow.ExecuteActivity(ctx, RetryPaymentActivity, input).Get(ctx, nil)
    if success {
        return nil // Payment succeeded
    }

    // Retry 2: After 7 days
    workflow.Sleep(ctx, 4*24*time.Hour)

    success = workflow.ExecuteActivity(ctx, RetryPaymentActivity, input).Get(ctx, nil)
    if success {
        return nil
    }

    // Retry 3: After 14 days - Final attempt
    workflow.Sleep(ctx, 7*24*time.Hour)

    success = workflow.ExecuteActivity(ctx, RetryPaymentActivity, input).Get(ctx, nil)
    if !success {
        // Cancel subscription after 14 days
        workflow.ExecuteActivity(ctx, CancelSubscriptionActivity, input)
    }

    return nil
}
```

### Scenario 6: Subscription Upgrade (Proration)

```go
// User upgrades from Basic to Pro mid-cycle
func UpgradeSubscription(subscriptionID uuid.UUID, newPriceID string) error {
    subscription, _ := subscriptionRepo.FindByID(ctx, subscriptionID)

    // Update in Stripe (automatic proration)
    stripeSubscription, _ := stripeClient.Subscriptions.Update(
        subscription.StripeSubscriptionID(),
        &stripe.SubscriptionParams{
            Items: []*stripe.SubscriptionItemsParams{
                {
                    ID:    stripe.String(subscription.Items[0].ID),
                    Price: stripe.String(newPriceID),
                },
            },
            ProrationBehavior: stripe.String("create_prorations"), // Pro-rate immediately
        },
    )

    // Update domain
    subscription.UpdatePrice(newPriceID)
    subscriptionRepo.Update(ctx, subscription)

    // Stripe automatically creates proration invoice
    // Event emitted: billing.subscription.updated
}
```

---

## Stripe Webhook Events

### Critical Events to Handle

```go
// infrastructure/http/handlers/stripe_webhook_handler.go

func (h *StripeWebhookHandler) HandleWebhook(c *gin.Context) {
    payload, _ := ioutil.ReadAll(c.Request.Body)

    // Verify webhook signature (REQUIRED for security)
    event, err := webhook.ConstructEvent(
        payload,
        c.GetHeader("Stripe-Signature"),
        webhookSecret,
    )

    switch event.Type {
    // Payment Intents
    case "payment_intent.succeeded":
        h.handlePaymentIntentSucceeded(event)
    case "payment_intent.payment_failed":
        h.handlePaymentIntentFailed(event)
    case "payment_intent.requires_action":
        h.handlePaymentIntentRequiresAction(event) // 3D Secure

    // Setup Intents (saving payment methods)
    case "setup_intent.succeeded":
        h.handleSetupIntentSucceeded(event)
    case "setup_intent.setup_failed":
        h.handleSetupIntentFailed(event)

    // Subscriptions
    case "customer.subscription.created":
        h.handleSubscriptionCreated(event)
    case "customer.subscription.updated":
        h.handleSubscriptionUpdated(event)
    case "customer.subscription.deleted":
        h.handleSubscriptionDeleted(event)
    case "customer.subscription.trial_will_end":
        h.handleTrialWillEnd(event) // 3 days before trial ends

    // Invoices
    case "invoice.created":
        h.handleInvoiceCreated(event)
    case "invoice.paid":
        h.handleInvoicePaid(event)
    case "invoice.payment_failed":
        h.handleInvoicePaymentFailed(event)
    case "invoice.upcoming":
        h.handleInvoiceUpcoming(event) // 7 days before renewal

    // Payment Methods
    case "payment_method.attached":
        h.handlePaymentMethodAttached(event)
    case "payment_method.detached":
        h.handlePaymentMethodDetached(event)

    // Customer
    case "customer.updated":
        h.handleCustomerUpdated(event)
    case "customer.deleted":
        h.handleCustomerDeleted(event)

    default:
        h.logger.Warn("Unhandled Stripe event", zap.String("type", event.Type))
    }

    c.JSON(http.StatusOK, gin.H{"received": true})
}
```

---

## Pricing Plans Architecture

### Stripe Products and Prices

```go
// One-time setup in Stripe Dashboard or via API

// Product 1: Basic Plan
product_basic := stripe.Product{
    ID:          "prod_basic",
    Name:        "Basic Plan",
    Description: "Up to 1,000 contacts, 5,000 messages/month",
}

// Price 1a: Monthly
price_basic_monthly := stripe.Price{
    ID:       "price_basic_monthly",
    Product:  "prod_basic",
    Currency: "usd",
    Recurring: {
        Interval:      "month",
        IntervalCount: 1,
    },
    UnitAmount: 2900, // $29.00
}

// Price 1b: Annual (with discount)
price_basic_annual := stripe.Price{
    ID:       "price_basic_annual",
    Product:  "prod_basic",
    Currency: "usd",
    Recurring: {
        Interval:      "year",
        IntervalCount: 1,
    },
    UnitAmount: 29000, // $290.00 (2 months free)
}

// Product 2: Pro Plan
product_pro := stripe.Product{
    ID:          "prod_pro",
    Name:        "Pro Plan",
    Description: "Up to 10,000 contacts, 50,000 messages/month",
}

// Price 2a: Monthly
price_pro_monthly := stripe.Price{
    ID:       "price_pro_monthly",
    Product:  "prod_pro",
    Currency: "usd",
    Recurring: {
        Interval:      "month",
        IntervalCount: 1,
    },
    UnitAmount: 9900, // $99.00
}

// Product 3: Usage-Based (Messages)
product_messages := stripe.Product{
    ID:          "prod_messages_overage",
    Name:        "Additional Messages",
    Description: "Pay-per-use messages beyond plan limit",
}

// Price 3a: Per message (metered)
price_messages_overage := stripe.Price{
    ID:       "price_messages_overage",
    Product:  "prod_messages_overage",
    Currency: "usd",
    Recurring: {
        Interval:       "month",
        UsageType:      "metered",
        AggregateUsage: "sum",
    },
    UnitAmount: 1, // $0.01 per message
}
```

### Domain: Plan Limits

```go
// internal/domain/billing/plan.go (NEW)

type Plan struct {
    id                uuid.UUID
    name              string
    stripePriceID     string
    maxContacts       int
    maxMessagesMonth  int
    maxProjects       int
    aiEnabled         bool
    analyticsEnabled  bool
    priceMonthly      int64 // cents
    priceAnnual       int64 // cents
}

var (
    PlanBasic = Plan{
        name:             "Basic",
        stripePriceID:    "price_basic_monthly",
        maxContacts:      1000,
        maxMessagesMonth: 5000,
        maxProjects:      1,
        aiEnabled:        false,
        analyticsEnabled: false,
        priceMonthly:     2900,
        priceAnnual:      29000,
    }

    PlanPro = Plan{
        name:             "Pro",
        stripePriceID:    "price_pro_monthly",
        maxContacts:      10000,
        maxMessagesMonth: 50000,
        maxProjects:      5,
        aiEnabled:        true,
        analyticsEnabled: true,
        priceMonthly:     9900,
        priceAnnual:      99000,
    }

    PlanEnterprise = Plan{
        name:             "Enterprise",
        stripePriceID:    "price_enterprise_monthly",
        maxContacts:      -1, // unlimited
        maxMessagesMonth: -1,
        maxProjects:      -1,
        aiEnabled:        true,
        analyticsEnabled: true,
        priceMonthly:     49900,
        priceAnnual:      499000,
    }
)

// Check if account can perform action
func (a *BillingAccount) CanSendMessage() bool {
    subscription := a.ActiveSubscription()
    if subscription == nil {
        return false
    }

    usage := a.CurrentMonthUsage()
    plan := subscription.Plan()

    // Unlimited plan
    if plan.maxMessagesMonth == -1 {
        return true
    }

    // Check if within limit
    return usage.MessagesSent < plan.maxMessagesMonth
}
```

---

## API Examples

### Create Billing Account

```http
POST /api/v1/billing/accounts
{
  "name": "Acme Corporation",
  "billing_email": "billing@acme.com"
}

Response:
{
  "id": "uuid",
  "name": "Acme Corporation",
  "billing_email": "billing@acme.com",
  "payment_status": "pending",
  "stripe_customer_id": "cus_XXXXXX",
  "created_at": "2025-10-10T15:00:00Z"
}
```

### Add Payment Method (with 3D Secure)

```http
POST /api/v1/billing/accounts/{id}/payment-methods
{
  "setup_intent_id": "seti_XXXXXX"  // From Stripe.js frontend
}

Response:
{
  "success": true,
  "payment_method": {
    "type": "card",
    "last_digits": "4242",
    "expires_at": "2026-12-31T23:59:59Z",
    "is_default": true
  }
}
```

### Create Subscription

```http
POST /api/v1/billing/accounts/{id}/subscriptions
{
  "price_id": "price_pro_monthly",
  "trial_days": 14
}

Response:
{
  "id": "uuid",
  "stripe_subscription_id": "sub_XXXXXX",
  "status": "trialing",
  "current_period_start": "2025-10-10T15:00:00Z",
  "current_period_end": "2025-11-10T15:00:00Z",
  "trial_end": "2025-10-24T15:00:00Z"
}
```

### Record Usage

```http
POST /api/v1/billing/accounts/{id}/usage
{
  "metric": "messages_sent",
  "quantity": 100,
  "timestamp": "2025-10-10T15:00:00Z"
}

Response:
{
  "success": true,
  "total_usage_this_period": 1250,
  "limit": 5000,
  "percentage_used": 25.0
}
```

### Get Invoices

```http
GET /api/v1/billing/accounts/{id}/invoices?limit=10

Response:
{
  "invoices": [
    {
      "id": "uuid",
      "stripe_invoice_id": "in_XXXXXX",
      "amount_due": 9900,
      "amount_paid": 9900,
      "currency": "usd",
      "status": "paid",
      "hosted_invoice_url": "https://invoice.stripe.com/i/...",
      "invoice_pdf": "https://invoice.stripe.com/i/.../pdf",
      "due_date": "2025-11-10T15:00:00Z",
      "paid_at": "2025-10-10T15:00:00Z"
    }
  ],
  "total": 12
}
```

---

## Performance Considerations

### Indexes

```sql
-- Billing Accounts
CREATE INDEX idx_billing_accounts_user ON billing_accounts(user_id);
CREATE INDEX idx_billing_accounts_stripe ON billing_accounts(stripe_customer_id);
CREATE INDEX idx_billing_accounts_status ON billing_accounts(payment_status);
CREATE INDEX idx_billing_accounts_suspended ON billing_accounts(suspended, suspended_at);

-- Subscriptions
CREATE INDEX idx_subscriptions_billing_account ON subscriptions(billing_account_id);
CREATE INDEX idx_subscriptions_stripe ON subscriptions(stripe_subscription_id);
CREATE INDEX idx_subscriptions_status ON subscriptions(status);
CREATE INDEX idx_subscriptions_period ON subscriptions(current_period_end) WHERE status = 'active';

-- Invoices
CREATE INDEX idx_invoices_billing_account ON invoices(billing_account_id);
CREATE INDEX idx_invoices_stripe ON invoices(stripe_invoice_id);
CREATE INDEX idx_invoices_status ON invoices(status);
CREATE INDEX idx_invoices_due_date ON invoices(due_date) WHERE status = 'open';
```

### Caching Strategy

```go
// Cache billing account by user (10 min TTL)
cacheKey := fmt.Sprintf("billing:user:%s", userID)
billingAccount, err := cache.Get(cacheKey)

// Cache active subscription (5 min TTL)
cacheKey := fmt.Sprintf("subscription:active:%s", billingAccountID)
subscription, err := cache.Get(cacheKey)

// Cache current usage (1 min TTL)
cacheKey := fmt.Sprintf("usage:current:%s", billingAccountID)
usage, err := cache.Get(cacheKey)
```

---

## Security & Compliance

### PCI-DSS Compliance

```go
// ✅ DO: Use Stripe.js / Payment Element (client-side)
// Card data NEVER touches your servers

// ❌ DON'T: Handle raw card numbers on backend
```

### Webhook Signature Verification

```go
// CRITICAL: Always verify webhook signatures
func VerifyStripeWebhook(payload []byte, signature string) (stripe.Event, error) {
    event, err := webhook.ConstructEvent(
        payload,
        signature,
        webhookSecret, // From environment variable
    )

    if err != nil {
        return stripe.Event{}, errors.New("invalid signature")
    }

    return event, nil
}
```

### Idempotency

```go
// Stripe API calls with idempotency keys
stripeClient.Subscriptions.New(&stripe.SubscriptionParams{
    IdempotencyKey: stripe.String(idempotencyKey), // UUID
    Customer:       stripe.String(customerID),
    // ...
})
```

---

## Implementation Status

### ✅ What's Implemented

1. **BillingAccount aggregate** - Basic lifecycle
2. **PaymentMethod value object** - Store payment method details
3. **Domain events** (5 events) - Account lifecycle
4. **Repository interface** - CRUD operations
5. **Unit tests** (~60% coverage)

### ❌ What's Missing (Critical)

1. **Stripe Integration** - No Stripe API client
2. **Subscription entity** - Not implemented
3. **Invoice entity** - Not implemented
4. **UsageMeter entity** - Not implemented
5. **PaymentIntent entity** - Not implemented
6. **Webhook handler** - No webhook receiver
7. **Dunning workflow** - No failed payment handling
8. **Plan limits** - No usage enforcement
9. **Proration** - No mid-cycle upgrades
10. **Multi-currency** - No currency support

---

## Suggested Implementation Roadmap

### Phase 1: Stripe Foundation (1 week)
- [ ] Install Stripe Go SDK
- [ ] Create Stripe API client wrapper
- [ ] Implement webhook handler with signature verification
- [ ] Create Customer on BillingAccount creation
- [ ] Implement SetupIntent for payment methods
- [ ] Handle 3D Secure authentication

### Phase 2: Subscription Management (1 week)
- [ ] Create Subscription entity
- [ ] Implement subscription CRUD
- [ ] Handle subscription lifecycle webhooks
- [ ] Create Invoice entity
- [ ] Handle invoice webhooks
- [ ] Generate invoice PDFs

### Phase 3: Usage-Based Billing (1 week)
- [ ] Create UsageMeter entity
- [ ] Implement real-time usage tracking
- [ ] Create Stripe Meters
- [ ] Report usage to Stripe
- [ ] Handle metered billing

### Phase 4: Dunning & Lifecycle (1 week)
- [ ] Create Temporal dunning workflow
- [ ] Implement retry logic (3, 7, 14 days)
- [ ] Email notifications for failed payments
- [ ] Account suspension/reactivation
- [ ] Subscription cancellation

### Phase 5: Advanced Features (1 week)
- [ ] Plan upgrades/downgrades with proration
- [ ] Annual subscriptions with discounts
- [ ] Multi-currency support
- [ ] Usage alerts (80%, 100% of limit)
- [ ] Admin dashboard for billing

---

## References

- [Billing Domain](../../internal/domain/billing/)
- [Billing Events](../../internal/domain/billing/events.go)
- [Billing Repository](../../internal/domain/billing/repository.go)
- [Stripe API Documentation](https://stripe.com/docs/api)
- [Stripe Billing Guide](https://stripe.com/docs/billing)
- [Stripe Webhooks](https://stripe.com/docs/webhooks)
- [Stripe 3D Secure](https://stripe.com/docs/payments/3d-secure)
- [Stripe Usage-Based Billing](https://stripe.com/docs/billing/subscriptions/usage-based)

---

**Next**: [Webhook Aggregate](webhook_aggregate.md) →
**Previous**: [Credential Aggregate](credential_aggregate.md) ←

---

## Summary

✅ **BillingAccount Current State**:
1. **Basic aggregate** - Account creation, payment methods, suspension
2. **5 domain events** - Account lifecycle events
3. **Simple status machine** - pending → active → suspended → canceled

❌ **Critical Missing Components**:
1. **Stripe Integration** - No API integration yet
2. **Subscription Management** - No subscription entity
3. **Usage-Based Billing** - No metering
4. **Webhook Handling** - No Stripe webhook receiver
5. **3D Secure / SCA** - No strong authentication

**Stripe Architecture Needed**:
- **Payment Intents** - 3D Secure compliant payments
- **Setup Intents** - Save payment methods for future use
- **Subscriptions** - Recurring billing with flexible billing mode (2025 API)
- **Usage Meters** - Real-time usage tracking (100M events/month)
- **Invoices** - Automatic invoice generation
- **Webhooks** - 20+ critical events to handle
- **Dunning** - Failed payment retry workflow

**Next Steps**: Implement complete Stripe integration following the architecture outlined above, including Payment Intents, Subscriptions, Usage Metering, and Webhook handling for production-ready SaaS billing.
