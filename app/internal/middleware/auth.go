package middleware

import (
	"encoding/json"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
)

// JWTAuth returns middleware that validates HS256 Bearer tokens.
// Protected routes return 401 with a WWW-Authenticate challenge on failure.
func JWTAuth(secret []byte) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if len(authHeader) < 8 || authHeader[:7] != "Bearer " {
				writeAuthErr(w, "missing or malformed authorization header")
				return
			}

			tokenStr := authHeader[7:]
			_, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrSignatureInvalid
				}
				return secret, nil
			}, jwt.WithValidMethods([]string{"HS256"}))

			if err != nil {
				writeAuthErr(w, "invalid or expired token")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func writeAuthErr(w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("WWW-Authenticate", `Bearer realm="crypto-asset-api"`)
	w.WriteHeader(http.StatusUnauthorized)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
