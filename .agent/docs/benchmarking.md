# Benchmarking & Fuzzing

F-Mesh's priority is simplicity and a clean API, not performance. So:
- **Benchmarks are regression tripwires**, not optimization targets. They catch a change that adds
  an allocation, a copy or quadratic behavior.
- **Fuzz targets are invariant guards**, mainly for the copy-on-write (CoW) rule.

## What we measure

Every benchmark calls `b.ReportAllocs()` and reports:

- `ns/op` — wall time. **Noisy** on GitHub-hosted runners; trust it only as a relative delta over
  many samples.
- `B/op` — bytes allocated per op.
- `allocs/op` — heap allocation count. The most stable metric on CI and the best CoW signal: a CoW
  method's allocation count says how many copies it makes. A jump usually means a stray copy — or
  in-place mutation. **Watch it first.**

## What exists

Three mesh benchmarks, deliberately few. Add a case to one before adding a fourth.

| Benchmark | Answers |
|---|---|
| `BenchmarkMeshRun/{wide,fan-in,fan-out,pipeline}/<n>` | what a run costs, per topology and size |
| `BenchmarkMeshConstruction/<n>` | what building a mesh costs |
| `BenchmarkMeshCyclesPerSecond/{scheduling-only,with-signal-movement}/<size>` | how fast the loop ticks, with and without signal movement |

Package benchmarks (`signal`, `port`) guard specific invariants — CoW allocation counts,
allocation-free port traversal — and are named for what they guard.

## Comparing runs: benchstat

Never compare two single runs by eye; variance swamps the signal. Use
[`benchstat`](https://pkg.go.dev/golang.org/x/perf/cmd/benchstat):

```bash
go test -run='^$' -bench=. -benchmem -count=8 ./... | tee new.txt
# ...same on the other revision into old.txt...
benchstat old.txt new.txt
```

- `-count=8` gives benchstat enough samples. `~` in the output means "no significant change" —
  do not act on it.
- CI runs this per PR (`.github/workflows/bench.yml`) and posts a sticky comment: significant
  deltas on top, the full table collapsed below. It is **advisory and never fails the check** —
  do not gate merges on runner noise.

## Writing a benchmark

- Always `b.ReportAllocs()`.
- Use `for b.Loop() { ... }`, not `for range b.N`. `b.Loop()` keeps setup before the loop out of the
  timer (no `b.ResetTimer()` in the common case), keeps inputs and results alive, and stops the
  compiler from removing the body as dead code.
- Keep per-iteration setup out of the measured region. If it cannot be hoisted, wrap it in
  `b.StopTimer()`/`b.StartTimer()`.
- Use `b.RunParallel` only when contended concurrency is what you test.
- A benchmark that deliberately stalls must say so and disable livelock detection
  (`WithoutLivelockDetection()`), as `scheduling-only` does.

## The headline metric: activation cycles per second

The key number is **activation cycles per second** — how fast the scheduler drives full run-loop
ticks with every component activating each cycle. It measures library overhead *on top of* the
activation function, so `BenchmarkMeshCyclesPerSecond` keeps the activation body near-empty and
reports `cycles/s` and `activations/s` (= cycles/s × size) via `b.ReportMetric`. Swept over tens,
hundreds and thousands of components:

- **scheduling-only** — activation returns `component.ErrWaitKeepingInputs`, so each component
  keeps its input and re-activates every cycle, with no outputs or pipes. The scheduling floor
  (worker pool, `WaitGroup`, result collection).
- **with-signal-movement** — activation copies input → output through a self-loop pipe, adding the
  per-cycle drain/flush/forward cost.

`with-signal-movement − scheduling-only` ≈ the cost of moving signals. Each `Run` sustains N cycles
with `WithCyclesLimit(N)` + `WithUnlimitedTime()`, so `ErrReachedMaxAllowedCycles` is the expected
stop, not a failure. Inputs can persist across `Run` calls (the "keep" kind retains them), so each `Run`
first clears every input and puts exactly one signal back (`primeInputs`).

## Size sweeps reveal complexity class

A fixed size is one point and hides scaling. Sweep sizes to see the curve:

```go
for _, n := range []int{10, 100, 1_000, 10_000} {
    b.Run(strconv.Itoa(n), func(b *testing.B) { /* ... */ })
}
```

`BenchmarkGroupBuild` shows this way that building a group by repeated `With` is O(n²) (each `With`
allocates a new slice). `BenchmarkMeshRun/wide` exercises the per-component scheduling cost at scale.

Mesh-scale gotchas:

- **Wide, not deep.** A pipeline of N components needs N cycles. N > `CyclesLimit` (default 1000)
  or wall time > `TimeLimit` (default 5s) stops the mesh early. Scale benchmarks use *wide* meshes
  (everything activates in one cycle) or raise the limits.
- **Shallow copy is payload-size independent.** Signal CoW copies the interface header, not the
  payload. `BenchmarkGroupPayloadSize` keeps `ns/op` flat as payload grows — a tripwire for an
  accidental deep copy.
- **Fan-in is quadratic** — see "Scaling characteristics" in [runtime.md](runtime.md).
  `BenchmarkMeshRun/fan-in` sweeps the curve; it turns linear only if forwarding ever batches
  appends.
- **Long benchmarks accumulate history.** `RuntimeInfo.Cycles` keeps every cycle's results for the
  whole `Run` (~100 B × components × cycles), so `B/op` in sustained-cycle benchmarks includes
  history growth.

## When to add one

Add or extend a benchmark when a change touches a hot path — the run loop (`run.go`
`runCycle`/`drainComponents`), the port drain/flush path, or any CoW method on
`signal.Signal`/`signal.Group` — especially if it could add an allocation or a copy.

## Fuzzing

Native Go fuzzing (`func FuzzXxx(f *testing.F)`) checks **properties**, not just crashes. The key
property is CoW: a mutating-style method returns a new value and leaves the receiver (payload,
metadata) untouched. Targets live beside the code they guard (`signal/signal_fuzz_test.go`,
`signal/group_fuzz_test.go`, `port/collection_fuzz_test.go`).

- The seed corpus runs under plain `go test ./...` (fast, deterministic) — treat it like a table
  test.
- Explore with a time budget: `go test -run='^$' -fuzz=FuzzSignalCoW -fuzztime=30s ./signal/`
- Failing inputs land in `testdata/fuzz/<Target>/` — **keep them in the repo** (the user commits);
  they become permanent regression cases.
- `nil` is a valid payload — include it in seeds.

Fuzzing is not in the per-PR workflow (unbounded time). Run it locally or on a schedule.
