package repository

import (
	"link/internal/board/entity"
)

type BoardRepository interface {
	//보드 정보 관련
	CreateBoard(board *entity.Board, boardUsers []entity.BoardUser) error
	GetBoardByID(boardID uint) (*entity.Board, error)
	GetBoardsByProjectID(projectID uint) ([]entity.Board, error)
	UpdateBoard(board *entity.Board) error
	//보드 사용자 관련
	AddUserToBoard(boardUser *entity.BoardUser) error
	CheckBoardUserRole(boardID uint, userID uint) (int, error)
	GetBoardUsersByBoardID(boardID uint) ([]entity.BoardUser, error)
	//컬럼 관련
	CreateBoardColumn(boardColumn *entity.BoardColumn) error
	//카드 관련

	//카드 담당자 관련

}
