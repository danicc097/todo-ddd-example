package redis

import (
	"github.com/hashicorp/go-msgpack/codec"

	"github.com/danicc097/todo-ddd-example/internal/modules/todo/domain"
)

type TodoCacheCodec struct {
	handle *codec.MsgpackHandle
}

func NewTodoCacheCodec() *TodoCacheCodec {
	return &TodoCacheCodec{
		handle: &codec.MsgpackHandle{},
	}
}

func (c *TodoCacheCodec) Marshal(t *domain.Todo) ([]byte, error) {
	dto := ToTodoCacheDTO(t)

	var b []byte

	enc := codec.NewEncoderBytes(&b, c.handle)
	err := enc.Encode(dto)

	return b, err
}

func (c *TodoCacheCodec) Unmarshal(data []byte) (*domain.Todo, error) {
	var dto TodoCacheDTO

	dec := codec.NewDecoderBytes(data, c.handle)
	if err := dec.Decode(&dto); err != nil {
		return nil, err
	}

	return FromTodoCacheDTO(dto), nil
}
