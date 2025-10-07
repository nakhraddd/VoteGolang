package petition_routes

import (
	"VoteGolang/internals/app/logging"
	http2 "VoteGolang/internals/controller/http"
	"VoteGolang/internals/controller/http/response"
	petition_data2 "VoteGolang/internals/domain"
	"VoteGolang/internals/usecases/petittion_usecase"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
)

type PetitionHandler struct {
	usecase      petittion_usecase.PetitionUseCase
	TokenManager *petition_data2.JwtToken
	KafkaLogger  *logging.KafkaLogger
}

func NewPetitionHandler(usecase petittion_usecase.PetitionUseCase, tokenManager *petition_data2.JwtToken, kafkaLogger *logging.KafkaLogger) *PetitionHandler {
	return &PetitionHandler{
		usecase:      usecase,
		TokenManager: tokenManager,
		KafkaLogger:  kafkaLogger,
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
	h.KafkaLogger.Log("INFO", fmt.Sprintf("Petition create attempt from %s", r.RemoteAddr))
	token, err := http2.ExtractTokenFromRequest(r)
	if err != nil {
		response.JSON(w, http.StatusUnauthorized, false, "Unauthorized, missing tokens: "+err.Error(), nil)
		return
	}

	payload := &petition_data2.JwtClaims{}
	_, err = jwt.ParseWithClaims(token, payload, func(t *jwt.Token) (interface{}, error) {
		return h.TokenManager.Secret, nil
	})

	if err != nil {
		response.JSON(w, http.StatusUnauthorized, false, "Unauthorized, invalid tokens: "+err.Error(), nil)
		return
	}

	userID := payload.UserID
	if userID == 0 {
		response.JSON(w, http.StatusUnauthorized, false, "Unauthorized, invalid userID", nil)
		return
	}

	var p petition_data2.Petition
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		response.JSON(w, http.StatusBadRequest, false, "Invalid request body: "+err.Error(), nil)
		return
	}

	p.UserID = userID

	if err := h.usecase.CreatePetition(&p); err != nil {
		response.JSON(w, http.StatusInternalServerError, false, "Failed to create petition: "+err.Error(), nil)
		return
	}

	response.JSON(w, http.StatusCreated, true, "Petition created successfully", p)
	h.KafkaLogger.Log("INFO", fmt.Sprintf("Petition created by user %d: %s", userID, p.Title))
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
		response.JSON(w, http.StatusInternalServerError, false, "Failed to get all paginated petitions: "+err.Error(), petitions)
		return
	}
	response.JSON(w, http.StatusOK, true, "OK", petitions)
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
		response.JSON(w, http.StatusBadRequest, false, "Page number required", nil)
		return
	}

	pageStr := parts[3]
	page, err := strconv.Atoi(pageStr)
	if err != nil || page <= 0 {
		response.JSON(w, http.StatusBadRequest, false, "Invalid page number", nil)
		return
	}

	const limit = 1
	offset := (page - 1) * limit

	petitions, err := h.usecase.GetAllPetitionsPaginated(limit, offset)
	if err != nil {
		response.JSON(w, http.StatusInternalServerError, false, "Failed to get all paginated petitions: "+err.Error(), nil)
		return
	}

	response.JSON(w, http.StatusOK, true, "OK", petitions)
}

func (h *PetitionHandler) GetPetitionByID(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.JSON(w, http.StatusBadRequest, false, "Invalid ID:"+err.Error(), nil)
		return
	}

	petition, err := h.usecase.GetPetitionByID(uint(id))
	if err != nil {
		response.JSON(w, http.StatusNotFound, false, "Petition not found:"+err.Error(), nil)
		return
	}
	response.JSON(w, http.StatusOK, true, "OK", petition)
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
	h.KafkaLogger.Log("INFO", fmt.Sprintf("Petition vote attempt from %s", r.RemoteAddr))
	token, err := http2.ExtractTokenFromRequest(r)
	if err != nil {
		response.JSON(w, http.StatusUnauthorized, false, "Unauthorized, missing tokens: "+err.Error(), nil)
		return
	}

	payload := &petition_data2.JwtClaims{}
	_, err = jwt.ParseWithClaims(token, payload, func(t *jwt.Token) (interface{}, error) {
		return h.TokenManager.Secret, nil
	})
	if err != nil {
		response.JSON(w, http.StatusUnauthorized, false, "Unauthorized, missing tokens: "+err.Error(), nil)
		return
	}

	userID := payload.UserID
	if userID == 0 {
		response.JSON(w, http.StatusUnauthorized, false, "Unauthorized, missing userID", nil)
		return
	}

	var voteReq petition_data2.PetitionVoteRequest
	if err := json.NewDecoder(r.Body).Decode(&voteReq); err != nil {
		response.JSON(w, http.StatusBadRequest, false, "Invalid request body: "+err.Error(), nil)
		return
	}

	petition, err := h.usecase.GetPetitionByID(voteReq.PetitionID)
	if err != nil {
		response.JSON(w, http.StatusNotFound, false, "Petition not found: "+err.Error(), nil)
		return
	}

	if time.Now().After(petition.VotingDeadline) {
		response.JSON(w, http.StatusForbidden, false, "Voting period has ended", nil)
		return
	}

	totalVotes := petition.VotesInFavor + petition.VotesAgainst
	if totalVotes >= petition.Goal {
		response.JSON(w, http.StatusForbidden, false, "Vote goal has been reached", nil)
	}

	err = h.usecase.Vote(userID, voteReq.PetitionID, voteReq.VoteType)
	if err != nil {
		response.JSON(w, http.StatusInternalServerError, false, "Failed to vote: "+err.Error(), nil)
		return
	}

	response.JSON(w, http.StatusOK, true, "OK", petition)
}

func (h *PetitionHandler) DeletePetition(w http.ResponseWriter, r *http.Request) {
	token, err := http2.ExtractTokenFromRequest(r)
	if err != nil {
		response.JSON(w, http.StatusUnauthorized, false, "Unauthorized, missing tokens: "+err.Error(), nil)
		return
	}

	payload := &petition_data2.JwtClaims{}
	_, err = jwt.ParseWithClaims(token, payload, func(t *jwt.Token) (interface{}, error) {
		return h.TokenManager.Secret, nil
	})
	if err != nil {
		response.JSON(w, http.StatusUnauthorized, false, "Unauthorized, missing tokens: "+err.Error(), nil)
		return
	}

	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.JSON(w, http.StatusBadRequest, false, "Invalid ID"+err.Error(), nil)
		return
	}

	err = h.usecase.DeletePetition(uint(id))
	if err != nil {
		response.JSON(w, http.StatusInternalServerError, false, "Failed to delete petition: "+err.Error(), nil)
		return
	}

	response.JSON(w, http.StatusOK, true, "OK", nil)
	h.KafkaLogger.Log("INFO", fmt.Sprintf("Petition deleted: %d", id))
}
