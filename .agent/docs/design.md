# Design

The concept → type → package table and the execution loop are in `CLAUDE.md`. This doc holds the
invariants and per-package rules behind them.

## Invariants

**`*signal.Signal` and `*signal.Group` are copy-on-write.** Mutating methods return a new value and
never modify the receiver. `cloneSignal(s)` is the one clone primitive (nil-safe) — use it in every
CoW method. `meta.Meta.Clone()` (nil-receiver safe) is the same for metadata.

**Payload is shallow-copied.** Callers must treat reference payloads (map, slice, pointer) as
immutable. `nil` is a valid payload and must survive every CoW operation unchanged.

**`Signal.Payload()` cannot fail.** It returns `any`. A signal without a payload can only come from
building the zero value instead of calling `New` — a construction bug, not worth an error on the
most-called accessor. It reads as `nil`.
- Type checks: `Signal.As[T]()`, or `Group.FirstAs[T]()` straight from a port.
- "Was there a signal at all": `Group.First() == nil`.
- `Group.FirstPayload` **keeps** its error: an empty group is a real runtime state. Do not collapse it.

**`meta.Meta` mutates in place.** Do not make it CoW — `port`, `component`, `cycle` and the
group/collection types depend on mutation. Only `Clone()` and `Filter(pred)` return a new value.

**Errors are returned directly.** Fallible methods return `error` last. Infallible transforms
(`Filter`, `Map`, `With*` on `signal.Signal`) return their type directly. No type has a "poison
object" or chainable error field.

**Runs are deterministic.** With deterministic activation functions, the same input always gives
the same output. Three orderings ensure this. Do not add a fourth source of order.

1. **Within one port** — signals keep insertion order (FIFO); `signal.Group` is an ordered slice.
2. **Many upstreams into one port** — arrival follows upstream **component-name** order, because
   `drainComponents` walks `component.Collection.AllOrdered()`.
3. **Across the ports of one component** — traversal follows **port-name** order.

Keyed collections (`port.Collection`, `component.Collection`) embed `internal/collection.Keyed[T]`,
which keeps a name-sorted slice beside the name map. Rules:
- **Never range over the name map.** Map order leaks into flush order and `Signals()` results.
  Inside the packages range over `c.Each` (no allocation, no per-item lookup); outside use
  `AllOrdered()`.
- The sorted slice is rebuilt on membership change, never lazily on read. Ports are read from
  activation goroutines, and a lazy cache would turn a read into a write.
- Traversal must stay allocation-free; `port/collection_bench_test.go` guards it.

Not promised: the order activation functions *run* in. Components in a cycle activate concurrently.
Determinism comes from the order signals are *collected and delivered*, so activation functions
must not depend on shared mutable state.

**Fan-out shares pointers.** An output fans out the same `*Signal` pointers to every destination.
Do not add a deep copy to `ForwardSignals` or `Flush`.

**The signal payload stays `any`.** One group must carry mixed-type signals (an FBP requirement),
so `Signal.Payload()` cannot be generic and pipes cannot be typed. `Signal.As[T]()` and
`PayloadOrDefault[T]()` read a payload out; they do not make the flow typed.

**Generics are fine where they remove real duplication.** Weigh two things first:
- **Per-instance cost.** `meta.Meta` is exactly the map header it wraps. Do not add fields to it
  (a `self` pointer doubled every store to 16 bytes; a name for errors cost ~100 bytes per signal).
- **Godoc.** Methods promoted from an unexported generic type render with unresolved type
  parameters (`All() map[string]T`). A generic *method* on a concrete type renders normally — one
  reason `Meta` is not generic over its value type.

A generic that needs one hand-written forwarding method per call site has usually not paid off.
`internal/hook.Group[T]` sits exactly on that line **on purpose**: 13 hand-written registration
methods across the three hook levels wrap its 3 methods. That is the price of unexported `Hooks`
fields, which keep closures the only registration path. Do not "fix" either side.

`internal/collection` is shared the same way. `Slice[T]` backs the groups and `Keyed[T]` the
name-keyed collections, embedded behind unexported type aliases so the read surface promotes.
Slice plumbing uses package functions (`collection.Items`/`SetItems`/`AppendItems`,
`collection.Reset`), not methods, so no mutator promotes onto `signal.Group`'s CoW surface.

**Generic methods (Go 1.27) are for accessors where the caller picks the type.** Examples:
`Signal.As[T]()`, `Group.FirstAs[T]()`, `Group.ReducePayloads[A]()`,
`Group.ContainsPayload[T comparable]()`, `State.GetTyped[T]()`, `Signal.WithMeta[T]()` and the
typed `meta.Meta` surface (`Set[T]`, `Value[T]`, `ValueOrDefault[T]`, `ValueIs[T]`). They cost
nothing per instance and render normally in godoc. Two limits:
- A generic method cannot implement an interface method.
- It cannot name its receiver's type as a result. So the `Filter`/`Map` duplication across the
  collection types needs a self type parameter, not generic methods — it stays hand-written.

**Minimise `reflect`.** Use it only when no alternative exists. There is currently no use at all.

## Package notes

- **`signal`**
  - `payload` is `[]any{value}` (a one-element slice, so `nil` is valid).
  - Predicate combinators and metadata predicates (`HasMeta`, `HasAnyMeta`, `MetaEquals[T]`,
    `MetaContains`) live in `predicates.go`.
  - `ForEach`/`ForEachIf` return `error` only, as on every collection type ([naming.md](naming.md)).
  - Typed payload accessors are generic methods in `typed.go`: `As[T]()` (error on nil signal,
    missing payload or wrong type), `PayloadOrDefault[T](d)`, `AsGroup()` (a signal carrying a
    group; `As[*Group]()` is noisy from outside) and `AsNumber()` (loose `(float64, bool)`
    widening of `float64`/`float32`/`int`/`int64`/`uint64`, `bool` as 1/0). None panic; all are
    nil-receiver safe. `Group.FirstAs[T]()`/`FirstPayloadOrDefault(d)` do the same over the first
    signal.
  - **No per-type shorthands** (`AsInt`, …): `s.As[int]()` is as short, and `PayloadOrDefault`
    infers `T` from the default. The one exception is `Float64OrDefault`: an untyped `0` infers
    `int`, so `s.PayloadOrDefault(0)` silently returns the default for a float64 payload. Do not
    add the others back for symmetry.
- **`meta`** — one type, `Meta`: a `map[string]any` whose values are `string | float64`, enforced
  by the `Value` constraint on every typed method. Constructor: `New()`.
  - Writes: `Set[T](k, v)`, `SetMany[T](m)`; `T` is inferred. An untyped integer literal is a
    compile error (`int does not satisfy meta.Value`) — intended; write `1.0`.
  - Reads: `Value[T](k)` (error names the key, and both types on a mismatch), `ValueOrDefault(k,
    def)` and `ValueIs(k, v)` (both infer `T`; a wrong-type entry reads as absent), `Has(keys...)`
    (all present; true for none) and `HasAny(keys...)`. `Keys()` is sorted.
  - `Clone()` (nil-safe, used by the CoW types) and `Filter(pred)` (predicate sees the value as
    `any`) return a new store.
  - The constraint is exact on purpose — no `~`, no ints. A named `type Celsius float64` stored via
    `~float64` would fail every `Value[float64]` read. Accepting ints would reopen the
    `PayloadOrDefault(0)` trap. Do not widen it.
  - Deliberately absent: `Values()`, `Map`, `Every`/`Any`/`Count`/`ForEach`, `Merge`, `Has*From`.
    On a mixed-type store each would be untyped or half-typed, and none had a caller.
- **`port`**
  - `Flush()` fans out, then clears the source. It fires `OnSignalsDelivered` on the source once per
    pipe after each destination accepts — the only event naming both ends of a pipe
    (`OnSignalsAdded` fires on the destination and cannot see the sender).
  - Hook context structs on hot paths are built only behind `hook.Group.IsEmpty()`: a `Trigger`
    argument escapes to the heap even with no hooks, which would cost one allocation per pipe, per
    flush, per cycle.
  - `PipeTo` is output→input only, validates direction at call time, and returns `error`.
  - `wiring.go` holds the declarative multi-edge helpers: `MultiPipe(...Pipe)` registers
    connections, `MultiForward(ctx, ...Pair)` copies signals now. Both name the failing edge and
    report nil ports instead of dereferencing them.
- **Name lookups forgive silently — helpers taking port names must not.** `Collection.ByName`
  returns `nil` for an unknown name, `ByNames` skips unknown names, and `AllHaveSignals()`/`Every()`
  on the empty result is vacuously `true`. So `ByNames("typo").AllHaveSignals()` reports *ready*.
  A helper that takes port names must resolve every name first and report the one it cannot.
- **`component`**
  - `State` is `map[string]any` and persists across cycles and `Run`s ([runtime.md](runtime.md)).
  - Constructor: `component.New(name, opts...) (*Component, error)` with functional options.
  - Ports, two styles: by name (`WithInputs`/`AddInputs`, `WithIndexedInputs("i", 1, 3)` →
    `i1..i3`) or by attaching pre-built `port.NewInput` ports (`AttachInputPorts`).
  - `LoopbackPipe(out, in)` wires a component to itself; such a mesh stops naturally only once the
    component stops emitting into the loop.
  - `ErrWaitDroppingInputs`/`ErrWaitKeepingInputs` (both wrap `ErrWaitingForInputs`) are scheduler
    control flow, not failures.
  - `compose.go` holds `ActivationFunc` combinators — `Sequential`, `When` + `HasSignalsOn`,
    `RequireInputs`, `Pipeline` + `PipelineStage` — for a component's *own* activation. Behavior
    added from outside goes through hooks ([hooks.md](hooks.md)).
  - `When` skips; `RequireInputs` suspends and keeps. Never swap one for the other: skipping where
    waiting was meant drops the partial inputs at drain.
- **`hook`** — `internal/hook`, not public API. Generic `hook.Group[T]`: ordered, fail-fast
  `Trigger`. See [hooks.md](hooks.md).
- **`cycle`** — its group has its own `Any`/`Every`/`Count`, independent of `signal.Group`.

## Metadata on groups and collections

| Tier | Methods | Where |
|---|---|---|
| 1 — the container's own store | `Meta()`; change via `WithMeta` on `signal.Group` (CoW), via the live store elsewhere | every group/collection |
| 2a — batch on contents | `WithMetaOnEach(k, v)`, `WithoutMetaOnEach(keys...)` | `signal.Group` only (CoW) |

- Every group and collection carries its **own** `*meta.Meta`.
- Batch methods exist on `signal.Group` only. Mutating collections have none — iterate with
  `ForEach` and use each element's store. There is no cross-entity aggregation tier. Do not add
  either back.
- `signal.Group` batch methods keep the group's own metadata on the result (`copyGroupMeta`).
- **`signal.Group.Meta()` returns a clone**, like `Signal.Meta()`. The only way to a changed group
  is `WithMeta`. Never hand out the live store — it would break CoW.

## Comment hygiene

Comments must add information beyond the signature. If a comment would only restate the name,
omit it.

**Omit comments on:**
- Private builders with a self-explanatory name (e.g. `newActivationResultOK`)
- One-line setter bodies (e.g. `p.signals = sg`)
- Constructors where the doc would only paraphrase the name

**Write comments on:**
- Exported types and functions (Go doc convention)
- Non-obvious invariants, edge cases, design constraints
- Anything that would surprise a reader who does not know the decision

**Style:**
- Type and package comments say **what the type is**, not how its methods work (method names go
  stale).
- No usage guidance ("Use X to do Y") or examples in type comments; those go in method godocs or
  external docs.
- Method comments: one line where possible.

**Length — short and on point.** One line is the default. A rationale that truly needs more gets
a second short paragraph of two or three sentences. That is the ceiling.
- **State the constraint, not the story.** "A name no port has fails the activation instead of
  suspending forever" beats a paragraph on how a reader might get it wrong.
- **No file-header essays.** A file comment names what the file holds in one sentence.
- **Paragraphs belong in `docs/wiki/`.** The wiki teaches; godoc reminds. Name the concept in the
  comment and let the page carry it — a source copy is a second thing to sync, and it rots first.
- **No essay voice** ("which is the right number", "and that is the point").
- Explain the **non-obvious**: an invariant, a footgun, an ordering need, why the obvious code is
  wrong. Never narrate what the next line plainly does.

## Dead code policy

Do not keep unused **unexported** symbols "for future use". Remove them at once:
- Named slice types with no methods (e.g. `type Components []*Component`)
- Unreachable branches (e.g. a second `if len(x) == 0` guard after the first returned)
- Private helpers with no caller

They add noise, mislead readers and rot as the code around them changes.

**This policy does not cover exported symbols.** F-Mesh is a library. Its callers live in
`fmesh-examples` (one Go module) and `fmesh-export`, which no analysis of this repo can see.
"No in-repo caller" is the normal state of public API, not evidence it is dead. Before removing
anything exported, compile the downstream repos and get the user's decision — see
[downstream.md](downstream.md).

Exported **error sentinels** are the same trap: users call `errors.Is` on them, so "no `errors.Is`
in this repo" is expected.
