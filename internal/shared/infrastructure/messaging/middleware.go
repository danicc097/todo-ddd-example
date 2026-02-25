package messaging

import (
	"context"

	"github.com/google/uuid"
	"github.com/wagslane/go-rabbitmq"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"

	"github.com/danicc097/todo-ddd-example/internal/shared/causation"
)

// TraceAndCausationMiddleware extracts tracing and causation info from RabbitMQ headers.
func TraceAndCausationMiddleware(tracer trace.Tracer, next func(ctx context.Context, d rabbitmq.Delivery) error) func(ctx context.Context, d rabbitmq.Delivery) error {
	return func(ctx context.Context, d rabbitmq.Delivery) error {
		headers := make(map[string]string)

		for k, v := range d.Headers {
			if s, ok := v.(string); ok {
				headers[k] = s
			}
		}

		carrier := propagation.MapCarrier(headers)
		ctx = otel.GetTextMapPropagator().Extract(ctx, carrier)

		meta := causation.Metadata{
			CorrelationID: headers[causation.AttrCorrelationID],
			CausationID:   d.MessageId,
			UserIP:        headers[causation.AttrUserIP],
			UserAgent:     headers[causation.AttrUserAgent],
		}
		if uidStr := headers[causation.AttrUserID]; uidStr != "" {
			if uid, err := uuid.Parse(uidStr); err == nil {
				meta.UserID = uid
			}
		}

		ctx = causation.WithMetadata(ctx, meta)

		ctx, span := tracer.Start(ctx, "messaging.consume",
			trace.WithSpanKind(trace.SpanKindConsumer),
			trace.WithAttributes(
				attribute.String("messaging.system", "rabbitmq"),
				attribute.String("messaging.operation", "process"),
				attribute.String("messaging.message_id", d.MessageId),
				attribute.String("messaging.routing_key", d.RoutingKey),
			),
		)
		defer span.End()

		return next(ctx, d)
	}
}
