package usecase

import (
	_companyRepository "link/internal/company/repository"
	_departmentRepository "link/internal/department/repository"
	"link/internal/post/entity"
	_postRepository "link/internal/post/repository"
	_teamRepository "link/internal/team/repository"
	_userRepository "link/internal/user/repository"

	"link/pkg/common"
	"link/pkg/dto/req"
	"link/pkg/dto/res"
	"net/http"
	"time"
)

type PostUsecase interface {
	CreatePost(requestUserId uint, post *req.CreatePostRequest) (*res.CreatePostResponse, error)
}

type postUsecase struct {
	postRepo       _postRepository.PostRepository
	userRepo       _userRepository.UserRepository
	companyRepo    _companyRepository.CompanyRepository
	departmentRepo _departmentRepository.DepartmentRepository
	teamRepo       _teamRepository.TeamRepository
}

func NewPostUsecase(
	postRepo _postRepository.PostRepository,
	userRepo _userRepository.UserRepository,
	companyRepo _companyRepository.CompanyRepository,
	departmentRepo _departmentRepository.DepartmentRepository,
	teamRepo _teamRepository.TeamRepository) PostUsecase {
	return &postUsecase{postRepo: postRepo, userRepo: userRepo, companyRepo: companyRepo, departmentRepo: departmentRepo, teamRepo: teamRepo}
}

// TODO 게시물 생성,
func (uc *postUsecase) CreatePost(requestUserId uint, post *req.CreatePostRequest) (*res.CreatePostResponse, error) {
	//TODO requestUserId가 존재하는지 조회
	user, err := uc.userRepo.GetUserByID(requestUserId)
	if err != nil {
		return nil, common.NewError(http.StatusBadRequest, "사용자가 없습니다", err)
	}

	//TODO visibility는 PUBLIC, COMPANY, DEPARTMENT, TEAM 중 하나
	//TODO 만약에 게시물이 PUBLIC이라면 회사 조회 X
	//TODO 게시물이 회사만 공개 게시물이면 회사 ID만 엔티티에 넣으면됨 부서ID, 팀ID는 null
	//TODO 게시물이 부서만 공개 게시물이면 회사 ID, 부서 ID 엔티티에 넣으면됨 팀ID는 null
	//TODO 게시물이 팀만 공개 게시물이면 회사 ID, 부서 ID, 팀 ID 엔티티에 넣으면됨
	//TODO 같은 부서에 여러 팀이 있을 수 있고 다른 부서 에 각각 팀이 있을 수도 있고 그 로직을 생각해야함
	var companyId *uint
	var departmentIds []*uint
	var teamIds []*uint

	// 가시성에 따른 설정
	switch post.Visibility {
	case "PUBLIC":
		// PUBLIC일 경우, 회사, 부서, 팀 정보 필요 없음
	case "COMPANY":
		// 회사에만 공개
		companyId = user.UserProfile.CompanyID
	case "DEPARTMENT":
		// 부서에만 공개
		companyId = user.UserProfile.CompanyID
		for _, deptId := range post.DepartmentIds {
			departmentIds = append(departmentIds, deptId)
		}
		isAnonymous := bool(false)
		post.IsAnonymous = &isAnonymous
	case "TEAM":
		// 팀에만 공개
		companyId = user.UserProfile.CompanyID
		for _, teamID := range post.TeamIds {
			teamIds = append(teamIds, teamID)
		}
		isAnonymous := bool(false)
		post.IsAnonymous = &isAnonymous
	}

	//TODO PUBLIC이나 COMPANY일 경우에만 익명	설정 가능

	// 데이터 검증 (회사, 부서, 팀 존재 여부 확인)
	if companyId != nil {
		_, err = uc.companyRepo.GetCompanyByID(*companyId)
		if err != nil {
			return nil, common.NewError(http.StatusBadRequest, "회사가 없습니다", err)
		}
	}

	for _, deptId := range departmentIds {
		_, err = uc.departmentRepo.GetDepartmentByID(*companyId, *deptId)
		if err != nil {
			return nil, common.NewError(http.StatusBadRequest, "부서가 없습니다", err)
		}
	}

	for _, teamID := range teamIds {
		_, err = uc.teamRepo.GetTeamByID(*teamID)
		if err != nil {
			return nil, common.NewError(http.StatusBadRequest, "팀이 없습니다", err)
		}
	}

	//요청 가공 엔티티
	postEntity := &entity.Post{
		AuthorID:    *user.ID,
		Title:       post.Title,
		IsAnonymous: *post.IsAnonymous,
		Content:     post.Content,
		CompanyID:   companyId,

		CreatedAt: time.Now(),
	}

	postResponse, err := uc.postRepo.CreatePost(requestUserId, postEntity)
	if err != nil {
		return nil, common.NewError(http.StatusBadRequest, "게시물 생성 실패", err)
	}

	//TODO 응답 가공 엔티티
	response := &res.CreatePostResponse{
		Title:      postResponse.Title,
		Content:    postResponse.Content,
		AuthorName: *user.Name,
		CreatedAt:  postEntity.CreatedAt,
	}

	return response, nil
}
