# ChannelType Aggregate

**Last Updated**: 2025-10-10
**Status**: ✅ Complete and Production-Ready
**Lines of Code**: ~200
**Test Coverage**: Partial

---

## Overview

- **Purpose**: Defines available communication channel types
- **Location**: `internal/domain/channel_type/`
- **Entity**: `infrastructure/persistence/entities/channel_type.go`
- **Repository**: `infrastructure/persistence/gorm_channel_type_repository.go`
- **Aggregate Root**: `ChannelType`

**Business Problem**:
The ChannelType aggregate represents the **catalog of available communication platforms** in the CRM system. It defines which channel types are supported (WhatsApp, Telegram, Instagram, etc.) and their configuration requirements. Critical for:
- **Platform discovery** - List available integration options
- **Configuration validation** - Define required config for each type
- **Provider grouping** - Group channels by provider (Meta, Telegram, WAHA)
- **Feature toggling** - Enable/disable channel types system-wide
- **UI rendering** - Display channel options in setup wizards

---

## Domain Model

### Aggregate Root: ChannelType

```go
type ChannelType struct {
    id            int
    name          string                 // "waha", "whatsapp", "direct_ig", etc.
    description   string                 // Human-readable description
    provider      string                 // "waha", "meta", "telegram"
    configuration map[string]interface{} // Default config schema
    active        bool                   // System-wide enable/disable
    createdAt     time.Time
    updatedAt     time.Time
}
```

### Predefined Channel Types

```go
const (
    WAHA      = 1  // WhatsApp HTTP API (multi-device)
    WhatsApp  = 2  // Official WhatsApp Business API
    DirectIG  = 3  // Instagram Direct Messages
    Messenger = 4  // Facebook Messenger
    Telegram  = 5  // Telegram Bot API
)

var Names = map[int]string{
    WAHA:      "waha",
    WhatsApp:  "whatsapp",
    DirectIG:  "direct_ig",
    Messenger: "messenger",
    Telegram:  "telegram",
}
```

### Value Objects

#### Provider

```go
type Provider string
const (
    ProviderWAHA     Provider = "waha"     // WAHA (self-hosted)
    ProviderMeta     Provider = "meta"     // Meta platforms (WhatsApp, IG, Messenger)
    ProviderTelegram Provider = "telegram" // Telegram
)
```

### Business Invariants

1. **ID must be positive**
   - IDs are predefined constants (1-5)
   - New types added with sequential IDs

2. **Name and Provider required**
   - `name` cannot be empty
   - `provider` cannot be empty
   - Names must be lowercase with underscores

3. **Active state controls availability**
   - Inactive channel types cannot be used for new channels
   - Existing channels remain functional if type deactivated

4. **Configuration schema**
   - Each type defines its own config requirements
   - Config stored as JSON map

---

## Events Emitted

| Event | When | Purpose |
|-------|------|---------|
| `channel_type.created` | New channel type added | Register type |
| `channel_type.activated` | Channel type enabled | Allow creation |
| `channel_type.deactivated` | Channel type disabled | Block creation |

---

## Repository Interface

```go
type Repository interface {
    Save(ctx context.Context, ct *ChannelType) error
    FindByID(ctx context.Context, id int) (*ChannelType, error)
    FindByName(ctx context.Context, name string) (*ChannelType, error)
    FindActive(ctx context.Context) ([]*ChannelType, error)
    FindAll(ctx context.Context) ([]*ChannelType, error)
}
```

---

## Commands (CQRS)

### ✅ Implemented

1. **CreateChannelTypeCommand** - Add new channel type
2. **ActivateChannelTypeCommand** - Enable channel type
3. **DeactivateChannelTypeCommand** - Disable channel type
4. **UpdateConfigurationCommand** - Update default config

### ❌ Suggested

- **ValidateChannelConfigCommand** - Validate config against schema
- **MigrateChannelTypeCommand** - Migrate channels to new type

---

## Use Cases

### ✅ Implemented

1. **GetChannelTypeUseCase** - Retrieve by ID
2. **GetChannelTypeByNameUseCase** - Retrieve by name
3. **ListChannelTypesUseCase** - List all or active only
4. **GetAvailableChannelTypesUseCase** - List predefined types

### ❌ Suggested

5. **ValidateChannelConfigUseCase** - Validate config before save
6. **SuggestChannelTypeUseCase** - Recommend type based on needs

---

## Channel Type Details

### 1. WAHA (ID: 1)

```go
{
    ID:          1,
    Name:        "waha",
    DisplayName: "WhatsApp (WAHA)",
    Description: "WhatsApp HTTP API - Multi-device support",
    Provider:    "waha",
}
```

**Use Cases**: Small businesses, personal WhatsApp accounts, self-hosted

**Configuration Requirements**:
```json
{
  "baseURL": "https://waha.example.com",
  "sessionID": "session-name",
  "auth": {
    "type": "basic",
    "username": "admin",
    "password": "secret"
  }
}
```

### 2. WhatsApp Business (ID: 2)

```go
{
    ID:          2,
    Name:        "whatsapp",
    DisplayName: "WhatsApp Business",
    Description: "Official WhatsApp Business API",
    Provider:    "meta",
}
```

**Use Cases**: Large businesses, verified business accounts, high volume

**Configuration Requirements**:
```json
{
  "phoneNumberID": "123456789",
  "businessAccountID": "987654321",
  "accessToken": "EAAxxxxxxxxxxxxx",
  "apiVersion": "v17.0"
}
```

### 3. Instagram Direct (ID: 3)

```go
{
    ID:          3,
    Name:        "direct_ig",
    DisplayName: "Instagram Direct",
    Description: "Instagram Direct Messages",
    Provider:    "meta",
}
```

**Use Cases**: Instagram business accounts, influencer support

**Configuration Requirements**:
```json
{
  "instagramAccountID": "123456789",
  "pageAccessToken": "EAAxxxxxxxxxxxxx",
  "webhookToken": "verify-token"
}
```

### 4. Facebook Messenger (ID: 4)

```go
{
    ID:          4,
    Name:        "messenger",
    DisplayName: "Facebook Messenger",
    Description: "Facebook Messenger Platform",
    Provider:    "meta",
}
```

**Use Cases**: Facebook business pages, community management

**Configuration Requirements**:
```json
{
  "pageID": "123456789012345",
  "pageAccessToken": "EAAxxxxxxxxxxxxx",
  "appSecret": "app-secret-key",
  "webhookToken": "verify-token"
}
```

### 5. Telegram (ID: 5)

```go
{
    ID:          5,
    Name:        "telegram",
    DisplayName: "Telegram",
    Description: "Telegram Bot API",
    Provider:    "telegram",
}
```

**Use Cases**: Telegram communities, bot-based support

**Configuration Requirements**:
```json
{
  "botToken": "123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11",
  "botUsername": "@MyCompanyBot",
  "webhookURL": "https://crm.example.com/webhooks/telegram/{channelID}"
}
```

---

## Provider Grouping

### Meta Provider (WhatsApp, Instagram, Messenger)

```go
func (ct *ChannelType) IsMeta() bool {
    return ct.provider == "meta"
}
```

**Common Features**:
- OAuth-based authentication
- Webhook verification with app secret
- Centralized Facebook Business Manager
- Similar API structure

**Shared Configuration**:
- `pageAccessToken` or `accessToken`
- `webhookToken` for verification
- `appSecret` for security

### WAHA Provider

**Features**:
- Self-hosted solution
- QR code authentication
- Session management
- Multi-device support

### Telegram Provider

**Features**:
- Bot token authentication
- Simple webhook setup
- Rich bot API features

---

## API Examples

### List Available Channel Types

```http
GET /api/v1/channel-types/available

Response:
{
  "channel_types": [
    {
      "id": 1,
      "name": "waha",
      "display_name": "WhatsApp (WAHA)",
      "description": "WhatsApp HTTP API - Multi-device support",
      "provider": "waha"
    },
    {
      "id": 2,
      "name": "whatsapp",
      "display_name": "WhatsApp Business",
      "description": "Official WhatsApp Business API",
      "provider": "meta"
    },
    {
      "id": 3,
      "name": "direct_ig",
      "display_name": "Instagram Direct",
      "description": "Instagram Direct Messages",
      "provider": "meta"
    },
    {
      "id": 4,
      "name": "messenger",
      "display_name": "Facebook Messenger",
      "description": "Facebook Messenger Platform",
      "provider": "meta"
    },
    {
      "id": 5,
      "name": "telegram",
      "display_name": "Telegram",
      "description": "Telegram Bot API",
      "provider": "telegram"
    }
  ]
}
```

### Get Channel Type by ID

```http
GET /api/v1/channel-types/1

Response:
{
  "id": 1,
  "name": "waha",
  "description": "WhatsApp HTTP API - Multi-device support",
  "provider": "waha",
  "configuration": {},
  "is_active": true,
  "created_at": "2025-01-01T00:00:00Z",
  "updated_at": "2025-01-01T00:00:00Z"
}
```

### Get Channel Type by Name

```http
GET /api/v1/channel-types/name/whatsapp

Response:
{
  "id": 2,
  "name": "whatsapp",
  "description": "Official WhatsApp Business API",
  "provider": "meta",
  "configuration": {},
  "is_active": true,
  "created_at": "2025-01-01T00:00:00Z",
  "updated_at": "2025-01-01T00:00:00Z"
}
```

### List Active Channel Types

```http
GET /api/v1/channel-types?active_only=true

Response:
{
  "channel_types": [
    {
      "id": 1,
      "name": "waha",
      "description": "WhatsApp HTTP API - Multi-device support",
      "provider": "waha",
      "is_active": true
    },
    {
      "id": 2,
      "name": "whatsapp",
      "description": "Official WhatsApp Business API",
      "provider": "meta",
      "is_active": true
    }
  ],
  "total": 2
}
```

### Deactivate Channel Type

```http
POST /api/v1/channel-types/3/deactivate

Response:
{
  "success": true,
  "channel_type_id": 3,
  "is_active": false,
  "deactivated_at": "2025-10-10T15:00:00Z"
}
```

---

## Real-World Usage

### Channel Setup Wizard

```
1. User clicks "Add Channel" in UI
2. Frontend calls GET /api/v1/channel-types/available
3. Display channel type options (WAHA, WhatsApp, Telegram, etc.)
4. User selects "WhatsApp (WAHA)"
5. UI shows config form based on channel type requirements:
   - Base URL
   - Session ID
   - Auth credentials
6. User submits form
7. Backend creates Channel entity with type_id=1
```

### Configuration Validation

```go
// Validate WAHA config
func ValidateWAHAConfig(config map[string]interface{}) error {
    required := []string{"baseURL", "sessionID", "auth"}
    for _, field := range required {
        if _, ok := config[field]; !ok {
            return fmt.Errorf("missing required field: %s", field)
        }
    }
    return nil
}

// Validate Meta config (WhatsApp, IG, Messenger)
func ValidateMetaConfig(config map[string]interface{}) error {
    if _, ok := config["accessToken"]; !ok {
        if _, ok := config["pageAccessToken"]; !ok {
            return fmt.Errorf("missing access token")
        }
    }
    return nil
}
```

---

## Performance Considerations

### Indexes

```sql
-- ChannelTypes table
CREATE INDEX idx_channel_types_name ON channel_types(name);
CREATE INDEX idx_channel_types_provider ON channel_types(provider);
CREATE INDEX idx_channel_types_active ON channel_types(active) WHERE active = true;
```

### Caching Strategy

```go
// Cache all channel types (1 hour TTL)
// Channel types rarely change
cacheKey := "channel_types:all"
types, err := cache.Get(cacheKey)
if err != nil {
    types, _ = repo.FindAll(ctx)
    cache.Set(cacheKey, types, 1*time.Hour)
}

// Cache active types (5 min TTL)
activeCacheKey := "channel_types:active"
activeTypes, err := cache.Get(activeCacheKey)
```

---

## Relationships

### ChannelType → Channel (One-to-Many)

```go
// Find all channels of a specific type
channels, _ := channelRepo.FindByType(ctx, channel_type.WAHA)

// Count channels by type
count, _ := channelRepo.CountByType(ctx, channel_type.WhatsApp)
```

---

## Extension Points

### Adding New Channel Types

```go
// 1. Add constant
const (
    WAHA      = 1
    WhatsApp  = 2
    DirectIG  = 3
    Messenger = 4
    Telegram  = 5
    SMS       = 6  // New type
)

// 2. Add name mapping
var Names = map[int]string{
    WAHA:      "waha",
    WhatsApp:  "whatsapp",
    DirectIG:  "direct_ig",
    Messenger: "messenger",
    Telegram:  "telegram",
    SMS:       "sms",  // New type
}

// 3. Add to GetAvailableChannelTypesUseCase
{
    ID:          SMS,
    Name:        "sms",
    DisplayName: "SMS",
    Description: "SMS messaging via Twilio",
    Provider:    "twilio",
}

// 4. Create database migration
INSERT INTO channel_types (id, name, description, provider, active)
VALUES (6, 'sms', 'SMS messaging via Twilio', 'twilio', true);
```

---

## References

- [ChannelType Domain](../../internal/domain/channel_type/)
- [ChannelType Events](../../internal/domain/channel_type/events.go)
- [ChannelType Repository](../../infrastructure/persistence/gorm_channel_type_repository.go)
- [Get ChannelType Use Case](../../internal/application/channel_type/get_channel_type_usecase.go)

---

**Next**: [Broadcast Aggregate](broadcast_aggregate.md) →
**Previous**: [Channel Aggregate](channel_aggregate.md) ←

---

## Summary

✅ **ChannelType Aggregate Features**:
1. Predefined catalog of 5 channel types
2. Provider grouping (Meta, WAHA, Telegram)
3. Configuration schema definition
4. Active/inactive state management
5. Simple CRUD operations
6. Extensible for new channel types

The ChannelType aggregate is a **catalog/reference data** aggregate that defines the available communication platforms in the Ventros CRM system.
