package component

import (
	"errors"
	"fmt"
	"github.com/hovsep/fmesh/port"
	"log"
)

type ActivationFunc func(state State, inputs *port.Collection, outputs *port.Collection, log *log.Logger) error

// WithActivationFunc sets activation function
func (c *Component) WithActivationFunc(f ActivationFunc) *Component {
	if c.HasErr() {
		return c
	}

	c.f = f
	return c
}

// hasActivationFunction checks when activation function is set
func (c *Component) hasActivationFunction() bool {
	if c.HasErr() {
		return false
	}

	return c.f != nil
}

// MaybeActivate tries to run the activation function if all required conditions are met
func (c *Component) MaybeActivate() (activationResult *ActivationResult) {
	c.propagateChainErrors()

	if c.HasErr() {
		activationResult = NewActivationResult(c.Name()).WithErr(c.Err())
		return
	}

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

	if !c.Inputs().AnyHasSignals() {
		//No inputs set, stop here
		activationResult = c.newActivationResultNoInput()
		return
	}

	//Invoke the activation func
	err := c.f(c.state, c.Inputs(), c.Outputs(), c.Logger())

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
