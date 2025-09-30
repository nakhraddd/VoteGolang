package login_routes

import (
	"VoteGolang/internals/domain"
	"VoteGolang/internals/usecases/auth_usecase"
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

// Login @Summary Login and get access tokens
// @Tags Auth
// @Accept json
// @Produce json
// @Param credentials body auth.AuthRequest true "Username and Password"
// @Success 200 {object} map[string]string
// @Failure 401 {string} string "Unauthorized"
// @Router /login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req domain.AuthRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	accessToken, refreshToken, err := h.authUseCase.Login(req.Username, req.Password)
	if err != nil {
		http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(domain.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// Register @Summary Register a new user
// @Tags Auth
// @Accept json
// @Produce json
// @Param credentials body auth.AuthRequest true "Username and Password"
// @Success 200 {string} string "User registered successfully"
// @Failure 400 {string} string "Invalid Request"
// @Router /register [post]
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req domain.User

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	err := h.authUseCase.Register(&req)
	if err != nil {
		http.Error(w, "Failed to register user_repository: "+err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(map[string]string{"message": "User registered successfully"})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

//func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
//
//	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
//		http.Error(w, "Invalid request", http.StatusBadRequest)
//		return
//	}
//
//	err := h.authUseCase.Logout(&req)
//	if err != nil {
//		http.Error(w, "Failed to logout"+err.Error(), http.StatusBadRequest)
//	}
//}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req domain.RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	accessToken, refreshToken, err := h.authUseCase.Refresh(r.Context(), req.RefreshToken)
	if err != nil {
		http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(domain.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	})
}
