// Package hook provides a generic, type-safe hook system for F-Mesh.
//
// Hooks fire in insertion order. Hooks fired outside a run (construction,
// wiring, seeding) receive context.Background().
package hook

import (
	"context"
	"errors"
)

// Group is a generic ordered collection of hooks.
// It maintains insertion order and supports triggering all hooks in sequence.
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

// IsEmpty reports whether the group has no hooks. Hot paths check it before
// building a Trigger argument, which escapes to the heap even with no hooks.
func (g *Group[T]) IsEmpty() bool {
	return len(g.hooks) == 0
}

// Trigger executes all hooks in order with the provided argument.
// Returns the first error encountered (fail-fast). It is for hooks that run
// before an action and can stop it.
func (g *Group[T]) Trigger(ctx context.Context, arg T) error {
	for _, hook := range g.hooks {
		if err := hook(ctx, arg); err != nil {
			return err
		}
	}
	return nil
}

// TriggerAll executes every hook in order, even after one fails, and returns
// their errors joined. It is for hooks that observe an action already done:
// one failing observer must not hide the event from the others, such as a
// plugin that flushes in AfterRun.
func (g *Group[T]) TriggerAll(ctx context.Context, arg T) error {
	var errs []error
	for _, hook := range g.hooks {
		if err := hook(ctx, arg); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}
