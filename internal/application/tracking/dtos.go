package tracking

import (
	"time"

	"github.com/caloi/ventros-crm/internal/domain/tracking"
	"github.com/google/uuid"
)

// CreateTrackingDTO para criar novo tracking
type CreateTrackingDTO struct {
	ContactID      uuid.UUID
	SessionID      *uuid.UUID
	TenantID       string
	ProjectID      uuid.UUID
	Source         string // meta_ads, google_ads, etc
	Platform       string // instagram, facebook, etc
	Campaign       string
	AdID           string
	AdURL          string
	ClickID        string
	ConversionData string
	UTMSource      string
	UTMMedium      string
	UTMCampaign    string
	UTMTerm        string
	UTMContent     string
	Metadata       map[string]interface{}
}

// TrackingDTO representa um tracking para a API
type TrackingDTO struct {
	ID             uuid.UUID              `json:"id"`
	ContactID      uuid.UUID              `json:"contact_id"`
	SessionID      *uuid.UUID             `json:"session_id,omitempty"`
	TenantID       string                 `json:"tenant_id"`
	ProjectID      uuid.UUID              `json:"project_id"`
	Source         string                 `json:"source"`
	Platform       string                 `json:"platform"`
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
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
}

// ToDTO converte Tracking domain para TrackingDTO
func ToDTO(t *tracking.Tracking) TrackingDTO {
	return TrackingDTO{
		ID:             t.ID(),
		ContactID:      t.ContactID(),
		SessionID:      t.SessionID(),
		TenantID:       t.TenantID(),
		ProjectID:      t.ProjectID(),
		Source:         string(t.Source()),
		Platform:       string(t.Platform()),
		Campaign:       t.Campaign(),
		AdID:           t.AdID(),
		AdURL:          t.AdURL(),
		ClickID:        t.ClickID(),
		ConversionData: t.ConversionData(),
		UTMSource:      t.UTMSource(),
		UTMMedium:      t.UTMMedium(),
		UTMCampaign:    t.UTMCampaign(),
		UTMTerm:        t.UTMTerm(),
		UTMContent:     t.UTMContent(),
		Metadata:       t.Metadata(),
		CreatedAt:      t.CreatedAt(),
		UpdatedAt:      t.UpdatedAt(),
	}
}

// ToDTOList converte uma lista de Trackings para DTOs
func ToDTOList(trackings []*tracking.Tracking) []TrackingDTO {
	dtos := make([]TrackingDTO, 0, len(trackings))
	for _, t := range trackings {
		dtos = append(dtos, ToDTO(t))
	}
	return dtos
}

// EncodeTrackingRequest requisição para codificar tracking em mensagem
type EncodeTrackingRequest struct {
	TrackingID int64  `json:"tracking_id" binding:"required"`
	Message    string `json:"message" binding:"required"`
	Phone      string `json:"phone,omitempty"`
}

// EncodeTrackingResponse resposta da codificação
type EncodeTrackingResponse struct {
	Success         bool                   `json:"success"`
	TrackingID      int64                  `json:"tracking_id"`
	OriginalMessage string                 `json:"original_message"`
	TernaryEncoded  string                 `json:"ternary_encoded"`
	DecimalValue    int64                  `json:"decimal_value"`
	Phone           string                 `json:"phone,omitempty"`
	InvisibleCode   string                 `json:"invisible_code"`
	MessageWithCode string                 `json:"message_with_code"`
	WhatsAppLink    string                 `json:"whatsapp_link,omitempty"`
	Debug           map[string]interface{} `json:"debug,omitempty"`
}

// DecodeTrackingRequest requisição para decodificar mensagem
type DecodeTrackingRequest struct {
	Message string `json:"message" binding:"required"`
}

// DecodeTrackingResponse resposta da decodificação
type DecodeTrackingResponse struct {
	Success         bool                   `json:"success"`
	DecodedTernary  string                 `json:"decoded_ternary,omitempty"`
	DecodedDecimal  int64                  `json:"decoded_decimal,omitempty"`
	Confidence      string                 `json:"confidence"`
	Analysis        map[string]interface{} `json:"analysis,omitempty"`
	CleanMessage    string                 `json:"clean_message"`
	OriginalMessage string                 `json:"original_message"`
	Error           string                 `json:"error,omitempty"`
}
