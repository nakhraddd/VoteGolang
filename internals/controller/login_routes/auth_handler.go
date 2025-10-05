package login_routes

import (
	"VoteGolang/internals/controller/http/response"
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
		response.JSON(w, http.StatusBadRequest, false, "Invalid request", nil)
		return
	}

	accessToken, refreshToken, err := h.authUseCase.Login(req.Username, req.Password)
	if err != nil {
		response.JSON(w, http.StatusUnauthorized, false, "Unauthorized: "+err.Error(), nil)
		return
	}

	response.JSON(w, http.StatusOK, true, "OK", domain.TokenResponse{
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
	var req domain.User

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.JSON(w, http.StatusBadRequest, false, "Invalid request", nil)
		return
	}

	link, token, err := h.authUseCase.Register(r.Context(), &req)
	if err != nil {
		response.JSON(w, http.StatusBadRequest, false, "Failed to register user: "+err.Error(), nil)
		return
	}

	response.JSON(w, http.StatusCreated, true, "User registered successfully", map[string]string{
		"verify_link":  link,
		"verify_token": token,
	})
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req domain.RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.JSON(w, http.StatusBadRequest, false, "Invalid request", nil)
		return
	}

	accessToken, refreshToken, err := h.authUseCase.Refresh(r.Context(), req.RefreshToken)
	if err != nil {
		response.JSON(w, http.StatusUnauthorized, false, "Unauthorized: "+err.Error(), nil)
		return
	}

	response.JSON(w, http.StatusOK, true, "OK", domain.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	})
}

func (h *AuthHandler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		response.JSON(w, http.StatusBadRequest, false, "Missing token", nil)
		return
	}

	ctx := r.Context()
	err := h.authUseCase.VerifyEmail(ctx, token)
	if err != nil {
		response.JSON(w, http.StatusBadRequest, false, "Failed to verify email: "+err.Error(), nil)
		return
	}

	response.JSON(w, http.StatusOK, true, "Email verified successfully!", nil)

}
