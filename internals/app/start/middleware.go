package start

import (
	"VoteGolang/pkg/domain"
	"fmt"
	"net/http"
	"strings"
)

func JWTMiddleware(tokenManager domain.TokenManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenString := r.Header.Get("Authorization")
		if tokenString == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		tokenString = strings.TrimPrefix(tokenString, "Bearer ")

		isValid, err := tokenManager.Check(&domain.Session{}, tokenString)
		if err != nil || !isValid {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		fmt.Println("JWT Token is valid")
		http.DefaultServeMux.ServeHTTP(w, r)
	}
}
