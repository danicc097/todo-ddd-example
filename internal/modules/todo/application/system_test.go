package application_test

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wagslane/go-rabbitmq"

	"github.com/danicc097/todo-ddd-example/internal/infrastructure/cache"
	infraDB "github.com/danicc097/todo-ddd-example/internal/infrastructure/db"
	"github.com/danicc097/todo-ddd-example/internal/infrastructure/messaging"
	"github.com/danicc097/todo-ddd-example/internal/infrastructure/outbox"
	infraRabbit "github.com/danicc097/todo-ddd-example/internal/infrastructure/rabbitmq"
	infraRedis "github.com/danicc097/todo-ddd-example/internal/infrastructure/redis"
	"github.com/danicc097/todo-ddd-example/internal/modules/todo/application"
	"github.com/danicc097/todo-ddd-example/internal/modules/todo/domain"
	todoDecorator "github.com/danicc097/todo-ddd-example/internal/modules/todo/infrastructure/decorator"
	todoPg "github.com/danicc097/todo-ddd-example/internal/modules/todo/infrastructure/postgres"
	todoRedis "github.com/danicc097/todo-ddd-example/internal/modules/todo/infrastructure/redis"
	userAdapters "github.com/danicc097/todo-ddd-example/internal/modules/user/infrastructure/adapters"
	wsApp "github.com/danicc097/todo-ddd-example/internal/modules/workspace/application"
	wsAdapters "github.com/danicc097/todo-ddd-example/internal/modules/workspace/infrastructure/adapters"
	wsPg "github.com/danicc097/todo-ddd-example/internal/modules/workspace/infrastructure/postgres"
	"github.com/danicc097/todo-ddd-example/internal/shared/causation"
	sharedDomain "github.com/danicc097/todo-ddd-example/internal/shared/domain"
	"github.com/danicc097/todo-ddd-example/internal/shared/infrastructure/middleware"
	"github.com/danicc097/todo-ddd-example/internal/testfixtures"
	"github.com/danicc097/todo-ddd-example/internal/testutils"
)

type eventCatcher struct {
	sync.Mutex

	rmq   []rabbitmq.Delivery
	redis []*redis.Message
}

func (c *eventCatcher) HasRMQ(routingKey, bodySnippet string) bool {
	c.Lock()
	defer c.Unlock()

	for _, msg := range c.rmq {
		if msg.RoutingKey == routingKey && strings.Contains(string(msg.Body), bodySnippet) {
			return true
		}
	}

	return false
}

func (c *eventCatcher) HasRedis(snippet string) bool {
	c.Lock()
	defer c.Unlock()

	for _, msg := range c.redis {
		if strings.Contains(msg.Payload, snippet) {
			return true
		}
	}

	return false
}

func setupEnv(t *testing.T) (*pgxpool.Pool, *redis.Client, *rabbitmq.Conn, *testfixtures.Fixtures, string) {
	t.Helper()

	ctx := context.Background()

	pool := testutils.GetGlobalPostgresPool(t)
	rdb := testutils.GetGlobalRedis(t).Connect(ctx, t)
	rmq := testutils.GetGlobalRabbitMQ(t)
	rmqConn := rmq.Connect(ctx, t)
	t.Cleanup(func() { _ = rmqConn.Close() })

	fixtures := testfixtures.NewFixtures(pool)

	exchangeName := "test_dist_events_" + uuid.New().String()

	rabbitPub, err := infraRabbit.NewPublisher(rmqConn, exchangeName)
	require.NoError(t, err)
	t.Cleanup(func() { rabbitPub.Close() })

	redisPub := infraRedis.NewPublisher(rdb)
	multiBroker := messaging.NewMultiBroker(rabbitPub, redisPub)

	relay := outbox.NewRelay(pool, multiBroker)

	relayCtx, relayCancel := context.WithCancel(context.Background())
	go relay.Start(relayCtx) // global for test suite

	t.Cleanup(relayCancel)

	return pool, rdb, rmqConn, fixtures, exchangeName
}

func TestSystem_Integration(t *testing.T) {
	t.Parallel()

	pool, rdb, rmqConn, fixtures, exchangeName := setupEnv(t)
	rmqContainer := testutils.GetGlobalRabbitMQ(t)

	baseTodoRepo := todoPg.NewTodoRepo(pool)
	todoCodec := todoRedis.NewTodoCacheCodec()
	cachedTodoRepo := todoDecorator.NewTodoRepositoryCache(baseTodoRepo, rdb, 5*time.Minute, todoCodec)

	baseQueryService := todoPg.NewTodoQueryService(pool)
	queryModelCodec := cache.NewMsgpackCodec[*application.TodoReadModel]()
	cachedQueryService := todoDecorator.NewTodoQueryServiceCache(baseQueryService, rdb, 5*time.Minute, queryModelCodec)

	wsRepo := wsPg.NewWorkspaceRepo(pool)
	wsProv := wsAdapters.NewTodoWorkspaceProvider(wsRepo)

	createTodoBase := application.NewCreateTodoHandler(cachedTodoRepo, wsProv)
	createTodoHandler := middleware.Transactional(pool, createTodoBase)

	completeTodoBase := application.NewCompleteTodoHandler(cachedTodoRepo, wsProv)
	completeTodoHandler := middleware.Transactional(pool, completeTodoBase)

	t.Run("success commits db invalidates cache and publishes", func(t *testing.T) {
		t.Parallel()

		testCtx, cancel := context.WithCancel(context.Background())
		t.Cleanup(cancel)

		user := fixtures.RandomUser(testCtx, t)
		ws := fixtures.RandomWorkspace(testCtx, t, user.ID())
		userCtx := causation.WithMetadata(testCtx, causation.Metadata{UserID: user.ID().UUID()})

		rmqChan, rmqConsumer := rmqContainer.StartTestConsumer(t, rmqConn, exchangeName, "topic", "#")
		t.Cleanup(func() { rmqConsumer.Close() })

		pubsub := rdb.Subscribe(testCtx, messaging.Keys.WorkspaceTodoAPIUpdatesChannel(ws.ID().UUID()))

		t.Cleanup(func() { _ = pubsub.Close() })

		catcher := &eventCatcher{}

		go func() {
			for {
				select {
				case <-testCtx.Done():
					return
				case d, ok := <-rmqChan:
					if !ok {
						return
					}

					catcher.Lock()
					catcher.rmq = append(catcher.rmq, d)
					catcher.Unlock()
				}
			}
		}()

		go func() {
			for {
				select {
				case <-testCtx.Done():
					return
				case m, ok := <-pubsub.Channel():
					if !ok {
						return
					}

					catcher.Lock()
					catcher.redis = append(catcher.redis, m)
					catcher.Unlock()
				}
			}
		}()

		cmd := application.CreateTodoCommand{
			Title:       "Task",
			WorkspaceID: ws.ID(),
		}

		resp, err := createTodoHandler.Handle(userCtx, cmd)
		require.NoError(t, err)

		saved, err := cachedTodoRepo.FindByID(testCtx, resp.ID)
		require.NoError(t, err)
		assert.Equal(t, cmd.Title, saved.Title().String())

		require.Eventually(t, func() bool {
			return rdb.Exists(testCtx, cache.Keys.Todo(resp.ID)).Val() == 1
		}, 5*time.Second, 50*time.Millisecond, "todo cache missing")

		expectedKey := messaging.Keys.EventRoutingKey(sharedDomain.TodoCreated, resp.ID.UUID())

		require.Eventually(t, func() bool {
			return catcher.HasRMQ(expectedKey, resp.ID.UUID().String())
		}, 5*time.Second, 50*time.Millisecond, "rabbitmq message missing")

		require.Eventually(t, func() bool {
			return catcher.HasRedis(resp.ID.UUID().String())
		}, 5*time.Second, 50*time.Millisecond, "redis message missing")
	})

	t.Run("rollback prevents ghost events and cache eviction", func(t *testing.T) {
		t.Parallel()

		testCtx, cancel := context.WithCancel(context.Background())
		t.Cleanup(cancel)

		user := fixtures.RandomUser(testCtx, t)
		ws := fixtures.RandomWorkspace(testCtx, t, user.ID())
		todo := fixtures.RandomTodo(testCtx, t, ws.ID())
		userCtx := causation.WithMetadata(testCtx, causation.Metadata{UserID: user.ID().UUID()})

		_, err := cachedTodoRepo.FindByID(testCtx, todo.ID())
		require.NoError(t, err)

		redisKey := cache.Keys.Todo(todo.ID())

		require.Eventually(t, func() bool {
			return rdb.Exists(testCtx, redisKey).Val() == 1
		}, 5*time.Second, 50*time.Millisecond, "cache should be warmed")

		rmqChan, rmqConsumer := rmqContainer.StartTestConsumer(t, rmqConn, exchangeName, "topic", "#")
		t.Cleanup(func() { rmqConsumer.Close() })

		catcher := &eventCatcher{}

		go func() {
			for {
				select {
				case <-testCtx.Done():
					return
				case d, ok := <-rmqChan:
					if !ok {
						return
					}

					catcher.Lock()
					catcher.rmq = append(catcher.rmq, d)
					catcher.Unlock()
				}
			}
		}()

		_, err = infraDB.RunInTx(testCtx, pool, func(txCtx context.Context) (any, error) {
			require.NoError(t, todo.Complete())
			_ = cachedTodoRepo.Save(txCtx, todo)

			return nil, errors.New("simulated fault")
		})
		require.ErrorContains(t, err, "simulated fault")

		// guarantees the outbox relay has processed past our rollback attempt
		barrierCmd := application.CreateTodoCommand{Title: "Barrier", WorkspaceID: ws.ID()}
		barrierResp, err := createTodoHandler.Handle(userCtx, barrierCmd)
		require.NoError(t, err)

		barrierKey := messaging.Keys.EventRoutingKey(sharedDomain.TodoCreated, barrierResp.ID.UUID())

		require.Eventually(t, func() bool {
			return catcher.HasRMQ(barrierKey, barrierResp.ID.UUID().String())
		}, 5*time.Second, 50*time.Millisecond, "timeout waiting for barrier event")

		fresh, _ := baseTodoRepo.FindByID(testCtx, todo.ID())
		assert.Equal(t, domain.StatusPending, fresh.Status())

		require.Equal(t, int64(1), rdb.Exists(testCtx, redisKey).Val())

		expectedRolledBackKey := messaging.Keys.EventRoutingKey(sharedDomain.TodoCompleted, todo.ID().UUID())
		assert.False(t, catcher.HasRMQ(expectedRolledBackKey, todo.ID().UUID().String()), "rolled back event was published")
	})

	t.Run("entity update triggers targeted and collection cache invalidation", func(t *testing.T) {
		t.Parallel()

		testCtx, cancel := context.WithCancel(context.Background())
		t.Cleanup(cancel)

		user := fixtures.RandomUser(testCtx, t)
		ws := fixtures.RandomWorkspace(testCtx, t, user.ID())
		todo := fixtures.RandomTodo(testCtx, t, ws.ID())
		userCtx := causation.WithMetadata(testCtx, causation.Metadata{UserID: user.ID().UUID()})

		_, err := cachedTodoRepo.FindByID(testCtx, todo.ID())
		require.NoError(t, err)

		_, err = cachedQueryService.GetAllByWorkspace(testCtx, ws.ID(), 10, 0)
		require.NoError(t, err)

		entityRedisKey := cache.Keys.Todo(todo.ID())
		collectionRedisKey := cache.Keys.TodoWorkspaceCollectionPaginated(ws.ID(), 10, 0)

		require.Eventually(t, func() bool {
			return rdb.Exists(testCtx, entityRedisKey).Val() == 1 && rdb.Exists(testCtx, collectionRedisKey).Val() == 1
		}, 5*time.Second, 50*time.Millisecond, "caches should be primed")

		_, err = completeTodoHandler.Handle(userCtx, application.CompleteTodoCommand{ID: todo.ID()})
		require.NoError(t, err)

		require.Eventually(t, func() bool {
			return rdb.Exists(testCtx, entityRedisKey).Val() == 0 && rdb.Exists(testCtx, collectionRedisKey).Val() == 0
		}, 5*time.Second, 50*time.Millisecond, "caches should be fully invalidated")
	})

	t.Run("atomic multi event outbox publishing", func(t *testing.T) {
		t.Parallel()

		testCtx, cancel := context.WithCancel(context.Background())
		t.Cleanup(cancel)

		user := fixtures.RandomUser(testCtx, t)
		userCtx := causation.WithMetadata(testCtx, causation.Metadata{UserID: user.ID().UUID()})

		rmqChan, rmqConsumer := rmqContainer.StartTestConsumer(t, rmqConn, exchangeName, "topic", "#")
		t.Cleanup(func() { rmqConsumer.Close() })

		catcher := &eventCatcher{}

		go func() {
			for {
				select {
				case <-testCtx.Done():
					return
				case d, ok := <-rmqChan:
					if !ok {
						return
					}

					catcher.Lock()
					catcher.rmq = append(catcher.rmq, d)
					catcher.Unlock()
				}
			}
		}()

		userProv := userAdapters.NewWorkspaceUserProvider(fixtures.UserRepo)
		onboardHandler := middleware.Transactional(pool, wsApp.NewOnboardWorkspaceHandler(wsRepo, userProv))

		cmd := wsApp.OnboardWorkspaceCommand{
			Name:    "Atomic ws",
			OwnerID: user.ID(),
		}

		resp, err := onboardHandler.Handle(userCtx, cmd)
		require.NoError(t, err)

		wsKey := messaging.Keys.EventRoutingKey(sharedDomain.WorkspaceCreated, resp.ID.UUID())
		memKey := messaging.Keys.EventRoutingKey(sharedDomain.WorkspaceMemberAdded, resp.ID.UUID())

		require.Eventually(t, func() bool {
			return catcher.HasRMQ(wsKey, resp.ID.UUID().String()) && catcher.HasRMQ(memKey, resp.ID.UUID().String())
		}, 5*time.Second, 50*time.Millisecond, "failed to receive multiple events for aggregate")
	})
}
