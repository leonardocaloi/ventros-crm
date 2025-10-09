package crypto

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAESEncryptor(t *testing.T) {
	t.Run("should create encryptor with valid 32-byte key", func(t *testing.T) {
		key := make([]byte, 32)
		encryptor, err := NewAESEncryptor(key)

		require.NoError(t, err)
		assert.NotNil(t, encryptor)
	})

	t.Run("should reject key with invalid length", func(t *testing.T) {
		key := make([]byte, 16) // AES-128, not AES-256
		encryptor, err := NewAESEncryptor(key)

		assert.Error(t, err)
		assert.Nil(t, encryptor)
		assert.Contains(t, err.Error(), "must be exactly 32 bytes")
	})
}

func TestNewAESEncryptorFromBase64(t *testing.T) {
	t.Run("should create encryptor from valid base64 key", func(t *testing.T) {
		key := make([]byte, 32)
		base64Key := base64.StdEncoding.EncodeToString(key)

		encryptor, err := NewAESEncryptorFromBase64(base64Key)

		require.NoError(t, err)
		assert.NotNil(t, encryptor)
	})

	t.Run("should reject invalid base64", func(t *testing.T) {
		encryptor, err := NewAESEncryptorFromBase64("not-valid-base64!!!")

		assert.Error(t, err)
		assert.Nil(t, encryptor)
		assert.Contains(t, err.Error(), "invalid base64 key")
	})

	t.Run("should reject base64 with wrong key length", func(t *testing.T) {
		key := make([]byte, 16)
		base64Key := base64.StdEncoding.EncodeToString(key)

		encryptor, err := NewAESEncryptorFromBase64(base64Key)

		assert.Error(t, err)
		assert.Nil(t, encryptor)
		assert.Contains(t, err.Error(), "must be exactly 32 bytes")
	})
}

func TestEncryptDecrypt(t *testing.T) {
	key, err := GenerateKey()
	require.NoError(t, err)

	encryptor, err := NewAESEncryptor(key)
	require.NoError(t, err)

	t.Run("should encrypt and decrypt successfully", func(t *testing.T) {
		plaintext := "my-secret-api-key-12345"

		encrypted, err := encryptor.Encrypt(plaintext)
		require.NoError(t, err)
		assert.NotEmpty(t, encrypted.Ciphertext())
		assert.NotEmpty(t, encrypted.Nonce())

		decrypted, err := encryptor.Decrypt(encrypted)
		require.NoError(t, err)
		assert.Equal(t, plaintext, decrypted)
	})

	t.Run("should produce different ciphertext for same plaintext", func(t *testing.T) {
		plaintext := "same-secret"

		encrypted1, err := encryptor.Encrypt(plaintext)
		require.NoError(t, err)

		encrypted2, err := encryptor.Encrypt(plaintext)
		require.NoError(t, err)

		// Mesmo plaintext deve produzir ciphertext diferente (nonce aleatório)
		assert.NotEqual(t, encrypted1.Ciphertext(), encrypted2.Ciphertext())
		assert.NotEqual(t, encrypted1.Nonce(), encrypted2.Nonce())

		// Mas ambos devem descriptografar para o mesmo valor
		decrypted1, err := encryptor.Decrypt(encrypted1)
		require.NoError(t, err)

		decrypted2, err := encryptor.Decrypt(encrypted2)
		require.NoError(t, err)

		assert.Equal(t, plaintext, decrypted1)
		assert.Equal(t, plaintext, decrypted2)
	})

	t.Run("should handle unicode text", func(t *testing.T) {
		plaintext := "Texto com acentuação: çãõáéíóú 日本語 🔐"

		encrypted, err := encryptor.Encrypt(plaintext)
		require.NoError(t, err)

		decrypted, err := encryptor.Decrypt(encrypted)
		require.NoError(t, err)
		assert.Equal(t, plaintext, decrypted)
	})

	t.Run("should reject empty plaintext", func(t *testing.T) {
		encrypted, err := encryptor.Encrypt("")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "plaintext cannot be empty")
		assert.True(t, encrypted.IsEmpty())
	})

	t.Run("should fail decryption with wrong key", func(t *testing.T) {
		plaintext := "secret-data"

		encrypted, err := encryptor.Encrypt(plaintext)
		require.NoError(t, err)

		// Cria outro encryptor com chave diferente
		wrongKey, err := GenerateKey()
		require.NoError(t, err)

		wrongEncryptor, err := NewAESEncryptor(wrongKey)
		require.NoError(t, err)

		// Tentativa de descriptografar com chave errada deve falhar
		decrypted, err := wrongEncryptor.Decrypt(encrypted)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "decryption failed")
		assert.Empty(t, decrypted)
	})
}

func TestGenerateKey(t *testing.T) {
	t.Run("should generate 32-byte key", func(t *testing.T) {
		key, err := GenerateKey()

		require.NoError(t, err)
		assert.Len(t, key, 32)
	})

	t.Run("should generate different keys each time", func(t *testing.T) {
		key1, err := GenerateKey()
		require.NoError(t, err)

		key2, err := GenerateKey()
		require.NoError(t, err)

		assert.NotEqual(t, key1, key2)
	})
}

func TestGenerateKeyBase64(t *testing.T) {
	t.Run("should generate valid base64 key", func(t *testing.T) {
		base64Key, err := GenerateKeyBase64()

		require.NoError(t, err)
		assert.NotEmpty(t, base64Key)

		// Deve ser possível decodificar
		key, err := base64.StdEncoding.DecodeString(base64Key)
		require.NoError(t, err)
		assert.Len(t, key, 32)
	})

	t.Run("generated key should work with encryptor", func(t *testing.T) {
		base64Key, err := GenerateKeyBase64()
		require.NoError(t, err)

		encryptor, err := NewAESEncryptorFromBase64(base64Key)
		require.NoError(t, err)

		plaintext := "test-secret"
		encrypted, err := encryptor.Encrypt(plaintext)
		require.NoError(t, err)

		decrypted, err := encryptor.Decrypt(encrypted)
		require.NoError(t, err)
		assert.Equal(t, plaintext, decrypted)
	})
}
