package config

import (
	"fmt"
	"os"
)

// Config holds all application configuration
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	RabbitMQ RabbitMQConfig
	Temporal TemporalConfig
	Log      LogConfig
	Session  SessionConfig
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

// Load loads configuration from environment variables
func Load() *Config {
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
		Log: LogConfig{
		},
		Session: SessionConfig{
			DefaultTimeoutMinutes: getEnvInt("SESSION_DEFAULT_TIMEOUT_MINUTES", 30),
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
