package redis

import (
	"fmt"

	"github.com/ugorji/go/codec"

	"github.com/danicc097/todo-ddd-example/internal/infrastructure/cache"
	"github.com/danicc097/todo-ddd-example/internal/modules/todo/domain"
)

type TodoCacheCodec struct {
	handle *codec.MsgpackHandle
}

func NewTodoCacheCodec() *TodoCacheCodec {
	return &TodoCacheCodec{
		handle: cache.NewMsgpackHandle(),
	}
}

func (c *TodoCacheCodec) Marshal(t *domain.Todo) ([]byte, error) {
	dto := ToTodoCacheDTO(t)

	var b []byte

	enc := codec.NewEncoderBytes(&b, c.handle)

	err := enc.Encode(dto)
	if err != nil {
		return nil, fmt.Errorf("encode: %w", err)
	}

	return b, nil
}

func (c *TodoCacheCodec) Unmarshal(data []byte) (*domain.Todo, error) {
	var dto TodoCacheDTO

	dec := codec.NewDecoderBytes(data, c.handle)
	if err := dec.Decode(&dto); err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}

	return FromTodoCacheDTO(dto), nil
}
