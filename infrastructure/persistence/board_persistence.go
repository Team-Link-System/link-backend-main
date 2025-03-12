package persistence

import (
	"fmt"
	"link/infrastructure/model"
	"link/internal/board/entity"
	"link/internal/board/repository"
	"strings"

	"gorm.io/gorm"
)

type BoardPersistence struct {
	db *gorm.DB
}

func NewBoardPersistence(db *gorm.DB) repository.BoardRepository {
	return &BoardPersistence{db: db}
}

// ! 보드 관련
func (p *BoardPersistence) CreateBoard(board *entity.Board, boardUsers []entity.BoardUser) error {
	tx := p.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	//보드도 만들면서 사용자도 추가해야함
	boardModel := model.Board{
		Title:     board.Title,
		ProjectID: board.ProjectID,
		CreatedAt: board.CreatedAt,
		UpdatedAt: board.UpdatedAt,
	}

	if err := tx.Create(&boardModel).Error; err != nil {
		tx.Rollback()
		return err
	}

	board.ID = boardModel.ID

	for i := range boardUsers {
		boardUsers[i].BoardID = board.ID
	}

	userQuery := `INSERT INTO board_users (board_id, user_id, role) VALUES `
	userValues := []interface{}{}
	placeHolders := []string{}
	for i, boardUser := range boardUsers {

		userValues = append(userValues, board.ID, boardUser.UserID, boardUser.Role)
		placeHolders = append(placeHolders, fmt.Sprintf("($%d, $%d, $%d)", i*3+1, i*3+2, i*3+3))
	}

	userQuery += strings.Join(placeHolders, ", ")

	result := tx.Exec(userQuery, userValues...)
	if result.Error != nil {
		tx.Rollback()
		return result.Error
	}

	board.ID = boardModel.ID
	return tx.Commit().Error
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

func (p *BoardPersistence) UpdateBoard(board *entity.Board) error {
	if err := p.db.Model(&model.Board{}).Where("id = ?", board.ID).Updates(map[string]interface{}{
		"title":      board.Title,
		"project_id": board.ProjectID,
		"updated_at": board.UpdatedAt,
	}).Error; err != nil {
		return err
	}
	return nil
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
	if err := p.db.Select("role").Where("board_id = ? AND user_id = ?", boardID, userID).First(&boardUser).Error; err != nil {
		return 0, err
	}
	return boardUser.Role, nil
}

func (p *BoardPersistence) GetBoardUsersByBoardID(boardID uint) ([]entity.BoardUser, error) {
	var boardUsers []model.BoardUser
	if err := p.db.Where("board_id = ?", boardID).Find(&boardUsers).Error; err != nil {
		return nil, err
	}

	boardUsersEntity := make([]entity.BoardUser, len(boardUsers))
	for i, boardUser := range boardUsers {
		boardUsersEntity[i] = entity.BoardUser{
			BoardID: boardUser.BoardID,
			UserID:  boardUser.UserID,
			Role:    boardUser.Role,
		}
	}

	return boardUsersEntity, nil
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
