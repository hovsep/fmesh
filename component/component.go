package component

import (
	"fmt"
	"log"

	"github.com/hovsep/fmesh/labels"
	"github.com/hovsep/fmesh/port"
)

// Component defines a main building block of FMesh.
type Component struct {
	name         string
	description  string
	labels       *labels.Collection
	chainableErr error
	inputs       *port.Collection
	outputs      *port.Collection
	f            ActivationFunc
	logger       *log.Logger
	state        State
}

// New creates initialized component.
func New(name string) *Component {
	return &Component{
		name:         name,
		description:  "",
		labels:       labels.NewCollection(nil),
		chainableErr: nil,
		inputs: port.NewCollection().WithDefaultLabels(labels.Map{
			port.DirectionLabel: port.DirectionIn,
		}),
		outputs: port.NewCollection().WithDefaultLabels(labels.Map{
			port.DirectionLabel: port.DirectionOut,
		}),
		state: NewState(),
	}
}

// Name getter.
func (c *Component) Name() string {
	return c.name
}

// Description getter.
func (c *Component) Description() string {
	return c.description
}

// WithDescription sets a description.
func (c *Component) WithDescription(description string) *Component {
	if c.HasChainableErr() {
		return c
	}

	c.description = description
	return c
}

// Labels getter.
func (c *Component) Labels() *labels.Collection {
	if c.HasChainableErr() {
		return labels.NewCollection(nil).WithChainableErr(c.ChainableErr())
	}
	return c.labels
}

// WithLabels sets labels and returns the component.
func (c *Component) WithLabels(labels labels.Map) *Component {
	if c.HasChainableErr() {
		return c
	}
	c.labels.WithMany(labels)
	return c
}

// propagateChainErrors propagates up all chain errors that might have not been propagated yet.
func (c *Component) propagateChainErrors() {
	if c.Inputs().HasChainableErr() {
		c.WithChainableErr(c.Inputs().ChainableErr())
		return
	}

	if c.Outputs().HasChainableErr() {
		c.WithChainableErr(c.Outputs().ChainableErr())
		return
	}

	for _, p := range c.Inputs().PortsOrNil() {
		if p.HasChainableErr() {
			c.WithChainableErr(p.ChainableErr())
			return
		}
	}

	for _, p := range c.Outputs().PortsOrNil() {
		if p.HasChainableErr() {
			c.WithChainableErr(p.ChainableErr())
			return
		}
	}
}

// WithChainableErr sets a chainable error and returns the component.
func (c *Component) WithChainableErr(err error) *Component {
	c.chainableErr = err
	return c
}

// HasChainableErr returns true when a chainable error is set.
func (c *Component) HasChainableErr() bool {
	return c.chainableErr != nil
}

// ChainableErr returns chainable error.
func (c *Component) ChainableErr() error {
	return c.chainableErr
}

// WithLogger creates a new logger prefixed with component name.
func (c *Component) WithLogger(logger *log.Logger) *Component {
	if c.HasChainableErr() {
		return c
	}

	if logger == nil {
		return c
	}

	prefix := fmt.Sprintf("%s: %s ", c.Name(), logger.Prefix())
	c.logger = log.New(logger.Writer(), prefix, logger.Flags())
	return c
}

// Logger returns component logger.
func (c *Component) Logger() *log.Logger {
	return c.logger
}
