package handlers

import (
	"VoteGolang/internals/usecases"
	"encoding/json"
	"net/http"
)

type CandidateHandler struct {
	UseCase *usecases.CandidateUseCase
}

func NewCandidateHandler(uc *usecases.CandidateUseCase) *CandidateHandler {
	return &CandidateHandler{UseCase: uc}
}

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

func (h *CandidateHandler) Vote(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID")
	if userID == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	type voteRequest struct {
		CandidateID   uint   `json:"candidate_id"`
		CandidateType string `json:"type"`
	}

	var req voteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	err := h.UseCase.Vote(req.CandidateID, userID.(string), req.CandidateType)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Write([]byte("Vote successful"))
}
