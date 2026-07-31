# Downstream consumers

F-Mesh is a library. Its callers are outside this repository, so **"no caller in this repo" is not
evidence that a public symbol is unused.** Two sibling repos depend on the public API:

| Repo | Contents |
|---|---|
| [`hovsep/fmesh-examples`](https://github.com/hovsep/fmesh-examples) | Every documented example. **Five Go modules**, not one — the root plus `life/`, `ray_tracer/`, `simulation/` and `can_bus/advanced/`, each with its own `go.mod`. |
| [`hovsep/fmesh-graphviz`](https://github.com/hovsep/fmesh-graphviz) | The DOT exporter documented in wiki `701.-Export`. |

## Before deleting any exported symbol

Static analysis of this repo cannot see these consumers. A `gopls references` sweep will report a
method as used only by its own unit test, and that method may still be load-bearing downstream.
Grepping the sibling repos for `package.Symbol` is **not** sufficient either: it finds
`component.WithIndexedInputs` but is blind to method calls like `.FindAny(...)` or `.Scalars()`,
which is what most deletions actually remove.

The only reliable check is to compile them. See "Compile-checking downstream" below.

### This has already gone wrong once

A cleanup pass removed 107 exported symbols on the evidence that they had no in-repo caller. Six of
them were used downstream — `signal.Signal.WithScalars`, `fmesh.FMesh.Scalars`, `cycle.Group.Count`,
`component.Collection.FindAny`, `component.Collection.All` and `cycle.Group.ForEach`. All were
restored. `Signal.WithScalars` is the instructive one: it had **zero references anywhere in this
repo, not even a test**, which reads as the strongest possible case for deletion, and `ray_tracer`
calls it.

Two signals were better predictors than the reference count, and both were noticed and overruled at
the time:

- **Symmetry.** `Signal.WithLabels` survived and `Signal.WithScalars` did not; `FMesh.Labels()`
  survived and `FMesh.Scalars()` did not. A label/scalar pair where only one half looks unused
  almost always means the other half is used somewhere you cannot see.
- **Coherent feature sets.** Removing part of a documented group (the metadata tiers in
  `design.md`, the predicate combinators, the typed accessors) leaves a worse API than removing all
  of it or none.

## Compile-checking downstream

Run this before proposing any public API removal, and again before finishing.

```bash
# 1. Fetch the consumer (no need to clone into the repo)
mkdir -p /tmp/ex && curl -sL \
  https://codeload.github.com/hovsep/fmesh-examples/tar.gz/refs/heads/main \
  | tar -xz -C /tmp/ex --strip-components=1

# 2. Point it at the working copy, per module
cd /tmp/ex && go mod edit -replace github.com/hovsep/fmesh=/path/to/fmesh
go build -gcflags=-e ./...
```

Three things will mislead you if you skip them:

1. **`-gcflags=-e`** — without it the compiler stops at "too many errors" per package and the error
   list is silently truncated.
2. **Every module separately.** `./...` does not descend into nested modules, so a plain
   `go build ./...` at the root **never compiles `life/`, `ray_tracer/`, `simulation/` or
   `can_bus/advanced/`** — which is most of the interesting code.
3. **Diff against `main`, don't read the raw output.** `fmesh-examples` pins an old version and is
   already broken against `main` (it still calls `fm.Run()` with no context, and uses
   `PayloadOrDefault`, `PayloadOrNil`, `Component.AddLabel`). Ten root-module packages fail before
   you change anything. Build against a `main` worktree and against your branch, then compare, or
   you will not be able to tell your breakage from the pre-existing kind:

```bash
git worktree add -q --detach /tmp/base main
for r in /tmp/base /path/to/fmesh; do
  go mod edit -replace github.com/hovsep/fmesh=$r
  go build -gcflags=-e ./... 2>&1 \
    | grep -oE "has no field or method [A-Za-z]+|undefined: [a-z]+\.[A-Za-z]+" | sort -u > /tmp/e.$$
done   # then: comm -13 <baseline> <branch>
```

`fmesh-examples` also depends on `fmesh-graphviz`, and the root module's `internal/` package imports
it — so 13 top-level examples cannot be type-checked at all while the exporter fails to build. To
get past that, copy the module out of the module cache, patch it locally, and `replace` it. That
copy is a test harness only; never commit it.

## When a removal is still the right call

Breaking these repos is sometimes correct — they need a large migration for the context API
regardless. But it is the user's decision, not a detail to absorb into a cleanup. Report the exact
symbol list and who calls it, and note where the replacement is genuinely better:
`Components().All()` returned a map, so the exporter iterated it non-deterministically;
`AllOrdered()` is the better answer for anything that renders output.
