package mapper

// MapList applies a transformation function to each element of a slice.
func MapList[T any, R any](items []T, transform func(T) R) []R {
	result := make([]R, len(items))
	for i, item := range items {
		result[i] = transform(item)
	}

	return result
}
