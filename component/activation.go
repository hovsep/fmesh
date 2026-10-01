package component

import (
	"context"
	"errors"
	"fmt"
	"runtime/debug"

	"github.com/hovsep/fmesh/internal/hook"
	"github.com/hovsep/fmesh/port"
	"github.com/hovsep/fmesh/signal"
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
// Each stage recovers its own panic, so every hook group fires exactly once and
// a panic anywhere becomes a Panicked result. AfterActivation runs even when
// BeforeActivation failed.
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

	var failures []error
	for attempt := 1; ; attempt++ {
		var outputs map[*port.Port]*signal.Group
		if c.attempts > 1 {
			outputs = c.snapshotOutputs()
		}

		err := c.f(ctx, c)
		switch {
		case err == nil:
			return c.newActivationResultOK()
		case errors.Is(err, ErrWaitingForInputs):
			return c.newActivationResultWaitingForInputs(err)
		}

		if c.attempts > 1 {
			err = fmt.Errorf("attempt %d of %d: %w", attempt, c.attempts, err)
		}
		failures = append(failures, err)
		if attempt >= c.attempts || ctx.Err() != nil {
			return c.newActivationResultReturnedError(failures...)
		}
		if restoreErr := c.restoreOutputs(ctx, outputs); restoreErr != nil {
			return c.newActivationResultReturnedError(append(failures, restoreErr)...)
		}
	}
}

// WithRetry is a component option that runs the activation function up to
// attempts times while it returns an error. The activation fails only when
// every attempt failed, and its result lists each attempt's error. A panic is
// never retried, and neither is a waiting-for-inputs result. Retrying stops
// early when the context is done. Activation hooks fire once, around all the
// attempts.
//
// Signals a failed attempt put on the output ports are removed before the next
// attempt, so retries do not pile up outputs. The reset goes through the port:
// a changed output port is cleared (firing its OnClear hooks) and, if it held
// signals before the activation, refilled with them (firing OnSignalsAdded).
func WithRetry(attempts int) Option {
	return func(c *Component) error {
		if attempts < 1 {
			return fmt.Errorf("retry attempts must be at least 1, got %d", attempts)
		}
		c.attempts = attempts
		return nil
	}
}

// snapshotOutputs keeps each output port's signal group. Groups are
// copy-on-write, so a port an attempt did not touch keeps the same group.
func (c *Component) snapshotOutputs() map[*port.Port]*signal.Group {
	outputs := make(map[*port.Port]*signal.Group, c.Outputs().Len())
	for _, p := range c.Outputs().AllOrdered() {
		outputs[p] = p.Signals()
	}
	return outputs
}

// restoreOutputs puts back the output signals of a snapshot, touching only the
// ports a failed attempt changed.
func (c *Component) restoreOutputs(ctx context.Context, outputs map[*port.Port]*signal.Group) error {
	for _, p := range c.Outputs().AllOrdered() {
		before := outputs[p]
		if p.Signals() == before {
			continue
		}
		if err := p.Clear(ctx); err != nil {
			return fmt.Errorf("failed to reset output port %q between attempts: %w", p.Name(), err)
		}
		if before != nil {
			if err := p.PutSignalGroups(before); err != nil {
				return fmt.Errorf("failed to reset output port %q between attempts: %w", p.Name(), err)
			}
		}
	}
	return nil
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

// markHookFailed records a failed hook on the result. A panicking hook marks the
// result Panicked, and a Panicked result stays Panicked, so StopOnFirstPanic
// always sees the panic.
func markHookFailed(result *ActivationResult, stage string, err error) {
	if _, panicked := errors.AsType[*PanicError](err); panicked {
		result.SetActivationCode(ActivationCodePanicked)
	} else if result.Code() != ActivationCodePanicked {
		result.SetActivationCode(ActivationCodeHookFailed)
	}
	result.AddActivationError(fmt.Errorf("%s hook failed: %w", stage, err))
}
