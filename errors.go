package fmesh

import (
	"errors"
)

type ErrorHandlingStrategy int

const (
	// StopOnFirstErrorOrPanic stops the f-mesh on first error or panic
	StopOnFirstErrorOrPanic ErrorHandlingStrategy = iota

	// StopOnFirstPanic ignores errors, but stops the f-mesh on first panic
	StopOnFirstPanic

	// IgnoreAll allows to continue running the f-mesh regardless of how components finish their activation functions
	IgnoreAll
)

var (
	ErrHitAnErrorOrPanic                = errors.New("f-mesh hit an error or panic and will be stopped")
	ErrHitAPanic                        = errors.New("f-mesh hit a panic and will be stopped")
	ErrUnsupportedErrorHandlingStrategy = errors.New("unsupported error handling strategy")
	ErrReachedMaxAllowedCycles          = errors.New("reached max allowed cycles")
)
