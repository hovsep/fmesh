# CLAUDE.md

Guidance for Claude Code (claude.ai/code) when working in this repository.

## Authoritative docs

Read these before making changes. Do not duplicate or contradict them here.

- `.agent/docs/design.md` — invariants, per-package rules, comment and dead-code policy
- `.agent/docs/downstream.md` — sibling repos that use the public API; how to compile-check them
  **before removing any exported symbol**
- `.agent/docs/runtime.md` — run loop, activation, stop conditions, component state
- `.agent/docs/hooks.md` — hook levels (mesh/component/port), semantics, plugins
- `.agent/docs/naming.md` — `With`/`Set`/`Add` conventions, CoW vs mutating
- `.agent/docs/testing.md` — test style and required coverage
- `.agent/docs/benchmarking.md` — benchmarks, size sweeps, fuzzing, benchstat CI
- `.agent/docs/workflow.md` — safe editing and renames
- `docs/wiki/` — user-facing wiki source. `.github/workflows/wiki.yml` overwrites the GitHub wiki
  with it on every push to `main`. Edit pages here, never in the wiki UI.

## Commands

```bash
make check     # race + lint + fmt-check — exactly what CI runs
make test      # go test ./...
make race      # go test -race ./...  (run for scheduler/port changes)
make lint      # golangci-lint run ./...
make fix       # golangci-lint run --fix
make fmt       # go fmt ./...  (rewrites files)
make fmt-check # gofmt -l, fails if anything needs formatting
make bench     # go test -bench with -benchmem (no unit tests)
make deps      # go mod tidy
```

- Single test: `go test ./signal/ -run TestSignal_WithMeta -v`
- Single package: `go test ./component/...`
- Integration suites live in `integration_tests/<topic>/` and run as normal `go test`.

Before starting, run `make test` to confirm the baseline is green.
Before finishing, run **`make check`**. It is the CI gate, so green `make check` means a green
build. Key linters: `errcheck`, `govet` (shadow), `prealloc`, `dupl`, `gocyclo` (min-complexity
15), `testifylint`, `gosec`. Config: `.golangci.yml`. Go 1.27.

## Hard rules

- **Never `git commit`/`git push`, amend, or rewrite history.** Committing and pushing is always
  the user's job. Leave changes in the working tree.
- **API compatibility is not a concern.** F-Mesh is not used in production. Any public API may be
  changed or broken until this doc says otherwise. No deprecation shims or compat layers.
  **But breaking it is the user's call, not a side effect of a cleanup.** `fmesh-examples` (one Go
  module) and `fmesh-export` use the public API, and nothing in this repo references most of it.
  Compile them before removing an exported symbol — see `.agent/docs/downstream.md`.
- Ask before relaxing a constraint or adding a new pattern, helper or abstraction.
- **Release tags are plain semver** (`v1.12.0`). A suffix (`v1.11.1-Shirak`) makes a pre-release,
  which `go get` skips — 36 such tags were never served. The module proxy caches tags forever, so
  a wrong tag cannot be undone by deleting it. Codenames go in the GitHub release **title**. The
  module path stays `github.com/hovsep/fmesh`: v2+ would need a `/vN` suffix in `go.mod` and every
  internal import. `.github/workflows/release-guard.yml` enforces both.
- Breaking changes go in `CHANGELOG.md` under a `BREAKING` heading, with before/after code.

## Architecture in one pass

F-Mesh is a Flow-Based Programming framework. An app is a directed graph of **components**
connected by **pipes** between **ports**. Data flows as **signals**. Execution runs in discrete,
synchronized **cycles**. Priority: **simplicity and a clean API, not performance.**

The root package `fmesh` orchestrates. The graph primitives live in subpackages:

| Concept | Type | Package |
|---|---|---|
| Data packet | `*signal.Signal` | `signal` |
| Ordered signal collection | `*signal.Group` | `signal` |
| Data endpoint | `*port.Port` | `port` |
| Connection (output→input) | `PipeTo` | `port` |
| Building block | `*component.Component` | `component` |
| Execution tick | `*cycle.Cycle` | `cycle` |
| String / numeric metadata | `*meta.Meta` | `meta` |

**Execution loop** (`fmesh.go`: `Run(ctx)` → `runCycle` → `mustStop` → `drainComponents`). Each
cycle activates all ready components concurrently (one goroutine each) and collects their
`ActivationResult`s. The drain then clears inputs and flushes outputs through pipes. The mesh stops
naturally when no component activated in a cycle, or on the cycle limit, time limit, error strategy
or livelock. A component can report "waiting for inputs" and keep or drop its inputs. Details:
`runtime.md`.

**Config** is unexported (`config.go`) and set only through `With*` options on `fmesh.New`.
Defaults: cycles limit 1000, time limit 5s, error strategy `StopOnFirstErrorOrPanic` (the three
strategies are in `errors.go`).

**Context.** `Run(ctx)` passes a context to activation functions, hooks and the drain. `TimeLimit`
becomes a deadline on it. Context is always the first parameter, never a struct field.
Construction and seeding (`New`, `PipeTo`, `PutSignals`) take no context; their hooks get
`context.Background()`. Cancellation is cooperative and checked between cycles.

**The central invariant** (`design.md`): `signal.Signal` and `signal.Group` are
**copy-on-write** — mutating methods return a new value and never touch the receiver. The payload
is shallow-copied, and `nil` is a valid payload. `meta.Meta` and the `port`/`component`/`cycle`
types **mutate in place**. Naming follows this split: `With*`/`Without*` = CoW returning a new
value (or a constructor option); `Set*`/`Add*`/`Remove*` = mutating. Never mix them, with no
exception: mutating types have no metadata methods — `x.Meta().Set(k, v)` writes the live store.
`Signal.Payload()` returns `any` and cannot fail.

The signal **payload** stays `any`: FBP needs mixed-type signals in one group, so pipes cannot be
typed. Generics are fine elsewhere when they remove real duplication (the generic `meta.Meta`
methods are the model); weigh per-instance cost and godoc rendering first — see `design.md`.
Minimise `reflect`. No chainable error / "poison object" pattern: fallible methods return `error`
last; infallible transforms (`Filter`, `Map`, `With*`) return their type directly.
