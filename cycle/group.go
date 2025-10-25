package cycle

// Cycles contain the results of several activation cycles.
type Cycles []*Cycle

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

// With adds cycle results to existing collection.
func (g *Group) With(cycles ...*Cycle) *Group {
	newCycles := make(Cycles, len(g.cycles)+len(cycles))
	copy(newCycles, g.cycles)
	for i, c := range cycles {
		newCycles[len(g.cycles)+i] = c
	}
	return g.withCycles(newCycles)
}

// withSignals sets signals.
func (g *Group) withCycles(cycles Cycles) *Group {
	g.cycles = cycles
	return g
}

// Cycles getter.
func (g *Group) Cycles() (Cycles, error) {
	if g.HasChainableErr() {
		return nil, g.ChainableErr()
	}
	return g.cycles, nil
}

// CyclesOrNil returns signals or nil in case of any error.
func (g *Group) CyclesOrNil() Cycles {
	return g.CyclesOrDefault(nil)
}

// CyclesOrDefault returns signals or default in case of any error.
func (g *Group) CyclesOrDefault(defaultCycles Cycles) Cycles {
	signals, err := g.Cycles()
	if err != nil {
		return defaultCycles
	}
	return signals
}

// Len returns number of cycles in group.
func (g *Group) Len() int {
	return len(g.cycles)
}

// Last returns the latest cycle added to the group.
func (g *Group) Last() *Cycle {
	if g.Len() == 0 {
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
