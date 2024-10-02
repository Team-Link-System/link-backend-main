package interceptor

import (
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"link/internal/auth/usecase"
	"link/pkg/util"
)

type TokenInterceptor struct {
	authUsecase usecase.AuthUsecase
}

func NewTokenInterceptor(authUsecase usecase.AuthUsecase) *TokenInterceptor {
	return &TokenInterceptor{authUsecase: authUsecase}
}

// Access Token 검증 인터셉터
func (i *TokenInterceptor) AccessTokenInterceptor() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			// "Bearer " 접두사 제거
			tokenString := strings.TrimPrefix(authHeader, "Bearer ")

			// Access Token 검증
			claims, err := util.ValidateAccessToken(tokenString)
			if err == nil {
				// Access Token이 유효한 경우 email과 name을 Context에 설정
				c.Set("email", claims.Email)
				c.Set("userId", claims.UserId)
				c.Next() // Access Token이 유효하면 다음 핸들러로 진행
				return
			}
		}

		// Access Token이 유효하지 않으면 Refresh Token 검증으로 넘김
		c.Next() // 다음 단계(Refresh Token 검증)로 넘어감
	}
}

// Refresh Token 검증 및 Access Token 재발급 인터셉터
// Refresh Token 검증 및 Access Token 재발급 인터셉터
func (i *TokenInterceptor) RefreshTokenInterceptor() gin.HandlerFunc {
	return func(c *gin.Context) {
		if email := c.GetString("email"); email != "" {
			log.Println(email)
			c.Next()
			return
		}

		// Access Token이 없거나 유효하지 않은 경우 Refresh Token 검증
		refreshToken, err := c.Cookie("refreshToken")
		if err != nil || refreshToken == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "로그인이 필요합니다."})
			c.Abort()
			return
		}

		//TODO  Redis에 저장된 리프레시인지 검증 틀리면, 재로그인 하라고 해야함
		err = i.authUsecase.ValidateRefreshToken(refreshToken)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "유효하지 않은 Refresh Token입니다. 다시 로그인 해주세요."})
			c.Abort()
			return
		}

		claims, err := util.ValidateRefreshToken(refreshToken)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "유효하지 않은 Refresh Token입니다. 다시 로그인 해주세요."})
			c.Abort()
			return
		}

		// TODO 위의 검증 맞으면, 액세스 토큰 새로 발급해야함
		newAccessToken, err := util.GenerateAccessToken(claims.Name, claims.Email, claims.UserId)
		if err != nil {
			log.Printf("액세스 토큰 재발급 실패: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "액세스 토큰 재발급에 실패했습니다"})
			c.Abort()
			return
		}
		c.SetCookie("accessToken", newAccessToken, 1200, "/", "", false, false)

		c.Set("email", claims.Email)
		c.Set("userId", claims.UserId)
		c.Next()
	}
}
