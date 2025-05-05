package petition_routes

import (
	"VoteGolang/internals/data/petition_data"
	"VoteGolang/internals/usecases/petittion_usecase"
	"VoteGolang/internals/utils"
	"VoteGolang/pkg/domain"
	"encoding/json"
	"github.com/dgrijalva/jwt-go"
	"net/http"
	"strconv"
	"time"
)

type PetitionHandler struct {
	usecase      petittion_usecase.PetitionUseCase
	TokenManager *domain.JwtToken
}

func NewPetitionHandler(usecase petittion_usecase.PetitionUseCase, tokenManager *domain.JwtToken) *PetitionHandler {
	return &PetitionHandler{
		usecase:      usecase,
		TokenManager: tokenManager,
	}
}

// @Summary Create a petition_data
// @Tags Petition
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param petition_data body petition_data.Petition true "Petition Data"
// @Success 200 {string} string "Petition created"
// @Router /petition/create [post]
func (h *PetitionHandler) CreatePetition(w http.ResponseWriter, r *http.Request) {
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

	var p petition_data.Petition
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
// @Success 200 {array} petition_data.Petition
// @Router /petition/all [get]
func (h *PetitionHandler) GetAllPetitions(w http.ResponseWriter, r *http.Request) {
	petitions, err := h.usecase.GetAllPetitions()
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

// @Summary Vote on a petition_data
// @Tags Petition
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param petitionVote body petition_data.PetitionVoteRequest true "Petition petition_data data"
// @Success 200 {string} string "Voted on petition_data"
// @Failure 400 {string} string "Bad Request"
// @Router /petition/vote [post]
func (h *PetitionHandler) Vote(w http.ResponseWriter, r *http.Request) {
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

	var voteReq petition_data.PetitionVoteRequest
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
		return
	}

	err = h.usecase.Vote(userID, voteReq.PetitionID, voteReq.VoteType)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *PetitionHandler) DeletePetition(w http.ResponseWriter, r *http.Request) {
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
