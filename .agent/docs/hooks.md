# Hooks & plugins — extension points

Source: `hooks.go`, `component/hooks.go`, `port/hooks.go`, `hook/hook_group.go`, `component/plugin.go`.

## The hook primitive

`hook.Group[T]` (package `hook`) — the **one approved generic type** in the codebase. An ordered
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

`component.Plugin` interface: `GetName() string` + `Init(*Component) error`. Registered via the
`component.WithPlugins(...)` constructor option; `Init` runs during `component.New` (after all
other options), duplicate names are a construction error. Query with `c.PluginRegistered(name)`.
A plugin is just an initialization bundle — typically registers hooks/ports on the component.
Order inside `New`: options → plugin `Init`s → `OnCreation` hooks.
