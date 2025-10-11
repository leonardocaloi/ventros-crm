package ai

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"go.uber.org/zap"
)

// AudioSplitter quebra áudios/vídeos longos em partes menores usando ffmpeg
type AudioSplitter struct {
	logger       *zap.Logger
	ffmpegPath   string
	ffprobePath  string
	tempDir      string
}

// NewAudioSplitter cria um novo splitter de áudio
func NewAudioSplitter(logger *zap.Logger) *AudioSplitter {
	return &AudioSplitter{
		logger:      logger,
		ffmpegPath:  "ffmpeg",  // Assumindo que ffmpeg está no PATH
		ffprobePath: "ffprobe", // Assumindo que ffprobe está no PATH
		tempDir:     "/tmp/ventros-audio-splits",
	}
}

// SplitByS silence quebra áudio em partes usando detecção de silêncio
func (s *AudioSplitter) SplitBySilence(ctx context.Context, inputPath string, silenceThreshold float64) ([]string, error) {
	// Criar diretório temporário se não existir
	if err := os.MkdirAll(s.tempDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %w", err)
	}

	// 1. Detectar pontos de silêncio usando ffmpeg silencedetect
	silencePoints, err := s.detectSilence(ctx, inputPath, silenceThreshold)
	if err != nil {
		return nil, fmt.Errorf("failed to detect silence: %w", err)
	}

	s.logger.Info("Detected silence points",
		zap.String("input", inputPath),
		zap.Int("points", len(silencePoints)),
		zap.Float64("threshold", silenceThreshold))

	// Se não encontrou pontos de silêncio suficientes, retorna o arquivo original
	if len(silencePoints) < 2 {
		s.logger.Info("Not enough silence points, returning original file")
		return []string{inputPath}, nil
	}

	// 2. Quebrar áudio nos pontos de silêncio
	segments, err := s.splitAtPoints(ctx, inputPath, silencePoints)
	if err != nil {
		return nil, fmt.Errorf("failed to split audio: %w", err)
	}

	s.logger.Info("Audio split completed",
		zap.String("input", inputPath),
		zap.Int("segments", len(segments)))

	return segments, nil
}

// detectSilence detecta pontos de silêncio no áudio
func (s *AudioSplitter) detectSilence(ctx context.Context, inputPath string, threshold float64) ([]float64, error) {
	// Converter threshold (0-1) para dB (ffmpeg usa -60dB a 0dB)
	// threshold 0.3 -> -40dB aprox
	thresholdDB := -60 + (threshold * 60)

	// ffmpeg -i input.mp3 -af silencedetect=noise=-40dB:d=0.5 -f null -
	cmd := exec.CommandContext(ctx, s.ffmpegPath,
		"-i", inputPath,
		"-af", fmt.Sprintf("silencedetect=noise=%.1fdB:d=0.5", thresholdDB),
		"-f", "null",
		"-",
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		// ffmpeg retorna erro mesmo quando funciona (exit code 1)
		// Vamos ignorar e processar o output
		s.logger.Debug("ffmpeg silencedetect output (expected error)",
			zap.Error(err),
			zap.String("output", string(output)))
	}

	// Parsear output para extrair timestamps de silêncio
	points := s.parseSilenceOutput(string(output))

	return points, nil
}

// parseSilenceOutput parseia output do silencedetect
func (s *AudioSplitter) parseSilenceOutput(output string) []float64 {
	var points []float64

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		// Procurar por "silence_end: 123.456"
		if strings.Contains(line, "silence_end:") {
			parts := strings.Split(line, "silence_end:")
			if len(parts) < 2 {
				continue
			}

			// Extrair timestamp
			timestampStr := strings.TrimSpace(strings.Split(parts[1], "|")[0])
			timestamp, err := strconv.ParseFloat(timestampStr, 64)
			if err != nil {
				s.logger.Warn("Failed to parse silence timestamp",
					zap.String("timestamp", timestampStr),
					zap.Error(err))
				continue
			}

			points = append(points, timestamp)
		}
	}

	return points
}

// splitAtPoints quebra o áudio nos pontos especificados
func (s *AudioSplitter) splitAtPoints(ctx context.Context, inputPath string, points []float64) ([]string, error) {
	var segments []string

	// Obter duração total do áudio
	duration, err := s.getDuration(ctx, inputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get duration: %w", err)
	}

	// Adicionar início (0) e fim (duration) aos pontos
	allPoints := append([]float64{0}, points...)
	allPoints = append(allPoints, duration)

	// Criar segmentos
	baseFilename := filepath.Base(inputPath)
	ext := filepath.Ext(baseFilename)
	nameWithoutExt := strings.TrimSuffix(baseFilename, ext)

	for i := 0; i < len(allPoints)-1; i++ {
		start := allPoints[i]
		end := allPoints[i+1]

		// Evitar segmentos muito curtos (< 1 segundo)
		if end-start < 1.0 {
			continue
		}

		outputPath := filepath.Join(s.tempDir, fmt.Sprintf("%s_part_%03d%s", nameWithoutExt, i+1, ext))

		// ffmpeg -i input.mp3 -ss 0 -to 60 -c copy output.mp3
		cmd := exec.CommandContext(ctx, s.ffmpegPath,
			"-i", inputPath,
			"-ss", fmt.Sprintf("%.2f", start),
			"-to", fmt.Sprintf("%.2f", end),
			"-c", "copy",
			"-y", // Sobrescrever se existir
			outputPath,
		)

		if err := cmd.Run(); err != nil {
			s.logger.Error("Failed to create segment",
				zap.Error(err),
				zap.Float64("start", start),
				zap.Float64("end", end))
			continue
		}

		segments = append(segments, outputPath)

		s.logger.Debug("Created audio segment",
			zap.String("output", outputPath),
			zap.Float64("start", start),
			zap.Float64("end", end),
			zap.Float64("duration", end-start))
	}

	return segments, nil
}

// getDuration obtém duração do áudio/vídeo usando ffprobe
func (s *AudioSplitter) getDuration(ctx context.Context, inputPath string) (float64, error) {
	// ffprobe -v error -show_entries format=duration -of default=noprint_wrappers=1:nokey=1 input.mp3
	cmd := exec.CommandContext(ctx, s.ffprobePath,
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		inputPath,
	)

	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("failed to get duration: %w", err)
	}

	durationStr := strings.TrimSpace(string(output))
	duration, err := strconv.ParseFloat(durationStr, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse duration: %w", err)
	}

	return duration, nil
}

// CleanupSegments remove arquivos temporários de segmentos
func (s *AudioSplitter) CleanupSegments(segments []string) {
	for _, segment := range segments {
		if err := os.Remove(segment); err != nil {
			s.logger.Warn("Failed to remove segment file",
				zap.String("file", segment),
				zap.Error(err))
		}
	}
}
