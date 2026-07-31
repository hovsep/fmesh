package cycle

import (
	"slices"
)

// Group contains multiple activation cycles.
type Group struct {
	cycles []*Cycle
	// lenLimit caps how many cycles the group retains; 0 means unlimited.
	// When Add would exceed the limit, the oldest cycles are evicted.
	lenLimit int
}

// NewGroup creates a group of cycles.
func NewGroup() *Group {
	return &Group{
		cycles: make([]*Cycle, 0),
	}
}

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
// Note: Unlike other collections, cycle errors are NOT propagated to the group
// because cycles represent historical execution records - users need to access
// cycles that had errors to understand what happened.
func (g *Group) Add(cycles ...*Cycle) *Group {
	g.cycles = append(g.cycles, cycles...)
	g.evictExcess()
	return g
}

// evictExcess removes the oldest cycles beyond the configured length limit.
func (g *Group) evictExcess() {
	if g.lenLimit > 0 && g.Len() > g.lenLimit {
		g.RemoveOldest(g.Len() - g.lenLimit)
	}
}

// RemoveOldest removes the count oldest cycles (from the front of the group) in place and
// returns the group. count is clamped to [0, Len()]. Evicted cycles are cleared from the
// backing array so they become unreachable and can be garbage-collected.
func (g *Group) RemoveOldest(count int) *Group {
	if count <= 0 {
		return g
	}
	if count > g.Len() {
		count = g.Len()
	}

	copy(g.cycles, g.cycles[count:])
	clear(g.cycles[g.Len()-count:])
	g.cycles = g.cycles[:g.Len()-count]

	return g
}

// Len returns number of cycles in group.
func (g *Group) Len() int {
	return len(g.cycles)
}

// Last returns the most recent cycle added to the group.
// Returns nil if the group is empty.
func (g *Group) Last() *Cycle {
	if g.IsEmpty() {
		return nil
	}
	return g.cycles[g.Len()-1]
}

// IsEmpty returns true when there are no cycles in the group.
func (g *Group) IsEmpty() bool {
	return g.Len() == 0
}

// First returns the first cycle in the group.
// Returns nil if the group is empty.
func (g *Group) First() *Cycle {
	if g.IsEmpty() {
		return nil
	}
	return g.cycles[0]
}

// All returns a cloned slice of cycles. The slice is independent of the group;
// the *Cycle pointers inside are shared.
func (g *Group) All() []*Cycle {
	return slices.Clone(g.cycles)
}
