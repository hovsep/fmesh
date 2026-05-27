package component

import (
	"errors"
	"fmt"

	"github.com/hovsep/fmesh/hook"
)

// WithActivationFunc is a component option that sets the activation function.
func WithActivationFunc(f ActivationFunc) Option {
	return func(c *Component) error {
		c.f = f
		return nil
	}
}

// WithActivationFunc sets the activation function on the component and returns the component for chaining.
// This method can be called after construction; the option form is preferred when building inside New().
func (c *Component) WithActivationFunc(f ActivationFunc) *Component {
	c.f = f
	return c
}

// MaybeActivate tries to run the activation function if all required conditions are met.
func (c *Component) MaybeActivate() *ActivationResult {
	if !c.Inputs().AnyHasSignals() {
		return c.newActivationResultNoInput()
	}

	return c.activate()
}

// activate executes the activation function and manages hooks.
func (c *Component) activate() (result *ActivationResult) {
	if err := c.hooks.beforeActivation.Trigger(c); err != nil {
		result = c.newActivationResultHookFailed(fmt.Errorf("beforeActivation hook failed: %w", err))
		c.triggerAfterActivation(result)
		return result
	}

	defer func() {
		if r := recover(); r != nil {
			result = c.newActivationResultPanicked(fmt.Errorf("panicked with: %v", r))
			c.triggerHooksForResult(result, c.hooks.onPanic)
			c.triggerAfterActivation(result)
		}
	}()

	err := c.f(c)
	result = c.buildResultAndTriggerHook(err)
	c.triggerAfterActivation(result)

	return result
}

// buildResultAndTriggerHook creates the activation result and triggers the appropriate hook.
func (c *Component) buildResultAndTriggerHook(err error) *ActivationResult {
	if errors.Is(err, ErrWaitingForInputs) {
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
	if err := hookGroup.Trigger(&ActivationContext{Component: c, Result: result}); err != nil {
		result.WithActivationCode(ActivationCodeHookFailed).
			WithActivationError(fmt.Errorf("activation hook failed: %w", err))
	}
}

// triggerAfterActivation triggers the AfterActivation hook.
func (c *Component) triggerAfterActivation(result *ActivationResult) {
	if err := c.hooks.afterActivation.Trigger(&ActivationContext{Component: c, Result: result}); err != nil {
		result.WithActivationCode(ActivationCodeHookFailed).
			WithActivationError(fmt.Errorf("afterActivation hook failed: %w", err))
	}
}
