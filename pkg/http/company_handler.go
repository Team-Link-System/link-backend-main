package http

import (
	"link/internal/company/usecase"
	"link/pkg/common"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CompanyHandler struct {
	companyUsecase usecase.CompanyUsecase
}

func NewCompanyHandler(companyUsecase usecase.CompanyUsecase) *CompanyHandler {
	return &CompanyHandler{companyUsecase: companyUsecase}
}

// TODO 회사 전체 목록 불러오기 - 모든 사용자 사용 가능
func (h *CompanyHandler) GetAllCompanies(c *gin.Context) {
	companies, err := h.companyUsecase.GetAllCompanies()
	if err != nil {
		if appError, ok := err.(*common.AppError); ok {
			c.JSON(appError.StatusCode, common.NewError(appError.StatusCode, appError.Message))
		} else {
			c.JSON(http.StatusInternalServerError, common.NewError(http.StatusInternalServerError, "서버 에러"))
		}
		return
	}

	c.JSON(http.StatusOK, common.NewResponse(http.StatusOK, "회사 전체 목록 조회 성공", companies))
}

// TODO 회사 상세 조회 - 모든 사용자 사용가능 메서드 (회사에 속한 사용자들은 보여주면안됨)
func (h *CompanyHandler) GetCompanyInfo(c *gin.Context) {
	companyId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "잘못된 요청입니다"))
		return
	}

	company, err := h.companyUsecase.GetCompanyInfo(uint(companyId))
	if err != nil {
		if appError, ok := err.(*common.AppError); ok {
			c.JSON(appError.StatusCode, common.NewError(appError.StatusCode, appError.Message))
		} else {
			c.JSON(http.StatusInternalServerError, common.NewError(http.StatusInternalServerError, "서버 에러"))
		}
	}

	c.JSON(http.StatusOK, common.NewResponse(http.StatusOK, "회사 상세 조회 성공", company))
}

//TODO 회사 검색 - 모든 사용자 사용가능

//TODO 회사 생성은 관리자만 가능 -> 유저가 요청하는 것임(따로 admin도메인에 요청 핸들러만들것)

//TODO 회사 수정은 관리자만 가능 -> 유저가 요청하는 것임(따로 admin도메인에 요청 핸들러만들것)

//TODO 회사 삭제는 관리자만 가능 -> 유저가 요청하는 것임(따로 admin도메인에 요청 핸들러만들것)
