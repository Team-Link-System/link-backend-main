package repository

import "link/internal/company/entity"

type PositionRepository interface {
	CreatePosition(position *entity.Position) error
}
