package infrastructure

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/wagslane/go-rabbitmq"

	"github.com/danicc097/todo-ddd-example/internal"
	infraDB "github.com/danicc097/todo-ddd-example/internal/infrastructure/db"
	infraMessaging "github.com/danicc097/todo-ddd-example/internal/infrastructure/messaging"
	infraRabbit "github.com/danicc097/todo-ddd-example/internal/infrastructure/rabbitmq"
	infraRedis "github.com/danicc097/todo-ddd-example/internal/infrastructure/redis"
)

type Container struct {
	Pool        *pgxpool.Pool
	Redis       *redis.Client
	MQConn      *rabbitmq.Conn
	MultiBroker infraMessaging.Broker
}

func NewContainer(ctx context.Context, cfg *internal.AppConfig) (*Container, func(), error) {
	pool, err := infraDB.NewConnectionPool(ctx, infraDB.Config{
		User:     cfg.Postgres.User,
		Password: cfg.Postgres.Password,
		Host:     cfg.Postgres.Host,
		Port:     cfg.Postgres.Port,
		DBName:   cfg.Postgres.DBName,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	redisClient := infraRedis.NewClient(cfg.Redis.Addr)

	mqConn, err := infraRabbit.NewConnection(cfg.RabbitMQ.URL)
	if err != nil {
		pool.Close()
		redisClient.Close()

		return nil, nil, fmt.Errorf("failed to connect to rabbitmq: %w", err)
	}

	rabbitPub, err := infraRabbit.NewPublisher(mqConn, infraMessaging.Keys.TodoEventsExchange())
	if err != nil {
		pool.Close()
		redisClient.Close()
		mqConn.Close()

		return nil, nil, fmt.Errorf("failed to create rabbitmq publisher: %w", err)
	}

	redisPub := infraRedis.NewPublisher(redisClient)
	multiBroker := infraMessaging.NewMultiBroker(rabbitPub, redisPub)

	cleanup := func() {
		pool.Close()
		redisClient.Close()
		rabbitPub.Close()
		mqConn.Close()
	}

	return &Container{
		Pool:        pool,
		Redis:       redisClient,
		MQConn:      mqConn,
		MultiBroker: multiBroker,
	}, cleanup, nil
}
