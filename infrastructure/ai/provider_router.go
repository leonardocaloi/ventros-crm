package ai

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/caloi/ventros-crm/internal/domain/core/shared"
	"github.com/caloi/ventros-crm/internal/domain/crm/message_enrichment"
)

// ProviderRouter roteia requisições para o provider mais adequado
// Arquitetura híbrida baseada em pesquisa de qualidade:
// - LlamaParse: PDFs e documentos estruturados (rápido, ~6s)
// - Gemini Vision: Imagens, fotos, scans (alta acurácia 93-96%, robusto a ruído)
// - Groq Whisper: Áudio falado (PTT) - speech-to-text GRATUITO e 216x real-time
// - OpenAI Whisper: Fallback para Groq (pago mas confiável)
type ProviderRouter struct {
	logger                *zap.Logger
	llamaParseProvider    *LlamaParseProvider
	vertexProvider        *VertexVisionProvider
	groqWhisperProvider   *WhisperProvider // Groq (grátis, prioridade 1)
	openaiWhisperProvider *WhisperProvider // OpenAI (pago, fallback)
	mimeRegistry          shared.MimeTypeRegistry
}

// NewProviderRouter cria um novo router de providers
func NewProviderRouter(
	logger *zap.Logger,
	llamaParseProvider *LlamaParseProvider,
	vertexProvider *VertexVisionProvider,
	groqWhisperProvider *WhisperProvider, // Groq (grátis, pode ser nil)
	openaiWhisperProvider *WhisperProvider, // OpenAI (pago, pode ser nil)
	mimeRegistry shared.MimeTypeRegistry,
) *ProviderRouter {
	return &ProviderRouter{
		logger:                logger,
		llamaParseProvider:    llamaParseProvider,
		vertexProvider:        vertexProvider,
		groqWhisperProvider:   groqWhisperProvider,
		openaiWhisperProvider: openaiWhisperProvider,
		mimeRegistry:          mimeRegistry,
	}
}

// RouteRequest decide qual provider usar baseado no mimetype e contexto
// Retorna o provider escolhido e a razão da escolha
func (r *ProviderRouter) RouteRequest(
	mimeType string,
	contentType message_enrichment.EnrichmentContentType,
	isProfilePicture bool,
	isSpokenAudio bool, // Se é áudio falado (fala humana) - usa Whisper
) (Provider, string, error) {
	// REGRA 1: Foto de perfil SEMPRE usa Gemini Vision (melhor para análise visual + scoring)
	if isProfilePicture {
		r.logger.Info("Routing profile picture to Vertex Vision",
			zap.String("reason", "profile pictures require visual analysis and scoring"))
		return r.vertexProvider, "profile_picture_visual_analysis", nil
	}

	// Obter categoria do mimetype
	category, err := r.mimeRegistry.GetCategory(mimeType)
	if err != nil {
		return nil, "", fmt.Errorf("unsupported mime type: %s", mimeType)
	}

	// REGRA 2: Imagens standalone (JPG, PNG, etc) -> Gemini Vision
	// Razão: Gemini tem 93-96% acurácia em OCR de imagens, robusto a ruído/qualidade baixa
	// NOTA: PDFs com imagens usam LlamaParse (já faz OCR automaticamente)
	if category == shared.CategoryImage {
		r.logger.Info("Routing standalone image to Vertex Vision",
			zap.String("mime_type", mimeType),
			zap.String("reason", "Gemini Vision has 93-96% accuracy on image OCR, robust to noise"))
		return r.vertexProvider, "image_ocr_high_quality", nil
	}

	// REGRA 3: Áudio Falado (fala humana) -> Whisper (Groq → OpenAI → Gemini)
	// Razão: Whisper especializado em speech-to-text, melhor acurácia para voz
	// Estratégia: Groq (grátis) → OpenAI (pago) → Gemini (fallback final)
	if category == shared.CategoryAudio && isSpokenAudio {
		// PRIORIDADE 1: Groq Whisper (GRATUITO, 216x real-time)
		if r.groqWhisperProvider != nil && r.groqWhisperProvider.IsConfigured() {
			r.logger.Info("Routing spoken audio to Groq Whisper (FREE)",
				zap.String("mime_type", mimeType),
				zap.String("reason", "Groq Whisper is free and 216x real-time"))
			return r.groqWhisperProvider, "spoken_audio_groq_free", nil
		}

		// PRIORIDADE 2: OpenAI Whisper (PAGO, fallback se Groq falhar)
		if r.openaiWhisperProvider != nil && r.openaiWhisperProvider.IsConfigured() {
			r.logger.Info("Routing spoken audio to OpenAI Whisper (fallback)",
				zap.String("mime_type", mimeType),
				zap.String("reason", "Groq not available, using OpenAI Whisper (paid)"))
			return r.openaiWhisperProvider, "spoken_audio_openai_paid", nil
		}

		// PRIORIDADE 3: Gemini Vision (fallback final se nenhum Whisper configurado)
		r.logger.Warn("No Whisper provider configured, falling back to Gemini",
			zap.String("mime_type", mimeType))
		return r.vertexProvider, "spoken_audio_fallback_gemini", nil
	}

	// REGRA 3b: Outros Áudios/Vídeos -> Gemini Vision
	// Razão: Melhor para extração de frames + OCR de vídeos, áudio com contexto visual
	if category == shared.CategoryAudio {
		r.logger.Info("Routing audio/video to Vertex Vision",
			zap.String("mime_type", mimeType),
			zap.String("reason", "Gemini better for frame extraction and video OCR"))
		return r.vertexProvider, "video_frame_extraction", nil
	}

	// REGRA 4: PDFs e documentos estruturados -> LlamaParse
	// Razão: Rápido (~6s), otimizado para documentos, extrai estrutura + markdown
	// IMPORTANTE: PDFs com imagens também usam LlamaParse (OCR automático de imagens dentro do PDF)
	if category == shared.CategoryPDF ||
		category == shared.CategoryOffice ||
		category == shared.CategorySpreadsheet ||
		category == shared.CategoryPresentation ||
		category == shared.CategoryText {
		r.logger.Info("Routing document to LlamaParse",
			zap.String("mime_type", mimeType),
			zap.String("category", string(category)),
			zap.String("reason", "LlamaParse optimized for structured documents, fast ~6s, extracts markdown + OCR of embedded images"))
		return r.llamaParseProvider, "structured_document_parsing", nil
	}

	// Fallback: usar Gemini Vision (mais robusto para casos desconhecidos)
	r.logger.Warn("Unknown category, falling back to Vertex Vision",
		zap.String("mime_type", mimeType),
		zap.String("category", string(category)))
	return r.vertexProvider, "fallback_unknown_type", nil
}

// Process processa o conteúdo usando o provider mais adequado
// Abstração sobre roteamento + execução
// NOTA: mimeType pode ser vazio se contentType for EnrichmentTypeVoice (inferido do ContentType)
func (r *ProviderRouter) Process(
	ctx context.Context,
	mediaURL string,
	mimeType string,
	contentType message_enrichment.EnrichmentContentType,
	processingContext *string,
	isProfilePicture bool,
	isSpokenAudio bool, // Se é áudio de fala humana (usa Whisper)
) (*EnrichmentResult, error) {
	// Auto-detectar isSpokenAudio se contentType == EnrichmentTypeVoice
	if contentType == message_enrichment.EnrichmentTypeVoice {
		isSpokenAudio = true
	}

	// Auto-detectar isProfilePicture se context == "profile_picture"
	if processingContext != nil && *processingContext == "profile_picture" {
		isProfilePicture = true
	}

	// Se mimeType vazio, inferir do contentType
	if mimeType == "" {
		mimeType = r.inferMimeTypeFromContentType(contentType)
	}

	// Escolher provider
	provider, reason, err := r.RouteRequest(mimeType, contentType, isProfilePicture, isSpokenAudio)
	if err != nil {
		return nil, fmt.Errorf("failed to route request: %w", err)
	}

	r.logger.Info("Processing with selected provider",
		zap.String("provider", provider.Name()),
		zap.String("routing_reason", reason),
		zap.String("mime_type", mimeType),
		zap.String("content_type", string(contentType)),
		zap.Bool("is_profile_picture", isProfilePicture),
		zap.Bool("is_spoken_audio", isSpokenAudio))

	// Executar processamento
	result, err := provider.Process(ctx, mediaURL, contentType, processingContext)
	if err != nil {
		return nil, fmt.Errorf("provider %s failed: %w", provider.Name(), err)
	}

	// Adicionar metadata de roteamento
	if result.Metadata == nil {
		result.Metadata = make(map[string]interface{})
	}
	result.Metadata["routing_reason"] = reason
	result.Metadata["selected_provider"] = provider.Name()

	return result, nil
}

// inferMimeTypeFromContentType infere mimetype quando não disponível
func (r *ProviderRouter) inferMimeTypeFromContentType(contentType message_enrichment.EnrichmentContentType) string {
	switch contentType {
	case message_enrichment.EnrichmentTypeImage:
		return "image/jpeg" // Default
	case message_enrichment.EnrichmentTypeVoice, message_enrichment.EnrichmentTypeAudio:
		return "audio/mpeg" // Default (mp3)
	case message_enrichment.EnrichmentTypeVideo:
		return "video/mp4" // Default
	case message_enrichment.EnrichmentTypeDocument:
		return "application/pdf" // Default
	default:
		return "application/octet-stream" // Fallback genérico
	}
}

// GetProviderForMimeType retorna o provider recomendado para um mimetype específico
// Útil para pré-validação antes de processar
func (r *ProviderRouter) GetProviderForMimeType(mimeType string, isProfilePicture bool, isSpokenAudio bool) (string, error) {
	provider, reason, err := r.RouteRequest(
		mimeType,
		message_enrichment.EnrichmentTypeDocument, // Default
		isProfilePicture,
		isSpokenAudio,
	)
	if err != nil {
		return "", err
	}

	r.logger.Debug("Provider selection query",
		zap.String("mime_type", mimeType),
		zap.Bool("is_profile_picture", isProfilePicture),
		zap.Bool("is_spoken_audio", isSpokenAudio),
		zap.String("selected_provider", provider.Name()),
		zap.String("reason", reason))

	return provider.Name(), nil
}

// Provider interface que ambos LlamaParse e VertexVision devem implementar
type Provider interface {
	Name() string
	Process(ctx context.Context, mediaURL string, contentType message_enrichment.EnrichmentContentType, processingContext *string) (*EnrichmentResult, error)
	SupportsContentType(contentType message_enrichment.EnrichmentContentType) bool
	SupportsMimeType(mimeType string) bool
}
