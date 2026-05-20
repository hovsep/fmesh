# Signal & Labels Package API Cleanup

## Scope
Only `signal` and `labels` packages get new API. Other packages are adapted (compilation fixes, no public API changes).

---

## signal package

### signal/signal.go
- Add godoc on `cloneForMutation` / `cloneSignal` documenting the shallow-payload contract:
  payload is shallow-copied; mutable reference types (map, slice, pointer) are shared across
  derived signals. Users must treat payload as immutable once placed in a signal.

### signal/types.go
- Add `Reducer func(acc *Signal, s *Signal) *Signal`
- Add `PayloadReducer func(acc any, payload any) any`

### signal/group.go

| Change | Detail |
|---|---|
| Fix `Every` (was `AllMatch`) vacuous truth | Empty group → `true` |
| Fix `Filter` capacity hint | `make(Signals, 0, len(g.signals))` |
| Fix `Map` pointer aliasing | After `result := s.Map(m)`, if `result == s` then `result = cloneSignal(s)` |
| Rename `AnyMatch` → `Any` | |
| Rename `AllMatch` → `Every` | No conflict — `All()` stays as-is |
| Rename `CountMatch` → `Count` | |
| Rename `AddFromPayloads` → `AddPayloads` | Shorter, symmetric with `Add` |
| Remove `Without` | Replaced by `Filter(Not(p))` with predicate combinators |
| Add `Last() *Signal` | Symmetric with `First`; slices are ordered |
| Add `Join(other *Group) *Group` | Merges two groups; receiver unchanged |
| Add `Contains(s *Signal) bool` | Pointer identity check |
| Add `ContainsPayload(payload any) bool` | Uses `==`; comparable types only; documented |
| Add `ContainsPayloadFunc(eq func(any) bool) bool` | For non-comparable types |
| Add `Reduce(initial *Signal, fn Reducer) *Signal` | Full signal accumulation |
| Add `ReducePayloads(initial any, fn PayloadReducer) any` | Payload-only convenience |

### signal/predicates.go (new file)

Combinators:
```go
func Not(p Predicate) Predicate
func And(p1, p2 Predicate) Predicate
func Or(p1, p2 Predicate) Predicate
```

Label-aware constructors (return a Predicate):
```go
func HasLabel(name string) Predicate
func LabelEquals(name, value string) Predicate
func LabelContains(name, substr string) Predicate
func HasAllLabels(names ...string) Predicate
func HasAnyLabel(names ...string) Predicate
```

Usage examples:
```go
g.Filter(And(LabelEquals("env", "prod"), Not(HasLabel("skip"))))
g.Filter(Or(LabelContains("tag", "urgent"), HasAnyLabel("priority", "critical")))
```

### signal/group_test.go
- Rename `AddFromPayloads` → `AddPayloads`
- Remove `Without` tests; replace usage with `Filter(Not(p))`
- Rename `AnyMatch` → `Any`, `AllMatch` → `Every`, `CountMatch` → `Count`
- Update `Every` empty-group test case: `false` → `true`
- Add tests: `Last`, `Join`, `Contains`, `ContainsPayload`, `ContainsPayloadFunc`, `Reduce`, `ReducePayloads`

### signal/predicates_test.go (new file)
- Full coverage for all 8 combinator/constructor functions

### signal/immutability_test.go
- Add case: `Map` with identity mapper verifies no pointer aliasing in output

---

## labels package

### labels/labels.go

| Change | Detail |
|---|---|
| Rename `AnyMatch` → `Any` | Also update internal call in `HasAnyFrom` |
| Rename `AllMatch` → `Every` | Also update internal call in `HasAllFrom` |
| Rename `CountMatch` → `Count` | |
| Fix `ValueIs` double lookup | Single: `v, ok := c.labels[label]; return ok && v == value` |
| Add `Keys() []string` | Returns all label names, sorted for determinism |
| Add `Values() []string` | Returns all label values, sorted by key for determinism |
| Add `Merge(other *Collection) *Collection` | New collection; `other` wins on key conflict; non-mutating on both inputs |

Notes:
- Labels stays **mutable** — cannot change without breaking port.go (out of scope)
- `Every` vacuous truth is already correct in labels (empty → `true`). No logic change, just rename.
- `WithoutIf` not added — use `Filter(Not(pred))` style (labels has its own `Filter`)

### labels/labels_test.go
- Rename all `AnyMatch` → `Any`, `AllMatch` → `Every`, `CountMatch` → `Count` test names and calls
- Add tests: `Keys`, `Values`, `Merge`
- Add `ValueIs` edge cases (key exists with empty value, key absent)

---

## Adapting other packages (compilation fixes only, zero public API changes)

These two call sites reference `signal.Group` methods we are renaming:

| File | Line | Change |
|---|---|---|
| `integration_tests/component_hooks/basic_test.go` | 129 | `.Signals().AnyMatch(` → `.Signals().Any(` |
| `port/port_test.go` | 1051 | `.Signals().AllMatch(` → `.Signals().Every(` |

Note: `AnyMatch`/`AllMatch`/`CountMatch` in `cycle`, `component`, and `port` packages are those
packages' own independent methods on their own types. They are NOT being renamed.

---

## Decisions & rationale (for future reference)

- **No generics** — FBP requires mixed-type signal groups; `any` is the right fit
- **No deep-copy mode on ports** — signals shared by pointer is fmesh design; users treat payload as immutable
- **No Clone() on Signal** — all label-mutating methods already return new signals (copy-on-write)
- **No text DSL for label queries** — predicate combinators are type-safe and composable
- **Labels stays mutable** — port.go uses labels mutably; we cannot change port's public API
- **Without removed from signal.Group** — `Filter(Not(p))` with combinators is cleaner
- **WithoutIf not added to labels** — same reasoning
- **Every (not All) for AllMatch** — `All()` is already taken (returns all signals)
- **Join (not Concat/Merge) for group merging** — Concat is string-flavored; Join is clean
- **ContainsPayloadFunc separate method** — no reflect; users pass their own comparator
- **AllMatch vacuous truth fixed in signal** — empty group returns `true` (standard math semantics)
- **AllMatch vacuous truth already correct in labels** — loop falls through to `return true`
