package petition_routes

import (
	http2 "VoteGolang/internals/deliveries/http"
	"VoteGolang/internals/domain"
	"log"
	"net/http"
)

func RegisterPetitionRoutes(mux *http.ServeMux, handler *PetitionHandler, tokenManager domain.TokenManager) {
	logRequest := func(route string, handlerFunc http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			token, err := http2.ExtractTokenFromRequest(r)
			if err != nil {
				log.Printf("Failed to extract tokens: %v", err)
			} else {
				log.Printf("Accessing %s route | Method: %s | URL: %s | Token: %s", route, r.Method, r.URL.Path, token)
			}

			handlerFunc(w, r)
		}
	}

	mux.Handle("/petition/create", http2.JWTMiddleware(tokenManager)(
		logRequest("/petition/petition_repository/create", handler.CreatePetition),
	))

	mux.Handle("/petition/all", http2.JWTMiddleware(tokenManager)(
		logRequest("/petition/petition_repository/all", handler.GetAllPetitions),
	))
	mux.Handle("/petition/all/", http2.JWTMiddleware(tokenManager)(
		logRequest("/petition/petition_repository/all_by_page", handler.GetPetitionsByPage),
	))

	mux.Handle("/petition/get", http2.JWTMiddleware(tokenManager)(
		logRequest("/petition/petition_repository/get", handler.GetPetitionByID),
	))

	mux.Handle("/petition/vote", http2.JWTMiddleware(tokenManager)(
		logRequest("/petition/petition_repository/petition", handler.Vote),
	))

	mux.Handle("/petition/delete", http2.JWTMiddleware(tokenManager)(
		logRequest("/petition/petition_repository/delete", handler.DeletePetition),
	))
}
