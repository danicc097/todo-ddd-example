package testfixtures

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"

	"github.com/danicc097/todo-ddd-example/internal/modules/todo/domain"
	todoPg "github.com/danicc097/todo-ddd-example/internal/modules/todo/infrastructure/postgres"
	userDomain "github.com/danicc097/todo-ddd-example/internal/modules/user/domain"
	userPg "github.com/danicc097/todo-ddd-example/internal/modules/user/infrastructure/postgres"
	wsDomain "github.com/danicc097/todo-ddd-example/internal/modules/workspace/domain"
	wsPg "github.com/danicc097/todo-ddd-example/internal/modules/workspace/infrastructure/postgres"
	sharedPg "github.com/danicc097/todo-ddd-example/internal/shared/infrastructure/postgres"
)

// Fixtures help create randomized data for real DB testing.
type Fixtures struct {
	UserRepo      userDomain.UserRepository
	WorkspaceRepo wsDomain.WorkspaceRepository
	TodoRepo      domain.TodoRepository
	TagRepo       domain.TagRepository
}

func NewFixtures(pool *pgxpool.Pool) *Fixtures {
	uow := sharedPg.NewUnitOfWork(pool)

	return &Fixtures{
		UserRepo:      userPg.NewUserRepo(pool, uow),
		WorkspaceRepo: wsPg.NewWorkspaceRepo(pool, uow),
		TodoRepo:      todoPg.NewTodoRepo(pool, uow),
		TagRepo:       todoPg.NewTagRepo(pool),
	}
}

func (f *Fixtures) RandomUser(ctx context.Context, t *testing.T) *userDomain.User {
	t.Helper()

	uid := uuid.New().String()[:8]
	emailStr := fmt.Sprintf("user-%s@example.com", uid)
	email, err := userDomain.NewUserEmail(emailStr)
	require.NoError(t, err)

	name, _ := userDomain.NewUserName("Random User " + uid)
	user := userDomain.NewUser(email, name)
	require.NoError(t, f.UserRepo.Save(ctx, user))

	return user
}

func (f *Fixtures) RandomWorkspace(ctx context.Context, t *testing.T, ownerID userDomain.UserID) *wsDomain.Workspace {
	t.Helper()

	uid := uuid.New().String()[:8]
	name, _ := wsDomain.NewWorkspaceName("Workspace " + uid)
	desc, _ := wsDomain.NewWorkspaceDescription("Desc")
	ws := wsDomain.NewWorkspace(name, desc, ownerID)
	require.NoError(t, f.WorkspaceRepo.Save(ctx, ws))

	return ws
}

func (f *Fixtures) RandomTag(ctx context.Context, t *testing.T, wsID wsDomain.WorkspaceID) *domain.Tag {
	t.Helper()

	uid := uuid.New().String()[:8]
	name, err := domain.NewTagName("Tag-" + uid)
	require.NoError(t, err)

	tag := domain.NewTag(name, wsID)
	require.NoError(t, f.TagRepo.Save(ctx, tag))

	return tag
}

func (f *Fixtures) RandomTodo(ctx context.Context, t *testing.T, wsID wsDomain.WorkspaceID) *domain.Todo {
	t.Helper()

	uid := uuid.New().String()[:8]
	title, err := domain.NewTodoTitle("Todo " + uid)
	require.NoError(t, err)

	todo := domain.NewTodo(title, wsID)
	require.NoError(t, f.TodoRepo.Save(ctx, todo))

	return todo
}
