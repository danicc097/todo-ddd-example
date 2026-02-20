package redis

import (
	"github.com/ugorji/go/codec"

	"github.com/danicc097/todo-ddd-example/internal/infrastructure/cache"
	"github.com/danicc097/todo-ddd-example/internal/modules/todo/domain"
)

type TagCacheCodec struct {
	handle *codec.MsgpackHandle
}

func NewTagCacheCodec() *TagCacheCodec {
	return &TagCacheCodec{
		handle: cache.NewMsgpackHandle(),
	}
}

func (c *TagCacheCodec) Marshal(t *domain.Tag) ([]byte, error) {
	dto := ToTagCacheDTO(t)

	var b []byte

	enc := codec.NewEncoderBytes(&b, c.handle)
	err := enc.Encode(dto)

	return b, err
}

func (c *TagCacheCodec) Unmarshal(data []byte) (*domain.Tag, error) {
	var dto TagCacheDTO

	dec := codec.NewDecoderBytes(data, c.handle)
	if err := dec.Decode(&dto); err != nil {
		return nil, err
	}

	return FromTagCacheDTO(dto), nil
}
