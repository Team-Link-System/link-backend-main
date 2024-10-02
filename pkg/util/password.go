package util

import "golang.org/x/crypto/bcrypt"

// 비밀번호 해싱 함수
func HashPassword(password string) (string, error) {
	// bcrypt 해싱 (14는 해싱 강도를 의미)
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

// 비밀번호 검증 함수
func CheckPasswordHash(password, hashedPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}
