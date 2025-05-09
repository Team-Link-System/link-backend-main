package http

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/gin-gonic/gin"

	"link/internal/user/usecase"
	"link/pkg/common"
	"link/pkg/dto/req"
)

type UserHandler struct {
	userUsecase usecase.UserUsecase
}

// RegisterUserHandler는 회원가입 핸들러를 생성합니다.
func NewUserHandler(userUsecase usecase.UserUsecase) *UserHandler {
	return &UserHandler{userUsecase: userUsecase}
}

// ! 회원가입 핸들러

// TODO 회원가입할 때 userProfile 빈값들로 일단 생성
func (h *UserHandler) RegisterUser(c *gin.Context) {
	var request req.RegisterUserRequest

	// 요청 바디 검증
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "잘못된 요청입니다.", err))
		return
	}

	response, err := h.userUsecase.RegisterUser(&request)
	if err != nil {
		if appError, ok := err.(*common.AppError); ok {
			c.JSON(appError.StatusCode, common.NewError(appError.StatusCode, appError.Message, appError.Err))
		} else {
			c.JSON(http.StatusInternalServerError, common.NewError(http.StatusInternalServerError, "서버 에러", err))
		}
		return
	}
	// 성공 응답
	c.JSON(http.StatusCreated, common.NewResponse(http.StatusCreated, "회원가입 완료", response))
}

// ! 이메일 검증 핸들러
func (h *UserHandler) ValidateEmail(c *gin.Context) {
	// 쿼리 파라미터에서 이메일 추출
	email := c.Query("email")

	// 이메일 파라미터가 없는 경우, 잘못된 요청 처리
	if email == "" {
		c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "이메일이 입력되지 않았습니다.", nil))
		return
	}

	// 이메일 유효성 검증 처리
	if err := h.userUsecase.ValidateEmail(email); err != nil {
		if appError, ok := err.(*common.AppError); ok {
			c.JSON(appError.StatusCode, common.NewError(appError.StatusCode, appError.Message, appError.Err))
		} else {
			c.JSON(http.StatusInternalServerError, common.NewError(http.StatusInternalServerError, "서버 에러", err))
		}
		return
	}

	// 검증 성공 응답
	c.JSON(http.StatusOK, common.NewResponse(http.StatusOK, "이메일 사용 가능", nil))
}

// TODO 닉네임 중복확인
func (h *UserHandler) ValidateNickname(c *gin.Context) {
	nickname := c.Query("nickname")

	if nickname == "" {
		c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "닉네임이 입력되지 않았습니다", nil))
		return
	}

	err := h.userUsecase.ValidateNickname(nickname)
	if err != nil {
		if appError, ok := err.(*common.AppError); ok {
			c.JSON(appError.StatusCode, common.NewError(appError.StatusCode, appError.Message, appError.Err))
		} else {
			c.JSON(http.StatusInternalServerError, common.NewError(http.StatusInternalServerError, "서버 에러", err))
		}
		return
	}

	c.JSON(http.StatusOK, common.NewResponse(http.StatusOK, "사용 가능한 닉네임입니다", nil))
}

// TODO UserProfile 있으면
func (h *UserHandler) GetUserInfo(c *gin.Context) {
	//params 받아야지
	userId, exists := c.Get("userId")
	targetUserId := c.Param("id")
	if !exists {
		c.JSON(http.StatusUnauthorized, common.NewError(http.StatusUnauthorized, "인증되지 않은 요청입니다", fmt.Errorf("userId가 없습니다")))
		return
	}

	fmt.Println("targetUserId", targetUserId)

	targetUserIdUint, err := strconv.ParseUint(targetUserId, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "유효하지 않은 사용자 ID입니다", err))
		return
	}

	response, err := h.userUsecase.GetUserInfo(userId.(uint), uint(targetUserIdUint), "user")
	if err != nil {
		if appError, ok := err.(*common.AppError); ok {
			c.JSON(appError.StatusCode, common.NewError(appError.StatusCode, appError.Message, appError.Err))
		} else {
			c.JSON(http.StatusInternalServerError, common.NewError(http.StatusInternalServerError, "서버 에러", err))
		}
		return
	}

	c.JSON(http.StatusOK, common.NewResponse(http.StatusOK, "사용자 조회 성공", response))
}

func (h *UserHandler) UpdateUserInfo(c *gin.Context) {
	requestUserId, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, common.NewError(http.StatusUnauthorized, "인증되지 않은 요청입니다", nil))
		return
	}

	targetUserId := c.Param("id")
	targetUserIdUint, err := strconv.ParseUint(targetUserId, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "유효하지 않은 사용자 ID입니다", err))
		return
	}

	var request req.UpdateUserRequest
	if err := c.ShouldBind(&request); err != nil {
		c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "잘못된 요청입니다", err))
		return
	}

	// 미들웨어에서 처리한 이미지 URL 가져오기
	profileImageUrl, exists := c.Get("profile_image_url")
	if exists {
		imageURL := profileImageUrl.(string)
		request.Image = &imageURL
	}

	// DTO를 Usecase에 전달
	err = h.userUsecase.UpdateUserInfo(uint(targetUserIdUint), requestUserId.(uint), &request)
	if err != nil {
		if appError, ok := err.(*common.AppError); ok {
			c.JSON(appError.StatusCode, common.NewError(appError.StatusCode, appError.Message, appError.Err))
		} else {
			c.JSON(http.StatusInternalServerError, common.NewError(http.StatusInternalServerError, "서버 에러", err))
		}
		return
	}

	c.JSON(http.StatusOK, common.NewResponse(http.StatusOK, "사용자 정보 수정 성공", nil))
}

// 사용자 정보 삭제 핸들러
func (h *UserHandler) DeleteUser(c *gin.Context) {
	requestUserId, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, common.NewError(http.StatusUnauthorized, "인증되지 않은 요청입니다", nil))
		return
	}

	targetUserId := c.Param("id")
	targetUserIdUint, err := strconv.ParseUint(targetUserId, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "유효하지 않은 사용자 ID입니다", err))
		return
	}

	err = h.userUsecase.DeleteUser(uint(targetUserIdUint), requestUserId.(uint))
	if err != nil {
		if appError, ok := err.(*common.AppError); ok {
			c.JSON(appError.StatusCode, common.NewError(appError.StatusCode, appError.Message, appError.Err))
		} else {
			c.JSON(http.StatusInternalServerError, common.NewError(http.StatusInternalServerError, "서버 에러", err))
		}
		return
	}

	c.JSON(http.StatusOK, common.NewResponse(http.StatusOK, "사용자 정보 삭제 성공", nil))
}

// TODO 사용자 검색 핸들러
// 사용자 검색 핸들러
func (h *UserHandler) SearchUser(c *gin.Context) {
	requestUserId, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, common.NewError(http.StatusUnauthorized, "인증되지 않은 요청입니다", nil))
		return
	}

	// 검색 요청 데이터 받기
	searchTerm := c.Query("searchTerm")
	decodedSearchTerm, err := url.QueryUnescape(searchTerm)
	if err != nil {
		c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "검색어 인코딩 오류", err))
		return
	}

	// 사용자 검색
	//TODO 검색 조건 나중에 회사사람만 검색가능하도록 해야함
	users, err := h.userUsecase.SearchUser(requestUserId.(uint), decodedSearchTerm)
	if err != nil {
		if appError, ok := err.(*common.AppError); ok {
			c.JSON(appError.StatusCode, common.NewError(appError.StatusCode, appError.Message, appError.Err))
		} else {
			c.JSON(http.StatusInternalServerError, common.NewError(http.StatusInternalServerError, "서버 에러", err))
		}
		return
	}

	c.JSON(http.StatusOK, common.NewResponse(http.StatusOK, "사용자 검색 성공", users))
}

// TODO 본인이 속한 회사 사용자 리스트 가져오기
func (h *UserHandler) GetUserByCompany(c *gin.Context) {
	requestUserId, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, common.NewError(http.StatusUnauthorized, "인증되지 않은 요청입니다", nil))
		return
	}

	queryOptions := req.UserQuery{
		SortBy: req.UserSortBy(c.Query("sortby")),
		Order:  req.UserSortOrder(c.Query("order")),
	}

	response, err := h.userUsecase.GetUsersByCompany(requestUserId.(uint), &queryOptions)
	if err != nil {
		if appError, ok := err.(*common.AppError); ok {
			c.JSON(appError.StatusCode, common.NewError(appError.StatusCode, appError.Message, appError.Err))
		} else {
			c.JSON(http.StatusInternalServerError, common.NewError(http.StatusInternalServerError, "서버 에러", err))
		}
		return
	}

	// 회사 사용자 조회 성공 응답

	c.JSON(http.StatusOK, common.NewResponse(http.StatusOK, "회사 사용자 조회 성공", response))
}

// TODO 해당 부서에 속한 사용자 리스트 가져오기 (이후 디테일 잡을때)
func (h *UserHandler) GetUsersByDepartment(c *gin.Context) {
	departmentId := c.Param("departmentid")

	departmentIdUint, err := strconv.ParseUint(departmentId, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "유효하지 않은 부서 ID입니다", err))
		return
	}

	users, err := h.userUsecase.GetUsersByDepartment(uint(departmentIdUint))
	if err != nil {
		if appError, ok := err.(*common.AppError); ok {
			c.JSON(appError.StatusCode, common.NewError(appError.StatusCode, appError.Message, appError.Err))
		} else {
			c.JSON(http.StatusInternalServerError, common.NewError(http.StatusInternalServerError, "서버 에러", err))
		}
		return
	}

	c.JSON(http.StatusOK, common.NewResponse(http.StatusOK, "부서 사용자 조회 성공", users))
}
