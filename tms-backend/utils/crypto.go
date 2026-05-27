package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
)

// Algorithm documents the symmetric cipher in use.
const Algorithm = "AES-CTR"

// secretKey holds the AES-256 key after InitCrypto is called.
var secretKey []byte

// InitCrypto must be called once at startup (from main) before any
// Encrypt/Decrypt calls.  key must be between 1 and 32 characters; shorter
// keys are zero-padded, longer keys are truncated to 32 bytes.
// Returns an error if the key is empty.
func InitCrypto(key string) error {
	if key == "" {
		return errors.New("crypto key must not be empty")
	}
	b := []byte(key)
	if len(b) < 32 {
		padded := make([]byte, 32)
		copy(padded, b)
		b = padded
	} else if len(b) > 32 {
		b = b[:32]
	}
	secretKey = b
	return nil
}

// Encrypt encrypts plaintext using AES-256-CTR.
// Returns a struct with hex-encoded IV and ciphertext.
func Encrypt(plaintext string) (*struct {
	IV      *string `json:"iv"`
	Content *string `json:"content"`
}, error) {
	if secretKey == nil {
		return nil, errors.New("crypto not initialised: call utils.InitCrypto first")
	}
	block, err := aes.NewCipher(secretKey)
	if err != nil {
		return nil, err
	}
	iv := make([]byte, 16)
	if _, err = rand.Read(iv); err != nil {
		return nil, err
	}
	stream := cipher.NewCTR(block, iv)
	ciphertext := make([]byte, len(plaintext))
	stream.XORKeyStream(ciphertext, []byte(plaintext))

	hexIV := hex.EncodeToString(iv)
	hexContent := hex.EncodeToString(ciphertext)
	return &struct {
		IV      *string `json:"iv"`
		Content *string `json:"content"`
	}{IV: &hexIV, Content: &hexContent}, nil
}

// Decrypt reverses Encrypt.
func Decrypt(encryptedData *struct {
	IV      string `json:"iv"`
	Content string `json:"content"`
}) (string, error) {
	if secretKey == nil {
		return "", errors.New("crypto not initialised: call utils.InitCrypto first")
	}
	if encryptedData == nil {
		return "", errors.New("invalid encrypted data")
	}
	iv, err := hex.DecodeString(encryptedData.IV)
	if err != nil {
		return "", fmt.Errorf("failed to decode IV: %w", err)
	}
	ciphertext, err := hex.DecodeString(encryptedData.Content)
	if err != nil {
		return "", fmt.Errorf("failed to decode ciphertext: %w", err)
	}
	block, err := aes.NewCipher(secretKey)
	if err != nil {
		return "", err
	}
	stream := cipher.NewCTR(block, iv)
	plaintext := make([]byte, len(ciphertext))
	stream.XORKeyStream(plaintext, ciphertext)
	return string(plaintext), nil
}

// GenerateRandomToken returns a hex-encoded random token of the given byte length.
func GenerateRandomToken(length int) (string, error) {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// GeneratePassword returns a random 24-character hex password (12 random bytes).
func GeneratePassword() (string, error) {
	return GenerateRandomToken(12)
}

// EncryptSync is a synchronous wrapper around Encrypt for use in entity hooks
// where an error return is inconvenient. Returns empty strings on failure.
func EncryptSync(plaintext string) *struct {
	IV      *string `json:"iv"`
	Content *string `json:"content"`
} {
	result, err := Encrypt(plaintext)
	if err != nil {
		empty := ""
		return &struct {
			IV      *string `json:"iv"`
			Content *string `json:"content"`
		}{IV: &empty, Content: &empty}
	}
	return result
}
