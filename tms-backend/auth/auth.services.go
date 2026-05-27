package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/auditrakkr/tms-fullstack/tms-backend/auth/dtos"
	"github.com/auditrakkr/tms-fullstack/tms-backend/config"
	"github.com/auditrakkr/tms-fullstack/tms-backend/models"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"
)

// AuthService handles all authentication flows.
// userService is the UserLookup interface (ports.go) so this package does not
// import the concrete users package — the coupling is inverted.
type AuthService struct {
	userService    UserLookup
	googleConfig   dtos.GoogleProfileDto
	facebookConfig dtos.FacebookProfileDto
}

func NewAuthService(
	userService UserLookup,
	googleConfig dtos.GoogleProfileDto,
	facebookConfig dtos.FacebookProfileDto,
) *AuthService {
	return &AuthService{
		userService:    userService,
		googleConfig:   googleConfig,
		facebookConfig: facebookConfig,
	}
}

// ── Credential validation ──────────────────────────────────

// ValidateUser checks email+password and returns a sanitised user on success.
func (s *AuthService) ValidateUser(ctx context.Context, email, password string) (*models.User, error) {
	user, err := s.userService.FindByPrimaryEmailAddress(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, errors.New("invalid password")
	}
	user.Sanitize()
	return user, nil
}

// ── Token creation ─────────────────────────────────────────

// CreateAccessToken mints a short-lived HS256 JWT.
// issuer is typically c.Request.Host, provided by the handler layer.
func (s *AuthService) CreateAccessToken(user *models.User, issuer string) (string, error) {
	jwtCfg := config.AppConfig.JWT
	return jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": user.PrimaryEmailAddress,
		"sub":      buildSubClaim(user),
		"iss":      issuer,
		"iat":      time.Now().Unix(),
		"exp":      time.Now().Add(time.Second * time.Duration(jwtCfg.SecretKeyExpiration)).Unix(),
	}).SignedString([]byte(jwtCfg.Secret))
}

// CreateRefreshToken mints a long-lived token and persists its bcrypt hash.
func (s *AuthService) CreateRefreshToken(ctx context.Context, user *models.User, issuer string) (string, error) {
	jwtCfg := config.AppConfig.JWT
	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": user.PrimaryEmailAddress,
		"sub":      buildSubClaim(user),
		"iss":      issuer,
		"iat":      time.Now().Unix(),
		"exp":      time.Now().Add(time.Second * time.Duration(jwtCfg.RefreshSecretKeyExpiration)).Unix(),
	}).SignedString([]byte(jwtCfg.RefreshSecret))
	if err != nil {
		return "", fmt.Errorf("failed to sign refresh token: %w", err)
	}
	if err := s.userService.SetRefreshTokenHash(ctx, user.ID, signed); err != nil {
		return "", fmt.Errorf("failed to persist refresh token: %w", err)
	}
	return signed, nil
}

// ── Login / Logout / Refresh ───────────────────────────────

// Login validates credentials and returns a token pair.
func (s *AuthService) Login(ctx context.Context, email, password, issuer string) (*LoginResponse, error) {
	user, err := s.ValidateUser(ctx, email, password)
	if err != nil {
		return nil, err
	}
	return s.issueTokens(ctx, user, issuer)
}

// Logout clears the stored refresh token hash, invalidating the refresh token.
func (s *AuthService) Logout(ctx context.Context, userID uint) error {
	return s.userService.SetRefreshTokenHash(ctx, userID, "")
}

// RefreshToken validates an existing refresh token and issues a new access token.
func (s *AuthService) RefreshToken(ctx context.Context, refreshTokenStr, issuer string) (*LoginResponse, error) {
	jwtCfg := config.AppConfig.JWT
	token, err := jwt.Parse(refreshTokenStr, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(jwtCfg.RefreshSecret), nil
	})
	if err != nil || !token.Valid {
		return nil, errors.New("invalid refresh token")
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}
	username, _ := claims["username"].(string)
	user, err := s.userService.FindByPrimaryEmailAddress(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}
	accessToken, err := s.CreateAccessToken(user, issuer)
	if err != nil {
		return nil, fmt.Errorf("failed to create access token: %w", err)
	}
	return &LoginResponse{AccessToken: accessToken, RefreshToken: refreshTokenStr, User: user}, nil
}

// ── OAuth callbacks ────────────────────────────────────────

// GoogleCallback exchanges an OAuth2 authorisation code without PKCE (kept for
// backward compatibility with existing callers that do not use sessions).
func (s *AuthService) GoogleCallback(ctx context.Context, code, issuer string) (*LoginResponse, error) {
	return s.GoogleCallbackWithPKCE(ctx, code, "", issuer)
}

// GoogleCallbackWithPKCE exchanges an authorisation code, optionally verifying
// the PKCE code_verifier stored in the session by the handler layer.
// If verifier is empty, the exchange proceeds without PKCE.
func (s *AuthService) GoogleCallbackWithPKCE(ctx context.Context, code, verifier, issuer string) (*LoginResponse, error) {
	oauthCfg := config.GoogleOAuthConfig
	if oauthCfg == nil {
		return nil, errors.New("Google OAuth2 is not configured")
	}

	opts := []oauth2.AuthCodeOption{}
	if verifier != "" {
		opts = append(opts, oauth2.SetAuthURLParam("code_verifier", verifier))
	}

	token, err := oauthCfg.Exchange(ctx, code, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange Google OAuth2 code: %w", err)
	}

	client := oauthCfg.Client(ctx, token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch Google user info: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read Google user info: %w", err)
	}

	// Map the raw JSON to GoogleProfileDto.
	// The Google userinfo endpoint returns sub as the unique ID.
	var rawProfile struct {
		Sub           string `json:"sub"`
		GivenName     string `json:"given_name"`
		FamilyName    string `json:"family_name"`
		Email         string `json:"email"`
		EmailVerified bool   `json:"email_verified"`
		Picture       string `json:"picture"`
	}
	if err := json.Unmarshal(body, &rawProfile); err != nil {
		return nil, fmt.Errorf("failed to parse Google profile: %w", err)
	}

	profile := &dtos.GoogleProfileDto{
		GoogleID:      &rawProfile.Sub,
		GivenName:     &rawProfile.GivenName,
		FamilyName:    &rawProfile.FamilyName,
		Email:         &rawProfile.Email,
		EmailVerified: &rawProfile.EmailVerified,
		Picture:       &rawProfile.Picture,
	}

	user, err := s.userService.FindOrCreateGoogleUser(ctx, profile)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve Google user: %w", err)
	}
	return s.issueTokens(ctx, user, issuer)
}

// FacebookCallback exchanges an OAuth2 code, fetches the Facebook user
// profile, and returns a token pair.
func (s *AuthService) FacebookCallback(ctx context.Context, code, issuer string) (*LoginResponse, error) {
	oauthCfg := config.FBOAuthConfig
	if oauthCfg == nil {
		return nil, errors.New("Facebook OAuth2 is not configured")
	}

	token, err := oauthCfg.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange Facebook OAuth2 code: %w", err)
	}

	client := oauthCfg.Client(ctx, token)
	resp, err := client.Get("https://graph.facebook.com/me?fields=id,name,email,first_name,last_name,picture")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch Facebook user info: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read Facebook user info: %w", err)
	}

	var rawProfile struct {
		ID        string `json:"id"`
		Email     string `json:"email"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
	}
	if err := json.Unmarshal(body, &rawProfile); err != nil {
		return nil, fmt.Errorf("failed to parse Facebook profile: %w", err)
	}

	fbID := rawProfile.ID
	email := rawProfile.Email
	profile := &dtos.FacebookProfileDto{
		FacebookID:  &fbID,
		Email:       &email,
		DisplayName: strPtr(rawProfile.FirstName + " " + rawProfile.LastName),
		Name: &struct {
			FamilyName string `json:"familyName"`
			GivenName  string `json:"givenName"`
		}{
			FamilyName: rawProfile.LastName,
			GivenName:  rawProfile.FirstName,
		},
	}

	user, err := s.userService.FindOrCreateFacebookUser(ctx, profile)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve Facebook user: %w", err)
	}
	return s.issueTokens(ctx, user, issuer)
}

// ── Private helpers ────────────────────────────────────────

func (s *AuthService) issueTokens(ctx context.Context, user *models.User, issuer string) (*LoginResponse, error) {
	accessToken, err := s.CreateAccessToken(user, issuer)
	if err != nil {
		return nil, fmt.Errorf("failed to create access token: %w", err)
	}
	refreshToken, err := s.CreateRefreshToken(ctx, user, issuer)
	if err != nil {
		return nil, fmt.Errorf("failed to create refresh token: %w", err)
	}
	return &LoginResponse{AccessToken: accessToken, RefreshToken: refreshToken, User: user}, nil
}

func buildSubClaim(user *models.User) map[string]any {
	var roles []string
	for _, r := range user.Roles {
		roles = append(roles, r.Name)
	}
	return map[string]any{
		"id":         user.ID,
		"first_name": user.FirstName,
		"last_name":  user.LastName,
		"landlord":   user.Landlord,
		"roles":      roles,
	}
}

func strPtr(s string) *string { return &s }
