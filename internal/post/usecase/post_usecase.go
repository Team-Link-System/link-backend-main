package usecase

import (
	"link/internal/post/entity"
	"link/internal/post/repository"
)

type PostUsecase interface {
	CreatePost(post *entity.Post) error
}

type postUsecase struct {
	postRepo repository.PostRepository
}

func NewPostUsecase(postRepo repository.PostRepository) PostUsecase {
	return &postUsecase{postRepo: postRepo}
}

func (uc *postUsecase) CreatePost(post *entity.Post) error {
	uc.postRepo.CreatePost(post)
	return nil
}
