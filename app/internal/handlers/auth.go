package handlers

import (
	"crypto/subtle"
	"encoding/json"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// demoCredentials — hardcoded for demo only.
// Production: validate against an identity provider (e.g. AWS Cognito).
var demoCredentials = map[string]string{
	"demo": "demo",
}

type tokenRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// IssueToken returns a handler that validates credentials and issues a signed HS256 JWT.
func IssueToken(secret []byte) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req tokenRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
			return
		}
		if req.Username == "" || req.Password == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "username and password required"})
			return
		}

		expected, ok := demoCredentials[req.Username]
		// Constant-time comparison prevents timing-based username enumeration.
		if !ok || subtle.ConstantTimeCompare([]byte(expected), []byte(req.Password)) != 1 {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid credentials"})
			return
		}

		claims := jwt.MapClaims{
			"sub": req.Username,
			"iat": time.Now().Unix(),
			"exp": time.Now().Add(time.Hour).Unix(),
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		signed, err := token.SignedString(secret)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to issue token"})
			return
		}

		writeJSON(w, http.StatusOK, map[string]string{"token": signed})
	}
}
