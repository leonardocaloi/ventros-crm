package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"

	"github.com/caloi/ventros-crm/internal/domain/crm/credential"
)

// AESEncryptor implementa criptografia AES-256-GCM
type AESEncryptor struct {
	key []byte // 32 bytes para AES-256
}

// NewAESEncryptor cria um novo encryptor AES-256-GCM
// A chave deve ter 32 bytes (256 bits)
func NewAESEncryptor(key []byte) (*AESEncryptor, error) {
	if len(key) != 32 {
		return nil, errors.New("key must be exactly 32 bytes for AES-256")
	}

	return &AESEncryptor{
		key: key,
	}, nil
}

// NewAESEncryptorFromBase64 cria um encryptor a partir de uma chave base64
func NewAESEncryptorFromBase64(base64Key string) (*AESEncryptor, error) {
	key, err := base64.StdEncoding.DecodeString(base64Key)
	if err != nil {
		return nil, errors.New("invalid base64 key: " + err.Error())
	}

	return NewAESEncryptor(key)
}

// Encrypt criptografa um texto plano usando AES-256-GCM
func (e *AESEncryptor) Encrypt(plaintext string) (credential.EncryptedValue, error) {
	if plaintext == "" {
		return credential.EncryptedValue{}, errors.New("plaintext cannot be empty")
	}

	// Cria o cipher block
	block, err := aes.NewCipher(e.key)
	if err != nil {
		return credential.EncryptedValue{}, err
	}

	// Cria o GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return credential.EncryptedValue{}, err
	}

	// Gera um nonce aleatório (12 bytes é o tamanho padrão para GCM)
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return credential.EncryptedValue{}, err
	}

	// Criptografa
	ciphertext := gcm.Seal(nil, nonce, []byte(plaintext), nil)

	// Converte para base64
	ciphertextB64 := base64.StdEncoding.EncodeToString(ciphertext)
	nonceB64 := base64.StdEncoding.EncodeToString(nonce)

	return credential.NewEncryptedValue(ciphertextB64, nonceB64), nil
}

// Decrypt descriptografa um valor criptografado usando AES-256-GCM
func (e *AESEncryptor) Decrypt(encrypted credential.EncryptedValue) (string, error) {
	if encrypted.IsEmpty() {
		return "", errors.New("encrypted value is empty")
	}

	// Decodifica base64
	ciphertext, err := base64.StdEncoding.DecodeString(encrypted.Ciphertext())
	if err != nil {
		return "", errors.New("invalid ciphertext base64: " + err.Error())
	}

	nonce, err := base64.StdEncoding.DecodeString(encrypted.Nonce())
	if err != nil {
		return "", errors.New("invalid nonce base64: " + err.Error())
	}

	// Cria o cipher block
	block, err := aes.NewCipher(e.key)
	if err != nil {
		return "", err
	}

	// Cria o GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// Verifica o tamanho do nonce
	if len(nonce) != gcm.NonceSize() {
		return "", errors.New("invalid nonce size")
	}

	// Descriptografa
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", errors.New("decryption failed: " + err.Error())
	}

	return string(plaintext), nil
}

// GenerateKey gera uma chave aleatória de 32 bytes para AES-256
func GenerateKey() ([]byte, error) {
	key := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil, err
	}
	return key, nil
}

// GenerateKeyBase64 gera uma chave aleatória e retorna em base64
func GenerateKeyBase64() (string, error) {
	key, err := GenerateKey()
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(key), nil
}
