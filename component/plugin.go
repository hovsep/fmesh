package component

import (
	"fmt"

	"github.com/hovsep/fmesh/internal/plugin"
)

// Plugin defines the component plugin interface.
type Plugin interface {
	GetName() string
	Init(*Component) error
}

// WithPlugins is a component constructor option that adds plugins.
func WithPlugins(plugins ...Plugin) Option {
	return func(c *Component) error {
		for _, p := range plugins {
			if err := c.plugins.Add(p); err != nil {
				return err
			}
		}
		return nil
	}
}

// PluginRegistered returns true if the plugin is registered.
func (c *Component) PluginRegistered(name string) bool {
	return c.plugins.Has(name)
}

// initPlugins runs every registered plugin's Init, in name order.
func (c *Component) initPlugins() error {
	if err := c.plugins.InitAll(c); err != nil {
		return fmt.Errorf("component %q %w", c.name, err)
	}
	return nil
}

// newPlugins is a constructor for the component plugin registry.
func newPlugins() *plugin.Registry[*Component] {
	return plugin.NewRegistry[*Component]()
}
