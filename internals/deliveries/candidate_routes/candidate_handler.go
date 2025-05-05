package candidate_routes

import (
	"VoteGolang/internals/data"
	"VoteGolang/internals/usecases/candidate_usecase"
	"VoteGolang/internals/utils"
	"VoteGolang/pkg/domain"
	"encoding/json"
	"github.com/dgrijalva/jwt-go"
	"log"
	"net/http"
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
// @Success 200 {array} data.Candidate "List of candidates"
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /candidates [get]
func (h *CandidateHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	typ := r.URL.Query().Get("type")
	if typ == "" {
		http.Error(w, "type is required", http.StatusBadRequest)
		return
	}

	candidates, err := h.UseCase.GetAllByType(typ)
	if err != nil {
		http.Error(w, "failed to get candidates", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(candidates)
}

// @Summary Vote for a candidate
// @Tags Vote
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param vote body data.VoteRequest true "Candidate vote data"
// @Success 200 {string} string "Vote successful"
// @Failure 400 {string} string "Invalid request format or duplicate vote"
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

	var req data.VoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	err = h.UseCase.Vote(req.CandidateID, userID, req.CandidateType)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Write([]byte("Vote successful"))
}
