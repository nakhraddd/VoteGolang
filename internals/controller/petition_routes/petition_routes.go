package petition_routes

import (
	http2 "VoteGolang/internals/controller/http"
	"VoteGolang/internals/domain"
	"VoteGolang/internals/infrastructure/repositories"
	"log"
	"net/http"
)

func RegisterPetitionRoutes(mux *http.ServeMux, handler *PetitionHandler, tokenManager domain.TokenManager, rbacRepo *repositories.RBACRepository) {
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

	mux.Handle("/petition/create",
		http2.JWTMiddleware(tokenManager)(
			http2.RBACMiddleware(rbacRepo, "create_petition")(
				logRequest("/petition/petition_repository/create", handler.CreatePetition),
			),
		),
	)

	mux.Handle("/petition/all",
		http2.JWTMiddleware(tokenManager)(
			http2.RBACMiddleware(rbacRepo, "read_petition")(
				logRequest("/petition/petition_repository/all", handler.GetAllPetitions),
			),
		),
	)
	mux.Handle("/petition/all/",
		http2.JWTMiddleware(tokenManager)(
			http2.RBACMiddleware(rbacRepo, "read_petition")(
				logRequest("/petition/petition_repository/all_by_page", handler.GetPetitionsByPage),
			),
		),
	)

	mux.Handle("/petition/",
		http2.JWTMiddleware(tokenManager)(
			http2.RBACMiddleware(rbacRepo, "read_petition")(
				logRequest("/petition/petition_repository", handler.GetPetitionByID),
			),
		),
	)

	mux.Handle("/petition/vote",
		http2.JWTMiddleware(tokenManager)(
			http2.RBACMiddleware(rbacRepo, "vote")(
				logRequest("/petition/petition_repository/petition", handler.Vote),
			),
		),
	)

	mux.Handle("/petition/delete", http2.JWTMiddleware(tokenManager)(
		http2.RBACMiddleware(rbacRepo, "delete_petition")(
			logRequest("/petition/petition_repository/delete", handler.DeletePetition),
		),
	),
	)
}
