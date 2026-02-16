package redis

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	userDomain "github.com/danicc097/todo-ddd-example/internal/modules/user/domain"
)

var ErrReplayDetected = errors.New("TOTP code has already been used")

type TOTPGuard struct {
	client *redis.Client
}

func NewTOTPGuard(client *redis.Client) *TOTPGuard {
	return &TOTPGuard{client: client}
}

func (g *TOTPGuard) Consume(ctx context.Context, userID userDomain.UserID, code string) error {
	key := fmt.Sprintf("user:%s:used_totp:%s", userID.String(), code)

	acquired, err := g.client.SetNX(ctx, key, "consumed", 90*time.Second).Result()
	if err != nil {
		return err
	}

	if !acquired {
		return ErrReplayDetected
	}

	return nil
}
