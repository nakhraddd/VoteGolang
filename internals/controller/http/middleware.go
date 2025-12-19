package middleware

import (
	"VoteGolang/internals/domain"
	"VoteGolang/internals/infrastructure/repositories"
	"context"
	"net/http"
	"strings"
)

type contextKey string

const userIDKey contextKey = "userID"

func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func JWTMiddleware(tokenManager domain.TokenManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenStr, err := ExtractTokenFromRequest(r)
			if err != nil {
				http.Error(w, "Authorization header missing", http.StatusUnauthorized)
				return
			}

			// TokenManager does the validation
			userID, err := tokenManager.VerifyAccessToken(r.Context(), tokenStr)
			if err != nil {
				http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
				return
			}

			// Attaching userID to context
			ctx := context.WithValue(r.Context(), userIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RBACMiddleware(rbacRepo *repositories.RBACRepository, permission string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			val := r.Context().Value(userIDKey)
			if val == nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			userID, ok := val.(uint)
			if !ok {
				http.Error(w, "Invalid user ID type", http.StatusInternalServerError)
				return
			}

			if !rbacRepo.HasAccess(userID, permission) {
				http.Error(w, "Permission denied: no access", http.StatusForbidden)
				return
			}

			ctx := context.WithValue(r.Context(), userIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func ExtractTokenFromRequest(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", http.ErrNoCookie // unauthorized
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", http.ErrNoCookie
	}

	return parts[1], nil
}
