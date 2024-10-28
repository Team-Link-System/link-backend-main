package http

import (
	_adminUsecase "link/internal/admin/usecase"
	"strconv"

	"link/pkg/common"
	"link/pkg/dto/req"
	"link/pkg/dto/res"
	"link/pkg/util"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AdminHandler struct {
	adminUsecase _adminUsecase.AdminUsecase
}

func NewAdminHandler(adminUsecase _adminUsecase.AdminUsecase) *AdminHandler {
	return &AdminHandler{adminUsecase: adminUsecase}
}

// TODO Role 1,2만 가능
// TODO 운영자 등록 - 시스템 루트만 가능
func (h *AdminHandler) CreateAdmin(c *gin.Context) {
	//TODO 운영자 등록 로직 구현
	userId, exists := c.Get("userId") //루트 관리자인지 확인
	if !exists {
		c.JSON(http.StatusUnauthorized, common.NewError(http.StatusUnauthorized, "인증되지 않은 요청입니다"))
		return
	}

	var request req.AdminCreateAdminRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "잘못된 요청입니다."))
		return
	}

	admin, err := h.adminUsecase.RegisterAdmin(userId.(uint), &request)
	if err != nil {
		if appError, ok := err.(*common.AppError); ok {
			c.JSON(appError.StatusCode, common.NewError(appError.StatusCode, appError.Message))
		} else {
			c.JSON(http.StatusInternalServerError, common.NewError(http.StatusInternalServerError, "서버 에러"))
		}
		return
	}

	adminResponse := res.RegisterUserResponse{
		Name:     *admin.Name,
		Email:    *admin.Email,
		Phone:    *admin.Phone,
		Nickname: *admin.Nickname,
		Role:     uint(admin.Role),
	}

	c.JSON(http.StatusCreated, common.NewResponse(http.StatusCreated, "운영자 등록에 성공하였습니다.", adminResponse))
}

// ! 사용자 전체 조회 핸들러 - 관리자만
func (h *AdminHandler) GetAllUsers(c *gin.Context) {
	// 사용자 정보를 데이터베이스에서 조회
	userId, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, common.NewError(http.StatusUnauthorized, "인증되지 않은 요청입니다"))
		return
	}

	requestUserID, ok := userId.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, common.NewError(http.StatusInternalServerError, "사용자 ID 형식이 잘못되었습니다"))
		return
	}

	// 사용자 정보를 데이터베이스에서 조회
	users, err := h.adminUsecase.GetAllUsers(requestUserID)
	if err != nil {
		if appError, ok := err.(*common.AppError); ok {
			c.JSON(appError.StatusCode, common.NewError(appError.StatusCode, appError.Message))
		} else {
			c.JSON(http.StatusInternalServerError, common.NewError(http.StatusInternalServerError, "서버 에러"))
		}
		return
	}

	// 응답 구조체로 변환
	var response []res.GetAllUsersResponse
	// 그룹 이름 또는 ID를 문자열 배열로 변환
	for _, user := range users {
		// User 정보를 GetAllUsersResponse로 변환
		userResponse := res.GetAllUsersResponse{
			ID:              *user.ID,
			Name:            *user.Name,
			Email:           *user.Email, // 민감 정보 포함할지 여부에 따라 처리
			Phone:           *user.Phone,
			Role:            uint(user.Role),
			Image:           user.UserProfile.Image,
			Birthday:        user.UserProfile.Birthday,
			CompanyID:       util.GetValueOrDefault(user.UserProfile.CompanyID, 0),
			CompanyName:     util.GetFirstOrEmpty(util.ExtractValuesFromMapSlice[string]([]*map[string]interface{}{user.UserProfile.Company}, "name"), ""),
			DepartmentNames: util.ExtractValuesFromMapSlice[string](user.UserProfile.Departments, "name"),
			TeamNames:       util.ExtractValuesFromMapSlice[string](user.UserProfile.Teams, "name"),
			PositionId:      util.GetValueOrDefault(user.UserProfile.PositionId, 0),
			PositionName:    util.GetFirstOrEmpty(util.ExtractValuesFromMapSlice[string]([]*map[string]interface{}{user.UserProfile.Position}, "name"), ""),
			CreatedAt:       *user.CreatedAt,
			UpdatedAt:       *user.UpdatedAt,
			Nickname:        *user.Nickname,
		}

		response = append(response, userResponse)
	}

	// 응답으로 JSON 반환
	c.JSON(http.StatusOK, common.NewResponse(http.StatusOK, "사용자 목록 조회 성공", response))
}

// TODO 회사 생성 (관리자 isVerified = true)
func (h *AdminHandler) CreateCompany(c *gin.Context) {
	userId, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, common.NewError(http.StatusUnauthorized, "인증되지 않은 요청입니다"))
		return
	}

	requestUserID, ok := userId.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, common.NewError(http.StatusInternalServerError, "사용자 ID 형식이 잘못되었습니다"))
		return
	}

	var request req.AdminCreateCompanyRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "잘못된 요청입니다."))
		return
	}

	company, err := h.adminUsecase.CreateCompany(requestUserID, &request)
	if err != nil {
		if appError, ok := err.(*common.AppError); ok {
			c.JSON(appError.StatusCode, common.NewError(appError.StatusCode, appError.Message))
		} else {
			c.JSON(http.StatusInternalServerError, common.NewError(http.StatusInternalServerError, "서버 에러"))
		}
		return
	}

	c.JSON(http.StatusCreated, common.NewResponse(http.StatusCreated, "회사 등록에 성공하였습니다.", company))
}

// TODO 회사 삭제 - ADMIN
func (h *AdminHandler) DeleteCompany(c *gin.Context) {
	companyID, err := strconv.Atoi(c.Param("company_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "잘못된 요청입니다."))
		return
	}

	userId, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, common.NewError(http.StatusUnauthorized, "인증되지 않은 요청입니다"))
		return
	}

	requestUserID, ok := userId.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, common.NewError(http.StatusInternalServerError, "사용자 ID 형식이 잘못되었습니다"))
		return
	}

	deletedCompany, err := h.adminUsecase.DeleteCompany(requestUserID, uint(companyID))
	if err != nil {
		if appError, ok := err.(*common.AppError); ok {
			c.JSON(appError.StatusCode, common.NewError(appError.StatusCode, appError.Message))
		} else {
			c.JSON(http.StatusInternalServerError, common.NewError(http.StatusInternalServerError, "서버 에러"))
		}
		return
	}

	c.JSON(http.StatusOK, common.NewResponse(http.StatusOK, "회사 삭제에 성공하였습니다.", deletedCompany))
}

//TODO 아래서부터는 Role 3까지 가능
