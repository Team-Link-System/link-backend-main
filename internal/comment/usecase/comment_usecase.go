package usecase

type CommentUsecase interface {
	CreateComment(req req.CommentRequest) (*res.CreateCommentResponse, error)
}

type commentUsecase struct {
	commentRepo _commentRepo.CommentRepository
}

func NewCommentUsecase(commentRepo _commentRepo.CommentRepository) CommentUsecase {
	return &commentUsecase{commentRepo: commentRepo}
}

func (u *commentUsecase) CreateComment(req req.CommentRequest) (*res.CreateCommentResponse, error) {

	return nil, nil
}
