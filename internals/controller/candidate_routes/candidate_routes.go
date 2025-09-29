package candidate_routes

import (
	http2 "VoteGolang/internals/controller/http"
	"VoteGolang/internals/domain"
	"log"
	"net/http"
)

func RegisterCandidateRoutes(mux *http.ServeMux, handler *CandidateHandler, tokenManager domain.TokenManager) {
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

	mux.Handle("/candidates", http2.JWTMiddleware(tokenManager)(
		logRequest("/candidates", handler.GetAll),
	))
	mux.Handle("/candidates/", http2.JWTMiddleware(tokenManager)(
		logRequest("/candidates/candidates_repository/all_by_page", handler.GetCandidatesByPage),
	))

	mux.Handle("/vote", http2.JWTMiddleware(tokenManager)(
		logRequest("/candidate/vote", handler.Vote),
	))
}
