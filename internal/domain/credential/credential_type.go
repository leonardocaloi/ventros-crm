package credential

// CredentialType representa os tipos de credenciais suportadas
type CredentialType string

const (
	// Meta Integrations
	CredentialTypeMetaWhatsApp    CredentialType = "meta_whatsapp_cloud"
	CredentialTypeMetaAds         CredentialType = "meta_ads"
	CredentialTypeMetaConversions CredentialType = "meta_conversions_api"

	// Google Integrations
	CredentialTypeGoogleAds       CredentialType = "google_ads"
	CredentialTypeGoogleAnalytics CredentialType = "google_analytics"

	// Other Integrations
	CredentialTypeWebhook   CredentialType = "webhook_auth"
	CredentialTypeAPIKey    CredentialType = "api_key"
	CredentialTypeBasicAuth CredentialType = "basic_auth"

	// Internal
	CredentialTypeWAHA CredentialType = "waha_instance"
)

// IsValid verifica se o tipo de credencial é válido
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

// RequiresOAuth verifica se o tipo requer OAuth
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

// GetScopes retorna os scopes OAuth necessários
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

// String retorna a representação em string do tipo
func (t CredentialType) String() string {
	return string(t)
}
