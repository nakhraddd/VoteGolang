package candidate_routes

import (
	"VoteGolang/internals/data/candidate_data"
	"VoteGolang/internals/usecases/candidate_usecase"
	"VoteGolang/internals/utils"
	"VoteGolang/pkg/domain"
	"encoding/json"
	"github.com/dgrijalva/jwt-go"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type CandidateHandler struct {
	UseCase      *candidate_usecase.CandidateUseCase
	TokenManager *domain.JwtToken
}

func NewCandidateHandler(uc *candidate_usecase.CandidateUseCase, tokenManager *domain.JwtToken) *CandidateHandler {
	return &CandidateHandler{
		UseCase:      uc,
		TokenManager: tokenManager,
	}
}

// @Summary Get candidates by type
// @Tags Candidates
// @Produce json
// @Param type query string true "Candidate Type"
// @Security BearerAuth
// @Success 200 {array} candidate_data.Candidate "List of candidates"
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /candidates [get]
func (h *CandidateHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	typ := r.URL.Query().Get("type")
	if typ == "" {
		http.Error(w, "type is required", http.StatusBadRequest)
		return
	}

	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}
	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	candidates, err := h.UseCase.GetAllByTypePaginated(typ, limit, offset)
	if err != nil {
		http.Error(w, "failed to get candidates", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(candidates)
}

// @Summary Get candidates by type by page
// @Tags Candidates
// @Produce json
// @Param type query string true "Candidate Type"
// @Security BearerAuth
// @Success 200 {array} candidate_data.Candidate "List of candidates"
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /candidates/ [get]
func (h *CandidateHandler) GetCandidatesByPage(w http.ResponseWriter, r *http.Request) {
	// Get 'type' query parameter
	typ := r.URL.Query().Get("type")
	if typ == "" {
		http.Error(w, "type is required", http.StatusBadRequest)
		return
	}

	// Split the URL to get the page number
	path := r.URL.Path // example: /candidates/1
	parts := strings.Split(path, "/")
	if len(parts) < 3 {
		http.Error(w, "Page number required", http.StatusBadRequest)
		return
	}

	pageStr := parts[len(parts)-1] // Get the last part which is the page number
	page, err := strconv.Atoi(pageStr)
	if err != nil || page <= 0 {
		http.Error(w, "Invalid page number", http.StatusBadRequest)
		return
	}

	// Optionally, get 'limit' from query params, else default to 10
	limitStr := r.URL.Query().Get("limit")
	limit := 10 // default value
	if limitStr != "" {
		limit, err = strconv.Atoi(limitStr)
		if err != nil || limit <= 0 {
			http.Error(w, "Invalid limit", http.StatusBadRequest)
			return
		}
	}

	// Calculate offset for pagination
	offset := (page - 1) * limit

	// Get candidates using the paginated use case method
	candidates, err := h.UseCase.GetAllByTypePaginated(typ, limit, offset)
	if err != nil {
		http.Error(w, "Failed to get candidates", http.StatusInternalServerError)
		return
	}

	// Return the candidates
	json.NewEncoder(w).Encode(candidates)
}

// @Summary Vote for a candidate
// @Tags Candidates
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param candidate_data body candidate_data.VoteRequest true "Candidate vote data"
// @Success 200 {string} string "Vote successful"
// @Failure 400 {string} string "Invalid request format or duplicate petition_data"
// @Failure 401 {string} string "Unauthorized"
// @Router /vote [post]
func (h *CandidateHandler) Vote(w http.ResponseWriter, r *http.Request) {
	token, err := utils.ExtractTokenFromRequest(r)
	if err != nil {
		http.Error(w, "Authorization token missing", http.StatusUnauthorized)
		return
	}

	payload := &domain.JwtClaims{}
	_, err = jwt.ParseWithClaims(token, payload, func(t *jwt.Token) (interface{}, error) {
		return h.TokenManager.Secret, nil
	})
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	userID := payload.UserID
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	log.Printf("User ID from token: %s", userID)

	var req candidate_data.VoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	err = h.UseCase.Vote(req.CandidateID, userID, candidate_data.CandidateType(req.CandidateType))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Write([]byte("Vote successful"))
}
