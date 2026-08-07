package profiler

import (
	"context"
	"math"
	"runtime/metrics"
	"time"

	"github.com/hovsep/fmesh"
)

// Sampled metric names, in the order resourceSnapshot reads them back.
const (
	metricAllocBytes   = "/gc/heap/allocs:bytes"
	metricAllocObjects = "/gc/heap/allocs:objects"
	metricGCCycles     = "/gc/cycles/total:gc-cycles"
	metricLiveHeap     = "/memory/classes/heap/objects:bytes"
	metricGoroutines   = "/sched/goroutines:goroutines"
	metricCPUUser      = "/cpu/classes/user:cpu-seconds"
	metricCPUGC        = "/cpu/classes/gc/total:cpu-seconds"
)

// ResourceStat is the change in process-wide Go runtime counters across one
// measured window.
//
// These numbers describe the process, not the mesh. Every field is a delta
// between two runtime/metrics samples taken at the window's boundaries, so
// anything else the process did in that window -- other goroutines, other meshes,
// the test harness, an HTTP server in the same binary -- is counted too. Read
// them as what the Go runtime did while this mesh ran, never as what this mesh
// allocated. Only the timing and per-component numbers are attributable to the
// mesh itself.
type ResourceStat struct {
	// AllocBytes is heap bytes allocated during the window. Monotonic, so never
	// negative.
	AllocBytes uint64

	// AllocObjects is heap objects allocated during the window. It separates
	// many small signals from a few large payloads, which AllocBytes cannot.
	AllocObjects uint64

	// GCCycles is GC cycles completed during the window.
	GCCycles uint64

	// LiveHeapDelta is the change in live heap bytes, negative when the window
	// ended holding less than it started with.
	LiveHeapDelta int64

	// GoroutineDelta is the change in goroutine count. A mesh leaking activation
	// goroutines shows up here; one that does not returns to zero.
	GoroutineDelta int64

	// CPUUser and CPUGC are estimated CPU time in user Go code and in the
	// garbage collector.
	//
	// The runtime advances these only at GC mark termination, so a window that
	// completed no GC cycle reports zero for both however long it ran -- check
	// GCCycles before reading anything into them. They are overestimates,
	// comparable with each other but not with wall time or OS CPU accounting.
	CPUUser time.Duration
	CPUGC   time.Duration
}

// resourceSnapshot is one absolute reading of the sampled counters.
type resourceSnapshot struct {
	allocBytes   uint64
	allocObjects uint64
	gcCycles     uint64
	liveHeap     uint64
	goroutines   uint64
	cpuUser      float64
	cpuGC        float64
}

// since returns the delta from an earlier snapshot. The two gauges are signed
// because live heap and goroutine count both go down.
func (later resourceSnapshot) since(earlier resourceSnapshot) ResourceStat {
	return ResourceStat{
		AllocBytes:     later.allocBytes - earlier.allocBytes,
		AllocObjects:   later.allocObjects - earlier.allocObjects,
		GCCycles:       later.gcCycles - earlier.gcCycles,
		LiveHeapDelta:  signedDelta(later.liveHeap, earlier.liveHeap),
		GoroutineDelta: signedDelta(later.goroutines, earlier.goroutines),
		CPUUser:        seconds(later.cpuUser - earlier.cpuUser),
		CPUGC:          seconds(later.cpuGC - earlier.cpuGC),
	}
}

// signedDelta subtracts two gauge readings.
//
// Subtracting after converting to int64 would be shorter and wrong: either
// operand can exceed MaxInt64, and the conversion wraps silently. Doing the
// arithmetic in uint64 and clamping keeps an absurd reading absurd instead of
// turning it into a plausible number with the wrong sign.
func signedDelta(later, earlier uint64) int64 {
	if later >= earlier {
		grew := later - earlier
		if grew > math.MaxInt64 {
			return math.MaxInt64
		}
		return int64(grew)
	}

	shrank := earlier - later
	if shrank > math.MaxInt64 {
		return math.MinInt64
	}
	return -int64(shrank)
}

// add accumulates one window's delta into another, so runs pool the way the
// timing stats do.
func (s ResourceStat) add(other ResourceStat) ResourceStat {
	s.AllocBytes += other.AllocBytes
	s.AllocObjects += other.AllocObjects
	s.GCCycles += other.GCCycles
	s.LiveHeapDelta += other.LiveHeapDelta
	s.GoroutineDelta += other.GoroutineDelta
	s.CPUUser += other.CPUUser
	s.CPUGC += other.CPUGC
	return s
}

// seconds converts a cpu-seconds reading to a Duration.
func seconds(f float64) time.Duration {
	return time.Duration(f * float64(time.Second))
}

// resourceSampler holds the preallocated runtime/metrics sample set.
//
// runtime/metrics is used rather than runtime.ReadMemStats because ReadMemStats
// stops the world -- tens of microseconds per call, which on a cycle measured in
// microseconds would not be profiling but a different program, and would
// serialize the very concurrency being measured.
//
// The buffer is reused across boundaries without a lock. Every run and cycle hook
// fires from the goroutine running the mesh; only component and port hooks are
// concurrent, so sampling from one of those would break this silently.
type resourceSampler struct {
	samples []metrics.Sample
}

func newResourceSampler() *resourceSampler {
	names := []string{
		metricAllocBytes,
		metricAllocObjects,
		metricGCCycles,
		metricLiveHeap,
		metricGoroutines,
		metricCPUUser,
		metricCPUGC,
	}

	samples := make([]metrics.Sample, len(names))
	for i, name := range names {
		samples[i].Name = name
	}
	return &resourceSampler{samples: samples}
}

// read takes one reading. A metric the runtime does not recognize reports
// KindBad and is left at zero rather than panicking.
func (s *resourceSampler) read() resourceSnapshot {
	metrics.Read(s.samples)

	var snap resourceSnapshot
	for _, sample := range s.samples {
		switch sample.Name {
		case metricAllocBytes:
			snap.allocBytes = uint64Value(sample)
		case metricAllocObjects:
			snap.allocObjects = uint64Value(sample)
		case metricGCCycles:
			snap.gcCycles = uint64Value(sample)
		case metricLiveHeap:
			snap.liveHeap = uint64Value(sample)
		case metricGoroutines:
			snap.goroutines = uint64Value(sample)
		case metricCPUUser:
			snap.cpuUser = float64Value(sample)
		case metricCPUGC:
			snap.cpuGC = float64Value(sample)
		}
	}
	return snap
}

func uint64Value(s metrics.Sample) uint64 {
	if s.Value.Kind() != metrics.KindUint64 {
		return 0
	}
	return s.Value.Uint64()
}

func float64Value(s metrics.Sample) float64 {
	if s.Value.Kind() != metrics.KindFloat64 {
		return 0
	}
	return s.Value.Float64()
}

// initResources registers the sampling hooks. Cycle boundaries are sampled only
// when a timeline exists to hold the per-cycle numbers.
func (p *Plugin) initResources(hooks *fmesh.Hooks) {
	hooks.BeforeRun(func(context.Context, *fmesh.FMesh) error {
		// Read outside the lock, like every other measurement here.
		snapshot := p.sampler.read()

		p.mu.Lock()
		defer p.mu.Unlock()
		p.runSnapshot = snapshot
		return nil
	})
	hooks.AfterRun(func(context.Context, *fmesh.FMesh) error {
		snapshot := p.sampler.read()

		p.mu.Lock()
		defer p.mu.Unlock()
		p.resources = p.resources.add(snapshot.since(p.runSnapshot))
		return nil
	})

	if p.modes&ModeTimeline == 0 {
		return
	}

	hooks.BeforeCycle(func(context.Context, *fmesh.CycleContext) error {
		snapshot := p.sampler.read()

		p.mu.Lock()
		defer p.mu.Unlock()
		p.cycleSnapshot = snapshot
		return nil
	})
	hooks.AfterCycle(func(context.Context, *fmesh.CycleContext) error {
		snapshot := p.sampler.read()

		p.mu.Lock()
		defer p.mu.Unlock()
		if p.current >= 0 {
			p.timeline[p.current].Resources = snapshot.since(p.cycleSnapshot)
		}
		return nil
	})
}

// Resources reports the runtime counter deltas accumulated over every measured
// run. Zero unless ModeRuntime is enabled.
func (p *Plugin) Resources() ResourceStat {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.resources
}
