package infrastructure

import (
	"context"
	"time"

	"github.com/danicc097/todo-ddd-example/internal"
	"github.com/danicc097/todo-ddd-example/internal/infrastructure/cache"
	"github.com/danicc097/todo-ddd-example/internal/infrastructure/crypto"
	"github.com/danicc097/todo-ddd-example/internal/infrastructure/messaging"
	auditMem "github.com/danicc097/todo-ddd-example/internal/modules/audit/infrastructure/memory"
	authApp "github.com/danicc097/todo-ddd-example/internal/modules/auth/application"
	authHttp "github.com/danicc097/todo-ddd-example/internal/modules/auth/infrastructure/http"
	authPg "github.com/danicc097/todo-ddd-example/internal/modules/auth/infrastructure/postgres"
	authRedis "github.com/danicc097/todo-ddd-example/internal/modules/auth/infrastructure/redis"
	scheduleApp "github.com/danicc097/todo-ddd-example/internal/modules/schedule/application"
	scheduleHttp "github.com/danicc097/todo-ddd-example/internal/modules/schedule/infrastructure/http"
	schedulePg "github.com/danicc097/todo-ddd-example/internal/modules/schedule/infrastructure/postgres"
	todoApp "github.com/danicc097/todo-ddd-example/internal/modules/todo/application"
	todoDecorator "github.com/danicc097/todo-ddd-example/internal/modules/todo/infrastructure/decorator"
	todoHttp "github.com/danicc097/todo-ddd-example/internal/modules/todo/infrastructure/http"
	todoPg "github.com/danicc097/todo-ddd-example/internal/modules/todo/infrastructure/postgres"
	todoRedis "github.com/danicc097/todo-ddd-example/internal/modules/todo/infrastructure/redis"
	"github.com/danicc097/todo-ddd-example/internal/modules/todo/infrastructure/ws"
	userApp "github.com/danicc097/todo-ddd-example/internal/modules/user/application"
	userAdapters "github.com/danicc097/todo-ddd-example/internal/modules/user/infrastructure/adapters"
	userHttp "github.com/danicc097/todo-ddd-example/internal/modules/user/infrastructure/http"
	userPg "github.com/danicc097/todo-ddd-example/internal/modules/user/infrastructure/postgres"
	wsApp "github.com/danicc097/todo-ddd-example/internal/modules/workspace/application"
	wsAdapters "github.com/danicc097/todo-ddd-example/internal/modules/workspace/infrastructure/adapters"
	wsDecorator "github.com/danicc097/todo-ddd-example/internal/modules/workspace/infrastructure/decorator"
	wsHttp "github.com/danicc097/todo-ddd-example/internal/modules/workspace/infrastructure/http"
	wsPg "github.com/danicc097/todo-ddd-example/internal/modules/workspace/infrastructure/postgres"
	wsRedis "github.com/danicc097/todo-ddd-example/internal/modules/workspace/infrastructure/redis"
	sharedApp "github.com/danicc097/todo-ddd-example/internal/shared/application"
	sharedPg "github.com/danicc097/todo-ddd-example/internal/shared/infrastructure/postgres"
)

type Services struct {
	TodoHandler      *todoHttp.TodoHandler
	UserHandler      *userHttp.UserHandler
	WorkspaceHandler *wsHttp.WorkspaceHandler
	AuthHandler      *authHttp.AuthHandler
	ScheduleHandler  *scheduleHttp.ScheduleHandler

	TodoRepo              *todoPg.TodoRepo
	ScheduleRepo          *schedulePg.ScheduleRepo
	TokenProvider         *crypto.TokenProvider
	WorkspaceQueryService wsApp.WorkspaceQueryService
}

func NewServices(ctx context.Context, cfg *internal.AppConfig, container *Container) (*Services, error) {
	uow := sharedPg.NewUnitOfWork(container.Pool)

	baseTodoRepo := todoPg.NewTodoRepo(container.Pool, uow)
	todoCodec := todoRedis.NewTodoCacheCodec()
	cachedTodoRepo := todoDecorator.NewTodoRepositoryCache(
		baseTodoRepo,
		container.Redis,
		5*time.Minute,
		todoCodec,
	)
	todoRepo := todoPg.NewTodoRepositoryWithTracing(cachedTodoRepo, messaging.Keys.ServiceName())

	baseTagRepo := todoPg.NewTagRepo(container.Pool)
	tagCodec := todoRedis.NewTagCacheCodec()
	cachedTagRepo := todoDecorator.NewTagRepositoryCache(
		baseTagRepo,
		container.Redis,
		60*time.Minute,
		tagCodec,
	)
	tagRepo := todoPg.NewTagRepositoryWithTracing(cachedTagRepo, messaging.Keys.ServiceName())

	baseUserRepo := userPg.NewUserRepo(container.Pool, uow)
	userRepo := userPg.NewUserRepositoryWithTracing(baseUserRepo, messaging.Keys.ServiceName())

	auditRepo := auditMem.NewAuditRepository()
	baseWsRepo := wsPg.NewWorkspaceRepo(container.Pool, uow)
	wsCodec := wsRedis.NewWorkspaceCacheCodec()
	cachedWsRepo := wsDecorator.NewWorkspaceRepositoryCache(
		baseWsRepo,
		container.Redis,
		10*time.Minute,
		wsCodec,
	)
	auditedWsRepo := wsDecorator.NewWorkspaceAuditWrapper(cachedWsRepo, auditRepo)
	wsRepo := wsPg.NewWorkspaceRepositoryWithTracing(auditedWsRepo, messaging.Keys.ServiceName())

	tokenProvider, err := crypto.NewTokenProvider("private.pem", "public.pem", messaging.Keys.ServiceName())
	if err != nil {
		return nil, err
	}

	baseAuthRepo := authPg.NewAuthRepo(container.Pool, uow)
	authRepo := authPg.NewAuthRepositoryWithTracing(baseAuthRepo, messaging.Keys.ServiceName())

	masterKey := []byte(cfg.MFAMasterKey)
	totpGuard := authRedis.NewTOTPGuard(container.Redis)

	getUserUC := userApp.NewGetUserUseCase(userRepo)

	wsUserProvider := userAdapters.NewWorkspaceUserProvider(userRepo)

	onboardWsHandler := sharedApp.NewCommandDecoratorBuilder(
		wsApp.NewOnboardWorkspaceHandler(wsRepo, wsUserProvider), uow, "onboard-workspace",
	).Build()
	addWsMemberHandler := sharedApp.NewCommandDecoratorBuilder(
		wsApp.NewAddWorkspaceMemberHandler(wsRepo), uow, "add-workspace-member",
	).Build()
	removeWsMemberHandler := sharedApp.NewCommandDecoratorBuilder(
		wsApp.NewRemoveWorkspaceMemberHandler(wsRepo), uow, "remove-workspace-member",
	).Build()
	deleteWsHandler := sharedApp.NewCommandDecoratorBuilder(
		wsApp.NewDeleteWorkspaceHandler(wsRepo), uow, "delete-workspace",
	).Build()

	baseWorkspaceQueryService := wsPg.NewWorkspaceQueryService(container.Pool)
	workspaceQueryService := wsPg.NewWorkspaceQueryServiceWithTracing(baseWorkspaceQueryService, messaging.Keys.ServiceName())

	wsProv := wsAdapters.NewTodoWorkspaceProvider(wsRepo)

	hub := ws.NewTodoHub(container.Redis, workspaceQueryService)

	createTodoHandler := sharedApp.NewCommandDecoratorBuilder(
		todoApp.NewCreateTodoHandler(todoRepo, wsProv), uow, "create-todo",
	).Build()
	completeTodoHandler := sharedApp.NewCommandDecoratorBuilder(
		todoApp.NewCompleteTodoHandler(todoRepo, wsProv), uow, "complete-todo",
	).Build()
	createTagHandler := sharedApp.NewCommandDecoratorBuilder(
		todoApp.NewCreateTagHandler(tagRepo), uow, "create-tag",
	).Build()
	assignTagToTodoHandler := sharedApp.NewCommandDecoratorBuilder(
		todoApp.NewAssignTagToTodoHandler(todoRepo, tagRepo), uow, "assign-tag-to-todo",
	).Build()
	startFocusHandler := sharedApp.NewCommandDecoratorBuilder(
		todoApp.NewStartFocusHandler(todoRepo, wsProv), uow, "start-focus",
	).Build()
	stopFocusHandler := sharedApp.NewCommandDecoratorBuilder(
		todoApp.NewStopFocusHandler(todoRepo, wsProv), uow, "stop-focus",
	).Build()

	scheduleRepo := schedulePg.NewScheduleRepo(container.Pool, uow)
	commitTaskHandler := sharedApp.NewCommandDecoratorBuilder(
		scheduleApp.NewCommitTaskHandler(scheduleRepo, todoRepo), uow, "commit-task",
	).WithRetryOnConflict(3).Build()

	baseTodoQueryService := todoPg.NewTodoQueryService(container.Pool)
	todoReadModelCodec := cache.NewMsgpackCodec[*todoApp.TodoReadModel]()
	cachedTodoQueryService := todoDecorator.NewTodoQueryServiceCache(
		baseTodoQueryService,
		container.Redis,
		5*time.Minute,
		todoReadModelCodec,
	)
	todoQueryService := todoPg.NewTodoQueryServiceWithTracing(cachedTodoQueryService, messaging.Keys.ServiceName())

	passwordHasher := crypto.NewArgon2PasswordHasher()

	loginHandler := sharedApp.NewDecoratorBuilder(authApp.NewLoginHandler(userRepo, authRepo, tokenProvider.Issuer, passwordHasher)).
		WithLogging("login").
		Build()
	registerHandler := sharedApp.NewCommandDecoratorBuilder(
		authApp.NewRegisterHandler(userRepo, authRepo, passwordHasher), uow, "register",
	).Build()
	initiateTOTPHandler := sharedApp.NewCommandDecoratorBuilder(
		authApp.NewInitiateTOTPHandler(authRepo, masterKey), uow, "initiate-totp",
	).Build()
	verifyTOTPHandler := sharedApp.NewCommandDecoratorBuilder(
		authApp.NewVerifyTOTPHandler(authRepo, totpGuard, tokenProvider.Issuer, masterKey), uow, "verify-totp",
	).Build()

	th := todoHttp.NewTodoHandler(
		todoHttp.TodoUseCases{
			CreateTodo: createTodoHandler,
			Complete:   completeTodoHandler,
			CreateTag:  createTagHandler,
			AssignTag:  assignTagToTodoHandler,
			StartFocus: startFocusHandler,
			StopFocus:  stopFocusHandler,
		},
		todoQueryService,
		hub,
		container.Redis,
	)

	uh := userHttp.NewUserHandler(
		getUserUC,
		workspaceQueryService,
	)

	wh := wsHttp.NewWorkspaceHandler(
		wsHttp.WorkspaceUseCases{
			Onboard:      onboardWsHandler,
			AddMember:    addWsMemberHandler,
			RemoveMember: removeWsMemberHandler,
			Delete:       deleteWsHandler,
		},
		workspaceQueryService,
	)

	ah := authHttp.NewAuthHandler(authHttp.AuthUseCases{
		Login:        loginHandler,
		Register:     registerHandler,
		InitiateTOTP: initiateTOTPHandler,
		VerifyTOTP:   verifyTOTPHandler,
	})

	sh := scheduleHttp.NewScheduleHandler(scheduleHttp.ScheduleUseCases{
		CommitTask: commitTaskHandler,
	})

	return &Services{
		TodoHandler:           th,
		UserHandler:           uh,
		WorkspaceHandler:      wh,
		AuthHandler:           ah,
		ScheduleHandler:       sh,
		TodoRepo:              baseTodoRepo,
		ScheduleRepo:          scheduleRepo,
		TokenProvider:         tokenProvider,
		WorkspaceQueryService: workspaceQueryService,
	}, nil
}
