package application

import (
	"context"
	"fmt"
	"log/slog"
	"reflect"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type loggingDecorator[C any, R any] struct {
	base   RequestHandler[C, R]
	tracer trace.Tracer
}

func getHandlerName[C any, R any](handler RequestHandler[C, R]) string {
	// get a readable name for the span
	t := reflect.TypeOf(handler)
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}

	return t.Name()
}

// WithLogging wraps a RequestHandler with logging and tracing.
func WithLogging[C any, R any](base RequestHandler[C, R], tracerName string) RequestHandler[C, R] {
	return &loggingDecorator[C, R]{
		base:   base,
		tracer: otel.Tracer(tracerName),
	}
}

func (d *loggingDecorator[C, R]) Handle(ctx context.Context, cmd C) (R, error) {
	var zero R

	handlerName := getHandlerName(d.base)

	ctx, span := d.tracer.Start(ctx, handlerName, trace.WithAttributes(
		attribute.String("app.handler", handlerName),
	))
	defer span.End()

	slog.DebugContext(ctx, "starting use case "+handlerName)

	start := time.Now()
	res, err := d.base.Handle(ctx, cmd)
	latency := time.Since(start)

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		slog.ErrorContext(ctx, fmt.Sprintf("use case %s failed", handlerName), slog.String("error", err.Error()), slog.Duration("latency", latency))

		return zero, err
	}

	slog.DebugContext(ctx, fmt.Sprintf("use case %s finished", handlerName), slog.Duration("latency", latency))

	return res, nil
}
