package usecase

import (
	"fmt"
	"link/internal/auth/entity"
	_authRepo "link/internal/auth/repository"
	_userEntity "link/internal/user/entity"
	_userRepo "link/internal/user/repository"
	"link/pkg/util"
	"log"
	"time"
)

// AuthUsecase 인터페이스 정의
type AuthUsecase interface {
	SignIn(email, password string) (*_userEntity.User, *entity.Token, error) // 로그인 처리
	SignOut(email string) error                                              // 로그아웃 처리
	ValidateRefreshToken(refreshToken string) error                          // Refresh Token 검증 후 Access Token 재발급          // Redis에서 저장된 Refresh Token 조회
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

	err = u.authRepo.StoreRefreshToken(refreshToken, user.Email)
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

func (u *authUsecase) SignOut(refreshToken string) error {
	err := u.authRepo.DeleteRefreshToken(refreshToken)
	if err != nil {
		log.Printf("로그아웃 처리 오류: %v", err)
		return fmt.Errorf("로그아웃 처리에 실패했습니다")
	}
	return nil
}

// Refresh Token 검증 후 새로운 Access Token 발급
func (u *authUsecase) ValidateRefreshToken(refreshToken string) error {
	// Refresh Token 검증
	_, err := u.authRepo.GetEmailFromRefreshToken(refreshToken)
	if err != nil {
		log.Printf("리프레시 토큰 조회 오류: %v", err)
		return fmt.Errorf("로그인 필요")
	}
	return nil
}
