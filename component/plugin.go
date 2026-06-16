package component

import "fmt"

// Plugin defines the component plugin interface.
type Plugin interface {
	GetName() string
	Init(*Component) error
}

// WithPlugins is a component constructor option that initializes plugins.
func WithPlugins(plugins ...Plugin) Option {
	return func(c *Component) error {
		for _, plugin := range plugins {
			if err := plugin.Init(c); err != nil {
				return fmt.Errorf("failed to initialize plugin %s: %w", plugin.GetName(), err)
			}
		}
		return nil
	}
}
