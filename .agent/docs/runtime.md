# Runtime — execution model

How a mesh runs. Source: `fmesh.go` (`Run`, `runCycle`, `mustStop`, `drainComponents`),
`component/activation.go`, `component/activation_result.go`.

## Run loop

`FMesh.Run(ctx)` returns `(*RuntimeInfo, error)`. Steps:

0. **`cleanUpPreviousRun`** — clears all output ports (no signal build-up between runs), resets
   `RuntimeInfo` and starts the run clock. A mesh is re-runnable. It gets the caller's context,
   without the time-limit deadline.
1. **Time-limit deadline** — when `TimeLimit > 0`, `Run` wraps the context with
   `context.WithTimeout(ctx, TimeLimit)`. The limit is a real deadline that reaches activation
   functions, not only a check between cycles. It must come **after** the clock starts:
   `contextError` tells the time limit from a caller's cancellation by run duration, so the
   deadline must not fire before that duration reaches the limit. This context goes to every
   activation function, hook and the drain.
2. **`beforeRun` hooks** — include a **default hook that validates mesh structure** (parent mesh /
   parent component wiring; pipe destinations belong to the same mesh). It runs on **every** `Run`,
   in component-name order (deterministic errors). An empty mesh fails here with
   `fmesh.ErrNoComponents`, so it never runs a cycle and user `BeforeRun` hooks do not fire.
3. An already-canceled context stops here: **zero** cycles run.
4. **Loop**: `runCycle` → `mustStop` → `drainComponents`. Stop conditions are checked **before**
   the drain, so the final cycle's outputs are not flushed.
5. **`afterRun` hooks** fire in a defer. Their error is returned only when the run itself did not
   fail.

### Cycle history

`RuntimeInfo.Cycles` holds executed cycles — the main observability surface.
- By default it keeps every cycle for the whole run and grows without bound (see "Scaling").
- `fmesh.WithCyclesHistoryLimit(n)` keeps only the last `n` cycles. The container enforces it
  (`cycle.Group.SetLenLimit`, set in `newRuntimeInfo`): `Add` evicts the oldest, so the run loop
  cannot bypass it.
- Cycle *numbers* keep counting after eviction (the next number comes from the last cycle, not the
  group length).

For any other retention policy (first N, every k-th, errors only), use a mesh-level `AfterCycle`
hook, which gets each cycle as it completes. Pair it with `WithCyclesHistoryLimit(1)` to keep the
engine's own memory minimal. Keeping the first N cycles:

```go
var startup []*cycle.Cycle
fm.SetupHooks(func(h *fmesh.Hooks) {
    h.AfterCycle(func(ctx context.Context, cc *fmesh.CycleContext) error {
        if cc.Cycle.Number() <= N {
            startup = append(startup, cc.Cycle)
        }
        return nil
    })
})
```

The hook runs synchronously in the run loop. Keep it cheap, or hand the cycle to a buffered channel
read by your own goroutine (the same pattern streams history to disk or a store).

## One cycle (`runCycle`)

- `beforeCycle` hooks, then every component's `MaybeActivate(ctx)` in **its own goroutine**, then
  `wg.Wait()`, then `afterCycle` hooks. There is no ordering between components within a cycle.
- `MaybeActivate` returns `ActivationCodeNoInput` without running the function when **no input port
  has signals**. One signal on any input makes a component ready — the activation function must
  handle partial inputs itself (or return a waiting sentinel).
- Only results of ready components are recorded. `NoInput` results are never added to
  `Cycle.ActivationResults()`, so a missing `ByName(name)` entry means "no input that cycle".
  Consumers of `RuntimeInfo` (including hooks) must treat a nil result as not activated. Waiting,
  error, panic and `HookFailed` results are always recorded. This keeps sparse meshes (pipelines,
  rings) free of noise.
- The cycle is always added to `RuntimeInfo.Cycles`, even when it failed. It is added after
  `afterCycle` runs, so that hook sees it only as `cc.Cycle`, not as `Cycles.Last()`.

## Activation

`activate` runs three stages. Each recovers its own panic, so no stage re-runs:

1. `BeforeActivation` hooks. If one fails, the activation function is skipped and the result is
   `HookFailed`.
2. The activation function.
3. `AfterActivation` hooks — run exactly once, whatever happened before.

Result codes (`component/activation_result.go`): `OK`, `NoInput`, `ReturnedError`, `Panicked`,
`WaitingForInputsClear`, `WaitingForInputsKeep`, `HookFailed`.

- A panic in the function or a hook is recovered with its stack and becomes a `Panicked` result
  carrying a `*component.PanicError`. A component panic never crashes the mesh; the error strategy
  decides whether the run stops.
- How hook failures and hook panics re-code the result: [hooks.md](hooks.md).
- `IsError()` is true for `ReturnedError` and `HookFailed`, so hook failures stop the mesh under
  `StopOnFirstErrorOrPanic` and surface in `Run`'s error.

## Waiting-for-inputs protocol

Sentinels in `component/errors.go`, returned **by activation functions**. They are control flow,
not failures:

| Sentinel | Code | Inputs |
|---|---|---|
| `ErrWaitDroppingInputs` (or bare `ErrWaitingForInputs`) | `WaitingForInputsClear` | **cleared** at drain |
| `ErrWaitKeepingInputs` | `WaitingForInputsKeep` | **kept** for the next cycle (accumulating, e.g. a join waiting for its second operand) |

Both wrap `ErrWaitingForInputs`, so `errors.Is(err, component.ErrWaitingForInputs)` means "waiting,
either mode". A waiting component's outputs are not flushed, and waiting is not an error under any
strategy.

## Drain (`drainComponents`)

Runs after every non-final cycle, over activated components in **name order**
(`Collection.AllOrdered`), so fan-in order is deterministic. Two passes, which cannot be merged
(flushing delivers into downstream inputs, and a single pass would clear them):

1. Clear the inputs of every activated component, except `WaitingForInputsKeep`.
2. `FlushOutputs` on every activated component, except waiting ones.

`Flush` sends **the same `*Signal` pointers** to every connected input, then clears the source —
only if every delivery succeeded (errors are joined). Flushing a port with no signals or no pipes is
a no-op.

## Stop conditions (`mustStop`, in order)

1. **Cycle limit** (default **1000**; `WithUnlimitedCycles` removes it) →
   `ErrReachedMaxAllowedCycles`. The limit is **exact**: the number of cycles that execute, not
   limit+1. It also requires `HasActivatedComponents()`, so a last allowed cycle that was already
   empty falls through to the natural stop instead of a false error.
2. **Time limit** (default **5s**; `WithUnlimitedTime` removes it) → `ErrTimeLimitExceeded`. It is
   also the run-context deadline, so an activation function that respects its context is
   interrupted. One that ignores it runs to completion, and the mesh stops after it returns.
3. **Context canceled** → `ErrRunCanceled`, wrapping `ctx.Err()` (`errors.Is(err,
   context.Canceled)` works). Checked **before** the error strategy: a canceled run makes activation
   functions return `ctx.Err()`, which would otherwise look like ordinary errors and hide why the
   mesh stopped. A caller deadline shorter than `TimeLimit` is reported as `ErrRunCanceled`, not
   `ErrTimeLimitExceeded` — `contextError` tells them apart by elapsed time.
4. **Error strategy** (`WithErrorHandlingStrategy`, default `StopOnFirstErrorOrPanic`; an unknown
   one fails `New` with `ErrUnsupportedErrorHandlingStrategy`) — checked **before** the natural
   stop, so errors are never swallowed:
   - `StopOnFirstErrorOrPanic` → `ErrHitAnErrorOrPanic` (includes hook failures)
   - `StopOnFirstPanic` → errors ignored; panics stop with `ErrHitAPanic`
   - `IgnoreAll` → run until a natural stop or a limit
5. **Natural stop** — no component activated in the last cycle → `nil`. The normal end. A mesh
   with a loopback pipe or a self-feeding component stops naturally only once it stops emitting
   into the loop.
6. **Livelock** (`WithLivelockThreshold(n)`, default **2**; `WithoutLivelockDetection` disables) →
   `ErrLivelockDetected`. After the natural stop (a livelocked cycle activated something) and last
   (a real error is always the better explanation).
   - A cycle is *stalled* when every activated component returned `ErrWaitKeepingInputs` **and**
     the mesh's pending signal count (after activation) equals the count the previous drain left.
     Both halves are needed: the first alone flags a component a hook or its own activation
     function feeds or consumes; the second alone flags a busy but idempotent mesh. The baseline is
     taken after the drain, so `n` is exact: a stall that starts right after a productive cycle
     stops the run `n` cycles later, not `n+1`.
   - `n` stalled cycles in a row end the run. The error names each waiting component with its empty
     and non-empty input ports, plus a count of components with no input signals.
   - Waiting components are not drained, so a stalled cycle moves no signals, and the next cycle
     is identical **as long as activation depends only on inputs**. A waiter that proceeds because
     of its `State()`, the clock or the outside world is a false positive; so is a hook that swaps
     signals one for one. Waiters that *drop* inputs never count as stalled:
     next cycle they have no input (a natural stop). The pending count alone cannot tell, because
     it is taken before the drain clears the dropped inputs.

In tests that expect a limit: with defaults, an infinite busy mesh stops after 1000 cycles or 5s,
whichever comes first; a stalled mesh stops after 2 cycles via livelock detection.

### Cancellation is cooperative

Go cannot preempt a goroutine, so:
- The context is checked **between cycles**. A started cycle always runs to completion.
- An activation function that ignores its context blocks the mesh while it runs; `Run` cannot
  return before `wg.Wait()`.
- So a mesh is only as interruptible as its slowest activation function. Pass the context to
  anything that blocks; check `ctx.Err()` in long loops.
- An already-canceled context runs **zero** cycles.

### Panic reporting

- A recovered panic is a `*component.PanicError` with the component name, the panic value and the
  stack.
- `Error()` is **one line** (`panicked: <value>`), so logs and assertions stay readable. Get the
  stack with `errors.As` and `StackTrace()`; the run loop's debug hook prints it when `WithDebug` is
  on.
- `Unwrap` returns the panic value when it is an `error`, so `errors.Is` reaches what was thrown.
- To build a run error from a cycle, use `cycleFailures`, not `AllErrorsCombined`/
  `AllPanicsCombined` passed to `%w` directly: a nil error in `%w` renders `%!w(<nil>)`.

## Scaling characteristics (measured)

Measured July 2026 on an 8-core / 16 GiB arm64 laptop. The numbers are machine-specific; the
complexity classes are what lasts.

- **Width scales near-linearly.** ~1.5–4 µs of scheduler overhead per component per cycle, and
  ~300 B of heap per component. A 10⁶-component mesh builds in ~2 s and runs one wave in ~10 s. But
  `runCycle` starts one goroutine per component per cycle, ready or not, so at 10⁷ components the
  goroutine stacks alone (tens of GiB) risk OOM before speed is the problem.
- **Fan-in is O(N²).** Each delivery copies the destination port's whole signal group
  (`port.putSignals` → `signal.Group.With`). N outputs into one input port become impractical near
  N ≈ 10⁵ (tens of seconds in one drain). Guarded by `BenchmarkMeshRun/fan-in`.
- **Long runs are memory-bound, not time-bound.** Per-cycle cost stays flat (~10³–10⁴ cycles/s
  depending on width). But by default `RuntimeInfo.Cycles` keeps an `ActivationResult` for every
  component with input in every cycle (~100 B × components × cycles) and frees nothing during
  `Run`, although the loop only reads `Cycles.Last()`. 100 components × 10⁵ cycles already holds
  ~1 GiB in a dense mesh. Rule of thumb: keep components × cycles per `Run` under ~10⁸ on 16 GiB, or
  bound memory with `WithCyclesHistoryLimit`.

## Component state

`component.State` (`map[string]any`) persists across cycles and `Run`s. Only `ResetState()` or
`WithInitialState` resets it. It needs no locks: a component activates at most once per cycle, in
one goroutine.

API: `Has`, `Get`, `GetOrDefault`, `Set`, `SetIfAbsent`, `Upsert` (creates if missing), `Update`
(only if present), `UpdateAndGet`, `Delete`, and the generic `GetTyped[T](key)` (error on a missing
key or wrong type). There is deliberately no panicking `Must` twin — return the error from the
activation function.
