package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
)

var (
	// Algorithm is aes256-ctr
	Algorithm = "AES-CTR"

	//SecretKey from environment variable
	SecretKey []byte
)

func init() {
	// Get the secret key from the environment variable
	key := os.Getenv("SECRET_KEY_FOR_CRYPTO_ENCRYPTION")
	if key == "" {
		// Use a default key for development (not recommended for production)
		key = "default_encryption_key_32_bytes_123"
	}

	// Ensure key is exactly 32 bytes (for AES-256)
	if len(key) < 32 {
		key = key + string(make([]byte, 32-len(key)))
	} else if len(key) > 32 {
		key = key[:32]
	}

	SecretKey = []byte(key)
}

// Encrypt encrypts plaintext using AES-CTR
// Returns a struct with hex-encoded iv and content
func Encrypt(plaintext string) (*struct{
	IV      *string `json:"iv"`
	Content *string `json:"content"`
}, error) {
	// Create a new cipher block from the key
	block, err := aes.NewCipher(SecretKey)
	if err != nil {
		return nil, err
	}

	// Create a 16-byte initialization vector
	iv := make([]byte, 16)
	if _, err = rand.Read(iv); err != nil {
		return nil, err
	}

	// Create a stream cipher using CTR mode
	stream := cipher.NewCTR(block, iv)

	// Create buffer for ciphertext
	ciphertext := make([]byte, len(plaintext))

	// Encrypt the data
	stream.XORKeyStream(ciphertext, []byte(plaintext))

	// Convert to hex encoding (matching Node.js implementation)
	hexIV := hex.EncodeToString(iv)
	hexCiphertext := hex.EncodeToString(ciphertext)

	return &struct{
		IV      *string `json:"iv"`
		Content *string `json:"content"`
	}{
		IV:      &hexIV,
		Content: &hexCiphertext,
	}, nil
}

// Decrypt decrypts ciphertext using AES-CTR
func Decrypt(encryptedData *struct{
	IV      string `json:"iv"`
	Content string `json:"content"`
}) (string, error) {
	if encryptedData == nil {
		return "", errors.New("invalid encrypted data")
	}

	// Decode from hex
	iv, err := hex.DecodeString(encryptedData.IV)
	if err != nil {
		return "", fmt.Errorf("failed to decode IV: %w", err)
	}

	ciphertext, err := hex.DecodeString(encryptedData.Content)
	if err != nil {
		return "", fmt.Errorf("failed to decode ciphertext: %w", err)
	}

	// Create a new cipher block from the key
	block, err := aes.NewCipher(SecretKey)
	if err != nil {
		return "", err
	}

	// Create a stream cipher using CTR mode
	stream := cipher.NewCTR(block, iv)

	// Create buffer for plaintext
	plaintext := make([]byte, len(ciphertext))

	// Decrypt the data
	stream.XORKeyStream(plaintext, ciphertext)

	return string(plaintext), nil
}

// GeneratePassword generates a random password
func GeneratePassword() (string, error) {
	// Generate 12 random bytes (24 hex chars)
	randomBytes := make([]byte, 12)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}

	// Convert to hex
	return hex.EncodeToString(randomBytes), nil
}

// EncryptSync is a synchronous version of Encrypt
// This is meant to match the NestJS implementation for use in entities
func EncryptSync(plaintext string) *struct{
	IV     *string `json:"iv"`
	Content *string `json:"content"`
} {
	encrypted, err := Encrypt(plaintext)
	if err != nil {
		// In sync version, we return empty result on error
		empty := ""
		return &struct{
			IV      *string `json:"iv"`
			Content *string `json:"content"`
		}{
			IV:      &empty,
			Content: &empty,
		}
	}
	return encrypted
}