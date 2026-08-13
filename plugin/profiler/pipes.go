package profiler

import (
	"cmp"
	"context"
	"slices"

	"github.com/hovsep/fmesh/component"
	"github.com/hovsep/fmesh/port"
)

// Flow is how many signals moved, over however many transfers.
type Flow struct {
	Transfers int
	Signals   int
	Min       int
	Max       int
}

// Avg is the mean batch size, or zero if nothing ever moved.
func (f Flow) Avg() float64 {
	if f.Transfers == 0 {
		return 0
	}
	return float64(f.Signals) / float64(f.Transfers)
}

// with returns the flow updated to include one more transfer of n signals.
func (f Flow) with(n int) Flow {
	if f.Transfers == 0 || n < f.Min {
		f.Min = n
	}
	if n > f.Max {
		f.Max = n
	}
	f.Transfers++
	f.Signals += n
	return f
}

// PipeStat is one pipe's throughput.
//
// Each end is labeled "component.port". A port with no parent component is
// labeled by its own name, which only happens outside a mesh -- and only there
// can two pipes share a label, since port names are unique within a component.
type PipeStat struct {
	Source      string
	Destination string
	Flow
}

// Pipe is the edge as one string: "producer.out -> consumer.in".
func (s PipeStat) Pipe() string {
	return s.Source + " -> " + s.Destination
}

// pipeKey identifies a pipe by its endpoints.
//
// Ports have no IDs and a port name is unique only within its component, so the
// pointer pair is the identity. Deriving the printable label on read instead
// keeps string building off the delivery path.
type pipeKey struct {
	source      *port.Port
	destination *port.Port
}

// instrumentPipes counts the signals crossing one component's outbound pipes.
//
// Pipes are registered as well as counted so that one which never fires still
// appears -- a pipe that carried nothing is the coldest pipe there is, and it
// cannot be reported by a hook that only ever fires on traffic. Pipes wired after
// the component arrives are caught by OnOutboundPipe; output ports added later
// are not, the same limitation Autowire has.
func (p *Plugin) instrumentPipes(c *component.Component) error {
	return c.Outputs().ForEach(func(out *port.Port) error {
		if err := out.Pipes().ForEach(func(in *port.Port) error {
			p.registerPipe(out, in)
			return nil
		}); err != nil {
			return err
		}

		out.SetupHooks(func(hooks *port.Hooks) {
			hooks.OnOutboundPipe(func(_ context.Context, wired *port.OutboundPipeContext) error {
				p.registerPipe(wired.SourcePort, wired.DestinationPort)
				return nil
			})
			hooks.OnSignalsDelivered(func(_ context.Context, delivered *port.SignalsDeliveredContext) error {
				p.recordDelivery(delivered.SourcePort, delivered.DestinationPort, len(delivered.SignalsDelivered))
				return nil
			})
		})
		return nil
	})
}

// registerPipe records a pipe that has carried nothing yet.
func (p *Plugin) registerPipe(source, destination *port.Port) {
	key := pipeKey{source: source, destination: destination}

	p.mu.Lock()
	defer p.mu.Unlock()

	if _, known := p.pipes[key]; !known {
		p.pipes[key] = Flow{}
	}
}

// recordDelivery folds one transfer of n signals into a pipe's flow, and into
// the cycle whose drain moved them.
func (p *Plugin) recordDelivery(source, destination *port.Port, n int) {
	key := pipeKey{source: source, destination: destination}

	p.mu.Lock()
	defer p.mu.Unlock()

	p.pipes[key] = p.pipes[key].with(n)

	// Signals seeded before a run or moved by a hand-written Flush belong to no
	// cycle; charging them to the last one would inflate it.
	if p.current >= 0 {
		p.timeline[p.current].SignalsMoved += n
	}
}

// pipeEnd labels one end of a pipe.
func pipeEnd(p *port.Port) string {
	if parent := p.ParentComponent(); parent != nil {
		return parent.Name() + "." + p.Name()
	}
	return p.Name()
}

// Pipes reports each pipe's throughput, most signals first.
//
// A pipe that carried nothing sorts last, which is what makes the cold ones
// findable: there is no separate accessor for them.
func (p *Plugin) Pipes() []PipeStat {
	p.mu.Lock()
	stats := make([]PipeStat, 0, len(p.pipes))
	for key, flow := range p.pipes {
		stats = append(stats, PipeStat{
			Source:      pipeEnd(key.source),
			Destination: pipeEnd(key.destination),
			Flow:        flow,
		})
	}
	p.mu.Unlock()

	slices.SortFunc(stats, func(a, b PipeStat) int {
		if c := cmp.Compare(b.Signals, a.Signals); c != 0 {
			return c
		}
		if c := cmp.Compare(b.Transfers, a.Transfers); c != 0 {
			return c
		}
		return comparePipeLabels(a, b)
	})
	return stats
}

// TopNPipes returns the n pipes that carried a batch most often, busiest first.
// An n of zero or less returns nothing.
//
// As with [Plugin.TopN], this is how often rather than how much: a pipe firing
// every cycle with one signal shapes a mesh more than one that moved a thousand
// signals once.
func (p *Plugin) TopNPipes(n int) []PipeStat {
	stats := p.Pipes()
	slices.SortFunc(stats, func(a, b PipeStat) int {
		if c := cmp.Compare(b.Transfers, a.Transfers); c != 0 {
			return c
		}
		return comparePipeLabels(a, b)
	})
	return stats[:max(0, min(n, len(stats)))]
}

// comparePipeLabels breaks a tie on the endpoint names, so a report cannot
// reorder itself between runs.
func comparePipeLabels(a, b PipeStat) int {
	if c := cmp.Compare(a.Source, b.Source); c != 0 {
		return c
	}
	return cmp.Compare(a.Destination, b.Destination)
}
