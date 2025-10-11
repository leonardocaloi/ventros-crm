package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"cloud.google.com/go/auth"
	"go.uber.org/zap"
	"google.golang.org/genai"

	"github.com/caloi/ventros-crm/internal/domain/crm/message_enrichment"
)

// VertexVisionProvider implementa Vision usando Vertex AI com Service Account
// Suporta: Gemini, Claude (Anthropic), e outros modelos via Vertex AI
type VertexVisionProvider struct {
	logger         *zap.Logger
	client         *genai.Client
	model          string // Default: gemini-1.5-flash
	projectID      string
	location       string
	promptRegistry *VisionPromptRegistry
}

// VertexVisionConfig configuração do Vertex AI provider
type VertexVisionConfig struct {
	ProjectID          string // Google Cloud Project ID
	Location           string // Região (ex: us-central1)
	ServiceAccountPath string // Path para arquivo JSON do Service Account
	Model              string // Default: gemini-1.5-flash
}

// NewVertexVisionProvider cria um novo provider Vertex AI com Service Account
func NewVertexVisionProvider(logger *zap.Logger, config VertexVisionConfig) (*VertexVisionProvider, error) {
	// Defaults
	if config.Location == "" {
		config.Location = "us-central1"
	}
	if config.Model == "" {
		config.Model = "gemini-1.5-flash" // Rápido e barato
	}

	// Validar configuração
	if config.ProjectID == "" {
		return nil, fmt.Errorf("vertex project ID is required")
	}
	if config.ServiceAccountPath == "" {
		return nil, fmt.Errorf("service account path is required")
	}

	// Ler Service Account JSON
	key, err := os.ReadFile(config.ServiceAccountPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read service account key: %w", err)
	}

	// Parse Service Account
	var serviceAccount struct {
		ClientEmail string `json:"client_email"`
		PrivateKey  string `json:"private_key"`
		TokenURI    string `json:"token_uri"`
		ProjectID   string `json:"project_id"`
	}
	if err := json.Unmarshal(key, &serviceAccount); err != nil {
		return nil, fmt.Errorf("invalid service account JSON: %w", err)
	}

	// Criar 2-legged OAuth token provider (tokens duram 1h)
	tp, err := auth.New2LOTokenProvider(&auth.Options2LO{
		Email:      serviceAccount.ClientEmail,
		PrivateKey: []byte(serviceAccount.PrivateKey),
		TokenURL:   serviceAccount.TokenURI,
		Scopes:     []string{"https://www.googleapis.com/auth/cloud-platform"},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create 2LO token provider: %w", err)
	}

	// Criar credentials usando token provider
	credentials := auth.NewCredentials(&auth.CredentialsOptions{
		TokenProvider: tp,
		JSON:          key,
	})

	// Criar cliente Vertex AI (GenAI SDK)
	ctx := context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		Project:     config.ProjectID,
		Location:    config.Location,
		Backend:     genai.BackendVertexAI,
		Credentials: credentials,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create vertex AI client: %w", err)
	}

	logger.Info("Vertex AI Vision provider initialized",
		zap.String("project_id", config.ProjectID),
		zap.String("location", config.Location),
		zap.String("model", config.Model),
		zap.String("service_account_email", serviceAccount.ClientEmail))

	return &VertexVisionProvider{
		logger:         logger,
		client:         client,
		model:          config.Model,
		projectID:      config.ProjectID,
		location:       config.Location,
		promptRegistry: NewVisionPromptRegistry(),
	}, nil
}

// Process processa imagem e retorna OCR + descrição usando Vertex AI
func (p *VertexVisionProvider) Process(
	ctx context.Context,
	mediaURL string,
	contentType message_enrichment.EnrichmentContentType,
	visionContext *string,
) (*EnrichmentResult, error) {
	startTime := time.Now()

	// Determinar contexto (default: chat_message)
	contextStr := string(ContextChatMessage)
	if visionContext != nil && *visionContext != "" {
		contextStr = *visionContext
	}

	p.logger.Info("Processing image with Vertex AI",
		zap.String("media_url", mediaURL),
		zap.String("model", p.model),
		zap.String("context", contextStr))

	// Obter prompt baseado no contexto
	prompt := p.promptRegistry.GetPromptText(VisionPromptContext(contextStr))

	// Criar request para Gemini Vision via Vertex AI
	// Usando Content com texto + imagem
	textPart := genai.NewPartFromText(prompt)
	imagePart := genai.NewPartFromURI(mediaURL, "image/jpeg")

	contents := []*genai.Content{
		genai.NewContentFromParts([]*genai.Part{textPart, imagePart}, genai.RoleUser),
	}

	// Configurar parâmetros de geração
	config := &genai.GenerateContentConfig{
		Temperature:     genai.Ptr(float32(0.1)), // Baixa criatividade para ser mais factual
		TopK:            genai.Ptr(float32(40)),
		TopP:            genai.Ptr(float32(0.95)),
		MaxOutputTokens: 1024,
	}

	// Chamar API
	response, err := p.client.Models.GenerateContent(ctx, p.model, contents, config)
	if err != nil {
		return nil, fmt.Errorf("failed to generate content: %w", err)
	}

	// Extrair texto da resposta
	extractedText := ""
	if len(response.Candidates) > 0 && len(response.Candidates[0].Content.Parts) > 0 {
		// Concatenar todos os parts de texto
		for _, part := range response.Candidates[0].Content.Parts {
			if part.Text != "" {
				extractedText += part.Text
			}
		}
	}

	if extractedText == "" {
		return nil, fmt.Errorf("no content in response")
	}

	processingTime := time.Since(startTime)

	// Extract metadata
	metadata := map[string]interface{}{
		"model":    p.model,
		"provider": "vertex-ai",
		"project":  p.projectID,
		"location": p.location,
		"context":  contextStr,
	}

	// Adicionar usage metadata se disponível
	if response.UsageMetadata != nil {
		metadata["prompt_tokens"] = response.UsageMetadata.PromptTokenCount
		metadata["candidates_tokens"] = response.UsageMetadata.CandidatesTokenCount
		metadata["total_tokens"] = response.UsageMetadata.TotalTokenCount
	}

	p.logger.Info("Image processed successfully with Vertex AI",
		zap.String("media_url", mediaURL),
		zap.String("context", contextStr),
		zap.Int("text_length", len(extractedText)),
		zap.Duration("processing_time", processingTime))

	return &EnrichmentResult{
		ExtractedText:  extractedText,
		Metadata:       metadata,
		ProcessingTime: processingTime,
	}, nil
}

// Name retorna o nome do provider
func (p *VertexVisionProvider) Name() string {
	return "vertex-vision"
}

// SupportsContentType verifica se suporta o tipo de conteúdo
func (p *VertexVisionProvider) SupportsContentType(contentType message_enrichment.EnrichmentContentType) bool {
	return contentType == message_enrichment.EnrichmentTypeImage
}

// SupportsMimeType verifica se suporta um mimetype específico
// Vertex Vision suporta imagens e vídeos para análise visual
func (p *VertexVisionProvider) SupportsMimeType(mimeType string) bool {
	supportedMimes := map[string]bool{
		// Imagens
		"image/jpeg":    true,
		"image/png":     true,
		"image/gif":     true,
		"image/bmp":     true,
		"image/webp":    true,
		"image/tiff":    true,
		"image/svg+xml": true,
		// Vídeos (para frame extraction)
		"video/mp4":  true,
		"video/mpeg": true,
		// Áudio (para transcrição)
		"audio/mpeg": true,
		"audio/mp4":  true,
		"audio/wav":  true,
		"audio/webm": true,
	}
	return supportedMimes[mimeType]
}

// Close fecha o cliente (deve ser chamado no shutdown)
// Note: genai.Client does not have a Close() method in the current SDK version
func (p *VertexVisionProvider) Close() error {
	// O SDK atual não requer close explícito
	return nil
}

// IsConfigured verifica se o provider está configurado
func (p *VertexVisionProvider) IsConfigured() bool {
	return p.client != nil
}

// ValidateConfig valida a configuração do provider
func (p *VertexVisionProvider) ValidateConfig() error {
	if p.projectID == "" {
		return fmt.Errorf("vertex project ID is required")
	}
	if p.location == "" {
		return fmt.Errorf("vertex location is required")
	}
	if p.client == nil {
		return fmt.Errorf("vertex AI client not initialized")
	}
	return nil
}
