package component

import "fmt"

// State is a key-value storage that persists between activation cycles of a component.
// It allows storing and retrieving arbitrary data using string keys.
//
// State is a plain map with no locking. That is safe for the common case —
// each component has its own instance and activates on one goroutine at a
// time — but not in general: sharing one State between components, or touching
// it from port hooks (which can fire on other components' activation
// goroutines), is a data race.
type State map[string]any

// newState creates a new component state.
func newState() State {
	return make(State)
}

// WithInitialState is a component constructor option that initializes the component state.
func WithInitialState(init func(state State)) Option {
	return func(c *Component) error {
		if init != nil {
			init(c.state)
		}
		return nil
	}
}

// State returns the component's state.
func (c *Component) State() State {
	return c.state
}

// ResetState resets the component state.
func (c *Component) ResetState() {
	c.state = newState()
}

// Has checks if the given key exists in the state.
func (s State) Has(key string) bool {
	_, exists := s[key]
	return exists
}

// Get returns the value by key or nil when the key does not exist.
func (s State) Get(key string) any {
	return s[key]
}

// GetOrDefault returns the value by key or defaultValue when the key does not exist.
func (s State) GetOrDefault(key string, defaultValue any) any {
	if value, exists := s[key]; exists {
		return value
	}

	return defaultValue
}

// Set sets the value for the given key.
func (s State) Set(key string, value any) {
	s[key] = value
}

// Delete deletes the key.
func (s State) Delete(key string) {
	delete(s, key)
}

// GetTyped returns the value under key as T, or an error when the key is
// missing or holds another type.
func (s State) GetTyped[T any](key string) (T, error) {
	var zero T
	val, exists := s[key]
	if !exists {
		return zero, fmt.Errorf("state key %q not found", key)
	}
	typed, ok := val.(T)
	if !ok {
		return zero, fmt.Errorf("state key %q is %T, not %T", key, val, zero)
	}
	return typed, nil
}

// SetIfAbsent sets the value for the given key only if the key does not already exist.
// Returns true if the value was set, false if the key was already present.
func (s State) SetIfAbsent(key string, value any) bool {
	if _, exists := s[key]; exists {
		return false
	}
	s[key] = value
	return true
}

// Upsert applies the given function to the value associated with the key,
// replacing it with the result.
// Creates a new key if not exists.
func (s State) Upsert(key string, fn func(old any) any) {
	s[key] = fn(s[key])
}

// Update applies the given function to the value associated with the key
// only if the key exists in the state.
// Returns true if the key existed and the function was applied, false otherwise.
func (s State) Update(key string, fn func(old any) any) bool {
	old, exists := s[key]
	if !exists {
		return false
	}
	s[key] = fn(old)
	return true
}

// UpdateAndGet is Upsert plus the value it stored, for the common case where the
// new value is needed immediately — a counter that has to be emitted in the same
// activation, say. Reading it back with Get would work but repeats the key, and
// a second map lookup is a second chance to mistype it.
//
// Like Upsert and unlike Update, this creates the key when absent: fn receives
// nil and its result is stored, so the returned value is always the one now in
// the state.
func (s State) UpdateAndGet(key string, fn func(old any) any) any {
	updated := fn(s[key])
	s[key] = updated
	return updated
}
