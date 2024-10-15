package repository

type AuthRepository interface {
	StoreRefreshToken(refreshToken, userId string) error
	DeleteRefreshToken(userId string) error
	GetRefreshToken(userId string) (string, error)
}
