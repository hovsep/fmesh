package fmesh

import (
	"errors"
)

// ErrorHandlingStrategy defines the strategy for handling errors in run-time.
type ErrorHandlingStrategy int

const (
	// StopOnFirstErrorOrPanic stops the f-mesh on the first error or panic.
	StopOnFirstErrorOrPanic ErrorHandlingStrategy = iota

	// StopOnFirstPanic ignores errors but stops the f-mesh on first panic.
	StopOnFirstPanic

	// IgnoreAll allows continuing running the f-mesh regardless of how components finish their activation functions.
	IgnoreAll
)

var (
	// ErrHitAnErrorOrPanic is returned when f-mesh hit an error or panic and will be stopped.
	ErrHitAnErrorOrPanic = errors.New("f-mesh hit an error or panic and will be stopped")
	// ErrHitAPanic is returned when f-mesh hit a panic and will be stopped.
	ErrHitAPanic = errors.New("f-mesh hit a panic and will be stopped")
	// ErrUnsupportedErrorHandlingStrategy is returned when an unsupported error handling strategy is used.
	ErrUnsupportedErrorHandlingStrategy = errors.New("unsupported error handling strategy")
	// ErrReachedMaxAllowedCycles is returned when the maximum number of allowed cycles is reached.
	ErrReachedMaxAllowedCycles = errors.New("reached max allowed cycles")
	// ErrTimeLimitExceeded is returned when the time limit is exceeded.
	ErrTimeLimitExceeded = errors.New("time limit exceeded")
	// ErrLivelockDetected is returned when the mesh stopped making progress: every
	// component that activated was waiting for inputs and no signal moved, for
	// LivelockThreshold consecutive cycles. The wrapped message names the stuck
	// components and, for each, which of its input ports are empty.
	ErrLivelockDetected = errors.New("livelock detected")
	// ErrRunCanceled is returned when the context passed to Run is canceled or
	// hits a deadline of its own. It wraps the underlying context error, so
	// errors.Is(err, context.Canceled) and context.DeadlineExceeded both work.
	// A time limit configured on the mesh reports ErrTimeLimitExceeded instead.
	ErrRunCanceled         = errors.New("run canceled")
	errFailedToRunCycle    = errors.New("failed to run cycle")
	errNoComponents        = errors.New("no components found")
	errFailedToClearInputs = errors.New("failed to clear input ports")
	// ErrFailedToDrain is returned when failed to drain.
	ErrFailedToDrain = errors.New("failed to drain")
)
