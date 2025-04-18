package http

import (
	"link/pkg/common"
	"link/pkg/dto/req"
	"net/http"

	_positionUsecase "github.com/danggeun/danggeun-server/internal/position/usecase"
	"github.com/gin-gonic/gin"
)

type PositionHandler struct {
	positionUsecase _positionUsecase.PositionUsecase
}

func NewPositionHandler(positionUsecase _positionUsecase.PositionUsecase) *PositionHandler {
	return &PositionHandler{positionUsecase: positionUsecase}
}

func (h *PositionHandler) CreatePosition(c *gin.Context) {

	userId, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, common.NewError(http.StatusUnauthorized, "인증되지 않은 요청입니다.", nil))
		return
	}

	requestUserId, ok := userId.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, common.NewError(http.StatusInternalServerError, "사용자 ID 형식이 잘못되었습니다", nil))
		return
	}

	companyId := c.Param("companyId")
	if companyId == "" {
		c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "회사 ID가 필요합니다", nil))
		return
	}

	var request req.CreatePositionRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "잘못된 요청입니다", err))
		return
	}

	if err := h.positionUsecase.CreatePosition(requestUserId, companyId, request); err != nil {
		if appError, ok := err.(*common.AppError); ok {
			c.JSON(appError.StatusCode, common.NewError(appError.StatusCode, appError.Message, appError.Err))
		} else {
			c.JSON(http.StatusInternalServerError, common.NewError(http.StatusInternalServerError, "직책 생성 실패", err))
		}
		return
	}

	c.JSON(http.StatusCreated, common.NewResponse(http.StatusCreated, "직책 생성 성공", nil))
}

// func (h *PositionHandler) GetPositions(c *gin.Context) {

// }

// func (h *PositionHandler) GetPosition(c *gin.Context) {

// }

// func (h *PositionHandler) UpdatePosition(c *gin.Context) {

// }

// func (h *PositionHandler) DeletePosition(c *gin.Context) {

// }
