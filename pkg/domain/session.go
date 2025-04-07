package domain

type Session struct {
	ID     string
	UserID string
	Expiry int64
}
