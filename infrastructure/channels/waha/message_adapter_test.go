package waha_test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/caloi/ventros-crm/infrastructure/channels/waha"
	"github.com/caloi/ventros-crm/internal/domain/crm/message"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func loadTestData(t *testing.T, filename string) waha.WAHAMessageEvent {
	data, err := os.ReadFile("testdata/" + filename)
	require.NoError(t, err, "failed to read test data file")
	
	var event waha.WAHAMessageEvent
	err = json.Unmarshal(data, &event)
	require.NoError(t, err, "failed to unmarshal test data")
	
	return event
}

func TestMessageAdapter_ToContentType_RealImageMessage(t *testing.T) {
	// Arrange
	adapter := waha.NewMessageAdapter()
	event := loadTestData(t, "real_image_message.json")
	
	// Act
	contentType, err := adapter.ToContentType(event)
	
	// Assert
	require.NoError(t, err)
	assert.Equal(t, message.ContentTypeImage, contentType)
}

func TestMessageAdapter_ToContentType_RealTextMessage(t *testing.T) {
	// Arrange
	adapter := waha.NewMessageAdapter()
	event := loadTestData(t, "real_text_message.json")
	
	// Act
	contentType, err := adapter.ToContentType(event)
	
	// Assert
	require.NoError(t, err)
	assert.Equal(t, message.ContentTypeText, contentType)
}

func TestMessageAdapter_ExtractText_FromImageMessage(t *testing.T) {
	// Arrange
	adapter := waha.NewMessageAdapter()
	event := loadTestData(t, "real_image_message.json")
	
	// Act
	text := adapter.ExtractText(event)
	
	// Assert
	// Imagem sem caption deve retornar vazio
	assert.Equal(t, "", text)
}

func TestMessageAdapter_ExtractText_FromTextMessage(t *testing.T) {
	// Arrange
	adapter := waha.NewMessageAdapter()
	event := loadTestData(t, "real_text_message.json")
	
	// Act
	text := adapter.ExtractText(event)
	
	// Assert
	assert.Equal(t, "No momento não vou poder", text)
}

func TestMessageAdapter_ExtractMediaURL_FromImageMessage(t *testing.T) {
	// Arrange
	adapter := waha.NewMessageAdapter()
	event := loadTestData(t, "real_image_message.json")
	
	// Act
	url := adapter.ExtractMediaURL(event)
	
	// Assert
	require.NotNil(t, url)
	assert.Contains(t, *url, "storage.googleapis.com/waha-ventros")
}

func TestMessageAdapter_ExtractMediaURL_FromTextMessage(t *testing.T) {
	// Arrange
	adapter := waha.NewMessageAdapter()
	event := loadTestData(t, "real_text_message.json")
	
	// Act
	url := adapter.ExtractMediaURL(event)
	
	// Assert
	assert.Nil(t, url)
}

func TestMessageAdapter_ExtractMimeType_FromImageMessage(t *testing.T) {
	// Arrange
	adapter := waha.NewMessageAdapter()
	event := loadTestData(t, "real_image_message.json")
	
	// Act
	mimetype := adapter.ExtractMimeType(event)
	
	// Assert
	require.NotNil(t, mimetype)
	assert.Equal(t, "image/jpeg", *mimetype)
}

func TestMessageAdapter_ExtractContactPhone_Standard(t *testing.T) {
	// Arrange
	adapter := waha.NewMessageAdapter()
	event := loadTestData(t, "real_text_message.json")
	
	// Act
	phone := adapter.ExtractContactPhone(event)
	
	// Assert
	assert.Equal(t, "554498699850", phone)
}

func TestMessageAdapter_ExtractContactPhone_CUs(t *testing.T) {
	// Arrange
	adapter := waha.NewMessageAdapter()
	event := loadTestData(t, "real_image_message.json")
	
	// Act
	phone := adapter.ExtractContactPhone(event)
	
	// Assert
	assert.Equal(t, "554499223925", phone)
}

func TestMessageAdapter_ExtractTrackingData_NoTracking(t *testing.T) {
	// Arrange
	adapter := waha.NewMessageAdapter()
	event := loadTestData(t, "real_text_message.json")
	
	// Act
	tracking := adapter.ExtractTrackingData(event)
	
	// Assert
	assert.Empty(t, tracking)
}

func TestMessageAdapter_ExtractTrackingData_FBAds(t *testing.T) {
	// Arrange
	adapter := waha.NewMessageAdapter()
	event := loadTestData(t, "fb_ads_message.json")
	
	// Act
	tracking := adapter.ExtractTrackingData(event)
	
	// Assert
	require.NotEmpty(t, tracking)
	assert.Equal(t, "ctwa_ad", tracking["conversion_source"])
	assert.Equal(t, "instagram", tracking["conversion_app"])
	assert.Equal(t, "ad", tracking["ad_source_type"])
	assert.Equal(t, "120232639087330785", tracking["ad_source_id"])
	assert.Equal(t, "instagram", tracking["ad_source_app"])
	assert.Equal(t, "https://www.instagram.com/p/DOqYjmWDOCx/", tracking["ad_source_url"])
	assert.Contains(t, tracking["ctwa_clid"], "Afcg5DA4aj8pfp0faj5HIKtKi2vOUGt4")
}

func TestMessageAdapter_IsFromAd_NoAd(t *testing.T) {
	// Arrange
	adapter := waha.NewMessageAdapter()
	event := loadTestData(t, "real_text_message.json")
	
	// Act
	isFromAd := adapter.IsFromAd(event)
	
	// Assert
	assert.False(t, isFromAd)
}

func TestMessageAdapter_IsFromAd_FBAds(t *testing.T) {
	// Arrange
	adapter := waha.NewMessageAdapter()
	event := loadTestData(t, "fb_ads_message.json")
	
	// Act
	isFromAd := adapter.IsFromAd(event)
	
	// Assert
	assert.True(t, isFromAd)
}

func TestMessageAdapter_FullFlow_TextMessage(t *testing.T) {
	// Arrange
	adapter := waha.NewMessageAdapter()
	event := loadTestData(t, "real_text_message.json")
	
	// Act - simula o fluxo completo
	contentType, err := adapter.ToContentType(event)
	require.NoError(t, err)
	
	phone := adapter.ExtractContactPhone(event)
	text := adapter.ExtractText(event)
	mediaURL := adapter.ExtractMediaURL(event)
	
	// Assert
	assert.Equal(t, message.ContentTypeText, contentType)
	assert.Equal(t, "554498699850", phone)
	assert.Equal(t, "No momento não vou poder", text)
	assert.Nil(t, mediaURL)
	
	// Verificações adicionais
	assert.Equal(t, "Ana Alves", event.Payload.Data.Info.PushName)
	assert.False(t, event.Payload.FromMe)
	assert.False(t, event.Payload.Data.Info.IsGroup)
}

func TestMessageAdapter_FullFlow_ImageMessage(t *testing.T) {
	// Arrange
	adapter := waha.NewMessageAdapter()
	event := loadTestData(t, "real_image_message.json")
	
	// Act - simula o fluxo completo
	contentType, err := adapter.ToContentType(event)
	require.NoError(t, err)
	
	phone := adapter.ExtractContactPhone(event)
	text := adapter.ExtractText(event)
	mediaURL := adapter.ExtractMediaURL(event)
	mimetype := adapter.ExtractMimeType(event)
	
	// Assert
	assert.Equal(t, message.ContentTypeImage, contentType)
	assert.Equal(t, "554499223925", phone)
	assert.Equal(t, "", text) // Sem caption
	assert.NotNil(t, mediaURL)
	assert.Contains(t, *mediaURL, "storage.googleapis.com")
	assert.NotNil(t, mimetype)
	assert.Equal(t, "image/jpeg", *mimetype)
	
	// Verificações adicionais
	assert.Equal(t, "Aline Fatobene", event.Payload.Data.Info.PushName)
	assert.False(t, event.Payload.FromMe)
	assert.True(t, event.Payload.HasMedia)
}

func TestMessageAdapter_SessionMetadata(t *testing.T) {
	// Arrange
	event := loadTestData(t, "real_text_message.json")
	
	// Assert - verifica metadata da sessão
	assert.Equal(t, "ask-dermato-imersao", event.Session)
	assert.Equal(t, "imersao", event.Metadata["number"])
	assert.Equal(t, "ASK Dermatologia", event.Metadata["customer"])
	assert.Equal(t, "2025.9.5", event.Environment.Version)
	assert.Equal(t, "GOWS", event.Environment.Engine)
}

// Benchmark para medir performance do adapter
func BenchmarkMessageAdapter_ToContentType(b *testing.B) {
	data, _ := os.ReadFile("testdata/real_text_message.json")
	var event waha.WAHAMessageEvent
	json.Unmarshal(data, &event)
	
	adapter := waha.NewMessageAdapter()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		adapter.ToContentType(event)
	}
}

func TestMessageAdapter_ToContentType_FBAdsMessage(t *testing.T) {
	// Arrange
	adapter := waha.NewMessageAdapter()
	event := loadTestData(t, "fb_ads_message.json")
	
	// Act
	contentType, err := adapter.ToContentType(event)
	
	// Assert
	require.NoError(t, err)
	assert.Equal(t, message.ContentTypeText, contentType)
}

func TestMessageAdapter_ExtractText_FBAdsMessage(t *testing.T) {
	// Arrange
	adapter := waha.NewMessageAdapter()
	event := loadTestData(t, "fb_ads_message.json")
	
	// Act
	text := adapter.ExtractText(event)
	
	// Assert
	assert.Equal(t, "Olá! Tenho interesse na imersão e queria mais informações, por favor.", text)
}

func TestMessageAdapter_ExtractContactPhone_FBAdsMessage(t *testing.T) {
	// Arrange
	adapter := waha.NewMessageAdapter()
	event := loadTestData(t, "fb_ads_message.json")
	
	// Act
	phone := adapter.ExtractContactPhone(event)
	
	// Assert
	assert.Equal(t, "554498211518", phone)
}

func TestMessageAdapter_FullFlow_FBAdsMessage(t *testing.T) {
	// Arrange
	adapter := waha.NewMessageAdapter()
	event := loadTestData(t, "fb_ads_message.json")
	
	// Act - simula o fluxo completo
	contentType, err := adapter.ToContentType(event)
	require.NoError(t, err)
	
	phone := adapter.ExtractContactPhone(event)
	text := adapter.ExtractText(event)
	mediaURL := adapter.ExtractMediaURL(event)
	tracking := adapter.ExtractTrackingData(event)
	isFromAd := adapter.IsFromAd(event)
	
	// Assert - Dados básicos da mensagem
	assert.Equal(t, message.ContentTypeText, contentType)
	assert.Equal(t, "554498211518", phone)
	assert.Equal(t, "Olá! Tenho interesse na imersão e queria mais informações, por favor.", text)
	assert.Nil(t, mediaURL)
	
	// Assert - Detecção de anúncio
	assert.True(t, isFromAd, "Message should be identified as from ad")
	
	// Assert - Dados de tracking
	require.NotEmpty(t, tracking, "Tracking data should be present")
	assert.Equal(t, "ctwa_ad", tracking["conversion_source"])
	assert.Equal(t, "instagram", tracking["conversion_app"])
	assert.Equal(t, "ad", tracking["ad_source_type"])
	assert.Equal(t, "120232639087330785", tracking["ad_source_id"])
	assert.Equal(t, "instagram", tracking["ad_source_app"])
	assert.Equal(t, "https://www.instagram.com/p/DOqYjmWDOCx/", tracking["ad_source_url"])
	assert.NotEmpty(t, tracking["ctwa_clid"], "Click ID should be present")
	
	// Assert - Metadata da sessão
	assert.Equal(t, "ask-dermato-imersao", event.Session)
	assert.Equal(t, "imersao", event.Metadata["number"])
	assert.Equal(t, "ASK Dermatologia", event.Metadata["customer"])
	
	// Assert - Informações do contato
	assert.Equal(t, "Nardin", event.Payload.Data.Info.PushName)
	assert.False(t, event.Payload.FromMe)
	assert.False(t, event.Payload.Data.Info.IsGroup)
}

func TestMessageAdapter_ExtractAdContextInfo(t *testing.T) {
	// Arrange
	adapter := waha.NewMessageAdapter()
	event := loadTestData(t, "fb_ads_message.json")
	
	// Act
	tracking := adapter.ExtractTrackingData(event)
	
	// Assert - Verifica campos específicos do contexto de anúncio
	assert.NotEmpty(t, tracking)
	
	// Conversão
	assert.Equal(t, "ctwa_ad", tracking["conversion_source"], "Entry point conversion source")
	assert.Equal(t, "instagram", tracking["conversion_app"], "Conversion app")
	
	// External Ad Reply
	assert.Equal(t, "ad", tracking["ad_source_type"], "Ad source type")
	assert.Equal(t, "120232639087330785", tracking["ad_source_id"], "Ad ID on Facebook/Instagram")
	assert.Equal(t, "instagram", tracking["ad_source_app"], "Source platform")
	assert.Equal(t, "https://www.instagram.com/p/DOqYjmWDOCx/", tracking["ad_source_url"], "Instagram post URL")
	
	// CTWA Click ID - importante para tracking de conversão
	ctwaClid := tracking["ctwa_clid"]
	assert.NotEmpty(t, ctwaClid, "CTWA Click ID should be present for conversion tracking")
	assert.Greater(t, len(ctwaClid), 50, "Click ID should be a long encoded string")
}

func BenchmarkMessageAdapter_ExtractAll(b *testing.B) {
	data, _ := os.ReadFile("testdata/real_image_message.json")
	var event waha.WAHAMessageEvent
	json.Unmarshal(data, &event)
	
	adapter := waha.NewMessageAdapter()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		adapter.ToContentType(event)
		adapter.ExtractContactPhone(event)
		adapter.ExtractText(event)
		adapter.ExtractMediaURL(event)
		adapter.ExtractMimeType(event)
	}
}

func BenchmarkMessageAdapter_ExtractTracking(b *testing.B) {
	data, _ := os.ReadFile("testdata/fb_ads_message.json")
	var event waha.WAHAMessageEvent
	json.Unmarshal(data, &event)
	
	adapter := waha.NewMessageAdapter()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		adapter.ExtractTrackingData(event)
		adapter.IsFromAd(event)
	}
}
