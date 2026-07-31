# Design

Architecture overview (the concept → type → package table and the execution loop) lives in
`CLAUDE.md`; this doc covers the invariants and per-package rules behind it.

## Invariants

**`*Signal` and `*signal.Group` are copy-on-write.** Mutating methods return a new value; receiver is never modified. `cloneSignal(s)` is the single clone primitive — nil-safe, use it in all CoW methods. `cloneScalars(s)` / `cloneLabels(c)` are the analogous clone helpers for metadata.

**Payload is shallow-copied.** Mutable reference payloads (map, slice, pointer) must be treated as immutable by the caller. `nil` is a valid payload and must survive all CoW operations unchanged.

**`Signal.Payload()` cannot fail.** It returns `any`. Since `nil` is a valid payload, the only way to hold a signal without one is to build the zero value instead of calling `New` — a construction bug, not a runtime condition, and not worth an error return on the most-called accessor in the library. Such a signal reads as `nil`. Type checking is `signal.As[T]`; "was there a signal at all" is `Group.First() == nil`. `Group.FirstPayload` **keeps** its error, because an empty group is a real runtime state rather than a construction bug — do not collapse it.

**`meta.Labels` and `meta.Scalars` are mutable.** They mutate in place. Do not make them CoW — `port`, `component`, `cycle`, and the Group/Collection types depend on mutation. The one exception is `Merge(other)` on both types, which returns a new value.

**Errors are returned directly.** Methods that can fail return `error` as the last return value. Infallible methods (transformations like `Filter`, `Map`, `With*` on `signal.Signal`) return their type directly for fluency. There is no "poison object" or chainable error field on any type.

**Runs are deterministic.** Given deterministic activation functions, a mesh produces identical
output for identical input, every time. Three orderings uphold this, and nothing may introduce a
fourth source of order:

1. **Within one port** — signals keep insertion order (FIFO). `signal.Group` is an ordered slice.
2. **Multiple upstreams into one port** — arrival follows upstream **component-name** order, because
   `drainComponents` iterates `component.Collection.AllOrdered()`.
3. **Across the ports of one component** — traversal follows **port-name** order, because every
   `port.Collection` traversal goes through `AllOrdered()`.

The third one was map iteration order until it was fixed: `Inputs().Signals()` returned four
different orders across 200 identical runs, and the order output ports flush in decides what a
shared downstream port receives. **Never range over `c.ports` directly** — inside the package range over `c.each` (allocation-free,
no per-port map lookup), outside it use `AllOrdered()`. The collection keeps the sorted `[]*Port`
alongside the name map and rebuilds it on membership change, not lazily on read: ports are read from
activation goroutines, and a lazily filled cache would turn a read into a write. Traversal must stay
allocation-free — `port/collection_bench_test.go` guards that.

Note what this does *not* promise: components within a cycle activate concurrently, so the order
their activation functions *run* in is unspecified and always will be. Determinism comes from the
order signals are *collected and delivered*, which is why activation functions must not depend on
shared mutable state.

**Fan-out shares pointers.** Output→input fan-out forwards the same `*Signal` pointers to all destinations. Do not add deep-copy to `ForwardSignals` or `Flush`.

**No generics in data flow.** `signal`/`meta` use `any`/`float64`; FBP requires mixed-type signal flows in one group. Approved generics elsewhere: `hook.Group[T]` (typed hook registry), `component.MustGetTyped[T]` (state accessor), and `signal.As[T]`/`signal.AsOrDefault[T]` (payload accessors — they read a payload out, they do not make the flow typed). Do not add more without approval.

**Minimise `reflect`.** Only when no alternative exists. Current approved use: `reflect.TypeOf(payload).Comparable()` in `ContainsPayload` — always nil-guard before calling `.Comparable()`.

## Package notes

- **`signal`** — `payload` is `[]any{value}` (single-element slice so `nil` is valid). Predicate combinators and label constructors live in `predicates.go`. `ForEach`/`ForEachIf` return `error` only (as on every collection type — see [naming.md](naming.md)). Typed payload accessors live in `typed.go`: `As[T]` (error on nil signal / missing payload / wrong type), `AsOrDefault[T]`, fallible per-type shorthands over `As`, and `AsNumber` (loose `(float64, bool)` widening — `float64`/`float32`/`int`/`int64`/`uint64`, `bool` as 1/0). None of them panic; that is the point of having them. There are deliberately **no** `AsIntOrDefault`-style shorthands: `AsOrDefault` infers `T` from the default and is shorter. The sole exception is `AsFloat64OrDefault`, which exists because an untyped `0` infers `int`, so `AsOrDefault(s, 0)` silently returns the default for a float64 payload — do not "restore symmetry" by adding the others back.
- **`meta`** — `Labels` (string k/v) and `Scalars` (string→float64). `Keys()`/`Values()` return sorted slices for determinism. `Merge(other)` is the one non-mutating method on both types. `Every(pred)` on empty = `true` (vacuous truth). `ForEach` returns `error`. Constructors: `NewLabels()`, `NewScalars()`.
- **`port`** — `Flush()` fans out then clears source. `PipeTo` is output→input only. Both return `error`. `PipeTo` validates direction at call time. `wiring.go` holds the declarative multi-edge helpers: `Pipe`/`MultiPipe` (registers connections) and `Pair`/`MultiForward` (copies signals now); both name the failing edge and report nil ports instead of dereferencing them.
- **Name lookups are silently forgiving — helpers taking port names must not be.** `Collection.ByName` returns `nil` for a name no port has, `Collection.ByNames` skips such names entirely, and `AllHaveSignals()`/`Every()` on the resulting empty collection is vacuously `true`. So `ByNames("typo").AllHaveSignals()` reports *ready*. Any helper that accepts port names as strings must resolve every name before asking anything about signals, and report the name it could not resolve.
- **`component`** — `State` is `map[string]any`, persistent across cycles and across `Run`s (see [runtime.md](runtime.md)). Constructors use functional options: `component.New(name, opts...) (*Component, error)`. Ports come in two creation styles: name-based (`WithInputs`/`AddInputs`, `WithIndexedInputs("i", 1, 3)` → `i1..i3`) and attach-based (`AttachInputPorts` for pre-built `port.NewInput` ports with options). `LoopbackPipe(out, in)` wires a component to itself (such a mesh never stops naturally). `ErrWaitingForInputs`/`ErrWaitKeepingInputs` are scheduler control-flow sentinels, not failures. `compose.go` holds the `ActivationFunc` combinators — `Sequential`, `When`+`HasSignalsOn`, `RequireInputs`, `Pipeline`+`PipelineStage` — which compose a component's *own* activation (contrast `OnActivation` hooks, which are for behavior added from outside; see [hooks.md](hooks.md)). `When` skips, `RequireInputs` suspends and keeps: never substitute one for the other, as skipping where waiting was meant drops the partial inputs at drain time.
- **`hook`** — lives at `internal/hook` (not public API). Generic `hook.Group[T]`, ordered, fail-fast `Trigger`. Three hook levels (mesh/component/port); see [hooks.md](hooks.md).
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

**Style:**
- Type and package-level comments state **what the type is**, not how its methods work — method names go stale
- No usage guidance ("Use X to do Y") or examples in type definition comments; those belong in method godocs or external docs
- Method comments: one line where possible

**Length — short and on point:**

One line is the default. A rationale that genuinely needs more gets a second short paragraph of
two or three sentences, and that is the ceiling.

- **State the constraint, not the story.** "A name no port has fails the activation instead of
  suspending forever" beats a paragraph reconstructing how a reader might get it wrong.
- **No file-header essays.** A file-level comment names what the file holds in a sentence. If a
  file needs several paragraphs to introduce itself, the material is documentation, not a comment.
- **A paragraph belongs in `docs/wiki/`.** The wiki teaches — narrative, examples, when-to-use-which
  tables. Godoc reminds. Name the concept in the comment and let the page carry it; duplicating it
  in source means two things to keep in sync, and the source copy is the one that rots.
- **Cut the prose voice.** Comments that argue with the reader ("which is the right number", "and
  that is the point", "the whole reason this exists") are essay, not documentation.
- Explain the **non-obvious**: an invariant, a footgun, an ordering requirement, a reason the
  obvious implementation is wrong. Never narrate what the next line plainly does.

## Dead code policy

Do not keep unused **unexported** symbols "for future use". Remove them immediately:
- Named slice type aliases that carry no methods (e.g. `type Components []*Component`)
- Unreachable branches (e.g. a second `if len(x) == 0` guard after the first already returned)
- Private helpers with no caller

These create noise, mislead readers, and rot silently as the surrounding code evolves.

**Exported symbols are different, and this policy does not cover them.** F-Mesh is a library: its
callers live in `fmesh-examples` (five separate Go modules) and `fmesh-graphviz`, which no analysis
of this repo can see. "No in-repo caller" is not evidence a public symbol is dead — it is the normal
state of a public API. Before removing anything exported, compile the downstream repos and get the
user's decision; see [downstream.md](downstream.md), which records the six symbols a previous
cleanup removed on exactly that mistaken reasoning.

Exported **error sentinels** are the same trap in miniature: "nothing calls `errors.Is` on it in
this repo" is expected, because users are the ones who check them.
