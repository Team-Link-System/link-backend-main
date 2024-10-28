package http

// import (
// 	"link/internal/team/usecase"
// 	"link/pkg/common"
// 	"net/http"

// 	"github.com/gin-gonic/gin"
// )

// type TeamHandler struct {
// 	// teamUsecase usecase.TeamUsecase
// }

// func NewTeamHandler(teamUsecase usecase.TeamUsecase) *TeamHandler {
// 	return &TeamHandler{teamUsecase: teamUsecase}
// }

// // 회사의 모든 팀 리스트 받기 요청 유저가 해당 회사에 속해있어야함
// func (h *TeamHandler) GetTeamsByCompany(c *gin.Context) {
// 	userId, exists := c.Get("userId")
// 	if !exists {
// 		c.JSON(http.StatusUnauthorized, common.NewError(http.StatusUnauthorized, "인증되지 않은 요청입니다"))
// 		return
// 	}

// 	response, err := h.teamUsecase.GetTeamsByCompany(userId.(uint))
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, common.NewError(http.StatusInternalServerError, "서버 에러"))
// 		return
// 	}

// 	c.JSON(http.StatusOK, response)
// }

// //해당 부서에 속한 팀리스트 받기
