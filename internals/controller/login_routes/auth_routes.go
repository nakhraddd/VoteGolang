package login_routes

import (
	"VoteGolang/internals/domain"
	"log"
	"net/http"
)

func AuthorizationRoutes(mux *http.ServeMux, authHandler *AuthHandler, tokenManager domain.TokenManager) {

	//login_routes
	mux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		logAuth(w, r, "/login_routes")
		authHandler.Login(w, r)
	})

	mux.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		logAuth(w, r, "/register")
		authHandler.Register(w, r)
	})

	mux.HandleFunc("/refresh", func(w http.ResponseWriter, r *http.Request) {
		logAuth(w, r, "/refresh")
		authHandler.Refresh(w, r)
	})

	mux.HandleFunc("/verify-email", func(w http.ResponseWriter, r *http.Request) {
		logAuth(w, r, "/verify-email")
		authHandler.VerifyEmail(w, r)
	})

}

func logAuth(w http.ResponseWriter, r *http.Request, route string) {
	log.Printf("Attempting to access %s route, Method: %s, URL: %s", route, r.Method, r.URL.Path)
}
