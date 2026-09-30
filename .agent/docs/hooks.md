# Hooks & plugins — extension points

Source: `hooks.go`, `component/hooks.go`, `port/hooks.go`, `internal/hook/hook_group.go`,
`plugin.go`, `component/plugin.go`, `internal/plugin/registry.go`, `plugin/`.

## The hook primitive

`hook.Group[T]` (package `internal/hook` — not part of the public API). Generics are confined to
`internal/` containers — `hook.Group[T]` and `plugin.Registry[T]` are **the only two approved
generic types** in the codebase; nothing public is generic. An ordered
slice of `func(context.Context, T) error`; `Trigger(ctx, arg)` runs all in insertion order,
**fail-fast** on first error.
Registration is chainable and happens through `SetupHooks(func(*Hooks))` closures (or the
`WithHooks` constructor option on components) — the `Hooks` structs' fields are unexported, so
closures are the only registration path.

## Context is the first argument, never a struct field

Every hook takes `context.Context` first, exactly like an `ActivationFunc`. The per-event context
structs (`*CycleContext`, `*ActivationContext`, the port ones) carry **data about the event**, never
the context itself — a context stored in a struct is the `containedctx` anti-pattern, and these
structs outlive no run anyway.

Which context a hook receives depends on the path that fired it:

| Fired from | Context |
|---|---|
| Anything inside `Run` — cycles, activations, drain | the run context (with the `TimeLimit` deadline applied) |
| `component.New`, `fm.AddComponents`, `PipeTo` | `context.Background()` — there is no run to cancel yet |
| `Port.PutSignals` / `PutPayloads` | `context.Background()` — seeding is a construction-time act, and requiring a context there would put `context.Background()` in every setup line and every activation function that writes an output |

That last row is the one deliberate seam: `OnSignalsAdded` sees the run context when signals arrive
through a pipe during drain, and `context.Background()` when a caller puts them directly.

## Three hook levels

| Level | Registration | Hooks | Context type |
|---|---|---|---|
| Mesh | `fm.SetupHooks(...)` | `OnComponentAdded`, `BeforeRun`, `AfterRun`, `BeforeCycle`, `AfterCycle` | `*FMesh` / `*CycleContext` / `*ComponentAddedContext` |
| Component | `component.WithHooks(...)` option or `c.SetupHooks(...)` | `OnCreation`, `BeforeActivation`, `AfterActivation` | `*Component` / `*ActivationContext` |
| Port | `p.SetupHooks(...)` | `OnSignalsAdded`, `OnSignalsDelivered`, `OnClear`, `OnInboundPipe`, `OnOutboundPipe` | per-event context structs |

## Semantics worth knowing

- **Only two activation hooks, each fired exactly once**: `BeforeActivation`, then the
  activation function, then `AfterActivation`. There are no per-outcome hooks — they were
  removed because a panicking outcome hook fired a second outcome hook (`OnPanic`) and re-ran
  `AfterActivation`. `AfterActivation` sees every outcome and tells them apart by
  `ctx.Result.Code()`.
- **`AfterActivation` always runs** — success, error, panic, waiting, or a failed
  `BeforeActivation`; a `finally` block.
- **Hook panics are recovered.** Hooks run on the activation goroutine, where an escaped panic
  would kill the process. Each stage (`BeforeActivation`, the function, `AfterActivation`)
  recovers its own panic, so no stage re-runs; a panicking hook re-codes the result to
  `ActivationCodePanicked` (so `StopOnFirstPanic` stops on it) with a `*component.PanicError`.
- Behavior added to a component from **outside** (a plugin) goes through these two hooks;
  `component/compose.go` combinators are for a component composing its own activation function.
- A **failing hook poisons the result**: the activation result is re-coded to
  `ActivationCodeHookFailed` with the hook error attached. `HookFailed` results count as
  activation errors (`IsError()`), so under `StopOnFirstErrorOrPanic` the mesh stops and the
  hook error surfaces in `Run()`'s return. A failing `beforeCycle`/`afterCycle` hook aborts the
  run (the error is wrapped inline: `failed to run cycle: beforeCycle hook failed: …`); a
  failing `onComponentAdded` hook fails `AddComponents`.
- The mesh registers **two default hooks**: a `BeforeRun` hook that validates mesh structure on
  every run (empty mesh → `ErrNoComponents`; see runtime.md), and an `AfterCycle` hook
  (`logActivationResultsInDebug`) that logs each cycle's activation results — including panic
  stack traces — when debug mode is on. Debug logging lives there deliberately, so the scheduler
  carries no logging code. Don't clear either group without re-adding the equivalent.
- Port hooks fire on every `PutSignals`/`PutPayloads`/`Clear`, including the scheduler's own
  drain-phase forwarding and clearing — keep them cheap and side-effect-aware. They may fire
  from concurrent activation goroutines, so they must be concurrency-safe when touching shared
  state. When an `OnSignalsAdded` hook fails, the port rolls back to its previous signals.

## Plugins

Two levels, same shape: `Name() string` + `Init(T) error`, registered via a `WithPlugins(...)`
constructor option, duplicate names are a construction error, queried with `PluginRegistered(name)`.
A plugin is just an initialization bundle — typically registers hooks.

| Level | Interface | Registration | `Init` receives | Query |
|---|---|---|---|---|
| Component | `component.Plugin` | `component.WithPlugins(...)` | `*Component` | `c.PluginRegistered(name)` |
| Mesh | `fmesh.Plugin` | `fmesh.WithPlugins(...)` | `*FMesh` | `fm.PluginRegistered(name)` |

Storage, the duplicate-name check and the init order are shared: both levels hold an
`internal/plugin.Registry[T]` and call its `InitAll`. Only the public interfaces and the
`WithPlugins` options (which must be typed as each level's `Option`) live at the level itself — so
the two cannot drift, and a third level would not be a third copy.

- **Both levels `Init` in sorted name order**, not map order — plugins register hooks, hooks fire in
  registration order, so init order is observable and must not vary between runs of the same binary.
- **Component**: `Init` runs during `component.New` after all other options. It may modify the
  component's ports and metadata as well as its hooks.
  Order inside `New`: options → plugin `Init`s → `OnCreation` hooks.
- **Mesh**: `Init` runs during `fmesh.New` after the options *and* after `runtimeInfo` is built, so
  a plugin sees a fully configured mesh.

**A mesh plugin cannot walk components in `Init`** — a mesh is constructed empty and filled by
`AddComponents` afterwards, so there is nothing to walk yet. To instrument components, register an
`OnComponentAdded` hook and attach component hooks to each one as it arrives. That is how a mesh
plugin observes every activation without any component knowing it exists, and it is the intended
pattern for profiling, tracing, and assertion plugins.

Two consequences of the arrival-hook pattern, both learned the hard way in `plugin/`:

- A component is only ever looked at **on arrival**, so ports added afterwards
  (`AddInputs`/`AddOutputs`) are invisible to a plugin that wires or inspects ports.
- Instrumentation hooks fire from the **concurrent activation goroutines**. Guard shared state, and
  take clock readings *outside* the plugin's own lock — timing from inside the critical section
  charges each component for the contention caused by all the others.

## Bundled plugins (`plugin/`)

The packages under `plugin/` ship mesh-level plugins built on exactly the pattern above. Each plugin
is its own package (`plugin` itself holds only a doc comment); they import `fmesh`, so `fmesh` can
never import them.

| Plugin | What it does |
|---|---|
| `plugin/autowire` — `autowire.Prefixed(prefix)` / `Broadcast(name)` / `BroadcastAs(out, in)` / `&autowire.Plugin{InputNameFor: ...}` | Pipes ports by naming convention, in both directions on every arrival, so `AddComponents` order does not matter. Each convention is a separate plugin instance with its own `PluginName`. |
| `plugin/profiler` — `profiler.New(modes...)` | Mesh-centric measurement: run/cycle/activation timing, per-pipe throughput, per-cycle timeline. Every number is mesh-attributable — it deliberately measures nothing process-wide. |
