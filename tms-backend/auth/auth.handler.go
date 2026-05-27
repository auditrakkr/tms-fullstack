package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"net/http"

	"github.com/auditrakkr/tms-fullstack/tms-backend/config"
	"github.com/auditrakkr/tms-fullstack/tms-backend/global"
	"github.com/auditrakkr/tms-fullstack/tms-backend/middleware"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
)

// Session key constants for PKCE state.
const (
	sessionKeyCodeVerifier = "google_pkce_code_verifier"
	sessionKeyOAuthState   = "oauth_state"
)

// loginRequest is the JSON body for POST /auth/login.
type loginRequest struct {
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// refreshRequest is the JSON body for POST /auth/refresh.
type refreshRequest struct {
	RefreshToken string `json:"refreshToken" validate:"required"`
}

// AuthHandler wires HTTP transport to AuthService.
type AuthHandler struct {
	svc *AuthService
}

func NewAuthHandler(svc *AuthService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

// Login godoc
// @Summary      Log in with email and password
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body body loginRequest true "Credentials"
// @Success      200  {object}  LoginResponse
// @Failure      400  {object}  global.AppError
// @Failure      401  {object}  global.AppError
// @Router       /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if !middleware.BindAndValidate(c, &req) {
		return
	}
	resp, err := h.svc.Login(c.Request.Context(), req.Email, req.Password, c.Request.Host)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// Logout godoc
// @Summary      Invalidate the current session's refresh token
// @Tags         auth
// @Security     BearerAuth
// @Success      200  {object}  map[string]string
// @Failure      401  {object}  global.AppError
// @Router       /auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	userID, exists := c.Get(middleware.CtxUserID)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "unauthorized"})
		return
	}
	if err := h.svc.Logout(c.Request.Context(), userID.(uint)); err != nil {
		global.RespondWithError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "logged out successfully"})
}

// Refresh godoc
// @Summary      Exchange a refresh token for a new access token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body body refreshRequest true "Refresh token"
// @Success      200  {object}  LoginResponse
// @Failure      401  {object}  global.AppError
// @Router       /auth/refresh [post]
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req refreshRequest
	if !middleware.BindAndValidate(c, &req) {
		return
	}
	resp, err := h.svc.RefreshToken(c.Request.Context(), req.RefreshToken, c.Request.Host)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// ── Google OAuth2 + PKCE ──────────────────────────────────────────────────────

// GoogleRedirect redirects the user to the Google OAuth2 consent screen.
// A PKCE code_verifier is generated per-request and stored in the session;
// the corresponding code_challenge (S256) is sent to Google.
func (h *AuthHandler) GoogleRedirect(c *gin.Context) {
	if config.GoogleOAuthConfig == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"message": "Google OAuth2 not configured"})
		return
	}

	// Generate PKCE code verifier (43–128 random chars, URL-safe base64).
	verifier, err := generateCodeVerifier()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to generate PKCE verifier"})
		return
	}

	// Derive the S256 challenge from the verifier.
	challenge := s256Challenge(verifier)

	// Persist verifier in session so the callback can retrieve it.
	session := sessions.Default(c)
	session.Set(sessionKeyCodeVerifier, verifier)
	if err := session.Save(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to save session"})
		return
	}

	url := config.GoogleOAuthConfig.AuthCodeURL(
		"state",
		oauth2.AccessTypeOffline,
		oauth2.SetAuthURLParam("code_challenge", challenge),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
	)
	c.Redirect(http.StatusTemporaryRedirect, url)
}

// GoogleCallback handles the redirect from Google after user consent.
// It retrieves the PKCE verifier from the session and passes it to the
// token exchange so the server can prove it initiated the flow.
func (h *AuthHandler) GoogleCallback(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "missing code parameter"})
		return
	}

	// Retrieve the verifier stored by GoogleRedirect.
	session := sessions.Default(c)
	verifierRaw := session.Get(sessionKeyCodeVerifier)
	verifier, _ := verifierRaw.(string)
	// Clear the verifier from session after single use.
	session.Delete(sessionKeyCodeVerifier)
	_ = session.Save()

	resp, err := h.svc.GoogleCallbackWithPKCE(c.Request.Context(), code, verifier, c.Request.Host)
	if err != nil {
		global.RespondWithError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// ── Facebook OAuth2 ───────────────────────────────────────────────────────────

// FacebookRedirect redirects the client to the Facebook OAuth2 consent screen.
func (h *AuthHandler) FacebookRedirect(c *gin.Context) {
	if config.FBOAuthConfig == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"message": "Facebook OAuth2 not configured"})
		return
	}
	url := config.FBOAuthConfig.AuthCodeURL("state")
	c.Redirect(http.StatusTemporaryRedirect, url)
}

// FacebookCallback handles the redirect from Facebook after user consent.
func (h *AuthHandler) FacebookCallback(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "missing code parameter"})
		return
	}
	resp, err := h.svc.FacebookCallback(c.Request.Context(), code, c.Request.Host)
	if err != nil {
		global.RespondWithError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// ── PKCE helpers ──────────────────────────────────────────────────────────────

// generateCodeVerifier creates a 32-byte cryptographically random code_verifier
// encoded as URL-safe base64 (no padding), as specified by RFC 7636.
func generateCodeVerifier() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// s256Challenge computes the S256 code_challenge from a verifier:
// BASE64URL(SHA256(verifier)).
func s256Challenge(verifier string) string {
	h := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(h[:])
}
