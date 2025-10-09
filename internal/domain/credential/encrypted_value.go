package credential

// EncryptedValue representa um valor criptografado com AES-256-GCM
type EncryptedValue struct {
	ciphertext string // Base64 encoded
	nonce      string // Base64 encoded (for AES-GCM)
}

// NewEncryptedValue cria um novo valor criptografado
func NewEncryptedValue(ciphertext, nonce string) EncryptedValue {
	return EncryptedValue{
		ciphertext: ciphertext,
		nonce:      nonce,
	}
}

// Ciphertext retorna o texto cifrado
func (e EncryptedValue) Ciphertext() string {
	return e.ciphertext
}

// Nonce retorna o nonce usado na criptografia
func (e EncryptedValue) Nonce() string {
	return e.nonce
}

// IsEmpty verifica se o valor está vazio
func (e EncryptedValue) IsEmpty() bool {
	return e.ciphertext == "" && e.nonce == ""
}

// Encryptor define a interface para criptografia
type Encryptor interface {
	// Encrypt criptografa um texto plano
	Encrypt(plaintext string) (EncryptedValue, error)

	// Decrypt descriptografa um valor criptografado
	Decrypt(encrypted EncryptedValue) (string, error)
}
