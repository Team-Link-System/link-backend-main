package util

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const accessTokenExp = time.Hour * 24
const refreshTokenExp = time.Hour * 24 * 5

var accessTokenSecret = []byte(os.Getenv("ACCESS_TOKEN_SECRET"))
var refreshTokenSecret = []byte(os.Getenv("REFRESH_TOKEN_SECRET"))

// Claims 구조체 - 사용자 정보를 토큰에 담음
type Claims struct {
	Name   string `json:"name"`
	Email  string `json:"email"`
	UserId uint   `json:"userId"`
	jwt.RegisteredClaims
}

func GenerateAccessToken(name string, email string, userId uint) (string, error) {
	return generateToken(name, email, userId, accessTokenExp, accessTokenSecret)
}

func GenerateRefreshToken(name string, email string, userId uint) (string, error) {
	return generateToken(name, email, userId, refreshTokenExp, refreshTokenSecret)
}

func generateToken(name string, email string, userId uint, expiration time.Duration, secret []byte) (string, error) {
	expirationTime := time.Now().Add(expiration) // 토큰 생성 시 유효 기간을 계산

	claims := &Claims{
		Name:   name,
		Email:  email,
		UserId: userId,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime), // JWT 표준 형식으로 변환
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}

// TODO 토큰 검증
func ValidateAccessToken(tokenString string) (*Claims, error) {
	return validateToken(tokenString, accessTokenSecret)
}

func ValidateRefreshToken(tokenString string) (*Claims, error) {
	return validateToken(tokenString, refreshTokenSecret)
}

func validateToken(tokenString string, secret []byte) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return secret, nil
	})

	if err != nil || !token.Valid {
		log.Printf("유효하지 않은 토큰:  %v", err)
		return nil, fmt.Errorf("유효하지 않은 토큰입니다")
	}

	return claims, nil
}
