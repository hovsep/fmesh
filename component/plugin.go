package component

import "fmt"

// Plugin defines the component plugin interface.
type Plugin interface {
	GetName() string
	Init(*Component) error
}

// Plugins defines a container of component plugins.
type Plugins map[string]Plugin

// newPlugins is a constructor for Plugins.
func newPlugins() Plugins {
	return make(Plugins)
}

// WithPlugins is a component constructor option that adds plugins.
func WithPlugins(plugins ...Plugin) Option {
	return func(c *Component) error {
		for _, plugin := range plugins {
			if _, exists := c.plugins[plugin.GetName()]; exists {
				return fmt.Errorf("plugin %s already registered", plugin.GetName())
			}
			c.plugins[plugin.GetName()] = plugin
		}
		return nil
	}
}

// PluginRegistered returns true if the plugin is registered.
func (c *Component) PluginRegistered(name string) bool {
	return c.plugins[name] != nil
}
