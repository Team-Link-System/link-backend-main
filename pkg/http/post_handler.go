package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

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
	if request.Title == "" {
		fmt.Printf("제목이 없습니다.")
		c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "제목이 없습니다.", nil))
		return
	} else if request.Content == "" {
		fmt.Printf("내용이 없습니다.")
		c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "내용이 없습니다.", nil))
		return
	}

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
	// 인증된 사용자 확인
	userId, exists := c.Get("userId")
	if !exists {
		fmt.Printf("인증되지 않은 사용자입니다.")
		c.JSON(http.StatusUnauthorized, common.NewError(http.StatusUnauthorized, "인증되지 않은 사용자입니다.", nil))
		return
	}

	// 게시물 조회 파라미터 처리
	category := strings.ToUpper(c.DefaultQuery("category", "PUBLIC"))
	if category != "PUBLIC" && category != "COMPANY" && category != "DEPARTMENT" {
		category = "PUBLIC"
	}

	viewType := strings.ToUpper(c.DefaultQuery("view_type", "INFINITE"))
	if viewType != "INFINITE" && viewType != "PAGINATION" {
		viewType = "INFINITE"
	}

	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if err != nil || limit < 1 || limit > 100 {
		limit = 10
	}

	sort := c.DefaultQuery("sort", "created_at")
	if sort != "created_at" && sort != "like_count" && sort != "comments_count" && sort != "id" {
		sort = "created_at"
	}

	order := c.DefaultQuery("order", "desc")
	if order != "asc" && order != "desc" {
		order = "desc"
	}

	cursorParam := c.Query("cursor")
	var cursor *req.Cursor

	if viewType == "INFINITE" && cursorParam == "" {
		cursor = nil //첫요청
		page = 1
	} else if viewType == "INFINITE" {
		var tempCursor req.Cursor
		if err := json.Unmarshal([]byte(cursorParam), &tempCursor); err != nil {
			fmt.Printf("커서 파싱 실패: %v", err)
			c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "유효하지 않은 커서 값입니다.", err))
			return
		}
		//TODO kst로 받은걸 utc로 변환
		//            "next_cursor": "2024-11-20 11:36:59",

		if sort == "created_at" && tempCursor.CreatedAt == "" {
			fmt.Printf("커서는 sort와 같은 값이 있어야 합니다.")
			c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "커서는 sort와 같은 값이 있어야 합니다.", nil))
			return
		} else if sort == "like_count" && tempCursor.LikeCount == "" {
			fmt.Printf("커서는 sort와 같은 값이 있어야 합니다.")
			c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "커서는 sort와 같은 값이 있어야 합니다.", nil))
			return
		} else if sort == "comments_count" && tempCursor.CommentsCount == "" {
			fmt.Printf("커서는 sort와 같은 값이 있어야 합니다.")
			c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "커서는 sort와 같은 값이 있어야 합니다.", nil))
			return
		} else if sort == "id" && tempCursor.ID == "" {
			fmt.Printf("커서는 sort와 같은 값이 있어야 합니다.")
			c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "커서는 sort와 같은 값이 있어야 합니다.", nil))
			return
		}

		cursor = &tempCursor

	} else if viewType == "PAGINATION" && cursorParam != "" {
		fmt.Printf("페이지네이션 타입인데 커서가 있습니다.")
		c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "페이지네이션 타입인데 커서가 있습니다.", nil))
		return
	}

	var companyId, departmentId uint
	if category == "COMPANY" {
		companyIdValue, _ := strconv.ParseUint(c.DefaultQuery("company_id", "0"), 10, 32)
		companyId = uint(companyIdValue)
		if companyId == 0 {
			fmt.Printf("회사 게시물 조회 시 company_id가 필요합니다.")
			c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "회사 게시물 조회 시 company_id가 필요합니다.", nil))
			return
		}
	} else if category == "DEPARTMENT" {
		departmentIdValue, _ := strconv.ParseUint(c.DefaultQuery("department_id", "0"), 10, 32)
		departmentId = uint(departmentIdValue)
		if departmentId == 0 {
			fmt.Printf("부서 게시물 조회 시 department_id가 필요합니다.")
			c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "부서 게시물 조회 시 department_id가 필요합니다.", nil))
			return
		}
	} else if category == "PUBLIC" && (companyId != 0 || departmentId != 0) {
		fmt.Printf("PUBLIC 게시물은 company_id와 department_id가 없어야 합니다.")
		c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "PUBLIC 게시물은 company_id와 department_id가 없어야 합니다.", nil))
		return
	}

	queryParams := req.GetPostQueryParams{
		Category:     category,
		Page:         page,
		Limit:        limit,
		Order:        order,
		Sort:         sort,
		ViewType:     viewType,
		Cursor:       cursor,
		CompanyId:    companyId,
		DepartmentId: departmentId,
	}

	// 게시물 조회
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

// TODO 게시물 상세보기 - 전체 사용자
func (h *PostHandler) GetPost(c *gin.Context) {
	userId, exists := c.Get("userId")
	if !exists {
		fmt.Printf("인증되지 않은 사용자입니다.")
		c.JSON(http.StatusUnauthorized, common.NewError(http.StatusUnauthorized, "인증되지 않은 사용자입니다.", nil))
		return
	}

	postId, err := strconv.Atoi(c.Param("postid"))
	if err != nil {
		fmt.Printf("게시물 아이디 처리 실패: %v", err)
		c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "게시물 아이디 처리 실패", err))
		return
	}

	post, err := h.postUsecase.GetPost(userId.(uint), uint(postId))
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

	c.JSON(http.StatusOK, common.NewResponse(http.StatusOK, "게시물 조회 완료", post))
}

// TODO 게시물 삭제
func (h *PostHandler) DeletePost(c *gin.Context) {
	userId, exists := c.Get("userId")
	if !exists {
		fmt.Printf("인증되지 않은 사용자입니다.")
		c.JSON(http.StatusUnauthorized, common.NewError(http.StatusUnauthorized, "인증되지 않은 사용자입니다.", nil))
		return
	}

	postId, err := strconv.Atoi(c.Param("postid"))
	if err != nil {
		fmt.Printf("게시물 아이디 처리 실패: %v", err)
		c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "게시물 아이디 처리 실패", err))
		return
	}

	err = h.postUsecase.DeletePost(userId.(uint), uint(postId))
	if err != nil {
		if appError, ok := err.(*common.AppError); ok {
			c.JSON(appError.StatusCode, common.NewError(appError.StatusCode, appError.Message, appError.Err))
		} else {
			fmt.Printf("게시물 삭제 실패: %v", err)
			c.JSON(http.StatusInternalServerError, common.NewError(http.StatusInternalServerError, "서버 에러", err))
		}
		return
	}

	c.JSON(http.StatusOK, common.NewResponse(http.StatusOK, "게시물 삭제 완료", nil))
}
