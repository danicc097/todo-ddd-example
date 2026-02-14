package redis

import (
	"github.com/hashicorp/go-msgpack/codec"

	api "github.com/danicc097/todo-ddd-example/internal/generated/api"
)

type APITodoCacheCodec struct {
	handle *codec.MsgpackHandle
}

func NewAPITodoCacheCodec() *APITodoCacheCodec {
	return &APITodoCacheCodec{
		handle: &codec.MsgpackHandle{},
	}
}

func (c *APITodoCacheCodec) Marshal(t *api.Todo) ([]byte, error) {
	var b []byte

	enc := codec.NewEncoderBytes(&b, c.handle)
	err := enc.Encode(t)

	return b, err
}

func (c *APITodoCacheCodec) Unmarshal(data []byte) (*api.Todo, error) {
	var dto api.Todo

	dec := codec.NewDecoderBytes(data, c.handle)
	if err := dec.Decode(&dto); err != nil {
		return nil, err
	}

	return &dto, nil
}
