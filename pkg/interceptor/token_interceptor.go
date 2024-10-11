package interceptor

import (
	"fmt"
	"log"
	"net/http"

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

// TODO CROSS-STIE 요청일 경우 OPTIONS 요청은 토큰검증없이 바로 처리
// Access Token 검증 인터셉터
// Access Token 검증 인터셉터
func (i *TokenInterceptor) AccessTokenInterceptor() gin.HandlerFunc {
	return func(c *gin.Context) {
		// OPTIONS 요청은 인증 없이 바로 처리

		accessToken, _ := c.Cookie("accessToken")
		fmt.Println("accessToken")
		fmt.Println(accessToken)
		if accessToken != "" {
			// Access Token 검증
			claims, err := util.ValidateAccessToken(accessToken)
			if err == nil {
				// Access Token이 유효한 경우 email과 userId를 Context에 설정
				c.Set("email", claims.Email)
				c.Set("userId", claims.UserId)
				c.Next() // Access Token이 유효하면 다음 핸들러로 진행
				return
			}
		}
		// Access Token이 유효하지 않으면 다음 단계로 넘어감 (Refresh Token 검증)
		c.Next()
	}
}

// Refresh Token 검증 및 Access Token 재발급 인터셉터
func (i *TokenInterceptor) RefreshTokenInterceptor() gin.HandlerFunc {
	return func(c *gin.Context) {
		// OPTIONS 요청은 인증 없이 바로 처리
		// OPTIONS 요청은 인증 없이 바로 처리
		email, _ := c.Get("email")
		userId, _ := c.Get("userId")

		// Access Token이 유효하다면 바로 다음 단계로
		if email != nil && userId != nil {
			c.Next()
			return
		}

		// Access Token이 없거나 유효하지 않은 경우 Refresh Token 검증
		refreshToken, err := c.Cookie("refreshToken")
		fmt.Println("refreshToken")
		fmt.Println(refreshToken)
		if err != nil || refreshToken == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "로그인이 필요합니다."})
			c.Abort()
			return
		}

		// Redis에 저장된 리프레시 토큰 검증
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

		// Access Token 재발급
		newAccessToken, err := util.GenerateAccessToken(claims.Name, claims.Email, claims.UserId)
		if err != nil {
			log.Printf("액세스 토큰 재발급 실패: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "액세스 토큰 재발급에 실패했습니다"})
			c.Abort()
			return
		}
		c.SetCookie("accessToken", newAccessToken, 1200, "/", "", false, true)
		// 새로운 Access Token 정보로 Context 설정
		c.Set("email", claims.Email)
		c.Set("userId", claims.UserId)
		c.Next()
	}
}
