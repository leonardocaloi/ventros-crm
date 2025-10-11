package ai

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/caloi/ventros-crm/internal/domain/crm/message_enrichment"
)

// VisionProvider implementa OCR + descrição de imagens usando GPT-4 Vision ou Gemini Vision
type VisionProvider struct {
	logger         *zap.Logger
	apiKey         string
	apiURL         string // OpenAI: https://api.openai.com/v1/chat/completions
	model          string // gpt-4-vision-preview, gpt-4o, gemini-pro-vision
	provider       string // openai ou google
	httpClient     *http.Client
	promptRegistry *VisionPromptRegistry
}

// VisionConfig configuração do Vision provider
type VisionConfig struct {
	// Gemini API (ai.google.dev) - Free tier com API Key
	APIKey string // API Key do Gemini Developer API

	// Vertex AI (cloud.google.com) - Enterprise com Service Account
	VertexProjectID      string // Google Cloud Project ID
	VertexLocation       string // Região (ex: us-central1)
	VertexServiceAccount string // Path para JSON do Service Account

	Provider   string // "gemini" (free tier) ou "vertex" (enterprise)
	Model      string
	TimeoutSec int
}

// NewVisionProvider cria um novo provider Vision
// Default: Gemini (melhor custo-benefício para imagem/vídeo)
func NewVisionProvider(logger *zap.Logger, config VisionConfig) *VisionProvider {
	// Default para Gemini (recomendado para uso geral)
	if config.Provider == "" {
		config.Provider = "gemini"
	}
	if config.Model == "" {
		if config.Provider == "gemini" {
			config.Model = "gemini-1.5-flash" // Rápido e barato
		} else {
			config.Model = "gpt-4o" // Fallback OpenAI
		}
	}
	if config.TimeoutSec == 0 {
		config.TimeoutSec = 60
	}

	apiURL := "https://generativelanguage.googleapis.com/v1beta/models/" + config.Model + ":generateContent"
	if config.Provider == "openai" {
		apiURL = "https://api.openai.com/v1/chat/completions"
	}

	return &VisionProvider{
		logger:   logger,
		apiKey:   config.APIKey,
		apiURL:   apiURL,
		model:    config.Model,
		provider: config.Provider,
		httpClient: &http.Client{
			Timeout: time.Duration(config.TimeoutSec) * time.Second,
		},
		promptRegistry: NewVisionPromptRegistry(),
	}
}

// Process processa imagem e retorna OCR + descrição
func (p *VisionProvider) Process(
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

	p.logger.Info("Processing image with Vision",
		zap.String("media_url", mediaURL),
		zap.String("provider", p.provider),
		zap.String("model", p.model),
		zap.String("context", contextStr))

	// 1. Download da imagem
	imageData, err := p.downloadImage(ctx, mediaURL)
	if err != nil {
		return nil, fmt.Errorf("failed to download image: %w", err)
	}

	// 2. Obter prompt baseado no contexto
	prompt := p.promptRegistry.GetPromptText(VisionPromptContext(contextStr))

	// 3. Processar com Vision API (default: Gemini)
	var extractedText string
	var metadata map[string]interface{}

	if p.provider == "gemini" {
		extractedText, metadata, err = p.processWithGemini(ctx, imageData, mediaURL, prompt)
	} else {
		extractedText, metadata, err = p.processWithOpenAI(ctx, imageData, mediaURL, prompt)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to process image: %w", err)
	}

	processingTime := time.Since(startTime)

	p.logger.Info("Image processed successfully",
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
func (p *VisionProvider) Name() string {
	return "vision"
}

// SupportsContentType verifica se suporta o tipo de conteúdo
func (p *VisionProvider) SupportsContentType(contentType message_enrichment.EnrichmentContentType) bool {
	return contentType == message_enrichment.EnrichmentTypeImage
}

// downloadImage baixa a imagem do URL
func (p *VisionProvider) downloadImage(ctx context.Context, mediaURL string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", mediaURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	return data, nil
}

// processWithOpenAI processa imagem usando OpenAI GPT-4 Vision
func (p *VisionProvider) processWithOpenAI(ctx context.Context, imageData []byte, mediaURL string, prompt string) (string, map[string]interface{}, error) {
	// Encode image to base64
	imageBase64 := base64.StdEncoding.EncodeToString(imageData)

	// Create request body
	requestBody := map[string]interface{}{
		"model": p.model,
		"messages": []map[string]interface{}{
			{
				"role": "user",
				"content": []map[string]interface{}{
					{
						"type": "text",
						"text": prompt,
					},
					{
						"type": "image_url",
						"image_url": map[string]string{
							"url": "data:image/jpeg;base64," + imageBase64,
						},
					},
				},
			},
		},
		"max_tokens": 1000,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return "", nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", p.apiURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+p.apiKey)
	req.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := p.httpClient.Do(req)
	if err != nil {
		return "", nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(respBody))
	}

	// Parse response
	var result OpenAIVisionResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if len(result.Choices) == 0 {
		return "", nil, fmt.Errorf("no choices in response")
	}

	extractedText := result.Choices[0].Message.Content

	// Extract metadata
	metadata := map[string]interface{}{
		"model":         p.model,
		"provider":      "openai",
		"image_size":    len(imageData),
		"prompt_tokens": result.Usage.PromptTokens,
		"total_tokens":  result.Usage.TotalTokens,
	}

	return extractedText, metadata, nil
}

// processWithGemini processa imagem usando Google Gemini Vision
func (p *VisionProvider) processWithGemini(ctx context.Context, imageData []byte, mediaURL string, prompt string) (string, map[string]interface{}, error) {
	// Encode image to base64
	imageBase64 := base64.StdEncoding.EncodeToString(imageData)

	// Criar request body com prompt customizado por contexto
	requestBody := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"parts": []map[string]interface{}{
					{
						"text": prompt,
					},
					{
						"inline_data": map[string]string{
							"mime_type": "image/jpeg",
							"data":      imageBase64,
						},
					},
				},
			},
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return "", nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	apiURL := p.apiURL + "?key=" + p.apiKey
	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := p.httpClient.Do(req)
	if err != nil {
		return "", nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(respBody))
	}

	// Parse response
	var result GeminiVisionResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if len(result.Candidates) == 0 || len(result.Candidates[0].Content.Parts) == 0 {
		return "", nil, fmt.Errorf("no content in response")
	}

	extractedText := result.Candidates[0].Content.Parts[0].Text

	// Extract metadata
	metadata := map[string]interface{}{
		"model":      p.model,
		"provider":   "google",
		"image_size": len(imageData),
	}

	return extractedText, metadata, nil
}

// OpenAIVisionResponse estrutura da resposta da API OpenAI Vision
type OpenAIVisionResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// GeminiVisionResponse estrutura da resposta da API Gemini
type GeminiVisionResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
			Role string `json:"role"`
		} `json:"content"`
		FinishReason  string `json:"finishReason"`
		Index         int    `json:"index"`
		SafetyRatings []struct {
			Category    string `json:"category"`
			Probability string `json:"probability"`
		} `json:"safetyRatings"`
	} `json:"candidates"`
}

// IsConfigured verifica se o provider está configurado
func (p *VisionProvider) IsConfigured() bool {
	return p.apiKey != ""
}

// ValidateConfig valida a configuração do provider
func (p *VisionProvider) ValidateConfig() error {
	if p.apiKey == "" {
		return fmt.Errorf("vision API key is required")
	}

	if p.provider != "openai" && p.provider != "google" {
		return fmt.Errorf("invalid provider: %s (must be 'openai' or 'google')", p.provider)
	}

	return nil
}
