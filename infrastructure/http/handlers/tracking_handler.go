package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ventros/crm/infrastructure/http/dto"
	"github.com/ventros/crm/internal/application/tracking"
	domainTracking "github.com/ventros/crm/internal/domain/crm/tracking"
	"go.uber.org/zap"
)

// TrackingHandler gerencia as rotas de trackings
type TrackingHandler struct {
	createUseCase              *tracking.CreateTrackingUseCase
	getUseCase                 *tracking.GetTrackingUseCase
	getContactTrackingsUseCase *tracking.GetContactTrackingsUseCase
	encodeUseCase              *tracking.EncodeTrackingUseCase
	decodeUseCase              *tracking.DecodeTrackingUseCase
	logger                     *zap.Logger
}

// NewTrackingHandler cria uma nova instância do handler
func NewTrackingHandler(
	createUseCase *tracking.CreateTrackingUseCase,
	getUseCase *tracking.GetTrackingUseCase,
	getContactTrackingsUseCase *tracking.GetContactTrackingsUseCase,
	logger *zap.Logger,
) *TrackingHandler {
	return &TrackingHandler{
		createUseCase:              createUseCase,
		getUseCase:                 getUseCase,
		getContactTrackingsUseCase: getContactTrackingsUseCase,
		encodeUseCase:              tracking.NewEncodeTrackingUseCase(logger),
		decodeUseCase:              tracking.NewDecodeTrackingUseCase(logger),
		logger:                     logger,
	}
}

// CreateTracking godoc
//
//	@Summary		Cria um novo tracking de conversão
//	@Description	Cria um novo registro de tracking para rastrear origem de contato (anúncios, campanhas, etc)
//	@Tags			CRM - Tracking
//	@Accept			json
//	@Produce		json
//	@Param			request	body		CreateTrackingRequest	true	"Dados do tracking"
//	@Success		201		{object}	tracking.TrackingDTO
//	@Failure		400		{object}	dto.ErrorResponse
//	@Failure		500		{object}	dto.ErrorResponse
//	@Router			/api/v1/trackings [post]
//	@Security		BearerAuth
func (h *TrackingHandler) CreateTracking(c *gin.Context) {
	var req CreateTrackingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	// Parse ContactID
	contactID, err := uuid.Parse(req.ContactID)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid contact_id"})
		return
	}

	// Parse SessionID se fornecido
	var sessionID *uuid.UUID
	if req.SessionID != "" {
		parsed, err := uuid.Parse(req.SessionID)
		if err != nil {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid session_id"})
			return
		}
		sessionID = &parsed
	}

	// Parse ProjectID
	projectID, err := uuid.Parse(req.ProjectID)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid project_id"})
		return
	}

	// Obter TenantID do contexto
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "tenant_id not found in context"})
		return
	}

	createDTO := tracking.CreateTrackingDTO{
		ContactID:      contactID,
		SessionID:      sessionID,
		TenantID:       tenantID.(string),
		ProjectID:      projectID,
		Source:         req.Source,
		Platform:       req.Platform,
		Campaign:       req.Campaign,
		AdID:           req.AdID,
		AdURL:          req.AdURL,
		ClickID:        req.ClickID,
		ConversionData: req.ConversionData,
		UTMSource:      req.UTMSource,
		UTMMedium:      req.UTMMedium,
		UTMCampaign:    req.UTMCampaign,
		UTMTerm:        req.UTMTerm,
		UTMContent:     req.UTMContent,
		Metadata:       req.Metadata,
	}

	result, err := h.createUseCase.Execute(c.Request.Context(), createDTO)
	if err != nil {
		h.logger.Error("Failed to create tracking", zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, result)
}

// GetTracking godoc
//
//	@Summary		Busca tracking por ID
//	@Description	Retorna um tracking específico por ID
//	@Tags			CRM - Tracking
//	@Produce		json
//	@Param			id	path		string	true	"ID do tracking"
//	@Success		200	{object}	tracking.TrackingDTO
//	@Failure		400	{object}	dto.ErrorResponse
//	@Failure		404	{object}	dto.ErrorResponse
//	@Router			/api/v1/trackings/{id} [get]
//	@Security		BearerAuth
func (h *TrackingHandler) GetTracking(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid tracking ID"})
		return
	}

	result, err := h.getUseCase.Execute(c.Request.Context(), id)
	if err != nil {
		if err.Error() == "tracking not found" {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: "tracking not found"})
			return
		}
		h.logger.Error("Failed to get tracking", zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetContactTrackings godoc
//
//	@Summary		Busca trackings de um contato
//	@Description	Retorna todos os trackings de um contato específico
//	@Tags			CRM - Tracking
//	@Produce		json
//	@Param			contact_id	path		string	true	"ID do contato"
//	@Success		200			{array}		tracking.TrackingDTO
//	@Failure		400			{object}	dto.ErrorResponse
//	@Failure		500			{object}	dto.ErrorResponse
//	@Router			/api/v1/contacts/{contact_id}/trackings [get]
//	@Security		BearerAuth
func (h *TrackingHandler) GetContactTrackings(c *gin.Context) {
	contactID, err := uuid.Parse(c.Param("contact_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid contact ID"})
		return
	}

	result, err := h.getContactTrackingsUseCase.Execute(c.Request.Context(), contactID)
	if err != nil {
		h.logger.Error("Failed to get contact trackings", zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetTrackingEnums godoc
//
//	@Summary		Lista todos os enums disponíveis para tracking
//	@Description	Retorna todos os valores válidos de enums para construir trackings (plataformas, mediums, táticas, formatos, etc)
//	@Tags			CRM - Tracking
//	@Produce		json
//	@Success		200	{object}	TrackingEnumsResponse
//	@Router			/api/v1/trackings/enums [get]
//	@Security		BearerAuth
func (h *TrackingHandler) GetTrackingEnums(c *gin.Context) {
	c.JSON(http.StatusOK, GetAllTrackingEnums())
}

// CreateTrackingRequest representa a requisição para criar tracking
type CreateTrackingRequest struct {
	ContactID      string                 `json:"contact_id" binding:"required"`
	SessionID      string                 `json:"session_id,omitempty"`
	ProjectID      string                 `json:"project_id" binding:"required"`
	Source         string                 `json:"source" binding:"required"`   // meta_ads, google_ads, etc
	Platform       string                 `json:"platform" binding:"required"` // instagram, facebook, etc
	Campaign       string                 `json:"campaign,omitempty"`
	AdID           string                 `json:"ad_id,omitempty"`
	AdURL          string                 `json:"ad_url,omitempty"`
	ClickID        string                 `json:"click_id,omitempty"`
	ConversionData string                 `json:"conversion_data,omitempty"`
	UTMSource      string                 `json:"utm_source,omitempty"`
	UTMMedium      string                 `json:"utm_medium,omitempty"`
	UTMCampaign    string                 `json:"utm_campaign,omitempty"`
	UTMTerm        string                 `json:"utm_term,omitempty"`
	UTMContent     string                 `json:"utm_content,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// TrackingEnumsResponse representa todos os enums disponíveis para tracking
type TrackingEnumsResponse struct {
	Platforms             []EnumValue                          `json:"platforms"`
	MetaSources           []EnumValue                          `json:"meta_sources"`
	GoogleSources         []EnumValue                          `json:"google_sources"`
	MktDiretoSources      []EnumValue                          `json:"mkt_direto_sources"`
	OfflineSources        []EnumValue                          `json:"offline_sources"`
	Mediums               []EnumValue                          `json:"mediums"`
	MarketingTactics      []EnumValue                          `json:"marketing_tactics"`
	CreativeFormats       []EnumValue                          `json:"creative_formats"`
	PlatformCompatibility map[string]PlatformCompatibilityInfo `json:"platform_compatibility"`
}

// EnumValue representa um valor de enum com metadados
type EnumValue struct {
	Value       string `json:"value" example:"meta"`
	Label       string `json:"label" example:"Meta (Facebook/Instagram)"`
	Description string `json:"description" example:"Facebook, Instagram, Messenger, Audience Network"`
}

// PlatformCompatibilityInfo descreve compatibilidades de uma plataforma
type PlatformCompatibilityInfo struct {
	ValidMediums []string `json:"valid_mediums"`
	Description  string   `json:"description"`
}

// GetAllTrackingEnums retorna todos os enums disponíveis
func GetAllTrackingEnums() TrackingEnumsResponse {
	return TrackingEnumsResponse{
		Platforms: []EnumValue{
			{Value: string(domainTracking.PlatformMktDireto), Label: "Marketing Direto", Description: "Influencers, disparos, afiliados"},
			{Value: string(domainTracking.UTMPlatformMeta), Label: "Meta (Facebook/Instagram)", Description: "Facebook, Instagram, Messenger, Audience Network"},
			{Value: string(domainTracking.UTMPlatformGoogle), Label: "Google", Description: "Search, Display, YouTube, Gmail"},
			{Value: string(domainTracking.UTMPlatformTikTok), Label: "TikTok", Description: "TikTok Ads"},
			{Value: string(domainTracking.UTMPlatformLinkedIn), Label: "LinkedIn", Description: "LinkedIn Ads"},
			{Value: string(domainTracking.UTMPlatformOffline), Label: "Offline", Description: "TV, impresso, outdoor, eventos"},
			{Value: string(domainTracking.UTMPlatformOther), Label: "Outros", Description: "Outras fontes"},
		},
		MetaSources: []EnumValue{
			{Value: string(domainTracking.MetaFacebook), Label: "Facebook", Description: "Facebook Ads"},
			{Value: string(domainTracking.MetaInstagram), Label: "Instagram", Description: "Instagram Ads"},
			{Value: string(domainTracking.MetaMessenger), Label: "Messenger", Description: "Messenger Ads"},
			{Value: string(domainTracking.MetaAudienceNetwork), Label: "Audience Network", Description: "Meta Audience Network"},
		},
		GoogleSources: []EnumValue{
			{Value: string(domainTracking.GoogleSearch), Label: "Google Search", Description: "Google Search Ads"},
			{Value: string(domainTracking.GoogleDisplay), Label: "Google Display", Description: "Google Display Network"},
			{Value: string(domainTracking.GoogleYouTube), Label: "YouTube", Description: "YouTube Ads"},
			{Value: string(domainTracking.GoogleGmail), Label: "Gmail", Description: "Gmail Ads"},
		},
		MktDiretoSources: []EnumValue{
			{Value: string(domainTracking.MktInfluencer), Label: "Influencer", Description: "Marketing de Influência"},
			{Value: string(domainTracking.MktDisparo), Label: "Disparo", Description: "Disparos em massa"},
			{Value: string(domainTracking.MktAffiliate), Label: "Affiliate", Description: "Marketing de Afiliados"},
		},
		OfflineSources: []EnumValue{
			{Value: string(domainTracking.OfflineTV), Label: "TV", Description: "Televisão"},
			{Value: string(domainTracking.OfflineImpresso), Label: "Impresso", Description: "Mídia impressa"},
			{Value: string(domainTracking.OfflineOutdoor), Label: "Outdoor", Description: "Outdoor/OOH"},
			{Value: string(domainTracking.OfflineEvento), Label: "Evento", Description: "Eventos presenciais"},
		},
		Mediums: []EnumValue{
			{Value: string(domainTracking.MediumPaidSocial), Label: "Paid Social", Description: "Mídia paga em redes sociais"},
			{Value: string(domainTracking.MediumOrganicSocial), Label: "Organic Social", Description: "Mídia orgânica em redes sociais"},
			{Value: string(domainTracking.MediumPaidSearch), Label: "Paid Search", Description: "Busca paga (SEM)"},
			{Value: string(domainTracking.MediumOrganicSearch), Label: "Organic Search", Description: "Busca orgânica (SEO)"},
			{Value: string(domainTracking.MediumDisplay), Label: "Display", Description: "Anúncios display/banner"},
			{Value: string(domainTracking.MediumVideo), Label: "Video", Description: "Anúncios em vídeo"},
			{Value: string(domainTracking.MediumEmail), Label: "Email", Description: "Email marketing"},
			{Value: string(domainTracking.MediumSMS), Label: "SMS", Description: "SMS marketing"},
			{Value: string(domainTracking.MediumWhatsApp), Label: "WhatsApp", Description: "WhatsApp marketing"},
			{Value: string(domainTracking.MediumDirect), Label: "Direct", Description: "Tráfego direto"},
			{Value: string(domainTracking.MediumOffline), Label: "Offline", Description: "Canais offline"},
			{Value: string(domainTracking.MediumReferral), Label: "Referral", Description: "Tráfego de referência"},
			{Value: string(domainTracking.MediumOther), Label: "Other", Description: "Outros canais"},
		},
		MarketingTactics: []EnumValue{
			{Value: string(domainTracking.TacticProspecting), Label: "Prospecting", Description: "Prospecção de novos clientes"},
			{Value: string(domainTracking.TacticRemarketing), Label: "Remarketing", Description: "Remarketing/Retargeting"},
			{Value: string(domainTracking.TacticDisparo), Label: "Disparo", Description: "Disparos em massa"},
			{Value: string(domainTracking.TacticCommentAutomation), Label: "Comment Automation", Description: "Automação de comentários"},
			{Value: string(domainTracking.TacticAbandonedCart), Label: "Abandoned Cart", Description: "Carrinho abandonado"},
			{Value: string(domainTracking.TacticUpsell), Label: "Upsell", Description: "Venda adicional"},
			{Value: string(domainTracking.TacticCrossSell), Label: "Cross-sell", Description: "Venda cruzada"},
			{Value: string(domainTracking.TacticRetention), Label: "Retention", Description: "Retenção de clientes"},
			{Value: string(domainTracking.TacticReactivation), Label: "Reactivation", Description: "Reativação de clientes"},
		},
		CreativeFormats: []EnumValue{
			{Value: string(domainTracking.FormatCarrossel), Label: "Carrossel", Description: "Anúncio em formato carrossel"},
			{Value: string(domainTracking.FormatBannerEstatico), Label: "Banner Estático", Description: "Banner/imagem estática"},
			{Value: string(domainTracking.FormatVideo), Label: "Video", Description: "Vídeo"},
			{Value: string(domainTracking.FormatVideoPreRoll), Label: "Video Pre-roll", Description: "Vídeo pre-roll"},
			{Value: string(domainTracking.FormatStoriesFullscreen), Label: "Stories Fullscreen", Description: "Stories tela cheia"},
			{Value: string(domainTracking.FormatStories), Label: "Stories", Description: "Stories"},
			{Value: string(domainTracking.FormatReels), Label: "Reels", Description: "Reels"},
			{Value: string(domainTracking.FormatImagemUnica), Label: "Imagem Única", Description: "Imagem única"},
			{Value: string(domainTracking.FormatTexto), Label: "Texto", Description: "Somente texto"},
			{Value: string(domainTracking.FormatCollection), Label: "Collection", Description: "Anúncio collection"},
		},
		PlatformCompatibility: map[string]PlatformCompatibilityInfo{
			"meta": {
				ValidMediums: []string{"paid-social", "organic-social"},
				Description:  "Plataforma Meta (Facebook/Instagram) suporta apenas mediums sociais",
			},
			"google": {
				ValidMediums: []string{"paid-search", "organic", "display", "video"},
				Description:  "Google suporta search, display e video mediums",
			},
			"tiktok": {
				ValidMediums: []string{"paid-social", "organic-social"},
				Description:  "TikTok suporta apenas mediums sociais",
			},
			"linkedin": {
				ValidMediums: []string{"paid-social", "organic-social"},
				Description:  "LinkedIn suporta apenas mediums sociais",
			},
			"mkt-direto": {
				ValidMediums: []string{"email", "sms", "whatsapp"},
				Description:  "Marketing direto suporta apenas canais de mensageria",
			},
			"offline": {
				ValidMediums: []string{"offline"},
				Description:  "Canais offline usam apenas medium offline",
			},
		},
	}
}

// EncodeTracking godoc
//
//	@Summary		Codifica tracking ID em mensagem WhatsApp
//	@Description	Insere código invisível ternário em mensagem para rastreamento
//	@Tags			CRM - Tracking
//	@Accept			json
//	@Produce		json
//	@Param			request	body		tracking.EncodeTrackingRequest	true	"Dados para encode"
//	@Success		200		{object}	tracking.EncodeTrackingResponse
//	@Failure		400		{object}	dto.ErrorResponse
//	@Failure		500		{object}	dto.ErrorResponse
//	@Router			/api/v1/trackings/encode [post]
//	@Security		BearerAuth
func (h *TrackingHandler) EncodeTracking(c *gin.Context) {
	var req tracking.EncodeTrackingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	result, err := h.encodeUseCase.Execute(req)
	if err != nil {
		h.logger.Error("Failed to encode tracking", zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// DecodeTracking godoc
//
//	@Summary		Decodifica mensagem WhatsApp para extrair tracking ID
//	@Description	Extrai código invisível ternário de mensagem para identificar tracking
//	@Tags			CRM - Tracking
//	@Accept			json
//	@Produce		json
//	@Param			request	body		tracking.DecodeTrackingRequest	true	"Mensagem para decode"
//	@Success		200		{object}	tracking.DecodeTrackingResponse
//	@Failure		400		{object}	dto.ErrorResponse
//	@Router			/api/v1/trackings/decode [post]
//	@Security		BearerAuth
func (h *TrackingHandler) DecodeTracking(c *gin.Context) {
	var req tracking.DecodeTrackingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	result, err := h.decodeUseCase.Execute(req)
	if err != nil {
		h.logger.Error("Failed to decode tracking", zap.Error(err))
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}
