package application

import (
	"context"

	"github.com/google/uuid"

	"github.com/danicc097/todo-ddd-example/internal/infrastructure/db"
	"github.com/danicc097/todo-ddd-example/internal/modules/todo/domain"
)

type CreateTagUseCase struct {
	tm db.TransactionManager
}

func NewCreateTagUseCase(tm db.TransactionManager) *CreateTagUseCase {
	return &CreateTagUseCase{tm: tm}
}

func (uc *CreateTagUseCase) Execute(ctx context.Context, name string) (uuid.UUID, error) {
	tn, err := domain.NewTagName(name)
	if err != nil {
		return uuid.UUID{}, err
	}

	tag := domain.NewTag(tn)

	err = uc.tm.Exec(ctx, func(p db.RepositoryProvider) error {
		return p.Tag().Save(ctx, tag)
	})
	if err != nil {
		return uuid.UUID{}, err
	}

	return tag.ID(), nil
}
