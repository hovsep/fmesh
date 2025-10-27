package component

import (
	"fmt"
)

// Map is a map of components.
type Map map[string]*Component

// Collection is a collection of components with useful methods.
type Collection struct {
	chainableErr error
	components   Map
}

// NewCollection creates empty collection.
func NewCollection() *Collection {
	return &Collection{
		chainableErr: nil,
		components:   make(Map),
	}
}

// ByName returns a component by its name.
func (c *Collection) ByName(name string) *Component {
	if c.HasChainableErr() {
		return New("").WithChainableErr(c.ChainableErr())
	}

	component, ok := c.components[name]

	if !ok {
		c.WithChainableErr(fmt.Errorf("%w, component name: %s", errNotFound, name))
		return New("").WithChainableErr(c.ChainableErr())
	}

	return component
}

// ByLabelValue returns all components that have a given label with a given value.
func (c *Collection) ByLabelValue(label, value string) *Collection {
	if c.HasChainableErr() {
		return NewCollection().WithChainableErr(c.ChainableErr())
	}

	selectedComponents := NewCollection()

	for _, component := range c.components {
		if component.labels.ValueIs(label, value) {
			selectedComponents = selectedComponents.With(component)
		}
	}

	return selectedComponents
}

// With adds components and returns the collection.
func (c *Collection) With(components ...*Component) *Collection {
	if c.HasChainableErr() {
		return c
	}

	for _, component := range components {
		c.components[component.Name()] = component

		if component.HasChainableErr() {
			return c.WithChainableErr(component.ChainableErr())
		}
	}

	return c
}

// WithChainableErr sets a chainable error and returns the collection.
func (c *Collection) WithChainableErr(err error) *Collection {
	c.chainableErr = err
	return c
}

// HasChainableErr returns true when a chainable error is set.
func (c *Collection) HasChainableErr() bool {
	return c.chainableErr != nil
}

// ChainableErr returns chainable error.
func (c *Collection) ChainableErr() error {
	return c.chainableErr
}

// Len returns the number of ports in a collection.
func (c *Collection) Len() int {
	return len(c.components)
}

// All returns underlying components map.
func (c *Collection) All() (Map, error) {
	if c.HasChainableErr() {
		return nil, c.ChainableErr()
	}
	return c.components, nil
}

// AllOrDefault returns all components or a default value in case of any error.
func (c *Collection) AllOrDefault(defaultValue Map) Map {
	if c.HasChainableErr() {
		return defaultValue
	}
	return c.components
}

// AllOrNil returns all components or nil in case of any error.
func (c *Collection) AllOrNil() Map {
	return c.AllOrDefault(nil)
}

// One returns an arbitrary component from the collection without guaranteeing order.
// This is useful when the collection is expected to contain exactly one component.
// If the collection is empty, it sets an error and returns a placeholder component.
func (c *Collection) One() *Component {
	if c.HasChainableErr() {
		return New("").WithChainableErr(c.ChainableErr())
	}

	for _, component := range c.components {
		return component
	}

	c.WithChainableErr(errNotFound)
	return New("").WithChainableErr(c.ChainableErr())
}
