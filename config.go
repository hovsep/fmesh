package fmesh

import (
	"errors"
	"time"
)

// Config defines the configuration for the f-mesh.
type Config struct {
	// ErrorHandlingStrategy defines how f-mesh will handle errors and panics.
	ErrorHandlingStrategy ErrorHandlingStrategy

	// Debug enables debug mode, which logs additional detailed information for troubleshooting and analysis.
	Debug bool

	// CyclesLimit defines the maximum number of activation cycles.
	// 0 means no limit (use WithUnlimitedCycles to express this explicitly).
	CyclesLimit int

	// TimeLimit defines the maximum duration F-Mesh can run.
	// The limit is checked between activation cycles: a cycle that is already
	// running is never interrupted, so a long or blocking activation function
	// can exceed the limit.
	// 0 means no limit (use WithUnlimitedTime to express this explicitly).
	TimeLimit time.Duration

	// CyclesHistoryLimit defines the maximum number of past cycles retained in
	// RuntimeInfo.Cycles; 0 means unlimited (the default).
	CyclesHistoryLimit int

	// LivelockThreshold is how many consecutive stalled cycles end the run with
	// ErrLivelockDetected. A cycle is stalled when every component that activated
	// was waiting for inputs and the mesh's pending signals did not change, which
	// means the next cycle would be identical to this one.
	//
	// The default of 2 rather than 1 leaves room for a BeforeCycle hook that feeds
	// the mesh: such a hook changes the pending signal count, which resets the
	// counter anyway, but a threshold above one keeps the detector conservative.
	// 0 disables detection (use WithoutLivelockDetection to say so explicitly).
	LivelockThreshold int
}

// newDefaultConfig returns a safe default configuration.
func newDefaultConfig() Config {
	return Config{
		ErrorHandlingStrategy: StopOnFirstErrorOrPanic,
		CyclesLimit:           1000,
		Debug:                 false,
		TimeLimit:             5 * time.Second,
		CyclesHistoryLimit:    0,
		LivelockThreshold:     2,
	}
}

// setPositive assigns a config field that must be greater than 0.
//
// Zero is reserved as "no limit" on every one of these fields, which is why it is
// rejected rather than accepted: a caller passing 0 almost always means a limit,
// not the removal of one, and the message names the option that does remove it.
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

// WithConfig is an FMesh option that replaces the entire configuration.
func WithConfig(config Config) Option {
	return func(fm *FMesh) error {
		fm.config = config
		return nil
	}
}

// WithErrorHandlingStrategy is an FMesh option that sets the error handling strategy.
func WithErrorHandlingStrategy(s ErrorHandlingStrategy) Option {
	return func(fm *FMesh) error {
		fm.config.ErrorHandlingStrategy = s
		return nil
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
