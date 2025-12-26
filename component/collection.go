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

// NewCollection creates an empty collection.
func NewCollection() *Collection {
	return &Collection{
		chainableErr: nil,
		components:   make(Map),
	}
}

// ByName returns a component by its name.
func (c *Collection) ByName(name string) *Component {
	if c.HasChainableErr() {
		return New("n/a").WithChainableErr(c.ChainableErr())
	}

	component, ok := c.components[name]

	if !ok {
		c.WithChainableErr(fmt.Errorf("%w, component name: %s", errNotFound, name))
		return New("n/a").WithChainableErr(c.ChainableErr())
	}

	return component
}

// Add adds components and returns the collection.
func (c *Collection) Add(components ...*Component) *Collection {
	if c.HasChainableErr() {
		return c
	}

	for _, component := range components {
		if c.AnyMatch(func(existingComponent *Component) bool {
			return existingComponent.Name() == component.Name()
		}) {
			return c.WithChainableErr(fmt.Errorf("component with name '%s' already exists", component.Name()))
		}
		c.components[component.Name()] = component

		if component.HasChainableErr() {
			return c.WithChainableErr(component.ChainableErr())
		}
	}

	return c
}

// Without removes components by name and returns the collection.
func (c *Collection) Without(names ...string) *Collection {
	if c.HasChainableErr() {
		return c
	}

	for _, name := range names {
		delete(c.components, name)
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

// ChainableErr returns the chainable error.
func (c *Collection) ChainableErr() error {
	return c.chainableErr
}

// Len returns the number of components in the collection.
func (c *Collection) Len() int {
	if c.HasChainableErr() {
		return 0
	}
	return len(c.components)
}

// IsEmpty returns true when there are no components in the collection.
func (c *Collection) IsEmpty() bool {
	return c.Len() == 0
}

// Any returns any arbitrary component from the collection.
// Note: Map iteration order is not guaranteed, so this may return different items on each call.
func (c *Collection) Any() *Component {
	if c.HasChainableErr() {
		return New("n/a").WithChainableErr(c.ChainableErr())
	}
	if c.IsEmpty() {
		c.WithChainableErr(ErrNoComponentsInCollection)
		return New("n/a").WithChainableErr(c.ChainableErr())
	}
	// Get arbitrary component from map (order not guaranteed)
	for _, comp := range c.components {
		return comp
	}
	return New("n/a").WithChainableErr(errUnexpectedErrorGettingComponent)
}

// All returns all components as a map.
func (c *Collection) All() (Map, error) {
	if c.HasChainableErr() {
		return nil, c.ChainableErr()
	}
	return c.components, nil
}

// AllMatch returns true if all components match the predicate.
func (c *Collection) AllMatch(predicate Predicate) bool {
	if c.HasChainableErr() {
		return false
	}
	for _, comp := range c.components {
		if !predicate(comp) {
			return false
		}
	}
	return true
}

// AnyMatch returns true if any component matches the predicate.
func (c *Collection) AnyMatch(predicate Predicate) bool {
	if c.HasChainableErr() {
		return false
	}
	for _, comp := range c.components {
		if predicate(comp) {
			return true
		}
	}
	return false
}

// CountMatch returns the number of components that match the predicate.
func (c *Collection) CountMatch(predicate Predicate) int {
	if c.HasChainableErr() {
		return 0
	}
	count := 0
	for _, comp := range c.components {
		if predicate(comp) {
			count++
		}
	}
	return count
}

// FindAny returns any arbitrary component that matches the predicate.
// Note: Map iteration order is not guaranteed, so this may return different items on each call.
func (c *Collection) FindAny(predicate Predicate) *Component {
	if c.HasChainableErr() {
		return New("n/a").WithChainableErr(c.ChainableErr())
	}
	for _, comp := range c.components {
		if predicate(comp) {
			return comp
		}
	}
	c.WithChainableErr(ErrNoComponentMatchesPredicate)
	return New("n/a").WithChainableErr(c.ChainableErr())
}

// Filter returns a new collection with components that match the predicate.
func (c *Collection) Filter(predicate Predicate) *Collection {
	if c.HasChainableErr() {
		return NewCollection().WithChainableErr(c.ChainableErr())
	}
	filtered := NewCollection()
	for _, comp := range c.components {
		if predicate(comp) {
			filtered = filtered.Add(comp)
			if filtered.HasChainableErr() {
				return filtered
			}
		}
	}
	return filtered
}

// Map returns a new collection with components transformed by the mapper function.
func (c *Collection) Map(mapper Mapper) *Collection {
	if c.HasChainableErr() {
		return NewCollection().WithChainableErr(c.ChainableErr())
	}
	mapped := NewCollection()
	for _, comp := range c.components {
		transformedComp := mapper(comp)
		if transformedComp != nil {
			mapped = mapped.Add(transformedComp)
			if mapped.HasChainableErr() {
				return mapped
			}
		}
	}
	return mapped
}

// ForEach applies the action to each component and returns the collection for chaining.
func (c *Collection) ForEach(action func(*Component) error) *Collection {
	if c.HasChainableErr() {
		return c
	}
	for _, comp := range c.components {
		if err := action(comp); err != nil {
			c.chainableErr = err
			return c
		}
	}
	return c
}

// Clear removes all components from the collection.
func (c *Collection) Clear() *Collection {
	if c.HasChainableErr() {
		return c
	}
	c.components = make(Map)
	return c
}
