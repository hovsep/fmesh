package component

import (
	"errors"
	"fmt"

	"github.com/hovsep/fmesh/hook"
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
func (c *Component) MaybeActivate() *ActivationResult {
	c.propagateChainErrors()

	if c.HasChainableErr() {
		return NewActivationResult(c.Name()).WithChainableErr(c.ChainableErr())
	}

	if !c.hasActivationFunction() {
		return c.newActivationResultNoFunction()
	}

	if !c.Inputs().AnyHasSignals() {
		return c.newActivationResultNoInput()
	}

	return c.activate()
}

// activate executes the activation function and manages hooks.
func (c *Component) activate() (result *ActivationResult) {
	c.hooks.beforeActivation.Trigger(c)

	defer func() {
		if r := recover(); r != nil {
			result = c.newActivationResultPanicked(fmt.Errorf("panicked with: %v", r))
			c.triggerHooksForResult(result, c.hooks.onPanic)
			c.hooks.afterActivation.Trigger(&ActivationContext{Component: c, Result: result})
		}
	}()

	err := c.f(c)
	result = c.buildResultAndTriggerHook(err)
	c.hooks.afterActivation.Trigger(&ActivationContext{Component: c, Result: result})

	return result
}

// buildResultAndTriggerHook creates the activation result and triggers the appropriate hook.
func (c *Component) buildResultAndTriggerHook(err error) *ActivationResult {
	if errors.Is(err, errWaitingForInputs) {
		result := c.newActivationResultWaitingForInputs(err)
		c.triggerHooksForResult(result, c.hooks.onWaitingForInputs)
		return result
	}

	if err != nil {
		result := c.newActivationResultReturnedError(err)
		c.triggerHooksForResult(result, c.hooks.onError)
		return result
	}

	result := c.newActivationResultOK()
	c.triggerHooksForResult(result, c.hooks.onSuccess)
	return result
}

// triggerHooksForResult triggers the outcome-specific hook with the activation context.
func (c *Component) triggerHooksForResult(result *ActivationResult, hookGroup *hook.Group[*ActivationContext]) {
	hookGroup.Trigger(&ActivationContext{Component: c, Result: result})
}
