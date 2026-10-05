// Package hook provides a generic, type-safe hook system for F-Mesh.
//
// Hooks allow extending framework behavior at specific execution points without
// modifying core logic. All hooks maintain insertion order and support chainable operations.
//
// Every hook takes a context as its first argument, so a hook that does I/O
// participates in the cancellation of the run that triggered it. Hooks fired
// outside a run (construction, wiring, seeding) receive context.Background().
package hook

import (
	"context"
	"errors"
)

// Group is a generic ordered collection of hooks.
// It maintains insertion order and supports triggering all hooks in sequence.
type Group[T any] struct {
	hooks []func(context.Context, T) error
	// runAll makes Trigger run every hook even after one fails.
	runAll bool
}

// NewGroup creates a fail-fast hook group, for hooks that run before an action
// and can stop it: the first error skips the remaining hooks.
func NewGroup[T any]() *Group[T] {
	return &Group[T]{}
}

// NewObserverGroup creates a hook group for hooks that run after an action and
// cannot stop it: every hook runs, and their errors are joined. One failing
// observer must not hide the event from the others, such as a plugin that
// flushes in AfterRun.
func NewObserverGroup[T any]() *Group[T] {
	return &Group[T]{runAll: true}
}

// Add appends a hook to the group, maintaining insertion order.
func (g *Group[T]) Add(hook func(context.Context, T) error) *Group[T] {
	g.hooks = append(g.hooks, hook)
	return g
}

// IsEmpty reports whether the group has no hooks.
//
// Hot paths use it to skip building a context struct that nothing will read:
// Trigger's argument reaches an indirect call, so it escapes to the heap whether
// or not any hook is registered.
func (g *Group[T]) IsEmpty() bool {
	return len(g.hooks) == 0
}

// Trigger executes the hooks in order with the provided argument. A fail-fast
// group returns the first error; an observer group runs every hook and returns
// their errors joined.
func (g *Group[T]) Trigger(ctx context.Context, arg T) error {
	var errs []error
	for _, hook := range g.hooks {
		if err := hook(ctx, arg); err != nil {
			if !g.runAll {
				return err
			}
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}
