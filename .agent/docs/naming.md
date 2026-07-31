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

**There is no exception to this.** Mutating types (`FMesh`, `Component`, `Port`, the collections,
`cycle.Cycle`) carry **no** metadata methods at all — `Labels()` and `Scalars()` return the live
store, and you mutate that: `fm.Labels().Set(k, v)`, not `fm.AddLabel(k, v)`. Roughly 40 delegating
methods were deleted to make the rule absolute, and chaining did not disappear with them — it moved
to `*meta.Labels`, which is where it belongs.

Use `Set` for **everything else** that is a plain `field = value; return receiver` mutating method, whether exported or unexported:
- Exported example: `cycle.SetNumber`, `component.SetLogger` (marks the logger as custom so the mesh never overrides it)
- Unexported example: `port.setSignals`, `port.setPorts`

**No dual-form duplication**: if a capability has a `With*` constructor option, do **not** also add a `Set*` method for the same capability, unless genuine post-construction mutation is needed. Logger is the one exception: `component.WithLogger`/`component.SetLogger` mark the logger custom, and `component.InheritLogger` (called by `fmesh.AddComponents`) sets the mesh logger only on components without a custom one.

## Metadata operations

| Type kind | How metadata is changed |
|---|---|
| CoW (`signal.Signal`, `signal.Group`) | `WithLabel`/`WithLabels`/`WithOnlyLabels`/`WithoutLabels`/`WithNoLabels`, and the `*Scalar*` equivalents. These return a **new value** — they are the only way to get a modified signal, so they are load-bearing. |
| Mutating (`fmesh.FMesh`, `component.Component`, `port.Port`, the collections, `cycle.Cycle`) | No methods. Go through the store: `x.Labels().Set(k, v)`, `.SetMany(m)`, `.Remove(names...)`, `.Clear()`, and the same on `x.Scalars()`. |

Replacing all labels is `x.Labels().Clear().SetMany(m)` — two calls, and it reads as what it does.

`meta.Labels` (mutating): `Set`, `SetMany`, `Remove`, `Clear`. `Merge(other)` returns a new collection.
`meta.Scalars` (mutating): `Set`, `SetMany`, `Remove`, `Clear`, `Scale`. `Merge(other)` returns a new collection.

## Group/Collection metadata batch methods

| Method | Effect |
|---|---|
| `WithLabelOnEach(k, v)` / `WithScalarOnEach(k, v)` | **CoW only** (`signal.Group`): returns a new group with the metadata set on each contained signal |
| `SetLabelOnEach(k, v)` / `SetScalarOnEach(k, v)` | **Mutating collections**: sets metadata on each contained element in place |
| `RemoveLabelOnEach(names...)` / `RemoveScalarOnEach(names...)` | Removes metadata from each contained element |

A collection's own metadata is reached the same way as any other mutating type: `c.Labels().Set(k, v)`.

For `signal.Group` (fully CoW): the batch methods return a new group and preserve the group's own metadata via `copyGroupMeta`.

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

Component-level: `component.HasSignalsOn(names...)` returns `func(*Component) bool`, the predicate
`component.When` takes.

## Combinators

Functions that take and return an `ActivationFunc` are named after what they do to the activation,
with no prefix: `Sequential`, `When`, `RequireInputs`, `Pipeline`. `With*` stays reserved for
constructor options, so `WithActivationFunc(Sequential(...))` reads as option-wrapping-value rather
than two options.

## Stuttering

Do not repeat the package name in a type or function name. Within the `meta` package, use
`Predicate`, `Mapper`, `ScalarPredicate` (not `LabelPredicate` / `MetaPredicate`). Within the
`component` package, use `ResultPredicate` and `ResultMapper` for activation-result types (not
`ActivationResultPredicate` / `ActivationResultMapper`).
