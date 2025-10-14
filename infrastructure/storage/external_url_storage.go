package storage

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"time"

	"github.com/ventros/crm/internal/domain/storage"
	"go.uber.org/zap"
)

// ExternalURLStorage é uma implementação "passthrough" que aceita URLs externas.
//
// USO ATUAL:
// - Validar URLs fornecidas pelo usuário
// - Retornar a mesma URL (sem upload real)
// - Permitir que frontend já use endpoint correto
//
// MIGRAÇÃO FUTURA:
// - Trocar por GCSStorage, S3Storage, MinIOStorage
// - Zero breaking changes no código
// - Apenas troca injeção de dependência
//
// Exemplo:
//
//	// Agora (ExternalURLStorage)
//	storage := storage.NewExternalURLStorage(logger)
//	url, _ := storage.Upload(ctx, nil, "", storage.UploadOptions{
//	    Metadata: map[string]string{
//	        "external_url": "https://storage.googleapis.com/bucket/file.jpg"
//	    },
//	})
//	// url = "https://storage.googleapis.com/bucket/file.jpg"
//
//	// Depois (GCSStorage)
//	storage := storage.NewGCSStorage(client, bucket, logger)
//	url, _ := storage.Upload(ctx, fileReader, "media/file.jpg", opts)
//	// url = "https://storage.googleapis.com/my-bucket/media/file.jpg"
type ExternalURLStorage struct {
	logger *zap.Logger
}

// NewExternalURLStorage cria uma nova instância de ExternalURLStorage
func NewExternalURLStorage(logger *zap.Logger) *ExternalURLStorage {
	return &ExternalURLStorage{
		logger: logger,
	}
}

// Upload valida e retorna uma URL externa.
//
// IMPORTANTE: Não faz upload real!
// A URL deve ser fornecida em opts.Metadata["external_url"]
//
// Validações:
// - URL deve ser válida (parse)
// - URL deve ser HTTP/HTTPS
// - URL deve ser acessível (opcional)
//
// Uso:
//
//	url, err := storage.Upload(ctx, nil, "", storage.UploadOptions{
//	    Metadata: map[string]string{
//	        "external_url": "https://cdn.example.com/image.jpg",
//	    },
//	})
func (s *ExternalURLStorage) Upload(ctx context.Context, file io.Reader, path string, opts storage.UploadOptions) (string, error) {
	// Extrai URL externa do metadata
	externalURL, ok := opts.Metadata["external_url"]
	if !ok || externalURL == "" {
		return "", &storage.StorageError{
			Op:      "Upload",
			Path:    path,
			Message: "external_url required in metadata (ExternalURLStorage is passthrough)",
		}
	}

	// Valida URL
	parsedURL, err := url.Parse(externalURL)
	if err != nil {
		return "", &storage.StorageError{
			Op:      "Upload",
			Path:    path,
			Err:     err,
			Message: fmt.Sprintf("invalid URL: %s", externalURL),
		}
	}

	// Valida scheme (apenas HTTP/HTTPS)
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return "", &storage.StorageError{
			Op:      "Upload",
			Path:    path,
			Message: fmt.Sprintf("invalid URL scheme: %s (must be http or https)", parsedURL.Scheme),
		}
	}

	s.logger.Debug("ExternalURLStorage: URL validated (passthrough)",
		zap.String("url", externalURL),
		zap.String("path", path),
		zap.String("content_type", opts.ContentType))

	// Retorna a mesma URL (passthrough)
	return externalURL, nil
}

// GetSignedURL retorna a mesma URL (sem expiração)
//
// NOTA: URLs externas não têm signed URLs.
// Esta implementação apenas retorna a URL original.
func (s *ExternalURLStorage) GetSignedURL(ctx context.Context, path string, expiry time.Duration) (string, error) {
	s.logger.Warn("ExternalURLStorage: GetSignedURL called but not supported (passthrough)",
		zap.String("path", path),
		zap.Duration("expiry", expiry))

	// Como não temos storage real, retornamos o path como URL
	// (assumindo que é uma URL completa)
	return path, nil
}

// Delete não faz nada (não temos controle sobre URLs externas)
func (s *ExternalURLStorage) Delete(ctx context.Context, path string) error {
	s.logger.Debug("ExternalURLStorage: Delete called but not supported (passthrough)",
		zap.String("path", path))

	// Não faz nada, mas não falha
	return nil
}

// Exists sempre retorna true (assumimos que URL externa existe)
func (s *ExternalURLStorage) Exists(ctx context.Context, path string) (bool, error) {
	s.logger.Debug("ExternalURLStorage: Exists called (passthrough)",
		zap.String("path", path))

	// Sempre assume que existe
	// Para validar de verdade, precisaria fazer HTTP HEAD request
	return true, nil
}

// Ensure ExternalURLStorage implements storage.Storage
var _ storage.Storage = (*ExternalURLStorage)(nil)
