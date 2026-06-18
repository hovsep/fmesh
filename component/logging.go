package component

import (
	"fmt"
	"log"
	"os"
)

func newDefaultLogger(componentName string) *log.Logger {
	return log.New(os.Stdout, fmt.Sprintf("%s: ", componentName), log.LstdFlags|log.Lmsgprefix)
}

// Logger returns the component's logger.
func (c *Component) Logger() *log.Logger {
	return c.logger
}

// WithLogger is a component constructor option that creates a new logger prefixed with component name.
func WithLogger(logger *log.Logger) Option {
	return func(c *Component) error {
		c.logger = logger
		return nil
	}
}
