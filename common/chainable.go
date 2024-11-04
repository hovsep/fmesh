package common

type Chainable struct {
	err error
}

// NewChainable initialises new chainable
func NewChainable() *Chainable {
	return &Chainable{}
}

// SetErr sets chainable error
func (c *Chainable) SetErr(err error) {
	c.err = err
}

// HasErr returns true when chainable has error
func (c *Chainable) HasErr() bool {
	return c.err != nil
}

// Err returns chainable error
func (c *Chainable) Err() error {
	return c.err
}
