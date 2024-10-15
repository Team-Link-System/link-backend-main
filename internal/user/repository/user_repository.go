package repository

import (
	"link/internal/user/entity"
	"link/pkg/dto/req"
)

// TODO 추상화
type UserRepository interface {
	CreateUser(user *entity.User) error
	GetUserByEmail(email string) (*entity.User, error)
	GetAllUsers(requestUserId uint) ([]entity.User, error)
	GetUserByID(id uint) (*entity.User, error)
	GetUserByIds(ids []uint) ([]entity.User, error)
	UpdateUser(id uint, updates map[string]interface{}) error
	DeleteUser(id uint) error
	SearchUser(request req.SearchUserRequest) ([]entity.User, error)
	GetUsersByDepartment(departmentId uint) ([]entity.User, error)
	UpdateUserOnlineStatus(userId uint, online bool) error
}
