package domain

type TokenManager interface {
	Create(*Session, int64) (string, error)
	Check(*Session, string) (bool, error)
}
