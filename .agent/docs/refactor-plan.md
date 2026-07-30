# Refactor plan

Transient working document — delete when the phases are done. Standing rules that come out of
this work belong in `design.md` / `naming.md` / `runtime.md`, not here.

Phases are ordered so that semantic changes land before mechanical ones, and so each phase's
diagnostics are available while debugging the next. Phases 1–5 are all breaking; they land on
`main` and ship as **one** release with a migration guide.

## Decisions taken

| Decision | Choice |
|---|---|
| Module path | Stays `github.com/hovsep/fmesh`. No `/v2`, ever. |
| Versioning | v1 forever, minor bumps only — including for breaking changes. |
| Type safety | Untyped signals are by design. Stop claiming otherwise. |
| Output written before a wait | Known caveat, documented, risk on users. Not fixed. |
| Payload immutability | Contract, documented, risk on users. Not enforced. |
| Nil-returning port lookups | Unchanged. |
| Golden rule (no inputs → no activation) | Unchanged, promoted to README. |

## Phase 0 — unbreak distribution

Independent of every other phase. Nothing here touches library code.

**Root cause:** commit `cab6d9c` was once tagged plain `v1.3.0`; proxy.golang.org cached it
permanently, then the tag was renamed to `v1.3.0-tigranakert`. Every other tag carries a codename
suffix, which semver reads as a *pre-release*. Go's `@latest` picks the highest **release**
version and only falls back to pre-releases when there are none — so the single plain `v1.3.0`
beats all 36 codenamed tags regardless of their numbers. Deleting tags on GitHub cannot fix this;
the proxy is immutable.

**Fix:** publish one plain-semver release above `v1.3.0`.

- [ ] Tag and push `v1.12.0` (user runs this — never the agent):
      `git tag v1.12.0 && git push origin v1.12.0`
- [ ] Verify: `go list -m github.com/hovsep/fmesh@latest` → `v1.12.0`
- [ ] Delete the malformed local tag `v.1.4.1` (leading dot)
- [ ] Codenames move to the GitHub Release **title**; the tag stays plain forever
- [ ] `CLAUDE.md` hard rule:
      > **Releases:** a release tag is plain semver (`v1.12.0`) — a suffix makes it a pre-release
      > and invisible to `go get`. Codenames go in the GitHub Release title. The module path stays
      > `github.com/hovsep/fmesh`; there will be no `/vN`.
- [ ] New `.github/workflows/release-guard.yml`: on `push: tags: ['v*']`, fail unless the tag
      matches `^v[0-9]+\.[0-9]+\.[0-9]+$`
- [ ] New `CHANGELOG.md` with explicit `### BREAKING` sections per release
- [ ] README notice: minor versions may contain breaking changes until further notice; pin an
      exact version if that matters to you

**Done when:** a fresh `go get github.com/hovsep/fmesh` + the README quick start compiles and runs.

## Phase 1 — context ✅ DONE

Breaking. Changes what an activation *is*, so everything after builds on it.

- [x] `FMesh.Run()` → `Run(ctx context.Context) (*RuntimeInfo, error)`
- [x] `TimeLimit` implemented as `context.WithTimeout` inside `Run`, so it propagates into
      activation functions instead of only being checked between cycles
- [x] Cancellation checked between cycles in `mustStop`; new `ErrRunCanceled` wrapping `ctx.Err()`
- [x] `component.ActivationFunc` → `func(ctx context.Context, this *Component) error`
- [x] Thread ctx through `MaybeActivate` / `activate` / `sequentialActivationFunc`
- [x] `component/compose.go` combinators (`Sequential`, `When`, `RequireInputs`, `Pipeline`) take ctx
- [x] Hooks take ctx as their **first parameter** (no `Ctx` struct field, no `RunContext`):
      `hook.Group[T].Trigger(ctx, arg)`; hooks are `func(context.Context, T) error`. This replaced
      the planned `RunContext` struct — putting a context in a struct is the `containedctx`
      anti-pattern that the same phase argues against for `Component`.
      - **Runtime path takes ctx**: `Run`, `ActivationFunc`, `MaybeActivate`, `Flush`, `Clear`,
        `FlushOutputs`, `ClearInputs`/`ClearOutputs`, `Forward*`, `MultiForward`.
      - **Construction and seeding stay ctx-free**: `New`, `PipeTo`, `PutSignals`, `PutPayloads`.
        Their hooks receive `context.Background()`; requiring a context there would put
        `context.Background()` in every setup line for no gain.
      - `ErrRunCancelled` → **`ErrRunCanceled`**: the repo's `misspell` linter enforces US
        spelling, which also matches `context.Canceled`.
- [x] Plugin `Init` signatures unchanged (construction-time, not run-time)
- [x] Update: all tests, `integration_tests/**`, benchmarks, `example_test.go`, README, `docs/wiki/**`

**Semantics to document (`runtime.md` + wiki 401):**
cancellation is cooperative and checked between cycles; a running cycle always completes; an
activation function that ignores ctx still blocks the mesh. Go cannot preempt.

**Done:** `integration_tests/constraints/context_test.go` covers cancellation mid-run, an
already-cancelled context running zero cycles, a caller deadline reported as cancellation rather
than as a time limit, `WithTimeLimit` interrupting a ctx-respecting activation function, ignoring
the context still working, and every hook level receiving the run context.
`make test`, `make race` and `make lint` are clean; the README quick start compiles and runs.

## Phase 2 — determinism

Makes phases 3–5 testable: a nondeterministic runtime can't be refactored with confidence.

Today `port.Collection` iterates a `map`, so `Inputs().Signals()` produced 4 distinct orderings
across 200 identical runs. `component.Collection` already solves this with `AllOrdered()`.

- [ ] `port.Collection` gains a cached sorted-name slice, invalidated on `Add`/`Remove`
      (ports are added at construction, so it is built once)
- [ ] `Signals()`, `ForEach`, `Flush`, `PutSignals*` iterate in port-name order
- [ ] Add `port.Collection.AllOrdered()` mirroring `component.Collection`
- [ ] Document the three ordering guarantees in `design.md` + wiki 401:
      1. Within one port: FIFO
      2. Multiple upstreams → one port: upstream component-name order (already true)
      3. Across ports of one component: port-name order (new)
- [ ] Determinism test: N identical runs of a fan-in mesh produce byte-identical output

**Done when:** the 200-run probe yields exactly one ordering, and the README can claim
*"identical input, identical output, every run."*

## Phase 3 — livelock detection

Currently two components waiting on each other burn 1001 cycles and report the generic
`reached max allowed cycles`.

**Why it's decidable, not heuristic:** waiting components are never drained, so they emit nothing;
drop-mode waiters clear their inputs and stop activating (self-resolving). A durable livelock is
therefore exactly: every activated component in the cycle returned a waiting result, with at least
one keep-mode waiter. Nothing moved; the next cycle is bit-identical.

- [ ] `cycle.Cycle.AllActivatedAreWaiting() bool`
- [ ] Consecutive-stalled-cycle counter on the run (threshold, not 1, because a `BeforeCycle` hook
      may legitimately inject signals)
- [ ] `WithLivelockThreshold(n)`, default 2, `0` disables
- [ ] `ErrLivelockDetected` naming the stuck components and, for each, which input ports are empty
- [ ] Rename the wait sentinels so neither behaviour is the unmarked default:
      - `ErrWaitingForInputs` stays as the parent sentinel for `errors.Is`
      - `ErrWaitingForInputs` (drop) → `ErrWaitDroppingInputs`
      - `ErrWaitingForInputsKeep` → `ErrWaitKeepingInputs`

**Done when:** the mutual-wait mesh stops in ~3 cycles with an error naming both components and
their empty ports.

## Phase 4 — API diet and naming harmonisation

One change fixes both: ~460 exported symbols, and a `naming.md` exception large enough to swallow
the rule it qualifies.

The exception exists only because mutating types have `With*` metadata methods that mutate. Those
methods are pure delegation — `fm.AddLabel(k,v)` is `fm.labels.Set(k,v); return fm` — and
`Labels()` on those types already returns the live store.

- [ ] Delete metadata methods on **mutating** types (`FMesh`, `Component`, `Port`,
      `port.Collection`, `cycle.Cycle`, `cycle.Group`): `AddLabel(s)`, `SetLabels`, `ClearLabels`,
      `RemoveLabels` and the scalar equivalents — ~60 methods.
      `fm.AddLabel("env","prod")` → `fm.Labels().Set("env","prod")` (still chainable on `*Labels`)
- [ ] **Keep** `With*`/`Without*` on CoW types (`signal.Signal`, `signal.Group`) — load-bearing there
- [ ] **Keep** constructor options `WithLabel`/`WithScalar` — correct use of `With`
- [ ] Renames:
      | Now | → | Why |
      |---|---|---|
      | `port.Collection.Without(names)` | `Remove(names)` | mutates in place |
      | `Collection.WithParentComponent` | `SetParentComponent` | mutates, returns receiver |
      | `Collection.With*OnEach` | `Set*OnEach` | mutates (CoW `signal.Group` keeps `With*OnEach`) |
      | `Collection.PutSignals` | `PutSignalsOnEach` | it broadcasts to every port |
      | `Collection.PipeTo` | `PipeEachTo` | it builds a full cross-product |
- [ ] Collapse payload accessors — `Payload()` errors only for a zero-value `Signal`, which is a
      construction bug, not a runtime condition:
      - `Signal.Payload() (any, error)` → `Payload() any`
      - delete `PayloadOrNil`, `PayloadOrDefault`, `Group.FirstPayloadOrNil`,
        `Group.FirstPayloadOrDefault`; `Group.FirstPayload() any` (nil for empty group)
      - `As[T]` / `AsOrDefault[T]` remain the typed path
- [ ] Remove the exception paragraph from `naming.md`; the rule becomes absolute:
      **`With` means CoW-returns-new, or a constructor Option. Nothing else.**
- [ ] Update the duplicated naming rule in `CONTRIBUTING.md`

**Done when:** ~70–80 exported symbols are gone, `naming.md` has no exceptions, no functionality lost.

## Phase 5 — diagnostics

- [ ] Fix `%!w(<nil>)` in `fmesh.go` `mustStop` — `%w` is handed `AllPanicsCombined()`, which is nil
      when there are no panics. This string reaches user logs today. Build from non-nil parts with
      `errors.Join`.
- [ ] `component.PanicError{Component, Value, Stack}`; `Error()` returns one readable line,
      `StackTrace()` returns the trace. Today the full `debug.Stack()` is formatted into the error
      string, producing multi-KB single-line errors.

## Phase 6 — docs and release

- [ ] README:
      - delete the "type-safe API" claim; add an **Untyped by design** section (mixed payloads in
        one pipe is the point; mismatches surface at read time; prefer `As[T]` over `AsOrDefault`)
      - resolve the contradiction: "Stream processing" in Use Cases vs. "not suitable for
        long-running components" in Limitations — cut the former
      - replace "no surprises" with the determinism guarantee earned in Phase 2
      - promote the four rules a user needs before writing component #2: the golden rule, fan-in
        ordering, the payload immutability contract, cancellation semantics
- [ ] `signal.AsOrDefault` docs demote it to "when a fallback is genuinely correct"; `As[T]` becomes
      the documented default (this is what silently turned `int64(1000)` into `0`)
- [ ] New wiki **Caveats** page, risks explicitly on users:
      - output written before returning a wait error is retained and re-emitted on the next
        activation — do not write outputs before deciding to wait
      - fan-out shares payload pointers; treat payloads as immutable, read-only, and produce new
        signals rather than mutating received ones; run mesh tests with `-race`
      - one component's error discards siblings' completed work (see backlog)
- [ ] CI step compiling ```go blocks extracted from `docs/wiki/**` (the README is already covered by
      `example_test.go`; the wiki is not covered at all)
- [ ] Migration guide + `CHANGELOG.md` `BREAKING` entry, then cut the single release carrying
      phases 1–5

## Backlog (GitHub issues, not this plan)

- **Drain on error.** `Run` calls `mustStop()` before `drainComponents()`, so a stopping error
  discards successful siblings' outputs. Swapping the order means downstream sees partial data
  instead — wants to be a policy, `WithDrainOnError(bool)`.
- **Transactional activation.** Snapshot output ports before the activation function, restore on
  wait/error/panic. Independent of declarative waiting — preserves dynamic waiting entirely.
  Deferred, not rejected.
- **`signal.Cloner`** — opt-in per-destination payload cloning on fan-out.
- **Debug-mode payload mutation detector** — fingerprint payloads around each activation.
- **Descriptive port-lookup panics** — `component "concat" has no input port "inn" (have: i1, i2)`
  instead of a nil dereference.
- **`WithMaxConcurrency(n)`** — peak goroutines currently equals ready components, unbounded.
  (Verified: no leak, goroutines are reaped every cycle. Purely a resource-cap knob.)
- **`port.WithExpectedType[T]()`** — opt-in per-port type assertion at the pipe boundary.
- **Sequential drain** is the throughput bottleneck at scale, not goroutine spawn (5000 components:
  109ms actual vs 2.2ms ideal). Deprioritised — simplicity over performance.
