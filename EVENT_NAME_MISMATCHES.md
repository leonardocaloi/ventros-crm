# 🚨 CRITICAL: Event Name Mismatches

## Problem
Events defined in domain layer use different names than what's mapped in `domain_event_bus.go`.
This means these events will NEVER be sent to webhooks because they don't match!

## Mismatches Found

| # | Actual Event Name (Domain) | Mapped Name (EventBus) | Status |
|---|---------------------------|------------------------|--------|
| 1 | `credential.oauth_refreshed` | `credential.oauth_token_refreshed` | ❌ MISMATCH |
| 2 | `channel.pipeline.associated` | `channel.pipeline_associated` | ❌ MISMATCH |
| 3 | `status.created` | `pipeline.status.created` | ❌ MISMATCH |
| 4 | `status.updated` | `pipeline.status.updated` | ❌ MISMATCH |
| 5 | `status.activated` | `pipeline.status.activated` | ❌ MISMATCH |
| 6 | `status.deactivated` | `pipeline.status.deactivated` | ❌ MISMATCH |
| 7 | `pipeline.status_added` | `pipeline.status.added` | ❌ MISMATCH |
| 8 | `pipeline.status_removed` | `pipeline.status.removed` | ❌ MISMATCH |
| 9 | `automation_rule.triggered` | `automation.rule_triggered` | ❌ MISMATCH |
| 10 | `automation_rule.executed` | `automation.rule_executed` | ❌ MISMATCH |
| 11 | `automation_rule.failed` | `automation.rule_failed` | ❌ MISMATCH |
| 12 | `message.ai.process_image_requested` | `ai.process_image_requested` | ❌ MISMATCH |
| 13 | `message.ai.process_video_requested` | `ai.process_video_requested` | ❌ MISMATCH |
| 14 | `message.ai.process_audio_requested` | `ai.process_audio_requested` | ❌ MISMATCH |
| 15 | `message.ai.process_voice_requested` | `ai.process_voice_requested` | ❌ MISMATCH |

## Impact

**HIGH SEVERITY**: 15 out of ~90 events (16.7%) will NEVER reach webhooks due to these mismatches.

Affected features:
- ❌ Credential OAuth refresh notifications
- ❌ Pipeline status management events
- ❌ Automation rule execution tracking
- ❌ AI processing requests for media
- ❌ Channel-pipeline associations

## Solution

Fix the `mapDomainToBusinessEvents()` function in `/home/caloi/ventros-crm/infrastructure/messaging/domain_event_bus.go` to use the ACTUAL event names from the domain layer.

## Files to Update

1. `/home/caloi/ventros-crm/infrastructure/messaging/domain_event_bus.go` - Update mappings (lines 204-444)

## Verification Command

```bash
# Find all event names in domain
grep -r "NewBaseEvent(\"" internal/domain/ | sed 's/.*NewBaseEvent("\([^"]*\)".*/\1/' | sort | uniq

# Compare with mapped events in domain_event_bus.go
grep -A 1 "^[[:space:]]*case \"" infrastructure/messaging/domain_event_bus.go | grep "case" | sed 's/.*case "\([^"]*\)".*/\1/' | sort | uniq
```

## Date Found
2025-10-09

## Status
🔴 CRITICAL - Needs immediate fix
