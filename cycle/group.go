package cycle

import "slices"

// Group contains multiple activation cycles.
type Group struct {
	cycles []*Cycle
}

// NewGroup creates a group of cycles.
func NewGroup() *Group {
	return &Group{
		cycles: make([]*Cycle, 0),
	}
}

// Add adds cycles to the group and returns it.
// Note: Unlike other collections, cycle errors are NOT propagated to the group
// because cycles represent historical execution records - users need to access
// cycles that had errors to understand what happened.
func (g *Group) Add(cycles ...*Cycle) *Group {
	newCycles := make([]*Cycle, len(g.cycles)+len(cycles))
	copy(newCycles, g.cycles)
	for i, c := range cycles {
		newCycles[len(g.cycles)+i] = c
	}
	g.cycles = newCycles
	return g
}

// Without removes cycles matching the predicate and returns a new group.
func (g *Group) Without(predicate Predicate) *Group {
	return g.Filter(func(c *Cycle) bool {
		return !predicate(c)
	})
}

// ForEach applies the action to each cycle. Returns the first error encountered.
func (g *Group) ForEach(action func(*Cycle) error) error {
	for _, c := range g.cycles {
		if err := action(c); err != nil {
			return err
		}
	}
	return nil
}

// ForEachIf applies the action only to cycles that match the predicate.
func (g *Group) ForEachIf(predicate Predicate, action func(*Cycle) error) error {
	for _, c := range g.cycles {
		if predicate(c) {
			if err := action(c); err != nil {
				return err
			}
		}
	}
	return nil
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

// Find returns the first cycle matching the predicate, or nil if none match.
func (g *Group) Find(predicate Predicate) *Cycle {
	for _, c := range g.cycles {
		if predicate(c) {
			return c
		}
	}
	return nil
}

// First returns the first cycle in the group.
// Returns nil if the group is empty.
func (g *Group) First() *Cycle {
	if g.IsEmpty() {
		return nil
	}
	return g.cycles[0]
}

// All returns all cycles as a slice.
func (g *Group) All() ([]*Cycle, error) {
	return g.cycles, nil
}

// Every returns true if all cycles match the predicate.
func (g *Group) Every(predicate Predicate) bool {
	for _, cyc := range g.cycles {
		if !predicate(cyc) {
			return false
		}
	}
	return true
}

// Any returns true if any cycle matches the predicate.
func (g *Group) Any(predicate Predicate) bool {
	return slices.ContainsFunc(g.cycles, predicate)
}

// Count returns the number of cycles that match the predicate.
func (g *Group) Count(predicate Predicate) int {
	count := 0
	for _, cyc := range g.cycles {
		if predicate(cyc) {
			count++
		}
	}
	return count
}

// Filter returns a new group with cycles that match the predicate.
func (g *Group) Filter(predicate Predicate) *Group {
	filtered := NewGroup()
	for _, cyc := range g.cycles {
		if predicate(cyc) {
			filtered = filtered.Add(cyc)
		}
	}
	return filtered
}

// MapIf is like Map but only applies the mapper to cycles that match the predicate.
// Non-matching cycles are kept as-is. Nil mapper results are dropped.
func (g *Group) MapIf(predicate Predicate, mapper Mapper) *Group {
	mapped := NewGroup()
	for _, c := range g.cycles {
		if predicate(c) {
			if transformedCyc := mapper(c); transformedCyc != nil {
				mapped = mapped.Add(transformedCyc)
			}
		} else {
			mapped = mapped.Add(c)
		}
	}
	return mapped
}

// Map returns a new group with cycles transformed by the mapper function.
func (g *Group) Map(mapper Mapper) *Group {
	mapped := NewGroup()
	for _, cyc := range g.cycles {
		if transformedCyc := mapper(cyc); transformedCyc != nil {
			mapped = mapped.Add(transformedCyc)
		}
	}
	return mapped
}
