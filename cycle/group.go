package cycle

import "github.com/hovsep/fmesh/common"

// Cycles contains the results of several activation cycles
type Cycles []*Cycle

// Group contains multiple activation cycles
type Group struct {
	*common.Chainable
	cycles Cycles
}

// NewGroup creates a group of cycles
func NewGroup() *Group {
	newGroup := &Group{
		Chainable: common.NewChainable(),
	}
	cycles := make(Cycles, 0)
	return newGroup.withCycles(cycles)
}

// With adds cycle results to existing collection
func (g *Group) With(cycles ...*Cycle) *Group {
	newCycles := make(Cycles, len(g.cycles)+len(cycles))
	copy(newCycles, g.cycles)
	for i, c := range cycles {
		newCycles[len(g.cycles)+i] = c
	}
	return g.withCycles(newCycles)
}

// withSignals sets signals
func (g *Group) withCycles(cycles Cycles) *Group {
	g.cycles = cycles
	return g
}

// Cycles getter
func (g *Group) Cycles() (Cycles, error) {
	if g.HasErr() {
		return nil, g.Err()
	}
	return g.cycles, nil
}

// CyclesOrNil returns signals or nil in case of any error
func (g *Group) CyclesOrNil() Cycles {
	return g.CyclesOrDefault(nil)
}

// CyclesOrDefault returns signals or default in case of any error
func (g *Group) CyclesOrDefault(defaultCycles Cycles) Cycles {
	signals, err := g.Cycles()
	if err != nil {
		return defaultCycles
	}
	return signals
}

// Len returns number of cycles in group
func (g *Group) Len() int {
	return len(g.cycles)
}

// Last returns the latest cycle added to the group
func (g *Group) Last() *Cycle {
	if g.Len() == 0 {
		return New().WithErr(errNoCyclesInGroup)
	}

	return g.cycles[g.Len()-1]
}
