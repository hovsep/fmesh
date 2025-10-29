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

// WithInitialState sets the initial state for the component.
// Use this to initialize state values before the first activation.
// State persists across activation cycles, allowing components to remember data.
//
// Example:
//
//	counter := component.New("counter").
//	    AddInputs("trigger").
//	    WithInitialState(func(state component.State) {
//	        state.Set("count", 0)
//	        state.Set("total", 0)
//	    }).
//	    WithActivationFunc(func(this *component.Component) error {
//	        count := this.State().Get("count").(int)
//	        this.State().Set("count", count+1)
//	        return nil
//	    })
func (c *Component) WithInitialState(init func(state State)) *Component {
	if init != nil {
		init(c.state)
	}

	return c
}

// State returns the component's state for reading and writing persistent data.
// Use this in your activation function to access data that persists across activations.
//
// Example (in activation function):
//
//	count := this.State().Get("count").(int)
//	this.State().Set("count", count+1)
//	this.Logger().Printf("Processed %d items total", count)
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
