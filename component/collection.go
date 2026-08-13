package component

import (
	"github.com/hovsep/fmesh/internal/collection"
	"github.com/hovsep/fmesh/meta"
)

// keyedComponents hides the embedded field name so it cannot be reached or
// reassigned from outside; only the base's exported methods promote.
type keyedComponents = collection.Keyed[*Component]

// Collection is a collection of components with useful methods.
// Every traversal goes in component-name order, matching port.Collection —
// see .agent/docs/design.md for the ordering guarantees this upholds.
type Collection struct {
	*keyedComponents
	labels  *meta.Labels
	scalars *meta.Scalars
}

// NewCollection creates an empty collection.
func NewCollection() *Collection {
	return &Collection{
		keyedComponents: collection.NewKeyed[*Component]("component"),
		labels:          meta.NewLabels(),
		scalars:         meta.NewScalars(),
	}
}

// Labels returns the collection's own labels store.
func (c *Collection) Labels() *meta.Labels { return c.labels }

// Scalars returns the collection's own scalars store.
func (c *Collection) Scalars() *meta.Scalars { return c.scalars }

// Remove deletes components by name and returns the collection.
func (c *Collection) Remove(names ...string) *Collection {
	c.keyedComponents.Remove(names...)
	return c
}

// Any returns the first component in the collection by name order.
// Returns nil if the collection is empty. Stable across runs.
func (c *Collection) Any() *Component {
	for comp := range c.Each {
		return comp
	}
	return nil
}

// Filter returns a new collection with components that match the predicate.
func (c *Collection) Filter(predicate Predicate) *Collection {
	matched := make([]*Component, 0, c.Len())
	for comp := range c.Each {
		if predicate(comp) {
			matched = append(matched, comp)
		}
	}
	filtered := NewCollection()
	_ = filtered.Add(matched...) // components come from existing collection — names are unique by construction
	return filtered
}

// Map returns a new collection with components transformed by the mapper function.
// Returns an error if a mapped component has a duplicate name.
func (c *Collection) Map(mapper Mapper) (*Collection, error) {
	transformed := make([]*Component, 0, c.Len())
	for comp := range c.Each {
		if transformedComp := mapper(comp); transformedComp != nil {
			transformed = append(transformed, transformedComp)
		}
	}
	mapped := NewCollection()
	if err := mapped.Add(transformed...); err != nil {
		return nil, err
	}
	return mapped, nil
}

// Clear removes all components from the collection.
func (c *Collection) Clear() *Collection {
	collection.Reset(c.keyedComponents)
	return c
}
