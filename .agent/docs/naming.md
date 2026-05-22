# Naming

## CoW vs mutating

| Semantics | Prefix |
|---|---|
| Copy-on-write (returns new value) | `With` / `Without` |
| Mutating (modifies receiver) | `Add` / `Remove` / `Set` |

Never mix. When adding a new method: does it return a new value or mutate? Pick the prefix
accordingly. Internal-only mutating helpers use `set…` (unexported).

## Label operations by type

| | CoW (`signal.Signal`) | Mutating (`port.Port`, `component.Component`) |
|---|---|---|
| Add/update one | `WithLabel(k, v)` | `AddLabel(k, v)` |
| Add/update many | `WithLabels(map)` | `AddLabels(map)` |
| Replace all | `WithOnlyLabels(map)` | `SetLabels(map)` |
| Remove specific | `WithoutLabels(names...)` | `RemoveLabels(names...)` |
| Remove all | `WithNoLabels()` | `ClearLabels()` |

`labels.Collection` (mutating): `Add`, `AddMany`, `Remove`, `Clear`. `Merge(other)` is the one
exception — returns a new collection.

## Collection/group operations

`Any(p)`, `Every(p)`, `Count(p)`, `Map`, `MapIf`, `Filter`, `ForEach`, `ForEachIf`, `Reduce`,
`ReducePayloads`, `Join`.

## Error returns

Methods that can fail return `error` as their last return value. Methods that are genuinely
infallible (e.g. `Filter`, `Map`, `signal.Signal` builders) keep their fluent return type — no
need to wrap `error` where nothing can go wrong.

`ForEach` on all collection types returns `error` (stops on first).
`signal.Group.ForEach` returns `(*Group, error)` — the processed group plus any error.

## Predicates

Prefer combinators over inline closures: `Not`, `And`, `Or`, `HasLabel`, `LabelEquals`,
`LabelContains`, `HasAllLabels`, `HasAnyLabel`.

## Stuttering

Do not repeat the package name in a type or function name. Within the `labels` package, use
`Predicate` and `Mapper` (not `LabelPredicate` / `LabelMapper`). Within the `component`
package, use `ResultPredicate` and `ResultMapper` for activation-result types (not
`ActivationResultPredicate` / `ActivationResultMapper`).
