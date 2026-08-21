# profiler — where the mesh spends its time

A mesh plugin that measures the mesh, and only the mesh: whole runs, single cycles, every
component's activations, and the traffic on every pipe. There is deliberately no process-wide
measurement — no CPU, heap or goroutine sampling — because those numbers cannot be attributed to
the mesh, and a profiler that mixes attributable numbers with ambient ones invites misreading.

The per-component numbers are the reason it exists: a Go CPU profile of a mesh is dominated by
the scheduler and barely distinguishes one component from another, because every component's work
is the same handful of runtime calls. Timing activations directly names the slow one.

```go
import "github.com/hovsep/fmesh/plugin/profiler"

prof := profiler.New()

fm, err := fmesh.New("mesh", fmesh.WithPlugins(prof))
// ... AddComponents, Run ...

fmt.Println(prof.Report())
```

```
runs:   1, total 2.41ms, avg 2.41ms
cycles: 7, total 2.28ms, avg 325µs

component                           count        total          avg          min          max
tokenizer                               7       1.42ms        202µs        180µs        310µs
counter                                 7        612µs         87µs         71µs        140µs
```

| Method | Returns |
|--------|---------|
| `Runs()` | `Stat` for whole runs |
| `Cycles()` | `Stat` for individual cycles |
| `Components()` | `[]ComponentStat`, slowest total first |
| `TopN(n)` | the `n` busiest components, by activation count |
| `Report()` | the table above, as a string |
| `Reset()` | discards everything measured so far |

`Stat` carries `Count`, `Total`, `Min`, `Max` and an `Avg()` method. Stats accumulate across
runs, which is what you want when comparing a mesh against itself; call `Reset()` between runs
that should not be pooled. One profiler belongs to one mesh.

`TopN` and `Components()` answer different questions. `Components()` sorts by total time — *what
is slow*. `TopN` sorts by activation count — *what is hot*. A component that activates every cycle
and does almost nothing can matter more than one that is individually slow but rarely runs.

## Choosing what to measure

Time is one dimension of three. The others are off by default, because their costs differ and an
always-on measurement distorts what it measures — the same reason Go's own block and mutex
profiles have to be switched on deliberately while CPU and heap do not.

```go
prof := profiler.New(profiler.ModeTiming | profiler.ModeThroughput)
everything := profiler.New(profiler.ModeAll)
```

| Mode | Measures | Cost |
|------|----------|------|
| `profiler.ModeTiming` | runs, cycles, component activations | the default |
| `profiler.ModeThroughput` | signals through each pipe | a hook on every output port, fired once per pipe per flush |
| `profiler.ModeTimeline` | one record per cycle | retains a record per cycle, bounded by `SetTimelineLimit` |

`profiler.New()` with no arguments measures `ModeTiming`. `Modes()` reports what a profiler is
measuring.

## Pipe throughput — which connections carry the traffic

With `ModeThroughput`, the report grows a second table:

```
pipe                                                  transfers    signals      avg    min    max
tokenizer.out -> counter.in                                   5         25     5.00      5      5
counter.out -> sink.in                                        5          5     1.00      1      1
tokenizer.debug -> logger.in                                  0          0     0.00      0      0
```

| Method | Returns |
|--------|---------|
| `Pipes()` | `[]PipeStat`, most signals first |
| `TopNPipes(n)` | the `n` pipes that carried a batch most often |

`Flow` carries `Transfers`, `Signals`, `Min`, `Max` and an `Avg()` batch size. `PipeStat` adds
the `Source` and `Destination` labels — `component.port` — and a `Pipe()` method rendering the
edge as one string.

There is no accessor for the coldest pipes because `Pipes()` is volume-sorted: they are its tail.
That last row above is the point — a pipe that was wired and never carried anything. Pipes are
*registered* as well as counted, so one that never fired still appears, which no hook firing only
on traffic could report.

`TopNPipes` is to `Pipes()` what `TopN` is to `Components()`: how often, not how much. A pipe
firing every cycle with a single signal shapes a mesh more than one that moved a thousand signals
once.

Pipes wired *after* `AddComponents` are measured — the hook lives on the port, not the pipe.
Output ports added later with `AddOutputs` are not, which is the limitation every mesh plugin
shares: a plugin only ever sees a component as it arrives.

## Timeline — any cycle-level stat, against the cycle number

With `ModeTimeline`, `Timeline()` returns one `CycleRecord` per cycle — the raw rows for a chart,
with no formatting imposed:

```go
for _, r := range prof.Timeline() {
    fmt.Println(r.Number, r.Duration, r.Activations, r.SignalsMoved)
}
```

`Run` and `Number` are the x-axis; `Duration`, `Activations`, `Errors`, `Panics`, `Waiting` and
`SignalsMoved` are y-axes. `Run` matters because cycle numbers restart at 1 on every run while
the timeline accumulates — `Number` alone is not unique.

Two things about the shape of this data are worth knowing before reading a chart of it:

- A cycle's drain runs *after* the cycle, so `SignalsMoved` counts what the drain following that
  cycle delivered. The **last cycle of a run always reports zero**: the mesh stops before
  draining it.
- A `Duration` of zero on a record means the cycle never finished — a failing `BeforeCycle` hook
  — rather than an instant cycle.

`SetTimelineLimit(n)` caps retention, defaulting to 10,000 records (about 1 MB); zero means
unlimited. It is a ceiling rather than an exact count, because eviction drops the oldest in
chunks.

## Goroutine labels — attributing what the profiler does not measure

The profiler measures nothing process-wide, but it makes the process-wide tools attributable:
every activation runs on a goroutine labeled with `fmesh.mesh` and `fmesh.component`
(`runtime/pprof` labels), whatever modes are enabled. A CPU profile taken while the mesh runs can
then be focused on one component —

```
go tool pprof -tagfocus=fmesh.component=tokenizer cpu.prof
```

— and since Go 1.27 goroutine tracebacks carry the labels in their header, so the stack a
`component.PanicError` keeps names the component that panicked:

```
goroutine 12 [running] {fmesh.component: tokenizer, fmesh.mesh: mesh}:
```

Labels set on the run context by the caller (`pprof.Do`) are kept; the mesh's are added to them.

## More

- [Plugins wiki page](https://github.com/hovsep/fmesh/wiki/502.-Plugins) — how mesh plugins work.
- [API reference](https://pkg.go.dev/github.com/hovsep/fmesh/plugin/profiler).
