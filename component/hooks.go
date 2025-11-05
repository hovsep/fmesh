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
func (h *Hooks) BeforeActivation(fn func(*Component)) {
	h.beforeActivation.Add(fn)
}

// OnSuccess registers a hook called when activation succeeds.
func (h *Hooks) OnSuccess(fn func(*ActivationContext)) {
	h.onSuccess.Add(fn)
}

// OnError registers a hook called when activation returns an error.
func (h *Hooks) OnError(fn func(*ActivationContext)) {
	h.onError.Add(fn)
}

// OnPanic registers a hook called when activation panics.
func (h *Hooks) OnPanic(fn func(*ActivationContext)) {
	h.onPanic.Add(fn)
}

// OnWaitingForInputs registers a hook to be called when component is waiting for inputs.
// Check ctx.Result.Code() to distinguish between Clear and Keep modes.
func (h *Hooks) OnWaitingForInputs(fn func(*ActivationContext)) {
	h.onWaitingForInputs.Add(fn)
}

// AfterActivation registers a hook to be called after activation completes (always).
// This runs regardless of success/error/panic/waiting - like a finally block.
func (h *Hooks) AfterActivation(fn func(*ActivationContext)) {
	h.afterActivation.Add(fn)
}
