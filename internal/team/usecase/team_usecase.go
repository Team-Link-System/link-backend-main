package usecase

// import (
// 	"link/internal/team/repository"
// )

// type TeamUsecase interface {
// 	// GetTeamsByCompany(userId uint) ([]res.GetTeamsByCompanyResponse, error)
// }

// type teamUsecase struct {
// 	teamRepository repository.TeamRepository
// }

// func NewTeamUsecase(teamRepository repository.TeamRepository) *TeamUsecase {
// 	return &TeamUsecase{teamRepository: teamRepository}
// }

// // func (u *teamUsecase) GetTeamsByCompany(userId uint) ([]res.GetTeamsByCompanyResponse, error) {

// // 	teams, err := u.teamRepository.GetTeamsByCompany(userId)
// // 	if err != nil {
// // 		return nil, err
// // 	}

// // 	response := make([]res.GetTeamsByCompanyResponse, len(teams))

// // 	for i, team := range teams {
// // 		response[i] = res.GetTeamsByCompanyResponse{
// // 			ID:   team.ID,
// // 			Name: team.Name,
// // 		}
// // 	}
// // 	return response, nil
// // }
