package hook

// Group is a generic ordered collection of hooks.
// It maintains insertion order and supports triggering all hooks in sequence.
type Group[T any] struct {
	hooks        []func(T)
	chainableErr error
}

// NewGroup creates a new empty hook group.
func NewGroup[T any]() *Group[T] {
	return &Group[T]{
		hooks:        make([]func(T), 0),
		chainableErr: nil,
	}
}

// Add appends a hook to the group, maintaining insertion order.
// Returns the hook group for chaining.
func (hg *Group[T]) Add(hook func(T)) *Group[T] {
	if hg.HasChainableErr() {
		return hg
	}
	hg.hooks = append(hg.hooks, hook)
	return hg
}

// All returns all hooks in the group as a slice.
// Useful for introspection or advanced use cases.
func (hg *Group[T]) All() ([]func(T), error) {
	if hg.HasChainableErr() {
		return nil, hg.ChainableErr()
	}
	return hg.hooks, nil
}

// Trigger executes all hooks in order with the provided argument.
func (hg *Group[T]) Trigger(arg T) {
	hg.ForEach(func(hook func(T)) {
		hook(arg)
	})
}

// ForEach applies an action to each hook function and returns the group for chaining.
// Note: This operates on the hook functions themselves, not their results.
// Most users should use Trigger() instead.
func (hg *Group[T]) ForEach(action func(func(T))) *Group[T] {
	if hg.HasChainableErr() {
		return hg
	}
	for _, hook := range hg.hooks {
		action(hook)
	}
	return hg
}

// Clear removes all hooks from the group and returns it for chaining.
// Useful for testing or resetting hook state.
func (hg *Group[T]) Clear() *Group[T] {
	if hg.HasChainableErr() {
		return hg
	}
	hg.hooks = make([]func(T), 0)
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

// WithChainableErr sets a chainable error and returns the hook group.
func (hg *Group[T]) WithChainableErr(err error) *Group[T] {
	hg.chainableErr = err
	return hg
}

// HasChainableErr returns true when a chainable error is set.
func (hg *Group[T]) HasChainableErr() bool {
	return hg.chainableErr != nil
}

// ChainableErr returns the chainable error.
func (hg *Group[T]) ChainableErr() error {
	return hg.chainableErr
}
