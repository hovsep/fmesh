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

**Labels and scalars are one metadata store.** `meta.Labels` and `meta.Scalars` are gone; every
entity carries a single `*meta.Meta` whose values are `string` or `float64`, chosen per entry.
Writes infer the type; reads name it (or infer it from a default). Go 1.27 generic methods make
this possible.

```go
// before
fm.Labels().Set("env", "prod")
fm.Scalars().Set("version", 2)
sig = sig.WithLabel("priority", "high").WithScalar("weight", 0.85)
w := sig.Scalars().ValueOrDefault("weight", 0)
v, err := c.Labels().Value("env")

// after
fm.Meta().Set("env", "prod").Set("version", 2.0)
sig = sig.WithMeta("priority", "high").WithMeta("weight", 0.85)
w := sig.Meta().ValueOrDefault("weight", 0.0)
v, err := c.Meta().Value[string]("env")
```

The full rename, on every type that had the pair:

| Before | After |
|---|---|
| `Labels()`, `Scalars()` | `Meta()` |
| `meta.NewLabels()`, `meta.NewScalars()` | `meta.New()` |
| `WithLabel`, `WithScalar` (signal, group, and the `fmesh`/`component`/`port` options) | `WithMeta` |
| `Signal.WithLabels`, `Signal.WithScalars` | `Signal.WithMetaMany` |
| `Signal.WithoutLabels`, `Signal.WithoutScalars` | `Signal.WithoutMeta` |
| `Group.WithLabelOnEach`, `Group.WithScalarOnEach` | `Group.WithMetaOnEach` |
| `Group.RemoveLabelOnEach`, `Group.RemoveScalarOnEach` | `Group.WithoutMetaOnEach` |
| `signal.HasLabel`, `signal.HasAllLabels` | `signal.HasMeta(keys...)` |
| `signal.HasAnyLabel`, `signal.LabelEquals`, `signal.LabelContains` | `signal.HasAnyMeta`, `signal.MetaEquals`, `signal.MetaContains` |
| `Labels.HasAll` | `Meta.Has(keys...)` |
| `Value(k)` | `Value[string](k)` / `Value[float64](k)` |

Two things the compiler now enforces. An untyped integer literal does not satisfy `meta.Value`:
`Set("n", 1)` and `ValueOrDefault("n", 0)` are compile errors — write `1.0` and `0.0`. And a key
holds one type at a time: `Value[float64]` on a string entry is an error naming both types, while
`ValueOrDefault`/`ValueIs` treat it as absent.

One thing the compiler does not catch: labels and scalars had separate key spaces, so a label and
a scalar could share a name. Now they are one key, and the later write wins —
`WithLabel("x", "a").WithScalar("x", 1)` migrates to `WithMeta("x", "a").WithMeta("x", 1.0)`, which
keeps only `1.0`. Rename one of them.

Removed with no replacement, for simplicity: `Signal.WithOnlyLabels`/`WithNoLabels` and the scalar
twins (`WithoutMeta(s.Meta().Keys()...)` is the rare case spelled out), `port.Group.SetLabel`/
`SetScalar` (use `g.Meta().Set`), and on the store `Values`, `Map`, `Merge`, `Every`, `Any`,
`Count`, `ForEach`, `HasAllFrom`, `HasAnyFrom` and the `Mapper`/`ScalarPredicate` types. `Filter`
stays, with a `func(key string, value any) bool` predicate; `Clone` is new.

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
`signal.ErrNoPayload`; `Signal.PayloadOrDefault` returns below as a typed generic method. `Group.AllPayloads()` likewise returns `[]any`
with no error.

`Group.FirstPayload()` **keeps** its error, along with `FirstPayloadOrNil`: an empty group is an
ordinary runtime state, not a construction bug. `FirstPayloadOrDefault` is now generic and
`FirstAs[T]()` is its error-returning sibling — see the next entry.

**Requires Go 1.27.** The typed accessors below are generic methods, a Go 1.27 language feature,
so consumers build with Go 1.27 or later.

**Typed payload accessors are methods on `*Signal`.** They were package functions because a
method could not have its own type parameter; now it can.

```go
n, err := signal.As[int](sig)              // before
n, err := sig.As[int]()                    // after

g, err := signal.AsGroup(sig)              // before
g, err := sig.AsGroup()                    // after

v, ok := signal.AsNumber(sig)              // before
v, ok := sig.AsNumber()                    // after
```

The defaulting form is renamed as well — `PayloadOrDefault` reads as a noun like the rest of the
family (`GetOrDefault`, `ValueOrDefault`, `FirstPayloadOrDefault`), where `AsOrDefault` did not.
All are safe on a nil `*Signal`, as before. The per-type shorthands are gone — a method with an
explicit type argument is as short as they were:

| Before | After |
|---|---|
| `signal.AsOrDefault(sig, d)` | `sig.PayloadOrDefault(d)` |
| `signal.AsFloat64OrDefault(sig, d)` | `sig.Float64OrDefault(d)` |
| `signal.AsInt(sig)` | `sig.As[int]()` |
| `signal.AsString(sig)` | `sig.As[string]()` |
| `signal.AsBool(sig)` | `sig.As[bool]()` |
| `signal.AsFloat64(sig)` | `sig.As[float64]()` |

The common read straight from a port has typed forms on the group, so `First()` need not be
spelled out:

```go
n, err := signal.As[int](this.InputByName("in").Signals().First())   // before
n, err := this.InputByName("in").Signals().FirstAs[int]()           // after; ErrNoSignalsInGroup when empty

n := this.InputByName("in").Signals().FirstPayloadOrDefault(0)       // after, was signal.AsOrDefault(...First(), 0)
```

`Group.FirstPayloadOrDefault` itself is generic now — `T` is inferred from the default, so the
assertion on its result goes away, and a nil or wrong-type payload yields the default. For the old
any-in/any-out behaviour spell the type argument: `FirstPayloadOrDefault[any](x)`.

```go
n := this.InputByName("in").Signals().FirstPayloadOrDefault(0).(int)   // before
n := this.InputByName("in").Signals().FirstPayloadOrDefault(0)         // after
```

**`Group.ReducePayloads` is generic in the accumulator.** No more casting `acc` on every step.

```go
s := g.ReducePayloads("", func(acc, payload any) any { return acc.(string) + payload.(string) }).(string) // before
s := g.ReducePayloads("", func(acc string, payload any) string { return acc + payload.(string) })         // after
```

`signal.PayloadReducer` is removed with it; `Reducer` (over signals) is unchanged.

**`Group.ContainsPayload` takes a comparable type parameter and returns only `bool`.** Passing a
slice or map no longer compiles, which is what the runtime error used to tell you;
`signal.ErrPayloadNotComparable` is removed. `ContainsPayloadFunc` remains for those types.

```go
found, err := g.ContainsPayload(v)    // before
found := g.ContainsPayload(v)         // after
found := g.ContainsPayload[any](nil)  // after: a nil payload needs the type argument spelled out
```

**`component.MustGetTyped` is replaced by the `State.GetTyped[T]` method**, which returns an error
instead of panicking — the same shape as every other accessor.

| Before | After |
|---|---|
| `cur := component.MustGetTyped[string](this.State(), "current")` | `cur, err := this.State().GetTyped[string]("current")` |

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
| `plugin.ProfileTiming` / `ProfileThroughput` / `ProfileTimeline` / `ProfileAll` | `profiler.ModeTiming` / `ModeThroughput` / `ModeTimeline` / `ModeAll` |
| `plugin.ProfileRuntime`, `plugin.ResourceStat` | removed — see the mesh-centric entry below |
| `plugin.Stat`, `ComponentStat`, `Flow`, `PipeStat`, `CycleRecord` | `profiler.Stat`, `ComponentStat`, … |
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

**The profiler is mesh-centric: the runtime-resources dimension is gone.** `ProfileRuntime`
(latterly `profiler.ModeRuntime`), `ResourceStat`, `Resources()` and `CycleRecord.Resources` are
removed. They sampled Go runtime counters — heap, GC, goroutines, CPU — which are process-wide by
nature: every number included whatever else the process was doing, and the report had to carry a
"not mesh-attributed" disclaimer. Everything the profiler now reports is attributable to the mesh.
`profiler.ModeAll` is `ModeTiming | ModeThroughput | ModeTimeline`.

**Plugins implement `Name()`, not `GetName()`.** Both plugin interfaces (`fmesh.Plugin`,
`component.Plugin`) now follow Go's getter convention. `autowire.Plugin`'s convention field is
renamed `Name` → `InputNameFor` to make room (and because it names what the function answers:
the input-port name an output should be wired to).

```go
func (p *MyPlugin) GetName() string { return "my-plugin" }   // before
func (p *MyPlugin) Name() string { return "my-plugin" }      // after

&autowire.Plugin{Name: func(...) string {...}}               // before
&autowire.Plugin{InputNameFor: func(...) string {...}}       // after
```

**`port.Group`'s own-metadata setters are `Set*`, because they mutate.** The `With` prefix is
reserved for copy-on-write and constructor options; `port.Group` was the one type violating that.

```go
group.WithLabel("k", "v").WithScalar("s", 1)   // before
group.SetLabel("k", "v").SetScalar("s", 1)     // after
```

**`ActivationResultCollection.Without` is now `Remove`.** It deletes from the receiver in place;
`Without` on the other collection types returns a new value.

**`signal.Group.Labels()` and `Scalars()` return copies.** `signal.Group` is copy-on-write, and
these were the one back door: mutating the returned store changed the group in place. They now
match `Signal.Labels()`; mutate via `WithLabel`/`WithScalar`, which return a new group.

```go
g.Labels().Set("k", "v")   // before: mutated g. Now: mutates a copy, g unchanged
g = g.WithLabel("k", "v")  // after
```

**Removed API.** Each of these had no callers in the repo, the examples, or the exporter:

- Batch metadata on the mutating collections — `SetLabelOnEach`, `SetScalarOnEach`,
  `RemoveLabelOnEach`, `RemoveScalarOnEach` on `port.Group`, `port.Collection`,
  `component.Collection` and `cycle.Group`. Iterate with `ForEach` and use each element's own
  store. The copy-on-write batch methods on `signal.Group` (`WithLabelOnEach` …) stay.
- Scalar statistics — `meta.Scalars.Min`/`Max`/`Sum`/`Average`/`Scale`, and `signal.Group`'s
  cross-signal aggregation `SumScalar`/`MinScalar`/`MaxScalar`/`AvgScalar` with its sentinel
  `signal.ErrScalarNotFoundInGroup`. Scalars are a metadata store, not a statistics library;
  aggregate with `ForEach` where needed.
- `port.NewIndexedInputGroup` / `NewIndexedOutputGroup` — indexed ports are created on components
  via `component.WithIndexedInputs` / `WithIndexedOutputs`, which never went through these.

### Added

- `signal.Group.FirstAs[T]()` — the first payload as `T`, with an empty group and a wrong type
  handled like `Signal.As`.
- The profiler labels every activation goroutine with `fmesh.mesh` and `fmesh.component`
  (`runtime/pprof` labels), so a CPU profile can be focused on one component
  (`go tool pprof -tagfocus=fmesh.component=NAME`) and, on Go 1.27, the stack a
  `component.PanicError` keeps names the component in its header.
- `fmesh.ErrRunCanceled`, wrapping `ctx.Err()` so `errors.Is(err, context.Canceled)` works.
- `fmesh.ErrLivelockDetected` plus `WithLivelockThreshold(n)` and `WithoutLivelockDetection()`.
  A mesh whose components all wait on each other now stops in 2 cycles with an error naming the
  stuck components and their empty input ports, instead of burning the cycle budget and reporting
  `reached max allowed cycles`.
- `port.Collection.AllOrdered()`, mirroring `component.Collection`.
- `component.State.UpdateAndGet(key, fn)` — `Upsert` plus the value it stored, for when the new
  value is needed in the same activation. Like `Upsert` and unlike `Update`, it creates the key
  when absent, so the returned value is always the one now in the state.

  ```go
  n := this.State().UpdateAndGet("seen", func(old any) any {
      if old == nil {
          return 1
      }
      return old.(int) + 1
  })
  return this.OutputByName("count").PutPayloads(n)
  ```
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

- **Go 1.27 is the minimum.** Generic methods (the typed accessors above) need it. The scheduler
  now spawns activation goroutines with `sync.WaitGroup.Go`.
- **Runs are deterministic.** Every `port.Collection` traversal goes in port-name order, so a mesh
  with deterministic activation functions produces identical output for identical input. Previously
  a component reading `Inputs().Signals()` saw its ports in map order — 200 identical runs produced
  four different answers. Traversal is also ~90% faster and allocation-free.
- **`TimeLimit` is a real deadline.** `Run` derives `context.WithTimeout`, so the limit reaches the
  calls inside your activation functions rather than only being checked between cycles. A
  `WithTimeLimit(100ms)` mesh containing a 3s blocking call used to run for the full 3s.
- A caller-supplied deadline shorter than the mesh's own `TimeLimit` is reported as
  `ErrRunCanceled`, not `ErrTimeLimitExceeded`.
- **A port-name typo says so.** `InputByName`, `OutputByName` and `Collection.ByName` still return
  `nil` for a name no port has — that is deliberate. What changed is what happens next: `Signals`,
  `HasSignals`, `PutSignals`, `PutPayloads`, `PutSignalGroups`, `Flush` and `Pipes` now panic with
  `port.ErrNilPort` instead of dereferencing the nil several frames later.

  ```
  panicked: invalid memory address or nil pointer dereference          // before
  panicked: port is nil: InputByName, OutputByName and Collection.ByName
            return nil for a name no port has                          // after
  ```

  Inside an activation the panic is recovered as before, so the run error names the component too,
  and `errors.Is(err, port.ErrNilPort)` reaches it. `PipeTo` is unchanged: `validatePipe` already
  returned `ErrNilPort` for a nil source, and it still returns rather than panics.
- **An empty mesh fails validation, not the first cycle.** `Run` on a mesh with no components now
  returns the sentinel `fmesh.ErrNoComponents` from structure validation, before any cycle runs.
  Previously it returned an unmatchable ad-hoc error after recording a phantom cycle in
  `RuntimeInfo`. Since validation runs first, user `BeforeRun` hooks no longer fire for an empty
  mesh.
- **`CyclesLimit` is the number of cycles that execute.** A limit of 1000 used to run 1001 cycles
  before stopping. A mesh whose last allowed cycle was already empty now stops naturally instead of
  reporting `ErrReachedMaxAllowedCycles`.
- Drain failures wrap inline (`failed to drain: ...`) instead of `errors.Join`'s newline format, so
  they print like every other stop reason. `errors.Is(err, ErrFailedToDrain)` still matches.
- Under `StopOnFirstPanic`, the stop error is built by the same helper as
  `StopOnFirstErrorOrPanic`, so it now also names any activation errors from the failing cycle.
- Debug-mode activation logging moved from the run loop into a default `AfterCycle` hook — same
  output, but the scheduler no longer carries logging code.
- **`component.Collection` traversal is deterministic.** `Any`, `FindAny`, `Every`, `AnyMatch`,
  `Count`, `Filter`, `Map` and `ForEach` now go in component-name order, matching
  `port.Collection`; `Any`/`FindAny` no longer return different components on identical calls.
  `AllOrdered` reads a cached sorted slice instead of sorting per call.

### Fixed

- A failed run's error lists the failing components in name order. `Cycle.AllErrorsCombined` /
  `AllPanicsCombined` joined them in map order, so the same failure printed differently on
  different runs. `ActivationResultCollection.AllOrdered()` is new and is what they iterate.
- `profiler.Reset()` called during a run no longer folds a garbage duration into the run stat:
  it clears the in-flight start times, and an end without a start is ignored.
- `autowire` reports a missing `InputNameFor` by that name, not as "Name".
- Run errors no longer contain `%!w(<nil>)`. A cycle with activation errors but no panics used to
  end its message with `activation panics: %!w(<nil>)` — a nil passed to `%w`.
- Panic errors are no longer multi-kilobyte single-line strings with a stack trace formatted into
  the message.
- `Test_MultipleRun/runtime_info_duration_is_per_run` no longer flakes under `-race`: it compared a
  50ms sleep against a fixed 100ms ceiling.
- The panic cases in `TestComponent_MaybeActivate` were asserting nothing — the table only compared
  errors when `IsError()`, which is false for panics.
- `cycle.Group.Filter`/`Map`/`MapIf` no longer drop the group's own length limit, labels and
  scalars from the returned group.
- `signal.Group.Join(nil)` no longer panics; a nil group joins as empty, matching `With`'s
  treatment of nil signals.
- The no-hook fast path no longer allocates: the hook-context structs for `OnSignalsAdded`,
  `OnClear` and the component activation hooks are only built when a hook is registered, saving
  four heap allocations per component per cycle.

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
7. Rename `GetName()` → `Name()` on your plugins; `autowire.Plugin{Name: ...}` → `InputNameFor`.
8. Drop any use of the profiler's runtime-resources dimension (`ProfileRuntime`/`ModeRuntime`,
   `ResourceStat`, `Resources()`); reach for `runtime/metrics` directly if you need process
   numbers.
9. If you mutated a `signal.Group`'s own metadata through `Labels()`/`Scalars()`, switch to
   `WithLabel`/`WithScalar` and assign the result.
10. Set `go 1.27` in your `go.mod`.
11. Turn `signal.As[T](sig)` / `signal.AsOrDefault(sig, d)` / `AsFloat64OrDefault` / `AsGroup` /
    `AsNumber` into method calls on the signal (`sig.As[T]()`, `sig.PayloadOrDefault(d)`,
    `sig.Float64OrDefault(d)`, …); replace `AsInt`/`AsString`/`AsBool`/`AsFloat64` with
    `sig.As[T]()`. Reads straight from a port can use `Signals().FirstAs[T]()` /
    `FirstPayloadOrDefault(d)`; drop the `.(T)` after `FirstPayloadOrDefault`.
12. Give `ReducePayloads` a typed accumulator and drop the `.(T)` on the result; drop the error
    from `ContainsPayload` (and the type argument `[any]` if you pass `nil`).
13. `component.MustGetTyped[T](state, k)` → `v, err := state.GetTyped[T](k)`.
14. Run `go build ./...` — the compiler finds every remaining site.
15. Run your mesh tests with `-race` (see [603. Caveats](https://github.com/hovsep/fmesh/wiki/603.-Caveats)).

## Earlier releases

See the [releases page](https://github.com/hovsep/fmesh/releases).
