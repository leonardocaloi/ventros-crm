package tracking

import "fmt"

type UTMSourcePlatform string

const (
	PlatformMktDireto UTMSourcePlatform = "mkt-direto"

	UTMPlatformMeta     UTMSourcePlatform = "meta"
	UTMPlatformGoogle   UTMSourcePlatform = "google"
	UTMPlatformTikTok   UTMSourcePlatform = "tiktok"
	UTMPlatformLinkedIn UTMSourcePlatform = "linkedin"

	UTMPlatformOffline UTMSourcePlatform = "offline"

	UTMPlatformOther UTMSourcePlatform = "other"
)

type UTMSourceMeta string

const (
	MetaFacebook        UTMSourceMeta = "facebook"
	MetaInstagram       UTMSourceMeta = "instagram"
	MetaMessenger       UTMSourceMeta = "messenger"
	MetaAudienceNetwork UTMSourceMeta = "audience-network"
)

type UTMSourceGoogle string

const (
	GoogleSearch  UTMSourceGoogle = "search"
	GoogleDisplay UTMSourceGoogle = "display"
	GoogleYouTube UTMSourceGoogle = "youtube"
	GoogleGmail   UTMSourceGoogle = "gmail"
)

type UTMSourceMktDireto string

const (
	MktInfluencer UTMSourceMktDireto = "influencer"
	MktDisparo    UTMSourceMktDireto = "disparo"
	MktAffiliate  UTMSourceMktDireto = "affiliate"
)

type UTMSourceOffline string

const (
	OfflineTV       UTMSourceOffline = "tv"
	OfflineImpresso UTMSourceOffline = "impresso"
	OfflineOutdoor  UTMSourceOffline = "outdoor"
	OfflineEvento   UTMSourceOffline = "evento"
)

type UTMMedium string

const (
	MediumPaidSocial    UTMMedium = "paid-social"
	MediumOrganicSocial UTMMedium = "organic-social"

	MediumPaidSearch    UTMMedium = "paid-search"
	MediumOrganicSearch UTMMedium = "organic"

	MediumDisplay UTMMedium = "display"
	MediumVideo   UTMMedium = "video"

	MediumEmail    UTMMedium = "email"
	MediumSMS      UTMMedium = "sms"
	MediumWhatsApp UTMMedium = "whatsapp"

	MediumDirect UTMMedium = "direct"

	MediumOffline UTMMedium = "offline"

	MediumReferral UTMMedium = "referral"
	MediumOther    UTMMedium = "other"
)

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

type UTMStandard struct {
	SourcePlatform UTMSourcePlatform
	Source         string
	Medium         UTMMedium

	Campaign        string
	MarketingTactic UTMMarketingTactic

	Term string

	Content        string
	CreativeFormat UTMCreativeFormat
}

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

	if err := u.validatePlatformMediumCompatibility(); err != nil {
		return err
	}

	return nil
}

func (u *UTMStandard) validatePlatformMediumCompatibility() error {
	switch u.SourcePlatform {
	case UTMPlatformMeta, UTMPlatformTikTok, UTMPlatformLinkedIn:
		if u.Medium != MediumPaidSocial && u.Medium != MediumOrganicSocial {
			return fmt.Errorf("platform %s should use paid-social or organic-social medium", u.SourcePlatform)
		}
	case UTMPlatformGoogle:
		if u.Medium != MediumPaidSearch && u.Medium != MediumOrganicSearch &&
			u.Medium != MediumDisplay && u.Medium != MediumVideo {
			return fmt.Errorf("platform google should use search, display or video mediums")
		}
	case PlatformMktDireto:
		if u.Medium != MediumEmail && u.Medium != MediumSMS && u.Medium != MediumWhatsApp {
			return fmt.Errorf("platform mkt-direto should use email, sms or whatsapp medium")
		}
	case UTMPlatformOffline:
		if u.Medium != MediumOffline {
			return fmt.Errorf("platform offline should use offline medium")
		}
	}

	return nil
}

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

func IsValidSource(platform UTMSourcePlatform, source string) bool {
	validSources := GetValidSourcesForPlatform(platform)
	for _, valid := range validSources {
		if valid == source {
			return true
		}
	}
	return false
}

func IsValidMedium(platform UTMSourcePlatform, medium UTMMedium) bool {
	validMediums := GetValidMediumsForPlatform(platform)
	for _, valid := range validMediums {
		if valid == medium {
			return true
		}
	}
	return false
}
