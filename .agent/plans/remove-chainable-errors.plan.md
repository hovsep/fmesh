# Remove Chainable Error Machinery

## Goal

Replace the "poison object" chainable error pattern with idiomatic Go error returns.
Infallible methods keep fluent return types. Fallible methods return `error`.
Constructors use functional options.

## Decisions

- `ForEach` on all collections: stop on first error, return `error`
- `signal.Group.ForEach`: returns `(*Group, error)` — processed group + error
- `signal.Signal` builders (`WithLabel`, `MapPayload`, etc.): stay `*Signal` (genuinely infallible)
- `labels.Collection` writers (`Add`, `Remove`, `Clear`, `Filter`, `Map`, `Merge`): stay `*Collection` (infallible)
- `port.Collection` / `port.Group` / `component.Collection` infallible transforms (`Filter`, `Map`, `Without`): stay fluent
- No `Must` helpers in production or test code
- `NewWithConfig` removed — replaced by `New(name, WithConfig(cfg))`

---

## Phase 1 — Remove chainable error machinery

### 1.1 Remove three-method contract from all types

Remove `chainableErr error` field, `WithChainableErr`, `HasChainableErr`, `ChainableErr` from:

| Type | File |
|---|---|
| `FMesh` | `fmesh.go` |
| `Component` | `component/component.go` |
| `Port` | `port/port.go` |
| `Signal` | `signal/signal.go` |
| `signal.Group` | `signal/group.go` |
| `port.Collection` | `port/collection.go` |
| `port.Group` | `port/group.go` |
| `component.Collection` | `component/collection.go` |
| `labels.Collection` | `labels/labels.go` |
| `cycle.Cycle` | `cycle/cycle.go` |
| `cycle.Group` | `cycle/group.go` |
| `ActivationResult` | `component/activation_result.go` |
| `ActivationResultCollection` | `component/activation_result_collection.go` |

### 1.2 Remove propagateChainErrors

Remove `propagateChainErrors()` from `component/component.go:116`.
Remove its call site in `MaybeActivate()`.

### 1.3 Remove All() chainable error guards

All `All()` methods currently guard with `if c.HasChainableErr() { return nil, c.chainableErr }`.
Remove these guards — `All()` returns the map/slice directly.

### 1.4 Delete chainable error integration test

Delete `integration_tests/errorhandling/chainable_api_test.go`.
Replace with a new `integration_tests/errorhandling/error_returns_test.go` that tests:
- Error returned from `fmesh.New` with invalid options
- Error returned from `port.PipeTo` with wrong direction
- Error returned from `component.Collection.Add` on duplicate name
- Error propagated all the way to `Run()` via standard `error` wrapping

---

## Phase 2 — Functional options for constructors

### 2.1 `component` package

```go
type Option func(*Component) error

func New(name string, opts ...Option) (*Component, error)

func WithDescription(s string) Option
func WithInputs(names ...string) Option
func WithOutputs(names ...string) Option
func WithIndexedInputs(prefix string, start, end int) Option
func WithIndexedOutputs(prefix string, start, end int) Option
func WithActivationFunc(f ActivationFunc) Option
func WithHooks(setup func(*Hooks)) Option
func WithLabels(m labels.Map) Option
```

### 2.2 `port` package

```go
type Option func(*Port) error

func NewInput(name string, opts ...Option) (*Port, error)
func NewOutput(name string, opts ...Option) (*Port, error)

func WithDescription(s string) Option
func WithLabel(k, v string) Option
```

### 2.3 `fmesh` package

```go
type Option func(*FMesh) error

func New(name string, opts ...Option) (*FMesh, error)
// NewWithConfig removed — use New(name, WithConfig(cfg))

func WithComponents(cs ...*component.Component) Option
func WithConfig(cfg *Config) Option
func WithHooks(setup func(*Hooks)) Option
func WithDescription(s string) Option
```

### 2.4 Update all call sites

Every test and integration test that calls `component.New("x").AddInputs(...)` etc. must be
updated to `component.New("x", component.WithInputs(...))`.

---

## Phase 3 — Mutation / operational methods return `error`

### 3.1 `port.Port`

| Method | Before | After |
|---|---|---|
| `PipeTo(dests ...*Port)` | `*Port` | `error` |
| `PutSignals(sigs ...*Signal)` | `*Port` | `error` |
| `PutPayloads(payloads ...any)` | `*Port` | `error` |
| `PutSignalGroups(groups ...*signal.Group)` | `*Port` | `error` |
| `Flush()` | `*Port` | `error` |
| `Clear()` | `*Port` | `error` |

### 3.2 `port.Collection`

| Method | Before | After |
|---|---|---|
| `Add(ports ...*Port)` | `*Collection` | `error` |
| `PipeTo(dests ...*Port)` | `*Collection` | `error` |
| `Flush()` | `*Collection` | `error` |
| `PutSignals(sigs ...*Signal)` | `*Collection` | `error` |
| `ForEach(func(*Port) error)` | `*Collection` | `error` |

### 3.3 `port.Group`

| Method | Before | After |
|---|---|---|
| `Add(ports ...*Port)` | `*Group` | `error` |
| `ForEach(func(*Port) error)` | `*Group` | `error` |
| `ForEachIf(Predicate, func(*Port) error)` | `*Group` | `error` |

### 3.4 `component.Collection`

| Method | Before | After |
|---|---|---|
| `Add(cs ...*Component)` | `*Collection` | `error` |
| `ForEach(func(*Component) error)` | `*Collection` | `error` |

### 3.5 `fmesh`

| Method | Before | After |
|---|---|---|
| `AddComponents(cs ...*Component)` | `*FMesh` | `error` |

### 3.6 `signal.Group`

| Method | Before | After |
|---|---|---|
| `ForEach(func(*Signal) error)` | `*Group` | `(*Group, error)` |

### 3.7 `labels.Collection`

| Method | Before | After |
|---|---|---|
| `ForEach(func(k, v string) error)` | `*Collection` | `error` |

### 3.8 `component`

| Method | Before | After |
|---|---|---|
| `LoopbackPipe(out, in string)` | `(void)` | `error` |

### 3.9 Internal runtime wiring

Update `fmesh.go` and `component/io.go` to thread errors from `ForEach` / `Flush` /
`AddComponents` through normally. Specifically:
- `fmesh.go` cycle loop: propagate `ForEach` errors through `Run()` return
- `component/io.go:FlushOutputs`: return error from `Outputs().Flush()`
- `fmesh.go:validateBeforeRun`: use `ForEach` returning `error` instead of chainable guard

---

## Phase 4 — Renames, naming consistency, secondary issues

### 4a. Rename `AnyMatch`/`AllMatch`/`CountMatch` → `Any`/`Every`/`Count`

Files:
- `port/collection.go`
- `port/group.go`
- `cycle/group.go`
- `component/collection.go`
- `component/activation_result_collection.go`

Update all call sites including integration tests.
This completes the migration tracked in `naming.md`.

### 4b. Fix `Direction` underlying type

Change `type Direction bool` to `type Direction int` with iota:

```go
type Direction int

const (
    DirectionUndefined Direction = iota
    DirectionIn
    DirectionOut
)
```

Zero value becomes `DirectionUndefined` (safe default) instead of `DirectionOut` (surprising).
Update all comparisons and port constructors.

### 4c. Remove stutter: `labels.LabelPredicate` / `labels.LabelMapper`

Rename in `labels/types.go`:
- `LabelPredicate` → `Predicate`
- `LabelMapper` → `Mapper`

Update all call sites.

### 4d. Fix `ActivationResultPredicate` / `ActivationResultMapper` naming

In `component/types.go`, align with the `Predicate`/`Mapper` names already used for `Component`:
- `ActivationResultPredicate` → `ResultPredicate`
- `ActivationResultMapper` → `ResultMapper`

Update all call sites in `component/activation_result_collection.go`.

### 4e. Fix `With…` prefix on mutating methods

`WithDescription`, `WithParentComponent`, `WithParentMesh` mutate the receiver but use the
`With` prefix (which per naming.md means CoW). After Phase 2 these are no longer in the public
API. Rename remaining internal uses:
- `WithDescription` → `setDescription` (unexported, used only internally after functional options land)
- `WithParentComponent` → `setParentComponent` (unexported)
- `WithParentMesh` → `setParentMesh` (unexported)

### 4f. Fix error strings starting with capital letters

`component/io.go:60` and `component/io.go:90` — lowercase the leading function-name prefix:

```go
// Before:
fmt.Errorf("AttachInputPorts: port '%s' is not an input port ...")
// After:
fmt.Errorf("port %q is not an input port (use port.NewInput): %w", ...)
```

### 4g. Rename `GetTyped[T]` → `MustGetTyped[T]`

`component/state.go:69,73` — exported function panics without `Must` in the name.
Rename to `MustGetTyped[T]` to communicate Must-semantics to callers.
Update all call sites.

### 4h. Fix `ContainsPayload` panic

`signal/group.go:117` — exported method panics on non-comparable payload without warning.
Split into:
- `ContainsPayload(payload any) (bool, error)` — safe, returns error for non-comparable types
- `MustContainsPayload(payload any) bool` — panics, for callers that know payload is comparable

Existing call sites switch to `MustContainsPayload`.

### 4i. Resolve @TODO comments

- `component/component.go:28` — decide on name validation; remove TODO regardless
- `fmesh.go:229` — decide on clearing outputs on cycle error; remove TODO regardless
- `fmesh.go:266` — decide on fine-grained port keep control; remove TODO regardless

---

## Checklist

- [ ] Phase 1: Remove chainable error machinery from all 13 types
- [ ] Phase 1: Delete `chainable_api_test.go`, add `error_returns_test.go`
- [ ] Phase 2: Functional options on `component.New`, `port.NewInput`, `port.NewOutput`, `fmesh.New`
- [ ] Phase 2: Remove `NewWithConfig`; update all call sites
- [ ] Phase 3: All method signature changes + call site updates
- [ ] Phase 3: Runtime wiring in `fmesh.go` and `component/io.go`
- [ ] Phase 4a: `AnyMatch`/`AllMatch`/`CountMatch` → `Any`/`Every`/`Count` in 5 files
- [ ] Phase 4b: `Direction` underlying type to `int`/iota
- [ ] Phase 4c: `LabelPredicate`/`LabelMapper` stutter
- [ ] Phase 4d: `ActivationResultPredicate`/`ActivationResultMapper` rename
- [ ] Phase 4e: `With…` → `set…` for internal mutating methods
- [ ] Phase 4f: Error string casing
- [ ] Phase 4g: `GetTyped` → `MustGetTyped`
- [ ] Phase 4h: `ContainsPayload` safe + Must split
- [ ] Phase 4i: Resolve @TODO comments
- [ ] `make test && make lint && make fmt` green after each phase
