package fmesh

import (
	"log"
	"os"
)

// Logger returns the F-Mesh logger.
func (fm *FMesh) Logger() *log.Logger {
	return fm.config.Logger
}

func getDefaultLogger() *log.Logger {
	logger := log.Default()
	logger.SetOutput(os.Stdout)
	logger.SetFlags(log.LstdFlags | log.Lmsgprefix)
	return logger
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
