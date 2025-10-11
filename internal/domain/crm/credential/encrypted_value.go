package credential

type EncryptedValue struct {
	ciphertext string
	nonce      string
}

func NewEncryptedValue(ciphertext, nonce string) EncryptedValue {
	return EncryptedValue{
		ciphertext: ciphertext,
		nonce:      nonce,
	}
}

func (e EncryptedValue) Ciphertext() string {
	return e.ciphertext
}

func (e EncryptedValue) Nonce() string {
	return e.nonce
}

func (e EncryptedValue) IsEmpty() bool {
	return e.ciphertext == "" && e.nonce == ""
}

type Encryptor interface {
	Encrypt(plaintext string) (EncryptedValue, error)

	Decrypt(encrypted EncryptedValue) (string, error)
}
