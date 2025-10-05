package login_routes

import (
	"VoteGolang/internals/app/logging"
	"VoteGolang/internals/domain"
	"fmt"
	"net/http"
)

func AuthorizationRoutes(mux *http.ServeMux, authHandler *AuthHandler, tokenManager domain.TokenManager, kafkaLogger *logging.KafkaLogger) {

	//login_routes
	mux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		kafkaLogger.Log(fmt.Sprintf("Accessed /login route from %s", r.RemoteAddr))
		authHandler.Login(w, r)
	})

	mux.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		kafkaLogger.Log(fmt.Sprintf("Accessed /register route from %s", r.RemoteAddr))
		authHandler.Register(w, r)
	})

	mux.HandleFunc("/refresh", func(w http.ResponseWriter, r *http.Request) {
		kafkaLogger.Log(fmt.Sprintf("Accessed /refresh route from %s", r.RemoteAddr))
		authHandler.Refresh(w, r)
	})

	mux.HandleFunc("/verify-email", func(w http.ResponseWriter, r *http.Request) {
		kafkaLogger.Log(fmt.Sprintf("Accessed /verify-email route from %s", r.RemoteAddr))
		authHandler.VerifyEmail(w, r)
	})

}
