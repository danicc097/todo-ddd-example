package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/sync/singleflight"
)

var (
	sfGroup singleflight.Group
	tracer  = otel.Tracer("cache-manager")
)

func GetOrFetch[T any](
	ctx context.Context,
	rdb *redis.Client,
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

	val, err := rdb.Get(ctx, key).Bytes()
	if err == nil {
		if decoded, decodeErr := codec.Unmarshal(val); decodeErr == nil {
			span.SetAttributes(attribute.Bool("cache.hit", true))
			return decoded, nil
		}
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

			bgCtx, asyncSpan := tracer.Start(bgCtx, "cache.async_update", trace.WithSpanKind(trace.SpanKindInternal))
			defer asyncSpan.End()

			rdb.Set(bgCtx, key, b, ttl)

			for _, tag := range tags {
				tagKey := Keys.TagSet(tag)
				rdb.SAdd(bgCtx, tagKey, key)

				if ttl > 0 {
					rdb.Expire(bgCtx, tagKey, ttl)
				}
			}
		}
	}()

	return result, nil
}

func InvalidateTag(ctx context.Context, rdb *redis.Client, tag string) error {
	tagKey := Keys.TagSet(tag)

	keys, err := rdb.SMembers(ctx, tagKey).Result()
	if err != nil {
		return err
	}

	if len(keys) > 0 {
		if err := rdb.Del(ctx, keys...).Err(); err != nil {
			return err
		}
	}

	return rdb.Del(ctx, tagKey).Err()
}
