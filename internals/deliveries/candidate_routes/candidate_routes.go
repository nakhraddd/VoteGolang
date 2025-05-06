package candidate_routes

import (
	"VoteGolang/internals/utils"
	"VoteGolang/pkg/domain"
	"log"
	"net/http"
)

func RegisterCandidateRoutes(mux *http.ServeMux, handler *CandidateHandler, tokenManager domain.TokenManager) {
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

	mux.Handle("/candidates", utils.JWTMiddleware(tokenManager)(
		logRequest("/candidates", handler.GetAll),
	))
	mux.Handle("/candidates/", utils.JWTMiddleware(tokenManager)(
		logRequest("/candidates/candidates_repository/all_by_page", handler.GetCandidatesByPage),
	))

	mux.Handle("/vote", utils.JWTMiddleware(tokenManager)(
		logRequest("/candidate/vote", handler.Vote),
	))
}
