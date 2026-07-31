package component

import (
	"context"
	"errors"
	"fmt"
	"runtime/debug"

	"github.com/hovsep/fmesh/internal/hook"
)

// WithActivationFunc is a component option that sets the activation function.
func WithActivationFunc(f ActivationFunc) Option {
	return func(c *Component) error {
		c.f = f
		return nil
	}
}

// MaybeActivate tries to run the activation function if all required conditions are met.
// The context is handed to the activation function and to every hook fired around it.
func (c *Component) MaybeActivate(ctx context.Context) *ActivationResult {
	if !c.Inputs().AnyHasSignals() {
		return c.newActivationResultNoInput()
	}

	return c.activate(ctx)
}

// activate executes the activation function and manages hooks.
func (c *Component) activate(ctx context.Context) (result *ActivationResult) {
	if err := c.hooks.beforeActivation.Trigger(ctx, c); err != nil {
		result = c.newActivationResultHookFailed(fmt.Errorf("beforeActivation hook failed: %w", err))
		c.triggerAfterActivation(ctx, result)
		return result
	}

	defer func() {
		if r := recover(); r != nil {
			result = c.newActivationResultPanicked(&PanicError{
				ComponentName: c.Name(),
				Value:         r,
				Stack:         debug.Stack(),
			})
			c.triggerHooksForResult(ctx, result, c.hooks.onPanic)
			c.triggerAfterActivation(ctx, result)
		}
	}()

	// The component's own function runs first, then any OnActivation hooks, as a
	// single sequence — so a hook returning ErrWaitingForInputs suspends the
	// component exactly as the function itself would.
	onActivation := c.hooks.onActivation.All()
	funcs := make([]ActivationFunc, 0, 1+len(onActivation))
	funcs = append(funcs, c.f)
	for _, hook := range onActivation {
		funcs = append(funcs, hook)
	}

	err := Sequential(funcs...)(ctx, c)
	result = c.buildResultAndTriggerHook(ctx, err)
	c.triggerAfterActivation(ctx, result)

	return result
}

// buildResultAndTriggerHook creates the activation result and triggers the appropriate hook.
func (c *Component) buildResultAndTriggerHook(ctx context.Context, err error) *ActivationResult {
	if errors.Is(err, ErrWaitingForInputs) {
		result := c.newActivationResultWaitingForInputs(err)
		c.triggerHooksForResult(ctx, result, c.hooks.onWaitingForInputs)
		return result
	}

	if err != nil {
		result := c.newActivationResultReturnedError(err)
		c.triggerHooksForResult(ctx, result, c.hooks.onError)
		return result
	}

	result := c.newActivationResultOK()
	c.triggerHooksForResult(ctx, result, c.hooks.onSuccess)
	return result
}

// triggerHooksForResult triggers the outcome-specific hook with the activation context.
func (c *Component) triggerHooksForResult(ctx context.Context, result *ActivationResult, hookGroup *hook.Group[*ActivationContext]) {
	if err := hookGroup.Trigger(ctx, &ActivationContext{Component: c, Result: result}); err != nil {
		result.SetActivationCode(ActivationCodeHookFailed).
			AddActivationError(fmt.Errorf("activation hook failed: %w", err))
	}
}

// triggerAfterActivation triggers the AfterActivation hook.
func (c *Component) triggerAfterActivation(ctx context.Context, result *ActivationResult) {
	if err := c.hooks.afterActivation.Trigger(ctx, &ActivationContext{Component: c, Result: result}); err != nil {
		result.SetActivationCode(ActivationCodeHookFailed).
			AddActivationError(fmt.Errorf("afterActivation hook failed: %w", err))
	}
}
