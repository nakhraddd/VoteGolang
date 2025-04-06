package start

import (
	"VoteGolang/internals/usecases"
	"VoteGolang/pkg/domain"
	"net/http"
)

func RegisterRoutes(authUseCase *usecases.AuthUseCase, tokenManager domain.TokenManager) {
	http.HandleFunc("/login", usecases.NewAuthHandler(authUseCase).Login)

	http.Handle("/protected", JWTMiddleware(tokenManager)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("You have access to the protected route"))
	})))
}
