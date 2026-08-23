package tests

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"r2apimangahost/backend/internal/auth"
	"r2apimangahost/backend/internal/config"
	customMiddleware "r2apimangahost/backend/internal/middleware"

	"github.com/golang-jwt/jwt/v5"
)

func TestGenerateAndValidateToken(t *testing.T) {
	service := auth.NewAuthService(&config.Config{JWTSecret: "0123456789abcdef0123456789abcdef"})
	token, err := service.GenerateToken("user-id", "admin", "admin")
	if err != nil {
		t.Fatal(err)
	}
	claims, err := service.ValidateToken(token)
	if err != nil {
		t.Fatal(err)
	}
	if claims.UserID != "user-id" || claims.Role != "admin" {
		t.Fatalf("unexpected claims: %#v", claims)
	}
}

func TestValidateTokenRejectsWrongIssuer(t *testing.T) {
	secret := "0123456789abcdef0123456789abcdef"
	service := auth.NewAuthService(&config.Config{JWTSecret: secret})
	claims := auth.Claims{
		UserID: "user-id",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "another-service",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.ValidateToken(token); err == nil {
		t.Fatal("expected token with the wrong issuer to be rejected")
	}
}

func TestValidateRejectsWeakSecrets(t *testing.T) {
	cfg := &config.Config{
		AdminUsername:         "admin",
		AdminPassword:         "short",
		JWTSecret:             "short",
		AllowedOrigins:        "https://manga.example.com",
		RateLimitGeneralRPM:   1,
		RateLimitGeneralBurst: 1,
		RateLimitLoginRPM:     1,
		RateLimitLoginBurst:   1,
		RateLimitAdminRPM:     1,
		RateLimitAdminBurst:   1,
	}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected weak credentials to fail validation")
	}
}

func TestAllowedOriginList(t *testing.T) {
	cfg := &config.Config{AllowedOrigins: "https://manga.example.com/, http://localhost:5173"}
	origins := cfg.AllowedOriginList()
	if len(origins) != 2 || origins[0] != "https://manga.example.com" {
		t.Fatalf("unexpected origins: %#v", origins)
	}
}

func TestRateLimiterRejectsBurstAndIgnoresUntrustedForwardedIP(t *testing.T) {
	limiter := customMiddleware.NewIPRateLimiter(60, 2, false)
	handler := limiter.Handler(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	for i, forwardedIP := range []string{"198.51.100.1", "198.51.100.2", "198.51.100.3"} {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "192.0.2.10:4321"
		req.Header.Set("X-Forwarded-For", forwardedIP)
		res := httptest.NewRecorder()
		handler.ServeHTTP(res, req)

		want := http.StatusNoContent
		if i == 2 {
			want = http.StatusTooManyRequests
		}
		if res.Code != want {
			t.Fatalf("request %d: got status %d, want %d", i+1, res.Code, want)
		}
	}
}

func TestRateLimiterHeaders(t *testing.T) {
	limiter := customMiddleware.NewIPRateLimiter(300, 60, false)
	handler := limiter.Handler(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "192.0.2.20:4321"
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if got := res.Header().Get("RateLimit-Limit"); got != "60" {
		t.Fatalf("RateLimit-Limit = %q, want %q", got, "60")
	}
	if got := res.Header().Get("RateLimit-Policy"); got != "300;w=60;burst=60" {
		t.Fatalf("RateLimit-Policy = %q", got)
	}
}

func TestSecurityHeaders(t *testing.T) {
	handler := customMiddleware.SecurityHeaders(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if got := res.Header().Get("X-Content-Type-Options"); got != "nosniff" {
		t.Fatalf("X-Content-Type-Options = %q", got)
	}
	if got := res.Header().Get("Cache-Control"); got != "no-store" {
		t.Fatalf("Cache-Control = %q", got)
	}
}
