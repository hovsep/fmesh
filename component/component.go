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
	c.inputs = c.Inputs().Add(port.NewGroup(portNames...)...)
	return c
}

// WithOutputs adds output ports
func (c *Component) WithOutputs(portNames ...string) *Component {
	c.outputs = c.Outputs().Add(port.NewGroup(portNames...)...)
	return c
}

// WithInputsIndexed creates multiple prefixed ports
func (c *Component) WithInputsIndexed(prefix string, startIndex int, endIndex int) *Component {
	c.inputs = c.Inputs().AddIndexed(prefix, startIndex, endIndex)
	return c
}

// WithOutputsIndexed creates multiple prefixed ports
func (c *Component) WithOutputsIndexed(prefix string, startIndex int, endIndex int) *Component {
	c.outputs = c.Outputs().AddIndexed(prefix, startIndex, endIndex)
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
	stateBeforeActivation := c.getStateSnapshot()

	defer func() {
		if r := recover(); r != nil {
			activationResult = c.newActivationResultPanicked(fmt.Errorf("panicked with: %v", r)).
				WithStateBefore(stateBeforeActivation).
				WithStateAfter(c.getStateSnapshot())
		}
	}()

	if !c.hasActivationFunction() {
		//Activation function is not set (maybe useful while the mesh is under development)
		activationResult = c.newActivationResultNoFunction().
			WithStateBefore(stateBeforeActivation).
			WithStateAfter(c.getStateSnapshot())

		return
	}

	if !c.inputs.AnyHasSignals() {
		//No inputs set, stop here
		activationResult = c.newActivationResultNoInput().
			WithStateBefore(stateBeforeActivation).
			WithStateAfter(c.getStateSnapshot())
		return
	}

	//Invoke the activation func
	err := c.f(c.Inputs(), c.Outputs())

	if errors.Is(err, errWaitingForInputs) {
		activationResult = c.newActivationResultWaitingForInput().
			WithStateBefore(stateBeforeActivation).
			WithStateAfter(c.getStateSnapshot())

		return
	}

	if err != nil {
		activationResult = c.newActivationResultReturnedError(err).
			WithStateBefore(stateBeforeActivation).
			WithStateAfter(c.getStateSnapshot())

		return
	}

	activationResult = c.newActivationResultOK().
		WithStateBefore(stateBeforeActivation).
		WithStateAfter(c.getStateSnapshot())

	return
}

// FlushInputs ...
// @TODO: hide this method from user
func (c *Component) FlushInputs(activationResult *ActivationResult, keepInputSignals bool) {
	c.Inputs().Flush()
	if !keepInputSignals {
		// Inputs can not be just cleared, instead we remove signals which
		// have been used (been set on inputs) during the last activation cycle
		// thus not affecting ones the component could have been received from i2i pipes
		for portName, p := range c.Inputs() {
			p.DisposeSignals(activationResult.StateBefore().InputPortsMetadata()[portName].SignalBufferLen)
		}
	}
}

// FlushOutputs ...
// @TODO: hide this method from user
func (c *Component) FlushOutputs(activationResult *ActivationResult) {
	for portName, p := range c.Outputs() {
		p.FlushAndDispose(activationResult.StateAfter().OutputPortsMetadata()[portName].SignalBufferLen)
	}
}
