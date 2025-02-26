package usecase

import (
	"fmt"
	"link/pkg/common"
	"link/pkg/dto/res"
	"net/http"
	"strconv"
	"strings"
	"time"

	_postRepo "link/internal/post/repository"
	_statRepo "link/internal/stat/repository"
	_userRepo "link/internal/user/repository"
)

type StatUsecase interface {
	//user관련
	GetCurrentOnlineUsers(requestUserId uint) (*res.GetCurrentOnlineUsersResponse, error)

	GetTodayPostStat(companyId uint) (*res.GetTodayPostStatResponse, error)

	GetPopularPostStat(companyId uint, period string, visibility string) (*res.GetPopularPostStatResponse, error)
}

type statUsecase struct {
	userRepo _userRepo.UserRepository
	postRepo _postRepo.PostRepository
	statRepo _statRepo.StatRepository
}

func NewStatUsecase(
	userRepo _userRepo.UserRepository,
	postRepo _postRepo.PostRepository,
	statRepo _statRepo.StatRepository,
) StatUsecase {
	return &statUsecase{userRepo: userRepo, postRepo: postRepo, statRepo: statRepo}
}

// TODO 현재 접속중인 사용자 수
func (u *statUsecase) GetCurrentOnlineUsers(requestUserId uint) (*res.GetCurrentOnlineUsersResponse, error) {

	user, err := u.userRepo.GetUserByID(requestUserId)
	if err != nil {
		fmt.Printf("사용자 조회에 실패했습니다: %v", err)
		return nil, common.NewError(http.StatusBadRequest, "사용자 조회에 실패했습니다", err)
	}

	//TODO 회사 사용자 수 조회
	userIds, err := u.userRepo.GetUsersIdsByCompany(*user.UserProfile.CompanyID)
	if err != nil {
		fmt.Printf("회사 사용자 조회에 실패했습니다: %v", err)
		return nil, common.NewError(http.StatusNotFound, "회사 사용자 조회에 실패했습니다", err)
	}

	onlineStatusMap, err := u.userRepo.GetCacheUsers(userIds, []string{"is_online"})
	if err != nil {
		fmt.Printf("온라인 상태 조회에 실패했습니다: %v", err)
		return nil, common.NewError(http.StatusNotFound, "온라인 상태 조회에 실패했습니다", err)
	}

	onlineCount := 0
	for _, user := range onlineStatusMap {
		if status, exists := user["is_online"]; exists {
			if strStatus, ok := status.(string); ok {
				if boolStatus, err := strconv.ParseBool(strStatus); err == nil && boolStatus {
					onlineCount++
				}
			}
		}
	}

	return &res.GetCurrentOnlineUsersResponse{
		OnlineUsers:      onlineCount,
		TotalCompanyUser: len(userIds),
	}, nil
}

func (uc *statUsecase) GetTodayPostStat(requestUserId uint) (*res.GetTodayPostStatResponse, error) {
	user, err := uc.userRepo.GetUserByID(requestUserId)
	if err != nil {
		fmt.Printf("사용자 조회 실패: %v", err)
		return nil, common.NewError(http.StatusBadRequest, "사용자가 없습니다", err)
	}

	if user.UserProfile.CompanyID == nil {
		fmt.Printf("사용자의 회사 정보가 없습니다")
		return nil, common.NewError(http.StatusBadRequest, "사용자의 회사 정보가 없습니다", nil)
	}

	//TODO post 집계 데이터 조회
	postsStat, err := uc.statRepo.GetTodayPostStat(*user.UserProfile.CompanyID)
	if err != nil {
		fmt.Printf("게시물 집계 데이터 조회 실패: %v", err)
		return nil, common.NewError(http.StatusBadRequest, "게시물 집계 데이터 조회 실패", err)
	}

	// Response 구조체에 데이터 매핑
	response := &res.GetTodayPostStatResponse{
		TotalCompanyPostCount:    postsStat.TotalCompanyPostCount,
		TotalDepartmentPostCount: postsStat.TotalDepartmentPostCount,
		DepartmentPost:           []res.DepartmentPostStat{},
	}

	for _, department := range postsStat.DepartmentStats {
		response.DepartmentPost = append(response.DepartmentPost, res.DepartmentPostStat{
			DepartmentId:   department.DepartmentId,
			DepartmentName: department.DepartmentName,
			PostCount:      department.PostCount,
		})
	}

	return response, nil
}

func (uc *statUsecase) GetPopularPostStat(requestUserId uint, period string, visibility string) (*res.GetPopularPostStatResponse, error) {
	user, err := uc.userRepo.GetUserByID(requestUserId)
	if err != nil {
		fmt.Printf("사용자 조회 실패: %v", err)
		return nil, common.NewError(http.StatusBadRequest, "사용자가 없습니다", err)
	}

	if strings.ToLower(visibility) == "public" {
		visibility = "public"
	} else if strings.ToLower(visibility) == "company" {
		if user.UserProfile.CompanyID == nil {
			fmt.Printf("사용자의 회사 정보가 없습니다")
			return nil, common.NewError(http.StatusBadRequest, "사용자의 회사 정보가 없습니다", nil)
		}
	} else if strings.ToLower(visibility) == "department" {
		if len(user.UserProfile.Departments) == 0 {
			fmt.Printf("사용자의 부서 정보가 없습니다")
			return nil, common.NewError(http.StatusBadRequest, "사용자의 부서 정보가 없습니다", nil)
		}
	}

	// //TODO post 집계 데이터 조회
	postsStat, err := uc.statRepo.GetPopularPost(visibility, period)
	if err != nil {
		fmt.Printf("게시물 집계 데이터 조회 실패: %v", err)
		return nil, common.NewError(http.StatusBadRequest, err.Error(), err)
	}

	response := &res.GetPopularPostStatResponse{
		Period:     period,
		Visibility: visibility,
		CreatedAt:  postsStat.CreatedAt,
		Posts:      []res.PostPayload{},
	}

	for _, post := range postsStat.Posts {
		response.Posts = append(response.Posts, res.PostPayload{
			Rank:   post.Rank,
			PostId: post.PostId,
			UserId: post.UserId,
			Title:  post.Title,
			// Content:       post.Content,
			IsAnonymous:   post.IsAnonymous,
			Visibility:    post.Visibility,
			CreatedAt:     post.CreatedAt.Format(time.DateTime),
			UpdatedAt:     post.UpdatedAt.Format(time.DateTime),
			TotalViews:    post.TotalViews,
			TotalLikes:    post.TotalLikes,
			TotalComments: post.TotalComments,
			Score:         post.Score,
		})
	}

	return response, nil
}
