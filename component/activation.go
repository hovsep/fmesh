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

// activate runs the activation function between its hooks.
//
// Every stage recovers its own panics, so each hook group fires exactly once and
// a panic anywhere — in a hook or in the function — becomes a result instead of
// crashing the process from the activation goroutine. AfterActivation is the
// finally block: it runs even when BeforeActivation failed.
func (c *Component) activate(ctx context.Context) *ActivationResult {
	var result *ActivationResult
	if err := triggerRecovering(ctx, c, c.hooks.beforeActivation, c); err != nil {
		result = NewActivationResult(c.Name())
		markHookFailed(result, "beforeActivation", err)
	} else {
		result = c.runActivationFunc(ctx)
	}

	// The IsEmpty guard keeps the context struct off the heap in the common
	// no-hook case — it escapes whether or not anything reads it.
	if !c.hooks.afterActivation.IsEmpty() {
		if err := triggerRecovering(ctx, c, c.hooks.afterActivation, &ActivationContext{Component: c, Result: result}); err != nil {
			markHookFailed(result, "afterActivation", err)
		}
	}
	return result
}

// runActivationFunc runs the activation function and turns its outcome into a result.
func (c *Component) runActivationFunc(ctx context.Context) (result *ActivationResult) {
	defer func() {
		if r := recover(); r != nil {
			result = c.newActivationResultPanicked(c.panicError(r))
		}
	}()

	err := c.f(ctx, c)
	switch {
	case errors.Is(err, ErrWaitingForInputs):
		return c.newActivationResultWaitingForInputs(err)
	case err != nil:
		return c.newActivationResultReturnedError(err)
	default:
		return c.newActivationResultOK()
	}
}

// triggerRecovering triggers a hook group, turning a panic in any hook into a
// *PanicError.
func triggerRecovering[T any](ctx context.Context, c *Component, group *hook.Group[T], arg T) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = c.panicError(r)
		}
	}()
	return group.Trigger(ctx, arg)
}

// panicError captures a recovered value. Called from the deferred recover, the
// stack still holds the frames that panicked.
func (c *Component) panicError(r any) *PanicError {
	return &PanicError{
		ComponentName: c.Name(),
		Value:         r,
		Stack:         debug.Stack(),
	}
}

// markHookFailed records a failed hook on the result. A hook that panicked marks
// the result as a panic, so StopOnFirstPanic stops on it just as it does on a
// panicking activation function.
func markHookFailed(result *ActivationResult, stage string, err error) {
	code := ActivationCodeHookFailed
	if _, panicked := errors.AsType[*PanicError](err); panicked {
		code = ActivationCodePanicked
	}
	result.SetActivationCode(code).
		AddActivationError(fmt.Errorf("%s hook failed: %w", stage, err))
}
