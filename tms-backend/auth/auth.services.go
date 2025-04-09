package auth

import (
	"errors"
	"fmt"
	//"net/http"
	"time"

	"github.com/auditrakkr/tms-fullstack/tms-backend/auth/dtos"
	"github.com/auditrakkr/tms-fullstack/tms-backend/config"
	"github.com/auditrakkr/tms-fullstack/tms-backend/models"
	"github.com/auditrakkr/tms-fullstack/tms-backend/users"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userService *users.UserService
	googleConfig dtos.GoogleProfileDto
	facebookConfig dtos.FacebookProfileDto
}

func NewAuthService(userService *users.UserService, jwtConfig JwtConstants, googleConfig dtos.GoogleProfileDto, facebookConfig dtos.FacebookProfileDto) *AuthService {
	return &AuthService{
		userService: userService,
		googleConfig: googleConfig,
		facebookConfig: facebookConfig,
	}
}

func (s *AuthService) ValidateUser(email, password string) (*models.User, error) {
	user, err := s.userService.FindByPrimaryEmailAddress(email)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %v", err)
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))

	if err != nil {
		return nil, errors.New("invalid password")
	}

	// Remove sensitive information
	user.Sanitize()

	return user, nil
}

func (s *AuthService) CreateAccessToken(user *models.User, c *gin.Context) (string, error) {

	jwtConstants := config.AppConfig.JWT
	// Create token payload
	payload := AuthTokenPayload{
		Username: user.PrimaryEmailAddress,
		Sub: struct {
			ID        uint     `json:"id"`
			FirstName string   `json:"first_name"`
			LastName  string   `json:"last_name"`
			Landlord  bool     `json:"landlord"`
			Roles     []string `json:"roles"`

		}{
			ID:        user.ID,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Landlord:  user.Landlord,
		},
	}



	// Add roles if they exist
	for _, role := range user.Roles {
		payload.Sub.Roles = append(payload.Sub.Roles, role.Name)
	}

	// Create token with claims
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"username": payload.Username,
		"sub":      payload.Sub,
		"iss":      c.Request.Host,
		"iat":      time.Now().Unix(),
		"exp":      time.Now().Add(time.Second * time.Duration(jwtConstants.SecretKeyExpiration)).Unix(),
	})

	// Sign the token with private key
	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(jwtConstants.Secret))
	if err != nil {
		return "", err
	}

	signedToken, err := token.SignedString(privateKey)
	if err != nil {
		return "", err
	}

	return signedToken, nil
}

func (s *AuthService) CreateRefreshToken(user *models.User, c *gin.Context) (string, error) {
	jwtConstants := config.AppConfig.JWT

	// Create token payload similar to access token
	payload := AuthTokenPayload{
		Username: user.PrimaryEmailAddress,
		Sub: struct {
			ID        uint     `json:"id"`
			FirstName string   `json:"first_name"`
			LastName  string   `json:"last_name"`
			Landlord  bool     `json:"landlord"`
			Roles     []string `json:"roles"`

		}{
			ID:        user.ID,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Landlord:  user.Landlord,
		},
	}

	// Add roles if they exist
	for _, role := range user.Roles {
		payload.Sub.Roles = append(payload.Sub.Roles, role.Name)
	}

	// Create token with claims
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"username": payload.Username,
		"sub":      payload.Sub,
		"iss":      c.Request.Host,
		"iat":      time.Now().Unix(),
		"exp":      time.Now().Add(time.Second * time.Duration(jwtConstants.RefreshSecretKeyExpiration)).Unix(),
	})

	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(jwtConstants.RefreshSecret))
	if err != nil {
		return "", err
	}

	signedToken, err := token.SignedString(privateKey)
	if err != nil {
		return "", err
	}

	err = s.userService.SetRefreshTokenHash(user.ID, string(signedToken))
	if err != nil {
		return "", err
	}

	return signedToken, nil
}

// Login handles user login and returns access and refresh tokens
func (s *AuthService) Login(user *models.User)