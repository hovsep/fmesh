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

// With adds cycles to the group and returns it.
func (g *Group) With(cycles ...*Cycle) *Group {
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
func (g *Group) ForEach(action func(*Cycle)) *Group {
	if g.HasChainableErr() {
		return g
	}
	for _, c := range g.cycles {
		action(c)
	}
	return g
}

// withCycles sets cycles.
func (g *Group) withCycles(cycles Cycles) *Group {
	g.cycles = cycles
	return g
}

// Len returns number of cycles in group.
func (g *Group) Len() int {
	return len(g.cycles)
}

// Last returns the most recent cycle added to the group.
func (g *Group) Last() *Cycle {
	if g.IsEmpty() {
		return New().WithChainableErr(errNoCyclesInGroup)
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

// ChainableErr returns chainable error.
func (g *Group) ChainableErr() error {
	return g.chainableErr
}

// IsEmpty returns true when there are no cycles in the group.
func (g *Group) IsEmpty() bool {
	return g.Len() == 0
}

// First returns the first cycle in the group.
func (g *Group) First() *Cycle {
	if g.HasChainableErr() {
		return New().WithChainableErr(g.ChainableErr())
	}
	if g.IsEmpty() {
		g.WithChainableErr(errNoCyclesInGroup)
		return New().WithChainableErr(g.ChainableErr())
	}
	return g.cycles[0]
}

// FirstOrDefault returns the first cycle or the provided default.
func (g *Group) FirstOrDefault(defaultCycle *Cycle) *Cycle {
	if g.HasChainableErr() || g.IsEmpty() {
		return defaultCycle
	}
	return g.cycles[0]
}

// FirstOrNil returns the first cycle or nil.
func (g *Group) FirstOrNil() *Cycle {
	if g.HasChainableErr() || g.IsEmpty() {
		return nil
	}
	return g.cycles[0]
}

// AllAsSlice returns all cycles as Cycles wrapper type.
func (g *Group) AllAsSlice() (Cycles, error) {
	if g.HasChainableErr() {
		return nil, g.ChainableErr()
	}
	return g.cycles, nil
}

// AllAsSliceOrDefault returns all cycles as Cycles wrapper or the provided default.
func (g *Group) AllAsSliceOrDefault(defaultCycles Cycles) Cycles {
	cycles, err := g.AllAsSlice()
	if err != nil {
		return defaultCycles
	}
	return cycles
}

// AllAsSliceOrNil returns all cycles as Cycles wrapper or nil in case of error.
func (g *Group) AllAsSliceOrNil() Cycles {
	return g.AllAsSliceOrDefault(nil)
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

// NoneMatch returns true if no cycles match the predicate.
func (g *Group) NoneMatch(predicate Predicate) bool {
	return !g.AnyMatch(predicate)
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

// FirstMatch returns the first cycle that matches the predicate.
func (g *Group) FirstMatch(predicate Predicate) *Cycle {
	if g.HasChainableErr() {
		return New().WithChainableErr(g.ChainableErr())
	}
	for _, cyc := range g.cycles {
		if predicate(cyc) {
			return cyc
		}
	}
	g.WithChainableErr(errNoCyclesInGroup)
	return New().WithChainableErr(g.ChainableErr())
}

// Filter returns a new group with cycles that match the predicate.
func (g *Group) Filter(predicate Predicate) *Group {
	if g.HasChainableErr() {
		return NewGroup().WithChainableErr(g.ChainableErr())
	}
	filtered := NewGroup()
	for _, cyc := range g.cycles {
		if predicate(cyc) {
			filtered = filtered.With(cyc)
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
			mapped = mapped.With(transformedCyc)
			if mapped.HasChainableErr() {
				return mapped
			}
		}
	}
	return mapped
}
