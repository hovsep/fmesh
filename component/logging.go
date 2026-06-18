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

// WithLogger is a component constructor option that creates a new logger prefixed with component name.
func WithLogger(logger *log.Logger) Option {
	return func(c *Component) error {
		if logger == nil {
			return errors.New("logger cannot be nil")
		}
		c.logger = logger
		return nil
	}
}
