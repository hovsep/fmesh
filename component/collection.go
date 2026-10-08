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
	meta *meta.Meta
}

// NewCollection creates an empty collection.
func NewCollection() *Collection {
	return &Collection{
		keyedComponents: collection.NewKeyed[*Component]("component"),
		meta:            meta.New(),
	}
}

// Meta returns the collection's own metadata store.
func (c *Collection) Meta() *meta.Meta { return c.meta }

// Remove deletes components by name and returns the collection.
func (c *Collection) Remove(names ...string) *Collection {
	c.keyedComponents.Remove(names...)
	return c
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
