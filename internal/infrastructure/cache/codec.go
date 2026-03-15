package cache

import (
	"fmt"
	"reflect"

	"github.com/ugorji/go/codec"
)

// Codec defines how to marshal/unmarshal an entity for caching.
type Codec[T any] interface {
	Marshal(T) ([]byte, error)
	Unmarshal([]byte) (T, error)
}

func NewMsgpackHandle() *codec.MsgpackHandle {
	h := &codec.MsgpackHandle{}
	h.MapType = reflect.TypeFor[map[string]any]()
	h.RawToString = true

	h.TypeInfos = codec.NewTypeInfos([]string{"codec", "json"})

	return h
}

type MsgpackCodec[T any] struct {
	handle *codec.MsgpackHandle
}

func NewMsgpackCodec[T any]() *MsgpackCodec[T] {
	return &MsgpackCodec[T]{handle: NewMsgpackHandle()}
}

func (c *MsgpackCodec[T]) Marshal(v T) ([]byte, error) {
	var b []byte

	err := codec.NewEncoderBytes(&b, c.handle).Encode(v)
	if err != nil {
		return nil, fmt.Errorf("codec.Encode failed: %w", err)
	}

	return b, nil
}

func (c *MsgpackCodec[T]) Unmarshal(data []byte) (T, error) {
	var dest T

	err := codec.NewDecoderBytes(data, c.handle).Decode(&dest)
	if err != nil {
		return dest, fmt.Errorf("codec.Decode failed: %w", err)
	}

	return dest, nil
}

// CollectionCodec implements Codec for slices of any type.
type CollectionCodec[T any] struct {
	handle *codec.MsgpackHandle
}

func NewCollectionCodec[T any]() *CollectionCodec[T] {
	return &CollectionCodec[T]{handle: NewMsgpackHandle()}
}

func (c CollectionCodec[T]) Marshal(v []T) ([]byte, error) {
	var b []byte

	err := codec.NewEncoderBytes(&b, c.handle).Encode(v)
	if err != nil {
		return nil, fmt.Errorf("codec.Encode failed: %w", err)
	}

	return b, nil
}

func (c CollectionCodec[T]) Unmarshal(b []byte) ([]T, error) {
	var res []T

	err := codec.NewDecoderBytes(b, c.handle).Decode(&res)
	if err != nil {
		return nil, fmt.Errorf("codec.Decode failed: %w", err)
	}

	return res, nil
}
