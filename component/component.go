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
	inputPorts   *port.Collection
	outputPorts  *port.Collection
	f            ActivationFunc
	logger       *log.Logger
	state        State
	parentMesh   ParentMesh
}

// New creates a new component with the specified name.
// This is the starting point for building any component in your F-Mesh.
//
// Example:
//
//	processor := component.New("data-processor").
//	    WithDescription("Processes incoming data").
//	    AddInputs("data", "config").
//	    AddOutputs("result")
func New(name string) *Component {
	return &Component{
		name:         name,
		description:  "",
		labels:       labels.NewCollection(nil),
		chainableErr: nil,
		inputPorts:   port.NewCollection(),
		outputPorts:  port.NewCollection(),
		state:        NewState(),
	}
}

// Name returns the component's name.
func (c *Component) Name() string {
	return c.name
}

// Description returns the component's description.
func (c *Component) Description() string {
	return c.description
}

// WithDescription sets a human-readable description for the component.
// Use this to document what your component does. Helpful for debugging and documentation.
//
// Example:
//
//	c := component.New("validator").
//	    WithDescription("Validates user input against business rules")
func (c *Component) WithDescription(description string) *Component {
	if c.HasChainableErr() {
		return c
	}

	c.description = description
	return c
}

// Labels returns the component's labels collection.
func (c *Component) Labels() *labels.Collection {
	if c.HasChainableErr() {
		return labels.NewCollection(nil).WithChainableErr(c.ChainableErr())
	}
	return c.labels
}

// SetLabels replaces all labels and returns the component for chaining.
func (c *Component) SetLabels(labelMap labels.Map) *Component {
	if c.HasChainableErr() {
		return c
	}
	c.labels.Clear().WithMany(labelMap)
	return c
}

// AddLabels adds or updates labels and returns the component for chaining.
func (c *Component) AddLabels(labelMap labels.Map) *Component {
	if c.HasChainableErr() {
		return c
	}
	c.labels.WithMany(labelMap)
	return c
}

// AddLabel adds or updates a single label and returns the component for chaining.
func (c *Component) AddLabel(name, value string) *Component {
	if c.HasChainableErr() {
		return c
	}
	c.labels.With(name, value)
	return c
}

// ClearLabels removes all labels and returns the component for chaining.
func (c *Component) ClearLabels() *Component {
	if c.HasChainableErr() {
		return c
	}
	c.labels.Clear()
	return c
}

// WithoutLabels removes specific labels and returns the component for chaining.
func (c *Component) WithoutLabels(names ...string) *Component {
	if c.HasChainableErr() {
		return c
	}
	c.labels.Without(names...)
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

	inputPorts, err := c.Inputs().All()
	if err != nil {
		c.WithChainableErr(err)
		return
	}
	for _, p := range inputPorts {
		if p.HasChainableErr() {
			c.WithChainableErr(p.ChainableErr())
			return
		}
	}

	outputPorts, err := c.Outputs().All()
	if err != nil {
		c.WithChainableErr(err)
		return
	}
	for _, p := range outputPorts {
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

// ChainableErr returns the chainable error.
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

// Logger returns the component's logger for debugging and monitoring.
// Use this to log information during component activation.
//
// Example (in activation function):
//
//	this.Logger().Printf("Processing %d items", count)
//	this.Logger().Println("Validation completed successfully")
func (c *Component) Logger() *log.Logger {
	return c.logger
}

// ParentMesh returns the component's parent mesh.
func (c *Component) ParentMesh() ParentMesh {
	return c.parentMesh
}

// WithParentMesh sets parent mesh.
func (c *Component) WithParentMesh(parentMesh ParentMesh) *Component {
	c.parentMesh = parentMesh
	return c
}
