package auth_test

import (
	"context"
	"errors"
	"testing"

	authdtos "github.com/auditrakkr/tms-fullstack/tms-backend/auth/dtos"
	"github.com/auditrakkr/tms-fullstack/tms-backend/config"
	"github.com/auditrakkr/tms-fullstack/tms-backend/models"

	"github.com/auditrakkr/tms-fullstack/tms-backend/auth"
	"golang.org/x/crypto/bcrypt"
)

// ── Stub ─────────────────────────────────────────────────────────────────────

type stubUserLookup struct {
	user    *models.User
	findErr error
	setErr  error
}

func (s *stubUserLookup) FindByPrimaryEmailAddress(_ context.Context, _ string) (*models.User, error) {
	return s.user, s.findErr
}
func (s *stubUserLookup) SetRefreshTokenHash(_ context.Context, _ uint, _ string) error {
	return s.setErr
}
func (s *stubUserLookup) FindByGoogleID(_ context.Context, _ string) (*models.User, error) {
	return s.user, s.findErr
}
func (s *stubUserLookup) FindByFacebookID(_ context.Context, _ string) (*models.User, error) {
	return s.user, s.findErr
}
func (s *stubUserLookup) FindOrCreateGoogleUser(_ context.Context, _ *authdtos.GoogleProfileDto) (*models.User, error) {
	return s.user, s.findErr
}
func (s *stubUserLookup) FindOrCreateFacebookUser(_ context.Context, _ *authdtos.FacebookProfileDto) (*models.User, error) {
	return s.user, s.findErr
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func mustHashPassword(t *testing.T, pw string) string {
	t.Helper()
	h, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("bcrypt: %v", err)
	}
	return string(h)
}

func newTestConfig() {
	config.AppConfig = &config.Config{}
	config.AppConfig.JWT.Secret = "test-secret-32-bytes-padding!!!!"
	config.AppConfig.JWT.SecretKeyExpiration = 3600
	config.AppConfig.JWT.RefreshSecret = "test-refresh-secret-32-bytes!!!"
	config.AppConfig.JWT.RefreshSecretKeyExpiration = 86400
}

// ── Tests ─────────────────────────────────────────────────────────────────────

func TestValidateUser_Success(t *testing.T) {
	newTestConfig()
	const password = "correcthorse"
	user := &models.User{
		FirstName:           "Ada",
		LastName:            "Lovelace",
		PrimaryEmailAddress: "ada@example.com",
		PasswordHash:        mustHashPassword(t, password),
	}
	user.ID = 1

	svc := auth.NewAuthService(&stubUserLookup{user: user}, authdtos.GoogleProfileDto{}, authdtos.FacebookProfileDto{})
	got, err := svc.ValidateUser(context.Background(), "ada@example.com", password)
	if err != nil {
		t.Fatalf("ValidateUser: %v", err)
	}
	if got.PrimaryEmailAddress != user.PrimaryEmailAddress {
		t.Errorf("email: got %q, want %q", got.PrimaryEmailAddress, user.PrimaryEmailAddress)
	}
	// Sanitize should have cleared the hash.
	if got.PasswordHash != "" {
		t.Errorf("PasswordHash should be empty after Sanitize, got %q", got.PasswordHash)
	}
}

func TestValidateUser_WrongPassword(t *testing.T) {
	newTestConfig()
	user := &models.User{
		PrimaryEmailAddress: "x@example.com",
		PasswordHash:        mustHashPassword(t, "correct"),
	}
	user.ID = 1

	svc := auth.NewAuthService(&stubUserLookup{user: user}, authdtos.GoogleProfileDto{}, authdtos.FacebookProfileDto{})
	_, err := svc.ValidateUser(context.Background(), "x@example.com", "wrong")
	if err == nil {
		t.Fatal("expected error for wrong password, got nil")
	}
}

func TestValidateUser_UserNotFound(t *testing.T) {
	newTestConfig()
	svc := auth.NewAuthService(
		&stubUserLookup{findErr: errors.New("not found")},
		authdtos.GoogleProfileDto{}, authdtos.FacebookProfileDto{},
	)
	_, err := svc.ValidateUser(context.Background(), "missing@example.com", "pw")
	if err == nil {
		t.Fatal("expected error when user not found, got nil")
	}
}

func TestCreateAccessToken_RoundTrip(t *testing.T) {
	newTestConfig()
	user := &models.User{FirstName: "Ada", LastName: "Lovelace", PrimaryEmailAddress: "ada@example.com"}
	user.ID = 42

	svc := auth.NewAuthService(&stubUserLookup{}, authdtos.GoogleProfileDto{}, authdtos.FacebookProfileDto{})
	token, err := svc.CreateAccessToken(user, "localhost")
	if err != nil {
		t.Fatalf("CreateAccessToken: %v", err)
	}
	if token == "" {
		t.Fatal("expected non-empty token")
	}
}

func TestLogin_Success(t *testing.T) {
	newTestConfig()
	const password = "s3cr3t"
	user := &models.User{
		FirstName:           "Turing",
		LastName:            "Alan",
		PrimaryEmailAddress: "alan@example.com",
		PasswordHash:        mustHashPassword(t, password),
	}
	user.ID = 2

	svc := auth.NewAuthService(&stubUserLookup{user: user}, authdtos.GoogleProfileDto{}, authdtos.FacebookProfileDto{})
	resp, err := svc.Login(context.Background(), "alan@example.com", password, "localhost")
	if err != nil {
		t.Fatalf("Login: %v", err)
	}
	if resp.AccessToken == "" {
		t.Error("expected non-empty AccessToken")
	}
	if resp.RefreshToken == "" {
		t.Error("expected non-empty RefreshToken")
	}
}

func TestLogout_ClearsToken(t *testing.T) {
	newTestConfig()
	stub := &stubUserLookup{}
	svc := auth.NewAuthService(stub, authdtos.GoogleProfileDto{}, authdtos.FacebookProfileDto{})
	if err := svc.Logout(context.Background(), 1); err != nil {
		t.Fatalf("Logout: %v", err)
	}
}
