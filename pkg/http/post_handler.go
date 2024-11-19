package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"link/internal/post/usecase"
	"link/pkg/common"
	"link/pkg/dto/req"
)

type PostHandler struct {
	postUsecase usecase.PostUsecase
}

func NewPostHandler(postUsecase usecase.PostUsecase) *PostHandler {
	return &PostHandler{postUsecase: postUsecase}
}

// TODO 게시물 생성 - 전체 사용자 게시물
func (h *PostHandler) CreatePost(c *gin.Context) {
	//TODO 게시물 생성
	userId, exists := c.Get("userId")
	if !exists {
		fmt.Printf("인증되지 않은 사용자입니다.")
		c.JSON(http.StatusUnauthorized, common.NewError(http.StatusUnauthorized, "인증되지 않은 사용자입니다.", nil))
		return
	}

	var request req.CreatePostRequest
	if err := c.ShouldBind(&request); err != nil {
		fmt.Printf("잘못된 요청입니다: %v", err)
		c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "잘못된 요청입니다", err))
		return
	}

	postImageUrls, exists := c.Get("post_image_urls")
	if exists {
		imageUrls, ok := postImageUrls.([]string)
		if !ok {
			fmt.Printf("이미지 처리 실패")
			c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "이미지 처리 실패", nil))
			return
		}
		if len(imageUrls) > 0 {
			ptrUrls := make([]*string, len(imageUrls))
			for i := range imageUrls {
				ptrUrls[i] = &imageUrls[i]
			}
			request.Images = ptrUrls
		}
	}

	err := h.postUsecase.CreatePost(userId.(uint), &request)
	if err != nil {
		if appError, ok := err.(*common.AppError); ok {
			c.JSON(appError.StatusCode, common.NewError(appError.StatusCode, appError.Message, appError.Err))
		} else {
			c.JSON(http.StatusInternalServerError, common.NewError(http.StatusInternalServerError, "서버 에러", err))
		}
		return
	}

	c.JSON(http.StatusOK, common.NewResponse(http.StatusOK, "게시물 생성 완료", nil))
}

// TODO 게시물 리스트 조회
func (h *PostHandler) GetPosts(c *gin.Context) {
	//TODO 게시물 리스트 조회
	userId, exists := c.Get("userId")
	if !exists {
		fmt.Printf("인증되지 않은 사용자입니다.")
		c.JSON(http.StatusUnauthorized, common.NewError(http.StatusUnauthorized, "인증되지 않은 사용자입니다.", nil))
		return
	}

	//TODO 게시물 조회 쿼리 파라미터
	category := c.DefaultQuery("category", "public")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if limit < 1 || limit > 100 {
		limit = 10
	}
	order := c.DefaultQuery("order", "desc")
	if order != "asc" && order != "desc" {
		order = "desc"
	}
	sort := c.DefaultQuery("sort", "created_at")
	if sort != "created_at" && sort != "like_count" && sort != "comments_count" {
		sort = "created_at"
	}
	cursorParam := c.DefaultQuery("cursor", "")

	var cursor *req.Cursor
	if cursorParam != "" {
		if err := json.Unmarshal([]byte(cursorParam), &cursor); err != nil {
			fmt.Printf("커서 파싱 실패: %v", err)
			c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "유효하지 않은 커서 값입니다.", err))
			return
		}
	}

	queryParams := req.GetPostQueryParams{
		Category: category,
		Page:     page,
		Limit:    limit,
		Order:    order,
		Sort:     sort,
		Cursor:   cursor,
	}

	if category == "COMPANY" || category == "DEPARTMENT" {
		companyId, _ := strconv.ParseUint(c.DefaultQuery("company_id", "0"), 10, 32)
		departmentId, _ := strconv.ParseUint(c.DefaultQuery("department_id", "0"), 10, 32)
		queryParams.CompanyId = uint(companyId)
		queryParams.DepartmentId = uint(departmentId)
	}

	posts, err := h.postUsecase.GetPosts(userId.(uint), queryParams)
	if err != nil {
		if appError, ok := err.(*common.AppError); ok {
			fmt.Printf("게시물 조회 실패: %v", err)
			c.JSON(appError.StatusCode, common.NewError(appError.StatusCode, appError.Message, appError.Err))
		} else {
			fmt.Printf("게시물 조회 실패: %v", err)
			c.JSON(http.StatusInternalServerError, common.NewError(http.StatusInternalServerError, "서버 에러", err))
		}
		return
	}

	c.JSON(http.StatusOK, common.NewResponse(http.StatusOK, "게시물 조회 완료", posts))
}
