package component

import (
	"errors"
	"fmt"
)

// ActivationFunc is the function that will be called when the component is activated.
type ActivationFunc func(this *Component) error

// WithActivationFunc sets the activation function.
func (c *Component) WithActivationFunc(f ActivationFunc) *Component {
	if c.HasChainableErr() {
		return c
	}

	c.f = f
	return c
}

// hasActivationFunction checks when the activation function is set.
func (c *Component) hasActivationFunction() bool {
	if c.HasChainableErr() {
		return false
	}

	return c.f != nil
}

// MaybeActivate tries to run the activation function if all required conditions are met.
func (c *Component) MaybeActivate() (activationResult *ActivationResult) {
	c.propagateChainErrors()

	if c.HasChainableErr() {
		return NewActivationResult(c.Name()).WithChainableErr(c.ChainableErr())
	}

	defer func() {
		if r := recover(); r != nil {
			activationResult = c.newActivationResultPanicked(fmt.Errorf("panicked with: %v", r))
		}
	}()

	if !c.hasActivationFunction() {
		// Activation function is not set (maybe useful while the mesh is under development)
		return c.newActivationResultNoFunction()
	}

	if !c.Inputs().AnyHasSignals() {
		// No inputs set, stop here
		return c.newActivationResultNoInput()
	}

	// Invoke the activation func
	err := c.f(c)

	if errors.Is(err, errWaitingForInputs) {
		return c.newActivationResultWaitingForInputs(err)
	}

	if err != nil {
		return c.newActivationResultReturnedError(err)
	}

	return c.newActivationResultOK()
}
