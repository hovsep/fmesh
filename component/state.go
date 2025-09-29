package component

// State is a key-value storage that persists between activation cycles of a component.
// It allows storing and retrieving arbitrary data using string keys.
//
// This type is inherently thread-safe as each component has a unique instance of State,
// and no two instances of the same component exist concurrently.
type State map[string]any

// NewState creates clean state.
func NewState() State {
	return make(State)
}

// WithInitialState sets initial state (optional).
func (c *Component) WithInitialState(init func(state State)) *Component {
	if init != nil {
		init(c.state)
	}

	return c
}

// State returns current state.
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

// Set upserts the given key value.
func (s State) Set(key string, value any) {
	s[key] = value
}

// Delete deletes the key.
func (s State) Delete(key string) {
	delete(s, key)
}
