package fmesh

import (
	"errors"
	"fmt"
	"time"
)

// config is the mesh configuration. It is set only through the With* options,
// which start from the defaults: 0 means "no limit" on every limit, so a zero
// must never stand for "not set".
type config struct {
	// ErrorHandlingStrategy defines how f-mesh will handle errors and panics.
	ErrorHandlingStrategy ErrorHandlingStrategy

	// Debug enables debug mode, which logs additional detailed information for troubleshooting and analysis.
	Debug bool

	// CyclesLimit defines the maximum number of activation cycles.
	// 0 means no limit (use WithUnlimitedCycles to express this explicitly).
	CyclesLimit int

	// TimeLimit is the maximum duration of a Run. It becomes a deadline on the
	// run context; a running cycle is never interrupted.
	// 0 means no limit (use WithUnlimitedTime to express this explicitly).
	TimeLimit time.Duration

	// CyclesHistoryLimit defines the maximum number of past cycles retained in
	// RuntimeInfo.Cycles; 0 means unlimited (the default).
	CyclesHistoryLimit int

	// LivelockThreshold is how many consecutive stalled cycles end the run with
	// ErrLivelockDetected (see livelockDetector.detect). A waiter that proceeds
	// on State, the clock or the outside world needs a higher threshold or none.
	// 0 disables detection (use WithoutLivelockDetection to say so explicitly).
	LivelockThreshold int
}

// newDefaultConfig returns a safe default configuration.
func newDefaultConfig() config {
	return config{
		ErrorHandlingStrategy: StopOnFirstErrorOrPanic,
		CyclesLimit:           1000,
		TimeLimit:             5 * time.Second,
		LivelockThreshold:     2,
	}
}

// setPositive assigns a limit that must be greater than 0. Zero means "no limit",
// so it is rejected: msg names the option that removes the limit explicitly.
func setPositive(value int, field *int, msg string) error {
	if value <= 0 {
		return errors.New(msg)
	}
	*field = value
	return nil
}

// WithLivelockThreshold is an FMesh option that sets how many consecutive stalled
// cycles end the run with ErrLivelockDetected. threshold must be greater than 0.
// Use WithoutLivelockDetection to turn detection off.
func WithLivelockThreshold(threshold int) Option {
	return func(fm *FMesh) error {
		return setPositive(threshold, &fm.config.LivelockThreshold,
			"livelock threshold must be greater than 0, use WithoutLivelockDetection() to disable detection")
	}
}

// WithoutLivelockDetection is an FMesh option that disables livelock detection.
// A livelocked mesh then runs until it hits the cycle or time limit.
func WithoutLivelockDetection() Option {
	return func(fm *FMesh) error { fm.config.LivelockThreshold = 0; return nil }
}

// WithErrorHandlingStrategy is an FMesh option that sets the error handling strategy.
// An unknown strategy fails New with ErrUnsupportedErrorHandlingStrategy.
func WithErrorHandlingStrategy(s ErrorHandlingStrategy) Option {
	return func(fm *FMesh) error {
		switch s {
		case StopOnFirstErrorOrPanic, StopOnFirstPanic, IgnoreAll:
			fm.config.ErrorHandlingStrategy = s
			return nil
		default:
			return fmt.Errorf("%w: %d", ErrUnsupportedErrorHandlingStrategy, s)
		}
	}
}

// WithCyclesLimit is an FMesh option that sets the maximum number of activation cycles.
// limit must be greater than 0. Use WithUnlimitedCycles to remove the cycle limit.
func WithCyclesLimit(limit int) Option {
	return func(fm *FMesh) error {
		return setPositive(limit, &fm.config.CyclesLimit,
			"cycles limit must be greater than 0, use WithUnlimitedCycles() to remove the limit")
	}
}

// WithUnlimitedCycles is an FMesh option that removes the cycle limit.
func WithUnlimitedCycles() Option {
	return func(fm *FMesh) error { fm.config.CyclesLimit = 0; return nil }
}

// WithTimeLimit is an FMesh option that sets the maximum duration the mesh can run.
// d must be greater than 0. Use WithUnlimitedTime to remove the time limit.
func WithTimeLimit(d time.Duration) Option {
	return func(fm *FMesh) error {
		if d <= 0 {
			return errors.New("time limit must be greater than 0, use WithUnlimitedTime() to remove the limit")
		}
		fm.config.TimeLimit = d
		return nil
	}
}

// WithUnlimitedTime is an FMesh option that removes the time limit.
func WithUnlimitedTime() Option {
	return func(fm *FMesh) error { fm.config.TimeLimit = 0; return nil }
}

// WithDebug is an FMesh option that enables or disables debug mode.
func WithDebug(enabled bool) Option {
	return func(fm *FMesh) error {
		fm.config.Debug = enabled
		return nil
	}
}

// WithCyclesHistoryLimit is an FMesh option that sets the maximum number of past cycles
// retained in RuntimeInfo.Cycles. limit must be greater than 0. Use WithUnlimitedCyclesHistory
// to remove the limit.
func WithCyclesHistoryLimit(limit int) Option {
	return func(fm *FMesh) error {
		return setPositive(limit, &fm.config.CyclesHistoryLimit,
			"cycles history limit must be greater than 0, use WithUnlimitedCyclesHistory() to remove the limit")
	}
}

// WithUnlimitedCyclesHistory is an FMesh option that removes the cycles history retention limit.
func WithUnlimitedCyclesHistory() Option {
	return func(fm *FMesh) error { fm.config.CyclesHistoryLimit = 0; return nil }
}
