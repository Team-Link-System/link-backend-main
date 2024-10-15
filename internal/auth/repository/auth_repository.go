package repository

type AuthRepository interface {
	StoreRefreshToken(mergeKey, refreshToken string) error
	DeleteRefreshToken(mergeKey string) error
	GetRefreshToken(mergeKey string) (string, error)
}
