package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/peterzhang086/crypto-asset-api/internal/middleware"
)

var testSecret = []byte("test-secret-for-unit-tests-32b!")

func okHandler(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) }

func signedToken(secret []byte, expOffset time.Duration) string {
	claims := jwt.MapClaims{
		"sub": "test",
		"exp": time.Now().Add(expOffset).Unix(),
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, _ := t.SignedString(secret)
	return s
}

func TestJWTAuth_ValidToken(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+signedToken(testSecret, time.Hour))
	w := httptest.NewRecorder()

	middleware.JWTAuth(testSecret)(http.HandlerFunc(okHandler)).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", w.Code)
	}
}

func TestJWTAuth_MissingHeader(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	middleware.JWTAuth(testSecret)(http.HandlerFunc(okHandler)).ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d", w.Code)
	}
}

func TestJWTAuth_ExpiredToken(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+signedToken(testSecret, -time.Hour))
	w := httptest.NewRecorder()

	middleware.JWTAuth(testSecret)(http.HandlerFunc(okHandler)).ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d", w.Code)
	}
}

func TestJWTAuth_WrongSecret(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+signedToken([]byte("different-secret!!!!!!!!!!!!!!"), time.Hour))
	w := httptest.NewRecorder()

	middleware.JWTAuth(testSecret)(http.HandlerFunc(okHandler)).ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d", w.Code)
	}
}

func TestJWTAuth_NonBearerScheme(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Basic dXNlcjpwYXNz")
	w := httptest.NewRecorder()

	middleware.JWTAuth(testSecret)(http.HandlerFunc(okHandler)).ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d", w.Code)
	}
}
