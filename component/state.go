package component

type State map[string]any
type StateInitializerFunc func(state State)

// WithInitialState sets initial state (optional)
func (c *Component) WithInitialState(init StateInitializerFunc) *Component {
	c.state = make(map[string]any)
	init(c.state)
	return c
}

// State returns current state
func (c *Component) State() State {
	return c.state
}
