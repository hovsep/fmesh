# Naming

## CoW vs mutating

| Semantics | Prefix |
|---|---|
| Copy-on-write (returns new value) | `With` / `Without` |
| Mutating field setter (modifies receiver, returns receiver) | `Set` |
| Mutating collection modifier | `Add` / `Remove` |

Never mix. When adding a new method: does it return a new value or mutate? Pick the prefix
accordingly. Internal-only mutating helpers use `set…` (unexported).

## `With` vs `Set` — the exact rule

Use `With` **only** when the method is one of:
- **CoW**: clones the receiver (or a sub-value) and returns the new instance
- **Functional option constructor**: a free function returning an `Option` type (e.g. `port.WithDescription`, `component.WithActivationFunc`)
- **Builder that does real work beyond field assignment**: e.g. nil guard + prefix logic, iteration over child objects, appending to a slice

Use `Set` for **everything else** that is a plain `field = value; return receiver` mutating method, whether exported or unexported:
- Exported example: `cycle.SetNumber`, `component.SetLogger` (marks the logger as custom so the mesh never overrides it)
- Unexported example: `port.setSignals`, `port.setPorts`

**No dual-form duplication**: if a capability has a `With*` constructor option, do **not** also add a `Set*` method for the same capability, unless genuine post-construction mutation is needed. Logger is the one exception: `component.WithLogger`/`component.SetLogger` mark the logger custom, and `component.InheritLogger` (called by `fmesh.AddComponents`) sets the mesh logger only on components without a custom one.

## Label operations by type

| | CoW (`signal.Signal`, `signal.Group`) | Mutating (`port.Port`, `component.Component`, `fmesh.FMesh`) |
|---|---|---|
| Add/update one | `WithLabel(k, v)` | `AddLabel(k, v)` |
| Add/update many | `WithLabels(map)` | `AddLabels(map)` |
| Replace all | `WithOnlyLabels(map)` | `SetLabels(map)` |
| Remove specific | `WithoutLabels(names...)` | `RemoveLabels(names...)` |
| Remove all | `WithNoLabels()` | `ClearLabels()` |

## Scalar operations by type

| | CoW (`signal.Signal`, `signal.Group`) | Mutating (`port.Port`, `component.Component`, `fmesh.FMesh`) |
|---|---|---|
| Add/update one | `WithScalar(k, v)` | `AddScalar(k, v)` |
| Add/update many | `WithScalars(map)` | `AddScalars(map)` |
| Replace all | `WithOnlyScalars(map)` | `SetScalars(map)` |
| Remove specific | `WithoutScalars(names...)` | `RemoveScalars(names...)` |
| Remove all | `WithNoScalars()` | `ClearScalars()` |

`meta.Labels` (mutating): `Set`, `SetMany`, `Remove`, `Clear`. `Merge(other)` returns a new collection.
`meta.Scalars` (mutating): `Set`, `SetMany`, `Remove`, `Clear`, `Scale`. `Merge(other)` returns a new collection.

## Group/Collection metadata batch methods

| Method | Effect |
|---|---|
| `WithLabel(k, v)` / `WithScalar(k, v)` | Sets metadata on the Group/Collection **itself** |
| `WithLabelOnEach(k, v)` / `WithScalarOnEach(k, v)` | Sets metadata on each **contained element** |
| `RemoveLabelOnEach(names...)` / `RemoveScalarOnEach(names...)` | Removes metadata from each **contained element** |

For `signal.Group` (fully CoW): `WithLabel`/`WithScalar` return a new group with the group's own metadata cloned then updated. Batch methods also return a new group and preserve the group's own metadata via `copyGroupMeta`.
For other groups/collections (mutating): `WithLabel`/`WithScalar` mutate the receiver in place and return it. Batch methods do the same.

## Constructor options

`WithLabel(k, v)` and `WithScalar(k, v)` are `Option` functions available for all
constructors that accept options (`fmesh.New`, `component.New`, `port.NewInput`, `port.NewOutput`).

Constructor options use the `With` prefix and are passed to `New(...)`:

| Capability | Option |
|---|---|
| Activation function | `component.WithActivationFunc(f)` |
| Component description | `component.WithDescription(s)` |
| Initial state | `component.WithInitialState(fn)` |
| Logger | `component.WithLogger(l)` |

Post-construction `Set*` methods exist only where mutation is genuinely required after `New()` returns — currently `SetLogger`, `SetParentMesh`, and `InheritLogger` (called by `fmesh.AddComponents`; sets the mesh logger unless a custom one was set).

Mutating methods that *append* use `Add*` even on result types: `ActivationResult.AddActivationError` (appends to the error list).

## Collection/group operations

`Any(p)`, `Every(p)`, `Count(p)`, `Map`, `MapIf`, `Filter`, `ForEach`, `ForEachIf`, `Reduce`,
`ReducePayloads`, `Join`.

## Error returns

Methods that can fail return `error` as their last return value. Methods that are genuinely
infallible (e.g. `Filter`, `Map`, `signal.Signal` builders) keep their fluent return type — no
need to wrap `error` where nothing can go wrong.

`ForEach` on all collection types returns `error` (stops on first).

## Predicates

Prefer combinators over inline closures: `Not`, `And`, `Or`, `HasLabel`, `LabelEquals`,
`LabelContains`, `HasAllLabels`, `HasAnyLabel`.

## Stuttering

Do not repeat the package name in a type or function name. Within the `meta` package, use
`Predicate`, `Mapper`, `ScalarPredicate` (not `LabelPredicate` / `MetaPredicate`). Within the
`component` package, use `ResultPredicate` and `ResultMapper` for activation-result types (not
`ActivationResultPredicate` / `ActivationResultMapper`).
