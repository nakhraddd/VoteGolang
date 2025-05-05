package login_routes

import (
	"VoteGolang/internals/data"
	"VoteGolang/internals/usecases/auth_usecase"
	"VoteGolang/pkg/domain"
	"encoding/json"
	"net/http"
)

type AuthHandler struct {
	authUseCase  *auth_usecase.AuthUseCase
	tokenManager domain.TokenManager
}

func NewAuthHandler(authUseCase *auth_usecase.AuthUseCase, tokenManager domain.TokenManager) *AuthHandler {
	return &AuthHandler{
		authUseCase:  authUseCase,
		tokenManager: tokenManager,
	}
}

// @Summary Login and get access token
// @Tags Auth
// @Accept json
// @Produce json
// @Param credentials body data.AuthRequest true "Username and Password"
// @Success 200 {object} map[string]string
// @Failure 401 {string} string "Unauthorized"
// @Router /login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req data.AuthRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	token, err := h.authUseCase.Login(req.Username, req.Password)
	if err != nil {
		http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": token})
}

// @Summary Register a new user
// @Tags Auth
// @Accept json
// @Produce json
// @Param credentials body data.AuthRequest true "Username and Password"
// @Success 200 {string} string "User registered successfully"
// @Failure 400 {string} string "Invalid Request"
// @Router /register [post]
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req data.User

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	err := h.authUseCase.Register(&req)
	if err != nil {
		http.Error(w, "Failed to register user_repository: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "User registered successfully"})
}
