package profiler

import (
	"cmp"
	"context"
	"fmt"
	"runtime/pprof"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/hovsep/fmesh"
	"github.com/hovsep/fmesh/component"
)

// Mode selects which dimensions a [Plugin] measures. Modes are bit
// flags and combine with |.
//
// Dimensions are opt-in because their costs differ by orders of magnitude and an
// always-on dimension distorts what it measures. This mirrors how Go's own
// profiles work: CPU and heap are cheap enough to default on, block and mutex
// have to be switched on deliberately.
type Mode uint8

const (
	// ModeTiming measures wall-clock time: runs, cycles, and each component's
	// activations. The default.
	ModeTiming Mode = 1 << iota

	// ModeThroughput counts the signals moving through each pipe. Registers a
	// hook on every output port, fired once per pipe on every flush.
	ModeThroughput

	// ModeTimeline records one [CycleRecord] per cycle, for plotting any
	// cycle-level stat against the cycle number.
	ModeTimeline
)

// ModeAll enables every dimension.
const ModeAll = ModeTiming | ModeThroughput | ModeTimeline

// modeNames is ordered by bit position so String renders deterministically.
var modeNames = []struct {
	mode Mode
	name string
}{
	{ModeTiming, "timing"},
	{ModeThroughput, "throughput"},
	{ModeTimeline, "timeline"},
}

// String renders the enabled modes as "timing|timeline".
func (m Mode) String() string {
	if m == 0 {
		return "none"
	}

	enabled := make([]string, 0, len(modeNames))
	for _, known := range modeNames {
		if m&known.mode != 0 {
			enabled = append(enabled, known.name)
		}
	}
	return strings.Join(enabled, "|")
}

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

// Plugin measures where a mesh spends its time: whole runs, single cycles, and
// each component's activations.
//
// The component numbers are the interesting ones, and they are the reason this
// is a plugin rather than something you reach for a CPU profile to answer. A Go
// profile of a mesh is dominated by the scheduler and tells you almost nothing
// about which component is slow, because every component's work is the same
// handful of runtime calls. Timing activations directly names the culprit.
//
// What it measures beyond timing is selected with [Mode]: per-pipe throughput
// and a per-cycle timeline.
//
// A Plugin holds the measurements of one mesh: it tracks a single in-flight run
// and cycle, so sharing an instance between two meshes interleaves their numbers.
// Stats accumulate across runs of that mesh until Reset.
type Plugin struct {
	modes Mode

	mu         sync.Mutex
	run        Stat
	cycle      Stat
	components map[string]Stat

	runStarted   time.Time
	cycleStarted time.Time
	started      map[string]time.Time

	pipes map[pipeKey]Flow

	timeline      []CycleRecord
	timelineLimit int
	recordStarted time.Time
	// current indexes the in-flight cycle's record, or is negative between
	// cycles. The drain runs after AfterCycle, so signals moved by it must still
	// find the cycle they belong to.
	current  int
	runIndex int
}

// New returns a plugin measuring the given dimensions, which are OR'd together.
// With no arguments it measures [ModeTiming] alone.
func New(modes ...Mode) *Plugin {
	enabled := Mode(0)
	for _, m := range modes {
		enabled |= m
	}
	if enabled == 0 {
		enabled = ModeTiming
	}

	p := &Plugin{
		modes:         enabled,
		components:    make(map[string]Stat),
		started:       make(map[string]time.Time),
		pipes:         make(map[pipeKey]Flow),
		timelineLimit: defaultTimelineLimit,
		current:       -1,
	}
	return p
}

// Modes reports the dimensions this profiler measures.
func (p *Plugin) Modes() Mode { return p.modes }

// Name implements fmesh.Plugin.
func (p *Plugin) Name() string { return "profiler" }

// start stamps a begin time under the lock.
func (p *Plugin) start(at *time.Time) {
	p.mu.Lock()
	defer p.mu.Unlock()
	*at = time.Now()
}

// finish folds the elapsed time since started into stat. A zero start means
// the measurement was reset mid-flight; there is nothing to fold.
func (p *Plugin) finish(stat *Stat, started *time.Time) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if started.IsZero() {
		return
	}
	*stat = stat.with(time.Since(*started))
	*started = time.Time{}
}

// Init implements fmesh.Plugin.
//
// Modes are read once, here: a disabled dimension registers no hooks at all,
// rather than registering hooks that check whether they should do anything.
func (p *Plugin) Init(fm *fmesh.FMesh) error {
	fm.SetupHooks(func(hooks *fmesh.Hooks) {
		// First, so the label hook precedes the timing hook on every component
		// and stays outside the measured window.
		hooks.OnComponentAdded(func(_ context.Context, added *fmesh.ComponentAddedContext) error {
			labelGoroutines(added.FMesh.Name(), added.Component)
			return nil
		})
		if p.modes&ModeTiming != 0 {
			p.initTiming(hooks)
		}
		if p.modes&ModeTimeline != 0 {
			p.initTimeline(hooks)
		}
		if p.modes&ModeThroughput != 0 {
			hooks.OnComponentAdded(func(_ context.Context, added *fmesh.ComponentAddedContext) error {
				return p.instrumentPipes(added.Component)
			})
		}
	})
	return nil
}

// initTiming registers the wall-clock hooks.
func (p *Plugin) initTiming(hooks *fmesh.Hooks) {
	hooks.BeforeRun(func(context.Context, *fmesh.FMesh) error {
		p.start(&p.runStarted)
		return nil
	})
	hooks.AfterRun(func(context.Context, *fmesh.FMesh) error {
		p.finish(&p.run, &p.runStarted)
		return nil
	})
	hooks.BeforeCycle(func(context.Context, *fmesh.CycleContext) error {
		p.start(&p.cycleStarted)
		return nil
	})
	hooks.AfterCycle(func(context.Context, *fmesh.CycleContext) error {
		p.finish(&p.cycle, &p.cycleStarted)
		return nil
	})
	hooks.OnComponentAdded(func(_ context.Context, added *fmesh.ComponentAddedContext) error {
		p.instrument(added.Component)
		return nil
	})
}

// labelGoroutines tags each activation goroutine with runtime/pprof labels
// naming the mesh and the component. BeforeActivation runs on the goroutine the
// scheduler spawned for this activation, and that goroutine ends with it, so
// nothing needs resetting. CPU profiles can then be focused per component
// (go tool pprof -tagfocus=fmesh.component=NAME), and since Go 1.27 tracebacks,
// including a PanicError's stack, carry the labels in their header.
//
// WithLabels merges with whatever the run context already carries, so labels
// the caller set with pprof.Do survive.
func labelGoroutines(meshName string, c *component.Component) {
	labels := pprof.Labels("fmesh.mesh", meshName, "fmesh.component", c.Name())
	c.SetupHooks(func(hooks *component.Hooks) {
		hooks.BeforeActivation(func(ctx context.Context, _ *component.Component) error {
			pprof.SetGoroutineLabels(pprof.WithLabels(ctx, labels))
			return nil
		})
	})
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
// Do not stamp the start before its lock: that moves the BeforeActivation
// contention into the window, where it dominates.
func (p *Plugin) instrument(c *component.Component) {
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
func (p *Plugin) Runs() Stat {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.run
}

// Cycles reports the per-cycle timings.
func (p *Plugin) Cycles() Stat {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.cycle
}

// Components reports each component's activation timings, slowest total first.
func (p *Plugin) Components() []ComponentStat {
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
func (p *Plugin) TopN(n int) []ComponentStat {
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
func (p *Plugin) Reset() {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.run, p.cycle = Stat{}, Stat{}
	p.runStarted, p.cycleStarted, p.recordStarted = time.Time{}, time.Time{}, time.Time{}
	p.components = make(map[string]Stat)
	p.started = make(map[string]time.Time)

	// A registered pipe is topology, not measurement: zeroing rather than
	// dropping keeps a pipe that has never carried anything visible as the
	// coldest one there is.
	for key := range p.pipes {
		p.pipes[key] = Flow{}
	}

	p.timeline, p.current, p.runIndex = nil, -1, 0
}

// Report renders the profile as text, slowest component first.
//
// Each dimension beyond timing contributes a block only when its mode is
// enabled. The timeline is summarized rather than tabulated -- use [Plugin.Timeline] for the
// rows themselves.
func (p *Plugin) Report() string {
	var b strings.Builder

	if p.modes != ModeTiming {
		fmt.Fprintf(&b, "modes:  %v\n", p.modes)
	}

	runs, cycles := p.Runs(), p.Cycles()
	fmt.Fprintf(&b, "runs:   %d, total %v, avg %v\n", runs.Count, runs.Total, runs.Avg())
	fmt.Fprintf(&b, "cycles: %d, total %v, avg %v\n", cycles.Count, cycles.Total, cycles.Avg())
	fmt.Fprintf(&b, "\n%-32s %8s %12s %12s %12s %12s\n",
		"component", "count", "total", "avg", "min", "max")

	for _, s := range p.Components() {
		fmt.Fprintf(&b, "%-32s %8d %12v %12v %12v %12v\n",
			s.Component, s.Count, s.Total, s.Avg(), s.Min, s.Max)
	}

	if p.modes&ModeThroughput != 0 {
		p.reportPipes(&b)
	}
	if p.modes&ModeTimeline != 0 {
		fmt.Fprintf(&b, "\ntimeline: %d cycles recorded\n", len(p.Timeline()))
	}
	return b.String()
}

// reportPipes writes the throughput table, hottest pipe first.
func (p *Plugin) reportPipes(b *strings.Builder) {
	fmt.Fprintf(b, "\n%-52s %10s %10s %8s %6s %6s\n",
		"pipe", "transfers", "signals", "avg", "min", "max")

	for _, s := range p.Pipes() {
		fmt.Fprintf(b, "%-52s %10d %10d %8.2f %6d %6d\n",
			s.Pipe(), s.Transfers, s.Signals, s.Avg(), s.Min, s.Max)
	}
}
