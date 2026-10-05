# Hooks & plugins — extension points

Source: `hooks.go`, `component/hooks.go`, `component/activation.go`, `port/hooks.go`,
`internal/hook/hook_group.go`, `plugin.go`, `component/plugin.go`, `internal/plugin/registry.go`,
`plugin/`.

## The hook primitive

`hook.Group[T]` in `internal/hook` (not public API) is an ordered slice of
`func(context.Context, T) error`. `Trigger(ctx, arg)` runs them in insertion order and is
**fail-fast** on the first error.

Registration is chainable and goes through `SetupHooks(func(*Hooks))` closures (or the
`component.WithHooks` option). The `Hooks` structs' fields are unexported, so closures are the only
registration path.

## Context is the first argument, never a struct field

Every hook takes `context.Context` first, like an `ActivationFunc`. The event structs
(`*CycleContext`, `*ActivationContext`, the port ones) carry **data about the event**, never the
context — a context in a struct is the `containedctx` anti-pattern.

Which context a hook gets depends on what fired it:

| Fired from | Context |
|---|---|
| Anything inside `Run` — cycles, activations, drain | the run context (with the `TimeLimit` deadline) |
| `Run`'s cleanup of the previous run's outputs (`OnClear`) | the caller's context — the `TimeLimit` deadline starts after it |
| `component.New`, `fm.AddComponents`, `PipeTo` | `context.Background()` — no run to cancel yet |
| `Port.PutSignals` / `PutPayloads` | `context.Background()` — seeding is a construction-time act; a context there would put `context.Background()` in every setup line and every activation function that writes an output |

The last row is the one deliberate seam: `OnSignalsAdded` sees the run context when signals arrive
through a pipe during the drain, and `context.Background()` when a caller puts them directly.

## Three hook levels

| Level | Registration | Hooks | Argument type |
|---|---|---|---|
| Mesh | `fm.SetupHooks(...)` | `OnComponentAdded`, `BeforeRun`, `AfterRun`, `BeforeCycle`, `AfterCycle` | `*ComponentAddedContext` / `*FMesh` / `*CycleContext` |
| Component | `component.WithHooks(...)` or `c.SetupHooks(...)` | `OnCreation`, `BeforeActivation`, `AfterActivation` | `*Component` (first two) / `*ActivationContext` |
| Port | `p.SetupHooks(...)` | `OnSignalsAdded`, `OnSignalsDelivered`, `OnClear`, `OnInboundPipe`, `OnOutboundPipe` | per-event context structs |

## Activation hooks

- **Only two, each fired exactly once**: `BeforeActivation`, the activation function, then
  `AfterActivation`. There are no per-outcome hooks: `AfterActivation` sees every outcome and tells
  them apart by `ac.Result.Code()` (`ac` is the `*component.ActivationContext`). Do not add
  per-outcome hooks back — a second hook firing on one outcome makes hooks run twice.
- **`AfterActivation` always runs** — on success, error, panic, waiting, or a failed
  `BeforeActivation`. Treat it as a `finally` block.
- **`WithRetry` does not repeat activation hooks.** All attempts are one activation:
  `BeforeActivation` and `AfterActivation` fire once around them. The output reset between
  attempts goes through `Port.Clear` / `PutSignalGroups` on purpose, so the port's `OnClear` and
  `OnSignalsAdded` hooks do fire for each changed output port. That is documented, not a bug.
- **Hook panics are recovered.** Activation hooks run on the activation goroutine, where an
  escaped panic would kill the process. Each stage (`BeforeActivation`, the function, `AfterActivation`)
  recovers its own panic, so no stage re-runs. A panicking hook re-codes the result to
  `ActivationCodePanicked` with a `*component.PanicError` (so `StopOnFirstPanic` stops on it).
- **A failing hook poisons the result**: it is re-coded to `ActivationCodeHookFailed` with the hook
  error attached. A `Panicked` result is never re-coded to `HookFailed`. `HookFailed` counts as an
  activation error (`IsError()`), so under `StopOnFirstErrorOrPanic` the mesh stops and the error
  surfaces from `Run`.
- A failing `BeforeActivation` skips the activation function.
- Behavior added to a component from **outside** (e.g. a plugin) goes through these two hooks.
  `component/compose.go` combinators are for a component composing its own activation function.

## Other semantics

- A failing `BeforeCycle`/`AfterCycle` hook aborts the run (wrapped inline:
  `failed to run cycle: beforeCycle hook failed: …`). A failing `BeforeCycle` skips that cycle's
  `AfterCycle`; `AfterRun` still runs. A failing `OnComponentAdded` hook fails `AddComponents`
  but leaves every component added, and the components after it never reach the hook.
- **Only the activation goroutine recovers panics.** Mesh hooks, and port hooks fired from the
  drain or from a caller's own `PutSignals`/`PipeTo`/`Clear`, run on the caller's goroutine and
  their panics propagate out of `Run` (or `AddComponents`, `PipeTo`, ...). The same port hook
  therefore panics out of `Run` from the drain but becomes a `Panicked` result when the component
  itself puts the signals during activation.
- The mesh registers **two default hooks**. Do not clear either group without re-adding the
  equivalent:
  - a `BeforeRun` hook that validates mesh structure on every run (empty mesh →
    `ErrNoComponents`; see [runtime.md](runtime.md));
  - an `AfterCycle` hook (`logActivationResultsInDebug`) that logs each cycle's results, including
    panic stacks, when debug is on. Debug logging lives there so the scheduler has no logging code.
- Port hooks fire on every `PutSignals`/`PutPayloads`/`Clear`, including the scheduler's own drain
  forwarding and clearing. Keep them cheap and side-effect-aware. They may fire from concurrent
  activation goroutines, so they must be concurrency-safe around shared state. A failing
  `OnSignalsAdded` hook rolls the port back to its previous signals.

## Plugins

Two levels, same shape: `Name() string` + `Init(T) error`, registered with a `WithPlugins(...)`
option, queried with `PluginRegistered(name)`. A duplicate name is a construction error. A plugin is
just an initialization bundle — usually it registers hooks.

**Plugins are for behavior, not views.** Something that only reads a mesh (an exporter, a report)
is a plain value, not a plugin: exporters implement fmesh-export's `export.Exporter` (`Export(fm)`,
`ExportCycle(fm, c)`) and get cycles from `RuntimeInfo.Cycles` or an `AfterCycle` hook. Make
something a plugin only when it must take part in the run.

| Level | Interface | Registration | `Init` receives | Query |
|---|---|---|---|---|
| Component | `component.Plugin` | `component.WithPlugins(...)` | `*Component` | `c.PluginRegistered(name)` |
| Mesh | `fmesh.Plugin` | `fmesh.WithPlugins(...)` | `*FMesh` | `fm.PluginRegistered(name)` |

- Storage, the duplicate check and init order are shared: both levels hold an
  `internal/plugin.Registry[T]` and call its `InitAll`. Only the public interfaces and the
  `WithPlugins` options (typed as each level's `Option`) live at the level itself, so the two cannot
  drift.
- **Both levels `Init` in sorted name order**, not map order. Plugins register hooks and hooks fire
  in registration order, so init order is observable and must not vary between runs.
- **Component**: `Init` runs inside `component.New` after all other options, and may change the
  component's ports and metadata as well as its hooks. Order in `New`: options → plugin `Init`s →
  `OnCreation` hooks.
- **Mesh**: `Init` runs inside `fmesh.New` after the options *and* after `runtimeInfo` is built, so
  a plugin sees a fully configured mesh.

**A mesh plugin cannot walk components in `Init`** — a mesh starts empty and is filled by
`AddComponents` later. To instrument components, register an `OnComponentAdded` hook and attach
component hooks to each one as it arrives. This is the intended pattern for profiling, tracing and
assertion plugins. Two consequences:

- A component is seen only **on arrival**, so ports added later (`AddInputs`/`AddOutputs`) are
  invisible to a plugin that wires or inspects ports.
- Instrumentation hooks fire from **concurrent activation goroutines**. Guard shared state, and take
  clock readings *outside* the plugin's own lock — timing inside the critical section charges each
  component for everyone else's contention.

## Bundled plugins (`plugin/`)

Mesh-level plugins built on the pattern above. Each is its own package (`plugin` itself holds only a
doc comment). They import `fmesh`, so `fmesh` can never import them.

| Plugin | What it does |
|---|---|
| `plugin/autowire` — `autowire.Prefixed(prefix)` / `Broadcast(name)` / `BroadcastAs(out, in)` / `&autowire.Plugin{InputNameFor: ...}` | Pipes ports by naming convention, in both directions on every arrival, so `AddComponents` order does not matter. Each convention is a separate plugin instance with its own `PluginName`. |
| `plugin/profiler` — `profiler.New(modes...)` | Mesh-centric measurement: run/cycle/activation timing, per-pipe throughput, per-cycle timeline. Every number is mesh-attributable; it measures nothing process-wide. |
