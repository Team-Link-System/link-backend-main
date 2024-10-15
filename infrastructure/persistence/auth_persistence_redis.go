package persistence

import (
	"context"
	"log"
	"time"

	"github.com/go-redis/redis/v8"

	"link/internal/auth/repository"
)

type authPersistenceRedis struct {
	redisClient *redis.Client
}

func NewAuthPersistenceRedis(redisClient *redis.Client) repository.AuthRepository {
	return &authPersistenceRedis{redisClient: redisClient}
}

// refreshToken 저장
func (r *authPersistenceRedis) StoreRefreshToken(refreshToken, userId string) error {
	ctx := context.Background()

	// Redis에 Refresh Token 저장
	err := r.redisClient.Set(ctx, userId, refreshToken, time.Hour*24*5).Err() // 5일 유효
	if err != nil {
		log.Printf("Refresh Token Redis 저장 오류: %v", err)
		return err
	}

	return nil
}

// userId로 refreshToken 가져오기
func (r *authPersistenceRedis) GetRefreshToken(userId string) (string, error) {
	ctx := context.Background()

	refreshToken, err := r.redisClient.Get(ctx, userId).Result()
	if err != nil {
		return "", err
	}

	return refreshToken, nil
}

// refreshToken 삭제 (로그아웃)
func (r *authPersistenceRedis) DeleteRefreshToken(userId string) error {
	ctx := context.Background()

	// Redis에서 Refresh Token 삭제
	err := r.redisClient.Del(ctx, userId).Err()
	if err != nil {
		log.Printf("Redis Refresh Token 삭제 오류: %v", err)
		return err
	}

	return nil
}
