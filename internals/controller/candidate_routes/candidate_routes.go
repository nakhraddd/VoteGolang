package candidate_routes

import (
	http2 "VoteGolang/internals/controller/http"
	"VoteGolang/internals/domain"
	"VoteGolang/internals/infrastructure/repositories"
	"log"
	"net/http"
)

func RegisterCandidateRoutes(mux *http.ServeMux, handler *CandidateHandler, tokenManager domain.TokenManager, rbacRepo *repositories.RBACRepository) {
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

	mux.Handle("/candidates",
		http2.JWTMiddleware(tokenManager)(
			http2.RBACMiddleware(rbacRepo, "read_candidate")(
				logRequest("/candidates", handler.GetAll),
			),
		),
	)
	mux.Handle("/candidate/page",
		http2.JWTMiddleware(tokenManager)(
			http2.RBACMiddleware(rbacRepo, "read_candidate")(
				logRequest("/candidate/page", handler.GetCandidatesByPage),
			),
		),
	)

	mux.Handle("/candidate/",
		http2.JWTMiddleware(tokenManager)(
			http2.RBACMiddleware(rbacRepo, "read_candidate")(
				logRequest("/candidate/", handler.GetCandidateByID),
			),
		),
	)

	mux.Handle("/vote",
		http2.JWTMiddleware(tokenManager)(
			http2.RBACMiddleware(rbacRepo, "vote")(
				logRequest("/vote", handler.Vote),
			),
		),
	)

	mux.Handle("/candidate/create",
		http2.JWTMiddleware(tokenManager)(
			http2.RBACMiddleware(rbacRepo, "create_candidate")(
				logRequest("/candidate/create", handler.CreateCandidate),
			),
		),
	)

	mux.Handle("/candidate/delete",
		http2.JWTMiddleware(tokenManager)(
			http2.RBACMiddleware(rbacRepo, "delete_candidate")(
				logRequest("/candidate/delete", handler.DeleteCandidate),
			),
		),
	)

}
