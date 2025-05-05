package auth_data

// AuthRequest represents the body to login.
type AuthRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
