package component

import (
	"fmt"
	"github.com/hovsep/fmesh/common"
	"github.com/hovsep/fmesh/port"
	"log"
)

// Component defines a main building block of FMesh
type Component struct {
	common.NamedEntity
	common.DescribedEntity
	common.LabeledEntity
	*common.Chainable
	inputs  *port.Collection
	outputs *port.Collection
	f       ActivationFunc
	logger  *log.Logger
	state   State
}

// New creates initialized component
func New(name string) *Component {
	return &Component{
		NamedEntity:     common.NewNamedEntity(name),
		DescribedEntity: common.NewDescribedEntity(""),
		LabeledEntity:   common.NewLabeledEntity(nil),
		Chainable:       common.NewChainable(),
		inputs: port.NewCollection().WithDefaultLabels(common.LabelsCollection{
			port.DirectionLabel: port.DirectionIn,
		}),
		outputs: port.NewCollection().WithDefaultLabels(common.LabelsCollection{
			port.DirectionLabel: port.DirectionOut,
		}),
	}
}

// WithDescription sets a description
func (c *Component) WithDescription(description string) *Component {
	if c.HasErr() {
		return c
	}

	c.DescribedEntity = common.NewDescribedEntity(description)
	return c
}

// WithLabels sets labels and returns the component
func (c *Component) WithLabels(labels common.LabelsCollection) *Component {
	if c.HasErr() {
		return c
	}
	c.LabeledEntity.SetLabels(labels)
	return c
}

// propagateChainErrors propagates up all chain errors that might have not been propagated yet
func (c *Component) propagateChainErrors() {
	if c.Inputs().HasErr() {
		c.SetErr(c.Inputs().Err())
		return
	}

	if c.Outputs().HasErr() {
		c.SetErr(c.Outputs().Err())
		return
	}

	for _, p := range c.Inputs().PortsOrNil() {
		if p.HasErr() {
			c.SetErr(p.Err())
			return
		}
	}

	for _, p := range c.Outputs().PortsOrNil() {
		if p.HasErr() {
			c.SetErr(p.Err())
			return
		}
	}
}

// WithErr returns component with error
func (c *Component) WithErr(err error) *Component {
	c.SetErr(err)
	return c
}

// WithLogger creates a new logger prefixed with component name
func (c *Component) WithLogger(logger *log.Logger) *Component {
	if c.HasErr() {
		return c
	}

	if logger == nil {
		return c
	}

	prefix := fmt.Sprintf("%s : ", c.Name())
	c.logger = log.New(logger.Writer(), prefix, logger.Flags())
	return c
}

// Logger returns component logger
func (c *Component) Logger() *log.Logger {
	return c.logger
}
