package usecase

import (
	"encoding/json"
	"fmt"
	"link/internal/auth/entity"
	_authRepo "link/internal/auth/repository"
	_userRepo "link/internal/user/repository"

	"link/pkg/common"
	"link/pkg/dto/req"
	"link/pkg/dto/res"
	_nats "link/pkg/nats"
	_utils "link/pkg/util"
	"log"
	"net/http"
	"strconv"
	"time"
)

// AuthUsecase 인터페이스 정의
type AuthUsecase interface {
	SignIn(request *req.LoginRequest) (*res.LoginUserResponse, *entity.Token, error) // 로그인 처리
	SignOut(userId uint, email string) error                                         // 로그아웃 처리
	GetRefreshToken(userId uint, email string) (string, error)
}

// authUsecase 구조체 정의
type authUsecase struct {
	authRepo      _authRepo.AuthRepository // Redis와 상호작용하는 저장소
	userRepo      _userRepo.UserRepository // 사용자 정보 저장소
	natsPublisher *_nats.NatsPublisher
}

// NewAuthUsecase 생성자 함수
// userRepo 주입
func NewAuthUsecase(authRepo _authRepo.AuthRepository, userRepo _userRepo.UserRepository, publisher *_nats.NatsPublisher) AuthUsecase {
	return &authUsecase{authRepo: authRepo, userRepo: userRepo, natsPublisher: publisher} //TODO 사용자 정보 저장소 주입
}

func (u *authUsecase) SignIn(request *req.LoginRequest) (*res.LoginUserResponse, *entity.Token, error) {

	user, err := u.userRepo.GetUserByEmail(request.Email)
	if err != nil {
		log.Printf("사용자 조회 오류: %v", err)
		return nil, nil, common.NewError(http.StatusNotFound, "이메일 또는 비밀번호가 존재하지 않습니다", err)
	}

	if !_utils.CheckPasswordHash(request.Password, *user.Password) {

		log.Printf("비밀번호 불일치: %s", request.Email)
		return nil, nil, common.NewError(http.StatusNotFound, "이메일 또는 비밀번호가 일치하지 않습니다", err)
	}

	accessToken, err := _utils.GenerateAccessToken(*user.Name, *user.Email, *user.ID)
	if err != nil {

		log.Printf("액세스 토큰 생성 오류: %v", err)
		return nil, nil, common.NewError(http.StatusInternalServerError, "액세스 토큰 생성에 실패했습니다", err)
	}

	refreshToken, err := _utils.GenerateRefreshToken(*user.Name, *user.Email, *user.ID)
	if err != nil {
		log.Printf("리프레시 토큰 생성 오류: %v", err)
		return nil, nil, common.NewError(http.StatusInternalServerError, "리프레시 토큰 생성에 실패했습니다", err)
	}

	userIdStr := strconv.FormatUint(uint64(*user.ID), 10)
	//TODO userId:email 키값으로 레디스 저장
	mergeKey := fmt.Sprintf("%s:%s", userIdStr, *user.Email)
	err = u.authRepo.StoreRefreshToken(mergeKey, refreshToken)
	if err != nil {
		log.Printf("리프레시 토큰 저장 오류: %v", err)
		return nil, nil, common.NewError(http.StatusInternalServerError, "리프레시 토큰 저장에 실패했습니다", err)
	}

	natsData := map[string]interface{}{
		"topic": "link.event.user.signin",
		"payload": map[string]interface{}{
			"user_id":    _utils.GetValueOrDefault(user.ID, 0),
			"email":      _utils.GetValueOrDefault(user.Email, ""),
			"name":       _utils.GetValueOrDefault(user.Name, ""),
			"company_id": _utils.GetValueOrDefault(user.UserProfile.CompanyID, 0),
			"timestamp":  time.Now(),
		},
	}
	jsonData, err := json.Marshal(natsData)
	if err != nil {
		log.Printf("NATS 데이터 직렬화 오류: %v", err)
		return nil, nil, common.NewError(http.StatusInternalServerError, "NATS 데이터 직렬화에 실패했습니다", err)
	}
	u.natsPublisher.PublishEvent("link.event.user.signin", []byte(jsonData))

	return &res.LoginUserResponse{
			ID:           _utils.GetValueOrDefault(user.ID, 0),
			Email:        _utils.GetValueOrDefault(user.Email, ""),
			Name:         _utils.GetValueOrDefault(user.Name, ""),
			Role:         uint(_utils.GetValueOrDefault(&user.Role, 4)),
			CompanyID:    _utils.GetValueOrDefault(user.UserProfile.CompanyID, 0),
			ProfileImage: _utils.GetValueOrDefault(user.UserProfile.Image, ""),
		}, &entity.Token{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			ExpiresAt:    time.Now().Add(24 * time.Hour), // AccessToken의 만료 시간
		}, nil
}

func (u *authUsecase) SignOut(userId uint, email string) error {
	userIdStr := strconv.FormatUint(uint64(userId), 10)
	if userIdStr == "" {
		return common.NewError(http.StatusBadRequest, "userId가 유효하지 않습니다", fmt.Errorf("userId가 유효하지 않습니다"))
	}

	mergeKey := fmt.Sprintf("%s:%s", userIdStr, email)

	err := u.authRepo.DeleteRefreshToken(mergeKey)
	if err != nil {
		log.Printf("로그아웃 처리 오류: %v", err)
		return common.NewError(http.StatusInternalServerError, "로그아웃 처리에 실패했습니다", err)
	}
	return nil
}

// TODO 레디스에서 userId로 리프레시 토큰 가져오기
func (u *authUsecase) GetRefreshToken(userId uint, email string) (string, error) {
	userIdStr := strconv.FormatUint(uint64(userId), 10)
	if userIdStr == "" {
		return "", common.NewError(http.StatusBadRequest, "userId가 유효하지 않습니다", fmt.Errorf("userId가 유효하지 않습니다"))
	}

	mergeKey := fmt.Sprintf("%s:%s", userIdStr, email)

	refreshToken, err := u.authRepo.GetRefreshToken(mergeKey)
	if err != nil {
		log.Printf("리프레시 토큰 조회 오류: %v", err)
		return "", common.NewError(http.StatusInternalServerError, "로그인 필요", err)
	}
	return refreshToken, nil
}
