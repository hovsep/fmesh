# Naming

## CoW vs mutating

| Semantics | Prefix |
|---|---|
| Copy-on-write (returns a new value) | `With` / `Without` |
| Mutating field setter (modifies and returns the receiver) | `Set` |
| Mutating collection modifier | `Add` / `Remove` |

Never mix them. For a new method, ask: does it return a new value or mutate? Pick the prefix to
match. Internal mutating helpers use unexported `set…`.

## `With` vs `Set` — the exact rule

Use `With` **only** for:
- **CoW**: clones the receiver (or a sub-value) and returns the new instance.
- **Functional option**: a free function returning an `Option` (e.g. `port.WithDescription`,
  `component.WithActivationFunc`).
- **A builder doing real work beyond assignment**: e.g. a nil guard plus prefix logic, iterating
  child objects, appending to a slice.

Use `Set` for **everything else** that is a plain `field = value; return receiver` mutation,
exported or not:
- Exported: `cycle.SetNumber`, `component.SetLogger` (marks the logger custom, so the mesh never
  overrides it)
- Unexported: `port.setSignals`, `port.setPorts`

**No exception for metadata.** Mutating types (`FMesh`, `Component`, `Port`, the collections,
`cycle.Cycle`) have **no** metadata methods. `Meta()` returns the live store; mutate that:
`fm.Meta().Set(k, v)`. Chaining lives on `*meta.Meta`.

**No dual forms.** If a capability has a `With*` option, do **not** also add a `Set*` method,
unless mutation after construction is truly needed. Logger is the one exception:
`component.WithLogger`/`component.SetLogger` mark the logger custom, and `component.InheritLogger`
(called by `fmesh.AddComponents`) sets the mesh logger only on components without a custom one.

## Metadata operations

| Type kind | How metadata changes |
|---|---|
| CoW (`signal.Signal`) | `WithMeta(k, v)`, `WithMetaMany(m)`, `WithoutMeta(keys...)` return a **new value** — the only way to get a changed signal. `signal.Group` has `WithMeta` for its own store. |
| Mutating (`fmesh.FMesh`, `component.Component`, `port.Port`, the collections, `cycle.Cycle`) | No methods. Use the store: `x.Meta().Set(k, v)`, `.SetMany(m)`, `.Remove(keys...)`, `.Clear()`. |

- Replace all metadata: `x.Meta().Clear().SetMany(m)`.
- There is no `WithOnlyMeta`/`WithNoMeta` on signals: write `WithoutMeta(s.Meta().Keys()...)`.
- `meta.Meta` mutators: `Set`, `SetMany`, `Remove`, `Clear`. `Clone()` and `Filter(pred)` return a
  new store.

Batch metadata on contents (`WithMetaOnEach`, `WithoutMetaOnEach`) exists on `signal.Group` only;
see "Metadata on groups and collections" in [design.md](design.md). A collection's own metadata is
reached like any mutating type: `c.Meta().Set(k, v)`.

## Constructor options

Options use the `With` prefix and go to `New(...)`. `WithMeta(k, v)` is an option on every
constructor that takes options (`fmesh.New`, `component.New`, `port.NewInput`, `port.NewOutput`).

| Capability | Option |
|---|---|
| Activation function | `component.WithActivationFunc(f)` |
| Component description | `component.WithDescription(s)` |
| Initial state | `component.WithInitialState(fn)` |
| Logger | `component.WithLogger(l)` |
| Mesh config | `fmesh.WithCyclesLimit(n)`, `WithTimeLimit(d)`, `WithErrorHandlingStrategy(s)`, … (the only way to set config) |

Post-construction `Set*` methods exist only where mutation after `New()` is required, e.g.
`Component.SetLogger`, `InheritLogger`, `SetParentMesh`, `SetupHooks`, `port.Collection.SetParentComponent`,
`cycle.Cycle.SetNumber`, `cycle.Group.SetLenLimit`, and `ActivationResult.SetActivated` /
`SetCode`.

Mutating methods that *append* use `Add*`, even on result types: `ActivationResult.AddError`.

## Collection/group operations

`Any(p)`, `Every(p)`, `Count(p)`, `Map`, `MapIf`, `Filter`, `ForEach`, `ForEachIf`, `Reduce`,
`ReducePayloads`, `Join`, `Find(p)`, `First()`. The names mean the same on every group and
collection; keyed collections (`port.Collection`, `component.Collection`) traverse in name order,
so their `First()` is the first by name.

## Error returns

Fallible methods return `error` last. Truly infallible ones (`Filter`, `Map`, `signal.Signal`
builders) keep their fluent return type. `ForEach` on every collection type returns `error` (stops
on the first).

## Predicates

Prefer combinators over inline closures: `Not`, `And`, `Or`, `HasMeta`, `HasAnyMeta`,
`MetaEquals`, `MetaContains`. Component level: `component.HasSignalsOn(names...)` returns the
`func(*Component) bool` that `component.When` takes.

## Combinators

Functions that take and return an `ActivationFunc` are named for what they do, with no prefix:
`Sequential`, `When`, `RequireInputs`, `Pipeline`. `With*` stays reserved for options, so
`WithActivationFunc(Sequential(...))` reads as an option wrapping a value.

## Stuttering

Do not repeat the package name in a type or function name:
- `meta`: `Predicate`, `Value` (not `MetaPredicate`/`MetaValue`). `meta.Meta` itself is the
  accepted exception, like `context.Context`.
- `component`: `ResultPredicate` (not `ActivationResultPredicate`).
- Methods on a type do not repeat it: `ActivationResult.Err()`, `SetCode`, `AddError` (not
  `ActivationError()`, `SetActivationCode`, `AddActivationError`).
