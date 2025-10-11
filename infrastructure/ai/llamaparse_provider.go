package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"time"

	"go.uber.org/zap"

	"github.com/caloi/ventros-crm/internal/domain/message_enrichment"
	"github.com/caloi/ventros-crm/internal/domain/shared"
)

// LlamaParseProvider implementa OCR de documentos usando LlamaParse API
// Suporta: PDF, DOCX, XLSX, PPTX, Images, HTML, Audio (30+ formatos)
// Retorna: Markdown formatado
// Webhook: Resultado enviado via POST para webhook_url configurado
// Segue princípios SOLID com dependency injection para mimetypes
type LlamaParseProvider struct {
	logger         *zap.Logger
	apiKey         string
	apiURL         string // https://api.cloud.llamaindex.ai/api/v1/parsing/upload
	webhookURL     string // URL para receber resultado assíncrono (HTTPS obrigatório)
	httpClient     *http.Client
	mimeRegistry   shared.MimeTypeRegistry // Dependency injection (SOLID)
}

// LlamaParseConfig configuração do LlamaParse provider
type LlamaParseConfig struct {
	APIKey     string // LlamaCloud API Key
	WebhookURL string // URL webhook para receber resultado (HTTPS, <200 chars)
	TimeoutSec int    // Timeout para upload (default: 30s)
}

// NewLlamaParseProvider cria um novo provider LlamaParse com dependency injection
// Segue princípios SOLID: recebe MimeTypeRegistry injetado
func NewLlamaParseProvider(
	logger *zap.Logger,
	config LlamaParseConfig,
	mimeRegistry shared.MimeTypeRegistry,
) (*LlamaParseProvider, error) {
	if config.APIKey == "" {
		return nil, fmt.Errorf("llamaparse API key is required")
	}

	if config.WebhookURL == "" {
		return nil, fmt.Errorf("webhook URL is required for async parsing")
	}

	// Validar webhook URL (HTTPS obrigatório, <200 chars)
	if len(config.WebhookURL) > 200 {
		return nil, fmt.Errorf("webhook URL must be less than 200 characters")
	}

	if mimeRegistry == nil {
		return nil, fmt.Errorf("mime registry is required (dependency injection)")
	}

	if config.TimeoutSec == 0 {
		config.TimeoutSec = 30 // Default: 30s para upload (parsing é assíncrono)
	}

	logger.Info("LlamaParse provider initialized",
		zap.String("webhook_url", config.WebhookURL),
		zap.Int("timeout_sec", config.TimeoutSec),
		zap.Int("supported_mimetypes", len(mimeRegistry.GetSupportedMimeTypes())))

	return &LlamaParseProvider{
		logger:       logger,
		apiKey:       config.APIKey,
		apiURL:       "https://api.cloud.llamaindex.ai/api/v1/parsing/upload",
		webhookURL:   config.WebhookURL,
		mimeRegistry: mimeRegistry,
		httpClient: &http.Client{
			Timeout: time.Duration(config.TimeoutSec) * time.Second,
		},
	}, nil
}

// Process envia documento para LlamaParse (async via webhook)
// Retorna jobID imediatamente, resultado virá via webhook
func (p *LlamaParseProvider) Process(
	ctx context.Context,
	mediaURL string,
	contentType message_enrichment.EnrichmentContentType,
	processingContext *string,
) (*EnrichmentResult, error) {
	startTime := time.Now()

	p.logger.Info("Uploading document to LlamaParse",
		zap.String("media_url", mediaURL),
		zap.String("webhook_url", p.webhookURL))

	// 1. Download do documento
	documentData, filename, err := p.downloadDocument(ctx, mediaURL)
	if err != nil {
		return nil, fmt.Errorf("failed to download document: %w", err)
	}

	// 2. Upload para LlamaParse (assíncrono via webhook)
	jobID, err := p.uploadDocument(ctx, documentData, filename)
	if err != nil {
		return nil, fmt.Errorf("failed to upload to llamaparse: %w", err)
	}

	processingTime := time.Since(startTime)

	p.logger.Info("Document queued for parsing",
		zap.String("job_id", jobID),
		zap.String("webhook_url", p.webhookURL),
		zap.Duration("upload_time", processingTime))

	// Retornar jobID - resultado virá via webhook
	return &EnrichmentResult{
		ExtractedText: fmt.Sprintf("Document queued for parsing. Job ID: %s. Result will be sent to webhook.", jobID),
		Metadata: map[string]interface{}{
			"provider":    "llamaparse",
			"job_id":      jobID,
			"webhook_url": p.webhookURL,
			"status":      "queued",
			"file_size":   len(documentData),
		},
		ProcessingTime: processingTime,
	}, nil
}

// Name retorna o nome do provider
func (p *LlamaParseProvider) Name() string {
	return "llamaparse"
}

// SupportsContentType verifica se suporta o tipo de conteúdo
func (p *LlamaParseProvider) SupportsContentType(contentType message_enrichment.EnrichmentContentType) bool {
	return contentType == message_enrichment.EnrichmentTypeDocument ||
		contentType == message_enrichment.EnrichmentTypeImage || // OCR de imagens
		contentType == message_enrichment.EnrichmentTypeAudio    // Transcrição de áudio
}

// SupportsMimeType verifica se suporta o mime type usando registry injetado
// Segue princípios SOLID: delegação para MimeTypeRegistry
func (p *LlamaParseProvider) SupportsMimeType(mimeType string) bool {
	return p.mimeRegistry.IsSupported(mimeType)
}

// GetSupportedMimeTypes retorna todos os mimetypes suportados
func (p *LlamaParseProvider) GetSupportedMimeTypes() []string {
	return p.mimeRegistry.GetSupportedMimeTypes()
}

// GetMimeTypeInfo retorna informações sobre um mimetype específico
func (p *LlamaParseProvider) GetMimeTypeInfo(mimeType string) (*shared.MimeTypeInfo, error) {
	return p.mimeRegistry.GetInfo(mimeType)
}

// downloadDocument baixa o documento do URL
func (p *LlamaParseProvider) downloadDocument(ctx context.Context, mediaURL string) ([]byte, string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", mediaURL, nil)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read response: %w", err)
	}

	// Extrair filename do URL
	filename := filepath.Base(mediaURL)
	if filename == "." || filename == "/" {
		filename = "document.pdf" // Fallback
	}

	return data, filename, nil
}

// uploadDocument envia documento para LlamaParse API (async via webhook)
func (p *LlamaParseProvider) uploadDocument(
	ctx context.Context,
	documentData []byte,
	filename string,
) (string, error) {
	// Criar multipart form
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	// Adicionar webhook URL
	if err := writer.WriteField("webhook_url", p.webhookURL); err != nil {
		return "", fmt.Errorf("failed to write webhook_url field: %w", err)
	}

	// Adicionar arquivo
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return "", fmt.Errorf("failed to create form file: %w", err)
	}

	if _, err := io.Copy(part, bytes.NewReader(documentData)); err != nil {
		return "", fmt.Errorf("failed to copy file data: %w", err)
	}

	// Adicionar parsing instruction (opcional)
	instructions := "Extract all text, tables, and structure from this document. Return as markdown."
	if err := writer.WriteField("parsing_instruction", instructions); err != nil {
		return "", fmt.Errorf("failed to write instructions field: %w", err)
	}

	if err := writer.Close(); err != nil {
		return "", fmt.Errorf("failed to close writer: %w", err)
	}

	// Criar request
	req, err := http.NewRequestWithContext(ctx, "POST", p.apiURL, &requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+p.apiKey)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Enviar request
	resp, err := p.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(respBody))
	}

	// Parse response para obter job_id
	var result struct {
		ID    string `json:"id"`     // job_id
		JobID string `json:"job_id"` // alternativo
	}

	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	jobID := result.ID
	if jobID == "" {
		jobID = result.JobID
	}

	if jobID == "" {
		return "", fmt.Errorf("no job_id in response: %s", string(respBody))
	}

	return jobID, nil
}

// LlamaParseWebhookPayload payload recebido via webhook após parsing
// Este é o formato que será enviado via POST para o webhook_url configurado
type LlamaParseWebhookPayload struct {
	JobID  string `json:"job_id"` // ID do job de parsing
	Status string `json:"status"` // "SUCCESS" ou "ERROR"

	// Resultado do parsing (em caso de sucesso)
	Text     string             `json:"txt"`    // Texto bruto extraído
	Markdown string             `json:"md"`     // Texto formatado em Markdown
	Pages    []LlamaParsePage   `json:"pages"`  // Detalhes por página
	Images   []LlamaParseImageRef `json:"images"` // Referências de imagens

	// Metadata
	ParsedAt time.Time              `json:"parsed_at"`
	Error    string                 `json:"error,omitempty"`     // Mensagem de erro (se status=ERROR)
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// LlamaParsePage detalhes de uma página
type LlamaParsePage struct {
	PageNumber int                  `json:"page"`
	Text       string               `json:"text"`
	Markdown   string               `json:"md"`
	Images     []LlamaParseImageRef `json:"images"`
}

// LlamaParseImageRef referência de imagem extraída
type LlamaParseImageRef struct {
	Name   string `json:"name"`
	Path   string `json:"path"`
	Width  int    `json:"width,omitempty"`
	Height int    `json:"height,omitempty"`
}

// IsConfigured verifica se o provider está configurado
func (p *LlamaParseProvider) IsConfigured() bool {
	return p.apiKey != "" && p.webhookURL != ""
}

// ValidateConfig valida a configuração do provider
func (p *LlamaParseProvider) ValidateConfig() error {
	if p.apiKey == "" {
		return fmt.Errorf("llamaparse API key is required")
	}
	if p.webhookURL == "" {
		return fmt.Errorf("webhook URL is required")
	}
	if len(p.webhookURL) > 200 {
		return fmt.Errorf("webhook URL must be less than 200 characters")
	}
	return nil
}
