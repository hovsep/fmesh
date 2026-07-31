// Package hook provides a generic, type-safe hook system for F-Mesh.
//
// Hooks allow extending framework behavior at specific execution points without
// modifying core logic. All hooks maintain insertion order and support chainable operations.
//
// Every hook takes a context as its first argument, so a hook that does I/O
// participates in the cancellation of the run that triggered it. Hooks fired
// outside a run (construction, wiring, seeding) receive context.Background().
package hook

import "context"

// Group is a generic ordered collection of hooks.
// It maintains insertion order and supports triggering all hooks in sequence.
// Hooks return errors for fail-fast behavior.
type Group[T any] struct {
	hooks []func(context.Context, T) error
}

// NewGroup creates a new hook group.
func NewGroup[T any]() *Group[T] {
	return &Group[T]{}
}

// Add appends a hook to the group, maintaining insertion order.
func (g *Group[T]) Add(hook func(context.Context, T) error) *Group[T] {
	g.hooks = append(g.hooks, hook)
	return g
}

// All returns all hooks in the group.
func (g *Group[T]) All() []func(context.Context, T) error {
	return g.hooks
}

// Trigger executes all hooks in order with the provided argument.
// Returns the first error encountered (fail-fast).
func (g *Group[T]) Trigger(ctx context.Context, arg T) error {
	for _, hook := range g.hooks {
		if err := hook(ctx, arg); err != nil {
			return err
		}
	}
	return nil
}
