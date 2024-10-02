package repository

type AuthRepository interface {
	StoreRefreshToken(refreshToken, email string) error
	GetEmailFromRefreshToken(refreshToken string) (string, error)
	DeleteRefreshToken(refreshToken string) error
}
