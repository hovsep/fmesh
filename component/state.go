package component

// State is a key-value storage that persists between activation cycles of a component.
// It allows storing and retrieving arbitrary data using string keys.
//
// This type is inherently thread-safe as each component has a unique instance of State,
// and no two instances of the same component exist concurrently.
type State map[string]any

// NewState creates a clean state.
func NewState() State {
	return make(State)
}

// WithInitialState initializes the component state and returns the component for chaining.
func (c *Component) WithInitialState(init func(state State)) *Component {
	if init != nil {
		init(c.state)
	}

	return c
}

// State returns the component's state.
func (c *Component) State() State {
	return c.state
}

// ResetState cleans the state.
func (c *Component) ResetState() {
	c.state = NewState()
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

// GetTyped retrieves the value associated with the given key from the state
// and asserts it to type T.
func GetTyped[T any](s State, key string) T {
	val, exists := s[key]
	if !exists {
		panic("state key not found: " + key)
	}
	typed, ok := val.(T)
	if !ok {
		panic("state key has wrong type: " + key)
	}
	return typed
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
