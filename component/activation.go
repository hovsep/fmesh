package component

import (
	"errors"
	"fmt"
)

// WithActivationFunc sets the activation function and returns the component for chaining.
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

	if !c.hasActivationFunction() {
		// Activation function is not set (maybe useful while the mesh is under development)
		return c.newActivationResultNoFunction()
	}

	if !c.Inputs().AnyHasSignals() {
		// No inputs set, stop here
		return c.newActivationResultNoInput()
	}

	// Component will activate - trigger before hook
	c.hooks.beforeActivation.Trigger(c)

	// Panic recovery with hook support
	defer func() {
		if r := recover(); r != nil {
			activationResult = c.newActivationResultPanicked(fmt.Errorf("panicked with: %v", r))

			// Trigger panic hook
			ctx := &ActivationContext{
				Component: c,
				Result:    activationResult,
			}
			c.hooks.onPanic.Trigger(ctx)
			c.hooks.afterActivation.Trigger(ctx)
		}
	}()

	// Invoke the activation func
	err := c.f(c)

	// Build activation result and trigger outcome-specific hook
	if errors.Is(err, errWaitingForInputs) {
		activationResult = c.newActivationResultWaitingForInputs(err)
		c.hooks.onWaitingForInputs.Trigger(&ActivationContext{
			Component: c,
			Result:    activationResult,
		})
	} else if err != nil {
		activationResult = c.newActivationResultReturnedError(err)
		c.hooks.onError.Trigger(&ActivationContext{
			Component: c,
			Result:    activationResult,
		})
	} else {
		activationResult = c.newActivationResultOK()
		c.hooks.onSuccess.Trigger(&ActivationContext{
			Component: c,
			Result:    activationResult,
		})
	}

	// Always trigger after hook
	c.hooks.afterActivation.Trigger(&ActivationContext{
		Component: c,
		Result:    activationResult,
	})

	return activationResult
}
