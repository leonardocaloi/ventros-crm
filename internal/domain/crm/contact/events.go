package contact

import (
	"time"

	"github.com/caloi/ventros-crm/internal/domain/core/shared"
	"github.com/google/uuid"
)

type DomainEvent = shared.DomainEvent

type ContactCreatedEvent struct {
	shared.BaseEvent
	ContactID uuid.UUID
	ProjectID uuid.UUID
	TenantID  string
	Name      string
	CreatedAt time.Time
}

func NewContactCreatedEvent(contactID, projectID uuid.UUID, tenantID, name string) ContactCreatedEvent {
	return ContactCreatedEvent{
		BaseEvent: shared.NewBaseEvent("contact.created", time.Now()),
		ContactID: contactID,
		ProjectID: projectID,
		TenantID:  tenantID,
		Name:      name,
		CreatedAt: time.Now(),
	}
}

func (e ContactCreatedEvent) AggregateID() uuid.UUID  { return e.ContactID }
func (e ContactCreatedEvent) GetTenantID() string     { return e.TenantID }
func (e ContactCreatedEvent) GetProjectID() uuid.UUID { return e.ProjectID }

type ContactUpdatedEvent struct {
	shared.BaseEvent
	ContactID uuid.UUID
	UpdatedAt time.Time
}

func NewContactUpdatedEvent(contactID uuid.UUID) ContactUpdatedEvent {
	return ContactUpdatedEvent{
		BaseEvent: shared.NewBaseEvent("contact.updated", time.Now()),
		ContactID: contactID,
		UpdatedAt: time.Now(),
	}
}

type ContactProfilePictureUpdatedEvent struct {
	shared.BaseEvent
	ContactID         uuid.UUID
	TenantID          string
	ProfilePictureURL string
	FetchedAt         time.Time
}

func NewContactProfilePictureUpdatedEvent(contactID uuid.UUID, tenantID, profilePictureURL string) ContactProfilePictureUpdatedEvent {
	return ContactProfilePictureUpdatedEvent{
		BaseEvent:         shared.NewBaseEvent("contact.profile_picture_updated", time.Now()),
		ContactID:         contactID,
		TenantID:          tenantID,
		ProfilePictureURL: profilePictureURL,
		FetchedAt:         time.Now(),
	}
}

type ContactDeletedEvent struct {
	shared.BaseEvent
	ContactID uuid.UUID
	DeletedAt time.Time
}

func NewContactDeletedEvent(contactID uuid.UUID) ContactDeletedEvent {
	return ContactDeletedEvent{
		BaseEvent: shared.NewBaseEvent("contact.deleted", time.Now()),
		ContactID: contactID,
		DeletedAt: time.Now(),
	}
}

type ContactMergedEvent struct {
	shared.BaseEvent
	PrimaryContactID uuid.UUID
	MergedContactIDs []uuid.UUID
	MergeStrategy    string
	MergedAt         time.Time
}

func NewContactMergedEvent(primaryContactID uuid.UUID, mergedContactIDs []uuid.UUID, mergeStrategy string) ContactMergedEvent {
	return ContactMergedEvent{
		BaseEvent:        shared.NewBaseEvent("contact.merged", time.Now()),
		PrimaryContactID: primaryContactID,
		MergedContactIDs: mergedContactIDs,
		MergeStrategy:    mergeStrategy,
		MergedAt:         time.Now(),
	}
}

type ContactEnrichedEvent struct {
	shared.BaseEvent
	ContactID        uuid.UUID
	EnrichmentSource string
	EnrichedData     map[string]interface{}
	EnrichedAt       time.Time
}

func NewContactEnrichedEvent(contactID uuid.UUID, enrichmentSource string, enrichedData map[string]interface{}) ContactEnrichedEvent {
	return ContactEnrichedEvent{
		BaseEvent:        shared.NewBaseEvent("contact.enriched", time.Now()),
		ContactID:        contactID,
		EnrichmentSource: enrichmentSource,
		EnrichedData:     enrichedData,
		EnrichedAt:       time.Now(),
	}
}

type ContactNameChangedEvent struct {
	shared.BaseEvent
	ContactID uuid.UUID
	TenantID  string
	OldName   string
	NewName   string
	ChangedBy uuid.UUID
}

func NewContactNameChangedEvent(contactID uuid.UUID, tenantID, oldName, newName string, changedBy uuid.UUID) ContactNameChangedEvent {
	return ContactNameChangedEvent{
		BaseEvent: shared.NewBaseEvent("contact.name_changed", time.Now()),
		ContactID: contactID,
		TenantID:  tenantID,
		OldName:   oldName,
		NewName:   newName,
		ChangedBy: changedBy,
	}
}

type ContactEmailSetEvent struct {
	shared.BaseEvent
	ContactID uuid.UUID
	TenantID  string
	Email     string
	Verified  bool
}

func NewContactEmailSetEvent(contactID uuid.UUID, tenantID, email string, verified bool) ContactEmailSetEvent {
	return ContactEmailSetEvent{
		BaseEvent: shared.NewBaseEvent("contact.email_set", time.Now()),
		ContactID: contactID,
		TenantID:  tenantID,
		Email:     email,
		Verified:  verified,
	}
}

type ContactPhoneSetEvent struct {
	shared.BaseEvent
	ContactID uuid.UUID
	TenantID  string
	Phone     string
	Verified  bool
}

func NewContactPhoneSetEvent(contactID uuid.UUID, tenantID, phone string, verified bool) ContactPhoneSetEvent {
	return ContactPhoneSetEvent{
		BaseEvent: shared.NewBaseEvent("contact.phone_set", time.Now()),
		ContactID: contactID,
		TenantID:  tenantID,
		Phone:     phone,
		Verified:  verified,
	}
}

type ContactTagAddedEvent struct {
	shared.BaseEvent
	ContactID uuid.UUID
	TenantID  string
	Tag       string
}

func NewContactTagAddedEvent(contactID uuid.UUID, tenantID, tag string) ContactTagAddedEvent {
	return ContactTagAddedEvent{
		BaseEvent: shared.NewBaseEvent("contact.tag_added", time.Now()),
		ContactID: contactID,
		TenantID:  tenantID,
		Tag:       tag,
	}
}

type ContactTagRemovedEvent struct {
	shared.BaseEvent
	ContactID uuid.UUID
	TenantID  string
	Tag       string
}

func NewContactTagRemovedEvent(contactID uuid.UUID, tenantID, tag string) ContactTagRemovedEvent {
	return ContactTagRemovedEvent{
		BaseEvent: shared.NewBaseEvent("contact.tag_removed", time.Now()),
		ContactID: contactID,
		TenantID:  tenantID,
		Tag:       tag,
	}
}

type ContactTagsClearedEvent struct {
	shared.BaseEvent
	ContactID   uuid.UUID
	TenantID    string
	ClearedTags []string
}

func NewContactTagsClearedEvent(contactID uuid.UUID, tenantID string, tags []string) ContactTagsClearedEvent {
	return ContactTagsClearedEvent{
		BaseEvent:   shared.NewBaseEvent("contact.tags_cleared", time.Now()),
		ContactID:   contactID,
		TenantID:    tenantID,
		ClearedTags: tags,
	}
}

type ContactExternalIDSetEvent struct {
	shared.BaseEvent
	ContactID  uuid.UUID
	TenantID   string
	ExternalID string
	Source     string
}

func NewContactExternalIDSetEvent(contactID uuid.UUID, tenantID, externalID, source string) ContactExternalIDSetEvent {
	return ContactExternalIDSetEvent{
		BaseEvent:  shared.NewBaseEvent("contact.external_id_set", time.Now()),
		ContactID:  contactID,
		TenantID:   tenantID,
		ExternalID: externalID,
		Source:     source,
	}
}

type ContactLanguageChangedEvent struct {
	shared.BaseEvent
	ContactID   uuid.UUID
	TenantID    string
	OldLanguage string
	NewLanguage string
}

func NewContactLanguageChangedEvent(contactID uuid.UUID, tenantID, oldLanguage, newLanguage string) ContactLanguageChangedEvent {
	return ContactLanguageChangedEvent{
		BaseEvent:   shared.NewBaseEvent("contact.language_changed", time.Now()),
		ContactID:   contactID,
		TenantID:    tenantID,
		OldLanguage: oldLanguage,
		NewLanguage: newLanguage,
	}
}

type ContactTimezoneSetEvent struct {
	shared.BaseEvent
	ContactID uuid.UUID
	TenantID  string
	Timezone  string
}

func NewContactTimezoneSetEvent(contactID uuid.UUID, tenantID, timezone string) ContactTimezoneSetEvent {
	return ContactTimezoneSetEvent{
		BaseEvent: shared.NewBaseEvent("contact.timezone_set", time.Now()),
		ContactID: contactID,
		TenantID:  tenantID,
		Timezone:  timezone,
	}
}

type ContactInteractionRecordedEvent struct {
	shared.BaseEvent
	ContactID          uuid.UUID
	TenantID           string
	InteractionType    string
	IsFirstInteraction bool
	InteractedAt       time.Time
}

func NewContactInteractionRecordedEvent(contactID uuid.UUID, tenantID, interactionType string, isFirst bool) ContactInteractionRecordedEvent {
	return ContactInteractionRecordedEvent{
		BaseEvent:          shared.NewBaseEvent("contact.interaction_recorded", time.Now()),
		ContactID:          contactID,
		TenantID:           tenantID,
		InteractionType:    interactionType,
		IsFirstInteraction: isFirst,
		InteractedAt:       time.Now(),
	}
}

type ContactSourceChannelSetEvent struct {
	shared.BaseEvent
	ContactID     uuid.UUID
	TenantID      string
	SourceChannel string
}

func NewContactSourceChannelSetEvent(contactID uuid.UUID, tenantID, sourceChannel string) ContactSourceChannelSetEvent {
	return ContactSourceChannelSetEvent{
		BaseEvent:     shared.NewBaseEvent("contact.source_channel_set", time.Now()),
		ContactID:     contactID,
		TenantID:      tenantID,
		SourceChannel: sourceChannel,
	}
}

type AdConversionTrackedEvent struct {
	shared.BaseEvent
	ContactID uuid.UUID
	SessionID uuid.UUID
	TenantID  string

	ConversionSource string
	ConversionApp    string

	AdSourceType string
	AdSourceID   string
	AdSourceApp  string
	AdSourceURL  string

	CTWAClickID string

	ConversionData string
	ExternalSource string
	ExternalMedium string

	TrackedAt time.Time
}

func NewAdConversionTrackedEvent(
	contactID uuid.UUID,
	sessionID uuid.UUID,
	tenantID string,
	trackingData map[string]string,
) AdConversionTrackedEvent {
	return AdConversionTrackedEvent{
		BaseEvent:        shared.NewBaseEvent("tracking.message.meta_ads", time.Now()),
		ContactID:        contactID,
		SessionID:        sessionID,
		TenantID:         tenantID,
		ConversionSource: trackingData["conversion_source"],
		ConversionApp:    trackingData["conversion_app"],
		AdSourceType:     trackingData["ad_source_type"],
		AdSourceID:       trackingData["ad_source_id"],
		AdSourceApp:      trackingData["ad_source_app"],
		AdSourceURL:      trackingData["ad_source_url"],
		CTWAClickID:      trackingData["ctwa_clid"],
		ConversionData:   trackingData["conversion_data"],
		ExternalSource:   trackingData["external_source"],
		ExternalMedium:   trackingData["external_medium"],
		TrackedAt:        time.Now(),
	}
}

func (e AdConversionTrackedEvent) ToContactEventPayload() map[string]interface{} {
	payload := make(map[string]interface{})

	if e.ConversionSource != "" {
		payload["conversion_source"] = e.ConversionSource
	}
	if e.ConversionApp != "" {
		payload["conversion_app"] = e.ConversionApp
	}
	if e.AdSourceType != "" {
		payload["ad_source_type"] = e.AdSourceType
	}
	if e.AdSourceID != "" {
		payload["ad_source_id"] = e.AdSourceID
	}
	if e.AdSourceApp != "" {
		payload["ad_source_app"] = e.AdSourceApp
	}
	if e.AdSourceURL != "" {
		payload["ad_source_url"] = e.AdSourceURL
	}
	if e.CTWAClickID != "" {
		payload["ctwa_click_id"] = e.CTWAClickID
	}

	return payload
}

func (e AdConversionTrackedEvent) GetTitle() string {
	if e.AdSourceApp != "" {
		return "Message from " + e.AdSourceApp + " ad"
	}
	return "Message from ad"
}

func (e AdConversionTrackedEvent) GetDescription() string {
	if e.AdSourceID != "" {
		return "Contact came from ad campaign (ID: " + e.AdSourceID + ")"
	}
	return "Contact came from ad campaign"
}

type ContactPipelineStatusChangedEvent struct {
	shared.BaseEvent
	ContactID          uuid.UUID
	PipelineID         uuid.UUID
	PreviousStatusID   *uuid.UUID
	NewStatusID        uuid.UUID
	PreviousStatusName string
	NewStatusName      string
	TenantID           string
	ProjectID          uuid.UUID
	ChangedBy          *uuid.UUID
	Reason             string
	ChangedAt          time.Time
}

func NewContactPipelineStatusChangedEvent(
	contactID uuid.UUID,
	pipelineID uuid.UUID,
	previousStatusID *uuid.UUID,
	newStatusID uuid.UUID,
	previousStatusName string,
	newStatusName string,
	tenantID string,
	projectID uuid.UUID,
	changedBy *uuid.UUID,
	reason string,
) ContactPipelineStatusChangedEvent {
	return ContactPipelineStatusChangedEvent{
		BaseEvent:          shared.NewBaseEvent("contact.pipeline_status_changed", time.Now()),
		ContactID:          contactID,
		PipelineID:         pipelineID,
		PreviousStatusID:   previousStatusID,
		NewStatusID:        newStatusID,
		PreviousStatusName: previousStatusName,
		NewStatusName:      newStatusName,
		TenantID:           tenantID,
		ProjectID:          projectID,
		ChangedBy:          changedBy,
		Reason:             reason,
		ChangedAt:          time.Now(),
	}
}

func (e ContactPipelineStatusChangedEvent) IsFirstStatus() bool {
	return e.PreviousStatusID == nil
}

func (e ContactPipelineStatusChangedEvent) ToContactEventPayload() map[string]interface{} {
	payload := map[string]interface{}{
		"pipeline_id":     e.PipelineID.String(),
		"new_status_id":   e.NewStatusID.String(),
		"new_status_name": e.NewStatusName,
	}

	if e.PreviousStatusID != nil {
		payload["previous_status_id"] = e.PreviousStatusID.String()
		payload["previous_status_name"] = e.PreviousStatusName
	}

	if e.ChangedBy != nil {
		payload["changed_by"] = e.ChangedBy.String()
	}

	if e.Reason != "" {
		payload["reason"] = e.Reason
	}

	return payload
}

func (e ContactPipelineStatusChangedEvent) GetTitle() string {
	if e.IsFirstStatus() {
		return "Entered pipeline: " + e.NewStatusName
	}
	return "Status changed: " + e.PreviousStatusName + " → " + e.NewStatusName
}

func (e ContactPipelineStatusChangedEvent) GetDescription() string {
	if e.IsFirstStatus() {
		return "Contact entered the pipeline with status: " + e.NewStatusName
	}
	return "Contact moved from " + e.PreviousStatusName + " to " + e.NewStatusName
}
