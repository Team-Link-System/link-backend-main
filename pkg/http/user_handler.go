package http

import (
	"link/internal/user/entity"
	"link/internal/user/usecase"
	"log"
	"strconv"

	"link/pkg/dto/user/req"
	"link/pkg/dto/user/res"
	"link/pkg/interceptor"
	"net/http"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userUsecase usecase.UserUsecase
}

// RegisterUserHandler는 회원가입 핸들러를 생성합니다.
func NewUserHandler(userUsecase usecase.UserUsecase) *UserHandler {
	return &UserHandler{userUsecase: userUsecase}
}

// ! 회원가입 핸들러
func (h *UserHandler) RegisterUser(c *gin.Context) {
	var request req.RegisterUserRequest

	// 요청 바디 검증
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, interceptor.Error(http.StatusBadRequest, "잘못된 요청입니다."))
		return
	}

	// 유스케이스에 엔티티 전달
	user := &entity.User{
		Name:     request.Name,
		Email:    request.Email,
		Password: request.Password,
		Phone:    request.Phone,
	}

	createdUser, err := h.userUsecase.RegisterUser(user)
	if err != nil {
		c.JSON(http.StatusConflict, interceptor.Error(http.StatusConflict, err.Error()))
		return
	}

	// 유스케이스에서 반환된 엔티티를 응답 DTO로 변환
	response := res.RegisterUserResponse{
		ID:    createdUser.ID,
		Name:  createdUser.Name,
		Email: createdUser.Email,
		Phone: createdUser.Phone,
	}

	// 성공 응답
	c.JSON(http.StatusCreated, interceptor.Created("회원가입 완료", response))
}

// ! 이메일 검증 핸들러
func (h *UserHandler) ValidateEmail(c *gin.Context) {
	// 쿼리 파라미터에서 이메일 추출
	email := c.Query("email")

	// 이메일 파라미터가 없는 경우, 잘못된 요청 처리
	if email == "" {
		c.JSON(http.StatusBadRequest, interceptor.Error(http.StatusBadRequest, "이메일이 입력되지 않았습니다."))
		return
	}

	// 이메일 유효성 검증 처리
	if err := h.userUsecase.ValidateEmail(email); err != nil {
		c.JSON(http.StatusConflict, interceptor.Error(http.StatusConflict, err.Error()))
		return
	}

	// 검증 성공 응답
	c.JSON(http.StatusOK, interceptor.Success("이메일 사용 가능", nil))
}

// ! 사용자 전체 조회 핸들러 - 관리자만
func (h *UserHandler) GetAllUsers(c *gin.Context) {
	// 사용자 정보를 데이터베이스에서 조회
	userId, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, interceptor.Error(http.StatusUnauthorized, "인증되지 않은 요청입니다"))
		return
	}

	requestUserID, ok := userId.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, interceptor.Error(http.StatusInternalServerError, "사용자 ID 형식이 잘못되었습니다"))
		return
	}

	// 사용자 정보를 데이터베이스에서 조회
	users, err := h.userUsecase.GetAllUsers(requestUserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, interceptor.Error(http.StatusInternalServerError, err.Error()))
		return
	}

	// 응답 구조체로 변환
	var response []res.GetAllUsersResponse
	// 그룹 이름 또는 ID를 문자열 배열로 변환
	var groupNames []string
	for _, user := range users {
		// User 정보를 GetAllUsersResponse로 변환
		log.Printf("user: %v", user)
		userResponse := res.GetAllUsersResponse{
			ID:           user.ID,
			Name:         user.Name,
			Email:        user.Email, // 민감 정보 포함할지 여부에 따라 처리
			Phone:        user.Phone,
			Groups:       groupNames,        // 그룹 이름 배열
			DepartmentID: user.DepartmentID, // DepartmentID가 없으면 0으로 처리
			TeamID:       user.TeamID,       // TeamID가 없으면 0으로 처리
			Role:         uint(user.Role),
			CreatedAt:    user.CreatedAt,
			UpdatedAt:    user.UpdatedAt,
		}

		response = append(response, userResponse)
	}

	// 응답으로 JSON 반환
	c.JSON(http.StatusOK, interceptor.Success("사용자 목록 조회 성공", response))
}

func (h *UserHandler) GetUserInfo(c *gin.Context) {
	//params 받아야지
	userId, exists := c.Get("userId")
	targetUserId := c.Param("id")
	if !exists {
		c.JSON(http.StatusUnauthorized, interceptor.Error(http.StatusUnauthorized, "인증되지 않은 요청입니다"))
		return
	}

	requestUserID, ok := userId.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, interceptor.Error(http.StatusInternalServerError, "사용자 ID 형식이 잘못되었습니다"))
		return
	}

	targetUserIdUint, err := strconv.ParseUint(targetUserId, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, interceptor.Error(http.StatusBadRequest, "유효하지 않은 사용자 ID입니다"))
		return
	}

	user, err := h.userUsecase.GetUserInfo(uint(targetUserIdUint), requestUserID, "user")
	if err != nil {
		c.JSON(http.StatusInternalServerError, interceptor.Error(http.StatusInternalServerError, err.Error()))
	}

	var groupNames []string //TODO 일단 임시적으로 그룹 부서 팀 도메인 만들면 지워야함

	response := res.GetUserByIdResponse{
		ID:           user.ID,
		Email:        user.Email,
		Name:         user.Name,
		Phone:        user.Phone,
		Groups:       groupNames,
		DepartmentID: user.DepartmentID,
		TeamID:       user.TeamID,
		Role:         uint(user.Role),
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
	}

	c.JSON(http.StatusOK, interceptor.Success("사용자 조회 성공", response))
}

// 사용자 정보 수정 핸들러
func (h *UserHandler) UpdateUserInfo(c *gin.Context) {
	userId, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, interceptor.Error(http.StatusUnauthorized, "인증되지 않은 요청입니다"))
		return
	}

	requestUserId, ok := userId.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, interceptor.Error(http.StatusInternalServerError, "사용자 ID 형식이 잘못되었습니다"))
		return
	}

	targetUserId := c.Param("id")
	targetUserIdUint, err := strconv.ParseUint(targetUserId, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, interceptor.Error(http.StatusBadRequest, "유효하지 않은 사용자 ID입니다"))
		return
	}

	// JSON 요청 바인딩
	var request req.UpdateUserRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, interceptor.Error(http.StatusBadRequest, "잘못된 요청입니다."))
		return
	}

	// DTO를 Usecase에 전달
	err = h.userUsecase.UpdateUserInfo(uint(targetUserIdUint), requestUserId, request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, interceptor.Error(http.StatusInternalServerError, err.Error()))
		return
	}

	c.JSON(http.StatusOK, interceptor.Success("사용자 정보 수정 성공", nil))
}

// 사용자 정보 삭제 핸들러
func (h *UserHandler) DeleteUser(c *gin.Context) {
	userId, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, interceptor.Error(http.StatusUnauthorized, "인증되지 않은 요청입니다"))
		return
	}

	requestUserId, ok := userId.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, interceptor.Error(http.StatusInternalServerError, "사용자 ID 형식이 잘못되었습니다"))
		return
	}

	targetUserId := c.Param("id")
	targetUserIdUint, err := strconv.ParseUint(targetUserId, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, interceptor.Error(http.StatusBadRequest, "유효하지 않은 사용자 ID입니다"))
		return
	}

	err = h.userUsecase.DeleteUser(uint(targetUserIdUint), requestUserId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, interceptor.Error(http.StatusInternalServerError, err.Error()))
		return
	}

	c.JSON(http.StatusOK, interceptor.Success("사용자 정보 삭제 성공", nil))
}

// TODO 사용자 검색 핸들러
// 사용자 검색 핸들러
func (h *UserHandler) SearchUser(c *gin.Context) {
	// 검색 요청 데이터 받기
	searchReq := req.SearchUserRequest{
		Email: c.Query("email"),
		Name:  c.Query("name"),
	}

	// 쿼리 내용이 아무것도 없는 경우
	if searchReq.Email == "" && searchReq.Name == "" {
		c.JSON(http.StatusBadRequest, interceptor.Error(http.StatusBadRequest, "이메일 혹은 이름이 입력되지 않았습니다."))
		return
	}

	// 사용자 검색
	users, err := h.userUsecase.SearchUser(searchReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, interceptor.Error(http.StatusInternalServerError, err.Error()))
		return
	}

	if len(users) == 0 {
		c.JSON(http.StatusNotFound, interceptor.Error(http.StatusNotFound, "사용자를 찾을 수 없습니다"))
		return
	}

	var response []res.SearchUserResponse
	// 그룹 이름 또는 ID를 문자열 배열로 변환
	var groupNames []string
	for _, user := range users {
		log.Printf("user: %v", user)
		userResponse := res.SearchUserResponse{
			ID:           user.ID,
			Name:         user.Name,
			Email:        user.Email, // 민감 정보 포함할지 여부에 따라 처리
			Phone:        user.Phone,
			Groups:       groupNames,        // 그룹 이름 배열
			DepartmentID: user.DepartmentID, // DepartmentID가 없으면 0으로 처리
			TeamID:       user.TeamID,       // TeamID가 없으면 0으로 처리
			Role:         uint(user.Role),
			CreatedAt:    user.CreatedAt,
			UpdatedAt:    user.UpdatedAt,
		}

		response = append(response, userResponse)
	}

	c.JSON(http.StatusOK, interceptor.Success("사용자 검색 성공", response))
}
