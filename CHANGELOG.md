# Changelog

All notable changes to F-Mesh are documented here.

F-Mesh is pre-production. **Minor versions may contain breaking changes** until this notice is
removed; every one is listed under a `BREAKING` heading below. Pin an exact version if that matters
to you.

Release tags are plain semver (`v1.12.0`). The Armenian capital naming lives in the GitHub release
title — a suffix in the tag makes Go treat the release as a pre-release and hides it from `go get`.

## [Unreleased]

Context support, deterministic runs, livelock detection, a smaller API, and readable failures.
Everything in this entry is one migration; do it in one pass.

### BREAKING

**`Run` takes a context.**

```go
ri, err := fm.Run()            // before
ri, err := fm.Run(ctx)         // after
```

**Activation functions take a context.** It is the run context, narrowed by `TimeLimit`. Ignoring
it is safe — name it `_` if you do not need it.

```go
component.WithActivationFunc(func(this *component.Component) error { ... })                    // before
component.WithActivationFunc(func(ctx context.Context, this *component.Component) error { ... }) // after
```

**Hooks take a context first.** Every hook, at every level. Hook context structs are unchanged and
still carry the event data; the context is a parameter, never a struct field.

```go
h.BeforeRun(func(fm *fmesh.FMesh) error { ... })                          // before
h.BeforeRun(func(ctx context.Context, fm *fmesh.FMesh) error { ... })     // after
```

**Runtime-path methods take a context**: `Component.MaybeActivate`, `FlushOutputs`, `ClearInputs`,
`ClearOutputs`, `Port.Flush`, `Port.Clear`, `port.Collection.Flush`, `port.ForwardSignals`,
`ForwardWithFilter`, `ForwardWithMap`, `MultiForward`.

Construction and seeding do **not**: `New`, `PipeTo`, `PutSignals`, `PutPayloads`. Their hooks
receive `context.Background()`.

**Metadata methods are gone from mutating types.** `FMesh`, `Component`, `Port`, the collections
and `cycle.Cycle` no longer carry `AddLabel`/`AddLabels`/`SetLabels`/`ClearLabels`/`RemoveLabels`
or their scalar equivalents (38 methods). Go through the store, which is where chaining lives now:

```go
fm.AddLabel("env", "prod")            // before
fm.Labels().Set("env", "prod")        // after

c.SetLabels(m)                        // before
c.Labels().Clear().SetMany(m)         // after

p.RemoveScalars("a", "b")             // before
p.Scalars().Remove("a", "b")          // after
```

Copy-on-write types (`signal.Signal`, `signal.Group`) keep `WithLabel`/`WithScalar`/`Without*` —
there they are the only way to produce a modified value. Constructor options
(`component.WithLabel`, `port.WithScalar`, …) are unchanged, and are the tidiest replacement when
you were labelling a freshly constructed object.

**Renames.**

| Before | After | Why |
|---|---|---|
| `port.Collection.Without(names)` | `Remove(names)` | it mutates |
| `component.Collection.Without(names)` | `Remove(names)` | it mutates |
| `port.Collection.WithParentComponent` | `SetParentComponent` | it mutates |
| `port.Collection.PutSignals` | `PutSignalsOnEach` | it broadcasts to every port |
| `port.Collection.PipeTo` | `PipeEachTo` | it builds a full cross product |
| `With{Label,Scalar}OnEach` on mutating collections | `Set{Label,Scalar}OnEach` | they mutate |
| `component.ErrWaitingForInputs` (as a return) | `component.ErrWaitDroppingInputs` | the dropping mode is now named |
| `component.ErrWaitingForInputsKeep` | `component.ErrWaitKeepingInputs` | symmetry with the above |

`component.ErrWaitingForInputs` still exists as the sentinel both modes wrap — `errors.Is(err,
ErrWaitingForInputs)` continues to answer "is this component waiting". Returning it directly still
means drop.

`signal.Group.WithLabelOnEach`/`WithScalarOnEach` are **unchanged**: that type is copy-on-write.

**`Signal.Payload()` no longer returns an error.**

```go
payload, err := sig.Payload()     // before
payload := sig.Payload()          // after
```

It could only fail for a zero-value `Signal{}` — a construction bug, not a runtime condition.
`nil` is a valid payload and reads as `nil`. Removed with it: `Signal.PayloadOrNil`,
`Signal.PayloadOrDefault`, `signal.ErrNoPayload`. `Group.AllPayloads()` likewise returns `[]any`
with no error.

`Group.FirstPayload()` **keeps** its error, along with `FirstPayloadOrDefault`/`FirstPayloadOrNil`:
an empty group is an ordinary runtime state, not a construction bug.

**Panics are a typed error.** A recovered panic is now `*component.PanicError` whose `Error()` is
one line. Code matching on the old `"panicked with: … stack: …"` message must change.

```go
var panicErr *component.PanicError
if errors.As(err, &panicErr) {
    log.Printf("%s: %v\n%s", panicErr.ComponentName, panicErr.Value, panicErr.StackTrace())
}
```

**`ErrRunCanceled`, not `ErrRunCancelled`** — US spelling, matching `context.Canceled` and the
repo's linter.

**Each bundled plugin has its own package.** `plugin` now holds no code; it groups
`plugin/profiler` and `plugin/autowire`. Every symbol lost the prefix it only carried to stay
distinct inside one shared package.

```go
import "github.com/hovsep/fmesh/plugin"                    // before
import "github.com/hovsep/fmesh/plugin/profiler"           // after
import "github.com/hovsep/fmesh/plugin/autowire"           // after
```

| Before | After |
|---|---|
| `plugin.NewProfiler()` | `profiler.New()` |
| `plugin.Profiler` | `profiler.Plugin` |
| `plugin.ProfileMode` | `profiler.Mode` |
| `plugin.ProfileTiming` / `ProfileThroughput` / `ProfileTimeline` / `ProfileRuntime` / `ProfileAll` | `profiler.ModeTiming` / `ModeThroughput` / `ModeTimeline` / `ModeRuntime` / `ModeAll` |
| `plugin.Stat`, `ComponentStat`, `Flow`, `PipeStat`, `CycleRecord`, `ResourceStat` | `profiler.Stat`, `ComponentStat`, … |
| `plugin.Autowire` | `autowire.Plugin` |
| `plugin.AutowireBroadcast(name)` | `autowire.Broadcast(name)` |
| `plugin.AutowireBroadcastAs(out, in)` | `autowire.BroadcastAs(out, in)` |
| `plugin.AutowirePrefixed(prefix)` | `autowire.Prefixed(prefix)` |

```go
fm, err := fmesh.New("mesh", fmesh.WithPlugins(   // before
    plugin.NewProfiler(),
    plugin.AutowireBroadcast("time"),
))

fm, err := fmesh.New("mesh", fmesh.WithPlugins(   // after
    profiler.New(),
    autowire.Broadcast("time"),
))
```

The plugin *names* are unchanged, so `PluginRegistered("profiler")` and
`PluginRegistered("autowire:broadcast:tick->time")` still answer the same.

### Added

- `fmesh.ErrRunCanceled`, wrapping `ctx.Err()` so `errors.Is(err, context.Canceled)` works.
- `fmesh.ErrLivelockDetected` plus `WithLivelockThreshold(n)` and `WithoutLivelockDetection()`.
  A mesh whose components all wait on each other now stops in 2 cycles with an error naming the
  stuck components and their empty input ports, instead of burning the cycle budget and reporting
  `reached max allowed cycles`.
- `port.Collection.AllOrdered()`, mirroring `component.Collection`.
- `cycle.Cycle.AllActivatedAreWaiting()`.
- `component.PanicError` with `StackTrace()` and `Unwrap()`.
- Wiki page [603. Caveats](https://github.com/hovsep/fmesh/wiki/603.-Caveats).
- **`port.Hooks.OnSignalsDelivered`** and `port.SignalsDeliveredContext` — the first event that
  names both ends of a pipe. It fires on the *source* port, once per destination, after the
  destination accepted the signals. `OnSignalsAdded` fires on the destination and cannot say who
  sent the batch, or tell a pipe delivery apart from a hand-written `PutSignals`.

  ```go
  out.SetupHooks(func(h *port.Hooks) {
      h.OnSignalsDelivered(func(ctx context.Context, d *port.SignalsDeliveredContext) error {
          fmt.Printf("%s -> %s carried %d\n",
              d.SourcePort.Name(), d.DestinationPort.Name(), len(d.SignalsDelivered))
          return nil
      })
  })
  ```

  Unlike `OnSignalsAdded`, which gates the put and rolls the port back on failure, this hook is an
  observer: the delivery has already happened, so a failure fails the flush without undoing it.
- **`profiler.Mode`** — the profiler now measures four dimensions, selected by bit flag.
  `profiler.New()` with no arguments measures `ModeTiming` and behaves exactly as the old
  `plugin.NewProfiler()` did. `ModeThroughput`, `ModeTimeline` and `ModeRuntime` are opt-in, and
  `ModeAll` enables everything. Each dimension has a real cost, which is why none of them is on by
  default — the same reason Go's own block and mutex profiles must be switched on deliberately.

  ```go
  prof := profiler.New()                                       // timing only
  prof := profiler.New(profiler.ModeAll)                       // everything
  prof := profiler.New(profiler.ModeTiming, profiler.ModeRuntime)
  ```

- `profiler.Plugin.Pipes()` and `TopNPipes(n)`, with `profiler.Flow` and `profiler.PipeStat` —
  per-pipe signal throughput under `ModeThroughput`. `Pipes()` is volume-sorted, so the hottest
  pipes are at the head and the coldest at the tail; a pipe that was wired but never used still
  appears, with zero transfers.
- `profiler.Plugin.Timeline()`, `SetTimelineLimit(n)` and `profiler.CycleRecord` — one record per
  cycle under `ModeTimeline`, for plotting any cycle-level stat against the cycle number.
- `profiler.Plugin.Resources()` and `profiler.ResourceStat` — Go runtime deltas (heap, allocations,
  GC, goroutines, CPU) under `ModeRuntime`, sampled through `runtime/metrics`. They are
  process-wide, not mesh-attributed.
- `profiler.Plugin.Report()` grew a pipe table, a resources block and a timeline summary — each
  printed only when its mode is enabled, so a timing-only profiler's report is unchanged.

### Changed

- **Runs are deterministic.** Every `port.Collection` traversal goes in port-name order, so a mesh
  with deterministic activation functions produces identical output for identical input. Previously
  a component reading `Inputs().Signals()` saw its ports in map order — 200 identical runs produced
  four different answers. Traversal is also ~90% faster and allocation-free.
- **`TimeLimit` is a real deadline.** `Run` derives `context.WithTimeout`, so the limit reaches the
  calls inside your activation functions rather than only being checked between cycles. A
  `WithTimeLimit(100ms)` mesh containing a 3s blocking call used to run for the full 3s.
- A caller-supplied deadline shorter than the mesh's own `TimeLimit` is reported as
  `ErrRunCanceled`, not `ErrTimeLimitExceeded`.

### Fixed

- Run errors no longer contain `%!w(<nil>)`. A cycle with activation errors but no panics used to
  end its message with `activation panics: %!w(<nil>)` — a nil passed to `%w`.
- Panic errors are no longer multi-kilobyte single-line strings with a stack trace formatted into
  the message.
- `Test_MultipleRun/runtime_info_duration_is_per_run` no longer flakes under `-race`: it compared a
  50ms sleep against a fixed 100ms ceiling.
- The panic cases in `TestComponent_MaybeActivate` were asserting nothing — the table only compared
  errors when `IsError()`, which is false for panics.

### Migration checklist

1. `fm.Run()` → `fm.Run(ctx)`.
2. Add `ctx context.Context` as the first parameter of every activation function and every hook
   (`_ context.Context` where unused).
3. Replace metadata calls with `.Labels()` / `.Scalars()` store calls; move labelling of freshly
   constructed objects into constructor options.
4. Apply the renames in the table above.
5. Drop the error return from `Payload()` and `AllPayloads()` call sites.
6. Replace `ErrWaitingForInputsKeep` with `ErrWaitKeepingInputs`; if you returned bare
   `ErrWaitingForInputs`, say `ErrWaitDroppingInputs` instead.
7. Run `go build ./...` — the compiler finds every remaining site.
8. Run your mesh tests with `-race` (see [603. Caveats](https://github.com/hovsep/fmesh/wiki/603.-Caveats)).

## Earlier releases

See the [releases page](https://github.com/hovsep/fmesh/releases).
