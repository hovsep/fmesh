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

// LogDebug logs a debug message only when debug mode is enabled (no-op otherwise).
func (fm *FMesh) LogDebug(v ...any) {
	if !fm.IsDebug() {
		return
	}

	fm.Logger().Println(append([]any{"DEBUG:"}, v...)...)
}
