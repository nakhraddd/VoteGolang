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

type IDRequest struct {
	ID uint `json:"id" example:"1"`
}

// Структура для JSON-запроса
type candidateTypeRequest struct {
	Type string `json:"type" example:"manager"`
}

// Parse JSON body
type candidatesByPageRequest struct {
	Type  string `json:"type" example:"manager"`
	Page  int    `json:"page" example:"1"`
	Limit int    `json:"limit" example:"10"`
}

type searchRequest struct {
	Query string `json:"query"`
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
// @Param candidateType body candidateTypeRequest true "Candidate Type"
// @Security BearerAuth
// @Success 200 {array} candidate_data2.Candidate "List of candidates"
// @Failure 400 {object} response.JSONResponse "Bad Request"
// @Failure 500 {object} response.JSONResponse "Internal Server Error"
// @Router /candidates [post]
func (h *CandidateHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	// Проверяем, что метод — POST (или другой, если нужно)
	if r.Method != http.MethodPost {
		response.JSON(w, http.StatusMethodNotAllowed, false, "Method not allowed", nil)
		return
	}

	var req candidateTypeRequest

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

// @Summary Create a candidate
// @Tags Candidates
// @Accept json
// @Produce json
// @Param candidate body candidate_data2.Candidate true "Candidate data"
// @Security BearerAuth
// @Success 201 {object} candidate_data2.Candidate "Candidate created successfully"
// @Failure 400 {object} response.JSONResponse "Bad Request"
// @Failure 401 {object} response.JSONResponse "Unauthorized"
// @Failure 500 {object} response.JSONResponse "Internal Server Error"
// @Router /candidate/create [post]
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
// @Param candidateType body candidatesByPageRequest true "Candidates By Page"
// @Security BearerAuth
// @Success 200 {array} candidate_data2.Candidate "List of candidates"
// @Failure 400 {object} response.JSONResponse "Bad Request"
// @Failure 500 {object} response.JSONResponse "Internal Server Error"
// @Router /candidates/ [post]
func (h *CandidateHandler) GetCandidatesByPage(w http.ResponseWriter, r *http.Request) {
	// Only allow POST for JSON body
	if r.Method != http.MethodPost {
		response.JSON(w, http.StatusMethodNotAllowed, false, "Method not allowed", nil)
		return
	}

	var req candidatesByPageRequest

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

// @Summary Get candidate by ID
// @Tags Candidates
// @Accept json
// @Produce json
// @Param id body IDRequest true "Candidate ID"
// @Security BearerAuth
// @Success 200 {object} candidate_data2.Candidate "Candidate details"
// @Failure 400 {object} response.JSONResponse "Bad Request"
// @Failure 404 {object} response.JSONResponse "Candidate not found"
// @Router /candidate/ [post]
func (h *CandidateHandler) GetCandidateByID(w http.ResponseWriter, r *http.Request) {
	var req IDRequest
	// Decode JSON body
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.JSON(w, http.StatusBadRequest, false, "Invalid JSON body: "+err.Error(), nil)
		return
	}
	// Validate ID
	if req.ID == 0 {
		response.JSON(w, http.StatusBadRequest, false, "Missing or invalid ID", nil)
		return
	}
	// Fetch candidate
	candidate, err := h.UseCase.GetCandidateByID(req.ID)
	if err != nil {
		response.JSON(w, http.StatusNotFound, false, "Candidate not found: "+err.Error(), nil)
		return
	}
	response.JSON(w, http.StatusOK, true, "OK", candidate)
}

// @Summary Vote for a candidate
// @Tags Candidates
// @Accept json
// @Produce json
// @Param candidate body candidate_data2.VoteRequest true "Candidate vote data"
// @Security BearerAuth
// @Success 200 {object} response.JSONResponse "Vote successful"
// @Failure 400 {object} response.JSONResponse "Invalid request or duplicate petition"
// @Failure 401 {object} response.JSONResponse "Unauthorized"
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

// @Summary Delete a candidate
// @Tags Candidates
// @Accept json
// @Produce json
// @Param id body IDRequest true "Candidate ID"
// @Security BearerAuth
// @Success 200 {object} response.JSONResponse "Candidate deleted successfully"
// @Failure 400 {object} response.JSONResponse "Bad Request"
// @Failure 401 {object} response.JSONResponse "Unauthorized"
// @Failure 500 {object} response.JSONResponse "Internal Server Error"
// @Router /candidate/delete [delete]
func (h *CandidateHandler) DeleteCandidate(w http.ResponseWriter, r *http.Request) {
	// Enforce POST or DELETE
	if r.Method != http.MethodDelete && r.Method != http.MethodPost {
		response.JSON(w, http.StatusMethodNotAllowed, false, "Only DELETE or POST allowed", nil)
		return
	}

	// Authenticate user
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
		response.JSON(w, http.StatusUnauthorized, false, "Invalid token: "+err.Error(), nil)
		return
	}

	// Decode JSON body
	var req IDRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.JSON(w, http.StatusBadRequest, false, "Invalid JSON body: "+err.Error(), nil)
		return
	}
	if req.ID == 0 {
		response.JSON(w, http.StatusBadRequest, false, "Missing or invalid candidate ID", nil)
		return
	}

	// Perform delete
	if err := h.UseCase.DeleteCandidate(req.ID); err != nil {
		response.JSON(w, http.StatusInternalServerError, false, "Failed to delete candidate: "+err.Error(), nil)
		return
	}

	response.JSON(w, http.StatusOK, true, "Candidate deleted successfully", nil)
	h.KafkaLogger.Log("INFO", fmt.Sprintf("Candidate deleted: %d", req.ID))
}

// @Summary Search candidates by name
// @Tags Candidates
// @Accept json
// @Produce json
// @Param query body searchRequest true "Search query"
// @Security BearerAuth
// @Success 200 {array} candidate_data2.Candidate "Search results"
// @Failure 400 {object} response.JSONResponse "Query is required"
// @Failure 500 {object} response.JSONResponse "Search failed"
// @Router /candidate/search [post]
func (h *CandidateHandler) SearchCandidates(w http.ResponseWriter, r *http.Request) {
	var req searchRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Query == "" {
		response.JSON(w, http.StatusBadRequest, false, "Query is required", nil)
		return
	}

	results, err := h.UseCase.SearchRepo.Search(req.Query, "Name")
	if err != nil {
		response.JSON(w, http.StatusInternalServerError, false, "Search failed", err.Error())
		return
	}

	response.JSON(w, http.StatusOK, true, "Search results", results)
}
