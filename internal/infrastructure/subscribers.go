package infrastructure

import (
	"context"

	"github.com/wagslane/go-rabbitmq"
	"go.opentelemetry.io/otel"

	"github.com/danicc097/todo-ddd-example/internal/infrastructure/messaging"
	infraRabbit "github.com/danicc097/todo-ddd-example/internal/infrastructure/rabbitmq"
	scheduleApp "github.com/danicc097/todo-ddd-example/internal/modules/schedule/application"
	scheduleDomain "github.com/danicc097/todo-ddd-example/internal/modules/schedule/domain"
	sharedMessaging "github.com/danicc097/todo-ddd-example/internal/shared/infrastructure/messaging"
)

type Closer interface {
	Close()
}

func RegisterSubscribers(conn *rabbitmq.Conn, scheduleRepo scheduleDomain.ScheduleRepository) ([]Closer, error) {
	subscriber := infraRabbit.NewSubscriber(conn)
	scheduleTracer := otel.Tracer("schedule-consumer")
	todoDeletedHandler := scheduleApp.NewTodoDeletedEventHandler(scheduleRepo)

	mw := sharedMessaging.TraceAndCausationMiddleware(scheduleTracer, func(ctx context.Context, d rabbitmq.Delivery) error {
		return todoDeletedHandler.Handle(ctx, d.Body)
	})

	todoDeletedConsumer, err := subscriber.Subscribe(
		messaging.Keys.ScheduleTodoDeletedQueue(),
		messaging.Keys.TodoEventsExchange(),
		[]string{"todo.deleted.*"},
		mw,
	)
	if err != nil {
		return nil, err
	}

	return []Closer{todoDeletedConsumer}, nil
}
