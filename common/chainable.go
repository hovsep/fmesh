package common

type Chainable struct {
	err error
}

func NewChainable() *Chainable {
	return &Chainable{}
}

func (c *Chainable) SetChainError(err error) {
	c.err = err
}

func (c *Chainable) HasChainError() bool {
	return c.err != nil
}

func (c *Chainable) ChainError() error {
	return c.err
}
