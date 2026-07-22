# Runtime — execution model

How a mesh actually runs. Source: `fmesh.go` (`Run`, `runCycle`, `drainComponents`, `mustStop`),
`component/activation.go`, `component/activation_result.go`.

## Run loop

`FMesh.Run()`:

1. `cleanUpPreviousRun` — clears all output ports (prevents signal accumulation between runs), resets `RuntimeInfo`. A mesh is re-runnable.
2. `beforeRun` hooks — includes a **default hook that validates mesh structure** (parent-mesh/parent-component wiring, pipe destinations belong to the same mesh). Validation runs on **every** `Run`, in component-name order (deterministic errors).
3. Loop: `runCycle` → `mustStop` → `drainComponents`. Note the order — stop conditions are checked **before** draining, so the final cycle's outputs are not flushed.
4. `afterRun` hooks fire in a defer; an afterRun hook error is only surfaced when the run itself did not already fail.

`Run` returns `(*RuntimeInfo, error)`. `RuntimeInfo.Cycles` holds every executed cycle with all activation results — the primary observability surface. Note this history is retained for the whole run and grows without bound — see "Scaling characteristics" below.

## One cycle (`runCycle`)

- Every component gets `MaybeActivate()` called in **its own goroutine**; the cycle waits on a `WaitGroup`. There is no per-cycle ordering between components.
- `MaybeActivate` returns `ActivationCodeNoInput` without running `f` when **no input port has signals**. A single signal on any one input makes the component "ready" — the activation function must handle partial inputs itself (or return `ErrWaitingForInputs*`).
- The cycle is always appended to `RuntimeInfo.Cycles`, even when it errored.
- An empty mesh (`Run` with zero components) is a cycle error (`errNoComponents`).

## Activation result codes

`ActivationResultCode` (in `component/activation_result.go`): `OK`, `NoInput`,
`ReturnedError`, `Panicked`, `WaitingForInputsClear`, `WaitingForInputsKeep`, `HookFailed`.
Panics inside activation functions are recovered (with stack trace) and become `Panicked`
results — a component panic never crashes the mesh; the error strategy decides whether the run
stops. `IsError()` is true for both `ReturnedError` and `HookFailed` results, so component-level
hook failures stop the mesh under `StopOnFirstErrorOrPanic` and surface in `Run()`'s error.

## Waiting-for-inputs protocol

Control-flow sentinels in `component/errors.go` — returned **by activation functions**, not real failures:

- `ErrWaitingForInputs` — skip this cycle; the scheduler **clears** the component's inputs.
- `ErrWaitingForInputsKeep` — skip this cycle; inputs are **kept** for the next cycle (use when accumulating partial inputs, e.g. waiting for both operands).

A component that reported waiting is not drained (its outputs are not flushed) and does not count as an "error" under any strategy.

## Drain phase (`drainComponents`)

After each non-final cycle: clear inputs of activated components (except `WaitingForInputsKeep`),
then `FlushOutputs` on every component that activated (except those waiting for input).
Components are drained in **name order** (`Collection.AllOrdered`), so fan-in signal order is
deterministic. `Flush` fans out **the same `*Signal` pointers** to all connected inputs, then
clears the source port (only when all deliveries succeeded — errors are joined). Flushing a port
with no signals or no pipes is a no-op, not an error.

## Stop conditions (`mustStop`, checked in order)

1. Cycle limit hit (`config.CyclesLimit`, default **1000**; 0 = unlimited) → `ErrReachedMaxAllowedCycles`
2. Time limit hit (`config.TimeLimit`, default **5s**; 0 = unlimited) → `ErrTimeLimitExceeded`. Checked between cycles only — a running activation function is never interrupted.
3. Error strategy (`config.ErrorHandlingStrategy`, default `StopOnFirstErrorOrPanic`) — checked **before** the natural stop so errors are never swallowed:
   - `StopOnFirstErrorOrPanic` → stop with `ErrHitAnErrorOrPanic` (includes hook failures)
   - `StopOnFirstPanic` → errors ignored, panics stop with `ErrHitAPanic`
   - `IgnoreAll` → run until natural stop or a limit
4. **Natural stop**: no component activated in the last cycle → `nil` error. This is the normal termination path — a mesh with a loopback pipe or a self-feeding component never stops naturally.

When writing tests that expect limits to trigger, remember the defaults: an infinite mesh stops at cycle 1001 or 5s, whichever comes first.

## Scaling characteristics (measured)

Empirical envelope from stress experiments (July 2026, 8-core/16 GiB arm64 laptop). The
absolute numbers are machine-specific; the complexity classes are the durable part.

- **Width scales near-linearly.** ~1.5–4 µs of scheduler overhead per component per cycle
  and ~300 B of heap per component: a 10⁶-component mesh builds in ~2 s and runs a
  one-wave computation in ~10 s. But `runCycle` spawns one goroutine per component per
  cycle — ready or not — so at 10⁷ components the goroutine stacks alone (tens of GiB)
  are an OOM risk before speed becomes the problem.
- **Fan-in is O(N²).** `ForwardSignals` appends one signal at a time, and each append
  copies the destination port's whole signal group (`port.putSignals` →
  `signal.Group.With`). N outputs converging on a single input port become impractical
  around N ≈ 10⁵ (tens of seconds spent in one drain). Guarded by `BenchmarkMeshFanIn`.
- **Long-running meshes are memory-bound, not time-bound.** Per-cycle cost stays flat as
  cycle count grows (~10³–10⁴ cycles/s depending on width), but `RuntimeInfo.Cycles`
  retains an `ActivationResult` for every component in every cycle (~100 B × components ×
  cycles) and nothing is freed during `Run` — even though the run loop itself only reads
  `Cycles.Last()`. 100 components × 10⁵ cycles already holds ~1 GiB. Rule of thumb: keep
  components × cycles per `Run` under ~10⁸ on a 16 GiB machine.

## Component state

`component.State` (`map[string]any`) persists across cycles and across `Run`s; it is only reset
via `ResetState()` or `WithInitialState`. Safe without locks because a component activates at
most once per cycle in a single goroutine. Rich API: `Get`, `GetOrDefault`, `Set`, `SetIfAbsent`,
`Upsert` (creates if missing), `Update` (only if present), `Delete`, and the generic
`component.MustGetTyped[T](state, key)` (panics on missing key or wrong type — fine inside
activation functions, panics become `Panicked` results).
