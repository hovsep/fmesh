package component

import (
	"errors"
	"github.com/hovsep/fmesh/common"
)

type ComponentsMap map[string]*Component

// Collection is a collection of components with useful methods
type Collection struct {
	*common.Chainable
	components ComponentsMap
}

// NewCollection creates empty collection
func NewCollection() *Collection {
	return &Collection{
		Chainable:  common.NewChainable(),
		components: make(ComponentsMap),
	}
}

// ByName returns a component by its name
func (c *Collection) ByName(name string) *Component {
	if c.HasChainError() {
		return nil
	}

	component, ok := c.components[name]

	if !ok {
		c.SetChainError(errors.New("component not found"))
		return nil
	}

	return component
}

// With adds components and returns the collection
func (c *Collection) With(components ...*Component) *Collection {
	if c.HasChainError() {
		return c
	}

	for _, component := range components {
		c.components[component.Name()] = component

		if component.HasChainError() {
			return c.WithChainError(component.ChainError())
		}
	}

	return c
}

// WithChainError returns group with error
func (c *Collection) WithChainError(err error) *Collection {
	c.SetChainError(err)
	return c
}

// Len returns number of ports in collection
func (c *Collection) Len() int {
	return len(c.components)
}

func (c *Collection) Components() (ComponentsMap, error) {
	if c.HasChainError() {
		return nil, c.ChainError()
	}
	return c.components, nil
}
