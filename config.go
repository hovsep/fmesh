package fmesh

import "log"

const UnlimitedCycles = 0

type Config struct {
	// ErrorHandlingStrategy defines how f-mesh will handle errors and panics
	ErrorHandlingStrategy ErrorHandlingStrategy
	// CyclesLimit defines max number of activation cycles, 0 means no limit
	CyclesLimit int
	// Debug flag enabled debug mode, when additional information will be logged
	Debug  bool
	Logger *log.Logger
}

var defaultConfig = &Config{
	ErrorHandlingStrategy: StopOnFirstErrorOrPanic,
	CyclesLimit:           1000,
	Debug:                 false,
	Logger:                getDefaultLogger(),
}

// withConfig sets the configuration and returns the f-mesh
func (fm *FMesh) withConfig(config *Config) *FMesh {
	if fm.HasErr() {
		return fm
	}

	fm.config = config

	if fm.Logger() == nil {
		fm.config.Logger = getDefaultLogger()
	}
	return fm
}
