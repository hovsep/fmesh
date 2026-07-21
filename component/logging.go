package component

import (
	"errors"
	"log"
	"os"
)

// newDefaultLogger creates a new logger prefixed with component name.
func newDefaultLogger(componentName string) *log.Logger {
	return log.New(os.Stdout, componentName+": ", log.LstdFlags|log.Lmsgprefix)
}

// Logger returns the component's logger.
func (c *Component) Logger() *log.Logger {
	return c.logger
}

// WithLogger is a component constructor option that sets a custom logger.
// A custom logger is never overridden by the mesh logger.
func WithLogger(logger *log.Logger) Option {
	return func(c *Component) error {
		return c.SetLogger(logger)
	}
}

// SetLogger sets a custom logger. A custom logger is never overridden by the mesh logger.
func (c *Component) SetLogger(logger *log.Logger) error {
	if logger == nil {
		return errors.New("logger cannot be nil")
	}
	c.logger = logger
	c.customLogger = true
	return nil
}

// InheritLogger sets the given logger only when the component has no custom logger.
// Called by the mesh when the component is added, so components share the mesh logger by default.
func (c *Component) InheritLogger(logger *log.Logger) {
	if c.customLogger || logger == nil {
		return
	}
	c.logger = logger
}
