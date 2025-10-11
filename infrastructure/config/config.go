package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

// Config holds all application configuration
type Config struct {
	Server               ServerConfig
	Database             DatabaseConfig
	Redis                RedisConfig
	RabbitMQ             RabbitMQConfig
	Temporal             TemporalConfig
	Log                  LogConfig
	Session              SessionConfig
	Encryption           EncryptionConfig
	RateLimit            RateLimitConfig
	WAHA                 WAHAConfig
	AI                   AIConfig
	UseSagaOrchestration bool // Feature flag: Saga Orchestration (Temporal workflows)
}

type ServerConfig struct {
	Port string
	Env  string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

type RabbitMQConfig struct {
	URL string
}

type TemporalConfig struct {
	Host      string
	Namespace string
}

type LogConfig struct {
	Level string
}

type SessionConfig struct {
	DefaultTimeoutMinutes int // Timeout padrão para sessões em minutos
}

type EncryptionConfig struct {
	AESKey string // Base64-encoded 32-byte key for AES-256-GCM
}

type RateLimitConfig struct {
	// Global rate limits (per IP)
	GlobalMaxRequests   int // Maximum requests per window
	GlobalWindowSeconds int // Window duration in seconds

	// Auth endpoints (login, register)
	AuthMaxRequests   int
	AuthWindowSeconds int

	// Authenticated user endpoints
	AuthenticatedMaxRequests   int
	AuthenticatedWindowSeconds int

	// Webhook endpoints
	WebhookMaxRequests   int
	WebhookWindowSeconds int
}

// WAHAConfig holds WAHA-specific configuration
type WAHAConfig struct {
	// DefaultSessionID é o session ID padrão usado para operações WAHA
	// como converter números, checar status, etc.
	// Exemplo: "guilherme-batilani-suporte"
	DefaultSessionID string
}

// AIConfig holds AI provider configuration
type AIConfig struct {
	// Vertex AI (cloud.google.com) - Gemini Vision para imagens
	VertexProjectID      string // Google Cloud Project ID
	VertexLocation       string // Região (ex: us-central1)
	VertexServiceAccount string // Path para JSON do Service Account
	VertexModel          string // Model ID (ex: gemini-1.5-flash)

	// LlamaParse (documentos PDF, Word, Excel, PowerPoint)
	LlamaParseAPIKey     string // LlamaParse API key
	LlamaParseWebhookURL string // Webhook URL para callback
	LlamaParseModel      string // Model (opcional, default: "default")

	// Groq Whisper (áudio falado/PTT) - PRIORIDADE 1 - GRATUITO
	GroqAPIKey       string // Groq API key (gsk_...)
	GroqWhisperModel string // Model (default: whisper-large-v3-turbo)

	// OpenAI Whisper (áudio falado/PTT) - PRIORIDADE 2 - FALLBACK
	OpenAIAPIKey       string // OpenAI API key (sk_...)
	OpenAIWhisperModel string // Model (default: whisper-1)

	// Claude (text analysis) - opcional
	ClaudeAPIKey string // Anthropic API key

	// GPT-4 Vision (video) - opcional, disabled by default
	GPT4VisionAPIKey string // OpenAI API key for GPT-4 Vision
}

// Load loads configuration from environment variables
// Automatically loads .env file if it exists (development)
func Load() *Config {
	// Tenta carregar .env (ignora erro se não existir - útil em produção)
	_ = godotenv.Load()

	return &Config{
		Server: ServerConfig{
			Port: getEnv("PORT", "8080"),
			Env:  getEnv("ENV", "development"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "ventros"),
			Password: getEnv("DB_PASSWORD", "ventros123"),
			Name:     getEnv("DB_NAME", "ventros_crm"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       0,
		},
		RabbitMQ: RabbitMQConfig{
			URL: getRabbitMQURL(),
		},
		Temporal: TemporalConfig{
			Host:      getEnv("TEMPORAL_HOST", "localhost:7233"),
			Namespace: getEnv("TEMPORAL_NAMESPACE", "ventros-crm"),
		},
		Log:                  LogConfig{},
		Session:              SessionConfig{DefaultTimeoutMinutes: getEnvInt("SESSION_DEFAULT_TIMEOUT_MINUTES", 30)},
		Encryption:           EncryptionConfig{AESKey: getEnv("ENCRYPTION_AES_KEY", "")},
		UseSagaOrchestration: getEnv("USE_SAGA_ORCHESTRATION", "false") == "true",
		RateLimit: RateLimitConfig{
			GlobalMaxRequests:          getEnvInt("RATE_LIMIT_GLOBAL_MAX", 1000),
			GlobalWindowSeconds:        getEnvInt("RATE_LIMIT_GLOBAL_WINDOW", 60),
			AuthMaxRequests:            getEnvInt("RATE_LIMIT_AUTH_MAX", 10),
			AuthWindowSeconds:          getEnvInt("RATE_LIMIT_AUTH_WINDOW", 60),
			AuthenticatedMaxRequests:   getEnvInt("RATE_LIMIT_AUTHENTICATED_MAX", 500),
			AuthenticatedWindowSeconds: getEnvInt("RATE_LIMIT_AUTHENTICATED_WINDOW", 60),
			WebhookMaxRequests:         getEnvInt("RATE_LIMIT_WEBHOOK_MAX", 100),
			WebhookWindowSeconds:       getEnvInt("RATE_LIMIT_WEBHOOK_WINDOW", 60),
		},
		WAHA: WAHAConfig{
			DefaultSessionID: getEnv("WAHA_DEFAULT_SESSION_ID", ""),
		},
		AI: AIConfig{
			// Vertex AI (cloud.google.com) - Gemini Vision para imagens
			VertexProjectID:      getEnv("VERTEX_PROJECT_ID", ""),
			VertexLocation:       getEnv("VERTEX_LOCATION", "us-central1"),
			VertexServiceAccount: getEnv("VERTEX_SERVICE_ACCOUNT", ""),
			VertexModel:          getEnv("VERTEX_MODEL", "gemini-1.5-flash"),

			// LlamaParse (documentos PDF, Word, Excel, PowerPoint)
			LlamaParseAPIKey:     getEnv("LLAMAPARSE_API_KEY", ""),
			LlamaParseWebhookURL: getEnv("LLAMAPARSE_WEBHOOK_URL", ""),
			LlamaParseModel:      getEnv("LLAMAPARSE_MODEL", "default"),

			// Groq Whisper (áudio falado/PTT) - PRIORIDADE 1 - GRATUITO
			GroqAPIKey:       getEnv("GROQ_API_KEY", ""),
			GroqWhisperModel: getEnv("GROQ_WHISPER_MODEL", "whisper-large-v3-turbo"),

			// OpenAI Whisper (áudio falado/PTT) - PRIORIDADE 2 - FALLBACK
			OpenAIAPIKey:       getEnv("OPENAI_API_KEY", ""),
			OpenAIWhisperModel: getEnv("OPENAI_WHISPER_MODEL", "whisper-1"),

			// Claude (text analysis) - opcional
			ClaudeAPIKey: getEnv("CLAUDE_API_KEY", ""),

			// GPT-4 Vision (video) - opcional, disabled by default
			GPT4VisionAPIKey: getEnv("GPT4_VISION_API_KEY", ""),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		var intValue int
		if _, err := fmt.Sscanf(value, "%d", &intValue); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getRabbitMQURL constrói a URL do RabbitMQ usando variáveis de ambiente separadas
// ou retorna RABBITMQ_URL se estiver definida (para compatibilidade)
func getRabbitMQURL() string {
	// Se RABBITMQ_URL já está definida, usa ela
	if url := os.Getenv("RABBITMQ_URL"); url != "" {
		return url
	}

	// Caso contrário, constrói a URL usando variáveis separadas
	host := getEnv("RABBITMQ_HOST", "localhost")
	port := getEnv("RABBITMQ_PORT", "5672")
	user := getEnv("RABBITMQ_USER", "guest")
	password := getEnv("RABBITMQ_PASSWORD", "guest")

	return fmt.Sprintf("amqp://%s:%s@%s:%s/", user, password, host, port)
}
