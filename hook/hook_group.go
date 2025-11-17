// Package hook provides a generic, type-safe hook system for F-Mesh.
//
// Hooks allow extending framework behavior at specific execution points without
// modifying core logic. All hooks maintain insertion order and support chainable operations.
package hook

// Group is a generic ordered collection of hooks.
// It maintains insertion order and supports triggering all hooks in sequence.
// Hooks return errors for fail-fast behavior.
type Group[T any] struct {
	hooks        []func(T) error
	chainableErr error
}

// NewGroup creates a new hook group.
func NewGroup[T any]() *Group[T] {
	return &Group[T]{}
}

// Add appends a hook to the group, maintaining insertion order.
func (hg *Group[T]) Add(hook func(T) error) *Group[T] {
	if hg.HasChainableErr() {
		return hg
	}
	hg.hooks = append(hg.hooks, hook)
	return hg
}

// All returns all hooks in the group.
func (hg *Group[T]) All() ([]func(T) error, error) {
	if hg.HasChainableErr() {
		return nil, hg.ChainableErr()
	}
	return hg.hooks, nil
}

// Trigger executes all hooks in order with the provided argument.
// Returns the first error encountered (fail-fast).
func (hg *Group[T]) Trigger(arg T) error {
	if hg.HasChainableErr() {
		return hg.ChainableErr()
	}
	for _, hook := range hg.hooks {
		if err := hook(arg); err != nil {
			return err
		}
	}
	return nil
}

// ForEach applies an action to each hook function.
// Note: Most users should use Trigger() instead.
func (hg *Group[T]) ForEach(action func(func(T) error)) *Group[T] {
	if hg.HasChainableErr() {
		return hg
	}
	for _, hook := range hg.hooks {
		action(hook)
	}
	return hg
}

// Clear removes all hooks from the group.
func (hg *Group[T]) Clear() *Group[T] {
	if hg.HasChainableErr() {
		return hg
	}
	hg.hooks = make([]func(T) error, 0)
	return hg
}

// Len returns the number of hooks in the group.
func (hg *Group[T]) Len() int {
	return len(hg.hooks)
}

// IsEmpty returns true if the group has no hooks.
func (hg *Group[T]) IsEmpty() bool {
	return len(hg.hooks) == 0
}

// WithChainableErr sets a chainable error.
func (hg *Group[T]) WithChainableErr(err error) *Group[T] {
	hg.chainableErr = err
	return hg
}

// HasChainableErr returns true if a chainable error is set.
func (hg *Group[T]) HasChainableErr() bool {
	return hg.chainableErr != nil
}

// ChainableErr returns the current chainable error.
func (hg *Group[T]) ChainableErr() error {
	return hg.chainableErr
}
