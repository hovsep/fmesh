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

**The signal payload stays `any`.** This is an FBP requirement, not a style preference: one group has to carry mixed-type signals, so `Signal.Payload()` cannot be parameterised and pipes cannot be typed. `signal.As[T]`/`AsOrDefault[T]` read a payload back out; they do not make the flow typed.

**Generics are otherwise fine — use them where they remove real duplication.** The earlier blanket ban was lifted, and `meta.store[T comparable]` is the worked example: it holds the read surface and the unexported mutators that were identical between `Labels` and `Scalars`. Two things to weigh before reaching for one, both learned from that change:
- **Measure the per-instance cost.** A store is now exactly the map header it wraps. An earlier version carried a `self` pointer back to the embedding type so promoted mutators could return it — that doubled every store from 8 to 16 bytes, and signals own two apiece (~6% more bytes per mesh run). A still earlier draft stored a name for error messages, ~100 bytes per signal. Both were removed.
- **Watch the godoc.** Read methods promoted from an unexported generic type render with unresolved type parameters (`All() map[string]T`). Accepted for `meta`; see the note under Package notes before repeating it elsewhere.

A generic that ends up wrapped in one hand-written forwarding method per call site has usually not paid for itself. `internal/hook.Group[T]` lives at exactly that line **on purpose**: 18 hand-written registration wrappers across the three hook levels sit over its 4 methods, and they are the accepted price of keeping every `Hooks` struct's fields unexported so closures stay the only registration path. Do not "fix" either side.

The collection surfaces share `internal/collection` the same way: `Slice[T]` backs the groups and `Keyed[T]` the name-keyed collections, embedded (behind unexported type aliases) so the read surface promotes; slice plumbing goes through package functions (`collection.Items`/`SetItems`/`AppendItems`) rather than methods, so no mutator can promote onto `signal.Group`'s copy-on-write surface.

**Minimise `reflect`.** Only when no alternative exists. Current approved use: `reflect.TypeOf(payload).Comparable()` in `ContainsPayload` — always nil-guard before calling `.Comparable()`.

## Package notes

- **`signal`** — `payload` is `[]any{value}` (single-element slice so `nil` is valid). Predicate combinators and label constructors live in `predicates.go`. `ForEach`/`ForEachIf` return `error` only (as on every collection type — see [naming.md](naming.md)). Typed payload accessors live in `typed.go`: `As[T]` (error on nil signal / missing payload / wrong type), `AsOrDefault[T]`, fallible per-type shorthands over `As`, and `AsNumber` (loose `(float64, bool)` widening — `float64`/`float32`/`int`/`int64`/`uint64`, `bool` as 1/0). None of them panic; that is the point of having them. There are deliberately **no** `AsIntOrDefault`-style shorthands: `AsOrDefault` infers `T` from the default and is shorter. The sole exception is `AsFloat64OrDefault`, which exists because an untyped `0` infers `int`, so `AsOrDefault(s, 0)` silently returns the default for a float64 payload — do not "restore symmetry" by adding the others back.
- **`meta`** — `Labels` (string k/v) and `Scalars` (string→float64). `Keys()`/`Values()` return sorted slices for determinism. `Merge(other)` is the one non-mutating method on both types. `Every(pred)` on empty = `true` (vacuous truth). `ForEach` returns `error`. Constructors: `NewLabels()`, `NewScalars()`.
  Both embed the generic `store[T comparable]` in `store.go`, which holds the shared read surface
  plus unexported mutators. A store is exactly the map header it wraps — no `self` pointer, no
  name: an earlier version carried a pointer back to the embedding type so promoted mutators
  could return it, which doubled every store from 8 to 16 bytes with signals owning two apiece;
  a name for error messages measured ~100 extra bytes per signal. The price of dropping them is
  that the four chainable mutators (`Set`, `SetMany`, `Remove`, `Clear`) are declared on `Labels`
  and `Scalars`, wrapping the store's unexported ones — they need a concrete return type to keep
  `fm.Labels().Clear().SetMany(m)` compiling — and `Value` stays concrete so its error can name
  what was missing. Read methods are promoted; their return types never name the receiver.
  **Known and accepted:** `go doc` renders the promoted read methods with an unresolved type
  parameter (`All() map[string]T`). The methods are correct and callable; only the rendering is
  poor. Do **not** "fix" it by adding concrete forwarding methods on `Labels`/`Scalars` — that
  reintroduces the wrapper layer this consolidation removed. Decision taken 2026-07-31.
- **`port`** — `Flush()` fans out then clears source, firing `OnSignalsDelivered` on the source once per pipe after each destination accepts. That hook is the only event naming both ends of a pipe: `OnSignalsAdded` fires on the destination and cannot identify the sender. Its context struct is guarded by `hook.Group.IsEmpty()` because a `Trigger` argument escapes to the heap even with no hooks registered — on this path that would be one wasted allocation per pipe, per flush, per cycle. `PipeTo` is output→input only. Both return `error`. `PipeTo` validates direction at call time. `wiring.go` holds the declarative multi-edge helpers: `Pipe`/`MultiPipe` (registers connections) and `Pair`/`MultiForward` (copies signals now); both name the failing edge and report nil ports instead of dereferencing them.
- **Name lookups are silently forgiving — helpers taking port names must not be.** `Collection.ByName` returns `nil` for a name no port has, `Collection.ByNames` skips such names entirely, and `AllHaveSignals()`/`Every()` on the resulting empty collection is vacuously `true`. So `ByNames("typo").AllHaveSignals()` reports *ready*. Any helper that accepts port names as strings must resolve every name before asking anything about signals, and report the name it could not resolve.
- **`component`** — `State` is `map[string]any`, persistent across cycles and across `Run`s (see [runtime.md](runtime.md)). Constructors use functional options: `component.New(name, opts...) (*Component, error)`. Ports come in two creation styles: name-based (`WithInputs`/`AddInputs`, `WithIndexedInputs("i", 1, 3)` → `i1..i3`) and attach-based (`AttachInputPorts` for pre-built `port.NewInput` ports with options). `LoopbackPipe(out, in)` wires a component to itself (such a mesh never stops naturally). `ErrWaitingForInputs`/`ErrWaitKeepingInputs` are scheduler control-flow sentinels, not failures. `compose.go` holds the `ActivationFunc` combinators — `Sequential`, `When`+`HasSignalsOn`, `RequireInputs`, `Pipeline`+`PipelineStage` — which compose a component's *own* activation (contrast `OnActivation` hooks, which are for behavior added from outside; see [hooks.md](hooks.md)). `When` skips, `RequireInputs` suspends and keeps: never substitute one for the other, as skipping where waiting was meant drops the partial inputs at drain time.
- **`hook`** — lives at `internal/hook` (not public API). Generic `hook.Group[T]`, ordered, fail-fast `Trigger`. Three hook levels (mesh/component/port); see [hooks.md](hooks.md).
- **`cycle`** — has its own `Any`/`Every`/`Count` on its collection type, independent of `signal.Group`.

## Metadata tiers on groups/collections

Every Group and Collection type carries its **own** `*meta.Labels` and `*meta.Scalars` (Tier 1). Batch mutation of a container's **contents** (Tier 2a) exists on `signal.Group` **only** — the mutating collections lost their `Set{Label,Scalar}OnEach`/`Remove*OnEach` batch methods (iterate with `ForEach` and use each element's own store), and the cross-entity scalar aggregation tier (`Min/Max/Avg/SumScalar`, once Tier 2b) was removed with the scalar-statistics API. Do not reintroduce either.

| Tier | Methods | Where |
|---|---|---|
| 1 — entity's own | `Labels()`, `Scalars()`; mutate via `WithLabel`/`WithScalar` on `signal.Group` (CoW), `SetLabel`/`SetScalar` on `port.Group`, the live stores elsewhere | all groups/collections |
| 2a — batch on contents | `WithLabelOnEach(k,v)`, `WithScalarOnEach(k,v)`, `RemoveLabelOnEach(names...)`, `RemoveScalarOnEach(names...)` | `signal.Group` only (CoW) |

`signal.Group` batch methods (Tier 2a) preserve the group's own metadata on the returned group via `copyGroupMeta`.

**`signal.Group.Labels()`/`Scalars()` return clones.** The group is copy-on-write and the live stores were the one back door: mutating the returned store used to change the group in place. They now match `Signal.Labels()` — the only way to a modified group is `WithLabel`/`WithScalar`. Do not hand the live stores back out.

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
