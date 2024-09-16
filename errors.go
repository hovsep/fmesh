package fmesh

import (
	"errors"
)

type ErrorHandlingStrategy int

const (
	StopOnFirstError ErrorHandlingStrategy = iota
	StopOnFirstPanic
	IgnoreAll
)

var (
	ErrHitAnError                       = errors.New("f-mesh hit an error and will be stopped")
	ErrHitAPanic                        = errors.New("f-mesh hit a panic and will be stopped")
	ErrUnsupportedErrorHandlingStrategy = errors.New("unsupported error handling strategy")
)
