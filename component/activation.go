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
		// @TODO: shall we introduce new activation code?
		result = NewActivationResult(c.Name()).
			SetActivated(false).
			WithActivationCode(ActivationCodeUndefined).
			WithActivationError(fmt.Errorf("beforeActivation hook failed: %w", err))
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
// @TODO: maybe we need to do things separately.
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
		// Hook error takes precedence, wrap the activation result error if any
		if result.ActivationError() != nil {
			result.WithActivationError(fmt.Errorf("%w (hook also failed: %w)", result.ActivationError(), err))
		} else {
			result.WithActivationError(fmt.Errorf("activation hook failed: %w", err))
		}
	}
}

// triggerAfterActivation triggers the AfterActivation hook.
func (c *Component) triggerAfterActivation(result *ActivationResult) {
	if err := c.hooks.afterActivation.Trigger(&ActivationContext{Component: c, Result: result}); err != nil {
		// AfterActivation hook error is always appended to result
		if result.ActivationError() != nil {
			result.WithActivationError(fmt.Errorf("%w (afterActivation hook failed: %w)", result.ActivationError(), err))
		} else {
			result.WithActivationError(fmt.Errorf("afterActivation hook failed: %w", err))
		}
	}
}
