package login_routes

import (
	"VoteGolang/internals/app/logging"
	"VoteGolang/internals/domain"
	"VoteGolang/internals/usecases/auth_usecase"
	"encoding/json"
	"fmt"
	"net/http"
)

type AuthHandler struct {
	authUseCase  *auth_usecase.AuthUseCase
	tokenManager domain.TokenManager
	kafkaLogger  *logging.KafkaLogger
}

func NewAuthHandler(authUseCase *auth_usecase.AuthUseCase, tokenManager domain.TokenManager, kafkaLogger *logging.KafkaLogger) *AuthHandler {
	return &AuthHandler{
		authUseCase:  authUseCase,
		tokenManager: tokenManager,
		kafkaLogger:  kafkaLogger,
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
	h.kafkaLogger.Log(fmt.Sprintf("Login attempt from %s", r.RemoteAddr))
	var req domain.AuthRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	accessToken, refreshToken, err := h.authUseCase.Login(req.Username, req.Password)
	if err != nil {
		h.kafkaLogger.Log(fmt.Sprintf("Login failed for %s: %v", req.Username, err))
		http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
		return
	}

	h.kafkaLogger.Log(fmt.Sprintf("Login success for %s", req.Username))
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(domain.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	})
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
	h.kafkaLogger.Log(fmt.Sprintf("Register attempt from %s", r.RemoteAddr))
	var req domain.User

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	link, token, err := h.authUseCase.Register(r.Context(), &req)
	if err != nil {
		h.kafkaLogger.Log(fmt.Sprintf("Register failed for %s: %v", req.Username, err))
		http.Error(w, "Failed to register user: "+err.Error(), http.StatusBadRequest)
		return
	}

	h.kafkaLogger.Log(fmt.Sprintf("Register success for %s", req.Username))
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"message":      "User registered successfully",
		"verify_link":  link,
		"verify_token": token,
	})
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	h.kafkaLogger.Log(fmt.Sprintf("Refresh token attempt from %s", r.RemoteAddr))
	var req domain.RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	accessToken, refreshToken, err := h.authUseCase.Refresh(r.Context(), req.RefreshToken)
	if err != nil {
		h.kafkaLogger.Log(fmt.Sprintf("Refresh token failed: %v", err))
		http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
		return
	}

	h.kafkaLogger.Log("Refresh token success")
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(domain.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	})
}

func (h *AuthHandler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	h.kafkaLogger.Log(fmt.Sprintf("Email verification attempt from %s", r.RemoteAddr))
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "missing token", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	err := h.authUseCase.VerifyEmail(ctx, token)
	if err != nil {
		h.kafkaLogger.Log(fmt.Sprintf("Email verification failed: %v", err))
		http.Error(w, fmt.Sprintf("verification failed: %v", err), http.StatusBadRequest)
		return
	}

	h.kafkaLogger.Log("Email verification success")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"message": "Email verified successfully!",
	})
}
