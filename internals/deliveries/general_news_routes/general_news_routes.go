package general_news_routes

import (
	"VoteGolang/internals/utils"
	"VoteGolang/pkg/domain"
	"log"
	"net/http"
)

func RegisterGeneralNewsRoutes(mux *http.ServeMux, handler *GeneralNewsHandler, tokenManager domain.TokenManager) {
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

	mux.Handle("/general_news", utils.JWTMiddleware(tokenManager)(
		logRequest("/general_news_repository", handler.GetAll),
	))
	mux.Handle("/general_news/get", utils.JWTMiddleware(tokenManager)(
		logRequest("/general_news_repository/get", handler.GetByID),
	))
	mux.Handle("/general_news/create", utils.JWTMiddleware(tokenManager)(
		logRequest("/general_news_repository/create", handler.Create),
	))

	mux.Handle("/general_news/update", utils.JWTMiddleware(tokenManager)(
		logRequest("/general_news_repository/update", handler.Update),
	))

	mux.Handle("/general_news/delete", utils.JWTMiddleware(tokenManager)(
		logRequest("/general_news_repository/delete", handler.Delete),
	))
}
