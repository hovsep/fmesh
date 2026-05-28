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
		state:       NewState(),
		hooks:       NewHooks(),
	}
	for _, opt := range opts {
		if err := opt(c); err != nil {
			return nil, fmt.Errorf("component %q option failed: %w", name, err)
		}
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

// SetDescription sets the component description.
func (c *Component) SetDescription(description string) *Component {
	c.description = description
	return c
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

// WithLabelOption is a component constructor option that adds or updates a single label.
func WithLabelOption(name, value string) Option {
	return func(c *Component) error {
		c.labels.Set(name, value)
		return nil
	}
}

// WithScalarOption is a component constructor option that adds or updates a single scalar.
func WithScalarOption(name string, value float64) Option {
	return func(c *Component) error {
		c.scalars.Set(name, value)
		return nil
	}
}

// WithLogger creates a new logger prefixed with component name.
func (c *Component) WithLogger(logger *log.Logger) *Component {
	if logger == nil {
		return c
	}

	prefix := fmt.Sprintf("%s: %s ", c.Name(), logger.Prefix())
	c.logger = log.New(logger.Writer(), prefix, logger.Flags())
	return c
}

// Logger returns the component's logger.
func (c *Component) Logger() *log.Logger {
	return c.logger
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

// SetupHooks configures hooks for the component using a closure.
// All hook registration happens inside the provided function.
func (c *Component) SetupHooks(configure func(*Hooks)) *Component {
	configure(c.hooks)
	return c
}

// ValidateBeforeAddingToMesh checks if the component is good to be added into mesh.
func (c *Component) ValidateBeforeAddingToMesh() error {
	if c.f == nil {
		return errors.New("activation function is not set")
	}

	return nil
}

// ValidateBeforeActivating checks if the component is good to be activated.
func (c *Component) ValidateBeforeActivating() error {
	if c.ParentMesh() == nil {
		return errors.New("parent mesh is not set")
	}

	return nil
}
