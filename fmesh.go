package fmesh

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/hovsep/fmesh/component"
	"github.com/hovsep/fmesh/internal/plugin"
	"github.com/hovsep/fmesh/meta"
)

// Option is a functional option for configuring an FMesh during construction.
type Option func(*FMesh) error

// FMesh is the functional mesh.
type FMesh struct {
	name        string
	description string
	meta        *meta.Meta
	components  *component.Collection
	runtimeInfo *RuntimeInfo
	logger      *log.Logger
	config      config
	hooks       *Hooks
	plugins     *plugin.Registry[*FMesh]
	livelock    livelockDetector // per run; reset by cleanUpPreviousRun

	// Per-cycle scratch, reused so a cycle does not allocate them: the components
	// ready to activate, and one result slot per component in name order.
	ready []readyComponent
	slots []*component.ActivationResult
}

// New creates a new F-Mesh with the default configuration and applies any provided options.
func New(name string, opts ...Option) (*FMesh, error) {
	fm := &FMesh{
		name:       name,
		meta:       meta.New(),
		components: component.NewCollection(),
		logger:     newDefaultLogger(name),
		config:     newDefaultConfig(),
		hooks:      newHooks(),
		plugins:    plugin.NewRegistry[*FMesh](),
	}
	for _, opt := range opts {
		if err := opt(fm); err != nil {
			return nil, fmt.Errorf("fmesh %q option failed: %w", name, err)
		}
	}
	// Built after the options so the runtime info picks up the configured
	// history limit (Run rebuilds it the same way).
	fm.runtimeInfo = newRuntimeInfo(fm.config.CyclesHistoryLimit)

	// Last, so a plugin sees the fully configured mesh and can rely on anything
	// the options set up.
	if err := fm.initPlugins(); err != nil {
		return nil, err
	}
	return fm, nil
}

// Name returns the name of the F-Mesh.
func (fm *FMesh) Name() string {
	return fm.name
}

// Description returns the description of the F-Mesh.
func (fm *FMesh) Description() string {
	return fm.description
}

// Components returns all components in the mesh.
func (fm *FMesh) Components() *component.Collection {
	return fm.components
}

// ComponentByName returns a component by name.
func (fm *FMesh) ComponentByName(name string) *component.Component {
	return fm.Components().ByName(name)
}

// WithDescription is a constructor option that sets a description on the mesh.
func WithDescription(description string) Option {
	return func(fm *FMesh) error {
		fm.description = description
		return nil
	}
}

// Meta returns the mesh's metadata store.
func (fm *FMesh) Meta() *meta.Meta {
	return fm.meta
}

// WithMeta is a constructor option that adds or updates one metadata entry on the mesh.
func WithMeta[T meta.Value](key string, value T) Option {
	return func(fm *FMesh) error {
		fm.meta.Set(key, value)
		return nil
	}
}

// AddComponents adds components to the mesh. Returns an error if any component is invalid or has a
// duplicate name; then no component is added or changed. The OnComponentAdded hooks run after all
// components are added, for every component even when one fails, and a failing hook leaves them
// added.
func (fm *FMesh) AddComponents(components ...*component.Component) error {
	for _, c := range components {
		if err := c.Validate(); err != nil {
			return fmt.Errorf("failed to add component %q: %w", c.Name(), err)
		}
	}

	if err := fm.components.Add(components...); err != nil {
		return fmt.Errorf("failed to add components to mesh: %w", err)
	}

	for _, c := range components {
		c.SetParentMesh(fm)
		c.InheritLogger(fm.logger)
	}

	var hookErrs []error
	for _, c := range components {
		// Components are added outside a run, so there is no run context yet.
		if err := fm.hooks.onComponentAdded.TriggerAll(context.Background(), &ComponentAddedContext{FMesh: fm, Component: c}); err != nil {
			hookErrs = append(hookErrs, fmt.Errorf("onComponentAdded hook failed for component %q: %w", c.Name(), err))
		}
	}

	fm.LogDebug("%d components added to mesh", fm.Components().Len())
	return errors.Join(hookErrs...)
}

// SetupHooks configures hooks for the mesh using a closure.
func (fm *FMesh) SetupHooks(configure func(*Hooks)) *FMesh {
	configure(fm.hooks)
	return fm
}
