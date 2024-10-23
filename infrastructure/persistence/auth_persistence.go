package persistence

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis/v8"

	"link/internal/auth/repository"
)

type authPersistence struct {
	redisClient *redis.Client
}

func NewAuthPersistence(redisClient *redis.Client) repository.AuthRepository {
	return &authPersistence{redisClient: redisClient}
}

// refreshToken 저장
func (r *authPersistence) StoreRefreshToken(mergeKey, refreshToken string) error {
	ctx := context.Background()

	key := fmt.Sprintf("session:%s", mergeKey)

	// Redis에 Refresh Token 저장
	err := r.redisClient.Set(ctx, key, refreshToken, time.Hour*24*5).Err() // 5일 유효
	if err != nil {
		log.Printf("Refresh Token Redis 저장 오류: %v", err)
		return err
	}

	return nil
}

// userId로 refreshToken 가져오기
func (r *authPersistence) GetRefreshToken(mergeKey string) (string, error) {
	ctx := context.Background()

	key := fmt.Sprintf("session:%s", mergeKey)

	refreshToken, err := r.redisClient.Get(ctx, key).Result()
	if err != nil {
		return "", err
	}

	return refreshToken, nil
}

// refreshToken 삭제 (로그아웃)
func (r *authPersistence) DeleteRefreshToken(mergeKey string) error {
	ctx := context.Background()

	key := fmt.Sprintf("session:%s", mergeKey)

	// Redis에서 Refresh Token 삭제
	err := r.redisClient.Del(ctx, key).Err()
	if err != nil {
		log.Printf("Redis Refresh Token 삭제 오류: %v", err)
		return err
	}

	return nil
}
