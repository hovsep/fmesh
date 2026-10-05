package port

import (
	"github.com/hovsep/fmesh/internal/collection"
	"github.com/hovsep/fmesh/meta"
)

// portSlice hides the embedded field name so it cannot be reached or
// reassigned from outside; only the base's exported read methods promote.
type portSlice = collection.Slice[*Port]

// Group represents a list of ports.
// It can carry multiple ports with the same name and has no lookup methods.
type Group struct {
	portSlice
	meta *meta.Meta
}

// NewGroup creates an empty group.
func NewGroup() *Group {
	return &Group{
		meta: meta.New(),
	}
}

// NewInputGroup creates a group of input ports with the given names.
func NewInputGroup(names ...string) *Group {
	return newGroupOfDirection(DirectionIn, names...)
}

// NewOutputGroup creates a group of output ports with the given names.
func NewOutputGroup(names ...string) *Group {
	return newGroupOfDirection(DirectionOut, names...)
}

func newGroupOfDirection(direction Direction, names ...string) *Group {
	ports := make([]*Port, len(names))
	for i, name := range names {
		ports[i], _ = newPort(direction, name) // no opts, never fails
	}
	return NewGroup().setPorts(ports)
}

func (g *Group) raw() []*Port { return collection.Items(&g.portSlice) }

// Meta returns the group's own metadata store.
func (g *Group) Meta() *meta.Meta { return g.meta }

// add appends ports to the group in place. Internal use only; always succeeds.
func (g *Group) add(ports ...*Port) {
	collection.AppendItems(&g.portSlice, ports...)
}

// Without removes ports matching the predicate and returns a new group.
func (g *Group) Without(predicate Predicate) *Group {
	return g.Filter(func(p *Port) bool {
		return !predicate(p)
	})
}

func (g *Group) setPorts(ports []*Port) *Group {
	collection.SetItems(&g.portSlice, ports)
	return g
}

// Filter returns a new group with ports that match the predicate.
func (g *Group) Filter(predicate Predicate) *Group {
	filtered := NewGroup()
	for _, port := range g.raw() {
		if predicate(port) {
			filtered.add(port)
		}
	}
	return filtered
}

// MapIf is like Map but only applies the mapper to ports that match the predicate.
// Non-matching ports are kept as-is. Nil mapper results are dropped.
func (g *Group) MapIf(predicate Predicate, mapper Mapper) *Group {
	mapped := NewGroup()
	for _, p := range g.raw() {
		if predicate(p) {
			if result := mapper(p); result != nil {
				mapped.add(result)
			}
		} else {
			mapped.add(p)
		}
	}
	return mapped
}

// Map returns a new group with ports transformed by the mapper function.
// Nil mapper results are dropped.
func (g *Group) Map(mapper Mapper) *Group {
	mapped := NewGroup()
	for _, port := range g.raw() {
		if result := mapper(port); result != nil {
			mapped.add(result)
		}
	}
	return mapped
}
