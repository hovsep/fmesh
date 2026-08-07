package profiler

import (
	"context"
	"slices"
	"time"

	"github.com/hovsep/fmesh"
	"github.com/hovsep/fmesh/component"
)

// defaultTimelineLimit bounds the timeline out of the box. A CycleRecord is
// roughly 112 bytes, so this is about 1 MB -- enough for any run worth plotting,
// and bounded for a mesh built with WithUnlimitedCycles.
const defaultTimelineLimit = 10_000

// timelineEvictionShare is the fraction of the timeline dropped when it fills.
// Evicting in one chunk rather than per cycle keeps the copy amortized: shifting
// the whole slice down by one on every cycle would move more memory than the
// rest of the profiler does combined.
const timelineEvictionShare = 4

// CycleRecord is one row of the per-cycle timeline: Run and Number are the
// x-axis, everything else is a y-axis.
//
// A cycle's drain runs after its AfterCycle hook, so SignalsMoved is filled in
// after Duration is. Two consequences: the last cycle of a run always reports
// zero signals moved, because the run stops before draining it; and a Duration of
// zero on a record means the cycle never reached AfterCycle, which is a failed
// hook rather than an instant cycle.
type CycleRecord struct {
	// Run is the 1-based index of the run this cycle belongs to. Cycle numbers
	// restart at 1 on every run, so Number alone is not a unique x-axis once
	// measurements have accumulated across runs.
	Run    int
	Number int

	Duration time.Duration

	// Activations counts components that activated; results with no input are
	// never recorded, so they are not counted here either.
	Activations int
	Errors      int
	Panics      int
	// Waiting counts components that asked to wait for more input, which is the
	// stat a livelock shows up in.
	Waiting int

	// SignalsMoved is the signals delivered by the drain that followed this
	// cycle. Zero unless ModeThroughput is enabled.
	SignalsMoved int

	// Resources is the runtime delta across this cycle. Zero unless both
	// ModeRuntime and ModeTimeline are enabled.
	Resources ResourceStat
}

// initTimeline registers the hooks that open and fill one record per cycle.
func (p *Plugin) initTimeline(hooks *fmesh.Hooks) {
	hooks.BeforeRun(func(context.Context, *fmesh.FMesh) error {
		p.mu.Lock()
		defer p.mu.Unlock()
		p.runIndex++
		return nil
	})
	hooks.BeforeCycle(func(_ context.Context, cycleCtx *fmesh.CycleContext) error {
		p.openRecord(cycleCtx.Cycle.Number())
		return nil
	})
	hooks.AfterCycle(func(_ context.Context, cycleCtx *fmesh.CycleContext) error {
		p.closeRecord(cycleCtx)
		return nil
	})
	hooks.AfterRun(func(context.Context, *fmesh.FMesh) error {
		p.mu.Lock()
		defer p.mu.Unlock()
		// Anything put on a port after the run belongs to no cycle.
		p.current = -1
		return nil
	})
}

// openRecord appends the record for a cycle about to run and points current at
// it. Eviction happens here, immediately before the append, so no stale index
// outlives the statement that produced it.
func (p *Plugin) openRecord(number int) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.timelineLimit > 0 && len(p.timeline) >= p.timelineLimit {
		p.timeline = slices.Delete(p.timeline, 0, max(1, p.timelineLimit/timelineEvictionShare))
	}

	p.timeline = append(p.timeline, CycleRecord{Run: p.runIndex, Number: number})
	p.current = len(p.timeline) - 1
	// Stamped here rather than read from the timing dimension, so a timeline
	// without ModeTiming still reports durations.
	p.recordStarted = time.Now()
}

// closeRecord fills in what the finished cycle knows about itself. The counts
// come from the cycle rather than from per-component hooks: the activation
// results are already collected, on this goroutine, with no lock to contend.
func (p *Plugin) closeRecord(cycleCtx *fmesh.CycleContext) {
	endedAt := time.Now()
	results := cycleCtx.Cycle.ActivationResults()

	activations := results.Count(func(r *component.ActivationResult) bool { return r.Activated() })
	errs := results.Count((*component.ActivationResult).IsError)
	panics := results.Count((*component.ActivationResult).IsPanic)
	waiting := results.Count(component.IsWaitingForInput)

	p.mu.Lock()
	defer p.mu.Unlock()

	if p.current < 0 {
		return
	}
	record := &p.timeline[p.current]
	record.Duration = endedAt.Sub(p.recordStarted)
	record.Activations = activations
	record.Errors = errs
	record.Panics = panics
	record.Waiting = waiting
}

// Timeline returns one record per cycle, oldest first. Empty unless
// ModeTimeline is enabled.
func (p *Plugin) Timeline() []CycleRecord {
	p.mu.Lock()
	defer p.mu.Unlock()
	return slices.Clone(p.timeline)
}

// SetTimelineLimit caps how many of the most recent cycle records are kept,
// dropping the oldest in batches when the cap is reached. Zero means unlimited.
//
// The cap is a ceiling, not a target: the retained count oscillates below it
// because eviction happens in chunks.
func (p *Plugin) SetTimelineLimit(limit int) *Plugin {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.timelineLimit = limit
	return p
}
