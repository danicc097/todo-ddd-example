package application

import (
	"context"

	"github.com/google/uuid"

	"github.com/danicc097/todo-ddd-example/internal/modules/todo/domain"
)

//go:generate go tool gowrap gen -g -i CreateTagUseCase -t ../../../../templates/transactional.gotmpl -o ../infrastructure/decorator/create_tag_tx.gen.go
type CreateTagUseCase interface {
	Execute(ctx context.Context, name string) (uuid.UUID, error)
}

type createTagUseCase struct {
	repo domain.TagRepository
}

func NewCreateTagUseCase(repo domain.TagRepository) CreateTagUseCase {
	return &createTagUseCase{repo: repo}
}

func (uc *createTagUseCase) Execute(ctx context.Context, name string) (uuid.UUID, error) {
	tn, err := domain.NewTagName(name)
	if err != nil {
		return uuid.UUID{}, err
	}

	tag := domain.NewTag(tn)

	if err := uc.repo.Save(ctx, tag); err != nil {
		return uuid.UUID{}, err
	}

	return tag.ID(), nil
}
