# Naming

## CoW vs mutating

| Semantics | Prefix |
|---|---|
| Copy-on-write (returns new value) | `With` / `Without` |
| Mutating (modifies receiver) | `Add` / `Remove` |

Never mix. When adding a new method: does it return a new value or mutate? Pick the prefix accordingly.

## Label operations by type

| | CoW (`signal.Signal`) | Mutating (`port.Port`, `component.Component`) |
|---|---|---|
| Add/update one | `WithLabel(k, v)` | `AddLabel(k, v)` |
| Add/update many | `WithLabels(map)` | `AddLabels(map)` |
| Replace all | `WithOnlyLabels(map)` | `SetLabels(map)` |
| Remove specific | `WithoutLabels(names...)` | `RemoveLabels(names...)` |
| Remove all | `WithNoLabels()` | `ClearLabels()` |

`labels.Collection` (mutating): `Add`, `AddMany`, `Remove`, `Clear`. `Merge(other)` is the one exception — returns a new collection.

## Collection/group operations

`Any(p)`, `Every(p)`, `Count(p)`, `Map`, `MapIf`, `Filter`, `ForEach`, `ForEachIf`, `Reduce`, `ReducePayloads`, `Join`.

> `cycle`, `port`, `component` collections still use `AnyMatch`/`AllMatch`/`CountMatch` — not yet renamed.

## Predicates

Prefer combinators over inline closures: `Not`, `And`, `Or`, `HasLabel`, `LabelEquals`, `LabelContains`, `HasAllLabels`, `HasAnyLabel`.
