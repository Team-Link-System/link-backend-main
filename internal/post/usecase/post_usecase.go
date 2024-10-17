package usecase

import (
	"link/internal/post/entity"
	"link/internal/post/repository"
)

type PostUsecase struct {
	postRepository repository.PostRepository
}

func NewPostUsecase(postRepository repository.PostRepository) *PostUsecase {
	return &PostUsecase{postRepository: postRepository}
}

func (uc *PostUsecase) CreatePost(post *entity.Post) error {
	return uc.postRepository.CreatePost(post)
}
