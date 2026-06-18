package fmesh

import (
	"errors"
	"fmt"
	"log"
	"os"
)

// Logger returns the mesh logger.
func (fm *FMesh) Logger() *log.Logger {
	return fm.logger
}

// newDefaultLogger creates a new logger prefixed with mesh name.
func newDefaultLogger(meshName string) *log.Logger {
	return log.New(os.Stdout, fmt.Sprintf("%s: ", meshName), log.LstdFlags|log.Lmsgprefix)
}

// IsDebug returns true when debug mode is enabled.
func (fm *FMesh) IsDebug() bool {
	return fm.config.Debug
}

// LogDebug logs a formatted debug message only when debug mode is enabled (no-op otherwise).
// The format string and args follow fmt.Sprintf conventions.
func (fm *FMesh) LogDebug(format string, args ...any) {
	if !fm.IsDebug() {
		return
	}

	fm.Logger().Printf("DEBUG: "+format, args...)
}

// WithLogger is an FMesh option that sets a custom logger.
func WithLogger(logger *log.Logger) Option {
	return func(fm *FMesh) error {
		if logger == nil {
			return errors.New("logger cannot be nil")
		}
		fm.logger = logger
		return nil
	}
}
