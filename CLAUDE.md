# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Authoritative docs

This repo already carries detailed agent guidance. Read it before making changes — do not
duplicate or contradict it here:

- `AGENTS.md` — top-level agent guide (layout, priorities, git rules)
- `.agent/docs/design.md` — architecture, invariants, per-package rules
- `.agent/docs/runtime.md` — run loop, activation lifecycle, stop conditions, component state
- `.agent/docs/hooks.md` — hook levels (mesh/component/port), semantics, plugins
- `.agent/docs/naming.md` — `With`/`Set`/`Add` conventions, CoW vs mutating
- `.agent/docs/testing.md` — test style and required coverage
- `.agent/docs/benchmarking.md` — benchmark best-practices, size sweeps, fuzzing, benchstat CI
- `.agent/docs/workflow.md` — safe editing/rename practices
- `.agent/plans/` — historical implementation plans (reference only)

## Commands

```bash
make test    # go test ./...
make race    # go test -race ./...   (concurrency is core — run this for scheduler/port changes)
make lint    # golangci-lint run ./...
make fix     # golangci-lint run --fix
make fmt     # go fmt ./...
make check   # race + lint
make bench   # go test -bench with -benchmem (no unit tests)
make deps    # go mod tidy
```

Single test: `go test ./signal/ -run TestSignal_WithLabel -v`
Single package: `go test ./component/...`
Integration suites live in `integration_tests/<topic>/` and run as normal `go test`.

Verify before finishing: `make test && make lint && make fmt`. Key linters enforced:
`errcheck`, `govet` (shadow), `prealloc`, `dupl`, `gocyclo` (min-complexity 15), `testifylint`,
`gosec`. Config: `.golangci.yml`. Go 1.26.

## Hard rules

- **Never `git commit`/`git push`, amend, or rewrite history.**
- **Public API changes require explicit user approval.** Adapting internal callers does not.
- Ask before relaxing a constraint or introducing a new pattern/helper/abstraction.

## Architecture in one pass

F-Mesh is a Flow-Based Programming framework. An app is a directed graph of **components**
connected by **pipes** between **ports**; data flows as **signals**; execution runs in discrete
synchronized **cycles**. Priority is **simplicity and a clean API, not performance.**

Root package `fmesh` orchestrates; the graph primitives live in subpackages:

| Concept | Type | Package |
|---|---|---|
| Data packet | `*signal.Signal` | `signal` |
| Ordered signal collection | `*signal.Group` | `signal` |
| Data endpoint | `*port.Port` | `port` |
| Connection (output→input) | `PipeTo` | `port` |
| Building block | `*component.Component` | `component` |
| Execution tick | `*cycle.Cycle` | `cycle` |
| String / numeric metadata | `*meta.Labels` / `*meta.Scalars` | `meta` |

**Execution loop** (`fmesh.go` `Run` → `runCycle` → `drainComponents`): each cycle activates all
ready components concurrently (one goroutine each, `sync.WaitGroup`), collects `ActivationResult`s,
then drains — clears inputs and flushes outputs through pipes to downstream inputs. The mesh stops
naturally when no component activated in a cycle, or on cycle limit / time limit / error strategy.
Components can signal "waiting for input" to keep their inputs for the next cycle. Config
(`config.go`): `CyclesLimit` default 1000, `TimeLimit` default 5s, `ErrorHandlingStrategy` default
`StopOnFirstErrorOrPanic` (see `errors.go` for the three strategies).

**The central invariant (`.agent/docs/design.md`):** `signal.Signal` and `signal.Group` are
**copy-on-write** — mutating methods return a new value, never touch the receiver; payload is
shallow-copied and `nil` is a valid payload. In contrast `meta.Labels`/`meta.Scalars` and the
`port`/`component`/`cycle` types **mutate in place**. This split drives the naming convention:
`With*`/`Without*` = CoW returning new; `Set*`/`Add*`/`Remove*` = mutating. Never mix them.

No generics (FBP needs mixed-type signals in one group); minimise `reflect`; no chainable
error/"poison object" pattern — fallible methods return `error` last, infallible transforms
(`Filter`, `Map`, `With*`) return their type directly for fluency.
