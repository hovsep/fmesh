package fmesh

import (
	"fmt"
	"strings"

	"github.com/hovsep/fmesh/component"
	"github.com/hovsep/fmesh/cycle"
	"github.com/hovsep/fmesh/port"
)

// livelockDetector is the per-run bookkeeping that tells a mesh making progress
// from one repeating the same cycle. A threshold of 0 disables it.
type livelockDetector struct {
	components     *component.Collection
	threshold      int
	stalledCycles  int
	pendingSignals int
}

// newLivelockDetector takes the seeded inputs as the baseline the first cycle is
// compared against, so a mesh that stalls immediately is caught in threshold
// cycles rather than threshold+1.
func newLivelockDetector(components *component.Collection, threshold int) livelockDetector {
	d := livelockDetector{components: components, threshold: threshold}
	d.takeBaseline()
	return d
}

// takeBaseline records the pending signal count the next check compares against.
// Called after the drain, so the signals a cycle delivered are counted in full.
func (d *livelockDetector) takeBaseline() {
	if d.threshold > 0 {
		d.pendingSignals = d.countPendingSignals()
	}
}

// countPendingSignals totals the signals sitting on input ports across the mesh.
// It is the cheapest thing that answers "did anything move?" — see detect.
func (d *livelockDetector) countPendingSignals() int {
	total := 0
	_ = d.components.ForEach(func(c *component.Component) error {
		return c.Inputs().ForEach(func(p *port.Port) error {
			total += p.Signals().Len()
			return nil
		})
	})
	return total
}

// detect reports whether the mesh has stopped making progress, and updates the
// stall count. Called once per cycle.
//
// A cycle is stalled when every component that activated was waiting and keeping
// its inputs *and* the pending signal count is what the previous drain left —
// both halves are needed: the first alone flags inputs a hook or an activation
// function changed, the second alone flags a busy-but-idempotent mesh. A waiter
// that drops its inputs never stalls: the count is taken before the drain clears
// them, so it would look unchanged.
func (d *livelockDetector) detect(lastCycle *cycle.Cycle) bool {
	if d.threshold <= 0 {
		return false
	}

	allKeeping := lastCycle.HasActivatedComponents() && lastCycle.ActivationResults().Every(component.WantsToKeepInputs)
	if allKeeping && d.countPendingSignals() == d.pendingSignals {
		d.stalledCycles++
	} else {
		d.stalledCycles = 0
	}

	return d.stalledCycles >= d.threshold
}

// error explains the stall by naming who is stuck and on what: the empty input
// ports of each waiting component point at the pipe that was never wired.
func (d *livelockDetector) error(lastCycle *cycle.Cycle) error {
	// Name enough components to see the pattern; count the rest.
	const maxNamed = 5

	var detail strings.Builder
	starved, named, waiting := 0, 0, 0
	for _, c := range d.components.AllOrdered() {
		// In a stalled cycle every recorded result is a wait keeping inputs, so a
		// component is either waiting or had no input and was not recorded.
		if lastCycle.ActivationResults().ByName(c.Name()) == nil {
			// Counting them is worth the line: in a mutual wait only the component
			// that happens to hold the signal shows up above, and the one starving
			// it is invisible.
			starved++
			continue
		}

		var empty, holding []string
		for _, p := range c.Inputs().AllOrdered() {
			if p.HasSignals() {
				holding = append(holding, p.Name())
			} else {
				empty = append(empty, p.Name())
			}
		}

		waiting++
		if named >= maxNamed {
			continue
		}
		named++

		fmt.Fprintf(&detail, "\n  %q is waiting (keeping inputs): empty input ports %v, holding signals on %v",
			c.Name(), empty, holding)
	}

	if waiting > named {
		fmt.Fprintf(&detail, "\n  ...and %d more waiting component(s)", waiting-named)
	}

	if starved > 0 {
		fmt.Fprintf(&detail, "\n  %d other component(s) have no input signals", starved)
	}

	return fmt.Errorf("%w: no progress for %d consecutive cycles, stopped at cycle #%d. "+
		"Every component that activated is waiting for inputs and no signals moved, "+
		"so every following cycle would be identical to this one%s",
		ErrLivelockDetected, d.stalledCycles, lastCycle.Number(), detail.String())
}
