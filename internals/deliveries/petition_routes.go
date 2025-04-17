package deliveries

import (
	"VoteGolang/internals/handlers"
	"VoteGolang/internals/utils"
	"VoteGolang/pkg/domain"
	"log"
	"net/http"
)

func RegisterPetitionRoutes(mux *http.ServeMux, handler *handlers.PetitionHandler, tokenManager domain.TokenManager) {
	logRequest := func(route string, handlerFunc http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			token, err := utils.ExtractTokenFromRequest(r)
			if err != nil {
				log.Printf("Failed to extract token: %v", err)
			} else {
				log.Printf("Accessing %s route | Method: %s | URL: %s | Token: %s", route, r.Method, r.URL.Path, token)
			}

			handlerFunc(w, r)
		}
	}

	mux.Handle("/vote/petition/create", utils.JWTMiddleware(tokenManager)(
		logRequest("/vote/petition/create", handler.CreatePetition),
	))

	mux.Handle("/vote/petition/all", utils.JWTMiddleware(tokenManager)(
		logRequest("/vote/petition/all", handler.GetAllPetitions),
	))

	mux.Handle("/vote/petition/get", utils.JWTMiddleware(tokenManager)(
		logRequest("/vote/petition/get", handler.GetPetitionByID),
	))

	mux.Handle("/vote/petition/vote", utils.JWTMiddleware(tokenManager)(
		logRequest("/vote/petition/vote", handler.Vote),
	))

	mux.Handle("/vote/petition/delete", utils.JWTMiddleware(tokenManager)(
		logRequest("/vote/petition/delete", handler.DeletePetition),
	))
}
