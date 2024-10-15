package usecase

import (
	"fmt"
	"link/internal/auth/entity"
	_authRepo "link/internal/auth/repository"
	_userEntity "link/internal/user/entity"
	_userRepo "link/internal/user/repository"
	"link/pkg/util"
	"log"
	"strconv"
	"time"
)

// AuthUsecase 인터페이스 정의
type AuthUsecase interface {
	SignIn(email, password string) (*_userEntity.User, *entity.Token, error) // 로그인 처리
	SignOut(userId uint) error                                               // 로그아웃 처리
	GetRefreshToken(userId uint) (string, error)
}

// authUsecase 구조체 정의
type authUsecase struct {
	authRepo _authRepo.AuthRepository // Redis와 상호작용하는 저장소
	userRepo _userRepo.UserRepository // 사용자 정보 저장소
}

// NewAuthUsecase 생성자 함수
// userRepo 주입
func NewAuthUsecase(authRepo _authRepo.AuthRepository, userRepo _userRepo.UserRepository) AuthUsecase {
	return &authUsecase{authRepo: authRepo, userRepo: userRepo} //TODO 사용자 정보 저장소 주입
}

func (u *authUsecase) SignIn(email, password string) (*_userEntity.User, *entity.Token, error) {
	user, err := u.userRepo.GetUserByEmail(email)
	if err != nil {
		log.Printf("사용자 조회 오류: %v", err)
		return nil, nil, fmt.Errorf("이메일 또는 비밀번호가 존재하지 않습니다")
	}

	if !util.CheckPasswordHash(password, user.Password) {
		log.Printf("비밀번호 불일치: %s", email)
		return nil, nil, fmt.Errorf("이메일 또는 비밀번호가 일치하지 않습니다")
	}

	accessToken, err := util.GenerateAccessToken(user.Name, user.Email, user.ID)
	if err != nil {
		log.Printf("액세스 토큰 생성 오류: %v", err)
		return nil, nil, fmt.Errorf("액세스 토큰 생성에 실패했습니다")
	}

	refreshToken, err := util.GenerateRefreshToken(user.Name, user.Email, user.ID)
	if err != nil {
		log.Printf("리프레시 토큰 생성 오류: %v", err)
		return nil, nil, fmt.Errorf("리프레시 토큰 생성에 실패했습니다")
	}

	userIdStr := strconv.FormatUint(uint64(user.ID), 10)
	err = u.authRepo.StoreRefreshToken(refreshToken, userIdStr)
	if err != nil {
		log.Printf("리프레시 토큰 저장 오류: %v", err)
		return nil, nil, fmt.Errorf("리프레시 토큰 저장에 실패했습니다")
	}

	return user, &entity.Token{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(24 * time.Hour), // AccessToken의 만료 시간
	}, nil
}

func (u *authUsecase) SignOut(userId uint) error {
	userIdStr := strconv.FormatUint(uint64(userId), 10)
	if userIdStr == "" {
		return fmt.Errorf("userId가 유효하지 않습니다")
	}

	err := u.authRepo.DeleteRefreshToken(userIdStr)
	if err != nil {
		log.Printf("로그아웃 처리 오류: %v", err)
		return fmt.Errorf("로그아웃 처리에 실패했습니다")
	}
	return nil
}

// TODO 레디스에서 userId로 리프레시 토큰 가져오기
func (u *authUsecase) GetRefreshToken(userId uint) (string, error) {
	userIdStr := strconv.FormatUint(uint64(userId), 10)
	fmt.Println("userIdStr:", userIdStr)
	if userIdStr == "" {
		return "", fmt.Errorf("userId가 유효하지 않습니다")
	}

	refreshToken, err := u.authRepo.GetRefreshToken(userIdStr)
	if err != nil {
		log.Printf("리프레시 토큰 조회 오류: %v", err)
		return "", fmt.Errorf("로그인 필요")
	}
	return refreshToken, nil
}
