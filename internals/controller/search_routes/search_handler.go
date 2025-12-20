package search_routes

import (
	"encoding/json"
	"net/http"

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

// SearchCandidates handles the request to search for candidates.
func (h *SearchHandler) SearchCandidates(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		http.Error(w, "query parameter 'q' is required", http.StatusBadRequest)
		return
	}

	results, err := h.searcher.SearchCandidates(query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

// SearchPetitions handles the request to search for petitions.
func (h *SearchHandler) SearchPetitions(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		http.Error(w, "query parameter 'q' is required", http.StatusBadRequest)
		return
	}

	results, err := h.searcher.SearchPetitions(query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}
