package usecase

import (
	"fmt"
	"link/internal/comment/entity"
	_commentRepo "link/internal/comment/repository"
	_departmentRepo "link/internal/department/repository"
	_postRepo "link/internal/post/repository"
	_userRepo "link/internal/user/repository"
	"link/pkg/common"
	"link/pkg/dto/req"
	"link/pkg/dto/res"
	_util "link/pkg/util"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type CommentUsecase interface {
	CreateComment(userId uint, req req.CommentRequest) error
	CreateReply(userId uint, req req.ReplyRequest) error
	GetComments(userId uint, queryParams req.GetCommentQueryParams) (*res.GetCommentsResponse, error)
	GetReplies(userId uint, queryParams req.GetReplyQueryParams) (*res.GetRepliesResponse, error)
	DeleteComment(userId uint, commentId uint) error
	UpdateComment(userId uint, commentId uint, req req.CommentUpdateRequest) error
}

type commentUsecase struct {
	commentRepo    _commentRepo.CommentRepository
	userRepo       _userRepo.UserRepository
	departmentRepo _departmentRepo.DepartmentRepository
	postRepo       _postRepo.PostRepository
}

func NewCommentUsecase(
	commentRepo _commentRepo.CommentRepository,
	userRepo _userRepo.UserRepository,
	departmentRepo _departmentRepo.DepartmentRepository,
	postRepo _postRepo.PostRepository,
) CommentUsecase {
	return &commentUsecase{
		commentRepo:    commentRepo,
		userRepo:       userRepo,
		postRepo:       postRepo,
		departmentRepo: departmentRepo,
	}
}

// TODO 댓글 생성
func (u *commentUsecase) CreateComment(userId uint, req req.CommentRequest) error {
	user, err := u.userRepo.GetUserByID(userId)
	if err != nil {
		fmt.Printf("사용자 조회 실패: %v", err)
		return common.NewError(http.StatusBadRequest, "사용자 조회 실패", err)
	}

	post, err := u.postRepo.GetPostByID(req.PostID)
	if err != nil {
		fmt.Printf("게시물 조회 실패: %v", err)
		return common.NewError(http.StatusBadRequest, "게시물 조회 실패", err)
	}

	if strings.ToUpper(post.Visibility) == "COMPANY" {
		if post.CompanyID == nil {
			fmt.Printf("해당 게시글은 회사 게시물이 아닙니다.")
			return common.NewError(http.StatusBadRequest, "해당 게시글은 회사 게시물이 아닙니다.", nil)
		}
		if user.UserProfile == nil || user.UserProfile.CompanyID == nil {
			fmt.Printf("사용자의 회사 정보가 없습니다.")
			return common.NewError(http.StatusBadRequest, "사용자의 회사 정보가 없습니다.", nil)
		}
		if *post.CompanyID != *user.UserProfile.CompanyID {
			fmt.Printf("회사 게시물에 대한 접근 권한이 없습니다.")
			return common.NewError(http.StatusForbidden, "회사 게시물에 대한 접근 권한이 없습니다.", nil)
		}

	} else if strings.ToUpper(post.Visibility) == "DEPARTMENT" {
		if post.CompanyID == nil {
			fmt.Printf("해당 게시글은 회사 게시물이 아닙니다.")
			return common.NewError(http.StatusBadRequest, "해당 게시글은 회사 게시물이 아닙니다.", nil)
		}
		if user.UserProfile == nil || user.UserProfile.CompanyID == nil {
			fmt.Printf("사용자의 회사 정보가 없습니다.")
			return common.NewError(http.StatusBadRequest, "사용자의 회사 정보가 없습니다.", nil)
		}

		if *post.CompanyID != *user.UserProfile.CompanyID {
			fmt.Printf("해당 회사의 부서 게시물에 대한 접근 권한이 없습니다.")
			return common.NewError(http.StatusForbidden, "해당 회사의 부서 게시물에 대한 접근 권한이 없습니다.", nil)
		}

		//TODO 해당 게시물이 속한 부서에 사용자가 속해있는지 확인
		if user.UserProfile.Departments == nil {
			fmt.Printf("해당 부서 게시물에 대한 접근 권한이 없습니다.")
			return common.NewError(http.StatusForbidden, "해당 부서 게시물에 대한 접근 권한이 없습니다.", nil)
		}

		//TODO post의 departments id 리스트에 해당 사용자의 부서 ids 리스트 중 속해있는지 확인
		userDeptIds := make(map[uint]struct{})
		for _, dept := range user.UserProfile.Departments {
			userDeptIds[(*dept)["id"].(uint)] = struct{}{}
		}

		hasAccess := false
		for _, dept := range *post.Departments {
			deptMap := dept.(map[string]interface{})
			if _, ok := userDeptIds[deptMap["id"].(uint)]; ok {
				hasAccess = true
				break
			}
		}

		if !hasAccess {
			fmt.Printf("부서 게시물에 대한 접근 권한이 없습니다.")
			return common.NewError(http.StatusForbidden, "부서 게시물에 대한 접근 권한이 없습니다.", nil)
		}

		if req.IsAnonymous != nil && *req.IsAnonymous {
			fmt.Printf("익명 댓글은 부서 게시물에 작성할 수 없습니다.")
			return common.NewError(http.StatusBadRequest, "익명 댓글은 부서 게시물에 작성할 수 없습니다.", nil)
		}
	}

	comment := &entity.Comment{
		PostID:      req.PostID,
		ParentID:    nil,
		UserID:      *user.ID,
		Content:     req.Content,
		IsAnonymous: req.IsAnonymous,
	}

	err = u.commentRepo.CreateComment(comment)
	if err != nil {
		fmt.Printf("댓글 생성 실패: %v", err)
		return common.NewError(http.StatusBadRequest, "댓글 생성 실패", err)
	}

	return nil
}

// TODO 해당 댓글에 대댓글 생성
func (u *commentUsecase) CreateReply(userId uint, req req.ReplyRequest) error {
	user, err := u.userRepo.GetUserByID(userId)
	if err != nil {
		fmt.Printf("사용자 조회 실패: %v", err)
		return common.NewError(http.StatusBadRequest, "사용자 조회 실패", err)
	}

	post, err := u.postRepo.GetPostByID(req.PostID)
	if err != nil {
		fmt.Printf("게시물 조회 실패: %v", err)
		return common.NewError(http.StatusBadRequest, "게시물 조회 실패", err)
	}

	//TODO 댓글 있는지 확인
	comment, err := u.commentRepo.GetCommentByID(req.ParentID)
	if err != nil {
		fmt.Printf("댓글 조회 실패: %v", err)
		return common.NewError(http.StatusBadRequest, "댓글 조회 실패", err)
	}

	//TODO 해당댓글이 해당 게시물의 댓글인지 확인
	if comment.PostID != post.ID {
		fmt.Printf("해당 게시물의 댓글이 아닙니다.")
		return common.NewError(http.StatusBadRequest, "해당 게시물의 댓글이 아닙니다.", nil)
	}

	if comment.ParentID != nil {
		fmt.Printf("대댓글에 댓글을 생성할 수 없습니다.")
		return common.NewError(http.StatusBadRequest, "대댓글에 댓글을 생성할 수 없습니다.", nil)
	}

	if strings.ToUpper(post.Visibility) == "COMPANY" && *post.CompanyID != *user.UserProfile.CompanyID {
		fmt.Printf("회사 게시물에 대한 접근 권한이 없습니다.")
		return common.NewError(http.StatusForbidden, "회사 게시물에 대한 접근 권한이 없습니다.", nil)
	} else if strings.ToUpper(post.Visibility) == "DEPARTMENT" {
		if *post.CompanyID != *user.UserProfile.CompanyID {
			fmt.Printf("해당 회사의 부서 게시물에 대한 접근 권한이 없습니다.")
			return common.NewError(http.StatusForbidden, "해당 회사의 부서 게시물에 대한 접근 권한이 없습니다.", nil)
		}

		//TODO 해당 게시물이 속한 부서에 사용자가 속해있는지 확인
		if user.UserProfile.Departments == nil {
			fmt.Printf("해당 부서 게시물에 대한 접근 권한이 없습니다.")
			return common.NewError(http.StatusForbidden, "해당 부서 게시물에 대한 접근 권한이 없습니다.", nil)
		}

		//TODO post의 departments id 리스트에 해당 사용자의 부서 ids 리스트 중 속해있는지 확인
		userDeptIds := make(map[uint]struct{})
		for _, dept := range user.UserProfile.Departments {
			userDeptIds[(*dept)["id"].(uint)] = struct{}{}
		}

		hasAccess := false
		for _, dept := range *post.Departments {
			deptMap := dept.(map[string]interface{})
			if _, ok := userDeptIds[deptMap["id"].(uint)]; ok {
				hasAccess = true
				break
			}
		}

		if !hasAccess {
			fmt.Printf("부서 게시물에 대한 접근 권한이 없습니다.")
			return common.NewError(http.StatusForbidden, "부서 게시물에 대한 접근 권한이 없습니다.", nil)
		}

		if req.IsAnonymous != nil && *req.IsAnonymous {
			fmt.Printf("익명 대댓글은 부서 게시물에 작성할 수 없습니다.")
			return common.NewError(http.StatusBadRequest, "익명 대댓글은 부서 게시물에 작성할 수 없습니다.", nil)
		}
	}

	reply := &entity.Comment{
		PostID:      req.PostID,
		ParentID:    &req.ParentID,
		UserID:      *user.ID,
		Content:     req.Content,
		IsAnonymous: req.IsAnonymous,
	}

	err = u.commentRepo.CreateComment(reply)
	if err != nil {
		fmt.Printf("대댓글 생성 실패: %v", err)
		return common.NewError(http.StatusBadRequest, "대댓글 생성 실패", err)
	}

	return nil
}

// TODO 해당 게시물 댓글 리스트 조회 - 커서기반 무한스크롤 정렬은 좋아요 갯수순, 날짜순 둘 중 하나가 가능해야함
func (u *commentUsecase) GetComments(userId uint, queryParams req.GetCommentQueryParams) (*res.GetCommentsResponse, error) {

	user, err := u.userRepo.GetUserByID(userId)
	if err != nil {
		fmt.Printf("사용자 조회 실패: %v", err)
		return nil, common.NewError(http.StatusBadRequest, "사용자 조회 실패", err)
	}

	post, err := u.postRepo.GetPostByID(queryParams.PostID)
	if err != nil {
		fmt.Printf("없는 게시물입니다: %v", err)
		return nil, common.NewError(http.StatusBadRequest, "없는 게시물입니다.", err)
	}

	//TODO 해당 게시글에 접근 권한이 있는지 확인해야함
	if strings.ToUpper(post.Visibility) == "COMPANY" && *post.CompanyID != *user.UserProfile.CompanyID {
		fmt.Printf("회사 게시물에 대한 접근 권한이 없습니다.")
		return nil, common.NewError(http.StatusForbidden, "회사 게시물에 대한 접근 권한이 없습니다.", nil)
	} else if strings.ToUpper(post.Visibility) == "DEPARTMENT" {
		//TODO 해당 게시물이 속한 부서에 사용자가 속해있는지 확인
		if user.UserProfile.Departments == nil {
			fmt.Printf("해당 부서 게시물에 대한 접근 권한이 없습니다.")
			return nil, common.NewError(http.StatusForbidden, "해당 부서 게시물에 대한 접근 권한이 없습니다.", nil)
		}

		userDeptIds := make(map[uint]struct{})
		for _, dept := range user.UserProfile.Departments {
			userDeptIds[(*dept)["id"].(uint)] = struct{}{}
		}

		hasAccess := false
		for _, dept := range *post.Departments {
			deptMap := dept.(map[string]interface{})
			if _, ok := userDeptIds[deptMap["id"].(uint)]; ok {
				hasAccess = true
				break
			}
		}

		if !hasAccess {
			fmt.Printf("부서 게시물에 대한 접근 권한이 없습니다.")
			return nil, common.NewError(http.StatusForbidden, "부서 게시물에 대한 접근 권한이 없습니다.", nil)
		}
	}

	queryOptions := map[string]interface{}{
		"page":    queryParams.Page,
		"limit":   queryParams.Limit,
		"sort":    queryParams.Sort,
		"order":   queryParams.Order,
		"post_id": queryParams.PostID,
		"cursor":  map[string]interface{}{},
	}

	if queryParams.Cursor != nil {
		if queryParams.Cursor.CreatedAt != "" {
			queryOptions["cursor"].(map[string]interface{})["created_at"] = queryParams.Cursor.CreatedAt
		} else if queryParams.Cursor.ID != "" {
			queryOptions["cursor"].(map[string]interface{})["id"] = queryParams.Cursor.ID
		} else if queryParams.Cursor.LikeCount != "" {
			queryOptions["cursor"].(map[string]interface{})["like_count"] = queryParams.Cursor.LikeCount
		}
	}

	meta, comments, err := u.commentRepo.GetCommentsByPostID(queryParams.PostID, queryOptions)
	if err != nil {
		fmt.Printf("댓글 조회 실패: %v", err)
		return nil, common.NewError(http.StatusBadRequest, "댓글 조회 실패", err)
	}

	var nextCursor string
	if len(comments) > 0 {
		lastComment := comments[len(comments)-1]
		if queryParams.Sort == "created_at" {
			nextCursor = _util.ParseKst(lastComment.CreatedAt).Format(time.DateTime)
		} else if queryParams.Sort == "id" {
			nextCursor = strconv.Itoa(int(lastComment.ID))
		}
	}

	commentRes := make([]*res.CommentResponse, len(comments))
	for i, comment := range comments {

		userName := "익명"
		var profileImage string
		if !*comment.IsAnonymous {
			userName = comment.UserName
			profileImage = comment.ProfileImage
		}

		commentRes[i] = &res.CommentResponse{
			CommentId:    comment.ID,
			UserId:       comment.UserID,
			UserName:     userName,
			ProfileImage: profileImage,
			Content:      comment.Content,
			IsAnonymous:  *comment.IsAnonymous,
			LikeCount:    comment.LikeCount,
			ReplyCount:   comment.ReplyCount,
			CreatedAt:    _util.ParseKst(comment.CreatedAt).Format(time.DateTime),
		}
	}

	return &res.GetCommentsResponse{
		Comments: commentRes,
		Meta: &res.CommentMeta{
			NextCursor: nextCursor,
			HasMore:    meta.HasMore,
			TotalCount: meta.TotalCount,
			PageSize:   meta.PageSize,
			PrevPage:   meta.PrevPage,
			NextPage:   meta.NextPage,
		},
	}, nil
}

// TODO 해당 댓글에 대한 대댓글 리스트 조회 - 얘는 오프셋 x 무조건 날짜 기반
func (u *commentUsecase) GetReplies(userId uint, queryParams req.GetReplyQueryParams) (*res.GetRepliesResponse, error) {
	user, err := u.userRepo.GetUserByID(userId)
	if err != nil {
		fmt.Printf("사용자 조회 실패: %v", err)
		return nil, common.NewError(http.StatusBadRequest, "사용자 조회 실패", err)
	}

	post, err := u.postRepo.GetPostByID(queryParams.PostID)
	if err != nil {
		fmt.Printf("없는 게시물입니다: %v", err)
		return nil, common.NewError(http.StatusBadRequest, "없는 게시물입니다.", err)
	}

	comment, err := u.commentRepo.GetCommentByID(queryParams.ParentID)
	if err != nil {
		fmt.Printf("없는 댓글입니다: %v", err)
		return nil, common.NewError(http.StatusBadRequest, "없는 댓글입니다.", err)
	}

	//TODO 해당 게시물에 접근권한 확인
	if strings.ToUpper(post.Visibility) == "COMPANY" && *post.CompanyID != *user.UserProfile.CompanyID {
		fmt.Printf("회사 게시물에 대한 접근 권한이 없습니다.")
		return nil, common.NewError(http.StatusForbidden, "회사 게시물에 대한 접근 권한이 없습니다.", nil)
	} else if strings.ToUpper(post.Visibility) == "DEPARTMENT" {
		//TODO 해당 게시물이 속한 부서에 사용자가 속해있는지 확인
		if user.UserProfile.Departments == nil {
			fmt.Printf("해당 부서 게시물에 대한 접근 권한이 없습니다.")
			return nil, common.NewError(http.StatusForbidden, "해당 부서 게시물에 대한 접근 권한이 없습니다.", nil)
		}

		userDeptIds := make(map[uint]struct{})
		for _, dept := range user.UserProfile.Departments {
			userDeptIds[(*dept)["id"].(uint)] = struct{}{}
		}

		hasAccess := false
		for _, dept := range *post.Departments {
			deptMap := dept.(map[string]interface{})
			if _, ok := userDeptIds[deptMap["id"].(uint)]; ok {
				hasAccess = true
				break
			}
		}

		if !hasAccess {
			fmt.Printf("부서 게시물에 대한 접근 권한이 없습니다.")
			return nil, common.NewError(http.StatusForbidden, "부서 게시물에 대한 접근 권한이 없습니다.", nil)
		}
	}

	queryOptions := map[string]interface{}{
		"page":      queryParams.Page,
		"limit":     queryParams.Limit,
		"sort":      queryParams.Sort,
		"order":     queryParams.Order,
		"post_id":   queryParams.PostID,
		"parent_id": comment.ID,
		"cursor":    map[string]interface{}{},
	}

	if queryParams.Cursor != nil {
		if queryParams.Cursor.CreatedAt != "" {
			queryOptions["cursor"].(map[string]interface{})["created_at"] = queryParams.Cursor.CreatedAt
		} else if queryParams.Cursor.ID != "" {
			queryOptions["cursor"].(map[string]interface{})["id"] = queryParams.Cursor.ID
		} else if queryParams.Cursor.LikeCount != "" {
			queryOptions["cursor"].(map[string]interface{})["like_count"] = queryParams.Cursor.LikeCount
		}
	}

	meta, replies, err := u.commentRepo.GetRepliesByParentID(queryParams.ParentID, queryOptions)
	if err != nil {
		fmt.Printf("대댓글 조회 실패: %v", err)
		return nil, common.NewError(http.StatusBadRequest, "대댓글 조회 실패", err)
	}

	var nextCursor string
	if len(replies) > 0 {
		lastReply := replies[len(replies)-1]
		if queryParams.Sort == "created_at" {
			nextCursor = _util.ParseKst(lastReply.CreatedAt).Format(time.DateTime)
		} else if queryParams.Sort == "id" {
			nextCursor = strconv.Itoa(int(lastReply.ID))
		}
	}

	replyRes := make([]*res.ReplyResponse, len(replies))
	for i, reply := range replies {

		userName := "익명"
		var profileImage string
		if !*reply.IsAnonymous {
			userName = reply.UserName
			profileImage = reply.ProfileImage
		}

		parentId := uint(0)
		if reply.ParentID != nil {
			parentId = *reply.ParentID
		}

		replyRes[i] = &res.ReplyResponse{
			CommentId:    reply.ID,
			UserId:       reply.UserID,
			UserName:     userName,
			ProfileImage: profileImage,
			ParentID:     parentId,
			Content:      reply.Content,
			LikeCount:    reply.LikeCount,
			IsAnonymous:  *reply.IsAnonymous,
			CreatedAt:    _util.ParseKst(reply.CreatedAt).Format(time.DateTime),
		}
	}

	return &res.GetRepliesResponse{
		Replies: replyRes,
		Meta: &res.CommentMeta{
			NextCursor: nextCursor,
			HasMore:    meta.HasMore,
			TotalCount: meta.TotalCount,
			PageSize:   meta.PageSize,
			PrevPage:   meta.PrevPage,
			NextPage:   meta.NextPage,
		},
	}, nil
}

// TODO 해당 댓글 삭제(이건 댓글 id 받아서 그냥 삭제) - 댓글, 대댓글
func (u *commentUsecase) DeleteComment(userId uint, commentId uint) error {
	user, err := u.userRepo.GetUserByID(userId)
	if err != nil {
		fmt.Printf("사용자 조회 실패: %v", err)
		return common.NewError(http.StatusBadRequest, "사용자 조회 실패", err)
	}

	comment, err := u.commentRepo.GetCommentByID(commentId)
	if err != nil {
		fmt.Printf("댓글 조회 실패: %v", err)
		return common.NewError(http.StatusBadRequest, "댓글 조회 실패", err)
	}

	//TODO 해당 댓글의 주인이 맞는지 확인
	if comment.UserID != *user.ID {
		fmt.Printf("해당 댓글의 주인이 아닙니다.")
		return common.NewError(http.StatusForbidden, "해당 댓글의 주인이 아닙니다.", nil)
	}

	err = u.commentRepo.DeleteComment(commentId)
	if err != nil {
		fmt.Printf("댓글 삭제 실패: %v", err)
		return common.NewError(http.StatusBadRequest, "댓글 삭제 실패", err)
	}

	return nil
}

// TODO 댓글 수정 (댓글 id 받아서 수정) parentId는 상관없이 내용만 수정
func (u *commentUsecase) UpdateComment(userId uint, commentId uint, request req.CommentUpdateRequest) error {
	user, err := u.userRepo.GetUserByID(userId)
	if err != nil {
		fmt.Printf("사용자 조회 실패: %v", err)
		return common.NewError(http.StatusBadRequest, "사용자 조회 실패", err)
	}

	comment, err := u.commentRepo.GetCommentByID(commentId)
	if err != nil {
		fmt.Printf("댓글 조회 실패: %v", err)
		return common.NewError(http.StatusBadRequest, "댓글 조회 실패", err)
	}

	//TODO 해당 댓글의 주인이 맞는지 확인
	if comment.UserID != *user.ID {
		fmt.Printf("해당 댓글의 주인이 아닙니다.")
		return common.NewError(http.StatusForbidden, "해당 댓글의 주인이 아닙니다.", nil)
	}

	//TODO isAnonymous는 public , company 게시물에만가능
	post, err := u.postRepo.GetPostByCommentID(commentId)
	if err != nil {
		fmt.Printf("게시물 조회 실패: %v", err)
		return common.NewError(http.StatusBadRequest, "게시물 조회 실패", err)
	}

	if strings.ToUpper(post.Visibility) != "PUBLIC" && strings.ToUpper(post.Visibility) != "COMPANY" {
		if *request.IsAnonymous {
			fmt.Printf("익명전환은 public , company 게시물에만 가능합니다.")
			return common.NewError(http.StatusForbidden, "익명전환은 public , company 게시물에만 가능합니다.", nil)
		}
	}

	updateComment := map[string]interface{}{}

	if request.Content != "" {
		updateComment["content"] = request.Content
	}
	if request.IsAnonymous != nil {
		updateComment["is_anonymous"] = *request.IsAnonymous
	}

	err = u.commentRepo.UpdateComment(commentId, updateComment)
	if err != nil {
		fmt.Printf("댓글 수정 실패: %v", err)
		return common.NewError(http.StatusBadRequest, "댓글 수정 실패", err)
	}

	return nil
}
