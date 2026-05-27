package utils_test

import (
	"strings"
	"testing"

	"github.com/auditrakkr/tms-fullstack/tms-backend/utils"
)

func TestInitCrypto_EmptyKey(t *testing.T) {
	if err := utils.InitCrypto(""); err == nil {
		t.Fatal("expected error for empty key, got nil")
	}
}

func TestEncryptDecrypt_RoundTrip(t *testing.T) {
	if err := utils.InitCrypto("test-secret-key-32-bytes-padding!"); err != nil {
		t.Fatalf("InitCrypto: %v", err)
	}

	tests := []struct {
		name      string
		plaintext string
	}{
		{"empty string", ""},
		{"short string", "hello"},
		{"long string", strings.Repeat("abc", 100)},
		{"unicode", "日本語テスト"},
		{"special chars", "!@#$%^&*()_+-=[]{}|;':\",./<>?"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encrypted, err := utils.Encrypt(tt.plaintext)
			if err != nil {
				t.Fatalf("Encrypt(%q): %v", tt.plaintext, err)
			}
			if encrypted == nil || encrypted.IV == nil || encrypted.Content == nil {
				t.Fatal("Encrypt returned nil result")
			}

			decrypted, err := utils.Decrypt(&struct {
				IV      string `json:"iv"`
				Content string `json:"content"`
			}{
				IV:      *encrypted.IV,
				Content: *encrypted.Content,
			})
			if err != nil {
				t.Fatalf("Decrypt: %v", err)
			}
			if decrypted != tt.plaintext {
				t.Errorf("round-trip: got %q, want %q", decrypted, tt.plaintext)
			}
		})
	}
}

func TestEncrypt_DifferentCiphertextEachCall(t *testing.T) {
	if err := utils.InitCrypto("another-test-key-for-non-determin!"); err != nil {
		t.Fatalf("InitCrypto: %v", err)
	}
	const input = "same plaintext"
	e1, _ := utils.Encrypt(input)
	e2, _ := utils.Encrypt(input)
	if *e1.IV == *e2.IV {
		t.Error("expected different IVs for successive Encrypt calls")
	}
	if *e1.Content == *e2.Content {
		t.Error("expected different ciphertexts for successive Encrypt calls (probabilistic)")
	}
}

func TestGenerateRandomToken(t *testing.T) {
	tok1, err := utils.GenerateRandomToken(32)
	if err != nil {
		t.Fatalf("GenerateRandomToken: %v", err)
	}
	if len(tok1) != 64 { // 32 bytes → 64 hex chars
		t.Errorf("token length: got %d, want 64", len(tok1))
	}

	tok2, _ := utils.GenerateRandomToken(32)
	if tok1 == tok2 {
		t.Error("two tokens should be different (probabilistic)")
	}
}

func TestDecrypt_InvalidIV(t *testing.T) {
	if err := utils.InitCrypto("test-secret-key-32-bytes-padding!"); err != nil {
		t.Fatalf("InitCrypto: %v", err)
	}
	_, err := utils.Decrypt(&struct {
		IV      string `json:"iv"`
		Content string `json:"content"`
	}{IV: "not-hex", Content: "cafebabe"})
	if err == nil {
		t.Fatal("expected error for invalid IV hex, got nil")
	}
}
