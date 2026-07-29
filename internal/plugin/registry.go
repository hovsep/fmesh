// Package plugin provides the storage and initialization order shared by the
// mesh- and component-level plugin systems.
//
// Not part of the public API: the two public Plugin interfaces
// (fmesh.Plugin, component.Plugin) and their WithPlugins options stay where
// users expect them. Only the parts that were identical at both levels — the
// by-name container, the duplicate check, and the init order — live here, so a
// third level cannot drift from the first two.
package plugin

import (
	"fmt"
	"maps"
	"slices"
)

// Plugin is the shape every plugin has: a name, and an Init that receives the
// thing being constructed.
type Plugin[T any] interface {
	GetName() string
	Init(T) error
}

// Registry is a by-name collection of plugins.
type Registry[T any] struct {
	plugins map[string]Plugin[T]
}

// NewRegistry creates an empty registry.
func NewRegistry[T any]() *Registry[T] {
	return &Registry[T]{plugins: make(map[string]Plugin[T])}
}

// Add registers plugins, failing on the first duplicate name.
func (r *Registry[T]) Add(plugins ...Plugin[T]) error {
	for _, p := range plugins {
		name := p.GetName()
		if _, exists := r.plugins[name]; exists {
			return fmt.Errorf("plugin %s already registered", name)
		}
		r.plugins[name] = p
	}
	return nil
}

// Has returns true if a plugin with that name is registered.
func (r *Registry[T]) Has(name string) bool {
	return r.plugins[name] != nil
}

// InitAll runs every plugin's Init against the target, in name order.
//
// The order is sorted rather than whatever the map yields because plugins
// register hooks, and hooks fire in registration order. Ranging a map would make
// the order in which two plugins observe the same event differ between runs of
// the same program.
func (r *Registry[T]) InitAll(target T) error {
	for _, name := range slices.Sorted(maps.Keys(r.plugins)) {
		if err := r.plugins[name].Init(target); err != nil {
			return fmt.Errorf("plugin %s initialization failed: %w", name, err)
		}
	}
	return nil
}
