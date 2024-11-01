package http

import (
	"fmt"
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

//! Rule 관리자 이상급으로 할 수 있는 메서드는 다 Admin이 붙음

// TODO Role 1,2만 가능
// TODO 운영자 등록 - 시스템 루트만 가능
func (h *AdminHandler) AdminCreateAdmin(c *gin.Context) {
	//TODO 운영자 등록 로직 구현
	userId, exists := c.Get("userId") //루트 관리자인지 확인
	if !exists {
		fmt.Printf("인증되지 않은 요청입니다")
		c.JSON(http.StatusUnauthorized, common.NewError(http.StatusUnauthorized, "인증되지 않은 요청입니다", nil))
		return
	}

	var request req.AdminCreateAdminRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		fmt.Printf("잘못된 요청입니다: %v", err)
		c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "잘못된 요청입니다.", err))
		return
	}

	admin, err := h.adminUsecase.AdminRegisterAdmin(userId.(uint), &request)
	if err != nil {
		if appError, ok := err.(*common.AppError); ok {
			fmt.Printf("운영자 등록 오류: %v", appError.Err)
			c.JSON(appError.StatusCode, common.NewError(appError.StatusCode, appError.Message, appError.Err))
		} else {
			fmt.Printf("운영자 등록 오류: %v", err)
			c.JSON(http.StatusInternalServerError, common.NewError(http.StatusInternalServerError, "서버 에러", err))
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

// ! 사용자 전체 조회 핸들러 - 관리자만 -> 얘도 나중에 쿼리 추가
func (h *AdminHandler) AdminGetAllUsers(c *gin.Context) {
	// 사용자 정보를 데이터베이스에서 조회
	userId, exists := c.Get("userId")
	if !exists {
		fmt.Printf("인증되지 않은 요청입니다")
		c.JSON(http.StatusUnauthorized, common.NewError(http.StatusUnauthorized, "인증되지 않은 요청입니다", nil))
		return
	}

	requestUserID, ok := userId.(uint)
	if !ok {
		fmt.Printf("사용자 ID 형식이 잘못되었습니다")
		c.JSON(http.StatusInternalServerError, common.NewError(http.StatusInternalServerError, "사용자 ID 형식이 잘못되었습니다", nil))
		return
	}

	// 사용자 정보를 데이터베이스에서 조회
	users, err := h.adminUsecase.AdminGetAllUsers(requestUserID)
	if err != nil {
		if appError, ok := err.(*common.AppError); ok {
			fmt.Printf("사용자 전체 조회 오류: %v", appError.Err)
			c.JSON(appError.StatusCode, common.NewError(appError.StatusCode, appError.Message, appError.Err))
		} else {
			fmt.Printf("사용자 전체 조회 오류: %v", err)
			c.JSON(http.StatusInternalServerError, common.NewError(http.StatusInternalServerError, "서버 에러", err))
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
func (h *AdminHandler) AdminCreateCompany(c *gin.Context) {
	userId, exists := c.Get("userId")
	if !exists {
		fmt.Printf("인증되지 않은 요청입니다")
		c.JSON(http.StatusUnauthorized, common.NewError(http.StatusUnauthorized, "인증되지 않은 요청입니다", nil))
		return
	}

	requestUserID, ok := userId.(uint)
	if !ok {
		fmt.Printf("사용자 ID 형식이 잘못되었습니다")
		c.JSON(http.StatusInternalServerError, common.NewError(http.StatusInternalServerError, "사용자 ID 형식이 잘못되었습니다", nil))
		return
	}

	var request req.AdminCreateCompanyRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		fmt.Printf("잘못된 요청입니다: %v", err)
		c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "잘못된 요청입니다.", err))
		return
	}

	company, err := h.adminUsecase.AdminCreateCompany(requestUserID, &request)
	if err != nil {
		if appError, ok := err.(*common.AppError); ok {
			fmt.Printf("회사 생성 오류: %v", appError.Err)
			c.JSON(appError.StatusCode, common.NewError(appError.StatusCode, appError.Message, appError.Err))
		} else {
			fmt.Printf("회사 생성 오류: %v", err)
			c.JSON(http.StatusInternalServerError, common.NewError(http.StatusInternalServerError, "서버 에러", err))
		}
		return
	}

	c.JSON(http.StatusCreated, common.NewResponse(http.StatusCreated, "회사 등록에 성공하였습니다.", company))
}

// TODO 회사 삭제 - ADMIN
func (h *AdminHandler) AdminDeleteCompany(c *gin.Context) {
	companyID, err := strconv.Atoi(c.Param("companyid"))
	if err != nil {
		fmt.Printf("잘못된 요청입니다: %v", err)
		c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "잘못된 요청입니다.", err))
		return
	}

	userId, exists := c.Get("userId")
	if !exists {
		fmt.Printf("인증되지 않은 요청입니다")
		c.JSON(http.StatusUnauthorized, common.NewError(http.StatusUnauthorized, "인증되지 않은 요청입니다", nil))
		return
	}

	requestUserID, ok := userId.(uint)
	if !ok {
		fmt.Printf("사용자 ID 형식이 잘못되었습니다")
		c.JSON(http.StatusInternalServerError, common.NewError(http.StatusInternalServerError, "사용자 ID 형식이 잘못되었습니다", nil))
		return
	}

	err = h.adminUsecase.AdminDeleteCompany(requestUserID, uint(companyID))
	if err != nil {
		if appError, ok := err.(*common.AppError); ok {
			fmt.Printf("회사 삭제 오류: %v", appError.Err)
			c.JSON(appError.StatusCode, common.NewError(appError.StatusCode, appError.Message, appError.Err))
		} else {
			fmt.Printf("회사 삭제 오류: %v", err)
			c.JSON(http.StatusInternalServerError, common.NewError(http.StatusInternalServerError, "서버 에러", err))
		}
		return
	}

	c.JSON(http.StatusOK, common.NewResponse(http.StatusOK, "회사 삭제에 성공하였습니다.", nil))
}

// TODO 회사 업데이트 - ADMIN
func (h *AdminHandler) AdminUpdateCompany(c *gin.Context) {
	adminUserId, exists := c.Get("userId")
	if !exists {
		fmt.Printf("인증되지 않은 요청입니다")
		c.JSON(http.StatusUnauthorized, common.NewError(http.StatusUnauthorized, "인증되지 않은 요청입니다", nil))
		return
	}

	var request req.AdminUpdateCompanyRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		fmt.Printf("잘못된 요청입니다: %v", err)
		c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "잘못된 요청입니다.", err))
		return
	}

	err := h.adminUsecase.AdminUpdateCompany(adminUserId.(uint), &request)
	if err != nil {
		if appError, ok := err.(*common.AppError); ok {
			fmt.Printf("회사 업데이트 오류: %v", appError.Err)
			c.JSON(appError.StatusCode, common.NewError(appError.StatusCode, appError.Message, appError.Err))
		} else {
			fmt.Printf("회사 업데이트 오류: %v", err)
			c.JSON(http.StatusInternalServerError, common.NewError(http.StatusInternalServerError, "서버 에러", err))
		}
		return
	}

	c.JSON(http.StatusOK, common.NewResponse(http.StatusOK, "회사 업데이트에 성공하였습니다.", nil))
}

// TODO 회사에 속한 모든 사용자 리스트 조회 (관리자만)
func (h *AdminHandler) AdminGetUsersByCompany(c *gin.Context) {
	adminUserId, exists := c.Get("userId")
	if !exists {
		fmt.Printf("인증되지 않은 요청입니다")
		c.JSON(http.StatusUnauthorized, common.NewError(http.StatusUnauthorized, "인증되지 않은 요청입니다", nil))
		return
	}

	companyID, err := strconv.Atoi(c.Param("companyid"))
	if err != nil {
		fmt.Printf("잘못된 요청입니다: %v", err)
		c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "잘못된 요청입니다.", err))
		return
	}

	//TODO query 스트링에 따라 조회가 달라지는 로직
	// 쿼리 옵션 설정
	queryOptions := req.UserQuery{
		SortBy: req.UserSortBy(c.Query("sortby")),
		Order:  req.UserSortOrder(c.Query("order")),
	}

	if queryOptions.SortBy == "" {
		queryOptions.SortBy = "birthday" // TODO 기본 값은 서비스 가입일
	}
	if queryOptions.Order != "asc" && queryOptions.Order != "desc" {
		queryOptions.Order = "asc" // TODO 기본 값은 오름차순
	}

	users, err := h.adminUsecase.AdminGetUsersByCompany(adminUserId.(uint), uint(companyID), &queryOptions)
	if err != nil {
		if appError, ok := err.(*common.AppError); ok {
			fmt.Printf("회사에 속한 사용자 목록 조회 오류: %v", appError.Err)
			c.JSON(appError.StatusCode, common.NewError(appError.StatusCode, appError.Message, appError.Err))
		} else {
			fmt.Printf("회사에 속한 사용자 목록 조회 오류: %v", err)
			c.JSON(http.StatusInternalServerError, common.NewError(http.StatusInternalServerError, "서버 에러", err))
		}
		return
	}

	c.JSON(http.StatusOK, common.NewResponse(http.StatusOK, "회사에 속한 사용자 목록 조회 성공", users))
}

// TODO 회사에 사용자 추가
func (h *AdminHandler) AdminAddUserToCompany(c *gin.Context) {
	adminUserId, exists := c.Get("userId")
	if !exists {
		fmt.Printf("인증되지 않은 요청입니다")
		c.JSON(http.StatusUnauthorized, common.NewError(http.StatusUnauthorized, "인증되지 않은 요청입니다", nil))
		return
	}

	var request req.AdminAddUserToCompanyRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		fmt.Printf("잘못된 요청입니다: %v", err)
		c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "잘못된 요청입니다.", err))
		return
	}

	err := h.adminUsecase.AdminAddUserToCompany(adminUserId.(uint), request.UserID, request.CompanyID)
	if err != nil {
		if appError, ok := err.(*common.AppError); ok {
			fmt.Printf("회사에 사용자 추가 오류: %v", appError.Err)
			c.JSON(appError.StatusCode, common.NewError(appError.StatusCode, appError.Message, appError.Err))
		} else {
			fmt.Printf("회사에 사용자 추가 오류: %v", err)
			c.JSON(http.StatusInternalServerError, common.NewError(http.StatusInternalServerError, "서버 에러", err))
		}
		return
	}

	c.JSON(http.StatusOK, common.NewResponse(http.StatusOK, "회사에 사용자 추가에 성공하였습니다.", nil))
}
