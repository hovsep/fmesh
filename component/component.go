package component

import (
	"errors"
	"fmt"
	"log"

	"github.com/hovsep/fmesh/meta"
	"github.com/hovsep/fmesh/port"
)

// Component defines a main building block of FMesh.
type Component struct {
	name        string
	description string
	labels      *meta.Labels
	scalars     *meta.Scalars
	inputPorts  *port.Collection
	outputPorts *port.Collection
	f           ActivationFunc
	logger      *log.Logger
	state       State
	parentMesh  ParentMesh
	hooks       *Hooks
	plugins     Plugins
}

// New creates a new component with the given name and options.
func New(name string, opts ...Option) (*Component, error) {
	c := &Component{
		name:        name,
		description: "",
		labels:      meta.NewLabels(),
		scalars:     meta.NewScalars(),
		inputPorts:  port.NewCollection(),
		outputPorts: port.NewCollection(),
		logger:      newDefaultLogger(name),
		state:       newState(),
		hooks:       newHooks(),
		plugins:     newPlugins(),
	}

	for _, opt := range opts {
		if err := opt(c); err != nil {
			return nil, fmt.Errorf("component %q option failed: %w", name, err)
		}
	}

	for pluginName, plugin := range c.plugins {
		if err := plugin.Init(c); err != nil {
			return nil, fmt.Errorf("component %q plugin %s initialization failed: %w", name, pluginName, err)
		}
	}

	if err := c.hooks.onCreation.Trigger(c); err != nil {
		return nil, fmt.Errorf("component %q on creation hook failed: %w", name, err)
	}

	return c, nil
}

// Name returns the component's name.
func (c *Component) Name() string {
	return c.name
}

// Description returns the component's description.
func (c *Component) Description() string {
	return c.description
}

// Labels returns the component's labels collection.
func (c *Component) Labels() *meta.Labels {
	return c.labels
}

// SetLabels replaces all labels.
func (c *Component) SetLabels(labelMap map[string]string) *Component {
	c.labels.Clear().SetMany(labelMap)
	return c
}

// AddLabels adds or updates labels.
func (c *Component) AddLabels(labelMap map[string]string) *Component {
	c.labels.SetMany(labelMap)
	return c
}

// AddLabel adds or updates a single label.
func (c *Component) AddLabel(name, value string) *Component {
	c.labels.Set(name, value)
	return c
}

// ClearLabels removes all labels.
func (c *Component) ClearLabels() *Component {
	c.labels.Clear()
	return c
}

// RemoveLabels removes specific labels.
func (c *Component) RemoveLabels(names ...string) *Component {
	c.labels.Remove(names...)
	return c
}

// Scalars returns the component's scalars store.
func (c *Component) Scalars() *meta.Scalars {
	return c.scalars
}

// SetScalars replaces all scalars.
func (c *Component) SetScalars(scalarsMap map[string]float64) *Component {
	c.scalars.Clear().SetMany(scalarsMap)
	return c
}

// AddScalars adds or updates scalars.
func (c *Component) AddScalars(scalarsMap map[string]float64) *Component {
	c.scalars.SetMany(scalarsMap)
	return c
}

// AddScalar adds or updates a single scalar.
func (c *Component) AddScalar(name string, value float64) *Component {
	c.scalars.Set(name, value)
	return c
}

// ClearScalars removes all scalars.
func (c *Component) ClearScalars() *Component {
	c.scalars.Clear()
	return c
}

// RemoveScalars removes specific scalars.
func (c *Component) RemoveScalars(names ...string) *Component {
	c.scalars.Remove(names...)
	return c
}

// WithDescription is a component constructor option that sets the description.
func WithDescription(description string) Option {
	return func(c *Component) error {
		c.description = description
		return nil
	}
}

// WithLabel is a component constructor option that adds or updates a single label.
func WithLabel(name, value string) Option {
	return func(c *Component) error {
		c.labels.Set(name, value)
		return nil
	}
}

// WithScalar is a component constructor option that adds or updates a single scalar.
func WithScalar(name string, value float64) Option {
	return func(c *Component) error {
		c.scalars.Set(name, value)
		return nil
	}
}

// ParentMesh returns the component's parent mesh.
func (c *Component) ParentMesh() ParentMesh {
	return c.parentMesh
}

// SetParentMesh sets parent mesh.
func (c *Component) SetParentMesh(parentMesh ParentMesh) *Component {
	c.parentMesh = parentMesh
	return c
}

// ValidateBeforeAddingToMesh checks if the component is good to be added into mesh.
func (c *Component) ValidateBeforeAddingToMesh() error {
	if c.f == nil {
		return errors.New("activation function is not set")
	}

	return nil
}
