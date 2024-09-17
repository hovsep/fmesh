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

// NewComponent creates a new empty component
func NewComponent(name string) *Component {
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
	c.inputs = c.inputs.Add(port.NewGroup(portNames...)...)
	return c
}

// WithOutputs adds output ports
func (c *Component) WithOutputs(portNames ...string) *Component {
	c.outputs = c.outputs.Add(port.NewGroup(portNames...)...)
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
func (c *Component) MaybeActivate() (activationResult *ActivationResult) {
	defer func() {
		if r := recover(); r != nil {
			//Clear inputs and exit
			c.inputs.ClearSignals()
			activationResult = c.newActivationCodePanicked(fmt.Errorf("panicked with: %v", r))
		}
	}()

	if !c.hasActivationFunction() {
		//Activation function is not set (maybe useful while the mesh is under development)
		activationResult = c.newActivationCodeNoFunction()

		return
	}

	if !c.inputs.AnyHasSignals() {
		//No inputs set, stop here
		activationResult = c.newActivationCodeNoInput()

		return
	}

	//Run the computation
	err := c.f(c.inputs, c.outputs)

	if errors.Is(err, errWaitingForInputs) {
		activationResult = c.newActivationCodeWaitingForInput()

		if !errors.Is(err, errWaitingForInputsKeep) {
			c.inputs.ClearSignals()
		}

		return
	}

	//Clear inputs
	c.inputs.ClearSignals()

	if err != nil {
		activationResult = c.newActivationCodeReturnedError(err)

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
