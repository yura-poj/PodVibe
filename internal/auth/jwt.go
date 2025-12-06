package auth

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// Claims describes JWT payload used across the service.
type Claims struct {
	UserID uint   `json:"sub"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// GenerateTokens creates access and refresh tokens.
func GenerateTokens(userID uint, role, secret string, accessTTL, refreshTTL time.Duration) (string, string, error) {
	now := time.Now()
	accessClaims := Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(accessTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
			Subject:   strconv.FormatUint(uint64(userID), 10),
		},
	}
	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims).SignedString([]byte(secret))
	if err != nil {
		return "", "", err
	}

	refreshClaims := accessClaims
	refreshClaims.RegisteredClaims.ExpiresAt = jwt.NewNumericDate(now.Add(refreshTTL))
	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString([]byte(secret))
	if err != nil {
		return "", "", err
	}
	return accessToken, refreshToken, nil
}

// ParseToken validates JWT and returns claims.
func ParseToken(tokenStr, secret string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}
	return nil, errors.New("invalid token")
}

// Middleware validates Authorization header and stores claims in context.
func Middleware(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "missing token"})
			return
		}
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "invalid auth header"})
			return
		}
		claims, err := ParseToken(parts[1], secret)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "invalid token"})
			return
		}
		c.Set(ContextUserID, claims.UserID)
		c.Set(ContextUserRole, claims.Role)
		c.Next()
	}
}

// RequireRole ensures the user has one of roles; expects Middleware executed before.
func RequireRole(roles ...string) gin.HandlerFunc {
	allowed := map[string]struct{}{}
	for _, r := range roles {
		allowed[r] = struct{}{}
	}
	return func(c *gin.Context) {
		role := c.GetString(ContextUserRole)
		if _, ok := allowed[role]; !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden", "message": "insufficient role"})
			return
		}
		c.Next()
	}
}

// Context keys.
const (
	ContextUserID   = "user_id"
	ContextUserRole = "user_role"
)

// UserID extracts authenticated user id from gin.Context.
func UserID(c *gin.Context) uint {
	if val, ok := c.Get(ContextUserID); ok {
		if id, ok := val.(uint); ok {
			return id
		}
	}
	return 0
}
