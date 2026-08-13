package port

import (
	"slices"

	"github.com/hovsep/fmesh/meta"
)

// Group represents a list of ports.
// It can carry multiple ports with the same name and has no lookup methods.
type Group struct {
	ports   []*Port
	labels  *meta.Labels
	scalars *meta.Scalars
}

// NewGroup creates an empty group.
func NewGroup() *Group {
	return &Group{
		labels:  meta.NewLabels(),
		scalars: meta.NewScalars(),
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
		ports[i] = newPortOfDirection(direction, name)
	}
	return NewGroup().setPorts(ports)
}

func newPortOfDirection(direction Direction, name string) *Port {
	p, _ := newPort(direction, name) // no opts, never fails
	return p
}

// Labels returns the group's own labels store.
func (g *Group) Labels() *meta.Labels { return g.labels }

// SetLabel adds or updates a single label on the group itself.
func (g *Group) SetLabel(name, value string) *Group { g.labels.Set(name, value); return g }

// Scalars returns the group's own scalars store.
func (g *Group) Scalars() *meta.Scalars { return g.scalars }

// SetScalar adds or updates a single scalar on the group itself.
func (g *Group) SetScalar(name string, value float64) *Group {
	g.scalars.Set(name, value)
	return g
}

// add appends ports to the group in place. Internal use only; always succeeds.
func (g *Group) add(ports ...*Port) {
	g.ports = append(g.ports, ports...)
}

// Without removes ports matching the predicate and returns a new group.
func (g *Group) Without(predicate Predicate) *Group {
	return g.Filter(func(p *Port) bool {
		return !predicate(p)
	})
}

// ForEach applies the action to each port. Returns the first error encountered.
func (g *Group) ForEach(action func(*Port) error) error {
	for _, p := range g.ports {
		if err := action(p); err != nil {
			return err
		}
	}
	return nil
}

// ForEachIf applies the action only to ports that match the predicate.
func (g *Group) ForEachIf(predicate Predicate, action func(*Port) error) error {
	for _, p := range g.ports {
		if predicate(p) {
			if err := action(p); err != nil {
				return err
			}
		}
	}
	return nil
}

func (g *Group) setPorts(ports []*Port) *Group {
	g.ports = ports
	return g
}

// All returns a cloned slice of ports. The slice is independent of the group;
// the *Port pointers inside are shared.
func (g *Group) All() []*Port {
	return slices.Clone(g.ports)
}

// Len returns the number of ports in a group.
func (g *Group) Len() int {
	return len(g.ports)
}

// IsEmpty returns true when there are no ports in the group.
func (g *Group) IsEmpty() bool {
	return g.Len() == 0
}

// Find returns the first port matching the predicate, or nil if none match.
func (g *Group) Find(predicate Predicate) *Port {
	for _, p := range g.ports {
		if predicate(p) {
			return p
		}
	}
	return nil
}

// First returns the first port in the group, or nil if empty.
func (g *Group) First() *Port {
	if g.IsEmpty() {
		return nil
	}
	return g.ports[0]
}

// Every returns true if all ports match the predicate.
func (g *Group) Every(predicate Predicate) bool {
	for _, port := range g.ports {
		if !predicate(port) {
			return false
		}
	}
	return true
}

// Any returns true if any port matches the predicate.
func (g *Group) Any(predicate Predicate) bool {
	return slices.ContainsFunc(g.ports, predicate)
}

// Count returns the number of ports that match the predicate.
func (g *Group) Count(predicate Predicate) int {
	count := 0
	for _, port := range g.ports {
		if predicate(port) {
			count++
		}
	}
	return count
}

// Filter returns a new group with ports that match the predicate.
func (g *Group) Filter(predicate Predicate) *Group {
	filtered := NewGroup()
	for _, port := range g.ports {
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
	for _, p := range g.ports {
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
	for _, port := range g.ports {
		if result := mapper(port); result != nil {
			mapped.add(result)
		}
	}
	return mapped
}
