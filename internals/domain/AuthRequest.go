package domain

// AuthRequest represents the body to login.
type AuthRequest struct {
	Username string `json:"username" example:"admin"`
	Password string `json:"password" example:"admin123"`
}
