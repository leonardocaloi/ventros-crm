package ai

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"go.uber.org/zap"

	"github.com/caloi/ventros-crm/internal/domain/crm/message_enrichment"
)

// FFmpegProvider implementa extração de áudio de vídeos usando FFmpeg
// Este provider NÃO faz transcrição - apenas extrai o áudio
// O áudio extraído deve ser processado por WhisperProvider depois
type FFmpegProvider struct {
	logger       *zap.Logger
	ffmpegPath   string
	tempDir      string
	httpClient   *http.Client
	audioCodec   string // Codec de áudio de saída (default: libmp3lame)
	audioFormat  string // Formato de áudio de saída (default: mp3)
	audioBitrate string // Bitrate de áudio (default: 128k)
}

// FFmpegConfig configuração do FFmpeg provider
type FFmpegConfig struct {
	FFmpegPath   string
	TempDir      string
	TimeoutSec   int
	AudioCodec   string
	AudioFormat  string
	AudioBitrate string
}

// NewFFmpegProvider cria um novo provider FFmpeg
func NewFFmpegProvider(logger *zap.Logger, config FFmpegConfig) *FFmpegProvider {
	if config.FFmpegPath == "" {
		config.FFmpegPath = "ffmpeg" // Assume ffmpeg no PATH
	}
	if config.TempDir == "" {
		config.TempDir = os.TempDir()
	}
	if config.TimeoutSec == 0 {
		config.TimeoutSec = 600 // 10 minutos (vídeos podem ser grandes)
	}
	if config.AudioCodec == "" {
		config.AudioCodec = "libmp3lame"
	}
	if config.AudioFormat == "" {
		config.AudioFormat = "mp3"
	}
	if config.AudioBitrate == "" {
		config.AudioBitrate = "128k"
	}

	return &FFmpegProvider{
		logger:       logger,
		ffmpegPath:   config.FFmpegPath,
		tempDir:      config.TempDir,
		audioCodec:   config.AudioCodec,
		audioFormat:  config.AudioFormat,
		audioBitrate: config.AudioBitrate,
		httpClient: &http.Client{
			Timeout: time.Duration(config.TimeoutSec) * time.Second,
		},
	}
}

// Process processa vídeo e extrai áudio
// IMPORTANTE: O texto extraído será vazio - este provider apenas prepara o áudio
// O áudio extraído deve ser processado por outro provider (Whisper) depois
// processingContext é ignorado (não usado para vídeo)
func (p *FFmpegProvider) Process(
	ctx context.Context,
	mediaURL string,
	contentType message_enrichment.EnrichmentContentType,
	processingContext *string,
) (*EnrichmentResult, error) {
	startTime := time.Now()

	p.logger.Info("Processing video with FFmpeg",
		zap.String("media_url", mediaURL))

	// 1. Download do vídeo
	videoData, err := p.downloadVideo(ctx, mediaURL)
	if err != nil {
		return nil, fmt.Errorf("failed to download video: %w", err)
	}

	// 2. Salvar vídeo temporário
	videoPath, err := p.saveToTemp(videoData, "video.mp4")
	if err != nil {
		return nil, fmt.Errorf("failed to save video: %w", err)
	}
	defer os.Remove(videoPath)

	// 3. Extrair áudio usando FFmpeg
	audioPath, duration, err := p.extractAudio(ctx, videoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to extract audio: %w", err)
	}
	defer os.Remove(audioPath)

	// 4. Ler áudio extraído
	audioData, err := os.ReadFile(audioPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read extracted audio: %w", err)
	}

	processingTime := time.Since(startTime)

	// Metadados incluem o áudio extraído (base64) para processamento posterior
	metadata := map[string]interface{}{
		"provider":        "ffmpeg",
		"video_size":      len(videoData),
		"audio_size":      len(audioData),
		"audio_format":    p.audioFormat,
		"audio_codec":     p.audioCodec,
		"audio_bitrate":   p.audioBitrate,
		"duration":        duration,
		"extracted_audio": audioPath, // Path temporário do áudio
	}

	p.logger.Info("Video audio extracted successfully",
		zap.String("media_url", mediaURL),
		zap.Int("audio_size", len(audioData)),
		zap.Float64("duration", duration),
		zap.Duration("processing_time", processingTime))

	// NOTA: extractedText fica vazio - o áudio precisa ser transcrito depois
	return &EnrichmentResult{
		ExtractedText:  fmt.Sprintf("[Áudio extraído do vídeo - duração: %.2fs]", duration),
		Metadata:       metadata,
		ProcessingTime: processingTime,
	}, nil
}

// Name retorna o nome do provider
func (p *FFmpegProvider) Name() string {
	return "ffmpeg"
}

// SupportsContentType verifica se suporta o tipo de conteúdo
func (p *FFmpegProvider) SupportsContentType(contentType message_enrichment.EnrichmentContentType) bool {
	return contentType == message_enrichment.EnrichmentTypeVideo
}

// downloadVideo baixa o vídeo do URL
func (p *FFmpegProvider) downloadVideo(ctx context.Context, mediaURL string) ([]byte, error) {
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

// saveToTemp salva dados em arquivo temporário
func (p *FFmpegProvider) saveToTemp(data []byte, filename string) (string, error) {
	tempFile := filepath.Join(p.tempDir, fmt.Sprintf("ffmpeg_%d_%s", time.Now().Unix(), filename))

	if err := os.WriteFile(tempFile, data, 0644); err != nil {
		return "", fmt.Errorf("failed to save temp file: %w", err)
	}

	p.logger.Debug("Saved temp file",
		zap.String("path", tempFile),
		zap.Int("size_bytes", len(data)))

	return tempFile, nil
}

// extractAudio extrai áudio do vídeo usando FFmpeg
func (p *FFmpegProvider) extractAudio(ctx context.Context, videoPath string) (string, float64, error) {
	// Path do áudio de saída
	audioPath := filepath.Join(p.tempDir, fmt.Sprintf("audio_%d.%s", time.Now().Unix(), p.audioFormat))

	// Construir comando FFmpeg
	args := []string{
		"-i", videoPath, // Input file
		"-vn",                   // Disable video
		"-acodec", p.audioCodec, // Audio codec
		"-ab", p.audioBitrate, // Audio bitrate
		"-ar", "16000", // Sample rate 16kHz (ideal para Whisper)
		"-ac", "1", // Mono
		"-y",      // Overwrite output
		audioPath, // Output file
	}

	p.logger.Debug("Running FFmpeg",
		zap.String("video_path", videoPath),
		zap.String("audio_path", audioPath),
		zap.Strings("args", args))

	// Executar FFmpeg com timeout
	cmd := exec.CommandContext(ctx, p.ffmpegPath, args...)

	// Capturar stderr para logs
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return "", 0, fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return "", 0, fmt.Errorf("failed to start ffmpeg: %w", err)
	}

	// Ler stderr (FFmpeg envia output para stderr)
	stderrData, _ := io.ReadAll(stderr)

	if err := cmd.Wait(); err != nil {
		p.logger.Error("FFmpeg failed",
			zap.Error(err),
			zap.String("stderr", string(stderrData)))
		return "", 0, fmt.Errorf("ffmpeg failed: %w", err)
	}

	// Verificar se áudio foi criado
	if _, err := os.Stat(audioPath); os.IsNotExist(err) {
		return "", 0, fmt.Errorf("audio file was not created")
	}

	// Obter duração do áudio (executar ffprobe ou estimar)
	duration := p.getAudioDuration(ctx, audioPath)

	p.logger.Info("Audio extracted successfully",
		zap.String("audio_path", audioPath),
		zap.Float64("duration", duration))

	return audioPath, duration, nil
}

// getAudioDuration obtém duração do áudio usando ffprobe
func (p *FFmpegProvider) getAudioDuration(ctx context.Context, audioPath string) float64 {
	// Tentar usar ffprobe
	ffprobePath := "ffprobe"

	cmd := exec.CommandContext(ctx, ffprobePath,
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		audioPath,
	)

	output, err := cmd.Output()
	if err != nil {
		p.logger.Warn("Failed to get duration with ffprobe", zap.Error(err))
		return 0
	}

	var duration float64
	if _, err := fmt.Sscanf(string(output), "%f", &duration); err != nil {
		p.logger.Warn("Failed to parse duration", zap.Error(err))
		return 0
	}

	return duration
}

// IsConfigured verifica se o provider está configurado
func (p *FFmpegProvider) IsConfigured() bool {
	// Verificar se FFmpeg está disponível
	cmd := exec.Command(p.ffmpegPath, "-version")
	if err := cmd.Run(); err != nil {
		return false
	}
	return true
}

// ValidateConfig valida a configuração do provider
func (p *FFmpegProvider) ValidateConfig() error {
	// Verificar se FFmpeg existe
	cmd := exec.Command(p.ffmpegPath, "-version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ffmpeg not found or not executable: %w", err)
	}

	// Verificar se temp dir existe e é gravável
	if err := os.MkdirAll(p.tempDir, 0755); err != nil {
		return fmt.Errorf("temp directory not accessible: %w", err)
	}

	return nil
}
