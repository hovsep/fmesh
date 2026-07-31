package component

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/hovsep/fmesh/internal/plugin"
	"github.com/hovsep/fmesh/meta"
	"github.com/hovsep/fmesh/port"
)

// Component defines a main building block of FMesh.
type Component struct {
	name         string
	description  string
	labels       *meta.Labels
	scalars      *meta.Scalars
	inputPorts   *port.Collection
	outputPorts  *port.Collection
	f            ActivationFunc
	logger       *log.Logger
	customLogger bool // true when the logger was set explicitly and must not be inherited from the mesh
	state        State
	parentMesh   ParentMesh
	hooks        *Hooks
	plugins      *plugin.Registry[*Component]
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

	if err := c.initPlugins(); err != nil {
		return nil, err
	}

	// Construction happens outside any run, so there is no run context to inherit.
	if err := c.hooks.onCreation.Trigger(context.Background(), c); err != nil {
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

// Scalars returns the component's scalars store.
func (c *Component) Scalars() *meta.Scalars {
	return c.scalars
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

	if err := c.Inputs().ForEach(func(p *port.Port) error {
		if p.ParentComponent() == nil {
			return fmt.Errorf("input port %q has no parent component", p.Name())
		}
		if p.ParentComponent() != c {
			return fmt.Errorf("input port %q has wrong parent component", p.Name())
		}
		return nil
	}); err != nil {
		return err
	}

	return c.Outputs().ForEach(func(p *port.Port) error {
		if p.ParentComponent() == nil {
			return fmt.Errorf("output port %q has no parent component", p.Name())
		}
		if p.ParentComponent() != c {
			return fmt.Errorf("output port %q has wrong parent component", p.Name())
		}
		return nil
	})
}
