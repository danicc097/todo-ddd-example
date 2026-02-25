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

	onboardWsHandler := wsApp.NewOnboardWorkspaceHandler(wsRepo, wsUserProvider, uow)
	addWsMemberHandler := wsApp.NewAddWorkspaceMemberHandler(wsRepo, uow)
	removeWsMemberHandler := wsApp.NewRemoveWorkspaceMemberHandler(wsRepo, uow)
	deleteWsHandler := wsApp.NewDeleteWorkspaceHandler(wsRepo, uow)

	baseWorkspaceQueryService := wsPg.NewWorkspaceQueryService(container.Pool)
	workspaceQueryService := wsPg.NewWorkspaceQueryServiceWithTracing(baseWorkspaceQueryService, messaging.Keys.ServiceName())

	wsProv := wsAdapters.NewTodoWorkspaceProvider(wsRepo)

	hub := ws.NewTodoHub(container.Redis, workspaceQueryService)

	createTodoHandler := todoApp.NewCreateTodoHandler(todoRepo, wsProv, uow)
	completeTodoHandler := todoApp.NewCompleteTodoHandler(todoRepo, wsProv, uow)
	createTagHandler := todoApp.NewCreateTagHandler(tagRepo, uow)
	assignTagToTodoHandler := todoApp.NewAssignTagToTodoHandler(todoRepo, tagRepo, uow)
	startFocusHandler := todoApp.NewStartFocusHandler(todoRepo, wsProv, uow)
	stopFocusHandler := todoApp.NewStopFocusHandler(todoRepo, wsProv, uow)

	scheduleRepo := schedulePg.NewScheduleRepo(container.Pool, uow)
	commitTaskHandler := sharedApp.Retry(
		scheduleApp.NewCommitTaskHandler(scheduleRepo, todoRepo, uow),
		3,
	)

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

	loginHandler := authApp.NewLoginHandler(userRepo, authRepo, tokenProvider.Issuer, passwordHasher)
	registerHandler := authApp.NewRegisterHandler(userRepo, authRepo, passwordHasher, uow)

	initiateTOTPHandler := authApp.NewInitiateTOTPHandler(authRepo, masterKey)
	verifyTOTPHandler := authApp.NewVerifyTOTPHandler(authRepo, totpGuard, tokenProvider.Issuer, masterKey)

	th := todoHttp.NewTodoHandler(
		createTodoHandler,
		completeTodoHandler,
		createTagHandler,
		assignTagToTodoHandler,
		startFocusHandler,
		stopFocusHandler,
		todoQueryService,
		hub,
		container.Redis,
	)

	uh := userHttp.NewUserHandler(
		getUserUC,
		workspaceQueryService,
	)

	wh := wsHttp.NewWorkspaceHandler(
		onboardWsHandler,
		addWsMemberHandler,
		removeWsMemberHandler,
		workspaceQueryService,
		deleteWsHandler,
	)

	ah := authHttp.NewAuthHandler(loginHandler, registerHandler, initiateTOTPHandler, verifyTOTPHandler)

	sh := scheduleHttp.NewScheduleHandler(commitTaskHandler)

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
