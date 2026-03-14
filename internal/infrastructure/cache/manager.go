package cache

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/sync/singleflight"

	"github.com/danicc097/todo-ddd-example/internal/shared/causation"
)

var (
	sfGroup singleflight.Group
	tracer  = otel.Tracer("cache-manager")
)

func GetOrFetch[T any](
	ctx context.Context,
	store Store,
	key string,
	ttl time.Duration,
	codec Codec[T],
	fetch func(context.Context) (T, error),
	tags ...string,
) (T, error) {
	ctx, span := tracer.Start(ctx, "cache.GetOrFetch", trace.WithAttributes(
		attribute.String("cache.key", key),
	))
	defer span.End()

	val, err := store.Get(ctx, key)
	if err == nil {
		if decoded, decodeErr := codec.Unmarshal(val); decodeErr == nil {
			span.SetAttributes(attribute.Bool("cache.hit", true))
			return decoded, nil
		}
	} else if !errors.Is(err, ErrCacheMiss) {
		slog.WarnContext(ctx, "cache get failed", slog.String("key", key), slog.String("error", err.Error()))
	}

	span.SetAttributes(attribute.Bool("cache.hit", false))

	v, err, _ := sfGroup.Do(key, func() (any, error) {
		fetchCtx, fetchSpan := tracer.Start(ctx, "cache.fetch_fallback")
		defer fetchSpan.End()

		return fetch(fetchCtx)
	})
	if err != nil {
		var zero T
		return zero, err
	}

	result, _ := v.(T)

	go func() {
		if b, marshalErr := codec.Marshal(result); marshalErr == nil {
			bgCtx := trace.ContextWithSpanContext(context.Background(), trace.SpanContextFromContext(ctx))
			bgCtx = causation.WithMetadata(bgCtx, causation.FromContext(ctx))

			bgCtx, asyncSpan := tracer.Start(bgCtx, "cache.async_update", trace.WithSpanKind(trace.SpanKindInternal))
			defer asyncSpan.End()

			_ = store.Set(bgCtx, key, b, ttl, tags...)
		}
	}()

	return result, nil
}
