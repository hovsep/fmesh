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

// newGroup creates an empty group.
func newGroup() *Group {
	return &Group{
		meta: meta.New(),
	}
}

// newInputGroup creates a group of input ports with the given names.
func newInputGroup(names ...string) *Group {
	return newGroupOfDirection(DirectionIn, names...)
}

// newOutputGroup creates a group of output ports with the given names.
func newOutputGroup(names ...string) *Group {
	return newGroupOfDirection(DirectionOut, names...)
}

func newGroupOfDirection(direction Direction, names ...string) *Group {
	ports := make([]*Port, len(names))
	for i, name := range names {
		ports[i], _ = newPort(direction, name) // no opts, never fails
	}
	return newGroup().setPorts(ports)
}

func (g *Group) raw() []*Port { return collection.Items(&g.portSlice) }

// Meta returns the group's own metadata store.
func (g *Group) Meta() *meta.Meta { return g.meta }

// add appends ports to the group in place. Internal use only; always succeeds.
func (g *Group) add(ports ...*Port) {
	collection.AppendItems(&g.portSlice, ports...)
}

func (g *Group) setPorts(ports []*Port) *Group {
	collection.SetItems(&g.portSlice, ports)
	return g
}

// Filter returns a new group with ports that match the predicate.
func (g *Group) Filter(predicate Predicate) *Group {
	filtered := newGroup()
	for _, port := range g.raw() {
		if predicate(port) {
			filtered.add(port)
		}
	}
	return filtered
}
