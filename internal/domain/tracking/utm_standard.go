package tracking

import "fmt"

// UTMSourcePlatform representa a plataforma macro de onde vem o tráfego
type UTMSourcePlatform string

const (
	// Plataformas de marketing direto
	PlatformMktDireto UTMSourcePlatform = "mkt-direto"

	// Redes sociais e ecossistemas
	UTMPlatformMeta     UTMSourcePlatform = "meta"     // Facebook, Instagram, Messenger, Audience Network
	UTMPlatformGoogle   UTMSourcePlatform = "google"   // Search, Display, YouTube, Gmail
	UTMPlatformTikTok   UTMSourcePlatform = "tiktok"   // TikTok Ads
	UTMPlatformLinkedIn UTMSourcePlatform = "linkedin" // LinkedIn Ads

	// Canais offline
	UTMPlatformOffline UTMSourcePlatform = "offline" // TV, impresso, outdoor, eventos

	// Outros
	UTMPlatformOther UTMSourcePlatform = "other"
)

// UTMSourceMeta representa fontes específicas dentro do ecossistema Meta
type UTMSourceMeta string

const (
	MetaFacebook        UTMSourceMeta = "facebook"
	MetaInstagram       UTMSourceMeta = "instagram"
	MetaMessenger       UTMSourceMeta = "messenger"
	MetaAudienceNetwork UTMSourceMeta = "audience-network"
)

// UTMSourceGoogle representa fontes específicas dentro do ecossistema Google
type UTMSourceGoogle string

const (
	GoogleSearch  UTMSourceGoogle = "search"
	GoogleDisplay UTMSourceGoogle = "display"
	GoogleYouTube UTMSourceGoogle = "youtube"
	GoogleGmail   UTMSourceGoogle = "gmail"
)

// UTMSourceMktDireto representa fontes de marketing direto
type UTMSourceMktDireto string

const (
	MktInfluencer UTMSourceMktDireto = "influencer"
	MktDisparo    UTMSourceMktDireto = "disparo"
	MktAffiliate  UTMSourceMktDireto = "affiliate"
)

// UTMSourceOffline representa fontes offline
type UTMSourceOffline string

const (
	OfflineTV       UTMSourceOffline = "tv"
	OfflineImpresso UTMSourceOffline = "impresso"
	OfflineOutdoor  UTMSourceOffline = "outdoor"
	OfflineEvento   UTMSourceOffline = "evento"
)

// UTMMedium representa como o tráfego chega
type UTMMedium string

const (
	// Social media
	MediumPaidSocial    UTMMedium = "paid-social"
	MediumOrganicSocial UTMMedium = "organic-social"

	// Search
	MediumPaidSearch    UTMMedium = "paid-search"
	MediumOrganicSearch UTMMedium = "organic"

	// Display & Video
	MediumDisplay UTMMedium = "display"
	MediumVideo   UTMMedium = "video"

	// Messaging
	MediumEmail    UTMMedium = "email"
	MediumSMS      UTMMedium = "sms"
	MediumWhatsApp UTMMedium = "whatsapp"

	// Direct
	MediumDirect UTMMedium = "direct"

	// Offline
	MediumOffline UTMMedium = "offline"

	// Other
	MediumReferral UTMMedium = "referral"
	MediumOther    UTMMedium = "other"
)

// UTMMarketingTactic representa a tática ou abordagem de marketing
type UTMMarketingTactic string

const (
	TacticProspecting       UTMMarketingTactic = "prospecting"
	TacticRemarketing       UTMMarketingTactic = "remarketing"
	TacticDisparo           UTMMarketingTactic = "disparo"
	TacticCommentAutomation UTMMarketingTactic = "comment_automation"
	TacticAbandonedCart     UTMMarketingTactic = "abandoned_cart"
	TacticUpsell            UTMMarketingTactic = "upsell"
	TacticCrossSell         UTMMarketingTactic = "cross_sell"
	TacticRetention         UTMMarketingTactic = "retention"
	TacticReactivation      UTMMarketingTactic = "reactivation"
)

// UTMCreativeFormat representa o formato do criativo
type UTMCreativeFormat string

const (
	FormatCarrossel         UTMCreativeFormat = "carrossel"
	FormatBannerEstatico    UTMCreativeFormat = "banner_estatico"
	FormatVideo             UTMCreativeFormat = "video"
	FormatVideoPreRoll      UTMCreativeFormat = "video_pre_roll"
	FormatStoriesFullscreen UTMCreativeFormat = "stories_fullscreen"
	FormatStories           UTMCreativeFormat = "stories"
	FormatReels             UTMCreativeFormat = "reels"
	FormatImagemUnica       UTMCreativeFormat = "imagem_unica"
	FormatTexto             UTMCreativeFormat = "texto"
	FormatCollection        UTMCreativeFormat = "collection"
)

// UTMStandard representa o modelo completo e padronizado de UTMs
type UTMStandard struct {
	// Hierarquia principal
	SourcePlatform UTMSourcePlatform // Plataforma macro (obrigatório)
	Source         string            // Fonte específica dentro da plataforma (obrigatório)
	Medium         UTMMedium         // Como o tráfego chega (obrigatório)

	// Detalhes da campanha
	Campaign        string             // Identificador da campanha (obrigatório)
	MarketingTactic UTMMarketingTactic // Tática de marketing (opcional)

	// Segmentação e targeting
	Term string // Palavra-chave ou público-alvo (opcional)

	// Detalhes do criativo
	Content        string            // ID do anúncio, criativo, disparo (opcional)
	CreativeFormat UTMCreativeFormat // Formato do criativo (opcional)
}

// Validate valida se os campos obrigatórios estão preenchidos e se são compatíveis
func (u *UTMStandard) Validate() error {
	if u.SourcePlatform == "" {
		return fmt.Errorf("utm_source_platform is required")
	}
	if u.Source == "" {
		return fmt.Errorf("utm_source is required")
	}
	if u.Medium == "" {
		return fmt.Errorf("utm_medium is required")
	}
	if u.Campaign == "" {
		return fmt.Errorf("utm_campaign is required")
	}

	// Valida compatibilidade entre source_platform e medium
	if err := u.validatePlatformMediumCompatibility(); err != nil {
		return err
	}

	return nil
}

// validatePlatformMediumCompatibility valida se o medium é compatível com a plataforma
func (u *UTMStandard) validatePlatformMediumCompatibility() error {
	switch u.SourcePlatform {
	case UTMPlatformMeta, UTMPlatformTikTok, UTMPlatformLinkedIn:
		// Social platforms devem usar social mediums
		if u.Medium != MediumPaidSocial && u.Medium != MediumOrganicSocial {
			return fmt.Errorf("platform %s should use paid-social or organic-social medium", u.SourcePlatform)
		}
	case UTMPlatformGoogle:
		// Google pode usar search, display, video
		if u.Medium != MediumPaidSearch && u.Medium != MediumOrganicSearch &&
			u.Medium != MediumDisplay && u.Medium != MediumVideo {
			return fmt.Errorf("platform google should use search, display or video mediums")
		}
	case PlatformMktDireto:
		// Marketing direto pode usar email, sms, whatsapp
		if u.Medium != MediumEmail && u.Medium != MediumSMS && u.Medium != MediumWhatsApp {
			return fmt.Errorf("platform mkt-direto should use email, sms or whatsapp medium")
		}
	case UTMPlatformOffline:
		// Offline deve usar offline medium
		if u.Medium != MediumOffline {
			return fmt.Errorf("platform offline should use offline medium")
		}
	}

	return nil
}

// GetValidSourcesForPlatform retorna as fontes válidas para uma plataforma
func GetValidSourcesForPlatform(platform UTMSourcePlatform) []string {
	switch platform {
	case UTMPlatformMeta:
		return []string{
			string(MetaFacebook),
			string(MetaInstagram),
			string(MetaMessenger),
			string(MetaAudienceNetwork),
		}
	case UTMPlatformGoogle:
		return []string{
			string(GoogleSearch),
			string(GoogleDisplay),
			string(GoogleYouTube),
			string(GoogleGmail),
		}
	case PlatformMktDireto:
		return []string{
			string(MktInfluencer),
			string(MktDisparo),
			string(MktAffiliate),
		}
	case UTMPlatformOffline:
		return []string{
			string(OfflineTV),
			string(OfflineImpresso),
			string(OfflineOutdoor),
			string(OfflineEvento),
		}
	case UTMPlatformTikTok:
		return []string{"tiktok"}
	case UTMPlatformLinkedIn:
		return []string{"linkedin"}
	default:
		return []string{}
	}
}

// GetValidMediumsForPlatform retorna os mediums válidos para uma plataforma
func GetValidMediumsForPlatform(platform UTMSourcePlatform) []UTMMedium {
	switch platform {
	case UTMPlatformMeta, UTMPlatformTikTok, UTMPlatformLinkedIn:
		return []UTMMedium{MediumPaidSocial, MediumOrganicSocial}
	case UTMPlatformGoogle:
		return []UTMMedium{MediumPaidSearch, MediumOrganicSearch, MediumDisplay, MediumVideo}
	case PlatformMktDireto:
		return []UTMMedium{MediumEmail, MediumSMS, MediumWhatsApp}
	case UTMPlatformOffline:
		return []UTMMedium{MediumOffline}
	default:
		return []UTMMedium{MediumOther}
	}
}

// IsValidSource verifica se a source é válida para a plataforma
func IsValidSource(platform UTMSourcePlatform, source string) bool {
	validSources := GetValidSourcesForPlatform(platform)
	for _, valid := range validSources {
		if valid == source {
			return true
		}
	}
	return false
}

// IsValidMedium verifica se o medium é válido para a plataforma
func IsValidMedium(platform UTMSourcePlatform, medium UTMMedium) bool {
	validMediums := GetValidMediumsForPlatform(platform)
	for _, valid := range validMediums {
		if valid == medium {
			return true
		}
	}
	return false
}
