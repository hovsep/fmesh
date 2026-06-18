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

	// TimeLimit defines the maximum duration F-Mesh can run before being forcefully stopped.
	// 0 means no limit (use WithUnlimitedTime to express this explicitly).
	TimeLimit time.Duration
}

// newDefaultConfig returns a safe default configuration.
func newDefaultConfig() Config {
	return Config{
		ErrorHandlingStrategy: StopOnFirstErrorOrPanic,
		CyclesLimit:           1000,
		Debug:                 false,
		TimeLimit:             5 * time.Second,
	}
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
		if limit <= 0 {
			return errors.New("cycles limit must be greater than 0, use WithUnlimitedCycles() to remove the limit")
		}
		fm.config.CyclesLimit = limit
		return nil
	}
}

// WithUnlimitedCycles is an FMesh option that removes the cycle limit.
func WithUnlimitedCycles() Option {
	return func(fm *FMesh) error {
		fm.config.CyclesLimit = 0
		return nil
	}
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
	return func(fm *FMesh) error {
		fm.config.TimeLimit = 0
		return nil
	}
}

// WithDebug is an FMesh option that enables or disables debug mode.
func WithDebug(enabled bool) Option {
	return func(fm *FMesh) error {
		fm.config.Debug = enabled
		return nil
	}
}
