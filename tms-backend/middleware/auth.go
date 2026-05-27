package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/auditrakkr/tms-fullstack/tms-backend/config"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

// Context keys for values set by JWTAuth.
const (
	CtxUserID   = "userID"
	CtxRoles    = "roles"
	CtxLandlord = "landlord"
	CtxEmail    = "email"
)

// JWTAuth validates the HS256 Bearer token in the Authorization header and
// injects the user's identity into the Gin context. Requests without a valid
// token receive 401.
func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		raw := c.GetHeader("Authorization")
		if raw == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "authorization header required"})
			return
		}
		parts := strings.SplitN(raw, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "malformed authorization header"})
			return
		}

		claims, err := parseJWT(parts[1])
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "invalid or expired token"})
			return
		}

		// Inject identity into context for downstream handlers.
		if sub, ok := claims["sub"].(map[string]any); ok {
			if id, ok := sub["id"].(float64); ok {
				c.Set(CtxUserID, uint(id))
			}
			if roles, ok := sub["roles"]; ok {
				c.Set(CtxRoles, roles)
			}
			if landlord, ok := sub["landlord"].(bool); ok {
				c.Set(CtxLandlord, landlord)
			}
		}
		if email, ok := claims["username"].(string); ok {
			c.Set(CtxEmail, email)
		}

		c.Next()
	}
}

func parseJWT(tokenStr string) (jwt.MapClaims, error) {
	secret := []byte(config.AppConfig.JWT.Secret)
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return secret, nil
	})
	if err != nil || !token.Valid {
		return nil, errors.New("invalid token")
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid claims")
	}
	return claims, nil
}
