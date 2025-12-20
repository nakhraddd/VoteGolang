package search_routes

import (
	"encoding/json"
	"net/http"
	"strings"

	"VoteGolang/internals/infrastructure/search"
)

// SearchHandler handles the search requests.
type SearchHandler struct {
	searcher search.Search
}

// NewSearchHandler creates a new SearchHandler instance.
func NewSearchHandler(searcher search.Search) *SearchHandler {
	return &SearchHandler{searcher: searcher}
}

// Search handles the search requests for different types.
func (h *SearchHandler) Search(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")

	// Extract search type from the URL path (e.g., "/search/candidates" -> "candidates")
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 2 {
		http.Error(w, "invalid search path", http.StatusBadRequest)
		return
	}
	searchType := pathParts[1]

	results, err := h.searcher.Search(searchType, query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}
