package search_routes

import (
	"net/http"

	"VoteGolang/internals/infrastructure/search"
)

// SetupRoutes sets up the search routes.
func SetupRoutes(mux *http.ServeMux, searcher search.Search) {
	handler := NewSearchHandler(searcher)

	mux.HandleFunc("/search/candidates", handler.SearchCandidates)
	mux.HandleFunc("/search/petitions", handler.SearchPetitions)
}
