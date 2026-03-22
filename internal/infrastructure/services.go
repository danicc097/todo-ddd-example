package infrastructure

import (
	"context"
	"time"

	"github.com/danicc097/todo-ddd-example/internal"
	"github.com/danicc097/todo-ddd-example/internal/infrastructure/cache"
	"github.com/danicc097/todo-ddd-example/internal/infrastructure/crypto"
	"github.com/danicc097/todo-ddd-example/internal/infrastructure/messaging"
	infraRedis "github.com/danicc097/todo-ddd-example/internal/infrastructure/redis"
	auditMem "github.com/danicc097/todo-ddd-example/internal/modules/audit/infrastructure/memory"
	authApp "github.com/danicc097/todo-ddd-example/internal/modules/auth/application"
	authDomain "github.com/danicc097/todo-ddd-example/internal/modules/auth/domain"
	authAdapters "github.com/danicc097/todo-ddd-example/internal/modules/auth/infrastructure/adapters"
	authPg "github.com/danicc097/todo-ddd-example/internal/modules/auth/infrastructure/postgres"
	authRedis "github.com/danicc097/todo-ddd-example/internal/modules/auth/infrastructure/redis"
	scheduleApp "github.com/danicc097/todo-ddd-example/internal/modules/schedule/application"
	scheduleDomain "github.com/danicc097/todo-ddd-example/internal/modules/schedule/domain"
	schedulePg "github.com/danicc097/todo-ddd-example/internal/modules/schedule/infrastructure/postgres"
	todoApp "github.com/danicc097/todo-ddd-example/internal/modules/todo/application"
	todoDomain "github.com/danicc097/todo-ddd-example/internal/modules/todo/domain"
	todoDecorator "github.com/danicc097/todo-ddd-example/internal/modules/todo/infrastructure/decorator"
	todoPg "github.com/danicc097/todo-ddd-example/internal/modules/todo/infrastructure/postgres"
	todoRedis "github.com/danicc097/todo-ddd-example/internal/modules/todo/infrastructure/redis"
	userApp "github.com/danicc097/todo-ddd-example/internal/modules/user/application"
	userDomain "github.com/danicc097/todo-ddd-example/internal/modules/user/domain"
	userAdapters "github.com/danicc097/todo-ddd-example/internal/modules/user/infrastructure/adapters"
	userPg "github.com/danicc097/todo-ddd-example/internal/modules/user/infrastructure/postgres"
	wsApp "github.com/danicc097/todo-ddd-example/internal/modules/workspace/application"
	wsDomain "github.com/danicc097/todo-ddd-example/internal/modules/workspace/domain"
	wsAdapters "github.com/danicc097/todo-ddd-example/internal/modules/workspace/infrastructure/adapters"
	wsDecorator "github.com/danicc097/todo-ddd-example/internal/modules/workspace/infrastructure/decorator"
	wsPg "github.com/danicc097/todo-ddd-example/internal/modules/workspace/infrastructure/postgres"
	wsRedis "github.com/danicc097/todo-ddd-example/internal/modules/workspace/infrastructure/redis"
	sharedApp "github.com/danicc097/todo-ddd-example/internal/shared/application"
	sharedPg "github.com/danicc097/todo-ddd-example/internal/shared/infrastructure/postgres"
)

type Services struct {
	Todo      todoApp.TodoUseCases
	Workspace wsApp.WorkspaceUseCases
	Auth      authApp.AuthUseCases
	Schedule  scheduleApp.ScheduleUseCases

	UserQuery      *userApp.GetUserUseCase
	TodoQuery      todoApp.TodoQueryService
	WorkspaceQuery wsApp.WorkspaceQueryService

	ScheduleRepo  scheduleDomain.ScheduleRepository
	TokenProvider *crypto.TokenProvider
}

func NewServices(ctx context.Context, cfg *internal.AppConfig, cnt *Container) (*Services, error) {
	uow, svcName := sharedPg.NewUnitOfWork(cnt.Pool), messaging.Keys.ServiceName()
	tokenProvider, _ := crypto.NewTokenProvider("private.pem", "public.pem", svcName)
	hasher, audit, totp := crypto.NewArgon2PasswordHasher(), auditMem.NewAuditRepository(), authRedis.NewTOTPGuard(cnt.Redis)
	cacheStore := infraRedis.NewCacheStore(cnt.Redis)
	encryptor := authAdapters.NewAESGCMEncryptor()
	appConfig := authAdapters.NewMessagingAppConfig()

	/** Repositories **/
	userRepo := sharedApp.Apply(userDomain.UserRepository(userPg.NewUserRepo(cnt.Pool, uow)),
		func(r userDomain.UserRepository) userDomain.UserRepository {
			return userPg.NewUserRepositoryWithTracing(r, svcName)
		})

	authRepo := sharedApp.Apply(authDomain.AuthRepository(authPg.NewAuthRepo(cnt.Pool, uow)),
		func(r authDomain.AuthRepository) authDomain.AuthRepository {
			return authPg.NewAuthRepositoryWithTracing(r, svcName)
		})

	todoRepo := sharedApp.Apply(todoDomain.TodoRepository(todoPg.NewTodoRepo(cnt.Pool, uow)),
		func(r todoDomain.TodoRepository) todoDomain.TodoRepository {
			return todoDecorator.NewTodoRepositoryCache(r, cacheStore, 5*time.Minute, todoRedis.NewTodoCacheCodec())
		},
		func(r todoDomain.TodoRepository) todoDomain.TodoRepository {
			return todoPg.NewTodoRepositoryWithTracing(r, svcName)
		})

	tagRepo := sharedApp.Apply(todoDomain.TagRepository(todoPg.NewTagRepo(cnt.Pool)),
		func(r todoDomain.TagRepository) todoDomain.TagRepository {
			return todoDecorator.NewTagRepositoryCache(r, cacheStore, 60*time.Minute, todoRedis.NewTagCacheCodec())
		},
		func(r todoDomain.TagRepository) todoDomain.TagRepository {
			return todoPg.NewTagRepositoryWithTracing(r, svcName)
		})

	wsRepo := sharedApp.Apply(wsDomain.WorkspaceRepository(wsPg.NewWorkspaceRepo(cnt.Pool, uow)),
		func(r wsDomain.WorkspaceRepository) wsDomain.WorkspaceRepository {
			return wsDecorator.NewWorkspaceRepositoryCache(r, cacheStore, 10*time.Minute, wsRedis.NewWorkspaceCacheCodec())
		},
		func(r wsDomain.WorkspaceRepository) wsDomain.WorkspaceRepository {
			return wsDecorator.NewWorkspaceAuditWrapper(r, audit)
		},
		func(r wsDomain.WorkspaceRepository) wsDomain.WorkspaceRepository {
			return wsPg.NewWorkspaceRepositoryWithTracing(r, svcName)
		})

	scheduleRepo := sharedApp.Apply(scheduleDomain.ScheduleRepository(schedulePg.NewScheduleRepo(cnt.Pool, uow)),
		func(r scheduleDomain.ScheduleRepository) scheduleDomain.ScheduleRepository { return r })

	/** Query **/
	wsQuery := sharedApp.Apply(wsPg.NewWorkspaceQueryService(cnt.Pool),
		func(qs wsApp.WorkspaceQueryService) wsApp.WorkspaceQueryService {
			return wsPg.NewWorkspaceQueryServiceWithTracing(qs, svcName)
		})

	todoQuery := sharedApp.Apply(todoPg.NewTodoQueryService(cnt.Pool),
		func(qs todoApp.TodoQueryService) todoApp.TodoQueryService {
			return todoDecorator.NewTodoQueryServiceCache(qs, cacheStore, 5*time.Minute, cache.NewMsgpackCodec[*todoApp.TodoReadModel]())
		},
		func(qs todoApp.TodoQueryService) todoApp.TodoQueryService {
			return todoPg.NewTodoQueryServiceWithTracing(qs, svcName)
		})

	/** Wiring **/
	wsProv := wsAdapters.NewTodoWorkspaceProvider(wsRepo)
	wsUserProv := userAdapters.NewWorkspaceUserProvider(userRepo)

	return &Services{
		Todo: todoApp.TodoUseCases{
			CreateTodo: sharedApp.BuildCommand(todoApp.NewCreateTodoHandler(todoRepo, wsProv), uow, "create-todo"),
			Complete:   sharedApp.BuildCommand(todoApp.NewCompleteTodoHandler(todoRepo, wsProv), uow, "complete-todo"),
			CreateTag:  sharedApp.BuildCommand(todoApp.NewCreateTagHandler(tagRepo), uow, "create-tag"),
			AssignTag:  sharedApp.BuildCommand(todoApp.NewAssignTagToTodoHandler(todoRepo, tagRepo), uow, "assign-tag-to-todo"),
			StartFocus: sharedApp.BuildCommand(todoApp.NewStartFocusHandler(todoRepo, wsProv), uow, "start-focus"),
			StopFocus:  sharedApp.BuildCommand(todoApp.NewStopFocusHandler(todoRepo, wsProv), uow, "stop-focus"),
		},
		Workspace: wsApp.WorkspaceUseCases{
			Onboard:      sharedApp.BuildCommand(wsApp.NewOnboardWorkspaceHandler(wsRepo, wsUserProv), uow, "onboard-workspace"),
			AddMember:    sharedApp.BuildCommand(wsApp.NewAddWorkspaceMemberHandler(wsRepo), uow, "add-workspace-member"),
			RemoveMember: sharedApp.BuildCommand(wsApp.NewRemoveWorkspaceMemberHandler(wsRepo), uow, "remove-workspace-member"),
			Delete:       sharedApp.BuildCommand(wsApp.NewDeleteWorkspaceHandler(wsRepo), uow, "delete-workspace"),
		},
		Auth: authApp.AuthUseCases{
			Login:        sharedApp.BuildQuery(authApp.NewLoginHandler(userRepo, authRepo, tokenProvider.Issuer, hasher), "login"),
			Register:     sharedApp.BuildCommand(authApp.NewRegisterHandler(userRepo, authRepo, hasher), uow, "register"),
			InitiateTOTP: sharedApp.BuildCommand(authApp.NewInitiateTOTPHandler(authRepo, encryptor, appConfig, []byte(cfg.MFAMasterKey)), uow, "initiate-totp"),
			VerifyTOTP:   sharedApp.BuildCommand(authApp.NewVerifyTOTPHandler(authRepo, totp, tokenProvider.Issuer, encryptor, []byte(cfg.MFAMasterKey)), uow, "verify-totp"),
		},
		Schedule: scheduleApp.ScheduleUseCases{
			CommitTask: sharedApp.BuildCommand(scheduleApp.NewCommitTaskHandler(scheduleRepo, todoRepo), uow, "commit-task"),
		},
		UserQuery:      userApp.NewGetUserUseCase(userRepo),
		TodoQuery:      todoQuery,
		WorkspaceQuery: wsQuery,
		ScheduleRepo:   scheduleRepo,
		TokenProvider:  tokenProvider,
	}, nil
}
