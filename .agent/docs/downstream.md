# Downstream consumers

F-Mesh is a library, and its callers live outside this repo. **"No caller in this repo" is not
evidence that a public symbol is unused.** Two sibling repos use the public API:

| Repo | Contents |
|---|---|
| [`hovsep/fmesh-examples`](https://github.com/hovsep/fmesh-examples) | Every documented example. One Go module (`basics/`, `patterns/`, `simulation/`, `graphics/`, `internal/`). Its `internal/` package imports `fmesh-graphviz/dot`. |
| [`hovsep/fmesh-graphviz`](https://github.com/hovsep/fmesh-graphviz) | The DOT exporter documented in wiki `701.-Export`. |

## Before deleting any exported symbol

Static analysis of this repo cannot see these callers:
- `gopls references` may show a method used only by its own unit test while downstream depends on it.
- Grepping the sibling repos for `package.Symbol` is **not** enough. It finds
  `component.WithIndexedInputs` but misses method calls like `.FindAny(...)`, which are most of
  what deletions remove.

**The only reliable check is to compile them** (next section).

Why this matters: one cleanup removed 107 exported symbols with no in-repo caller. Six were used
downstream and had to be restored; one had zero references here, not even a test. Two hints
predicted this better than reference counts:
- **Symmetry.** When only one half of a pair (`WithX`/`WithY`, `X()`/`Y()`) looks unused, the other
  half is almost always used somewhere you cannot see.
- **Coherent feature sets.** Removing part of a documented group (the metadata tiers in
  `design.md`, the predicate combinators, the typed accessors) leaves a worse API than removing
  all of it or none.

## Compile-checking downstream

Run this before proposing any public API removal, and again before finishing.

Both repos are usually checked out next to this one (`../fmesh-graphviz`, `../fmesh-examples`);
clone them if not. Neither has a `replace`. Add one for the check and drop it after — it must not
be committed.

```bash
# 1. The exporter first: the examples' internal/ package imports it, so while it
#    fails to build, every example calling internal.HandleGraphFlag fails too.
cd ../fmesh-graphviz
go mod edit -replace github.com/hovsep/fmesh=../fmesh
go build -gcflags=-e ./... && go vet ./...

# 2. Then the examples (one module), against both working copies
cd ../fmesh-examples
go mod edit -replace github.com/hovsep/fmesh=../fmesh -replace github.com/hovsep/fmesh-graphviz=../fmesh-graphviz
go build -gcflags=-e ./... && go vet ./...

# 3. Afterwards, in both
go mod edit -dropreplace github.com/hovsep/fmesh -dropreplace github.com/hovsep/fmesh-graphviz
```

Do not skip these three:

1. **`-gcflags=-e`** — without it the compiler stops at "too many errors" per package and the list
   is silently cut short.
2. **`go vet`, not only `go build`.** `go build ./...` skips `_test.go`, and much downstream API
   use is in tests (e.g. `simulation/life/*_test.go`). `go vet` type-checks tests without running
   them (`make test` there takes ~15 minutes).
3. **Diff against `main`.** `fmesh-examples` pins a released version, so it may already be broken
   against `main`. Build against a `main` worktree and against your branch, then compare:

```bash
git worktree add -q --detach /tmp/base main
for r in /tmp/base /path/to/fmesh; do
  go mod edit -replace github.com/hovsep/fmesh=$r
  go build -gcflags=-e ./... 2>&1 \
    | grep -oE "has no field or method [A-Za-z]+|undefined: [a-z]+\.[A-Za-z]+" | sort -u > /tmp/e.$(basename $r)
done
comm -13 /tmp/e.base /tmp/e.fmesh   # errors only your branch introduces
```

## When a removal is still right

Breaking these repos is sometimes correct, but it is the user's decision, not a detail to absorb
into a cleanup. Report the exact symbol list and who calls each one. Say where the replacement is
truly better — e.g. `Components().All()` returns a map, so iterating it is non-deterministic;
`AllOrdered()` is better for anything that renders output.
