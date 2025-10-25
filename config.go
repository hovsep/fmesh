package fmesh

import (
	"log"
	"time"
)

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

var defaultConfig = &Config{
	ErrorHandlingStrategy: StopOnFirstErrorOrPanic,
	CyclesLimit:           1000,
	Debug:                 false,
	Logger:                getDefaultLogger(),
	TimeLimit:             UnlimitedTime,
}

// withConfig sets the configuration and returns the f-mesh.
func (fm *FMesh) withConfig(config *Config) *FMesh {
	if fm.HasChainableErr() {
		return fm
	}

	fm.config = config

	if fm.Logger() == nil {
		fm.config.Logger = getDefaultLogger()
	}
	return fm
}
