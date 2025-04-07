package start

import (
	"VoteGolang/internals/usecases"
	"VoteGolang/pkg/domain"
	"net/http"
)

func RegisterRoutes(mux *http.ServeMux, authUseCase *usecases.AuthUseCase, tokenManager domain.TokenManager) {
	authHandler := usecases.NewAuthHandler(authUseCase, tokenManager)

	// Register routes with the passed mux
	mux.HandleFunc("/login", authHandler.Login)
	mux.HandleFunc("/register", authHandler.Register)

	// Protected route with middleware
	mux.Handle("/protected", JWTMiddleware(tokenManager)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("You have access to the protected route"))
	})))
}
