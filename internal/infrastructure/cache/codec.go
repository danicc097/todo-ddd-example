package cache

// Codec defines how to marshal/unmarshal an entity for caching.
type Codec[T any] interface {
	Marshal(T) ([]byte, error)
	Unmarshal([]byte) (T, error)
}
