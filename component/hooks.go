package component

import (
	"context"

	"github.com/hovsep/fmesh/internal/hook"
)

// ActivationContext provides context for activation hooks.
type ActivationContext struct {
	Component *Component
	Result    *ActivationResult
}

// Hooks is a registry of all hook types for Component.
type Hooks struct {
	onCreation       *hook.Group[*Component]
	beforeActivation *hook.Group[*Component]
	afterActivation  *hook.Group[*ActivationContext]
}

// newHooks creates a new hook registry.
func newHooks() *Hooks {
	return &Hooks{
		onCreation:       hook.NewGroup[*Component](),
		beforeActivation: hook.NewGroup[*Component](),
		afterActivation:  hook.NewGroup[*ActivationContext](),
	}
}

// OnCreation registers a hook called on state initialization.
func (h *Hooks) OnCreation(fn func(context.Context, *Component) error) *Hooks {
	h.onCreation.Add(fn)
	return h
}

// BeforeActivation registers a hook called before activation.
func (h *Hooks) BeforeActivation(fn func(context.Context, *Component) error) *Hooks {
	h.beforeActivation.Add(fn)
	return h
}

// AfterActivation registers a hook to be called after activation completes (always).
// This runs regardless of success/error/panic/waiting - like a finally block.
// Check Result.Code() on the context to react to one outcome only.
func (h *Hooks) AfterActivation(fn func(context.Context, *ActivationContext) error) *Hooks {
	h.afterActivation.Add(fn)
	return h
}

// SetupHooks configures hooks for the component using a closure.
// All hook registration happens inside the provided function.
func (c *Component) SetupHooks(configure func(*Hooks)) *Component {
	configure(c.hooks)
	return c
}

// WithHooks allows setting hooks during component creation.
func WithHooks(configure func(*Hooks)) Option {
	return func(c *Component) error {
		c.SetupHooks(configure)
		return nil
	}
}
