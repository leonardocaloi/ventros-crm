package tracking

import "fmt"

type TrackingBuilder struct {
	utm       *UTMStandard
	contactID string
	sessionID *string
	tenantID  string
	projectID string
	adID      string
	clickID   string
	metadata  map[string]interface{}
	errors    []error
}

func NewTrackingBuilder() *TrackingBuilder {
	return &TrackingBuilder{
		utm:      &UTMStandard{},
		metadata: make(map[string]interface{}),
		errors:   []error{},
	}
}

func (b *TrackingBuilder) WithContact(contactID, tenantID, projectID string) *TrackingBuilder {
	b.contactID = contactID
	b.tenantID = tenantID
	b.projectID = projectID
	return b
}

func (b *TrackingBuilder) WithSession(sessionID string) *TrackingBuilder {
	b.sessionID = &sessionID
	return b
}

func (b *TrackingBuilder) WithSourcePlatform(platform UTMSourcePlatform) *TrackingBuilder {
	b.utm.SourcePlatform = platform

	validPlatforms := []UTMSourcePlatform{
		PlatformMktDireto,
		UTMPlatformMeta,
		UTMPlatformGoogle,
		UTMPlatformTikTok,
		UTMPlatformLinkedIn,
		UTMPlatformOffline,
		UTMPlatformOther,
	}

	valid := false
	for _, p := range validPlatforms {
		if platform == p {
			valid = true
			break
		}
	}

	if !valid {
		b.errors = append(b.errors, fmt.Errorf("invalid source platform: %s", platform))
	}

	return b
}

func (b *TrackingBuilder) WithSource(source string) *TrackingBuilder {
	if b.utm.SourcePlatform == "" {
		b.errors = append(b.errors, fmt.Errorf("must set source_platform before source"))
		return b
	}

	if !IsValidSource(b.utm.SourcePlatform, source) {
		validSources := GetValidSourcesForPlatform(b.utm.SourcePlatform)
		b.errors = append(b.errors, fmt.Errorf("invalid source '%s' for platform '%s'. Valid sources: %v",
			source, b.utm.SourcePlatform, validSources))
	}

	b.utm.Source = source
	return b
}

func (b *TrackingBuilder) WithMedium(medium UTMMedium) *TrackingBuilder {
	if b.utm.SourcePlatform == "" {
		b.errors = append(b.errors, fmt.Errorf("must set source_platform before medium"))
		return b
	}

	if !IsValidMedium(b.utm.SourcePlatform, medium) {
		validMediums := GetValidMediumsForPlatform(b.utm.SourcePlatform)
		b.errors = append(b.errors, fmt.Errorf("invalid medium '%s' for platform '%s'. Valid mediums: %v",
			medium, b.utm.SourcePlatform, validMediums))
	}

	b.utm.Medium = medium
	return b
}

func (b *TrackingBuilder) WithCampaign(campaign string) *TrackingBuilder {
	if campaign == "" {
		b.errors = append(b.errors, fmt.Errorf("campaign cannot be empty"))
	}
	b.utm.Campaign = campaign
	return b
}

func (b *TrackingBuilder) WithMarketingTactic(tactic UTMMarketingTactic) *TrackingBuilder {
	b.utm.MarketingTactic = tactic
	return b
}

func (b *TrackingBuilder) WithTerm(term string) *TrackingBuilder {
	b.utm.Term = term
	return b
}

func (b *TrackingBuilder) WithContent(content string) *TrackingBuilder {
	b.utm.Content = content
	return b
}

func (b *TrackingBuilder) WithCreativeFormat(format UTMCreativeFormat) *TrackingBuilder {
	b.utm.CreativeFormat = format
	return b
}

func (b *TrackingBuilder) WithAdID(adID string) *TrackingBuilder {
	b.adID = adID

	if b.utm.Content == "" && adID != "" {
		b.utm.Content = "ad_id_" + adID
	}
	return b
}

func (b *TrackingBuilder) WithClickID(clickID string) *TrackingBuilder {
	b.clickID = clickID
	return b
}

func (b *TrackingBuilder) WithMetadata(key string, value interface{}) *TrackingBuilder {
	b.metadata[key] = value
	return b
}

func (b *TrackingBuilder) Validate() error {
	if b.contactID == "" {
		b.errors = append(b.errors, fmt.Errorf("contactID is required"))
	}
	if b.tenantID == "" {
		b.errors = append(b.errors, fmt.Errorf("tenantID is required"))
	}
	if b.projectID == "" {
		b.errors = append(b.errors, fmt.Errorf("projectID is required"))
	}

	if err := b.utm.Validate(); err != nil {
		b.errors = append(b.errors, err)
	}

	if len(b.errors) > 0 {
		errorMsg := "validation errors: "
		for i, err := range b.errors {
			if i > 0 {
				errorMsg += "; "
			}
			errorMsg += err.Error()
		}
		return fmt.Errorf("%s", errorMsg)
	}

	return nil
}

func (b *TrackingBuilder) Build() (*UTMStandard, map[string]interface{}, error) {
	if err := b.Validate(); err != nil {
		return nil, nil, err
	}

	result := make(map[string]interface{})
	for k, v := range b.metadata {
		result[k] = v
	}

	result["contact_id"] = b.contactID
	result["tenant_id"] = b.tenantID
	result["project_id"] = b.projectID

	if b.sessionID != nil {
		result["session_id"] = *b.sessionID
	}

	if b.adID != "" {
		result["ad_id"] = b.adID
	}

	if b.clickID != "" {
		result["click_id"] = b.clickID
	}

	return b.utm, result, nil
}

func (b *TrackingBuilder) BuildURL(baseURL string) (string, error) {
	if err := b.Validate(); err != nil {
		return "", err
	}

	url := baseURL
	if len(url) == 0 {
		return "", fmt.Errorf("base URL cannot be empty")
	}

	separator := "?"
	if contains(url, "?") {
		separator = "&"
	}

	params := []string{
		fmt.Sprintf("utm_source_platform=%s", b.utm.SourcePlatform),
		fmt.Sprintf("utm_source=%s", b.utm.Source),
		fmt.Sprintf("utm_medium=%s", b.utm.Medium),
		fmt.Sprintf("utm_campaign=%s", b.utm.Campaign),
	}

	if b.utm.MarketingTactic != "" {
		params = append(params, fmt.Sprintf("utm_marketing_tactic=%s", b.utm.MarketingTactic))
	}
	if b.utm.Term != "" {
		params = append(params, fmt.Sprintf("utm_term=%s", b.utm.Term))
	}
	if b.utm.Content != "" {
		params = append(params, fmt.Sprintf("utm_content=%s", b.utm.Content))
	}
	if b.utm.CreativeFormat != "" {
		params = append(params, fmt.Sprintf("utm_creative_format=%s", b.utm.CreativeFormat))
	}

	for i, param := range params {
		if i == 0 {
			url += separator + param
		} else {
			url += "&" + param
		}
	}

	return url, nil
}

func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 &&
		(len(s) >= len(substr) && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
