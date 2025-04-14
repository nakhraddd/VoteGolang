package deliveries

import (
	"VoteGolang/internals/services/auth"
	"VoteGolang/pkg/domain"
	"log"
	"net/http"
)

func LoginRegisterRoutes(mux *http.ServeMux, authHandler *auth.AuthHandler, tokenManager domain.TokenManager) {

	//login
	mux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		logLoginRegister(w, r, "/login")
		authHandler.Login(w, r)
	})

	mux.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		logLoginRegister(w, r, "/register")
		authHandler.Register(w, r)
	})
}

func logLoginRegister(w http.ResponseWriter, r *http.Request, route string) {
	log.Printf("Attempting to access %s route, Method: %s, URL: %s", route, r.Method, r.URL.Path)
}
