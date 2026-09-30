# profiler: where the mesh spends its time

A mesh plugin that measures the mesh and only the mesh: whole runs, single cycles, each
component's activations and the traffic on each pipe. It does no process-wide sampling (CPU, heap,
goroutines), because those numbers cannot be attributed to the mesh.

Why per-component numbers: a Go CPU profile of a mesh is mostly scheduler work. Every component
calls the same few runtime functions, so the profile barely tells them apart. Timing each
activation names the slow component directly.

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
| `Cycles()` | `Stat` for single cycles |
| `Components()` | `[]ComponentStat`, highest total time first: *what is slow* |
| `TopN(n)` | the `n` components with the most activations: *what is hot* |
| `Report()` | the tables above, as a string |
| `Reset()` | drops everything measured so far |

- `Stat` has `Count`, `Total`, `Min`, `Max` and an `Avg()` method.
- Stats add up across runs. Call `Reset()` between runs you do not want to pool.
- One profiler belongs to one mesh.
- Hot can matter more than slow: a component that runs every cycle and does little can cost more
  than a slow one that rarely runs.

## Choosing what to measure

Timing is on by default. The other two modes are opt-in, because each has a cost.

```go
prof := profiler.New(profiler.ModeTiming | profiler.ModeThroughput)
everything := profiler.New(profiler.ModeAll)
```

| Mode | Measures | Cost |
|------|----------|------|
| `profiler.ModeTiming` | runs, cycles, component activations | the default |
| `profiler.ModeThroughput` | signals through each pipe | a hook on every output port, fired once per pipe per flush |
| `profiler.ModeTimeline` | one record per cycle | keeps a record per cycle, capped by `SetTimelineLimit` |

`profiler.New()` with no arguments means `ModeTiming`. `Modes()` reports what a profiler measures.

## Pipe throughput

With `ModeThroughput`, the report adds a pipe table:

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

- `Flow` has `Transfers`, `Signals`, `Min`, `Max` and an `Avg()` batch size.
- `PipeStat` adds `Source` and `Destination` (as `component.port`) and a `Pipe()` method that
  renders the edge as one string.
- Every wired pipe is listed, even one that never carried a signal (the last row above). Unused
  pipes sort last in `Pipes()`, so the coldest pipes are at its tail.
- `TopNPipes` counts how often, not how much, like `TopN`.
- Pipes created after `AddComponents` are measured: the hook lives on the output port. Output
  ports added later with `AddOutputs` are not, because a plugin sees a component only when it is added.

## Timeline

With `ModeTimeline`, `Timeline()` returns one `CycleRecord` per cycle, ready to plot:

```go
for _, r := range prof.Timeline() {
    fmt.Println(r.Number, r.Duration, r.Activations, r.SignalsMoved)
}
```

- X-axis: `Run` and `Number`. Cycle numbers restart at 1 on each run, so use both.
- Y-axis: `Duration`, `Activations`, `Errors`, `Panics`, `Waiting`, `SignalsMoved`.
- `SignalsMoved` counts what the drain *after* that cycle delivered, and needs `ModeThroughput`.
  The **last cycle of a run always shows zero**: the mesh stops before draining it.
- A `Duration` of zero means the cycle never finished (a hook failed), not that it was instant.
- `SetTimelineLimit(n)` caps how many records are kept. The default is 10,000 (about 1 MB); zero
  means no limit. Old records are dropped in chunks, so the cap is a ceiling, not an exact count.

## Goroutine labels

The profiler does not measure the process, but it helps process-wide tools. Every activation runs
on a goroutine with the `runtime/pprof` labels `fmesh.mesh` and `fmesh.component`, in every mode.
So you can focus a CPU profile on one component:

```
go tool pprof -tagfocus=fmesh.component=tokenizer cpu.prof
```

Since Go 1.27, goroutine tracebacks show labels in the header. So the stack kept by a
`component.PanicError` names the component that panicked:

```
goroutine 12 [running] {fmesh.component: tokenizer, fmesh.mesh: mesh}:
```

Labels the caller set on the run context (with `pprof.Do`) are kept; the mesh adds its own.

## More

- [Plugins wiki page](https://github.com/hovsep/fmesh/wiki/502.-Plugins): how mesh plugins work.
- [API reference](https://pkg.go.dev/github.com/hovsep/fmesh/plugin/profiler).
