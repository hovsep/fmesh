# Agent Guide — F-Mesh

F-Mesh is a Flow-Based Programming (FBP) framework for Go. Apps are directed graphs of **components** connected by **ports** and **pipes**. Data flows as **signals**. Execution runs in discrete synchronized **cycles**; all eligible components run concurrently per cycle.

Priority: **simplicity and clean API** — not performance.

## Layout

```
fmesh.go / runtime.go / config.go / hooks.go / errors.go
signal/        port/          component/     cycle/
labels/        hook/          integration_tests/
.agent/plans/  (historical)   .agent/docs/   (reference)
```

## Reference docs

- [Design](.agent/docs/design.md) — architecture, invariants, package rules
- [Runtime](.agent/docs/runtime.md) — run loop, activation lifecycle, stop conditions, state
- [Hooks](.agent/docs/hooks.md) — hook levels, semantics, plugins
- [Naming](.agent/docs/naming.md) — CoW vs mutating, method naming
- [Testing](.agent/docs/testing.md) — style, what to cover
- [Workflow](.agent/docs/workflow.md) — how to work safely

## Before starting

Run `make test` — confirm baseline is green.

Public API changes require explicit user approval. Adapting callers does not.

## Verify

```bash
make test && make lint && make fmt
```

Key linters: `errcheck`, `govet` (shadow), `prealloc`, `dupl`, `testifylint`.

## Git

- Never `git commit` or `git push`
- Never amend or rewrite history
