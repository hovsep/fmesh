# Plan: meta package + Scalars

## Goal

1. Rename `labels` package → `meta` (resolves `labels.Labels` redundancy → `meta.Labels`)
2. Add `meta.Scalars` — a `map[string]float64` numeric metadata store
3. Attach `Labels` + `Scalars` to every object in the fmesh ecosystem
4. Two-tier API on groups/collections: entity's own metadata + batch mutation of contents
5. Cross-entity scalar aggregation on `signal.Group` and `port.Group`

## Naming rules

| Concept | Name |
|---|---|
| Package | `meta` |
| String metadata type | `meta.Labels` |
| Numeric metadata type | `meta.Scalars` |
| Constructors | `meta.NewLabels()`, `meta.NewScalars()` |
| Entity's own metadata | `obj.Labels()`, `obj.Scalars()`, `obj.WithLabel(k,v)`, `obj.WithScalar(k,v)` |
| Batch on contents | `obj.WithLabelOnEach(k,v)`, `obj.WithScalarOnEach(k,v)`, `obj.RemoveLabelOnEach(names...)`, `obj.RemoveScalarOnEach(names...)` |
| Cross-entity aggregation | `group.MinScalar(name)`, `group.MaxScalar(name)`, `group.AvgScalar(name)`, `group.SumScalar(name)` |

## Phase 1 — Rename labels → meta

- Rename directory `labels/` → `meta/`
- Module path: `github.com/hovsep/fmesh/labels` → `github.com/hovsep/fmesh/meta`
- Rename constructor: `New()` → `NewLabels()`
- Type `Labels` and all methods unchanged
- Update all imports (~12 files, mechanical)

## Phase 2 — meta.Scalars

New file `meta/scalars.go`.

```go
type Scalars struct { scalars map[string]float64 }

func NewScalars() *Scalars

// CRUD
Set(name string, v float64) *Scalars
SetMany(map[string]float64) *Scalars
Get(name string) (float64, bool)
GetOrDefault(name string, def float64) float64
Has(name string) bool
Remove(names ...string) *Scalars
Clear() *Scalars

// Inspection
All() map[string]float64
Keys() []string
Len() int
IsEmpty() bool

// Math
Min() (name string, v float64, ok bool)
Max() (name string, v float64, ok bool)
Sum(names ...string) float64            // all keys if none given
Average(names ...string) (float64, bool)
Scale(name string, factor float64) *Scalars

// Functional
Every(func(string, float64) bool) bool
Any(func(string, float64) bool) bool
Count(func(string, float64) bool) int
Filter(func(string, float64) bool) *Scalars
ForEach(func(string, float64) error) error
```

## Phase 3a — Add Scalars to Signal, Component, Port

All three already have Labels. Add `scalars *meta.Scalars` field.

**signal.Signal** (CoW style — matches existing label API):
```
WithScalar(name string, v float64) *Signal
WithScalars(map[string]float64) *Signal       // merge
WithOnlyScalars(map[string]float64) *Signal   // replace all
WithNoScalars() *Signal
WithoutScalars(names ...string) *Signal
Scalars() *meta.Scalars
```

**component.Component** (mutating style):
```
WithScalar(name string, v float64) *Component   // also Option
SetScalars(map[string]float64) *Component
AddScalars(map[string]float64) *Component
ClearScalars() *Component
RemoveScalars(names ...string) *Component
Scalars() *meta.Scalars
```

**port.Port** (mutating style):
```
WithScalar(name string, v float64) *Port       // also Option for NewInput/NewOutput
SetScalars(map[string]float64) *Port
AddScalars(map[string]float64) *Port
ClearScalars() *Port
RemoveScalars(names ...string) *Port
Scalars() *meta.Scalars
```

## Phase 3b — Add Labels + Scalars to FMesh

New fields: `labels *meta.Labels`, `scalars *meta.Scalars`
Also add `WithLabel` and `WithScalar` as constructor Options.
Mutating style matching Component/Port.

## Phase 3c — Add Labels + Scalars + batch methods to all Groups/Collections

Applies to: `signal.Group`, `port.Group`, `port.Collection`, `component.Collection`, `cycle.Group`

**Tier 1 — entity's own:**
```
Labels() *meta.Labels
Scalars() *meta.Scalars
WithLabel(k, v string) *T
WithScalar(k string, v float64) *T
```

**Tier 2 — batch on contents:**
```
WithLabelOnEach(k, v string) *T
WithScalarOnEach(k string, v float64) *T
RemoveLabelOnEach(names ...string) *T
RemoveScalarOnEach(names ...string) *T
```

## Phase 3d — Cross-entity aggregation on signal.Group and port.Group

```
MinScalar(name string) (float64, bool)
MaxScalar(name string) (float64, bool)
AvgScalar(name string) (float64, bool)
SumScalar(name string) float64
```

These iterate over contained elements, collect values for `name`, and aggregate.
`(float64, bool)` — bool is false if no element has that scalar name.
`SumScalar` returns 0 if no element has that name (sum of empty set = 0).

## Phase 4 — Tests

- `meta/scalars_test.go` — comprehensive unit tests for all Scalars methods
- `meta/labels_test.go` — update constructor call `New()` → `NewLabels()`
- All other test files: update `labels` import → `meta`, `labels.NewLabels()` → `meta.NewLabels()`
- Add scalar tests to `signal/signal_test.go`, `port/port_test.go`, `component/component_test.go`
- Add labels+scalars tests to group/collection test files
- New `integration_tests/meta/` — end-to-end test demonstrating scalars on signals flowing through a mesh

## Phase 5 — make test && make lint && make fmt

## Key decisions

- Value type: `float64` only (no `any`) — enables all math methods
- No `Axis` projection type — the scalar name IS the axis; `MinScalar("temp")` is sufficient
- No value-based shorthand methods (`Above`, `Below`) — `Filter` covers these
- Batch mutation via `WithLabelOnEach`/`WithScalarOnEach` shorthand AND existing `ForEach`
- Cross-entity aggregation only on `signal.Group` and `port.Group` (not `port.Collection`, `component.Collection`, `cycle.Group`)
- `cycle.Group` gets Labels+Scalars as entity (it's a Group) but no aggregation methods
- Constructor options `WithLabel`/`WithScalar` added to `fmesh.New`, `component.New`, `port.NewInput`, `port.NewOutput`
