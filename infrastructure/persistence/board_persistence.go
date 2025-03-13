package persistence

import (
	"fmt"
	"link/infrastructure/model"
	"link/internal/board/entity"
	"link/internal/board/repository"
	"strings"

	"github.com/google/uuid"
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

func (p *BoardPersistence) DeleteBoard(boardID uint) error {
	if err := p.db.Where("id = ?", boardID).Delete(&model.Board{}).Error; err != nil {
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

	boardColumn.Position = uint(maxPosition.MaxPos) + 1

	boardColumnModel := model.BoardColumn{
		ID:        boardColumn.ID,
		BoardID:   boardColumn.BoardID,
		Name:      boardColumn.Name,
		Position:  boardColumn.Position, //자등으로 다음 position 에 생성
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

func (p *BoardPersistence) GetBoardColumnByID(columnID uuid.UUID) (*entity.BoardColumn, error) {
	var boardColumn model.BoardColumn
	if err := p.db.Where("id = ?", columnID).First(&boardColumn).Error; err != nil {
		return nil, err
	}

	boardColumnEntity := &entity.BoardColumn{
		ID:        boardColumn.ID,
		Name:      boardColumn.Name,
		BoardID:   boardColumn.BoardID,
		Position:  boardColumn.Position,
		CreatedAt: boardColumn.CreatedAt,
		UpdatedAt: boardColumn.UpdatedAt,
	}

	return boardColumnEntity, nil
}

func (p *BoardPersistence) UpdateBoardColumn(boardColumn *entity.BoardColumn) error {
	if err := p.db.Model(&model.BoardColumn{}).Where("id = ?", boardColumn.ID).Updates(map[string]interface{}{
		"name":       boardColumn.Name,
		"position":   boardColumn.Position,
		"updated_at": boardColumn.UpdatedAt,
	}).Error; err != nil {
		return err
	}
	return nil
}

func (p *BoardPersistence) DeleteBoardColumn(columnID uuid.UUID) error {
	if err := p.db.Where("id = ?", columnID).Delete(&model.BoardColumn{}).Error; err != nil {
		return err
	}
	return nil
}

func (p *BoardPersistence) MoveBoardColumn(columnID uuid.UUID, newPosition uint) error {
	tx := p.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	//현재 컬럼 정보 조회
	var currentColumn model.BoardColumn
	if err := tx.Where("id = ?", columnID).First(&currentColumn).Error; err != nil {
		tx.Rollback()
		return err
	}

	currentPosition := currentColumn.Position
	boardID := currentColumn.BoardID

	//컬럼을 옮기면 , 옆의 컬럼들도 밀려나야함
	if currentPosition == newPosition {
		return tx.Commit().Error
	}

	//위치 이동 방향에 따라 다른 컬럼들의 위치 조정
	if currentPosition < newPosition {
		// 왼쪽으로 옮기는 경우
		if err := tx.Model(&model.BoardColumn{}).
			Where("board_id = ? AND position > ? AND position <= ?", boardID, currentPosition, newPosition).
			Update("position", gorm.Expr("position - 1")).Error; err != nil {
			tx.Rollback()
			return err
		}
	} else {
		// 오른쪽으로 옮기는 경우
		if err := tx.Model(&model.BoardColumn{}).
			Where("board_id = ? AND position >= ? AND position < ?", boardID, newPosition, currentPosition).
			Update("position", gorm.Expr("position + 1")).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	if err := tx.Model(&model.BoardColumn{}).Where("id = ?", columnID).Update("position", newPosition).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

// ! 카드 관련
func (p *BoardPersistence) CreateBoardCard(boardCard *entity.BoardCard) error {
	// 최대 포지션 조회
	tx := p.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	var maxPosition struct {
		MaxPos int
	}

	if err := tx.Model(&model.BoardCard{}).
		Select("COALESCE(MAX(position), -1) as max_pos").
		Where("board_column_id = ?", boardCard.BoardColumnID).
		Scan(&maxPosition).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 포지션이 지정되지 않은 경우 최대값 + 1로 설정
	if boardCard.Position == 0 {
		boardCard.Position = uint(maxPosition.MaxPos + 1)
	}

	// 모델 생성
	boardCardModel := model.BoardCard{
		ID:            boardCard.ID,
		BoardID:       boardCard.BoardID,
		BoardColumnID: boardCard.BoardColumnID,
		Name:          boardCard.Name,
		Content:       boardCard.Content,
		Position:      boardCard.Position,
		StartDate:     boardCard.StartDate,
		EndDate:       boardCard.EndDate,
		Version:       boardCard.Version,
		CreatedAt:     boardCard.CreatedAt,
		UpdatedAt:     boardCard.UpdatedAt,
	}

	if len(boardCard.Assignees) > 0 {
		if err := tx.Where("card_id = ?", boardCard.ID).Delete(&model.CardAssignee{}).Error; err != nil {
			tx.Rollback()
			return err
		}

		for _, userID := range boardCard.Assignees {
			assignees := model.CardAssignee{
				CardID: boardCard.ID,
				UserID: userID,
			}
			if err := tx.Create(&assignees).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
	}

	// 데이터베이스에 카드 생성
	if err := tx.Create(&boardCardModel).Error; err != nil {
		tx.Rollback()
		return err
	}

	boardCard.ID = boardCardModel.ID
	return tx.Commit().Error
}

func (p *BoardPersistence) GetBoardCardByID(cardID uuid.UUID) (*entity.BoardCard, error) {
	var boardCard model.BoardCard
	if err := p.db.Preload("Assignees").Where("id = ?", cardID).First(&boardCard).Error; err != nil {
		return nil, err
	}

	boardCardEntity := &entity.BoardCard{
		ID:            boardCard.ID,
		Name:          boardCard.Name,
		Content:       boardCard.Content,
		BoardID:       boardCard.BoardID,
		BoardColumnID: boardCard.BoardColumnID,
		Position:      boardCard.Position,
		StartDate:     boardCard.StartDate,
		EndDate:       boardCard.EndDate,
		Version:       boardCard.Version,
		CreatedAt:     boardCard.CreatedAt,
		UpdatedAt:     boardCard.UpdatedAt,
	}

	assignees := make([]model.CardAssignee, len(boardCard.Assignees))
	for i, cardAssignee := range boardCard.Assignees {
		assignees[i] = model.CardAssignee{
			CardID: cardAssignee.CardID,
			UserID: cardAssignee.UserID,
		}
	}

	return boardCardEntity, nil
}

func (p *BoardPersistence) UpdateBoardCard(boardCard *entity.BoardCard) error {
	tx := p.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 카드 기본 정보 업데이트
	if err := tx.Model(&model.BoardCard{}).Where("id = ?", boardCard.ID).Updates(map[string]interface{}{
		"name":       boardCard.Name,
		"content":    boardCard.Content,
		"position":   boardCard.Position,
		"start_date": boardCard.StartDate,
		"end_date":   boardCard.EndDate,
		"version":    boardCard.Version,
		"updated_at": boardCard.UpdatedAt,
	}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Assignees가 있는 경우에만 업데이트
	if len(boardCard.Assignees) > 0 {
		// 기존 할당자 삭제
		if err := tx.Where("card_id = ?", boardCard.ID).Delete(&model.CardAssignee{}).Error; err != nil {
			tx.Rollback()
			return err
		}

		// 새 할당자 추가
		assignees := make([]model.CardAssignee, len(boardCard.Assignees))
		for i, userID := range boardCard.Assignees {
			assignees[i] = model.CardAssignee{
				CardID: boardCard.ID,
				UserID: userID,
			}
		}

		if err := tx.Create(&assignees).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit().Error
}

func (p *BoardPersistence) DeleteBoardCard(cardID uuid.UUID) error {
	if err := p.db.Where("id = ?", cardID).Delete(&model.BoardCard{}).Error; err != nil {
		return err
	}
	return nil
}

func (p *BoardPersistence) MoveBoardCard(cardID uuid.UUID, toColumnID *uuid.UUID, newPosition *uint) error {
	tx := p.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 현재 카드 정보 조회
	var card model.BoardCard
	if err := tx.First(&card, cardID).Error; err != nil {
		tx.Rollback()
		return err
	}

	fromColumnID := card.BoardColumnID
	currentPosition := card.Position

	isColumnChanged := toColumnID != nil && fromColumnID != *toColumnID
	isPositionChanged := newPosition != nil && currentPosition != *newPosition

	if !isColumnChanged && !isPositionChanged {
		return tx.Commit().Error
	}

	targetColumnID := fromColumnID
	if isColumnChanged {
		targetColumnID = *toColumnID
	}

	targetPosition := currentPosition
	if isPositionChanged {
		targetPosition = *newPosition
	}

	// 다른 컬럼으로 이동하는 경우, 위치가 지정되지 않았다면 마지막 위치로 설정
	if isColumnChanged && !isPositionChanged {
		var maxPosition struct {
			MaxPos uint
		}
		if err := tx.Model(&model.BoardCard{}).
			Select("COALESCE(MAX(position), 0) as max_pos").
			Where("board_column_id = ?", targetColumnID).
			Scan(&maxPosition).Error; err != nil {
			tx.Rollback()
			return err
		}
		targetPosition = maxPosition.MaxPos + 1
	}

	if !isColumnChanged && isPositionChanged {
		if currentPosition < targetPosition {
			// 위로 이동: 현재 위치와 새 위치 사이의 카드들을 한 칸씩 아래로 이동
			if err := tx.Model(&model.BoardCard{}).
				Where("board_column_id = ? AND position > ? AND position <= ?",
					fromColumnID, currentPosition, targetPosition).
				Update("position", gorm.Expr("position - 1")).Error; err != nil {
				tx.Rollback()
				return err
			}
		} else {
			// 아래로 이동: 새 위치와 현재 위치 사이의 카드들을 한 칸씩 위로 이동
			if err := tx.Model(&model.BoardCard{}).
				Where("board_column_id = ? AND position >= ? AND position < ?",
					fromColumnID, targetPosition, currentPosition).
				Update("position", gorm.Expr("position + 1")).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
	} else if isColumnChanged {
		// 원래 컬럼에서 카드 제거 후 위치 조정
		if err := tx.Model(&model.BoardCard{}).
			Where("board_column_id = ? AND position > ?",
				fromColumnID, currentPosition).
			Update("position", gorm.Expr("position - 1")).Error; err != nil {
			tx.Rollback()
			return err
		}

		// 새 컬럼에서 공간 확보
		if err := tx.Model(&model.BoardCard{}).
			Where("board_column_id = ? AND position >= ?",
				targetColumnID, targetPosition).
			Update("position", gorm.Expr("position + 1")).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	// 이동하는 카드의 컬럼과 위치 업데이트
	updates := map[string]interface{}{
		"board_column_id": targetColumnID,
		"position":        targetPosition,
	}

	if err := tx.Model(&card).Updates(updates).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}
