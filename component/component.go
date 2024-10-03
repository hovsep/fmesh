package component

import (
	"errors"
	"fmt"
	"github.com/hovsep/fmesh/port"
)

type ActivationFunc func(inputs port.Collection, outputs port.Collection) error

// Component defines a main building block of FMesh
type Component struct {
	name        string
	description string
	inputs      port.Collection
	outputs     port.Collection
	f           ActivationFunc
}

// New creates initialized component
func New(name string) *Component {
	return &Component{
		name:    name,
		inputs:  port.NewCollection(),
		outputs: port.NewCollection(),
	}
}

// WithDescription sets a description
func (c *Component) WithDescription(description string) *Component {
	c.description = description
	return c
}

// WithInputs ads input ports
func (c *Component) WithInputs(portNames ...string) *Component {
	c.inputs = c.Inputs().With(port.NewGroup(portNames...)...)
	return c
}

// WithOutputs adds output ports
func (c *Component) WithOutputs(portNames ...string) *Component {
	c.outputs = c.Outputs().With(port.NewGroup(portNames...)...)
	return c
}

// WithInputsIndexed creates multiple prefixed ports
func (c *Component) WithInputsIndexed(prefix string, startIndex int, endIndex int) *Component {
	c.inputs = c.Inputs().WithIndexed(prefix, startIndex, endIndex)
	return c
}

// WithOutputsIndexed creates multiple prefixed ports
func (c *Component) WithOutputsIndexed(prefix string, startIndex int, endIndex int) *Component {
	c.outputs = c.Outputs().WithIndexed(prefix, startIndex, endIndex)
	return c
}

// WithActivationFunc sets activation function
func (c *Component) WithActivationFunc(f ActivationFunc) *Component {
	c.f = f
	return c
}

// Name getter
func (c *Component) Name() string {
	return c.name
}

// Description getter
func (c *Component) Description() string {
	return c.description
}

// Inputs getter
func (c *Component) Inputs() port.Collection {
	return c.inputs
}

// Outputs getter
func (c *Component) Outputs() port.Collection {
	return c.outputs
}

// hasActivationFunction checks when activation function is set
func (c *Component) hasActivationFunction() bool {
	return c.f != nil
}

// MaybeActivate tries to run the activation function if all required conditions are met
// @TODO: hide this method from user
func (c *Component) MaybeActivate() (activationResult *ActivationResult) {
	defer func() {
		if r := recover(); r != nil {
			activationResult = c.newActivationResultPanicked(fmt.Errorf("panicked with: %v", r))
		}
	}()

	if !c.hasActivationFunction() {
		//Activation function is not set (maybe useful while the mesh is under development)
		activationResult = c.newActivationResultNoFunction()
		return
	}

	if !c.inputs.AnyHasSignals() {
		//No inputs set, stop here
		activationResult = c.newActivationResultNoInput()
		return
	}

	//Invoke the activation func
	err := c.f(c.Inputs(), c.Outputs())

	if errors.Is(err, errWaitingForInputs) {
		activationResult = c.newActivationResultWaitingForInputs(err)
		return
	}

	if err != nil {
		activationResult = c.newActivationResultReturnedError(err)
		return
	}

	activationResult = c.newActivationResultOK()
	return
}

// FlushOutputs pushed signals out of the component outputs to pipes and clears outputs
func (c *Component) FlushOutputs() {
	for _, out := range c.outputs {
		out.Flush()
	}
}

// ClearInputs clears all input ports
func (c *Component) ClearInputs() {
	c.Inputs().Clear()
}
