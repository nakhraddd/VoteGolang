package start

import (
	"VoteGolang/pkg/domain"
	"net/http"
)

type contextKey string

const sessionKey contextKey = "session"

func JWTMiddleware(tokenManager domain.TokenManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := r.Header.Get("Authorization")
			if token == "" {
				http.Error(w, "Authorization header missing", http.StatusUnauthorized)
				return
			}

			valid, err := tokenManager.Check(r.Context(), token)
			if err != nil || !valid {
				http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
