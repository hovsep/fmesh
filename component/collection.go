package component

import (
	"fmt"
	"maps"
	"slices"
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

// Len returns the number of components in the collection.
func (c *Collection) Len() int {
	return len(c.components)
}

// IsEmpty returns true when there are no components in the collection.
func (c *Collection) IsEmpty() bool {
	return c.Len() == 0
}

// AllOrdered returns all components sorted by name.
// Use this when iteration order must be deterministic (e.g. the drain phase).
func (c *Collection) AllOrdered() []*Component {
	names := slices.Sorted(maps.Keys(c.components))
	ordered := make([]*Component, len(names))
	for i, name := range names {
		ordered[i] = c.components[name]
	}
	return ordered
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
