package http

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"link/internal/auth/usecase"
	"link/pkg/common"
	"link/pkg/dto/req"
	"link/pkg/util"
)

type AuthHandler struct {
	authUsecase usecase.AuthUsecase
}

func NewAuthHandler(authUsecase usecase.AuthUsecase) *AuthHandler {
	return &AuthHandler{authUsecase: authUsecase}
}

func (h *AuthHandler) SignIn(c *gin.Context) {
	var request req.LoginRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "잘못된 요청입니다"))
		return
	}

	response, token, err := h.authUsecase.SignIn(&request)
	if err != nil {
		if appError, ok := err.(*common.AppError); ok {
			c.JSON(appError.StatusCode, common.NewError(appError.StatusCode, appError.Message))
		} else {
			c.JSON(http.StatusInternalServerError, common.NewError(http.StatusInternalServerError, "서버 에러"))
		}
		return
	}

	//! 도메인 다를 때 사용
	authorization := fmt.Sprintf("Bearer %s", token.AccessToken)

	// c.SetCookie("accessToken", token.AccessToken, 1200, "/", "", false, true)
	c.Header("Authorization", authorization)
	c.JSON(http.StatusOK, common.NewResponse(http.StatusOK, "로그인 성공", response))
}

func (h *AuthHandler) SignOut(c *gin.Context) {
	// userId를 가져옵니다.
	userId, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, common.NewError(http.StatusUnauthorized, "인증되지 않은 요청입니다"))
		return
	}

	// email을 가져옵니다.
	email, exists := c.Get("email")
	if !exists {
		c.JSON(http.StatusUnauthorized, common.NewError(http.StatusUnauthorized, "인증되지 않은 요청입니다"))
		return
	}

	// 로그아웃 처리 로직 호출
	err := h.authUsecase.SignOut(userId.(uint), email.(string))
	if err != nil {
		if appError, ok := err.(*common.AppError); ok {
			c.JSON(appError.StatusCode, common.NewError(appError.StatusCode, appError.Message))
		} else {
			c.JSON(http.StatusInternalServerError, common.NewError(http.StatusInternalServerError, "서버 에러"))
		}
		return
	}

	// accessToken 쿠키 삭제
	// TODO 리프레시 토큰은 레디스에 저장되어 있음
	c.SetCookie("accessToken", "", -1, "/", "", false, true)
	c.JSON(http.StatusOK, common.NewResponse(http.StatusOK, "로그아웃 되었습니다", nil))
}

// TODO accessToken 재발급 핸들러
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	//TODO 액세스 토큰 재발급 로직 구현

	//TODO accessToken을 setTimeOut으로 넘겨옴(next서버에서 15분마다 넘김) -> validate가 됐다면 재발급 -> 재발급 된 accessToken을 쿠키에 저장

	//TODO 없다면 로그아웃 처리

	userId, exists := c.Get("userId")
	if !exists {
		log.Printf("유저 ID 가져오기 중 오류가 발생했습니다")
		c.JSON(http.StatusUnauthorized, common.NewError(http.StatusUnauthorized, "인증되지 않은 요청입니다"))
		return
	}

	email, exists := c.Get("email")
	if !exists {
		log.Printf("이메일 가져오기 중 오류가 발생했습니다")
		c.JSON(http.StatusUnauthorized, common.NewError(http.StatusUnauthorized, "인증되지 않은 요청입니다"))
		return
	}

	refreshToken, err := h.authUsecase.GetRefreshToken(userId.(uint), email.(string))
	if err != nil {
		log.Printf("리프레시 토큰 가져오기 중 오류가 발생했습니다: %v", err)
		if appError, ok := err.(*common.AppError); ok {
			c.JSON(appError.StatusCode, common.NewError(appError.StatusCode, appError.Message))
		} else {
			c.JSON(http.StatusUnauthorized, common.NewError(http.StatusUnauthorized, "인증되지 않은 요청입니다"))
		}
		return
	}

	//TODO refreshToken 으로 accessToken 재발급
	claims, err := util.ValidateRefreshToken(refreshToken)
	if err != nil {
		log.Printf("리프레시 토큰 검증 중 오류가 발생했습니다: %v", err)
		if appError, ok := err.(*common.AppError); ok {
			c.JSON(appError.StatusCode, common.NewError(appError.StatusCode, appError.Message))
		} else {
			c.JSON(http.StatusUnauthorized, common.NewError(http.StatusUnauthorized, "인증되지 않은 요청입니다"))
		}
		return
	}

	newAccessToken, err := util.GenerateAccessToken(claims.Name, claims.Email, claims.UserId)
	if err != nil {
		log.Printf("액세스 토큰 재발급 중 오류가 발생했습니다: %v", err)
		if appError, ok := err.(*common.AppError); ok {
			c.JSON(appError.StatusCode, common.NewError(appError.StatusCode, appError.Message))
		} else {
			c.JSON(http.StatusInternalServerError, common.NewError(http.StatusInternalServerError, "서버 에러"))
		}
		c.Abort()
		return
	}

	//TODO 재발급 된 accessToken을 헤더로 전송
	authorization := fmt.Sprintf("Bearer %s", newAccessToken)
	c.Header("Authorization", authorization)
	// c.SetCookie("accessToken", newAccessToken, 1200, "/", "", false, true)
	c.JSON(http.StatusOK, common.NewResponse(http.StatusOK, "액세스 토큰 재발급 성공", nil))
}
