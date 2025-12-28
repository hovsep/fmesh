package cycle

// Group contains multiple activation cycles.
type Group struct {
	chainableErr error
	cycles       Cycles
}

// NewGroup creates a group of cycles.
func NewGroup() *Group {
	newGroup := &Group{
		chainableErr: nil,
	}
	cycles := make(Cycles, 0)
	return newGroup.withCycles(cycles)
}

// Add adds cycles to the group and returns it.
func (g *Group) Add(cycles ...*Cycle) *Group {
	if g.HasChainableErr() {
		return g
	}
	newCycles := make(Cycles, len(g.cycles)+len(cycles))
	copy(newCycles, g.cycles)
	for i, c := range cycles {
		newCycles[len(g.cycles)+i] = c
	}
	return g.withCycles(newCycles)
}

// Without removes cycles matching the predicate and returns a new group.
func (g *Group) Without(predicate Predicate) *Group {
	if g.HasChainableErr() {
		return NewGroup().WithChainableErr(g.ChainableErr())
	}
	// Keep cycles that DON'T match the predicate
	return g.Filter(func(c *Cycle) bool {
		return !predicate(c)
	})
}

// ForEach applies the action to each cycle and returns the group for chaining.
func (g *Group) ForEach(action func(*Cycle) error) *Group {
	if g.HasChainableErr() {
		return g
	}
	for _, c := range g.cycles {
		if err := action(c); err != nil {
			g.chainableErr = err
			return g
		}
	}
	return g
}

// withCycles sets cycles.
func (g *Group) withCycles(cycles Cycles) *Group {
	g.cycles = cycles
	return g
}

// Len returns number of cycles in group.
// Returns 0 if the group has a chainable error.
func (g *Group) Len() int {
	if g.HasChainableErr() {
		return 0
	}
	return len(g.cycles)
}

// Last returns the most recent cycle added to the group.
// Returns nil if the group is empty or has an error.
func (g *Group) Last() *Cycle {
	if g.HasChainableErr() {
		return nil
	}
	if g.IsEmpty() {
		return nil
	}

	return g.cycles[g.Len()-1]
}

// WithChainableErr sets a chainable error and returns the group.
func (g *Group) WithChainableErr(err error) *Group {
	g.chainableErr = err
	return g
}

// HasChainableErr returns true when a chainable error is set.
func (g *Group) HasChainableErr() bool {
	return g.chainableErr != nil
}

// ChainableErr returns the chainable error.
func (g *Group) ChainableErr() error {
	return g.chainableErr
}

// IsEmpty returns true when there are no cycles in the group.
func (g *Group) IsEmpty() bool {
	return g.Len() == 0
}

// First returns the first cycle in the group.
// Returns nil if the group is empty or has an error.
func (g *Group) First() *Cycle {
	if g.HasChainableErr() {
		return nil
	}
	if g.IsEmpty() {
		return nil
	}
	return g.cycles[0]
}

// All returns all cycles as a slice.
func (g *Group) All() (Cycles, error) {
	if g.HasChainableErr() {
		return nil, g.ChainableErr()
	}
	return g.cycles, nil
}

// AllMatch returns true if all cycles match the predicate.
func (g *Group) AllMatch(predicate Predicate) bool {
	if g.HasChainableErr() {
		return false
	}
	for _, cyc := range g.cycles {
		if !predicate(cyc) {
			return false
		}
	}
	return true
}

// AnyMatch returns true if any cycle matches the predicate.
func (g *Group) AnyMatch(predicate Predicate) bool {
	if g.HasChainableErr() {
		return false
	}
	for _, cyc := range g.cycles {
		if predicate(cyc) {
			return true
		}
	}
	return false
}

// CountMatch returns the number of cycles that match the predicate.
func (g *Group) CountMatch(predicate Predicate) int {
	if g.HasChainableErr() {
		return 0
	}
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
	if g.HasChainableErr() {
		return NewGroup().WithChainableErr(g.ChainableErr())
	}
	filtered := NewGroup()
	for _, cyc := range g.cycles {
		if predicate(cyc) {
			filtered = filtered.Add(cyc)
			if filtered.HasChainableErr() {
				return filtered
			}
		}
	}
	return filtered
}

// Map returns a new group with cycles transformed by the mapper function.
func (g *Group) Map(mapper Mapper) *Group {
	if g.HasChainableErr() {
		return NewGroup().WithChainableErr(g.ChainableErr())
	}
	mapped := NewGroup()
	for _, cyc := range g.cycles {
		transformedCyc := mapper(cyc)
		if transformedCyc != nil {
			mapped = mapped.Add(transformedCyc)
			if mapped.HasChainableErr() {
				return mapped
			}
		}
	}
	return mapped
}
