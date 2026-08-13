package fmesh

import (
	"fmt"

	"github.com/hovsep/fmesh/internal/plugin"
)

// Plugin defines the mesh plugin interface — the home for cross-cutting
// concerns (component.Plugin is the per-component counterpart).
//
// Init cannot walk components: a mesh is constructed empty and filled by
// AddComponents afterwards. Register an OnComponentAdded hook instead and
// instrument each component as it arrives.
type Plugin interface {
	Name() string
	Init(*FMesh) error
}

// WithPlugins is a mesh constructor option that adds plugins.
func WithPlugins(plugins ...Plugin) Option {
	return func(fm *FMesh) error {
		for _, p := range plugins {
			if err := fm.plugins.Add(p); err != nil {
				return err
			}
		}
		return nil
	}
}

// PluginRegistered returns true if the plugin is registered.
func (fm *FMesh) PluginRegistered(name string) bool {
	return fm.plugins.Has(name)
}

// initPlugins runs every registered plugin's Init, in name order.
func (fm *FMesh) initPlugins() error {
	if err := fm.plugins.InitAll(fm); err != nil {
		return fmt.Errorf("fmesh %q %w", fm.name, err)
	}
	return nil
}

// newPlugins is a constructor for the mesh plugin registry.
func newPlugins() *plugin.Registry[*FMesh] {
	return plugin.NewRegistry[*FMesh]()
}
