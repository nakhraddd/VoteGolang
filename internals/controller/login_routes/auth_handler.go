package login_routes

import (
	"VoteGolang/internals/app/logging"
	"VoteGolang/internals/controller/http/response"
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

// Login godoc
// @Summary      User login
// @Description  Authenticates a user and returns access and refresh tokens.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        credentials  body      domain.AuthRequest  true  "User credentials"
// @Success      200  {object}  domain.TokenResponse
// @Failure      400  {object}  response.JSONResponse  "Invalid request"
// @Failure      401  {object}  response.JSONResponse  "Unauthorized"
// @Router       /login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	h.kafkaLogger.Log("INFO", fmt.Sprintf("Login attempt from %s", r.RemoteAddr))
	var req domain.AuthRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.JSON(w, http.StatusBadRequest, false, "Invalid request", nil)
		return
	}

	accessToken, refreshToken, isAdmin, err := h.authUseCase.Login(req.Username, req.Password)
	if err != nil {
		h.kafkaLogger.Log("WARN", fmt.Sprintf("Login failed for %s: %v", req.Username, err))
		response.JSON(w, http.StatusUnauthorized, false, "Unauthorized: "+err.Error(), nil)
		return
	}

	h.kafkaLogger.Log("INFO", fmt.Sprintf("Login success for %s", req.Username))

	tokenResp := domain.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	if isAdmin {
		tokenResp.IsAdmin = true
	}

	response.JSON(w, http.StatusOK, true, "OK", tokenResp)
}

// Register godoc
// @Summary      Register a new user
// @Description  Creates a new user account and sends an email verification link.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        user  body      domain.User  true  "User registration data"
// @Success      201  {object}  response.JSONResponse  "User registered successfully"
// @Failure      400  {object}  response.JSONResponse  "Invalid request"
// @Router       /register [post]
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	h.kafkaLogger.Log("INFO", fmt.Sprintf("Register attempt from %s", r.RemoteAddr))
	var req domain.User

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.JSON(w, http.StatusBadRequest, false, "Invalid request", nil)
		return
	}

	link, token, err := h.authUseCase.Register(r.Context(), &req)
	if err != nil {
		h.kafkaLogger.Log("WARN", fmt.Sprintf("Register failed for %s: %v", req.Username, err))
		response.JSON(w, http.StatusBadRequest, false, "Failed to register user: "+err.Error(), nil)
		return
	}

	h.kafkaLogger.Log("INFO", fmt.Sprintf("Register success for %s", req.Username))
	response.JSON(w, http.StatusCreated, true, "User registered successfully", map[string]string{
		"verify_link":  link,
		"verify_token": token,
	})
}

// Refresh godoc
// @Summary      Refresh access token
// @Description  Generates a new pair of access and refresh tokens using a valid refresh token.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        token  body      domain.RefreshRequest  true  "Refresh token"
// @Success      200  {object}  domain.TokenResponse
// @Failure      400  {object}  response.JSONResponse  "Invalid request"
// @Failure      401  {object}  response.JSONResponse  "Unauthorized"
// @Router       /refresh [post]
func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	h.kafkaLogger.Log("INFO", fmt.Sprintf("Refresh token attempt from %s", r.RemoteAddr))
	var req domain.RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.JSON(w, http.StatusBadRequest, false, "Invalid request", nil)
		return
	}

	accessToken, refreshToken, err := h.authUseCase.Refresh(r.Context(), req.RefreshToken)
	if err != nil {
		h.kafkaLogger.Log("WARN", fmt.Sprintf("Refresh token failed %v", err))
		response.JSON(w, http.StatusUnauthorized, false, "Unauthorized: "+err.Error(), nil)
		return
	}

	h.kafkaLogger.Log("INFO", fmt.Sprintf("Refresh token success"))
	response.JSON(w, http.StatusOK, true, "OK", domain.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	})
}

// VerifyEmail godoc
// @Summary      Verify email address
// @Description  Verifies the user's email using a token sent in the verification link.
// @Tags         Auth
// @Produce      json
// @Param        token  query     string  true  "Email verification token"
// @Success      200  {object}  response.JSONResponse  "Email verified successfully"
// @Failure      400  {object}  response.JSONResponse  "Invalid or missing token"
// @Router       /verify-email [get]
func (h *AuthHandler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	h.kafkaLogger.Log("INFO", fmt.Sprintf("Email verification attempt from %s", r.RemoteAddr))
	token := r.URL.Query().Get("token")
	if token == "" {
		response.JSON(w, http.StatusBadRequest, false, "Missing token", nil)
		return
	}

	ctx := r.Context()
	err := h.authUseCase.VerifyEmail(ctx, token)
	if err != nil {
		h.kafkaLogger.Log("WARN", fmt.Sprintf("Email verification failed: %v", err))
		response.JSON(w, http.StatusBadRequest, false, "Failed to verify email: "+err.Error(), nil)
		return
	}

	h.kafkaLogger.Log("INFO", "Email verification success")
	response.JSON(w, http.StatusOK, true, "Email verified successfully!", nil)

}
