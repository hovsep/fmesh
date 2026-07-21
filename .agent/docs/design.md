# Design

## Architecture

| Concept | Type | Package |
|---|---|---|
| Data packet | `*Signal` | `signal` |
| Ordered signal collection | `*signal.Group` | `signal` |
| String metadata | `*meta.Labels` | `meta` |
| Numeric metadata | `*meta.Scalars` | `meta` |
| Component | `*component.Component` | `component` |
| Data endpoint | `*port.Port` | `port` |
| Connection | pipe via `PipeTo` (output→input only) | `port` |
| Execution tick | `*cycle.Cycle` | `cycle` |

## Invariants

**`*Signal` and `*signal.Group` are copy-on-write.** Mutating methods return a new value; receiver is never modified. `cloneSignal(s)` is the single clone primitive — nil-safe, use it in all CoW methods. `cloneScalars(s)` / `cloneLabels(c)` are the analogous clone helpers for metadata.

**Payload is shallow-copied.** Mutable reference payloads (map, slice, pointer) must be treated as immutable by the caller. `nil` is a valid payload and must survive all CoW operations unchanged.

**`meta.Labels` and `meta.Scalars` are mutable.** They mutate in place. Do not make them CoW — `port`, `component`, `cycle`, and the Group/Collection types depend on mutation. The one exception is `Merge(other)` on both types, which returns a new value.

**Errors are returned directly.** Methods that can fail return `error` as the last return value. Infallible methods (transformations like `Filter`, `Map`, `With*` on `signal.Signal`) return their type directly for fluency. There is no "poison object" or chainable error field on any type.

**Fan-out shares pointers.** Output→input fan-out forwards the same `*Signal` pointers to all destinations. Do not add deep-copy to `ForwardSignals` or `Flush`.

**No generics in data flow.** `signal`/`meta` use `any`/`float64`; FBP requires mixed-type signal flows in one group. Approved generics elsewhere: `hook.Group[T]` (typed hook registry) and `component.MustGetTyped[T]` (state accessor). Do not add more without approval.

**Minimise `reflect`.** Only when no alternative exists. Current approved use: `reflect.TypeOf(payload).Comparable()` in `ContainsPayload` — always nil-guard before calling `.Comparable()`.

## Package notes

- **`signal`** — `payload` is `[]any{value}` (single-element slice so `nil` is valid). Predicate combinators and label constructors live in `predicates.go`. `ForEach` returns `(*Group, error)`.
- **`meta`** — `Labels` (string k/v) and `Scalars` (string→float64). `Keys()`/`Values()` return sorted slices for determinism. `Merge(other)` is the one non-mutating method on both types. `Every(pred)` on empty = `true` (vacuous truth). `ForEach` returns `error`. Constructors: `NewLabels()`, `NewScalars()`.
- **`port`** — `Flush()` fans out then clears source. `PipeTo` is output→input only. Both return `error`. `PipeTo` validates direction at call time.
- **`component`** — `State` is `map[string]any`, persistent across cycles and across `Run`s (see [runtime.md](runtime.md)). Constructors use functional options: `component.New(name, opts...) (*Component, error)`. Ports come in two creation styles: name-based (`WithInputs`/`AddInputs`, `WithIndexedInputs("i", 1, 3)` → `i1..i3`) and attach-based (`AttachInputPorts` for pre-built `port.NewInput` ports with options). `LoopbackPipe(out, in)` wires a component to itself (such a mesh never stops naturally). `ErrWaitingForInputs`/`ErrWaitingForInputsKeep` are scheduler control-flow sentinels, not failures.
- **`hook`** — generic `hook.Group[T]`, ordered, fail-fast `Trigger`. Three hook levels (mesh/component/port); see [hooks.md](hooks.md).
- **`cycle`** — has its own `Any`/`Every`/`Count` on its collection type, independent of `signal.Group`.

## Metadata tiers on groups/collections

Every Group and Collection type carries its **own** `*meta.Labels` and `*meta.Scalars` (Tier 1), plus batch mutation of its **contents** (Tier 2a). `signal.Group` additionally exposes cross-entity scalar aggregation (Tier 2b).

| Tier | Methods | Where |
|---|---|---|
| 1 — entity's own | `Labels()`, `Scalars()`, `WithLabel(k,v)`, `WithScalar(k,v)` | all groups/collections |
| 2a — batch on contents | `WithLabelOnEach(k,v)`, `WithScalarOnEach(k,v)`, `RemoveLabelOnEach(names...)`, `RemoveScalarOnEach(names...)` | all groups/collections |
| 2b — cross-entity aggregation | `MinScalar(name)`, `MaxScalar(name)`, `AvgScalar(name)`, `SumScalar(name)` | `signal.Group` only |

`signal.Group` batch methods (Tier 2a) preserve the group's own metadata on the returned group via `copyGroupMeta`. `MinScalar`/`MaxScalar`/`AvgScalar` return `(float64, error)` with the sentinel `signal.ErrScalarNotFoundInGroup` when no element has the named scalar; `SumScalar` always returns `float64` (0 when absent).

## Comment hygiene

Comments must add information beyond the signature. Omit a comment entirely rather than restate what the name already says.

**Omit comments on:**
- Private builder methods whose name is self-explanatory (e.g. `newActivationResultOK`)
- One-line setter bodies (e.g. `p.signals = sg`)
- Constructors where the doc would only paraphrase the function name

**Keep/write comments on:**
- Exported types and functions (required by Go doc convention)
- Non-obvious invariants, edge cases, or design constraints
- Anything that would surprise a reader unfamiliar with the decision

## Dead code policy

Do not keep unused exported symbols "for future use". Remove them immediately:
- Exported error vars with no callers (e.g. `ErrInvalidSignal`)
- Named slice type aliases that carry no methods (e.g. `type Components []*Component`)
- Unreachable branches (e.g. a second `if len(x) == 0` guard after the first already returned)

These create noise, mislead readers, and rot silently as the surrounding code evolves.
