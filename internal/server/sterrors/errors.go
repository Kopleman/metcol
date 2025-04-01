// Package sterrors collection on custom store errors
package sterrors

import "errors"

var ErrNotFound = errors.New("not found")
var ErrAlreadyExists = errors.New("already exists")
