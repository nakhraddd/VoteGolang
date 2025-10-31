package candidate_routes

import (
	"VoteGolang/internals/app/logging"
	http2 "VoteGolang/internals/controller/http"
	"VoteGolang/internals/controller/http/response"
	candidate_data2 "VoteGolang/internals/domain"
	"VoteGolang/internals/usecases/candidate_usecase"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/dgrijalva/jwt-go"
)

type CandidateHandler struct {
	UseCase      *candidate_usecase.CandidateUseCase
	TokenManager *candidate_data2.JwtToken
	KafkaLogger  *logging.KafkaLogger
}

func NewCandidateHandler(useCase *candidate_usecase.CandidateUseCase, tokenManager *candidate_data2.JwtToken, kafkaLogger *logging.KafkaLogger) *CandidateHandler {
	return &CandidateHandler{
		UseCase:      useCase,
		TokenManager: tokenManager,
		KafkaLogger:  kafkaLogger,
	}
}

// @Summary Get candidates by type
// @Tags Candidates
// @Produce json
// @Param type query string true "Candidate Type"
// @Security BearerAuth
// @Success 200 {array} candidate.Candidate "List of candidates"
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /candidates [get]
func (h *CandidateHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	// Проверяем, что метод — POST (или другой, если нужно)
	if r.Method != http.MethodPost {
		response.JSON(w, http.StatusMethodNotAllowed, false, "Method not allowed", nil)
		return
	}

	// Структура для JSON-запроса
	var req struct {
		Type string `json:"type"`
	}

	// Парсим JSON из тела запроса
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.JSON(w, http.StatusBadRequest, false, "Invalid JSON", err.Error())
		return
	}

	// Проверяем обязательное поле
	if req.Type == "" {
		response.JSON(w, http.StatusBadRequest, false, "Type is required", nil)
		return
	}

	// Вызываем use case
	candidates, err := h.UseCase.GetAllByType(req.Type)
	if err != nil {
		response.JSON(w, http.StatusInternalServerError, false, "Failed to get candidates", err.Error())
		return
	}

	response.JSON(w, http.StatusOK, true, "Candidates retrieved successfully", candidates)
}

func (h *CandidateHandler) CreateCandidate(w http.ResponseWriter, r *http.Request) {
	token, err := http2.ExtractTokenFromRequest(r)
	if err != nil {
		response.JSON(w, http.StatusUnauthorized, false, "Unauthorized, missing tokens: "+err.Error(), nil)
		return
	}
	payload := &candidate_data2.JwtClaims{}
	_, err = jwt.ParseWithClaims(token, payload, func(t *jwt.Token) (interface{}, error) {
		return h.TokenManager.Secret, nil
	})

	if err != nil {
		response.JSON(w, http.StatusUnauthorized, false, "Unauthorized, invalid tokens: "+err.Error(), nil)
		return
	}
	var candidate candidate_data2.Candidate
	if err := json.NewDecoder(r.Body).Decode(&candidate); err != nil {
		response.JSON(w, http.StatusBadRequest, false, "Invalid request: "+err.Error(), nil)
		return
	}

	if candidate.Type == "" || candidate.Name == "" {
		response.JSON(w, http.StatusBadRequest, false, "Missing candidate fields", nil)
		return
	}

	err = h.UseCase.CreateCandidate(&candidate)
	if err != nil {
		response.JSON(w, http.StatusInternalServerError, false, "Failed to create candidate: "+err.Error(), nil)
		return
	}

	response.JSON(w, http.StatusCreated, true, "Candidate created successfully", candidate)
}

// @Summary Get candidates by type by page
// @Tags Candidates
// @Produce json
// @Param type query string true "Candidate Type"
// @Security BearerAuth
// @Success 200 {array} candidate.Candidate "List of candidates"
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /candidates/ [get]
func (h *CandidateHandler) GetCandidatesByPage(w http.ResponseWriter, r *http.Request) {
	// Only allow POST for JSON body
	if r.Method != http.MethodPost {
		response.JSON(w, http.StatusMethodNotAllowed, false, "Method not allowed", nil)
		return
	}

	// Parse JSON body
	var req struct {
		Type  string `json:"type"`
		Page  int    `json:"page"`
		Limit int    `json:"limit"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.JSON(w, http.StatusBadRequest, false, "Invalid JSON body", err.Error())
		return
	}

	// Validate input
	if req.Type == "" {
		response.JSON(w, http.StatusBadRequest, false, "Type is required", nil)
		return
	}

	if req.Page <= 0 {
		response.JSON(w, http.StatusBadRequest, false, "Page must be greater than 0", nil)
		return
	}

	if req.Limit <= 0 {
		req.Limit = 10 // default limit
	}

	// Calculate offset
	offset := (req.Page - 1) * req.Limit

	// Fetch data
	candidates, err := h.UseCase.GetAllByTypePaginated(req.Type, req.Limit, offset)
	if err != nil {
		response.JSON(w, http.StatusInternalServerError, false, "Failed to get candidates", err.Error())
		return
	}

	// Return success response
	response.JSON(w, http.StatusOK, true, "Candidates retrieved successfully", candidates)
}

// @Summary Vote for a candidate
// @Tags Candidates
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param candidate body candidate.VoteRequest true "Candidate vote data"
// @Success 200 {string} string "Vote successful"
// @Failure 400 {string} string "Invalid request format or duplicate petition"
// @Failure 401 {string} string "Unauthorized"
// @Router /vote [post]
func (h *CandidateHandler) Vote(w http.ResponseWriter, r *http.Request) {
	h.KafkaLogger.Log("INFO", fmt.Sprintf("Candidate vote attempt from %s", r.RemoteAddr))

	token, err := http2.ExtractTokenFromRequest(r)
	if err != nil {
		response.JSON(w, http.StatusUnauthorized, false, "Unauthorized, authorization tokens missing: "+err.Error(), nil)
		return
	}

	payload := &candidate_data2.JwtClaims{}
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

	log.Printf("User ID from tokens: %d", userID)

	var req candidate_data2.VoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.JSON(w, http.StatusBadRequest, false, "Invalid request format: "+err.Error(), nil)
		return
	}

	err = h.UseCase.Vote(req.CandidateID, userID, candidate_data2.CandidateType(req.CandidateType))
	if err != nil {
		response.JSON(w, http.StatusBadRequest, false, "Failed to vote: "+err.Error(), nil)
		return
	}

	h.KafkaLogger.Log("INFO", fmt.Sprintf("Candidate vote success: user %d voted for candidate %d", userID, req.CandidateID))

	response.JSON(w, http.StatusOK, true, "Vote successfully", nil)
}

func (h *CandidateHandler) DeleteCandidate(w http.ResponseWriter, r *http.Request) {

}
