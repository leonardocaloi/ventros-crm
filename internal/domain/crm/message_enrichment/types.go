package message_enrichment

// EnrichmentContentType representa o tipo de conteúdo a ser enriquecido
type EnrichmentContentType string

const (
	// EnrichmentTypeAudio - Áudio geral (música, podcasts, etc)
	EnrichmentTypeAudio EnrichmentContentType = "audio"

	// EnrichmentTypeVoice - Mensagens de voz (PTT) - PRIORIDADE MÁXIMA
	EnrichmentTypeVoice EnrichmentContentType = "voice"

	// EnrichmentTypeImage - Imagens (fotos, screenshots, memes)
	EnrichmentTypeImage EnrichmentContentType = "image"

	// EnrichmentTypeVideo - Vídeos
	EnrichmentTypeVideo EnrichmentContentType = "video"

	// EnrichmentTypeDocument - Documentos (PDF, DOCX, XLSX, etc)
	EnrichmentTypeDocument EnrichmentContentType = "document"
)

// EnrichmentProvider representa o provedor de IA usado para enriquecimento
type EnrichmentProvider string

const (
	// ProviderWhisper - OpenAI Whisper (transcription)
	ProviderWhisper EnrichmentProvider = "whisper"

	// ProviderDeepgram - Deepgram (transcription with real-time)
	ProviderDeepgram EnrichmentProvider = "deepgram"

	// ProviderVision - GPT-4 Vision ou Gemini Vision (image description + OCR)
	ProviderVision EnrichmentProvider = "vision"

	// ProviderLlamaParse - LlamaParse (document parsing)
	ProviderLlamaParse EnrichmentProvider = "llamaparse"

	// ProviderFFmpeg - FFmpeg (video processing - extract audio + frames)
	ProviderFFmpeg EnrichmentProvider = "ffmpeg"

	// ProviderTesseract - Tesseract OCR (fallback para imagens simples)
	ProviderTesseract EnrichmentProvider = "tesseract"
)

// EnrichmentStatus representa o status do processamento
type EnrichmentStatus string

const (
	// StatusPending - Aguardando processamento
	StatusPending EnrichmentStatus = "pending"

	// StatusProcessing - Em processamento
	StatusProcessing EnrichmentStatus = "processing"

	// StatusCompleted - Processamento concluído com sucesso
	StatusCompleted EnrichmentStatus = "completed"

	// StatusFailed - Processamento falhou
	StatusFailed EnrichmentStatus = "failed"
)

// Priority retorna a prioridade do tipo de conteúdo para fila RabbitMQ
func (t EnrichmentContentType) Priority() uint8 {
	switch t {
	case EnrichmentTypeVoice:
		return 10 // Máxima prioridade (PTT)
	case EnrichmentTypeAudio:
		return 8
	case EnrichmentTypeImage:
		return 7
	case EnrichmentTypeDocument:
		return 6
	case EnrichmentTypeVideo:
		return 3 // Menor prioridade (processamento pesado)
	default:
		return 5
	}
}

// String returns the string representation
func (t EnrichmentContentType) String() string {
	return string(t)
}

// String returns the string representation
func (p EnrichmentProvider) String() string {
	return string(p)
}

// String returns the string representation
func (s EnrichmentStatus) String() string {
	return string(s)
}

// IsValid validates if content type is valid
func (t EnrichmentContentType) IsValid() bool {
	switch t {
	case EnrichmentTypeAudio, EnrichmentTypeVoice, EnrichmentTypeImage,
		EnrichmentTypeVideo, EnrichmentTypeDocument:
		return true
	}
	return false
}

// IsValid validates if provider is valid
func (p EnrichmentProvider) IsValid() bool {
	switch p {
	case ProviderWhisper, ProviderDeepgram, ProviderVision,
		ProviderLlamaParse, ProviderFFmpeg, ProviderTesseract:
		return true
	}
	return false
}

// IsValid validates if status is valid
func (s EnrichmentStatus) IsValid() bool {
	switch s {
	case StatusPending, StatusProcessing, StatusCompleted, StatusFailed:
		return true
	}
	return false
}

// IsFinal returns true if status is final (completed or failed)
func (s EnrichmentStatus) IsFinal() bool {
	return s == StatusCompleted || s == StatusFailed
}
