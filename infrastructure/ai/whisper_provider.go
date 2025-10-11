package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"

	"github.com/caloi/ventros-crm/internal/domain/crm/message_enrichment"
)

// WhisperProvider implementa transcrição de áudio usando OpenAI Whisper API
type WhisperProvider struct {
	logger     *zap.Logger
	apiKey     string
	apiURL     string // Default: https://api.openai.com/v1/audio/transcriptions
	model      string // Default: whisper-1
	httpClient *http.Client
}

// WhisperConfig configuração do Whisper provider
type WhisperConfig struct {
	APIKey     string
	APIURL     string
	Model      string
	TimeoutSec int
}

// NewWhisperProvider cria um novo provider Whisper
func NewWhisperProvider(logger *zap.Logger, config WhisperConfig) *WhisperProvider {
	if config.APIURL == "" {
		config.APIURL = "https://api.openai.com/v1/audio/transcriptions"
	}
	if config.Model == "" {
		config.Model = "whisper-1"
	}
	if config.TimeoutSec == 0 {
		config.TimeoutSec = 120 // 2 minutos default
	}

	return &WhisperProvider{
		logger: logger,
		apiKey: config.APIKey,
		apiURL: config.APIURL,
		model:  config.Model,
		httpClient: &http.Client{
			Timeout: time.Duration(config.TimeoutSec) * time.Second,
		},
	}
}

// Process processa áudio e retorna transcrição
// processingContext é ignorado (não usado para áudio)
func (p *WhisperProvider) Process(
	ctx context.Context,
	mediaURL string,
	contentType message_enrichment.EnrichmentContentType,
	processingContext *string,
) (*EnrichmentResult, error) {
	startTime := time.Now()

	p.logger.Info("Processing audio with Whisper",
		zap.String("media_url", mediaURL),
		zap.String("content_type", string(contentType)))

	// 1. Download do arquivo de áudio
	audioData, filename, err := p.downloadAudio(ctx, mediaURL)
	if err != nil {
		return nil, fmt.Errorf("failed to download audio: %w", err)
	}

	// 2. Transcrever usando Whisper API
	transcription, metadata, err := p.transcribeAudio(ctx, audioData, filename, contentType)
	if err != nil {
		return nil, fmt.Errorf("failed to transcribe audio: %w", err)
	}

	processingTime := time.Since(startTime)

	p.logger.Info("Audio transcribed successfully",
		zap.String("media_url", mediaURL),
		zap.Int("text_length", len(transcription)),
		zap.Duration("processing_time", processingTime))

	return &EnrichmentResult{
		ExtractedText:  transcription,
		Metadata:       metadata,
		ProcessingTime: processingTime,
	}, nil
}

// Name retorna o nome do provider
func (p *WhisperProvider) Name() string {
	return "whisper"
}

// SupportsContentType verifica se suporta o tipo de conteúdo
func (p *WhisperProvider) SupportsContentType(contentType message_enrichment.EnrichmentContentType) bool {
	return contentType == message_enrichment.EnrichmentTypeVoice ||
		contentType == message_enrichment.EnrichmentTypeAudio
}

// SupportsMimeType verifica se suporta um mimetype específico (áudios)
func (p *WhisperProvider) SupportsMimeType(mimeType string) bool {
	supportedMimes := map[string]bool{
		// Áudio formatos suportados pelo Whisper (até 25MB)
		"audio/mpeg":   true, // .mp3
		"audio/mp4":    true, // .m4a
		"audio/wav":    true, // .wav
		"audio/webm":   true, // .webm
		"audio/ogg":    true, // .ogg (WhatsApp PTT)
		"audio/flac":   true, // .flac
		"audio/x-flac": true, // .flac (alt)
		"video/mp4":    true, // .mp4 (áudio extraído)
		"video/mpeg":   true, // .mpeg (áudio extraído)
	}
	return supportedMimes[mimeType]
}

// downloadAudio baixa o arquivo de áudio do URL
func (p *WhisperProvider) downloadAudio(ctx context.Context, mediaURL string) ([]byte, string, error) {
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

	// Extrair nome do arquivo da URL (usado para determinar extensão)
	filename := filepath.Base(mediaURL)
	if filename == "." || filename == "/" {
		filename = "audio.ogg" // Fallback para WhatsApp PTT
	}

	return data, filename, nil
}

// transcribeAudio envia áudio para Whisper API e retorna transcrição
func (p *WhisperProvider) transcribeAudio(
	ctx context.Context,
	audioData []byte,
	filename string,
	contentType message_enrichment.EnrichmentContentType,
) (string, map[string]interface{}, error) {
	// Criar multipart form
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	// Adicionar arquivo de áudio
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return "", nil, fmt.Errorf("failed to create form file: %w", err)
	}

	if _, err := part.Write(audioData); err != nil {
		return "", nil, fmt.Errorf("failed to write audio data: %w", err)
	}

	// Adicionar model
	if err := writer.WriteField("model", p.model); err != nil {
		return "", nil, fmt.Errorf("failed to write model field: %w", err)
	}

	// Adicionar response_format (verbose_json para metadata)
	if err := writer.WriteField("response_format", "verbose_json"); err != nil {
		return "", nil, fmt.Errorf("failed to write response_format field: %w", err)
	}

	// Adicionar language hint (português) para melhor precisão
	if err := writer.WriteField("language", "pt"); err != nil {
		return "", nil, fmt.Errorf("failed to write language field: %w", err)
	}

	// Adicionar prompt hint para PTT (mensagens de voz curtas)
	if contentType == message_enrichment.EnrichmentTypeVoice {
		prompt := "Esta é uma mensagem de voz do WhatsApp."
		if err := writer.WriteField("prompt", prompt); err != nil {
			return "", nil, fmt.Errorf("failed to write prompt field: %w", err)
		}
	}

	if err := writer.Close(); err != nil {
		return "", nil, fmt.Errorf("failed to close writer: %w", err)
	}

	// Criar request
	req, err := http.NewRequestWithContext(ctx, "POST", p.apiURL, &requestBody)
	if err != nil {
		return "", nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+p.apiKey)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Enviar request
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
	var result WhisperResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Extrair metadados
	metadata := map[string]interface{}{
		"model":    p.model,
		"language": result.Language,
		"duration": result.Duration,
	}

	// Adicionar segments se disponível
	if len(result.Segments) > 0 {
		metadata["segment_count"] = len(result.Segments)
		metadata["segments"] = result.Segments
	}

	return result.Text, metadata, nil
}

// WhisperResponse estrutura da resposta da API Whisper (verbose_json)
type WhisperResponse struct {
	Text     string           `json:"text"`
	Language string           `json:"language"`
	Duration float64          `json:"duration"`
	Segments []WhisperSegment `json:"segments,omitempty"`
	Words    []WhisperWord    `json:"words,omitempty"`
}

// WhisperSegment representa um segmento de transcrição
type WhisperSegment struct {
	ID               int     `json:"id"`
	Seek             int     `json:"seek"`
	Start            float64 `json:"start"`
	End              float64 `json:"end"`
	Text             string  `json:"text"`
	Temperature      float64 `json:"temperature"`
	AvgLogprob       float64 `json:"avg_logprob"`
	CompressionRatio float64 `json:"compression_ratio"`
	NoSpeechProb     float64 `json:"no_speech_prob"`
}

// WhisperWord representa uma palavra individual (word-level timestamps)
type WhisperWord struct {
	Word  string  `json:"word"`
	Start float64 `json:"start"`
	End   float64 `json:"end"`
}

// IsConfigured verifica se o provider está configurado (tem API key)
func (p *WhisperProvider) IsConfigured() bool {
	return p.apiKey != ""
}

// ValidateConfig valida a configuração do provider
func (p *WhisperProvider) ValidateConfig() error {
	if p.apiKey == "" {
		return fmt.Errorf("whisper API key is required")
	}

	// Verificar se a API key é válida (formato básico)
	if len(p.apiKey) < 20 {
		return fmt.Errorf("whisper API key appears to be invalid")
	}

	return nil
}

// SaveAudioToTemp salva áudio em arquivo temporário (útil para debugging)
func (p *WhisperProvider) SaveAudioToTemp(audioData []byte, filename string) (string, error) {
	tempFile := filepath.Join(os.TempDir(), "whisper_"+filename)

	if err := os.WriteFile(tempFile, audioData, 0644); err != nil {
		return "", fmt.Errorf("failed to save temp file: %w", err)
	}

	p.logger.Debug("Audio saved to temp file",
		zap.String("path", tempFile),
		zap.Int("size_bytes", len(audioData)))

	return tempFile, nil
}
