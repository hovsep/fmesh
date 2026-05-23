package component

import (
	"fmt"
)

// Collection is a collection of components with useful methods.
type Collection struct {
	components map[string]*Component
}

// NewCollection creates an empty collection.
func NewCollection() *Collection {
	return &Collection{
		components: make(map[string]*Component),
	}
}

// ByName returns a component by its name.
// Returns nil if not found.
func (c *Collection) ByName(name string) *Component {
	return c.components[name]
}

// Add adds components and returns an error if a duplicate name is found.
func (c *Collection) Add(components ...*Component) error {
	for _, comp := range components {
		if _, exists := c.components[comp.Name()]; exists {
			return fmt.Errorf("component with name %q already exists", comp.Name())
		}
		c.components[comp.Name()] = comp
	}
	return nil
}

// Without removes components by name and returns the collection.
func (c *Collection) Without(names ...string) *Collection {
	for _, name := range names {
		delete(c.components, name)
	}
	return c
}

// Len returns the number of components in the collection.
func (c *Collection) Len() int {
	return len(c.components)
}

// IsEmpty returns true when there are no components in the collection.
func (c *Collection) IsEmpty() bool {
	return c.Len() == 0
}

// Any returns any arbitrary component from the collection.
// Returns nil if the collection is empty.
// Note: Map iteration order is not guaranteed, so this may return different items on each call.
func (c *Collection) Any() *Component {
	for _, comp := range c.components {
		return comp
	}
	return nil
}

// All returns all components as a map.
func (c *Collection) All() (map[string]*Component, error) {
	return c.components, nil
}

// Every returns true if all components match the predicate.
func (c *Collection) Every(predicate Predicate) bool {
	for _, comp := range c.components {
		if !predicate(comp) {
			return false
		}
	}
	return true
}

// AnyMatch returns true if any component matches the predicate.
// Note: AnyMatch is used instead of Any to avoid conflict with the no-arg Any() *Component method.
func (c *Collection) AnyMatch(predicate Predicate) bool {
	for _, comp := range c.components {
		if predicate(comp) {
			return true
		}
	}
	return false
}

// Count returns the number of components that match the predicate.
func (c *Collection) Count(predicate Predicate) int {
	count := 0
	for _, comp := range c.components {
		if predicate(comp) {
			count++
		}
	}
	return count
}

// FindAny returns any arbitrary component that matches the predicate.
// Returns nil if no match found.
// Note: Map iteration order is not guaranteed, so this may return different items on each call.
func (c *Collection) FindAny(predicate Predicate) *Component {
	for _, comp := range c.components {
		if predicate(comp) {
			return comp
		}
	}
	return nil
}

// Filter returns a new collection with components that match the predicate.
func (c *Collection) Filter(predicate Predicate) *Collection {
	filtered := NewCollection()
	for _, comp := range c.components {
		if predicate(comp) {
			filtered.components[comp.Name()] = comp
		}
	}
	return filtered
}

// Map returns a new collection with components transformed by the mapper function.
// Returns an error if a mapped component has a duplicate name.
func (c *Collection) Map(mapper Mapper) (*Collection, error) {
	mapped := NewCollection()
	for _, comp := range c.components {
		transformedComp := mapper(comp)
		if transformedComp != nil {
			if err := mapped.Add(transformedComp); err != nil {
				return nil, err
			}
		}
	}
	return mapped, nil
}

// ForEach applies the action to each component. Returns the first error encountered.
func (c *Collection) ForEach(action func(*Component) error) error {
	for _, comp := range c.components {
		if err := action(comp); err != nil {
			return err
		}
	}
	return nil
}

// Clear removes all components from the collection.
func (c *Collection) Clear() *Collection {
	c.components = make(map[string]*Component)
	return c
}
