# Design

## Architecture

| Concept | Type | Package |
|---|---|---|
| Data packet | `*Signal` | `signal` |
| Ordered signal collection | `*signal.Group` | `signal` |
| Key-value metadata | `*labels.Collection` | `labels` |
| Component | `*component.Component` | `component` |
| Data endpoint | `*port.Port` | `port` |
| Connection | pipe via `PipeTo` (output→input only) | `port` |
| Execution tick | `*cycle.Cycle` | `cycle` |

## Invariants

**`*Signal` and `*signal.Group` are copy-on-write.** Mutating methods return a new value; receiver is never modified. `cloneSignal(s)` is the single clone primitive — nil-safe, use it in all CoW methods.

**Payload is shallow-copied.** Mutable reference payloads (map, slice, pointer) must be treated as immutable by the caller. `nil` is a valid payload and must survive all CoW operations unchanged.

**`labels.Collection` is mutable.** It mutates in place. Do not make it CoW — `port` and `component` depend on mutation.

**Errors are returned directly.** Methods that can fail return `error` as the last return value. Infallible methods (transformations like `Filter`, `Map`, `With*` on `signal.Signal`) return their type directly for fluency. There is no "poison object" or chainable error field on any type.

**Fan-out shares pointers.** Output→input fan-out forwards the same `*Signal` pointers to all destinations. Do not add deep-copy to `ForwardSignals` or `Flush`.

**No generics.** `signal`/`labels` use `any`; FBP requires mixed-type signal flows in one group.

**Minimise `reflect`.** Only when no alternative exists. Current approved use: `reflect.TypeOf(payload).Comparable()` in `ContainsPayload` — always nil-guard before calling `.Comparable()`.

## Package notes

- **`signal`** — `payload` is `[]any{value}` (single-element slice so `nil` is valid). Predicate combinators and label constructors live in `predicates.go`. `ForEach` returns `(*Group, error)`.
- **`labels`** — `Keys()`/`Values()` return sorted slices for determinism. `Merge(other)` is the one non-mutating method. `Every(pred)` on empty = `true` (vacuous truth). `ForEach` returns `error`.
- **`port`** — `Flush()` fans out then clears source. `PipeTo` is output→input only. Both return `error`. `PipeTo` validates direction at call time.
- **`component`** — `State` is `map[string]any`, persistent across cycles. Constructors use functional options: `component.New(name, opts...) (*Component, error)`.
- **`cycle`** — has its own `Any`/`Every`/`Count` on its collection type, independent of `signal.Group`.
