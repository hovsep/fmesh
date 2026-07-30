// Package plugin provides mesh-level plugins that ship with F-Mesh.
//
// Each is a plain fmesh.Plugin: an initialization bundle that registers hooks
// and then gets out of the way. What they have in common is that they reach
// every component through the OnComponentAdded hook, so nothing in a mesh has to
// be written differently to be measured (Profiler) or wired (Autowire) -- both
// are things you add to a mesh rather than things components opt into.
package plugin

import (
	"cmp"
	"context"
	"fmt"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/hovsep/fmesh"
	"github.com/hovsep/fmesh/component"
)

// Stat is how long something took, over however many times it happened.
type Stat struct {
	Count int
	Total time.Duration
	Min   time.Duration
	Max   time.Duration
}

// Avg is the mean duration, or zero if it never happened.
func (s Stat) Avg() time.Duration {
	if s.Count == 0 {
		return 0
	}
	return s.Total / time.Duration(s.Count)
}

// with returns the stat updated to include one more observation.
func (s Stat) with(d time.Duration) Stat {
	if s.Count == 0 || d < s.Min {
		s.Min = d
	}
	if d > s.Max {
		s.Max = d
	}
	s.Count++
	s.Total += d
	return s
}

// Profiler measures where a mesh spends its time: whole runs, single cycles, and
// each component's activations.
//
// The component numbers are the interesting ones, and they are the reason this
// is a plugin rather than something you reach for a CPU profile to answer. A Go
// profile of a mesh is dominated by the scheduler and tells you almost nothing
// about which component is slow, because every component's work is the same
// handful of runtime calls. Timing activations directly names the culprit.
//
// A Profiler holds the timings of one mesh: it tracks a single in-flight run and
// cycle, so sharing an instance between two meshes interleaves their numbers.
// Stats accumulate across runs of that mesh until Reset.
type Profiler struct {
	mu         sync.Mutex
	run        Stat
	cycle      Stat
	components map[string]Stat

	runStarted   time.Time
	cycleStarted time.Time
	started      map[string]time.Time
}

// NewProfiler returns a profiler plugin.
func NewProfiler() *Profiler {
	return &Profiler{
		components: make(map[string]Stat),
		started:    make(map[string]time.Time),
	}
}

// GetName implements fmesh.Plugin.
func (p *Profiler) GetName() string { return "profiler" }

// Init implements fmesh.Plugin.
func (p *Profiler) Init(fm *fmesh.FMesh) error {
	fm.SetupHooks(func(hooks *fmesh.Hooks) {
		hooks.BeforeRun(func(context.Context, *fmesh.FMesh) error {
			p.mu.Lock()
			defer p.mu.Unlock()
			p.runStarted = time.Now()
			return nil
		})
		hooks.AfterRun(func(context.Context, *fmesh.FMesh) error {
			p.mu.Lock()
			defer p.mu.Unlock()
			p.run = p.run.with(time.Since(p.runStarted))
			return nil
		})
		hooks.BeforeCycle(func(context.Context, *fmesh.CycleContext) error {
			p.mu.Lock()
			defer p.mu.Unlock()
			p.cycleStarted = time.Now()
			return nil
		})
		hooks.AfterCycle(func(context.Context, *fmesh.CycleContext) error {
			p.mu.Lock()
			defer p.mu.Unlock()
			p.cycle = p.cycle.with(time.Since(p.cycleStarted))
			return nil
		})
		hooks.OnComponentAdded(func(_ context.Context, added *fmesh.ComponentAddedContext) error {
			p.instrument(added.Component)
			return nil
		})
	})
	return nil
}

// instrument times one component's activations.
//
// Components activate concurrently, so the start times live in a map under the
// same lock as the stats rather than in a field. Every component in a cycle
// therefore queues on this one mutex, and the goal is to keep that queueing
// outside the measured window: the start is stamped as late as possible (inside
// the lock, just before the activation returns to the scheduler) and the end as
// early as possible (before the lock, the instant the activation finished).
//
// Stamping the start *before* its lock instead looks symmetrical and is much
// worse -- it moves the BeforeActivation contention into the window, where it
// dominates. Measured on a 500-component cycle doing ~2us of work each: ~18us
// reported this way, ~180us with the start stamped early.
func (p *Profiler) instrument(c *component.Component) {
	c.SetupHooks(func(hooks *component.Hooks) {
		hooks.BeforeActivation(func(_ context.Context, this *component.Component) error {
			p.mu.Lock()
			defer p.mu.Unlock()
			p.started[this.Name()] = time.Now()
			return nil
		})
		hooks.AfterActivation(func(_ context.Context, activation *component.ActivationContext) error {
			endedAt := time.Now()

			p.mu.Lock()
			defer p.mu.Unlock()

			name := activation.Component.Name()
			startedAt, ok := p.started[name]
			if !ok {
				// AfterActivation is a finally block and can fire for an
				// activation that never began.
				return nil
			}
			delete(p.started, name)

			p.components[name] = p.components[name].with(endedAt.Sub(startedAt))
			return nil
		})
	})
}

// Runs reports the whole-run timings.
func (p *Profiler) Runs() Stat {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.run
}

// Cycles reports the per-cycle timings.
func (p *Profiler) Cycles() Stat {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.cycle
}

// Components reports each component's activation timings, slowest total first.
func (p *Profiler) Components() []ComponentStat {
	p.mu.Lock()
	defer p.mu.Unlock()

	stats := make([]ComponentStat, 0, len(p.components))
	for name, stat := range p.components {
		stats = append(stats, ComponentStat{Component: name, Stat: stat})
	}
	slices.SortFunc(stats, func(a, b ComponentStat) int {
		if c := cmp.Compare(b.Total, a.Total); c != 0 {
			return c
		}
		return cmp.Compare(a.Component, b.Component)
	})
	return stats
}

// ComponentStat is one component's activation timings.
type ComponentStat struct {
	Component string
	Stat
}

// TopN returns the n components that activated most often, busiest first.
// An n of zero or less returns nothing.
//
// "Hottest" and "slowest" are different questions and this answers the first:
// a component that activates on every cycle and does almost nothing can matter
// more than one that is individually slow but rarely runs.
func (p *Profiler) TopN(n int) []ComponentStat {
	stats := p.Components()
	slices.SortFunc(stats, func(a, b ComponentStat) int {
		if c := cmp.Compare(b.Count, a.Count); c != 0 {
			return c
		}
		return cmp.Compare(a.Component, b.Component)
	})
	return stats[:max(0, min(n, len(stats)))]
}

// Reset discards everything measured so far.
//
// Stats accumulate across runs, which is what you want when comparing a mesh
// against itself; call this between runs that should not be pooled.
func (p *Profiler) Reset() {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.run, p.cycle = Stat{}, Stat{}
	p.components = make(map[string]Stat)
	p.started = make(map[string]time.Time)
}

// Report renders the profile as a table, slowest component first.
func (p *Profiler) Report() string {
	var b strings.Builder

	runs, cycles := p.Runs(), p.Cycles()
	fmt.Fprintf(&b, "runs:   %d, total %v, avg %v\n", runs.Count, runs.Total, runs.Avg())
	fmt.Fprintf(&b, "cycles: %d, total %v, avg %v\n", cycles.Count, cycles.Total, cycles.Avg())
	fmt.Fprintf(&b, "\n%-32s %8s %12s %12s %12s %12s\n",
		"component", "count", "total", "avg", "min", "max")

	for _, s := range p.Components() {
		fmt.Fprintf(&b, "%-32s %8d %12v %12v %12v %12v\n",
			s.Component, s.Count, s.Total, s.Avg(), s.Min, s.Max)
	}
	return b.String()
}
