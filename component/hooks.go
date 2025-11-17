package component

import "github.com/hovsep/fmesh/hook"

// ActivationContext provides context for activation hooks.
type ActivationContext struct {
	Component *Component
	Result    *ActivationResult
}

// Hooks is a registry of all hook types for Component.
type Hooks struct {
	beforeActivation   *hook.Group[*Component]
	onSuccess          *hook.Group[*ActivationContext]
	onError            *hook.Group[*ActivationContext]
	onPanic            *hook.Group[*ActivationContext]
	onWaitingForInputs *hook.Group[*ActivationContext]
	afterActivation    *hook.Group[*ActivationContext]
}

// NewHooks creates a new hooks registry.
func NewHooks() *Hooks {
	return &Hooks{
		beforeActivation:   hook.NewGroup[*Component](),
		onSuccess:          hook.NewGroup[*ActivationContext](),
		onError:            hook.NewGroup[*ActivationContext](),
		onPanic:            hook.NewGroup[*ActivationContext](),
		onWaitingForInputs: hook.NewGroup[*ActivationContext](),
		afterActivation:    hook.NewGroup[*ActivationContext](),
	}
}

// BeforeActivation registers a hook called before activation.
// Returns the Hooks registry for method chaining.
func (h *Hooks) BeforeActivation(fn func(*Component) error) *Hooks {
	h.beforeActivation.Add(fn)
	return h
}

// OnSuccess registers a hook called when activation succeeds.
// Returns the Hooks registry for method chaining.
func (h *Hooks) OnSuccess(fn func(*ActivationContext) error) *Hooks {
	h.onSuccess.Add(fn)
	return h
}

// OnError registers a hook called when activation returns an error.
// Returns the Hooks registry for method chaining.
func (h *Hooks) OnError(fn func(*ActivationContext) error) *Hooks {
	h.onError.Add(fn)
	return h
}

// OnPanic registers a hook called when activation panics.
// Returns the Hooks registry for method chaining.
func (h *Hooks) OnPanic(fn func(*ActivationContext) error) *Hooks {
	h.onPanic.Add(fn)
	return h
}

// OnWaitingForInputs registers a hook to be called when component is waiting for inputs.
// Check ctx.Result.Code() to distinguish between Clear and Keep modes.
// Returns the Hooks registry for method chaining.
func (h *Hooks) OnWaitingForInputs(fn func(*ActivationContext) error) *Hooks {
	h.onWaitingForInputs.Add(fn)
	return h
}

// AfterActivation registers a hook to be called after activation completes (always).
// This runs regardless of success/error/panic/waiting - like a finally block.
// Returns the Hooks registry for method chaining.
func (h *Hooks) AfterActivation(fn func(*ActivationContext) error) *Hooks {
	h.afterActivation.Add(fn)
	return h
}
