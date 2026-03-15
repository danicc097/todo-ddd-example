package application

import "github.com/danicc097/todo-ddd-example/internal/shared/application"

type TodoQueryService interface {
	Save() // want "Arch violation: QueryService interface TodoQueryService has mutating method Save. Queries must be read-only."
	Get()
}

type MyQueryService struct {
	uow application.UnitOfWork // want "Arch violation: MyQueryService cannot have UnitOfWork as a field. Queries must be read-only."
}
