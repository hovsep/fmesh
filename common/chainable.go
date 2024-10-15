package common

type Chainable struct {
	err error
}

func (c *Chainable) SetError(err error) {
	c.err = err
}

func (c *Chainable) HasError() bool {
	return c.err != nil
}

func (c *Chainable) Error() error {
	return c.err
}
