package fmesh

// HookGroup is a generic ordered collection of hooks.
// It maintains insertion order and supports triggering all hooks in sequence.
type HookGroup[T any] struct {
	hooks        []func(T)
	chainableErr error
}

// NewHookGroup creates a new empty hook group.
func NewHookGroup[T any]() *HookGroup[T] {
	return &HookGroup[T]{
		hooks:        make([]func(T), 0),
		chainableErr: nil,
	}
}

// Add appends a hook to the group, maintaining insertion order.
// Returns the hook group for chaining.
func (hg *HookGroup[T]) Add(hook func(T)) *HookGroup[T] {
	if hg.HasChainableErr() {
		return hg
	}
	hg.hooks = append(hg.hooks, hook)
	return hg
}

// All returns all hooks in the group as a slice.
// Useful for introspection or advanced use cases.
func (hg *HookGroup[T]) All() ([]func(T), error) {
	if hg.HasChainableErr() {
		return nil, hg.ChainableErr()
	}
	return hg.hooks, nil
}

// Trigger executes all hooks in order with the provided argument.
func (hg *HookGroup[T]) Trigger(arg T) {
	if hg.HasChainableErr() {
		return
	}
	for _, hook := range hg.hooks {
		hook(arg)
	}
}

// ForEach applies an action to each hook function and returns the group for chaining.
// Note: This operates on the hook functions themselves, not their results.
// Most users should use Trigger() instead.
func (hg *HookGroup[T]) ForEach(action func(func(T))) *HookGroup[T] {
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
func (hg *HookGroup[T]) Clear() *HookGroup[T] {
	if hg.HasChainableErr() {
		return hg
	}
	hg.hooks = make([]func(T), 0)
	return hg
}

// Len returns the number of hooks in the group.
func (hg *HookGroup[T]) Len() int {
	return len(hg.hooks)
}

// IsEmpty returns true if the group has no hooks.
func (hg *HookGroup[T]) IsEmpty() bool {
	return len(hg.hooks) == 0
}

// WithChainableErr sets a chainable error and returns the hook group.
func (hg *HookGroup[T]) WithChainableErr(err error) *HookGroup[T] {
	hg.chainableErr = err
	return hg
}

// HasChainableErr returns true when a chainable error is set.
func (hg *HookGroup[T]) HasChainableErr() bool {
	return hg.chainableErr != nil
}

// ChainableErr returns the chainable error.
func (hg *HookGroup[T]) ChainableErr() error {
	return hg.chainableErr
}
