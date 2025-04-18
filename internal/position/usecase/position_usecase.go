package usecase

import (
	"link/pkg/dto/req"

	_positionRepo "link/internal/position/repository"
)

type PositionUsecase interface {
	CreatePosition(requestUserId uint, companyId string, request req.CreatePositionRequest) error
}

type positionUsecase struct {
	positionRepository _positionRepo.PositionRepository
}

func NewPositionUsecase(positionRepository _positionRepo.PositionRepository) PositionUsecase {
	return &positionUsecase{positionRepository: positionRepository}
}

func (u *positionUsecase) CreatePosition(requestUserId uint, companyId string, request req.CreatePositionRequest) error {

}
