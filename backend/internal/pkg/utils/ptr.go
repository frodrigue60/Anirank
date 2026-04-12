package utils

// Ptr returns a pointer to the value passed as an argument.
// This is useful for creating pointers to literals or constants in a single line.
func Ptr[T any](v T) *T {
	return &v
}
