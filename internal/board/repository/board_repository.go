package repository

import (
	"link/internal/board/entity"

	"github.com/google/uuid"
)

type BoardRepository interface {
	//보드 정보 관련
	CreateBoard(board *entity.Board, boardUsers []entity.BoardUser) error
	GetBoardByID(boardID uint) (*entity.Board, error)
	GetBoardsByProjectID(projectID uint) ([]entity.Board, error)
	UpdateBoard(board *entity.Board) error
	DeleteBoard(boardID uint) error
	//보드 사용자 관련
	AddUserToBoard(boardUser *entity.BoardUser) error
	CheckBoardUserRole(boardID uint, userID uint) (int, error)
	GetBoardUsersByBoardID(boardID uint) ([]entity.BoardUser, error)
	//컬럼 관련
	CreateBoardColumn(boardColumn *entity.BoardColumn) error
	GetBoardColumnByID(columnID uuid.UUID) (*entity.BoardColumn, error)
	UpdateBoardColumn(boardColumn *entity.BoardColumn) error
	DeleteBoardColumn(columnID uuid.UUID) error
	MoveBoardColumn(columnID uuid.UUID, position uint) error
	//카드 관련
	CreateBoardCard(boardCard *entity.BoardCard) error
	GetBoardCardByID(cardID uuid.UUID) (*entity.BoardCard, error)
	UpdateBoardCard(boardCard *entity.BoardCard) error
	DeleteBoardCard(cardID uuid.UUID) error
	MoveBoardCard(cardID uuid.UUID, toColumnID *uuid.UUID, newPosition *uint) error
}
