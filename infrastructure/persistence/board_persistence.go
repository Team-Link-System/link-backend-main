package persistence

import (
	"link/infrastructure/model"
	"link/internal/board/entity"
	"link/internal/board/repository"

	"gorm.io/gorm"
)

type BoardPersistence struct {
	db *gorm.DB
}

func NewBoardPersistence(db *gorm.DB) repository.BoardRepository {
	return &BoardPersistence{db: db}
}

// ! 보드 관련
func (p *BoardPersistence) CreateBoard(board *entity.Board) error {
	boardModel := model.Board{
		Title:     board.Title,
		ProjectID: board.ProjectID,
		CreatedAt: board.CreatedAt,
		UpdatedAt: board.UpdatedAt,
	}

	if err := p.db.Create(&boardModel).Error; err != nil {
		return err
	}

	board.ID = boardModel.ID
	return nil
}

func (p *BoardPersistence) GetBoardByID(boardID uint) (*entity.Board, error) {
	var board model.Board
	if err := p.db.Where("id = ?", boardID).First(&board).Error; err != nil {
		return nil, err
	}

	boardEntity := &entity.Board{
		ID:        board.ID,
		Title:     board.Title,
		ProjectID: board.ProjectID,
		CreatedAt: board.CreatedAt,
		UpdatedAt: board.UpdatedAt,
	}

	return boardEntity, nil
}

func (p *BoardPersistence) GetBoardsByProjectID(projectID uint) ([]entity.Board, error) {
	var boards []model.Board
	if err := p.db.Where("project_id = ?", projectID).Find(&boards).Error; err != nil {
		return nil, err
	}

	boardsEntity := make([]entity.Board, len(boards))
	for i, board := range boards {
		boardsEntity[i] = entity.Board{
			ID:        board.ID,
			Title:     board.Title,
			ProjectID: board.ProjectID,
			CreatedAt: board.CreatedAt,
			UpdatedAt: board.UpdatedAt,
		}
	}

	return boardsEntity, nil
}

// ! 보드 사용자 관련
func (p *BoardPersistence) AddUserToBoard(boardUser *entity.BoardUser) error {
	boardUserModel := model.BoardUser{
		BoardID: boardUser.BoardID,
		UserID:  boardUser.UserID,
		Role:    boardUser.Role,
	}

	if err := p.db.Create(&boardUserModel).Error; err != nil {
		return err
	}

	return nil
}

func (p *BoardPersistence) CheckBoardUserRole(boardID uint, userID uint) (int, error) {
	var boardUser model.BoardUser
	if err := p.db.Where("board_id = ? AND user_id = ?", boardID, userID).First(&boardUser).Error; err != nil {
		return 0, err
	}
	return boardUser.Role, nil
}

// ! 컬럼 관련
func (p *BoardPersistence) CreateBoardColumn(boardColumn *entity.BoardColumn) error {
	//최대 포지션
	var maxPosition struct {
		MaxPos int
	}

	if err := p.db.Model(&model.BoardColumn{}).
		Select("COALESCE(MAX(position), -1) as max_pos").
		Where("board_id = ?", boardColumn.BoardID).
		Scan(&maxPosition).Error; err != nil {
		return err
	}

	boardColumn.Position = maxPosition.MaxPos + 1

	boardColumnModel := model.BoardColumn{
		BoardID:   boardColumn.BoardID,
		Name:      boardColumn.Name,
		Position:  boardColumn.Position + 1, //자등으로 다음 position 에 생성
		CreatedAt: boardColumn.CreatedAt,
		UpdatedAt: boardColumn.UpdatedAt,
	}

	if err := p.db.Create(&boardColumnModel).Error; err != nil {
		return err
	}

	// ID와 Position 설정
	boardColumn.ID = boardColumnModel.ID
	boardColumn.Position = boardColumnModel.Position
	return nil
}
