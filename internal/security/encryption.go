package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/scrypt"
	"encoding/base64"
	"fmt"
	"io"
)

type Vault struct {
	password string
}

func NewVault(password string) *Vault {
	return &Vault{password: password}
}

func (v *Vault) Encrypt(data []byte) (string, error) {
	// Derive key from password
	salt := make([]byte, 32)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	key, err := scrypt.Key([]byte(v.password), salt, 32768, 8, 1, 32)
	if err != nil {
		return "", err
	}

	// Create cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// Generate nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	// Encrypt data
	ciphertext := gcm.Seal(nil, nonce, data, nil)

	// Combine salt + nonce + ciphertext
	result := append(salt, nonce...)
	result = append(result, ciphertext...)

	return base64.StdEncoding.EncodeToString(result), nil
}

func (v *Vault) Decrypt(encryptedData string) ([]byte, error) {
	data, err := base64.StdEncoding.DecodeString(encryptedData)
	if err != nil {
		return nil, err
	}

	if len(data) < 32+12 { // salt + nonce minimum
		return nil, fmt.Errorf("invalid encrypted data")
	}

	// Extract components
	salt := data[:32]
	nonce := data[32:44]
	ciphertext := data[44:]

	// Derive key
	key, err := scrypt.Key([]byte(v.password), salt, 32768, 8, 1, 32)
	if err != nil {
		return nil, err
	}

	// Create cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Decrypt
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

func GenerateSecurePassword() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}
