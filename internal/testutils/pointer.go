// Package testutils collection of utils to simplify test writing.
package testutils

// Pointer simple func to convert inline values to pointer
// simplify such case v:= 1; s := &s; -> s := Pointer(1).
func Pointer[T any](v T) *T {
	return &v
}
