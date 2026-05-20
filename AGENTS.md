# Agent Guide — F-Mesh

This file is the primary reference for AI coding agents working on this repository.
Read it fully before making any changes.

---

## What this project is

F-Mesh is a **Flow-Based Programming (FBP) framework** for Go.
Applications are built as directed graphs of independent **components** connected by **ports** and **pipes**.
Data packets called **signals** flow through the graph.
Execution runs in discrete synchronized **cycles** (ticks); all eligible components run concurrently per cycle.

The framework is **not optimized for performance**. It prioritizes **simplicity, readability, and clean API design**.

---

## Repository layout

```
fmesh.go              # Mesh orchestrator: Run() loop, cycle management, mustStop()
runtime.go            # RuntimeInfo: start/stop time, cycle list
config.go             # Configurable behavior: error strategy, cycle/time limits
hooks.go              # Mesh-level lifecycle hooks
errors.go             # Mesh-level error definitions

signal/               # Signal (data packet) and Group (ordered collection)
labels/               # Key-value string label collections (attached to signals or ports/components)
port/                 # Port, pipe connections, port collections and groups
component/            # Component, activation logic, state, I/O helpers
cycle/                # Cycle struct and ordered collection
hook/                 # Generic ordered callback group (reusable across packages)
integration_tests/    # End-to-end scenario tests (computation, piping, hooks, state, etc.)
.agent/plans/         # Implementation plans for pending work — check here before starting
```

---

## Core concepts and their types

| Concept | Type | Package |
|---|---|---|
| Data packet | `*Signal` | `signal` |
| Ordered signal collection | `*signal.Group` | `signal` |
| Key-value metadata | `*labels.Collection` | `labels` |
| Component processing unit | `*component.Component` | `component` |
| Data endpoint on component | `*port.Port` | `port` |
| Directed connection between ports | pipe (output→input via `PipeTo`) | `port` |
| One execution tick | `*cycle.Cycle` | `cycle` |
| Ordered cycle collection | `*cycle.Group` | `cycle` |
| Lifecycle callback list | `*hook.Group` | `hook` |

---

## Design principles — follow these at all times

### 1. Signals are immutable (copy-on-write)
All mutating methods on `*Signal` (e.g. `AddLabel`, `SetLabels`, `MapPayload`) return a **new** `*Signal`.
The receiver is never modified. Never add methods to `*Signal` that mutate in place.

**Payload is shallow-copied.** If the payload is a mutable reference type (map, slice, pointer to struct),
the user is responsible for treating it as immutable once placed in a signal. The framework does not
and cannot enforce deep immutability on `any` payloads.

### 2. Groups are immutable (copy-on-write)
All mutating methods on `*signal.Group` (e.g. `Add`, `Filter`, `Map`) return a **new** `*Group`.
The internal `[]*Signal` slice is cloned; individual `*Signal` pointers are shared (safe because signals are immutable).

### 3. Chainable error pattern
All builder/fluent methods store errors in a `chainableErr` field instead of returning `(T, error)`.
The chain short-circuits: if `HasChainableErr()` is true, methods are no-ops and return `self`.
This applies to `Signal`, `signal.Group`, `labels.Collection`, `Port`, and all other chainable types.

### 4. Labels are mutable
`labels.Collection` mutates in place (`Add`, `AddMany`, `Without`, `Clear`).
This is intentional — `port.Port` and `component.Component` hold labels as mutable state.
Do NOT make `labels.Collection` copy-on-write; it would break port and component packages.

### 5. Fan-out uses pointer sharing
When an output port fans out to multiple input ports, the same `*Signal` pointers are forwarded to all destinations.
This is a deliberate design choice — do not add deep-copy logic to `ForwardSignals` or `Flush`.
Users must treat signal payloads as immutable.

### 6. No generics
The signal/group API uses `any` for payload. This is intentional — FBP requires mixed-type signal flows
in a single group. Do not introduce generics into `signal` or `labels` packages.

### 7. No `reflect`
Do not use `reflect` anywhere in the framework. If a method needs to handle non-comparable types,
provide a `Func` variant that accepts a user-supplied comparator (e.g. `ContainsPayloadFunc`).

---

## Package-level rules

### `signal` package
- `Signal.payload` is `[]any{value}` — a single-element slice to allow `nil` as a valid payload
- `cloneForMutation()` is the internal helper for copy-on-write; always use it in mutating methods
- Predicate combinators (`Not`, `And`, `Or`) and label-aware constructors (`HasLabel`, `LabelEquals`, etc.)
  live in `signal/predicates.go` — use these instead of inline closures wherever possible
- `signal.Group` is ordered (backed by a slice); `First()` and `Last()` are meaningful

### `labels` package
- Backed by `map[string]string` — iteration order is not guaranteed; `Keys()` and `Values()` return sorted slices
- `Merge(other)` returns a new collection (non-mutating); all other write methods mutate in place
- `Every(pred)` returns `true` for an empty collection (vacuous truth) — this is correct and intentional

### `port` package
- `Port.Signals()` returns `*signal.Group` — the signal buffer
- `Flush()` fans out signals to all piped ports then clears the source port
- `PipeTo` only connects output→input; `validatePipe` enforces this
- Port labels are mutable (use `labels.Collection` directly)

### `component` package
- `ActivationFunc` is `func(*Component) error`
- `MaybeActivate()` decides readiness and runs the activation function
- `State` is `map[string]any` — persistent across cycles, intentionally untyped
- Components hold `*port.Collection` for inputs and outputs

### `cycle` package
- `cycle.Group` has its own `AnyMatch`/`AllMatch`/`CountMatch` — these are separate from `signal.Group`'s methods
  and are NOT renamed when signal/labels methods are renamed

### `component` and `port` packages
- Both have their own `AnyMatch`/`AllMatch`/`CountMatch` on their collection types
- These are independent of signal/labels and follow their own naming evolution

---

## Naming conventions

| Pattern | Convention |
|---|---|
| Predicate matching (signal.Group, labels.Collection) | `Any(p)`, `Every(p)`, `Count(p)` |
| Predicate matching (cycle, port, component collections) | `AnyMatch`, `AllMatch`, `CountMatch` (legacy, not yet renamed) |
| Add items to a group | `Add(signals...)`, `AddPayloads(payloads...)` |
| Merge two groups | `Join(other)` |
| Transform all items | `Map(mapper)`, `MapPayloads(mapper)` |
| Transform matching items | `MapIf(pred, mapper)`, `MapPayloadsIf(pred, mapper)` |
| Iterate with side effects | `ForEach(action)`, `ForEachIf(pred, action)` |
| Remove items | `Filter(Not(p))` — do not add `Without` convenience wrappers |
| Accumulate | `Reduce(initial, fn)`, `ReducePayloads(initial, fn)` |

---

## Testing conventions

- Unit tests live alongside source files (`*_test.go` in the same package, using `package signal` not `package signal_test`)
- Integration tests live in `integration_tests/` organized by scenario
- Use `github.com/stretchr/testify/assert` and `require`
- Table-driven tests are the standard pattern — always prefer them over repeated inline calls
- Immutability must be tested explicitly — see `signal/immutability_test.go` as the reference pattern
- A `mustAll()` helper pattern (panics on error, for test setup) is acceptable in `_test.go` files

### Do not invent test helpers
**Never create helper functions** (e.g. `check(t, ...)`, `assertNilPayload(t, ...)`) just to reduce
repetition inside a single test. Use plain `assert`/`require` calls directly. If a subtests-based
structure avoids repetition, use `t.Run(name, func(t *testing.T){...})` with inline assertions.
The only exception is a `mustXxx()` panic-on-error helper used strictly for **test setup** (building
fixtures), not for assertions. Ask the user before introducing any new test helper.

---

## How to verify your work

```bash
make test    # go test ./...
make lint    # golangci-lint run ./...
make fmt     # go fmt ./...
```

All three must pass before any change is considered done.
The linter is strict — check `.golangci.yml` for enabled rules.
Key linters to be aware of: `errcheck`, `govet` (with shadow), `prealloc`, `dupl`, `testifylint`.

---

## Git rules — non-negotiable

- **Never** run `git commit` on behalf of the user
- **Never** run `git push` on behalf of the user
- **Never** amend, rebase, or otherwise rewrite history
- Stage and show diffs freely; let the user decide when to commit

---

## Before starting any task

1. Check `.agent/plans/` for an existing plan covering the work
2. If a plan exists, follow it; if scope has changed, update the plan file first
3. Run `make test` to confirm the baseline is green before making changes
4. Understand which packages are in scope — changing a package's **public API** requires explicit approval;
   adapting internal callers to renamed/removed methods does not

---

## Active branch context

Branch `signal_immutability` — work in progress on signal/labels API cleanup.
See `.agent/plans/signal-labels-api-cleanup.plan.md` for the full implementation plan.
