package usecase

import (
	"encoding/json"
	"link/internal/board/entity"
	_boardRepo "link/internal/board/repository"
	_projectRepo "link/internal/project/repository"
	_userEntity "link/internal/user/entity"
	_userRepo "link/internal/user/repository"

	"link/pkg/common"
	"link/pkg/dto/req"
	"link/pkg/dto/res"
	"log"
	"net/http"
	"time"

	_nats "link/pkg/nats"

	"github.com/google/uuid"
)

type BoardUsecase interface {
	CreateBoard(userId uint, request *req.CreateBoardRequest) error
	GetBoard(userId uint, boardID uint) (*res.GetBoardResponse, error)
	GetBoards(userId uint, projectID uint) (*res.GetBoardsResponse, error)
	UpdateBoard(userId uint, boardID uint, request *req.UpdateBoardRequest) error
	DeleteBoard(userId uint, boardID uint) error

	AutoSaveBoard(userId uint, projectID uint, boardID uint, request *req.BoardStateUpdateReqeust) error
	GetKanbanBoard(userId uint, boardID uint) (*res.GetKanbanBoardResponse, error)
}

type boardUsecase struct {
	boardRepo     _boardRepo.BoardRepository
	userRepo      _userRepo.UserRepository
	projectRepo   _projectRepo.ProjectRepository
	natsPublisher *_nats.NatsPublisher
}

func NewBoardUsecase(
	boardRepo _boardRepo.BoardRepository,
	userRepo _userRepo.UserRepository,
	projectRepo _projectRepo.ProjectRepository,
	natsPublisher *_nats.NatsPublisher) BoardUsecase {
	return &boardUsecase{
		boardRepo:     boardRepo,
		userRepo:      userRepo,
		projectRepo:   projectRepo,
		natsPublisher: natsPublisher,
	}
}

// ! 보드 관련
func (u *boardUsecase) CreateBoard(userId uint, request *req.CreateBoardRequest) error {
	_, err := u.userRepo.GetUserByID(userId)
	if err != nil {
		return common.NewError(http.StatusBadRequest, "사용자 조회 실패", err)
	}
	hasAcess, err := u.projectRepo.GetProjectByID(userId, request.ProjectID)
	if err != nil {
		return common.NewError(http.StatusInternalServerError, "프로젝트 조회 실패", err)
	}

	if hasAcess == nil {
		return common.NewError(http.StatusForbidden, "프로젝트 접근 권한 없음", nil)
	}

	board := entity.Board{
		Title:     request.Title,
		ProjectID: request.ProjectID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	//userId를 제외한 나머지 projectUser들의 boardRole을 0으로 변경
	projectUsers, err := u.projectRepo.GetProjectUsers(request.ProjectID)
	if err != nil {
		return common.NewError(http.StatusInternalServerError, "프로젝트 사용자 조회 실패", err)
	}

	boardUsers := make([]entity.BoardUser, 0, len(projectUsers))
	for _, projectUser := range projectUsers {
		role := 0 // 기본적으로 읽기 권한 없음
		if projectUser.UserID == userId {
			role = 2 // 생성자는 관리자
		}

		boardUsers = append(boardUsers, entity.BoardUser{
			UserID:  projectUser.UserID,
			Role:    role,
			BoardID: board.ID,
		})
	}

	if err := u.boardRepo.CreateBoard(&board, boardUsers); err != nil {
		return common.NewError(http.StatusInternalServerError, "보드 생성 실패", err)
	}

	//TODO 더미로 생성 이후 삭제 필요
	defaultColums := []entity.BoardColumn{ // 일단 더미로 생성
		{
			ID:        uuid.New(),
			Name:      "To Do",
			BoardID:   board.ID,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:        uuid.New(),
			Name:      "In Progress",
			BoardID:   board.ID,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:        uuid.New(),
			Name:      "Done",
			BoardID:   board.ID,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	for _, column := range defaultColums {
		if err := u.boardRepo.CreateBoardColumn(&column); err != nil {
			return common.NewError(http.StatusInternalServerError, "보드 컬럼 생성 실패", err)
		}
	}

	user, _ := u.userRepo.GetUserByID(userId)

	//mongoDB 에 로그성 데이터는 nats로 전송
	docID := uuid.New().String()

	natsData := map[string]interface{}{
		"topic": "link.event.notification.board.create",
		"payload": map[string]interface{}{
			"doc_id":      docID,
			"board_id":    board.ID,
			"title":       board.Title,
			"project_id":  board.ProjectID,
			"created_at":  board.CreatedAt,
			"updated_at":  board.UpdatedAt,
			"user_id":     *user.ID,
			"user_name":   *user.Name,
			"alarm_type":  "BOARD",
			"target_type": "BOARD",
			"action":      "create.board",
			"timestamp":   time.Now(),
		},
	}

	jsonData, err := json.Marshal(natsData)
	if err != nil {
		log.Printf("NATS 데이터 직렬화 실패: %v", err)
		return common.NewError(http.StatusInternalServerError, "NATS 데이터 직렬화 실패", err)
	}

	go u.natsPublisher.PublishEvent("link.event.notification.board.create", jsonData)

	return nil
}

func (u *boardUsecase) GetBoards(userId uint, projectID uint) (*res.GetBoardsResponse, error) {
	_, err := u.userRepo.GetUserByID(userId)
	if err != nil {
		return nil, common.NewError(http.StatusInternalServerError, "사용자 조회 실패", err)
	}

	project, err := u.projectRepo.GetProjectByID(userId, projectID)
	if err != nil {
		return nil, common.NewError(http.StatusInternalServerError, "프로젝트 조회 실패", err)
	}

	boards, err := u.boardRepo.GetBoardsByProjectID(projectID)
	if err != nil {
		return nil, common.NewError(http.StatusInternalServerError, "보드 조회 실패", err)
	}

	boardsResponse := make([]res.GetBoardResponse, len(boards))
	for i, board := range boards {
		boardsResponse[i] = res.GetBoardResponse{
			BoardID:     board.ID,
			Title:       board.Title,
			ProjectID:   board.ProjectID,
			ProjectName: project.Name,
			CreatedAt:   board.CreatedAt,
			UpdatedAt:   board.UpdatedAt,
		}
	}

	return &res.GetBoardsResponse{
		Boards: boardsResponse,
	}, nil
}

func (u *boardUsecase) GetBoard(userId uint, boardID uint) (*res.GetBoardResponse, error) {

	_, err := u.userRepo.GetUserByID(userId)
	if err != nil {
		return nil, common.NewError(http.StatusInternalServerError, "사용자 조회 실패", err)
	}

	board, err := u.boardRepo.GetBoardByID(boardID)
	if err != nil {
		return nil, common.NewError(http.StatusInternalServerError, "보드 조회 실패", err)
	}

	checkBoardUserRole, err := u.boardRepo.CheckBoardUserRole(boardID, userId)
	if err != nil {
		return nil, common.NewError(http.StatusInternalServerError, "보드 사용자 권한 조회 실패", err)
	}

	project, err := u.projectRepo.GetProjectByID(userId, board.ProjectID)
	if err != nil {
		return nil, common.NewError(http.StatusInternalServerError, "프로젝트 조회 실패", err)
	}

	boardResponse := &res.GetBoardResponse{
		BoardID:       board.ID,
		Title:         board.Title,
		ProjectID:     board.ProjectID,
		ProjectName:   project.Name,
		UserBoardRole: &checkBoardUserRole,
		CreatedAt:     board.CreatedAt,
		UpdatedAt:     board.UpdatedAt,
	}

	return boardResponse, nil
}

func (u *boardUsecase) UpdateBoard(userId uint, boardID uint, request *req.UpdateBoardRequest) error {

	if request.ProjectID == nil && request.Title == "" {
		return common.NewError(http.StatusBadRequest, "프로젝트 ID 또는 제목이 필요합니다.", nil)
	}

	_, err := u.userRepo.GetUserByID(userId)
	if err != nil {
		return common.NewError(http.StatusInternalServerError, "사용자 조회 실패", err)
	}

	board, err := u.boardRepo.GetBoardByID(boardID)
	if err != nil {
		return common.NewError(http.StatusInternalServerError, "보드 조회 실패", err)
	}

	originalProjectId := board.ProjectID

	//보드에 대한 권한 확인
	checkBoardUserRole, err := u.boardRepo.CheckBoardUserRole(boardID, userId)
	if err != nil {
		return common.NewError(http.StatusInternalServerError, "보드 사용자 권한 조회 실패", err)
	}

	if request.ProjectID != nil && originalProjectId != *request.ProjectID {

		_, err := u.projectRepo.GetProjectByProjectID(*request.ProjectID)
		if err != nil {
			return common.NewError(http.StatusInternalServerError, "해당 보드를 옮기려는 프로젝트가 존재하지 않습니다.", err)
		}

		projectUsers, err := u.projectRepo.GetProjectUsers(*request.ProjectID)
		if err != nil {
			return common.NewError(http.StatusInternalServerError, "보드 사용자 조회 실패", err)
		}

		boardUsers, err := u.boardRepo.GetBoardUsersByBoardID(boardID)
		if err != nil {
			return common.NewError(http.StatusInternalServerError, "보드 사용자 조회 실패", err)
		}

		boardUsersIsExistInProjectMap := make(map[uint]bool) //보드에 있는 사용자들이 프로젝트에 존재하는지 확인
		for _, boardUser := range boardUsers {
			for _, projectUser := range projectUsers {
				if boardUser.UserID == projectUser.UserID {
					boardUsersIsExistInProjectMap[boardUser.UserID] = true
				}
			}
		}

		for _, projectUser := range projectUsers {
			if !boardUsersIsExistInProjectMap[projectUser.UserID] {
				log.Println("변경하려는 프로젝트에 포함되지 않는 유저가 있습니다. 먼저 프로젝트에 추가해주세요.")
				return common.NewError(http.StatusForbidden, "변경하려는 프로젝트에 포함되지 않는 유저가 있습니다. 먼저 프로젝트에 추가해주세요.", nil)
			}
		}

		board.ProjectID = *request.ProjectID
	}

	if checkBoardUserRole < entity.BoardRoleAdmin {
		return common.NewError(http.StatusForbidden, "해당 보드의 수정 권한이 없습니다.", nil)
	}

	if request.Title != "" {
		board.Title = request.Title
	}

	if err := u.boardRepo.UpdateBoard(board); err != nil {
		return common.NewError(http.StatusInternalServerError, "보드 업데이트 실패", err)
	}

	return nil
}

func (u *boardUsecase) DeleteBoard(userId uint, boardID uint) error {
	_, err := u.userRepo.GetUserByID(userId)
	if err != nil {
		return common.NewError(http.StatusInternalServerError, "사용자 조회 실패", err)
	}

	_, err = u.boardRepo.GetBoardByID(boardID)
	if err != nil {
		return common.NewError(http.StatusInternalServerError, "보드 조회 실패", err)
	}

	checkBoardUserRole, err := u.boardRepo.CheckBoardUserRole(boardID, userId)
	if err != nil {
		return common.NewError(http.StatusInternalServerError, "보드 사용자 권한 조회 실패", err)
	}

	if checkBoardUserRole < entity.BoardRoleAdmin {
		return common.NewError(http.StatusForbidden, "해당 보드의 삭제 권한이 없습니다.", nil)
	}

	if err := u.boardRepo.DeleteBoard(boardID); err != nil {
		return common.NewError(http.StatusInternalServerError, "보드 삭제 실패", err)
	}

	return nil
}

func (u *boardUsecase) AutoSaveBoard(userId uint, projectID uint, boardID uint, request *req.BoardStateUpdateReqeust) error {
	_, err := u.userRepo.GetUserByID(userId)
	if err != nil {
		return common.NewError(http.StatusInternalServerError, "사용자 조회 실패", err)
	}

	_, err = u.projectRepo.GetProjectByID(userId, projectID)
	if err != nil {
		return common.NewError(http.StatusInternalServerError, "프로젝트 조회 실패", err)
	}

	_, err = u.boardRepo.GetBoardByID(boardID)
	if err != nil {
		return common.NewError(http.StatusInternalServerError, "보드 조회 실패", err)
	}

	_, err = u.boardRepo.GetBoardUsersByBoardID(boardID)
	if err != nil {
		return common.NewError(http.StatusInternalServerError, "보드 사용자 조회 실패", err)
	}

	role, err := u.boardRepo.CheckBoardUserRole(boardID, userId)
	if err != nil {
		return common.NewError(http.StatusInternalServerError, "보드 사용자 권한 조회 실패", err)
	}

	if role < entity.BoardRoleMaintainer {
		return common.NewError(http.StatusForbidden, "해당 보드의 수정 권한이 없습니다.", nil)
	}

	if request.Changes == nil {
		return common.NewError(http.StatusBadRequest, "변경사항이 없습니다.", nil)
	}

	for _, change := range request.Changes {
		switch change.Type {
		case "column":
			if change.Action == "create" {

				// 새로운 컬럼 생성
				newColumn := entity.BoardColumn{
					ID:        *change.ColumnID,
					Name:      *change.Name,
					BoardID:   boardID,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}
				if err := u.boardRepo.CreateBoardColumn(&newColumn); err != nil {
					return common.NewError(http.StatusInternalServerError, "보드 컬럼 생성 실패", err)
				}
			} else if change.Action == "update" {
				// 컬럼 이름 변경
				column, err := u.boardRepo.GetBoardColumnByID(*change.ColumnID)
				if err != nil {
					return common.NewError(http.StatusInternalServerError, "보드 컬럼 조회 실패", err)
				}
				if change.Name != nil {
					column.Name = *change.Name
				}
				if err := u.boardRepo.UpdateBoardColumn(column); err != nil {
					return common.NewError(http.StatusInternalServerError, "보드 컬럼 업데이트 실패", err)
				}
			} else if change.Action == "delete" {
				// 컬럼 삭제
				if err := u.boardRepo.DeleteBoardColumn(*change.ColumnID); err != nil {
					return common.NewError(http.StatusInternalServerError, "보드 컬럼 삭제 실패", err)
				}
			} else if change.Action == "move" {
				// 컬럼 이동
				if change.Position != nil {
					if err := u.boardRepo.MoveBoardColumn(*change.ColumnID, *change.Position); err != nil {
						return common.NewError(http.StatusInternalServerError, "보드 컬럼 이동 실패", err)
					}
				}
			}
		case "card":
			if change.Action == "create" {

				column, err := u.boardRepo.GetBoardColumnByID(*change.ColumnID)
				if err != nil || column == nil {
					return common.NewError(http.StatusInternalServerError, "카드 생성 실패: 컬럼이 존재하지 않음", err)
				}
				// 새로운 카드 생성
				newCard := entity.BoardCard{
					ID:            change.CardID,
					Name:          *change.Name,
					Content:       *change.Content,
					BoardID:       boardID,
					BoardColumnID: *change.ColumnID,
					Assignees:     change.Assignees,
					CreatedAt:     time.Now(),
					UpdatedAt:     time.Now(),
				}

				if err := u.boardRepo.CreateBoardCard(&newCard); err != nil {
					return common.NewError(http.StatusInternalServerError, "보드 카드 생성 실패", err)
				}
			} else if change.Action == "update" {
				// 카드 업데이트
				card, err := u.boardRepo.GetBoardCardByID(change.CardID)
				if err != nil {
					return common.NewError(http.StatusInternalServerError, "보드 카드 조회 실패", err)
				}
				if change.Name != nil {
					card.Name = *change.Name
				}
				if change.Content != nil {
					card.Content = *change.Content
				}

				if change.Assignees != nil {
					card.Assignees = change.Assignees
				}

				loc, err := time.LoadLocation("Asia/Seoul")
				if err != nil {
					log.Printf("시간대 로드 실패: %v", err)
					return common.NewError(http.StatusBadRequest, "시간대 로드 실패", err)
				}
				if change.StartDate != nil {
					startTime, err := time.ParseInLocation("2006-01-02 15:04:05", *change.StartDate, loc)
					if err != nil {
						log.Printf("시작일 파싱 실패: %v", err)
						return common.NewError(http.StatusBadRequest, "시작일 파싱 실패", err)
					}
					card.StartDate = startTime
				}
				if change.EndDate != nil {
					endTime, err := time.ParseInLocation("2006-01-02 15:04:05", *change.EndDate, loc)
					if err != nil {
						log.Printf("종료일 파싱 실패: %v", err)
						return common.NewError(http.StatusBadRequest, "종료일 파싱 실패", err)
					}
					card.EndDate = endTime
				}

				card.StartDate = time.Now()
				card.EndDate = time.Now()

				if err := u.boardRepo.UpdateBoardCard(card); err != nil {
					return common.NewError(http.StatusInternalServerError, "보드 카드 업데이트 실패", err)
				}
			} else if change.Action == "delete" {
				// 카드 삭제
				if err := u.boardRepo.DeleteBoardCard(change.CardID); err != nil {
					return common.NewError(http.StatusInternalServerError, "보드 카드 삭제 실패", err)
				}
			} else if change.Action == "move" {
				if err := u.boardRepo.MoveBoardCard(change.CardID, change.ColumnID, change.Position); err != nil {
					return common.NewError(http.StatusInternalServerError, "보드 카드 이동 실패", err)
				}
			}
		}
	}

	return nil
}

// 칸반보드 렌더링 조회
func (u *boardUsecase) GetKanbanBoard(userId uint, boardID uint) (*res.GetKanbanBoardResponse, error) {
	_, err := u.userRepo.GetUserByID(userId)
	if err != nil {
		return nil, common.NewError(http.StatusInternalServerError, "사용자 조회 실패", err)
	}

	board, err := u.boardRepo.GetBoardByID(boardID)
	if err != nil {
		return nil, common.NewError(http.StatusInternalServerError, "보드 조회 실패", err)
	}

	userBoardRole, err := u.boardRepo.CheckBoardUserRole(boardID, userId)
	if err != nil {
		return nil, common.NewError(http.StatusForbidden, "해당 칸반보드에 접근할 수 없습니다.", err)
	}

	project, err := u.projectRepo.GetProjectByID(userId, board.ProjectID)
	if err != nil {
		return nil, common.NewError(http.StatusInternalServerError, "프로젝트 조회 실패", err)
	}

	columns, err := u.boardRepo.GetBoardColumnsByBoardID(boardID)
	if err != nil {
		return nil, common.NewError(http.StatusInternalServerError, "보드 컬럼 조회 실패", err)
	}

	// 컬럼 응답 구성
	columnsResponse := make([]res.GetKanbanBoardColumnResponse, len(columns))
	for i, column := range columns {
		// 컬럼의 카드 조회
		cards, err := u.boardRepo.GetBoardCardsByColumnID(column.ID)
		if err != nil {
			return nil, common.NewError(http.StatusInternalServerError, "보드 카드 조회 실패", err)
		}

		// 컬럼에 속한 카드 응답 구성
		cardsResponse := make([]res.GetKanbanBoardCardResponse, len(cards))
		for j, card := range cards {
			// 카드 할당자 조회
			assignees, err := u.boardRepo.GetCardAssignees(card.ID)
			if err != nil {
				return nil, common.NewError(http.StatusInternalServerError, "카드 할당자 조회 실패", err)
			}

			cardsResponse[j] = res.GetKanbanBoardCardResponse{
				ID:        card.ID,
				Name:      card.Name,
				Content:   card.Content,
				Position:  card.Position,
				StartDate: card.StartDate,
				EndDate:   card.EndDate,
				Version:   card.Version,
				CreatedAt: card.CreatedAt,
				UpdatedAt: card.UpdatedAt,
			}

			for _, assignee := range assignees {
				cardsResponse[j].Assignees = append(cardsResponse[j].Assignees, assignee.UserID)
			}
		}

		columnsResponse[i] = res.GetKanbanBoardColumnResponse{
			ID:        column.ID,
			Name:      column.Name,
			Position:  column.Position,
			Cards:     cardsResponse,
			CreatedAt: column.CreatedAt,
			UpdatedAt: column.UpdatedAt,
		}
	}

	// 보드 사용자 조회
	boardUsers, err := u.boardRepo.GetBoardUsersByBoardID(boardID)
	if err != nil {
		return nil, common.NewError(http.StatusInternalServerError, "보드 사용자 조회 실패", err)
	}

	// 사용자 정보 조회
	usersResponse := make([]res.GetKanbanBoardUserResponse, len(boardUsers))

	userIds := make([]uint, len(boardUsers))
	for i, boardUser := range boardUsers {
		userIds[i] = boardUser.UserID
	}

	users, err := u.userRepo.GetUserByIds(userIds)
	if err != nil {
		return nil, common.NewError(http.StatusInternalServerError, "사용자 정보 조회 실패", err)
	}

	userMap := make(map[uint]_userEntity.User)
	for _, user := range users {
		userMap[*user.ID] = user
	}

	for i, boardUser := range boardUsers {
		//! 아래 부분 db에 range로 조회하는 것이 더비효율적
		// user, err := u.userRepo.GetUserByID(boardUser.UserID)
		// if err != nil {
		// 	return nil, common.NewError(http.StatusInternalServerError, "사용자 정보 조회 실패", err)
		// }

		user, ok := userMap[boardUser.UserID]
		if !ok {
			return nil, common.NewError(http.StatusInternalServerError, "사용자 정보 조회 실패", nil)
		}

		var profileImage string
		if user.UserProfile != nil && user.UserProfile.Image != nil {
			profileImage = *user.UserProfile.Image
		} else {
			profileImage = "" // 기본값 설정
		}

		usersResponse[i] = res.GetKanbanBoardUserResponse{
			ID:           *user.ID,
			Name:         *user.Name,
			Email:        *user.Email,
			ProfileImage: profileImage,
			Role:         boardUser.Role,
		}
	}

	// 최종 응답 구성
	response := &res.GetKanbanBoardResponse{
		BoardID:       board.ID,
		Title:         board.Title,
		ProjectID:     board.ProjectID,
		ProjectName:   project.Name,
		UserBoardRole: &userBoardRole,
		CreatedAt:     board.CreatedAt,
		UpdatedAt:     board.UpdatedAt,
		Columns:       columnsResponse,
		BoardUsers:    usersResponse,
	}

	return response, nil
}
