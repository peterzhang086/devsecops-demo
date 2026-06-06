package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/peterzhang086/crypto-asset-api/internal/handlers"
)

var authTestSecret = []byte("test-secret-for-unit-tests-32b!")

func TestIssueToken_ValidCredentials(t *testing.T) {
	body := `{"username":"demo","password":"demo"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/token", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handlers.IssueToken(authTestSecret)(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", w.Code)
	}
	var resp map[string]string
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	if resp["token"] == "" {
		t.Fatal("expected non-empty token in response")
	}
}

func TestIssueToken_WrongPassword(t *testing.T) {
	body := `{"username":"demo","password":"wrong"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/token", strings.NewReader(body))
	w := httptest.NewRecorder()

	handlers.IssueToken(authTestSecret)(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d", w.Code)
	}
}

func TestIssueToken_UnknownUser(t *testing.T) {
	body := `{"username":"unknown","password":"demo"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/token", strings.NewReader(body))
	w := httptest.NewRecorder()

	handlers.IssueToken(authTestSecret)(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d", w.Code)
	}
}

func TestIssueToken_BadBody(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/auth/token", strings.NewReader("not-json"))
	w := httptest.NewRecorder()

	handlers.IssueToken(authTestSecret)(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", w.Code)
	}
}

func TestIssueToken_EmptyFields(t *testing.T) {
	body := `{"username":"","password":""}`
	req := httptest.NewRequest(http.MethodPost, "/auth/token", strings.NewReader(body))
	w := httptest.NewRecorder()

	handlers.IssueToken(authTestSecret)(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", w.Code)
	}
}
