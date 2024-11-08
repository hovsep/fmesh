package component

import (
	"fmt"
	"github.com/hovsep/fmesh/common"
)

// @TODO: make type unexported
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
	if c.HasErr() {
		return New("").WithErr(c.Err())
	}

	component, ok := c.components[name]

	if !ok {
		c.SetErr(fmt.Errorf("%w, component name: %s", errNotFound, name))
		return New("").WithErr(c.Err())
	}

	return component
}

// With adds components and returns the collection
func (c *Collection) With(components ...*Component) *Collection {
	if c.HasErr() {
		return c
	}

	for _, component := range components {
		c.components[component.Name()] = component

		if component.HasErr() {
			return c.WithErr(component.Err())
		}
	}

	return c
}

// WithErr returns group with error
func (c *Collection) WithErr(err error) *Collection {
	c.SetErr(err)
	return c
}

// Len returns number of ports in collection
func (c *Collection) Len() int {
	return len(c.components)
}

func (c *Collection) Components() (ComponentsMap, error) {
	if c.HasErr() {
		return nil, c.Err()
	}
	return c.components, nil
}
