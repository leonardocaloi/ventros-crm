package storage

import (
	"context"
	"io"
	"time"
)

// Storage é a interface para operações de armazenamento de mídia.
//
// Esta interface define o contrato para diferentes providers de storage
// (GCS, S3, MinIO, etc.) permitindo trocar a implementação sem afetar o código.
type Storage interface {
	// Upload faz upload de um arquivo e retorna a URL pública
	//
	// Parâmetros:
	// - ctx: Contexto para cancelamento/timeout
	// - file: Reader contendo os dados do arquivo
	// - path: Caminho onde o arquivo será armazenado (ex: "media/images/uuid.jpg")
	// - opts: Opções de upload (content-type, tamanho máximo, etc.)
	//
	// Retorna:
	// - URL pública do arquivo
	// - Erro se upload falhar
	Upload(ctx context.Context, file io.Reader, path string, opts UploadOptions) (string, error)

	// GetSignedURL gera uma URL temporária com acesso limitado
	//
	// Útil para:
	// - Compartilhar arquivos privados temporariamente
	// - URLs com expiração automática
	// - Controle de acesso granular
	//
	// Parâmetros:
	// - ctx: Contexto
	// - path: Caminho do arquivo
	// - expiry: Duração até a URL expirar
	//
	// Retorna:
	// - URL temporária (expira após 'expiry')
	// - Erro se geração falhar
	GetSignedURL(ctx context.Context, path string, expiry time.Duration) (string, error)

	// Delete remove um arquivo do storage
	//
	// Parâmetros:
	// - ctx: Contexto
	// - path: Caminho do arquivo a deletar
	//
	// Retorna:
	// - Erro se deleção falhar (ou nil se arquivo não existir)
	Delete(ctx context.Context, path string) error

	// Exists verifica se um arquivo existe no storage
	//
	// Parâmetros:
	// - ctx: Contexto
	// - path: Caminho do arquivo
	//
	// Retorna:
	// - true se arquivo existe
	// - false se não existe
	// - Erro se verificação falhar
	Exists(ctx context.Context, path string) (bool, error)
}

// UploadOptions define opções para upload de arquivo
type UploadOptions struct {
	// ContentType é o MIME type do arquivo (ex: "image/jpeg", "video/mp4")
	ContentType string

	// Public define se o arquivo deve ser público ou privado
	// - true: qualquer um pode acessar via URL
	// - false: requer signed URL
	Public bool

	// MaxSize é o tamanho máximo permitido em bytes
	// Se 0, não há limite
	MaxSize int64

	// Metadata adicional para o arquivo (key-value)
	// Útil para tags, descrições, etc.
	Metadata map[string]string
}

// UploadResult contém informações sobre o upload realizado
type UploadResult struct {
	// URL pública do arquivo
	URL string

	// Path onde o arquivo foi armazenado
	Path string

	// Size do arquivo em bytes
	Size int64

	// ContentType do arquivo
	ContentType string

	// UploadedAt timestamp do upload
	UploadedAt time.Time
}

// StorageError representa erros específicos de storage
type StorageError struct {
	Op      string // Operação que falhou (Upload, Delete, etc.)
	Path    string // Path do arquivo
	Err     error  // Erro original
	Message string // Mensagem customizada
}

func (e *StorageError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return e.Op + " failed for " + e.Path + ": " + e.Err.Error()
}

func (e *StorageError) Unwrap() error {
	return e.Err
}
