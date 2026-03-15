package redis

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/danicc097/todo-ddd-example/internal/infrastructure/cache"
	userDomain "github.com/danicc097/todo-ddd-example/internal/modules/user/domain"
)

var ErrReplayDetected = errors.New("TOTP code has already been used")

type TOTPGuard struct {
	client redis.UniversalClient
}

func NewTOTPGuard(client redis.UniversalClient) *TOTPGuard {
	return &TOTPGuard{client: client}
}

func (g *TOTPGuard) Consume(ctx context.Context, userID userDomain.UserID, code string) error {
	key := cache.Keys.TOTPUsedCode(userID, code)

	acquired, err := g.client.SetNX(ctx, key, "consumed", 90*time.Second).Result()
	if err != nil {
		return fmt.Errorf("setnx: %w", err)
	}

	if !acquired {
		return ErrReplayDetected
	}

	return nil
}
