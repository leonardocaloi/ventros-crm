package storage

import (
	"context"
	"fmt"
	"io"
	"time"

	"cloud.google.com/go/storage"
	domainstorage "github.com/ventros/crm/internal/domain/storage"
	"go.uber.org/zap"
	"google.golang.org/api/option"
)

// GCSStorage implementa storage.Storage usando Google Cloud Storage
//
// Dependências necessárias:
//
//	go get cloud.google.com/go/storage
//
// Setup:
//  1. Criar bucket no GCS
//  2. Configurar service account
//  3. Gerar credentials JSON
//  4. Set GOOGLE_APPLICATION_CREDENTIALS env var
//
// Exemplo de uso:
//
//	// Criar client
//	ctx := context.Background()
//	gcsClient, err := gcs.NewClient(ctx)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer gcsClient.Close()
//
//	// Criar storage
//	storage := NewGCSStorage(gcsClient, "my-bucket", "https://storage.googleapis.com", logger)
//
//	// Upload
//	file, _ := os.Open("image.jpg")
//	url, err := storage.Upload(ctx, file, "media/images/uuid.jpg", storage.UploadOptions{
//	    ContentType: "image/jpeg",
//	    Public: true,
//	})
//	// url = "https://storage.googleapis.com/my-bucket/media/images/uuid.jpg"
//
// Features a implementar:
// - Upload com retry automático
// - Signed URLs com expiração
// - Metadata customizada
// - ACL (public-read, private)
// - Lifecycle rules (auto-delete arquivos antigos)
// - CDN integration (Cloud CDN)
type GCSStorage struct {
	client    *storage.Client
	bucket    string
	baseURL   string
	logger    *zap.Logger
	projectID string
}

// NewGCSStorage cria uma nova instância de GCSStorage
func NewGCSStorage(ctx context.Context, bucket, projectID string, logger *zap.Logger, opts ...option.ClientOption) (*GCSStorage, error) {
	client, err := storage.NewClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCS client: %w", err)
	}

	baseURL := fmt.Sprintf("https://storage.googleapis.com/%s", bucket)

	return &GCSStorage{
		client:    client,
		bucket:    bucket,
		baseURL:   baseURL,
		logger:    logger,
		projectID: projectID,
	}, nil
}

// Close fecha o client do GCS
func (s *GCSStorage) Close() error {
	return s.client.Close()
}

// Upload faz upload de um arquivo para o GCS
func (s *GCSStorage) Upload(ctx context.Context, file io.Reader, path string, opts domainstorage.UploadOptions) (string, error) {
	// Criar writer
	obj := s.client.Bucket(s.bucket).Object(path)
	w := obj.NewWriter(ctx)
	w.ContentType = opts.ContentType

	// Set ACL based on public/private
	if opts.Public {
		w.ACL = []storage.ACLRule{{Entity: storage.AllUsers, Role: storage.RoleReader}}
	}

	// Set metadata
	if len(opts.Metadata) > 0 {
		w.Metadata = opts.Metadata
	}

	// Copiar dados
	bytesWritten, err := io.Copy(w, file)
	if err != nil {
		return "", &domainstorage.StorageError{
			Op:      "Upload",
			Path:    path,
			Err:     err,
			Message: "failed to copy file data to GCS",
		}
	}

	// Verificar tamanho máximo
	if opts.MaxSize > 0 && bytesWritten > opts.MaxSize {
		return "", &domainstorage.StorageError{
			Op:      "Upload",
			Path:    path,
			Message: fmt.Sprintf("file too large: %d bytes (max: %d)", bytesWritten, opts.MaxSize),
		}
	}

	// Fechar writer e verificar
	if err := w.Close(); err != nil {
		return "", &domainstorage.StorageError{
			Op:      "Upload",
			Path:    path,
			Err:     err,
			Message: "failed to close GCS writer",
		}
	}

	// Retornar URL pública
	url := fmt.Sprintf("%s/%s", s.baseURL, path)

	s.logger.Info("File uploaded to GCS",
		zap.String("path", path),
		zap.String("url", url),
		zap.Int64("bytes", bytesWritten),
		zap.String("content_type", opts.ContentType),
		zap.Bool("public", opts.Public))

	return url, nil
}

// GetSignedURL gera uma URL temporária com acesso limitado
func (s *GCSStorage) GetSignedURL(ctx context.Context, path string, expiry time.Duration) (string, error) {
	opts := &storage.SignedURLOptions{
		Scheme:  storage.SigningSchemeV4,
		Method:  "GET",
		Expires: time.Now().Add(expiry),
	}

	url, err := storage.SignedURL(s.bucket, path, opts)
	if err != nil {
		return "", &domainstorage.StorageError{
			Op:      "GetSignedURL",
			Path:    path,
			Err:     err,
			Message: "failed to generate signed URL",
		}
	}

	s.logger.Debug("Generated signed URL",
		zap.String("path", path),
		zap.Duration("expiry", expiry))

	return url, nil
}

// Delete remove um arquivo do GCS
func (s *GCSStorage) Delete(ctx context.Context, path string) error {
	obj := s.client.Bucket(s.bucket).Object(path)
	if err := obj.Delete(ctx); err != nil {
		if err == storage.ErrObjectNotExist {
			// Não falha se arquivo não existir
			s.logger.Debug("File already deleted or does not exist",
				zap.String("path", path))
			return nil
		}
		return &domainstorage.StorageError{
			Op:      "Delete",
			Path:    path,
			Err:     err,
			Message: "failed to delete object from GCS",
		}
	}

	s.logger.Info("File deleted from GCS",
		zap.String("path", path))

	return nil
}

// Exists verifica se um arquivo existe no GCS
func (s *GCSStorage) Exists(ctx context.Context, path string) (bool, error) {
	obj := s.client.Bucket(s.bucket).Object(path)
	_, err := obj.Attrs(ctx)
	if err == storage.ErrObjectNotExist {
		return false, nil
	}
	if err != nil {
		return false, &domainstorage.StorageError{
			Op:      "Exists",
			Path:    path,
			Err:     err,
			Message: "failed to check object existence",
		}
	}
	return true, nil
}

// Ensure GCSStorage implements domainstorage.Storage
var _ domainstorage.Storage = (*GCSStorage)(nil)

// ===============================================
// Guia de Implementação
// ===============================================

// Passo 1: Adicionar dependência
// go get cloud.google.com/go/storage

// Passo 2: Criar service account no GCP
// https://console.cloud.google.com/iam-admin/serviceaccounts

// Passo 3: Gerar credentials JSON
// - Ir em service account
// - Keys → Add Key → Create new key → JSON
// - Salvar arquivo (ex: gcs-credentials.json)

// Passo 4: Set environment variable
// export GOOGLE_APPLICATION_CREDENTIALS="/path/to/gcs-credentials.json"

// Passo 5: Criar bucket
// gsutil mb gs://ventros-crm-media

// Passo 6: Configurar CORS (se frontend fizer upload direto)
// gsutil cors set cors.json gs://ventros-crm-media

// Arquivo cors.json:
// [
//   {
//     "origin": ["https://app.ventros.cloud"],
//     "method": ["GET", "HEAD", "PUT", "POST"],
//     "responseHeader": ["Content-Type"],
//     "maxAgeSeconds": 3600
//   }
// ]

// Passo 7: Implementar funções acima

// Passo 8: Trocar injeção de dependência
// Antes:
//   storage := storage.NewExternalURLStorage(logger)
// Depois:
//   gcsClient, _ := gcs.NewClient(ctx)
//   storage := storage.NewGCSStorage(gcsClient, "ventros-crm-media", "", logger)

// ===============================================
// Features Avançadas
// ===============================================

// 1. Upload com retry exponential backoff
// import "github.com/cenkalti/backoff/v4"

// 2. Progress callback para upload grandes
// type ProgressFunc func(bytesWritten int64, totalBytes int64)

// 3. Multipart upload para arquivos grandes (> 5MB)

// 4. CDN integration (Cloud CDN)
// baseURL = "https://cdn.ventros.cloud"

// 5. Lifecycle policy (auto-delete arquivos antigos)
// gsutil lifecycle set lifecycle.json gs://ventros-crm-media

// 6. Thumbnail generation (Cloud Functions trigger)

// 7. Virus scanning (ClamAV integration)

// 8. Image optimization (compression, resize)
// import "github.com/disintegration/imaging"

// 9. Video transcoding (FFmpeg worker)

// 10. Analytics (track downloads, bandwidth)
