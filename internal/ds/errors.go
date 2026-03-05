package ds

import "errors"

var (
	ErrInvalidRuntime  = errors.New("invalid runtime type")
	ErrFunctionTimeout = errors.New("function timeout")
)
