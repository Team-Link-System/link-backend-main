package http

import (
	"link/internal/department/entity"
	"link/internal/department/usecase"
	"link/pkg/dto/department/req"
	"link/pkg/interceptor"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type DepartmentHandler struct {
	departmentUsecase usecase.DepartmentUsecase
}

func NewDepartmentHandler(departmentUsecase usecase.DepartmentUsecase) *DepartmentHandler {
	return &DepartmentHandler{departmentUsecase: departmentUsecase}
}

func (h *DepartmentHandler) CreateDepartment(c *gin.Context) {

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

	var request req.CreateDepartmentRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, interceptor.Error(http.StatusBadRequest, "잘못된 요청입니다."))
		return
	}

	var managerID *uint
	if request.ManagerID != 0 {
		managerID = &request.ManagerID
	}
	department := &entity.Department{
		Name:      request.Name,
		ManagerID: managerID,
	}

	createdDepartment, err := h.departmentUsecase.CreateDepartment(department, requestUserId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, interceptor.Error(http.StatusInternalServerError, err.Error()))
		return
	}

	c.JSON(http.StatusCreated, interceptor.Created("부서 생성 성공", createdDepartment))
}

// TODO 부서 목록 리스트
func (h *DepartmentHandler) GetDepartments(c *gin.Context) {
	departments, err := h.departmentUsecase.GetDepartments()
	if err != nil {
		c.JSON(http.StatusInternalServerError, interceptor.Error(http.StatusInternalServerError, err.Error()))
		return
	}
	c.JSON(http.StatusOK, interceptor.Success("부서 목록 조회 성공", departments))
}

// TODO 부서 상세 조회
func (h *DepartmentHandler) GetDepartment(c *gin.Context) {
	departmentID := c.Param("id")

	targetDepartmentID, err := strconv.ParseUint(departmentID, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, interceptor.Error(http.StatusBadRequest, "유효하지 않은 부서 ID입니다"))
		return
	}

	department, err := h.departmentUsecase.GetDepartment(uint(targetDepartmentID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, interceptor.Error(http.StatusInternalServerError, err.Error()))
		return
	}

	c.JSON(http.StatusOK, interceptor.Success("부서 상세 조회 성공", department))
}

// TODO 부서 수정 ( 관리자 )
func (h *DepartmentHandler) UpdateDepartment(c *gin.Context) {
	departmentID := c.Param("id")
	targetDepartmentID, err := strconv.ParseUint(departmentID, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, interceptor.Error(http.StatusBadRequest, "유효하지 않은 부서 ID입니다"))
		return
	}

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

	var request req.UpdateDepartmentRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, interceptor.Error(http.StatusBadRequest, "잘못된 요청입니다."))
		return
	}

	department := &entity.Department{
		Name:      request.Name,
		ManagerID: &request.ManagerID,
	}

	updatedDepartment, err := h.departmentUsecase.UpdateDepartment(uint(targetDepartmentID), department, requestUserId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, interceptor.Error(http.StatusInternalServerError, err.Error()))
		return
	}

	c.JSON(http.StatusOK, interceptor.Success("부서 수정 성공", updatedDepartment))
}

// TODO 부서 삭제 ( 관리자만 )
func (h *DepartmentHandler) DeleteDepartment(c *gin.Context) {
	departmentID := c.Param("id")
	targetDepartmentID, err := strconv.ParseUint(departmentID, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, interceptor.Error(http.StatusBadRequest, "유효하지 않은 부서 ID입니다"))
		return
	}

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

	err = h.departmentUsecase.DeleteDepartment(uint(targetDepartmentID), requestUserId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, interceptor.Error(http.StatusInternalServerError, err.Error()))
		return
	}

	c.JSON(http.StatusOK, interceptor.Success("부서 삭제 성공", nil))
}

//TODO 부서 수정 요청 저장 - 부서장만 요청 가능 - 내용은 바디에 있음

//TODO 부서 수정 요청 조회해서 수락 및 거절 ( 관리자만 )
