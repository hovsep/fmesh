package fmesh

import (
	"fmt"

	"github.com/hovsep/fmesh/internal/plugin"
)

// Plugin defines the mesh plugin interface.
//
// The component-level counterpart (component.Plugin) bundles initialization for
// one component. This one bundles it for a whole mesh, which makes it the
// natural home for anything cross-cutting: measuring, tracing, exporting,
// asserting. None of those belong to any single component, and all of them would
// otherwise have to be wired into every one by hand.
//
// A plugin that wants to reach the components cannot simply walk them in Init:
// a mesh is constructed empty and filled by AddComponents afterwards, so at Init
// time there is nothing to walk. It registers an OnComponentAdded hook instead
// and instruments each component as it arrives. That indirection is the whole
// trick, and it is why a mesh plugin can observe every activation in a mesh
// without a single component knowing it exists.
type Plugin interface {
	GetName() string
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
