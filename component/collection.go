package component

import (
	"fmt"

	"github.com/hovsep/fmesh/common"
)

// Map is a map of components
type Map map[string]*Component

// Collection is a collection of components with useful methods
type Collection struct {
	*common.Chainable
	components Map
}

// NewCollection creates empty collection
func NewCollection() *Collection {
	return &Collection{
		Chainable:  common.NewChainable(),
		components: make(Map),
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

// ByLabelValue returns all components which have a given label with given value
func (c *Collection) ByLabelValue(label, value string) *Collection {
	if c.HasErr() {
		return NewCollection().WithErr(c.Err())
	}

	selectedComponents := NewCollection()

	for _, component := range c.components {
		if component.LabelIs(label, value) {
			selectedComponents = selectedComponents.With(component)
		}
	}

	return selectedComponents
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

// Components returns underlying components map
func (c *Collection) Components() (Map, error) {
	if c.HasErr() {
		return nil, c.Err()
	}
	return c.components, nil
}

// First returns the first component in the collection (ORDER IS NOT GUARANTEED)
func (c *Collection) First() *Component {
	if c.HasErr() {
		return New("").WithErr(c.Err())
	}

	for _, component := range c.components {
		return component
	}

	c.SetErr(errNotFound)
	return New("").WithErr(c.Err())
}
