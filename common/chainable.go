package common

// Chainable is a base struct for chainable objects.
type Chainable struct {
	err error
}

// NewChainable initializes new chainable.
func NewChainable() *Chainable {
	return &Chainable{}
}

// SetErr sets chainable error.
func (c *Chainable) SetErr(err error) {
	c.err = err
}

// HasErr returns true when chainable has an error.
func (c *Chainable) HasErr() bool {
	return c.err != nil
}

// Err returns chainable error.
func (c *Chainable) Err() error {
	return c.err
}
