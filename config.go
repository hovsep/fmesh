package fmesh

import (
	"log"
	"time"
)

// @TODO: use functional options instead of such constants.
const (
	// UnlimitedCycles defines the maximum number of activation cycles, 0 means no limit.
	UnlimitedCycles = 0
	// UnlimitedTime defines the maximum duration F-Mesh can run before being forcefully stopped, 0 means no limit.
	UnlimitedTime = 0
)

// Config defines the configuration for the f-mesh.
type Config struct {
	// ErrorHandlingStrategy defines how f-mesh will handle errors and panics
	ErrorHandlingStrategy ErrorHandlingStrategy

	// Debug enables debug mode, which logs additional detailed information for troubleshooting and analysis.
	Debug bool

	Logger *log.Logger

	// CyclesLimit defines max number of activation cycles, 0 means no limit
	CyclesLimit int

	// TimeLimit defines the maximum duration F-Mesh can run before being forcefully stopped.
	// A value of 0 disables the time constraint, allowing indefinite execution.
	TimeLimit time.Duration
}

// @TODO: maybe we need to use struct, not pointer
// @TODO: Use functional options
// newDefaultConfig returns a new default configuration.
func newDefaultConfig() *Config {
	return &Config{
		ErrorHandlingStrategy: StopOnFirstErrorOrPanic,
		CyclesLimit:           1000,
		Debug:                 false,
		Logger:                getDefaultLogger(),
		TimeLimit:             UnlimitedTime,
	}
}

// WithConfig is an FMesh option that sets the configuration.
func WithConfig(config *Config) Option {
	return func(fm *FMesh) error {
		fm.config = config
		if fm.config.Logger == nil {
			fm.config.Logger = getDefaultLogger()
		}
		return nil
	}
}
