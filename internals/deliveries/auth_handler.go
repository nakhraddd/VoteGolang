package deliveries

import (
	"VoteGolang/internals/usecases"
	"encoding/json"
	"net/http"
)

type AuthHandler struct {
	UseCase *usecases.AuthUseCase
}

func NewAuthHandler(uc *usecases.AuthUseCase) *AuthHandler {
	return &AuthHandler{UseCase: uc}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var creds struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	token, err := h.UseCase.Login(creds.Username, creds.Password)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"token": token})
}
