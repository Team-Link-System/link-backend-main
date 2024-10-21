package http

import (
	_companyUsecase "link/internal/company/usecase"
	_userUsecase "link/internal/user/usecase"

	_companyEntity "link/internal/company/entity"
	_userEntity "link/internal/user/entity"
	"link/pkg/common"
	"link/pkg/dto/req"
	"link/pkg/dto/res"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AdminHandler struct {
	companyUsecase _companyUsecase.CompanyUsecase
	userUsecase    _userUsecase.UserUsecase
}

func NewAdminHandler(companyUsecase _companyUsecase.CompanyUsecase, userUsecase _userUsecase.UserUsecase) *AdminHandler {
	return &AdminHandler{companyUsecase: companyUsecase, userUsecase: userUsecase}
}

// TODO 운영자 등록 - 시스템 루트만 가능
func (h *AdminHandler) CreateAdmin(c *gin.Context) {
	//TODO 운영자 등록 로직 구현
	userId, exists := c.Get("userId") //루트 관리자인지 확인
	if !exists {
		c.JSON(http.StatusUnauthorized, common.NewError(http.StatusUnauthorized, "인증되지 않은 요청입니다"))
		return
	}

	var request req.CreateAdminRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "잘못된 요청입니다."))
		return
	}

	requestAdmin := &_userEntity.User{
		Email:    request.Email,
		Password: request.Password,
		Nickname: request.Nickname,
		Name:     request.Name,
		Phone:    request.Phone,
	}

	admin, err := h.userUsecase.RegisterAdmin(userId.(uint), requestAdmin)
	if err != nil {
		c.JSON(http.StatusInternalServerError, common.NewError(http.StatusInternalServerError, "운영자 등록에 실패하였습니다."))
		return
	}

	adminResponse := res.RegisterUserResponse{
		ID:       admin.ID,
		Name:     admin.Name,
		Email:    admin.Email,
		Phone:    admin.Phone,
		Nickname: admin.Nickname,
		Role:     uint(*admin.Role),
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
	users, err := h.userUsecase.GetAllUsers(requestUserID)
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
			ID:    user.ID,
			Name:  user.Name,
			Email: user.Email, // 민감 정보 포함할지 여부에 따라 처리
			Phone: user.Phone,
			Role:  uint(*user.Role),
			UserProfile: res.UserProfile{
				ID:           user.UserProfile.ID,
				Image:        user.UserProfile.Image,
				Birthday:     user.UserProfile.Birthday,
				CompanyID:    user.UserProfile.CompanyID,
				DepartmentID: user.UserProfile.DepartmentID,
				TeamID:       user.UserProfile.TeamID,
				PositionID:   user.UserProfile.PositionID,
			},
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
			Nickname:  user.Nickname,
		}

		response = append(response, userResponse)
	}

	// 응답으로 JSON 반환
	c.JSON(http.StatusOK, common.NewResponse(http.StatusOK, "사용자 목록 조회 성공", response))
}

// TODO 관리자 회사 등록
func (h *AdminHandler) CreateCompany(c *gin.Context) {
	var request req.CreateCompanyRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "잘못된 요청입니다."))
		return
	}

	company := &_companyEntity.Company{
		Name:                        request.Name,
		BusinessRegistrationNumber:  request.BusinessRegistrationNumber,
		RepresentativeName:          request.RepresentativeName,
		RepresentativePhoneNumber:   request.RepresentativePhoneNumber,
		RepresentativeEmail:         request.RepresentativeEmail,
		RepresentativeAddress:       request.RepresentativeAddress,
		RepresentativeAddressDetail: request.RepresentativeAddressDetail,
		RepresentativePostalCode:    request.RepresentativePostalCode,
	}

	company, err := h.companyUsecase.CreateCompany(company)
	if err != nil {
		c.JSON(http.StatusInternalServerError, common.NewError(http.StatusInternalServerError, "회사 등록에 실패하였습니다."))
		return
	}

	c.JSON(http.StatusCreated, common.NewResponse(http.StatusCreated, "회사 등록에 성공하였습니다.", company))
}
