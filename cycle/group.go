package cycle

import (
	"github.com/hovsep/fmesh/internal/collection"
	"github.com/hovsep/fmesh/meta"
)

// cycleSlice hides the embedded field name so it cannot be reached or
// reassigned from outside; only the base's exported read methods promote.
type cycleSlice = collection.Slice[*Cycle]

// Group contains multiple activation cycles.
type Group struct {
	cycleSlice
	// lenLimit caps how many cycles the group retains; 0 means unlimited.
	// When Add would exceed the limit, the oldest cycles are evicted.
	lenLimit int
	meta     *meta.Meta
}

// NewGroup creates a group of cycles.
func NewGroup() *Group {
	return &Group{
		meta: meta.New(),
	}
}

func (g *Group) raw() []*Cycle { return collection.Items(&g.cycleSlice) }

func (g *Group) replace(items []*Cycle) { collection.SetItems(&g.cycleSlice, items) }

// Meta returns the group's own metadata store.
func (g *Group) Meta() *meta.Meta { return g.meta }

// SetLenLimit caps how many cycles the group retains and returns the group.
// A limit of 0 (or negative) means unlimited. When an Add pushes the group past
// the limit, the oldest cycles are evicted; the limit is also applied
// immediately if the group already exceeds it.
func (g *Group) SetLenLimit(limit int) *Group {
	if limit < 0 {
		limit = 0
	}
	g.lenLimit = limit
	g.evictExcess()
	return g
}

// Add adds cycles to the group and returns it. When a length limit is set
// (SetLenLimit), adding beyond it evicts the oldest cycles.
func (g *Group) Add(cycles ...*Cycle) *Group {
	collection.AppendItems(&g.cycleSlice, cycles...)
	g.evictExcess()
	return g
}

// evictExcess drops the oldest cycles beyond the length limit. The front is
// sliced off rather than the rest shifted down, so evicting one cycle per Add is
// O(1) amortized; evicted slots are cleared so the cycles can be collected.
func (g *Group) evictExcess() {
	excess := g.Len() - g.lenLimit
	if g.lenLimit == 0 || excess <= 0 {
		return
	}
	cycles := g.raw()
	clear(cycles[:excess])
	g.replace(cycles[excess:])
}

// derive returns an empty group carrying the receiver's length limit and copies
// of its own metadata, so Filter/Map/MapIf results keep them.
func (g *Group) derive() *Group {
	derived := NewGroup()
	derived.lenLimit = g.lenLimit
	derived.meta = g.meta.Clone()
	return derived
}

// Without removes cycles matching the predicate and returns a new group.
func (g *Group) Without(predicate Predicate) *Group {
	return g.Filter(func(c *Cycle) bool {
		return !predicate(c)
	})
}

// Filter returns a new group with cycles that match the predicate. The group's
// own length limit and metadata are preserved on the returned group.
func (g *Group) Filter(predicate Predicate) *Group {
	filtered := g.derive()
	for _, cyc := range g.raw() {
		if predicate(cyc) {
			filtered.Add(cyc)
		}
	}
	return filtered
}

// MapIf is like Map but only applies the mapper to cycles that match the predicate.
// Non-matching cycles are kept as-is. Nil mapper results are dropped. The group's
// own length limit and metadata are preserved on the returned group.
func (g *Group) MapIf(predicate Predicate, mapper Mapper) *Group {
	mapped := g.derive()
	for _, c := range g.raw() {
		if predicate(c) {
			if transformedCyc := mapper(c); transformedCyc != nil {
				mapped.Add(transformedCyc)
			}
		} else {
			mapped.Add(c)
		}
	}
	return mapped
}

// Map returns a new group with cycles transformed by the mapper function. The
// group's own length limit and metadata are preserved on the returned group.
func (g *Group) Map(mapper Mapper) *Group {
	mapped := g.derive()
	for _, cyc := range g.raw() {
		if transformedCyc := mapper(cyc); transformedCyc != nil {
			mapped.Add(transformedCyc)
		}
	}
	return mapped
}
