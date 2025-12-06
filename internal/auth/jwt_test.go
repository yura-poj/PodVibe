package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestGenerateAndParseTokens(t *testing.T) {
	secret := "test-secret"
	access, refresh, err := GenerateTokens(42, "user", secret, 15*time.Minute, 7*24*time.Hour)
	if err != nil {
		t.Fatalf("GenerateTokens error: %v", err)
	}
	if access == "" || refresh == "" {
		t.Fatalf("expected non-empty tokens")
	}

	claims, err := ParseToken(access, secret)
	if err != nil {
		t.Fatalf("ParseToken error: %v", err)
	}
	if claims.UserID != 42 || claims.Role != "user" {
		t.Fatalf("unexpected claims: %+v", claims)
	}
	if claims.ExpiresAt == nil || time.Until(claims.ExpiresAt.Time) <= 0 {
		t.Fatalf("access token exp not set")
	}
}

func TestMiddlewareSetsContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secret := "another-secret"
	access, _, err := GenerateTokens(7, "admin", secret, time.Hour, 24*time.Hour)
	if err != nil {
		t.Fatalf("GenerateTokens error: %v", err)
	}

	r := gin.New()
	r.Use(Middleware(secret))
	r.GET("/protected", func(c *gin.Context) {
		uid := UserID(c)
		role := c.GetString(ContextUserRole)
		c.JSON(http.StatusOK, gin.H{"uid": uid, "role": role})
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+access)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	expectedBody := `{"role":"admin","uid":7}`
	if rec.Body.String() != expectedBody {
		t.Fatalf("unexpected body: %s", rec.Body.String())
	}
}
