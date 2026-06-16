// Package hook provides a generic, type-safe hook system for F-Mesh.
//
// Hooks allow extending framework behavior at specific execution points without
// modifying core logic. All hooks maintain insertion order and support chainable operations.
package hook

// Group is a generic ordered collection of hooks.
// It maintains insertion order and supports triggering all hooks in sequence.
// Hooks return errors for fail-fast behavior.
type Group[T any] struct {
	hooks []func(T) error
}

// NewGroup creates a new hook group.
func NewGroup[T any]() *Group[T] {
	return &Group[T]{}
}

// Add appends a hook to the group, maintaining insertion order.
func (g *Group[T]) Add(hook func(T) error) *Group[T] {
	g.hooks = append(g.hooks, hook)
	return g
}

// All returns all hooks in the group.
func (g *Group[T]) All() []func(T) error {
	return g.hooks
}

// Trigger executes all hooks in order with the provided argument.
// Returns the first error encountered (fail-fast).
func (g *Group[T]) Trigger(arg T) error {
	for _, hook := range g.hooks {
		if err := hook(arg); err != nil {
			return err
		}
	}
	return nil
}

// ForEach applies an action to each hook function.
// Note: Most users should use Trigger() instead.
func (g *Group[T]) ForEach(action func(func(T) error) error) *Group[T] {
	for _, hook := range g.hooks {
		if err := action(hook); err != nil {
			return g
		}
	}
	return g
}

// Clear removes all hooks from the group.
func (g *Group[T]) Clear() *Group[T] {
	g.hooks = make([]func(T) error, 0)
	return g
}

// Len returns the number of hooks in the group.
// Returns 0 if the group has a chainable error.
func (g *Group[T]) Len() int {
	return len(g.hooks)
}

// IsEmpty returns true if the group has no hooks.
func (g *Group[T]) IsEmpty() bool {
	return len(g.hooks) == 0
}
