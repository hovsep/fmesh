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
		return New("").WithChainableErr(c.ChainableErr())
	}
	if c.IsEmpty() {
		c.WithChainableErr(ErrNoComponentsInCollection)
		return New("").WithChainableErr(c.ChainableErr())
	}
	// Get arbitrary component from map (order not guaranteed)
	for _, comp := range c.components {
		return comp
	}
	return New("").WithChainableErr(errUnexpectedErrorGettingComponent)
}

// AnyOrDefault returns any arbitrary component or the provided default.
func (c *Collection) AnyOrDefault(defaultComponent *Component) *Component {
	if c.HasChainableErr() || c.IsEmpty() {
		return defaultComponent
	}
	for _, comp := range c.components {
		return comp
	}
	return defaultComponent
}

// AnyOrNil returns any arbitrary component or nil.
func (c *Collection) AnyOrNil() *Component {
	if c.HasChainableErr() || c.IsEmpty() {
		return nil
	}
	for _, comp := range c.components {
		return comp
	}
	return nil
}

// AllAsSlice returns all components as Components wrapper type.
func (c *Collection) AllAsSlice() (Components, error) {
	if c.HasChainableErr() {
		return nil, c.ChainableErr()
	}
	components := make([]*Component, 0, len(c.components))
	for _, comp := range c.components {
		components = append(components, comp)
	}
	return Components(components), nil
}

// AllAsSliceOrDefault returns all components as Components wrapper or the provided default.
func (c *Collection) AllAsSliceOrDefault(defaultComponents Components) Components {
	components, err := c.AllAsSlice()
	if err != nil {
		return defaultComponents
	}
	return components
}

// AllAsSliceOrNil returns all components as Components wrapper or nil in case of error.
func (c *Collection) AllAsSliceOrNil() Components {
	return c.AllAsSliceOrDefault(nil)
}

// AllAsMap returns all components as a map.
func (c *Collection) AllAsMap() (Map, error) {
	if c.HasChainableErr() {
		return nil, c.ChainableErr()
	}
	return c.components, nil
}

// AllAsMapOrDefault returns all components as map or the provided default.
func (c *Collection) AllAsMapOrDefault(defaultComponents Map) Map {
	components, err := c.AllAsMap()
	if err != nil {
		return defaultComponents
	}
	return components
}

// AllAsMapOrNil returns all components as map or nil in case of error.
func (c *Collection) AllAsMapOrNil() Map {
	return c.AllAsMapOrDefault(nil)
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

// NoneMatch returns true if no components match the predicate.
func (c *Collection) NoneMatch(predicate Predicate) bool {
	return !c.AnyMatch(predicate)
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

// AnyThatMatches returns any arbitrary component that matches the predicate.
// Note: Map iteration order is not guaranteed, so this may return different items on each call.
func (c *Collection) AnyThatMatches(predicate Predicate) *Component {
	if c.HasChainableErr() {
		return New("").WithChainableErr(c.ChainableErr())
	}
	for _, comp := range c.components {
		if predicate(comp) {
			return comp
		}
	}
	c.WithChainableErr(ErrNoComponentMatchesPredicate)
	return New("").WithChainableErr(c.ChainableErr())
}

// Filter returns a new collection with components that match the predicate.
func (c *Collection) Filter(predicate Predicate) *Collection {
	if c.HasChainableErr() {
		return NewCollection().WithChainableErr(c.ChainableErr())
	}
	filtered := NewCollection()
	for _, comp := range c.components {
		if predicate(comp) {
			filtered = filtered.With(comp)
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
			mapped = mapped.With(transformedComp)
			if mapped.HasChainableErr() {
				return mapped
			}
		}
	}
	return mapped
}
