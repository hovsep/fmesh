# Hooks & plugins — extension points

Source: `hooks.go`, `component/hooks.go`, `port/hooks.go`, `internal/hook/hook_group.go`,
`plugin.go`, `component/plugin.go`, `internal/plugin/registry.go`, `plugin/`.

## The hook primitive

`hook.Group[T]` (package `internal/hook` — not part of the public API). Generics are confined to
`internal/` containers — `hook.Group[T]` and `plugin.Registry[T]` are **the only two approved
generic types** in the codebase; nothing public is generic. An ordered
slice of `func(T) error`; `Trigger(arg)` runs all in insertion order, **fail-fast** on first error.
Registration is chainable and happens through `SetupHooks(func(*Hooks))` closures (or the
`WithHooks` constructor option on components) — the `Hooks` structs' fields are unexported, so
closures are the only registration path.

## Three hook levels

| Level | Registration | Hooks | Context type |
|---|---|---|---|
| Mesh | `fm.SetupHooks(...)` | `OnComponentAdded`, `BeforeRun`, `AfterRun`, `BeforeCycle`, `AfterCycle` | `*FMesh` / `*CycleContext` / `*ComponentAddedContext` |
| Component | `component.WithHooks(...)` option or `c.SetupHooks(...)` | `OnCreation`, `BeforeActivation`, `OnActivation`, `OnSuccess`, `OnError`, `OnPanic`, `OnWaitingForInputs`, `AfterActivation` | `*Component` / `*ActivationContext` |
| Port | `port.Hooks` via port options | `OnSignalsAdded`, `OnClear`, `OnInboundPipe`, `OnOutboundPipe` | per-event context structs |

## Semantics worth knowing

- **`OnActivation` is special**: its hooks are `ActivationFunc`s appended after the main
  activation function and run **sequentially in the same activation** — they share the error
  path (first error aborts the chain and becomes the activation error).
- **`OnActivation` hooks vs activation combinators**: `component/compose.go` also chains
  `ActivationFunc`s, but those are combinators — plain values passed to `WithActivationFunc`,
  with no registry, no name, and no initialization step. Use a hook when something *outside*
  the component adds behavior to it; use a combinator when the component composes its own.
- **`AfterActivation` always runs** — success, error, panic, or waiting; a `finally` block.
- Outcome hooks (`OnSuccess`/`OnError`/`OnPanic`/`OnWaitingForInputs`) fire before
  `AfterActivation`. Distinguish waiting modes via `ctx.Result.Code()`
  (`WaitingForInputsClear` vs `WaitingForInputsKeep`).
- A **failing hook poisons the result**: the activation result is re-coded to
  `ActivationCodeHookFailed` with the hook error attached. `HookFailed` results count as
  activation errors (`IsError()`), so under `StopOnFirstErrorOrPanic` the mesh stops and the
  hook error surfaces in `Run()`'s return. A failing `beforeCycle`/`afterCycle`
  hook aborts the run (`errFailedToRunCycle`); a failing `onComponentAdded` hook fails
  `AddComponents`.
- The mesh's **default `BeforeRun` hook validates mesh structure** on every run (see
  runtime.md). Don't clear the beforeRun group without re-adding equivalent validation.
- Port hooks fire on every `PutSignals`/`PutPayloads`/`Clear`, including the scheduler's own
  drain-phase forwarding and clearing — keep them cheap and side-effect-aware. They may fire
  from concurrent activation goroutines, so they must be concurrency-safe when touching shared
  state. When an `OnSignalsAdded` hook fails, the port rolls back to its previous signals.

## Plugins

Two levels, same shape: `GetName() string` + `Init(T) error`, registered via a `WithPlugins(...)`
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

Package `plugin` ships mesh-level plugins built on exactly the pattern above. It imports `fmesh`, so
`fmesh` can never import it.

| Plugin | What it does |
|---|---|
| `plugin.NewProfiler()` | Times whole runs, single cycles, and each component's activations. `Runs()`, `Cycles()`, `Components()`, `TopN(n)`, `Report()`, `Reset()`. One instance belongs to one mesh. |
| `plugin.AutowirePrefixed(prefix)` / `AutowireBroadcast(name)` / `AutowireBroadcastAs(out, in)` / `&Autowire{Name: ...}` | Pipes ports by naming convention, in both directions on every arrival, so `AddComponents` order does not matter. Each convention is a separate plugin instance with its own `PluginName`. |
