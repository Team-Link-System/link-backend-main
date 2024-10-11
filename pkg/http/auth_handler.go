package http

import (
	"link/internal/auth/usecase"
	"link/pkg/dto/req"
	"link/pkg/dto/res"
	"link/pkg/interceptor"
	"net/http"

	"github.com/gin-gonic/gin"
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
		c.JSON(http.StatusBadRequest, interceptor.Error(http.StatusBadRequest, "잘못된 요청입니다"))
		return
	}

	user, token, err := h.authUsecase.SignIn(request.Email, request.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, interceptor.Error(http.StatusUnauthorized, err.Error()))
		return
	}

	response := res.LoginUserResponse{
		ID:    user.ID,
		Email: user.Email,
		Name:  user.Name,
		Role:  uint(user.Role),
	}

	//! 도메인 다를 때 사용

	c.SetCookie("accessToken", token.AccessToken, 1200, "/", "", false, true)
	c.SetCookie("refreshToken", token.RefreshToken, 5*24*3600, "/", "", false, true)
	c.JSON(http.StatusOK, interceptor.Success("로그인 성공", response))
}

func (h *AuthHandler) SignOut(c *gin.Context) {
	refreshToken, err := c.Cookie("refreshToken")
	if err != nil {
		c.JSON(http.StatusUnauthorized, interceptor.Error(http.StatusUnauthorized, "인증되지 않은 요청입니다"))
		return
	}

	err = h.authUsecase.SignOut(refreshToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, interceptor.Error(http.StatusInternalServerError, "로그아웃 처리 중 오류가 발생했습니다"))
		return
	}

	c.SetCookie("accessToken", "", -1, "/", "", false, true)
	c.SetCookie("refreshToken", "", -1, "/", "", false, true)
	c.JSON(http.StatusOK, interceptor.Success("로그아웃 되었습니다", nil))
}
