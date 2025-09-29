package petition_routes

import (
	http2 "VoteGolang/internals/controller/http"
	petition_data2 "VoteGolang/internals/domain"
	"VoteGolang/internals/usecases/petittion_usecase"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
)

type PetitionHandler struct {
	usecase      petittion_usecase.PetitionUseCase
	TokenManager *petition_data2.JwtToken
}

func NewPetitionHandler(usecase petittion_usecase.PetitionUseCase, tokenManager *petition_data2.JwtToken) *PetitionHandler {
	return &PetitionHandler{
		usecase:      usecase,
		TokenManager: tokenManager,
	}
}

// @Summary Create a petition
// @Tags Petition
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param petition body petition.Petition true "Petition Data"
// @Success 200 {string} string "Petition created"
// @Router /petition/create [post]
func (h *PetitionHandler) CreatePetition(w http.ResponseWriter, r *http.Request) {
	token, err := http2.ExtractTokenFromRequest(r)
	if err != nil {
		http.Error(w, "Authorization tokens missing", http.StatusUnauthorized)
		return
	}

	payload := &petition_data2.JwtClaims{}
	_, err = jwt.ParseWithClaims(token, payload, func(t *jwt.Token) (interface{}, error) {
		return h.TokenManager.Secret, nil
	})

	if err != nil {
		http.Error(w, "Invalid tokens", http.StatusUnauthorized)
		return
	}

	userID := payload.UserID
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var p petition_data2.Petition
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	p.UserID = userID

	if err := h.usecase.CreatePetition(&p); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(p)
}

// @Summary Get all petitions
// @Tags Petition
// @Produce json
// @Security BearerAuth
// @Success 200 {array} petition.Petition
// @Router /petition/all [get]
func (h *PetitionHandler) GetAllPetitions(w http.ResponseWriter, r *http.Request) {
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

	petitions, err := h.usecase.GetAllPetitionsPaginated(limit, offset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(petitions)
}

// @Summary Get petitions by page
// @Tags Petition
// @Produce json
// @Security BearerAuth
// @Success 200 {array} petition.Petition
// @Router /petition/all/ [get]
func (h *PetitionHandler) GetPetitionsByPage(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path // example: /petition/all/3
	parts := strings.Split(path, "/")
	if len(parts) < 4 {
		http.Error(w, "Page number required", http.StatusBadRequest)
		return
	}

	pageStr := parts[3]
	page, err := strconv.Atoi(pageStr)
	if err != nil || page <= 0 {
		http.Error(w, "Invalid page number", http.StatusBadRequest)
		return
	}

	const limit = 1
	offset := (page - 1) * limit

	petitions, err := h.usecase.GetAllPetitionsPaginated(limit, offset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(petitions)
}

func (h *PetitionHandler) GetPetitionByID(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	petition, err := h.usecase.GetPetitionByID(uint(id))
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(petition)
}

// @Summary Vote on a petition
// @Tags Petition
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param petitionVote body petition.PetitionVoteRequest true "Petition petition data"
// @Success 200 {string} string "Voted on petition"
// @Failure 400 {string} string "Bad Request"
// @Router /petition/vote [post]
func (h *PetitionHandler) Vote(w http.ResponseWriter, r *http.Request) {
	token, err := http2.ExtractTokenFromRequest(r)
	if err != nil {
		http.Error(w, "Authorization tokens missing", http.StatusUnauthorized)
		return
	}

	payload := &petition_data2.JwtClaims{}
	_, err = jwt.ParseWithClaims(token, payload, func(t *jwt.Token) (interface{}, error) {
		return h.TokenManager.Secret, nil
	})
	if err != nil {
		http.Error(w, "Invalid tokens", http.StatusUnauthorized)
		return
	}

	userID := payload.UserID
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var voteReq petition_data2.PetitionVoteRequest
	if err := json.NewDecoder(r.Body).Decode(&voteReq); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	petition, err := h.usecase.GetPetitionByID(voteReq.PetitionID)
	if err != nil {
		http.Error(w, "Petition not found", http.StatusNotFound)
		return
	}

	if time.Now().After(petition.VotingDeadline) {
		http.Error(w, "Voting period has ended", http.StatusForbidden)
		return
	}

	totalVotes := petition.VotesInFavor + petition.VotesAgainst
	if totalVotes >= petition.Goal {
		http.Error(w, "Vote goal has been reached", http.StatusForbidden)
	}

	err = h.usecase.Vote(userID, voteReq.PetitionID, voteReq.VoteType)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *PetitionHandler) DeletePetition(w http.ResponseWriter, r *http.Request) {
	token, err := http2.ExtractTokenFromRequest(r)
	if err != nil {
		http.Error(w, "Authorization tokens missing", http.StatusUnauthorized)
		return
	}

	payload := &petition_data2.JwtClaims{}
	_, err = jwt.ParseWithClaims(token, payload, func(t *jwt.Token) (interface{}, error) {
		return h.TokenManager.Secret, nil
	})
	if err != nil {
		http.Error(w, "Invalid tokens", http.StatusUnauthorized)
		return
	}

	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	err = h.usecase.DeletePetition(uint(id))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
