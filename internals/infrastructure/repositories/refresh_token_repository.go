package repositories

type RefreshTokenRepository interface {
	Save(userID uint, token string, expiry int64) error
	Delete(userID uint, token string) error
	Exists(userID uint, token string) (bool, error)
}
