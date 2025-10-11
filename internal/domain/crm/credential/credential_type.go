package credential

type CredentialType string

const (
	CredentialTypeMetaWhatsApp    CredentialType = "meta_whatsapp_cloud"
	CredentialTypeMetaAds         CredentialType = "meta_ads"
	CredentialTypeMetaConversions CredentialType = "meta_conversions_api"

	CredentialTypeGoogleAds       CredentialType = "google_ads"
	CredentialTypeGoogleAnalytics CredentialType = "google_analytics"

	CredentialTypeWebhook   CredentialType = "webhook_auth"
	CredentialTypeAPIKey    CredentialType = "api_key"
	CredentialTypeBasicAuth CredentialType = "basic_auth"

	CredentialTypeWAHA CredentialType = "waha_instance"
)

func (t CredentialType) IsValid() bool {
	switch t {
	case CredentialTypeMetaWhatsApp,
		CredentialTypeMetaAds,
		CredentialTypeMetaConversions,
		CredentialTypeGoogleAds,
		CredentialTypeGoogleAnalytics,
		CredentialTypeWebhook,
		CredentialTypeAPIKey,
		CredentialTypeBasicAuth,
		CredentialTypeWAHA:
		return true
	default:
		return false
	}
}

func (t CredentialType) RequiresOAuth() bool {
	switch t {
	case CredentialTypeMetaWhatsApp,
		CredentialTypeMetaAds,
		CredentialTypeMetaConversions,
		CredentialTypeGoogleAds,
		CredentialTypeGoogleAnalytics:
		return true
	default:
		return false
	}
}

func (t CredentialType) GetScopes() []string {
	switch t {
	case CredentialTypeMetaWhatsApp:
		return []string{
			"whatsapp_business_management",
			"whatsapp_business_messaging",
		}
	case CredentialTypeMetaAds, CredentialTypeMetaConversions:
		return []string{
			"ads_management",
			"ads_read",
		}
	case CredentialTypeGoogleAds:
		return []string{
			"https://www.googleapis.com/auth/adwords",
		}
	case CredentialTypeGoogleAnalytics:
		return []string{
			"https://www.googleapis.com/auth/analytics.readonly",
		}
	default:
		return []string{}
	}
}

func (t CredentialType) String() string {
	return string(t)
}
